package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSummarizeRunAcceptsFlagsAfterInput(t *testing.T) {
	repo := repoRootForTest(t)
	dir := t.TempDir()
	summary := filepath.Join(dir, "summary.json")
	runMD := filepath.Join(dir, "run.md")
	var out, err bytes.Buffer
	code := runCLI([]string{
		"summarize-run",
		filepath.Join(repo, "fixtures/golden/go-port/detector/mock-query.json"),
		"--output", summary,
		"--markdown", runMD,
	}, &out, &err)
	if code != 0 {
		t.Fatalf("summarize code=%d stderr=%s", code, err.String())
	}
	if !fileExists(summary) || !fileExists(runMD) {
		t.Fatalf("summary artifacts missing")
	}
}

func TestSummarizeRunRendersFinalReportBody(t *testing.T) {
	dir := t.TempDir()
	candidates := filepath.Join(dir, "candidates.json")
	summary := filepath.Join(dir, "summary.json")
	runMD := filepath.Join(dir, "run.md")
	if err := os.WriteFile(candidates, []byte(`{"monitor":{"generated_at":"2026-05-25T18:00:00Z"},"signals":[]}`), 0o644); err != nil {
		t.Fatalf("write candidates: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "final_report.md"), []byte("Verdict: no opportunities after freshness review.\n\n- **Link:** `https://example.com/story`\n"), 0o644); err != nil {
		t.Fatalf("write final report: %v", err)
	}
	var out, errBuf bytes.Buffer
	code := runCLI([]string{"summarize-run", candidates, "--output", summary, "--markdown", runMD}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("summarize code=%d stderr=%s", code, errBuf.String())
	}
	data, err := os.ReadFile(runMD)
	if err != nil {
		t.Fatalf("read run.md: %v", err)
	}
	body := string(data)
	if !strings.Contains(body, "## Editorial Verdict") || !strings.Contains(body, "Verdict: no opportunities after freshness review.") {
		t.Fatalf("run.md did not include final_report.md body:\n%s", body)
	}
	if !strings.Contains(body, "[example.com](https://example.com/story)") {
		t.Fatalf("run.md did not convert final report URL to clickable link:\n%s", body)
	}
}

func TestSummarizeRunRendersSkimmableCandidateLinks(t *testing.T) {
	dir := t.TempDir()
	candidates := filepath.Join(dir, "candidates.json")
	summary := filepath.Join(dir, "summary.json")
	runMD := filepath.Join(dir, "run.md")
	body := `{
  "monitor": {"generated_at": "2026-05-25T18:00:00Z", "queries": ["AI workforce planning"], "sources_used": ["news_search"], "profile": {"company": "Orgvue"}},
  "diagnostics": {"total_scored_signals": 1},
  "signals": [
    {
      "id": "story",
      "title": "Wix confirms AI layoffs",
      "routing": {"lane": "profile_relevance", "queue_priority": 62.2},
      "story_size": {"band": "high", "score": 49, "confidence": "medium"},
      "mechanical_scores": {"profile_match": 0.1},
      "evidence": [
        {"source": "news_search", "title": "Wix confirms AI layoffs", "url": "https://example.com/wix", "published_at": "2026-05-25"}
      ]
    }
  ]
}`
	if err := os.WriteFile(candidates, []byte(body), 0o644); err != nil {
		t.Fatalf("write candidates: %v", err)
	}
	var out, errBuf bytes.Buffer
	code := runCLI([]string{"summarize-run", candidates, "--output", summary, "--markdown", runMD}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("summarize code=%d stderr=%s", code, errBuf.String())
	}
	data, err := os.ReadFile(runMD)
	if err != nil {
		t.Fatalf("read run.md: %v", err)
	}
	rendered := string(data)
	for _, want := range []string{"## Top News Today", "Size: high (49, medium confidence)", "Google News", "## Scan Context"} {
		if !strings.Contains(rendered, want) {
			t.Fatalf("run.md missing %q:\n%s", want, rendered)
		}
	}
	if strings.Contains(rendered, "## Appendix: Provenance") {
		t.Fatalf("run.md should not include provenance dump:\n%s", rendered)
	}
}

// Feeding render-run the RAW candidates.json must still produce a safe brief:
// the renderer drops hard-safety-flagged signals from the human scan even when
// no upstream filter removed them, and discloses that it did so.
func TestRenderRunGatesSafetyFlaggedSignals(t *testing.T) {
	dir := t.TempDir()
	candidates := filepath.Join(dir, "candidates.json")
	body := `{
  "monitor": {"generated_at": "2026-05-25T18:00:00Z", "profile": {"company": "Clearnym"}},
  "signals": [
    {"id": "tragedy", "title": "Apartment fire kills 3 in Dallas",
     "routing": {"lane": "profile_relevance_weak", "queue_priority": 66.0},
     "features": {"safety_flags": [{"type": "hard_safety_term", "term": "kills"}]},
     "evidence": [{"source": "news_search", "title": "Apartment fire kills 3", "url": "https://example.com/fire"}]},
    {"id": "ok", "title": "Connecticut bans location data sales",
     "routing": {"lane": "profile_relevance", "queue_priority": 42.4},
     "features": {"safety_flags": []},
     "evidence": [{"source": "news_search", "title": "CT bans data sales", "url": "https://example.com/ct"}]}
  ]
}`
	if err := os.WriteFile(candidates, []byte(body), 0o644); err != nil {
		t.Fatalf("write candidates: %v", err)
	}
	rendered := runRenderForTest(t, candidates)
	if strings.Contains(rendered, "Apartment fire kills 3") {
		t.Fatalf("safety-flagged signal leaked into the brief:\n%s", rendered)
	}
	if !strings.Contains(rendered, "Connecticut bans location data sales") {
		t.Fatalf("clean signal missing from the brief:\n%s", rendered)
	}
	if !strings.Contains(rendered, "brand-safety-flagged") {
		t.Fatalf("brief did not disclose the withheld safety-flagged signal:\n%s", rendered)
	}
}

// A signal the coarse pass rejected must never render, even when the caller
// hands render-run the pre-filter candidates.json. The renderer reads the
// coarse decisions file from the run dir as the gate.
func TestRenderRunGatesCoarseRejectedSignals(t *testing.T) {
	dir := t.TempDir()
	candidates := filepath.Join(dir, "candidates.json")
	body := `{
  "monitor": {"generated_at": "2026-05-25T18:00:00Z", "profile": {"company": "Clearnym"}},
  "signals": [
    {"id": "junk", "title": "Local police standoff downtown",
     "routing": {"lane": "profile_relevance_weak", "queue_priority": 50.0},
     "evidence": [{"source": "news_search", "title": "standoff", "url": "https://example.com/junk"}]},
    {"id": "keep", "title": "Connecticut bans location data sales",
     "routing": {"lane": "profile_relevance", "queue_priority": 42.4},
     "evidence": [{"source": "news_search", "title": "CT bans data sales", "url": "https://example.com/ct"}]}
  ]
}`
	if err := os.WriteFile(candidates, []byte(body), 0o644); err != nil {
		t.Fatalf("write candidates: %v", err)
	}
	decisions := `{"decisions": [
		{"signal_id": "junk", "decision": "reject", "reason": "keyword_collision"},
		{"signal_id": "keep", "decision": "keep", "reason": "relevant_news"}
	]}`
	if err := os.WriteFile(filepath.Join(dir, "coarse_relevance_decisions.json"), []byte(decisions), 0o644); err != nil {
		t.Fatalf("write decisions: %v", err)
	}
	rendered := runRenderForTest(t, candidates)
	if strings.Contains(rendered, "Local police standoff downtown") {
		t.Fatalf("coarse-rejected signal leaked into the brief:\n%s", rendered)
	}
	if !strings.Contains(rendered, "Connecticut bans location data sales") {
		t.Fatalf("kept signal missing from the brief:\n%s", rendered)
	}
	if !strings.Contains(rendered, "coarse-rejected") {
		t.Fatalf("brief did not disclose the withheld coarse-rejected signal:\n%s", rendered)
	}
}

func TestRenderRunNameAndDeprecatedAlias(t *testing.T) {
	dir := t.TempDir()
	candidates := filepath.Join(dir, "candidates.json")
	if err := os.WriteFile(candidates, []byte(`{"monitor":{"generated_at":"2026-05-25T18:00:00Z"},"signals":[]}`), 0o644); err != nil {
		t.Fatalf("write candidates: %v", err)
	}
	for _, name := range []string{"render-run", "summarize-run"} {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{name, candidates, "--output", filepath.Join(dir, name+".json"), "--markdown", filepath.Join(dir, name+".md")}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("%s code=%d stderr=%s", name, code, errBuf.String())
		}
	}
}

func runRenderForTest(t *testing.T, candidates string) string {
	t.Helper()
	dir := filepath.Dir(candidates)
	summary := filepath.Join(dir, "summary.json")
	runMD := filepath.Join(dir, "run.md")
	var out, errBuf bytes.Buffer
	code := runCLI([]string{"render-run", candidates, "--output", summary, "--markdown", runMD}, &out, &errBuf)
	if code != 0 {
		t.Fatalf("render-run code=%d stderr=%s", code, errBuf.String())
	}
	data, err := os.ReadFile(runMD)
	if err != nil {
		t.Fatalf("read run.md: %v", err)
	}
	return string(data)
}
