package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRunSummaryAcceptsFlagsAfterInput(t *testing.T) {
	repo := repoRootForTest(t)
	dir := t.TempDir()
	summary := filepath.Join(dir, "summary.json")
	var out, err bytes.Buffer
	code := runCLI([]string{
		"run-summary",
		filepath.Join(repo, "fixtures/golden/go-port/detector/mock-query.json"),
		"--output", summary,
	}, &out, &err)
	if code != 0 {
		t.Fatalf("run-summary code=%d stderr=%s", code, err.String())
	}
	if !fileExists(summary) {
		t.Fatalf("summary artifact missing")
	}
}

func TestRunSummaryReportsMarkdownArtifactMetadataOnly(t *testing.T) {
	dir := t.TempDir()
	candidates := filepath.Join(dir, "candidates.json")
	summary := filepath.Join(dir, "summary.json")
	if err := os.WriteFile(candidates, []byte(`{"monitor":{"generated_at":"2026-05-25T18:00:00Z"},"signals":[]}`), 0o644); err != nil {
		t.Fatalf("write candidates: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "final_report.md"), []byte("Verdict: no opportunities.\n"), 0o644); err != nil {
		t.Fatalf("write final report: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "run.md"), []byte("# Run Report\n"), 0o644); err != nil {
		t.Fatalf("write run report: %v", err)
	}
	var out, errBuf bytes.Buffer
	code := runCLI([]string{"run-summary", candidates, "--output", summary}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("run-summary code=%d stderr=%s", code, errBuf.String())
	}
	payload := readJSONForTest(t, summary)
	artifacts := valueOrEmptyMap(payload["artifacts"])
	runReport := valueOrEmptyMap(artifacts["run_report"])
	if !truthy(runReport["exists"], false) {
		t.Fatalf("summary reported run.md missing: %#v", runReport)
	}
	finalReport := valueOrEmptyMap(payload["final_report_file"])
	if !truthy(finalReport["exists"], false) {
		t.Fatalf("summary reported final_report.md missing: %#v", finalReport)
	}
	if _, ok := finalReport["content"]; ok {
		t.Fatalf("summary should not embed report prose: %#v", finalReport)
	}
}

func TestRunSummaryCountsDeterministicExclusionsWithoutFilteringSignals(t *testing.T) {
	dir := t.TempDir()
	candidates := filepath.Join(dir, "candidates.json")
	body := `{
  "monitor": {"generated_at": "2026-05-25T18:00:00Z", "profile": {"company": "Clearnym"}},
  "signals": [
    {"id": "tragedy", "title": "Apartment fire kills 3 in Dallas",
     "routing": {"lane": "profile_relevance_weak", "queue_priority": 66.0},
     "features": {"safety_flags": [{"type": "hard_safety_term", "term": "kills"}]},
     "evidence": [{"source": "news_search", "title": "Apartment fire kills 3", "url": "https://example.com/fire"}]},
    {"id": "junk", "title": "Local police standoff downtown",
     "routing": {"lane": "profile_relevance_weak", "queue_priority": 50.0},
     "evidence": [{"source": "news_search", "title": "standoff", "url": "https://example.com/junk"}]}
  ]
}`
	if err := os.WriteFile(candidates, []byte(body), 0o644); err != nil {
		t.Fatalf("write candidates: %v", err)
	}
	decisions := `{"decisions": [
		{"signal_id": "tragedy", "decision": "keep", "reason": "relevant_news"},
		{"signal_id": "junk", "decision": "reject", "reason": "new_prompt_owned_reason"}
	]}`
	if err := os.WriteFile(filepath.Join(dir, "coarse_relevance_decisions.json"), []byte(decisions), 0o644); err != nil {
		t.Fatalf("write decisions: %v", err)
	}
	payload := runSummaryForTest(t, candidates)
	signals := mapSlice(payload["top_signals"])
	if len(signals) != 2 {
		t.Fatalf("top_signals=%d, want unfiltered input signals", len(signals))
	}
	exclusions := valueOrEmptyMap(payload["deterministic_exclusions"])
	if exclusions["coarse_rejected_in_input"] != float64(1) || exclusions["safety_flagged_in_input"] != float64(1) {
		t.Fatalf("exclusions=%#v, want one coarse rejection and one safety flag", exclusions)
	}
}

func TestReportRenderCommandsRemoved(t *testing.T) {
	dir := t.TempDir()
	candidates := filepath.Join(dir, "candidates.json")
	if err := os.WriteFile(candidates, []byte(`{"monitor":{"generated_at":"2026-05-25T18:00:00Z"},"signals":[]}`), 0o644); err != nil {
		t.Fatalf("write candidates: %v", err)
	}
	for _, name := range []string{"render-run", "summarize-run"} {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{name, candidates}, &out, &errBuf)
		if code == 0 {
			t.Fatalf("%s unexpectedly succeeded", name)
		}
	}
}

func runSummaryForTest(t *testing.T, candidates string) map[string]any {
	t.Helper()
	var out, errBuf bytes.Buffer
	code := runCLI([]string{"run-summary", candidates}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("run-summary code=%d stderr=%s", code, errBuf.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
		t.Fatalf("invalid summary JSON: %s", out.String())
	}
	return payload
}

func readJSONForTest(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
	return payload
}
