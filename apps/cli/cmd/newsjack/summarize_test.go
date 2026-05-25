package main

import (
	"bytes"
	"path/filepath"
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
