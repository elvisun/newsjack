package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

func cmdSummarizeRun(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("summarize-run", flag.ContinueOnError)
	fs.SetOutput(stderr)
	outputPath := fs.String("output", "", "Path to write machine-readable summary JSON")
	markdownPath := fs.String("markdown", "", "Path to write Markdown report")
	briefPath := fs.String("brief", "", "Deprecated alias for --markdown")
	top := fs.Int("top", 25, "Number of selected and dropped signals to include")
	if err := fs.Parse(reorderIntermixedFlags(args, stringSet([]string{"output", "markdown", "brief", "top"}))); err != nil {
		return 2
	}
	if fs.NArg() != 1 || *outputPath == "" {
		return fail(stderr, errors.New("usage: newsjack summarize-run INPUT --output summary.json --markdown run.md"))
	}
	mdPath := firstString(*markdownPath, *briefPath)
	if mdPath == "" {
		return fail(stderr, errors.New("--markdown is required"))
	}
	payload, err := readJSONMap(fs.Arg(0))
	if err != nil {
		return fail(stderr, err)
	}
	summary := summarizeRun(payload, expandPath(fs.Arg(0)), maxInt(0, *top))
	if err := os.MkdirAll(filepath.Dir(expandPath(*outputPath)), 0o755); err != nil {
		return fail(stderr, err)
	}
	if err := os.MkdirAll(filepath.Dir(expandPath(mdPath)), 0o755); err != nil {
		return fail(stderr, err)
	}
	if err := os.WriteFile(expandPath(*outputPath), marshalJSON(summary), 0o644); err != nil {
		return fail(stderr, err)
	}
	if err := os.WriteFile(expandPath(mdPath), []byte(renderSummaryMarkdown(summary)), 0o644); err != nil {
		return fail(stderr, err)
	}
	return 0
}

func summarizeRun(payload map[string]any, inputPath string, top int) map[string]any {
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
	return map[string]any{
		"generated_at": time.Now().UTC().Format(time.RFC3339Nano),
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
		"selection":                valueOrEmptyMap(diagnostics["selection"]),
		"lanes":                    map[string]any{"scored": firstNonNil(diagnostics["signals_by_lane"], countByLanes(allScored)), "emitted": firstNonNil(diagnostics["emitted_by_lane"], countByLanes(signals)), "dropped_debug": countByLanes(dropped)},
		"sources":                  map[string]any{"evidence_by_source": firstNonNil(diagnostics["evidence_by_source"], countEvidenceSources(signals)), "source_errors": sourceErrors},
		"hygiene_rejections":       valueOrEmptyMap(diagnostics["hygiene_rejections"]),
		"coarse_relevance":         coarseRelevanceMap(payload),
		"coarse_relevance_file":    summarizeDecisions(paths["coarse_relevance_decisions"]),
		"relevant_candidates_file": summarizeTargeted(paths["relevant_candidates"]),
		"origin_findings_file":     summarizeOriginFindings(paths["origin_findings"]),
		"targeted_candidates_file": summarizeTargeted(paths["targeted_candidates"]),
		"final_report_file":        summarizeFinalReport(paths["final_report"]),
		"top_signals":              summarizeSignals(firstNSignals(signals, top)),
		"top_dropped_signals":      summarizeSignals(firstNSignals(dropped, top)),
	}
}

func renderSummaryMarkdown(summary map[string]any) string {
	monitor := valueOrEmptyMap(summary["monitor"])
	counts := valueOrEmptyMap(summary["counts"])
	profile := firstString(monitor["profile_name"], "Newsjack")
	var lines []string
	lines = append(lines, "# "+mdInline(profile)+" Newsjack Brief", "")
	lines = append(lines, renderTable([][2]any{{"Status", statusText(summary)}, {"Generated", formatDatetime(firstString(monitor["generated_at"], summary["generated_at"]))}, {"Detector candidates", counts["selected_unique_signals"]}})...)
	if finalReport := finalReportContent(summary); finalReport != "" {
		lines = append(lines, "", "## Editorial Verdict", "", finalReport)
	}
	lines = append(lines, "", "## Candidate Preview", "")
	for i, signal := range mapSlice(summary["top_signals"]) {
		lines = append(lines, fmt.Sprintf("%d. **%s**", i+1, mdInline(firstString(signal["title"], "(untitled)"))))
		lines = append(lines, fmt.Sprintf("   - Why surfaced: %s; queue %s, profile %s, major %s, story size %s.", label(signal["lane"]), fmtValue(signal["queue_priority"]), fmtValue(signal["profile_match"]), fmtValue(signal["major_news"]), formatStorySize(signal["story_size"])))
		for _, raw := range anySlice(signal["evidence"]) {
			if ev, ok := raw.(map[string]any); ok {
				lines = append(lines, "   - "+renderEvidenceLink(ev))
			}
		}
	}
	if len(mapSlice(summary["top_signals"])) == 0 {
		lines = append(lines, "- (none)")
	}
	lines = append(lines, "", "## What Was Scanned", "")
	lines = append(lines, fmt.Sprintf("- **Profile:** %s", mdInline(firstString(monitor["profile_name"], "(unknown)"))))
	lines = append(lines, fmt.Sprintf("- **Queries:** %s", mdInline(formatList(valueOrEmptyArray(monitor["queries"]), 8))))
	lines = append(lines, "", "## Appendix: Provenance", "", "### Pipeline", "")
	lines = append(lines, renderPipeline(valueOrEmptyArray(summary["pipeline"]))...)
	lines = append(lines, "", "### Detector Counts", "")
	lines = append(lines, renderTable([][2]any{{"scored", counts["total_scored_signals"]}, {"selected", counts["selected_unique_signals"]}, {"source_errors", counts["source_errors"]}})...)
	return strings.TrimRight(strings.Join(lines, "\n"), "\n") + "\n"
}

func statusText(summary map[string]any) string {
	finalReport := valueOrEmptyMap(summary["final_report_file"])
	if truthy(finalReport["exists"], false) {
		return "Editorial review complete"
	}
	return "Detector preview only"
}

func finalReportContent(summary map[string]any) string {
	finalReport := valueOrEmptyMap(summary["final_report_file"])
	if !truthy(finalReport["exists"], false) {
		return ""
	}
	return strings.TrimSpace(stringValue(finalReport["content"]))
}

func artifactPaths(runDir string) map[string]string {
	return map[string]string{
		"candidates":                 filepath.Join(runDir, "candidates.json"),
		"detector_summary":           filepath.Join(runDir, "summary.json"),
		"commands":                   filepath.Join(runDir, "commands.log"),
		"detector_stderr":            filepath.Join(runDir, "detector.stderr.log"),
		"coarse_relevance_decisions": filepath.Join(runDir, "coarse_relevance_decisions.json"),
		"relevant_candidates":        filepath.Join(runDir, "relevant_candidates.json"),
		"origin_findings":            filepath.Join(runDir, "origin_findings.json"),
		"targeted_candidates":        filepath.Join(runDir, "targeted_candidates.json"),
		"final_report":               filepath.Join(runDir, "final_report.md"),
		"run_markdown":               filepath.Join(runDir, "run.md"),
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
		stage("story_origin", paths["origin_findings"]),
		stage("freshness_gate", paths["targeted_candidates"]),
		stage("final_report", paths["final_report"]),
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

func summarizeFinalReport(path string) map[string]any {
	if !fileExists(path) {
		return map[string]any{"exists": false, "path": path}
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return map[string]any{"exists": true, "path": path, "error": err.Error()}
	}
	return map[string]any{"exists": true, "path": path, "bytes": len(data), "content": string(data)}
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

func formatStorySize(value any) string {
	storySize := valueOrEmptyMap(value)
	if len(storySize) == 0 {
		return "-"
	}
	band := firstString(storySize["band"], "unknown")
	score := fmtValue(storySize["score"])
	confidence := firstString(storySize["confidence"], "unknown")
	return fmt.Sprintf("%s/%s (%s)", band, score, confidence)
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

func renderEvidenceLink(ev map[string]any) string {
	source := label(ev["source"])
	title := mdInline(firstString(ev["title"], "(no title)"))
	u := strings.TrimSpace(stringValue(ev["url"]))
	suffix := ""
	if published := stringValue(ev["published_at"]); published != "" {
		suffix = " (" + mdInline(published) + ")"
	}
	if u != "" {
		return fmt.Sprintf("%s: [%s](%s)%s", source, escapeLinkText(title), u, suffix)
	}
	return source + ": " + title + suffix
}

func renderPipeline(stages []any) []string {
	var rows [][2]any
	for _, raw := range stages {
		if st, ok := raw.(map[string]any); ok {
			rows = append(rows, [2]any{st["stage"], fmt.Sprintf("%v - %v", st["status"], st["artifact"])})
		}
	}
	return renderTable(rows)
}

func renderTable(rows [][2]any) []string {
	if len(rows) == 0 {
		return []string{"- (none)"}
	}
	lines := []string{"| key | value |", "|---|---|"}
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("| %s | %s |", mdCell(row[0]), mdCell(row[1])))
	}
	return lines
}

func mdCell(v any) string { return strings.ReplaceAll(mdInline(v), "|", `\|`) }
func mdInline(v any) string {
	return strings.Join(strings.Fields(fmt.Sprint(v)), " ")
}
func label(v any) string {
	return strings.ReplaceAll(strings.ReplaceAll(mdInline(v), "_", " "), "-", " ")
}
func escapeLinkText(v string) string {
	return strings.ReplaceAll(strings.ReplaceAll(v, "[", `\[`), "]", `\]`)
}

func formatDatetime(v string) string {
	if v == "" || v == "<nil>" {
		return "(unknown)"
	}
	raw := strings.TrimSuffix(v, "Z") + strings.TrimPrefix("Z", "Z")
	if strings.HasSuffix(v, "Z") {
		raw = strings.TrimSuffix(v, "Z") + "+00:00"
	}
	if parsed, ok := parseTime(raw); ok {
		return parsed.UTC().Format("2006-01-02 15:04 UTC")
	}
	return v
}

func formatList(values []any, limit int) string {
	var clean []string
	for _, v := range values {
		if s := mdInline(v); s != "" {
			clean = append(clean, s)
		}
	}
	if len(clean) == 0 {
		return "(none)"
	}
	if len(clean) <= limit {
		return strings.Join(clean, ", ")
	}
	return strings.Join(clean[:limit], ", ") + fmt.Sprintf(", plus %d more", len(clean)-limit)
}

func fmtValue(v any) string {
	if v == nil {
		return "-"
	}
	if f, ok := numberValue(v); ok {
		return strconv.FormatFloat(f, 'g', 3, 64)
	}
	return fmt.Sprint(v)
}
