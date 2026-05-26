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
	if err := os.WriteFile(filepath.Join(dir, "final_report.md"), []byte("Verdict: no opportunities after freshness review.\n"), 0o644); err != nil {
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
}
