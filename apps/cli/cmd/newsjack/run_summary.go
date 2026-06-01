package main

import (
	"errors"
	"flag"
	"io"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// cmdRunSummary writes a deterministic machine-readable summary of a run
// artifact. It emits JSON only; human-facing report rendering belongs in skills.
func cmdRunSummary(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("run-summary", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outputPath := fs.String("output", "", "Path to write summary JSON; stdout when omitted")
	top := fs.Int("top", 25, "Number of input and debug-dropped signals to include")
	if err := fs.Parse(reorderIntermixedFlags(args, stringSet([]string{"output", "top"}))); err != nil {
		return 2
	}
	if fs.NArg() != 1 {
		return fail(stderr, errors.New("usage: newsjack run-summary INPUT [--output summary.json]"))
	}
	payload, err := readJSONMap(fs.Arg(0))
	if err != nil {
		return fail(stderr, err)
	}
	generatedAt := time.Now().UTC().Format(time.RFC3339Nano)
	summary := summarizeRunAt(payload, expandPath(fs.Arg(0)), maxInt(0, *top), generatedAt)
	data := marshalJSON(summary)
	if *outputPath == "" {
		stdout.Write(data)
		return 0
	}
	if err := os.MkdirAll(filepath.Dir(expandPath(*outputPath)), 0o755); err != nil {
		return fail(stderr, err)
	}
	if err := os.WriteFile(expandPath(*outputPath), data, 0o644); err != nil {
		return fail(stderr, err)
	}
	summary = summarizeRunAt(payload, expandPath(fs.Arg(0)), maxInt(0, *top), generatedAt)
	if err := os.WriteFile(expandPath(*outputPath), marshalJSON(summary), 0o644); err != nil {
		return fail(stderr, err)
	}
	return 0
}

func summarizeRunAt(payload map[string]any, inputPath string, top int, generatedAt string) map[string]any {
	signals := signalSlice(payload["signals"])
	diagnostics := valueOrEmptyMap(firstNonNil(payload["diagnostics"], payload["detector_diagnostics"]))
	debug := valueOrEmptyMap(payload["debug"])
	allScored := signalSlice(debug["all_scored_signals"])
	selectedIDs := map[string]bool{}
	for _, signal := range signals {
		if id := stringValue(signal["id"]); id != "" {
			selectedIDs[id] = true
		}
	}
	var allIDs []string
	var dropped, selectedDebug []map[string]any
	for _, signal := range allScored {
		id := stringValue(signal["id"])
		if id != "" {
			allIDs = append(allIDs, id)
		}
		if selectedIDs[id] {
			selectedDebug = append(selectedDebug, signal)
		} else {
			dropped = append(dropped, signal)
		}
	}
	runDir := filepath.Dir(inputPath)
	paths := artifactPaths(runDir)
	monitor := valueOrEmptyMap(payload["monitor"])
	sourceErrors := valueOrEmptyMap(payload["source_errors"])
	sort.SliceStable(dropped, func(i, j int) bool { return queuePriority(dropped[i]) > queuePriority(dropped[j]) })

	rejectedIDs := loadCoarseRejections(paths["coarse_relevance_decisions"])
	rejectedInInput, safetyInInput := countDeterministicExclusions(signals, rejectedIDs)
	return map[string]any{
		"generated_at": generatedAt,
		"input_path":   inputPath,
		"run_dir":      runDir,
		"artifacts":    artifactStatus(paths),
		"pipeline":     pipelineStatus(paths),
		"monitor": map[string]any{
			"name":              monitor["name"],
			"generated_at":      monitor["generated_at"],
			"profile_name":      profileName(monitor),
			"queries":           valueOrEmptyArray(monitor["queries"]),
			"feed_urls":         valueOrEmptyArray(monitor["feed_urls"]),
			"sources_requested": valueOrEmptyArray(monitor["sources_requested"]),
			"sources_used":      valueOrEmptyArray(monitor["sources_used"]),
			"lookback_days":     monitor["lookback_days"],
			"max_age_hours":     monitor["max_age_hours"],
			"depth":             monitor["depth"],
			"mock":              monitor["mock"],
		},
		"counts": map[string]any{
			"selected_unique_signals":        len(signals),
			"total_scored_signals":           firstNonNil(diagnostics["total_scored_signals"], len(allScored)),
			"total_emitted_signals":          firstNonNil(diagnostics["total_emitted_signals"], len(signals)),
			"debug_all_scored_rows":          len(allScored),
			"debug_unique_scored_signal_ids": len(stringSet(allIDs)),
			"debug_selected_rows":            len(selectedDebug),
			"debug_unselected_rows":          len(dropped),
			"debug_duplicate_scored_rows":    len(allIDs) - len(stringSet(allIDs)),
			"source_errors":                  len(sourceErrors),
		},
		"selection":                 valueOrEmptyMap(diagnostics["selection"]),
		"lanes":                     map[string]any{"scored": firstNonNil(diagnostics["signals_by_lane"], countByLanes(allScored)), "emitted": firstNonNil(diagnostics["emitted_by_lane"], countByLanes(signals)), "dropped_debug": countByLanes(dropped)},
		"sources":                   map[string]any{"evidence_by_source": firstNonNil(diagnostics["evidence_by_source"], countEvidenceSources(signals)), "source_errors": sourceErrors, "source_status": diagnostics["source_status"]},
		"hygiene_rejections":        valueOrEmptyMap(diagnostics["hygiene_rejections"]),
		"coarse_relevance":          coarseRelevanceMap(payload),
		"coarse_relevance_file":     summarizeDecisions(paths["coarse_relevance_decisions"]),
		"relevant_candidates_file":  summarizeTargeted(paths["relevant_candidates"]),
		"clustered_candidates_file": summarizeTargeted(paths["clustered_candidates"]),
		"origin_findings_file":      summarizeOriginFindings(paths["origin_findings"]),
		"targeted_candidates_file":  summarizeTargeted(paths["targeted_candidates"]),
		"triaged_candidates_file":   summarizeTargeted(paths["triaged_candidates"]),
		"final_report_file":         summarizeMarkdownArtifact(paths["final_report"]),
		"run_report_file":           summarizeMarkdownArtifact(paths["run_report"]),
		"top_signals":               summarizeSignals(firstNSignals(signals, top)),
		"top_dropped_signals":       summarizeSignals(firstNSignals(dropped, top)),
		"deterministic_exclusions": map[string]any{
			"coarse_rejected_in_input": rejectedInInput,
			"safety_flagged_in_input":  safetyInInput,
		},
	}
}

// loadCoarseRejections reads coarse_relevance_decisions.json from the run dir
// and returns the set of signal IDs the coarse pass rejected. Missing or
// unreadable file => empty set.
func loadCoarseRejections(path string) map[string]bool {
	out := map[string]bool{}
	if !fileExists(path) {
		return out
	}
	payload, err := readJSONMap(path)
	if err != nil {
		return out
	}
	for _, d := range mapSlice(payload["decisions"]) {
		if firstString(d["decision"], "") == "reject" {
			if id := stringValue(d["signal_id"]); id != "" {
				out[id] = true
			}
		}
	}
	return out
}

// hasHardSafetyFlag reports whether a signal carries a deterministic hard
// safety flag (tragedy/human-suffering term) from the detector.
func hasHardSafetyFlag(signal map[string]any) bool {
	features := valueOrEmptyMap(signal["features"])
	for _, raw := range anySlice(features["safety_flags"]) {
		if f, ok := raw.(map[string]any); ok {
			if firstString(f["type"], "") == "hard_safety_term" {
				return true
			}
		}
	}
	return false
}

func countDeterministicExclusions(signals []map[string]any, rejectedIDs map[string]bool) (rejected, safety int) {
	for _, signal := range signals {
		if id := stringValue(signal["id"]); id != "" && rejectedIDs[id] {
			rejected++
		}
		if hasHardSafetyFlag(signal) {
			safety++
		}
	}
	return rejected, safety
}

func artifactPaths(runDir string) map[string]string {
	return map[string]string{
		"candidates":                 filepath.Join(runDir, "candidates.json"),
		"detector_summary":           filepath.Join(runDir, "summary.json"),
		"scored_candidates":          filepath.Join(runDir, "scored_candidates.json"),
		"commands":                   filepath.Join(runDir, "commands.log"),
		"detector_stderr":            filepath.Join(runDir, "detector.stderr.log"),
		"coarse_relevance_decisions": filepath.Join(runDir, "coarse_relevance_decisions.json"),
		"relevant_candidates":        filepath.Join(runDir, "relevant_candidates.json"),
		"clustered_candidates":       filepath.Join(runDir, "clustered_candidates.json"),
		"origin_findings":            filepath.Join(runDir, "origin_findings.json"),
		"targeted_candidates":        filepath.Join(runDir, "targeted_candidates.json"),
		"triaged_candidates":         filepath.Join(runDir, "triaged_candidates.json"),
		"final_report":               filepath.Join(runDir, "final_report.md"),
		"run_report":                 filepath.Join(runDir, "run.md"),
	}
}

func artifactStatus(paths map[string]string) map[string]any {
	out := map[string]any{}
	for name, path := range paths {
		info, err := os.Stat(path)
		exists := err == nil
		size := int64(0)
		if exists {
			size = info.Size()
		}
		out[name] = map[string]any{"path": path, "exists": exists, "bytes": size}
	}
	return out
}

func pipelineStatus(paths map[string]string) []map[string]any {
	return []map[string]any{
		stage("detector", paths["candidates"]),
		stage("coarse_relevance", paths["coarse_relevance_decisions"]),
		stage("relevance_apply", paths["relevant_candidates"]),
		stage("clustering", paths["clustered_candidates"]),
		stage("story_origin", paths["origin_findings"]),
		stage("freshness_gate", paths["targeted_candidates"]),
		stage("standing_triage", paths["triaged_candidates"]),
		stage("final_report", paths["final_report"]),
		stage("run_report", paths["run_report"]),
	}
}

func stage(name, path string) map[string]any {
	status := "pending"
	if fileExists(path) {
		status = "done"
	}
	return map[string]any{"stage": name, "status": status, "artifact": filepath.Base(path)}
}

func summarizeDecisions(path string) map[string]any {
	if !fileExists(path) {
		return map[string]any{"exists": false, "path": path}
	}
	payload, err := readJSONMap(path)
	if err != nil {
		return map[string]any{"exists": true, "path": path, "error": err.Error()}
	}
	decisions := mapSlice(payload["decisions"])
	outcome := map[string]int{}
	reasons := map[string]int{}
	for _, d := range decisions {
		outcome[firstString(d["decision"], "unknown")]++
		reasons[firstString(d["reason"], "unknown")]++
	}
	return map[string]any{"exists": true, "path": path, "decision_count": len(decisions), "decisions_by_outcome": outcome, "decisions_by_reason": reasons}
}

func summarizeTargeted(path string) map[string]any {
	if !fileExists(path) {
		return map[string]any{"exists": false, "path": path}
	}
	payload, err := readJSONMap(path)
	if err != nil {
		return map[string]any{"exists": true, "path": path, "error": err.Error()}
	}
	gate := freshnessGateMap(payload)
	coarse := coarseRelevanceMap(payload)
	return map[string]any{"exists": true, "path": path, "selected_signals": len(signalSlice(payload["signals"])), "input_signals": firstNonNil(gate["input_signal_count"], coarse["input_signal_count"]), "rejected_signals": firstNonNil(gate["rejected_count"], coarse["rejected_count"])}
}

func summarizeOriginFindings(path string) map[string]any {
	if !fileExists(path) {
		return map[string]any{"exists": false, "path": path}
	}
	payload, err := readJSONAny(path)
	if err != nil {
		return map[string]any{"exists": true, "path": path, "error": err.Error()}
	}
	findings, err := normalizeOriginFindings(payload)
	if err != nil {
		return map[string]any{"exists": true, "path": path, "error": err.Error()}
	}
	return map[string]any{"exists": true, "path": path, "finding_count": len(findings)}
}

func summarizeMarkdownArtifact(path string) map[string]any {
	if !fileExists(path) {
		return map[string]any{"exists": false, "path": path}
	}
	info, err := os.Stat(path)
	if err != nil {
		return map[string]any{"exists": true, "path": path, "error": err.Error()}
	}
	return map[string]any{"exists": true, "path": path, "bytes": info.Size()}
}

func profileName(monitor map[string]any) any {
	profile := valueOrEmptyMap(monitor["profile"])
	if v := firstString(profile["name"], profile["company"], profile["client"]); v != "" {
		return v
	}
	return nil
}

func countEvidenceSources(signals []map[string]any) map[string]int {
	out := map[string]int{}
	for _, signal := range signals {
		for _, raw := range anySlice(signal["evidence"]) {
			if ev, ok := raw.(map[string]any); ok {
				out[firstString(ev["source"], "unknown")]++
			}
		}
	}
	return out
}

func summarizeSignals(signals []map[string]any) []map[string]any {
	var out []map[string]any
	for _, signal := range signals {
		routing := valueOrEmptyMap(signal["routing"])
		mech := valueOrEmptyMap(signal["mechanical_scores"])
		out = append(out, map[string]any{"id": signal["id"], "title": firstString(signal["title"], firstEvidenceValue(signal, "title")), "query": signal["query"], "lane": routing["lane"], "queue_priority": routing["queue_priority"], "decay_bucket": firstNonNil(signal["decay_bucket"], mech["decay_bucket"]), "profile_match": mech["profile_match"], "major_news": mech["major_news"], "momentum": mech["momentum"], "source_agreement": mech["source_agreement"], "story_size": signal["story_size"], "coarse_relevance": firstNonNil(signal["coarse_relevance"], signal["coarse_filter"], signal["cheap_filter"]), "story_origin": signal["story_origin"], "freshness_gate": signal["freshness_gate"], "evidence": summarizeEvidence(signal)})
	}
	return out
}

func coarseRelevanceMap(payload map[string]any) map[string]any {
	if coarse := valueOrEmptyMap(payload["coarse_relevance"]); len(coarse) > 0 {
		return coarse
	}
	if coarse := valueOrEmptyMap(payload["coarse_filter"]); len(coarse) > 0 {
		return coarse
	}
	return valueOrEmptyMap(payload["cheap_filter"])
}

func freshnessGateMap(payload map[string]any) map[string]any {
	return valueOrEmptyMap(payload["freshness_gate"])
}

func summarizeEvidence(signal map[string]any) []map[string]any {
	var out []map[string]any
	for _, raw := range anySlice(signal["evidence"]) {
		if ev, ok := raw.(map[string]any); ok {
			out = append(out, map[string]any{"source": ev["source"], "title": ev["title"], "url": ev["url"], "published_at": ev["published_at"], "author": ev["author"], "engagement": valueOrEmptyMap(ev["engagement"]), "metadata": valueOrEmptyMap(ev["metadata"])})
		}
	}
	return out
}

func firstEvidenceValue(signal map[string]any, key string) any {
	for _, raw := range anySlice(signal["evidence"]) {
		if ev, ok := raw.(map[string]any); ok && ev[key] != nil && stringValue(ev[key]) != "" {
			return ev[key]
		}
	}
	return nil
}

func firstNSignals(signals []map[string]any, n int) []map[string]any {
	if n < 0 {
		n = 0
	}
	if len(signals) > n {
		return signals[:n]
	}
	return signals
}
