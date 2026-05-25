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

var allowedDecisions = stringSet([]string{"keep", "monitor_only", "reject"})
var allowedReasons = stringSet([]string{"relevant_news", "plausible_client_bridge", "major_news_no_bridge", "keyword_collision", "not_news", "owned_docs_or_product_page", "seo_landing_page", "low_reach_x_post", "safety_risk", "duplicate", "off_beat", "no_profile_bridge"})

func cmdFilterApply(args []string, stdout, stderr io.Writer) int {
	var includes stringList
	fs := flag.NewFlagSet("filter-apply", flag.ContinueOnError)
	fs.SetOutput(stderr)
	candidatesPath := fs.String("candidates", "", "Detector JSON output")
	decisionsPath := fs.String("decisions", "", "Coarse-relevance decision JSON")
	outputPath := fs.String("output", "", "Output path")
	fs.Var(&includes, "include", "Decision to include. Repeatable.")
	allowMissing := fs.Bool("allow-missing", false, "Do not fail when a candidate has no decision")
	allowUnknown := fs.Bool("allow-unknown", false, "Do not fail when decisions reference unknown IDs")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *candidatesPath == "" || *decisionsPath == "" {
		return fail(stderr, errors.New("--candidates and --decisions are required"))
	}
	candidates, err := readJSONMap(*candidatesPath)
	if err != nil {
		return fail(stderr, err)
	}
	decisionsPayload, err := readJSONAny(*decisionsPath)
	if err != nil {
		return fail(stderr, err)
	}
	includeSet := map[string]bool{}
	if len(includes) == 0 {
		includeSet["keep"] = true
	} else {
		for _, inc := range includes {
			includeSet[inc] = true
		}
	}
	output, err := applyDecisions(candidates, decisionsPayload, includeSet, *allowMissing, *allowUnknown)
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

func applyDecisions(candidates map[string]any, decisionsPayload any, include map[string]bool, allowMissing, allowUnknown bool) (map[string]any, error) {
	signals := signalSlice(candidates["signals"])
	signalByID := map[string]map[string]any{}
	for _, signal := range signals {
		if id := stringValue(signal["id"]); id != "" {
			signalByID[id] = signal
		}
	}
	decisions, err := normalizeDecisions(decisionsPayload)
	if err != nil {
		return nil, err
	}
	decisionByID := map[string]map[string]any{}
	var errs []string
	for _, decision := range decisions {
		id := strings.TrimSpace(stringValue(decision["signal_id"]))
		if id == "" {
			errs = append(errs, "decision missing signal_id")
			continue
		}
		if _, ok := decisionByID[id]; ok {
			errs = append(errs, "duplicate decision for signal_id="+id)
			continue
		}
		if signalByID[id] == nil && !allowUnknown {
			errs = append(errs, "decision references unknown signal_id="+id)
			continue
		}
		normalized := normalizeDecision(decision)
		if !allowedDecisions[stringValue(normalized["decision"])] {
			errs = append(errs, fmt.Sprintf("%s: unsupported decision=%s", id, normalized["decision"]))
		}
		if !allowedReasons[stringValue(normalized["reason"])] {
			errs = append(errs, fmt.Sprintf("%s: unsupported reason=%s", id, normalized["reason"]))
		}
		decisionByID[id] = normalized
	}
	var missing []string
	for id := range signalByID {
		if decisionByID[id] == nil {
			missing = append(missing, id)
		}
	}
	sort.Strings(missing)
	if len(missing) > 0 && !allowMissing {
		sample := strings.Join(firstN(missing, 8), ", ")
		suffix := ""
		if len(missing) > 8 {
			suffix = "..."
		}
		errs = append(errs, fmt.Sprintf("missing decisions for %d signal(s): %s%s", len(missing), sample, suffix))
	}
	if len(errs) > 0 {
		return nil, errors.New(strings.Join(errs, "\n"))
	}
	var selected []map[string]any
	var rejected []map[string]any
	var missingSignals []map[string]any
	decisionCounts := map[string]int{}
	reasonCounts := map[string]int{}
	for _, d := range decisionByID {
		decisionCounts[stringValue(d["decision"])]++
		reasonCounts[stringValue(d["reason"])]++
	}
	for _, signal := range signals {
		id := stringValue(signal["id"])
		decision := decisionByID[id]
		if decision == nil {
			missingSignals = append(missingSignals, summarySignal(signal))
			continue
		}
		withDecision := cloneMap(signal)
		withDecision["coarse_relevance"] = decision
		if include[stringValue(decision["decision"])] {
			selected = append(selected, withDecision)
		} else {
			s := summarySignal(signal)
			s["decision"] = decision["decision"]
			s["reason"] = decision["reason"]
			s["rationale"] = decision["rationale"]
			rejected = append(rejected, s)
		}
	}
	included := mapKeys(include)
	sort.Strings(included)
	return map[string]any{
		"version":              1,
		"generated_at":         time.Now().UTC().Format(time.RFC3339Nano),
		"monitor":              valueOrEmptyMap(candidates["monitor"]),
		"signals":              selected,
		"coarse_relevance":     map[string]any{"input_signal_count": len(signals), "decision_count": len(decisionByID), "selected_count": len(selected), "rejected_count": len(rejected), "missing_count": len(missingSignals), "included_decisions": included, "decision_counts": sortedCountMap(decisionCounts), "reason_counts": sortedCountMap(reasonCounts), "rejected_signals": rejected, "missing_signals": missingSignals},
		"detector_diagnostics": valueOrEmptyMap(candidates["diagnostics"]),
		"source_errors":        valueOrEmptyMap(candidates["source_errors"]),
	}, nil
}

func normalizeDecisions(payload any) ([]map[string]any, error) {
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
		if arr, ok := m["decisions"].([]any); ok {
			var out []map[string]any
			for _, item := range arr {
				if d, ok := item.(map[string]any); ok {
					out = append(out, d)
				}
			}
			return out, nil
		}
	}
	return nil, errors.New("decisions JSON must be a list or an object with a decisions list")
}

func normalizeDecision(decision map[string]any) map[string]any {
	out := map[string]any{
		"signal_id":     strings.TrimSpace(stringValue(decision["signal_id"])),
		"decision":      strings.TrimSpace(stringValue(decision["decision"])),
		"reason":        strings.TrimSpace(stringValue(decision["reason"])),
		"rationale":     strings.TrimSpace(stringValue(decision["rationale"])),
		"confidence":    firstString(strings.TrimSpace(stringValue(decision["confidence"])), "medium"),
		"evidence_urls": toStringSlice(decision["evidence_urls"]),
	}
	return out
}

func summarySignal(signal map[string]any) map[string]any {
	var urls []string
	for _, item := range anySlice(signal["evidence"]) {
		if m, ok := item.(map[string]any); ok {
			if u := stringValue(m["url"]); u != "" {
				urls = append(urls, u)
			}
		}
	}
	return map[string]any{"signal_id": signal["id"], "signal_title": signal["title"], "sources": valueOrEmptyArray(signal["sources"]), "routing": valueOrEmptyMap(signal["routing"]), "evidence_urls": firstN(urls, 5)}
}
