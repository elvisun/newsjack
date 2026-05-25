package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"time"
)

type originApplyOptions struct {
	AllowMissing bool
	AllowUnknown bool
	RunTime      time.Time
	WindowHours  float64
}

func cmdOriginApply(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("origin-apply", flag.ContinueOnError)
	fs.SetOutput(stderr)
	candidatesPath := fs.String("candidates", "", "Relevant candidates JSON")
	originsPath := fs.String("origins", "", "Story-origin findings JSON")
	outputPath := fs.String("output", "", "Output path")
	runTimeRaw := fs.String("run-time", "", "Override run timestamp for freshness cutoff")
	windowHours := fs.Float64("window-hours", 24.0, "Freshness window in hours")
	allowMissing := fs.Bool("allow-missing", false, "Do not fail when a candidate has no origin finding")
	allowUnknown := fs.Bool("allow-unknown", false, "Do not fail when origin findings reference unknown IDs")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *candidatesPath == "" || *originsPath == "" {
		return fail(stderr, errors.New("--candidates and --origins are required"))
	}
	candidates, err := readJSONMap(*candidatesPath)
	if err != nil {
		return fail(stderr, err)
	}
	originsPayload, err := readJSONAny(*originsPath)
	if err != nil {
		return fail(stderr, err)
	}
	runTime, err := originRunTime(candidates, *runTimeRaw)
	if err != nil {
		return fail(stderr, err)
	}
	output, err := applyOriginFindings(candidates, originsPayload, originApplyOptions{
		AllowMissing: *allowMissing,
		AllowUnknown: *allowUnknown,
		RunTime:      runTime,
		WindowHours:  *windowHours,
	})
	if err != nil {
		return fail(stderr, err)
	}
	data := marshalJSON(output)
	if *outputPath != "" {
		if err := os.WriteFile(expandPath(*outputPath), data, 0o644); err != nil {
			return fail(stderr, err)
		}
	} else {
		stdout.Write(data)
	}
	return 0
}

func applyOriginFindings(candidates map[string]any, originsPayload any, opts originApplyOptions) (map[string]any, error) {
	if opts.WindowHours <= 0 {
		return nil, errors.New("--window-hours must be positive")
	}
	signals := signalSlice(candidates["signals"])
	signalByID := map[string]map[string]any{}
	for _, signal := range signals {
		if id := stringValue(signal["id"]); id != "" {
			signalByID[id] = signal
		}
	}
	findings, err := normalizeOriginFindings(originsPayload)
	if err != nil {
		return nil, err
	}
	findingByID := map[string]map[string]any{}
	var errs []string
	for _, finding := range findings {
		id := strings.TrimSpace(stringValue(finding["signal_id"]))
		if id == "" {
			errs = append(errs, "origin finding missing signal_id")
			continue
		}
		if _, ok := findingByID[id]; ok {
			errs = append(errs, "duplicate origin finding for signal_id="+id)
			continue
		}
		if signalByID[id] == nil && !opts.AllowUnknown {
			errs = append(errs, "origin finding references unknown signal_id="+id)
			continue
		}
		findingByID[id] = normalizeOriginFinding(finding)
	}
	var missing []string
	for id := range signalByID {
		if findingByID[id] == nil {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 && !opts.AllowMissing {
		sample := strings.Join(firstN(missing, 8), ", ")
		suffix := ""
		if len(missing) > 8 {
			suffix = "..."
		}
		errs = append(errs, fmt.Sprintf("missing origin findings for %d signal(s): %s%s", len(missing), sample, suffix))
	}
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, "\n"))
	}

	var selected, rejected, missingSignals []map[string]any
	statusCounts := map[string]int{}
	cutoff := opts.RunTime.Add(-time.Duration(opts.WindowHours * float64(time.Hour)))
	for _, signal := range signals {
		id := stringValue(signal["id"])
		finding := findingByID[id]
		if finding == nil {
			missingSignals = append(missingSignals, summarySignal(signal))
			continue
		}
		gate := computeFreshnessGate(finding, opts.RunTime, cutoff, opts.WindowHours)
		status := stringValue(gate["computed_status"])
		statusCounts[status]++
		withOrigin := cloneMap(signal)
		withOrigin["story_origin"] = finding
		withOrigin["freshness_gate"] = gate
		if status == "fresh" || status == "fresh_new_development" {
			selected = append(selected, withOrigin)
			continue
		}
		s := summarySignal(signal)
		s["story_origin"] = finding
		s["freshness_gate"] = gate
		rejected = append(rejected, s)
	}
	return map[string]any{
		"version":      1,
		"generated_at": time.Now().UTC().Format(time.RFC3339Nano),
		"monitor":      valueOrEmptyMap(candidates["monitor"]),
		"signals":      selected,
		"coarse_relevance": firstNonNil(
			candidates["coarse_relevance"],
			candidates["coarse_filter"],
			candidates["cheap_filter"],
		),
		"freshness_gate": map[string]any{
			"input_signal_count":      len(signals),
			"origin_finding_count":    len(findingByID),
			"selected_count":          len(selected),
			"rejected_count":          len(rejected),
			"missing_count":           len(missingSignals),
			"status_counts":           sortedCountMap(statusCounts),
			"freshness_window_hours":  opts.WindowHours,
			"run_generated_at":        opts.RunTime.UTC().Format(time.RFC3339Nano),
			"freshness_cutoff":        cutoff.UTC().Format(time.RFC3339Nano),
			"included_statuses":       []string{"fresh", "fresh_new_development"},
			"rejected_signals":        rejected,
			"missing_signals":         missingSignals,
			"deterministic_authority": true,
		},
		"detector_diagnostics": valueOrEmptyMap(candidates["detector_diagnostics"]),
		"source_errors":        valueOrEmptyMap(candidates["source_errors"]),
	}, nil
}

func normalizeOriginFindings(payload any) ([]map[string]any, error) {
	if arr, ok := payload.([]any); ok {
		var out []map[string]any
		for _, item := range arr {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		return out, nil
	}
	if m, ok := payload.(map[string]any); ok {
		for _, key := range []string{"findings", "origins", "story_origins"} {
			if arr, ok := m[key].([]any); ok {
				var out []map[string]any
				for _, item := range arr {
					if d, ok := item.(map[string]any); ok {
						out = append(out, d)
					}
				}
				return out, nil
			}
		}
	}
	return nil, errors.New("origin findings JSON must be a list or an object with findings/origins list")
}

func normalizeOriginFinding(finding map[string]any) map[string]any {
	out := cloneMap(finding)
	if fp := valueOrEmptyMap(finding["first_publication"]); len(fp) > 0 {
		for _, key := range []string{
			"status", "surfaced_article_published_at", "first_public_at", "original_url",
			"original_source", "canonical_coverage_url", "canonical_coverage_source",
			"canonical_coverage_published_at", "canonical_coverage_basis",
			"same_story_basis", "new_development", "new_development_at", "confidence",
			"timestamp_evidence", "evidence_urls", "rationale",
		} {
			if out[key] == nil {
				out[key] = fp[key]
			}
		}
	}
	return out
}

func computeFreshnessGate(origin map[string]any, runTime, cutoff time.Time, windowHours float64) map[string]any {
	workerStatus := firstString(origin["status"], valueOrEmptyMap(origin["first_publication"])["status"])
	newRaw := firstString(origin["new_development_at"], valueOrEmptyMap(origin["first_publication"])["new_development_at"])
	firstRaw := firstString(origin["first_public_at"], valueOrEmptyMap(origin["first_publication"])["first_public_at"])
	if parsed, precision, ok := parseOriginTimestamp(newRaw); ok {
		relation := compareOriginTimestamp(parsed, precision, runTime, cutoff)
		if relation == "fresh" {
			return freshnessGate("fresh_new_development", runTime, cutoff, windowHours, workerStatus, "new_development_at", newRaw, precision, "new_development_at is inside the freshness window")
		}
		if relation == "ambiguous" {
			return freshnessGate("freshness_unverified", runTime, cutoff, windowHours, workerStatus, "new_development_at", newRaw, precision, "new_development_at has date-only precision on the cutoff date")
		}
		if firstRaw == "" {
			return freshnessGate("stale", runTime, cutoff, windowHours, workerStatus, "new_development_at", newRaw, precision, "new_development_at is before the freshness cutoff")
		}
	}
	if parsed, precision, ok := parseOriginTimestamp(firstRaw); ok {
		relation := compareOriginTimestamp(parsed, precision, runTime, cutoff)
		switch relation {
		case "fresh":
			return freshnessGate("fresh", runTime, cutoff, windowHours, workerStatus, "first_public_at", firstRaw, precision, "first_public_at is inside the freshness window")
		case "stale":
			return freshnessGate("stale", runTime, cutoff, windowHours, workerStatus, "first_public_at", firstRaw, precision, "first_public_at is before the freshness cutoff")
		default:
			return freshnessGate("freshness_unverified", runTime, cutoff, windowHours, workerStatus, "first_public_at", firstRaw, precision, "first_public_at has date-only precision on the cutoff date")
		}
	}
	return freshnessGate("freshness_unverified", runTime, cutoff, windowHours, workerStatus, "", "", "", "first_public_at is missing or unparseable")
}

func freshnessGate(status string, runTime, cutoff time.Time, windowHours float64, workerStatus, basisField, basisValue, precision, rationale string) map[string]any {
	return map[string]any{
		"computed_status":         status,
		"worker_status":           nullableString(workerStatus),
		"run_generated_at":        runTime.UTC().Format(time.RFC3339Nano),
		"freshness_cutoff":        cutoff.UTC().Format(time.RFC3339Nano),
		"freshness_window_hours":  windowHours,
		"basis_field":             nullableString(basisField),
		"basis_value":             nullableString(basisValue),
		"basis_precision":         nullableString(precision),
		"rationale":               rationale,
		"deterministic_authority": true,
	}
}

func parseOriginTimestamp(value string) (time.Time, string, bool) {
	raw := strings.TrimSpace(value)
	if raw == "" || strings.EqualFold(raw, "null") {
		return time.Time{}, "", false
	}
	dateOnly := len(raw) == 10 && raw[4] == '-' && raw[7] == '-'
	parsed, ok := parseTime(raw)
	if !ok {
		return time.Time{}, "", false
	}
	if dateOnly {
		return parsed, "date", true
	}
	return parsed, "time", true
}

func compareOriginTimestamp(parsed time.Time, precision string, runTime, cutoff time.Time) string {
	parsed = parsed.UTC()
	runTime = runTime.UTC()
	cutoff = cutoff.UTC()
	if precision == "date" {
		day := parsed.Format("2006-01-02")
		runDay := runTime.Format("2006-01-02")
		cutoffDay := cutoff.Format("2006-01-02")
		switch {
		case day == runDay || day > cutoffDay:
			return "fresh"
		case day < cutoffDay:
			return "stale"
		default:
			return "ambiguous"
		}
	}
	if !parsed.Before(cutoff) {
		return "fresh"
	}
	return "stale"
}

func originRunTime(candidates map[string]any, override string) (time.Time, error) {
	if strings.TrimSpace(override) != "" {
		if parsed, ok := parseTime(override); ok {
			return parsed.UTC(), nil
		}
		return time.Time{}, fmt.Errorf("could not parse --run-time=%q", override)
	}
	monitor := valueOrEmptyMap(candidates["monitor"])
	for _, raw := range []string{stringValue(monitor["generated_at"]), stringValue(candidates["generated_at"])} {
		if parsed, ok := parseTime(raw); ok {
			return parsed.UTC(), nil
		}
	}
	return time.Now().UTC(), nil
}
