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
