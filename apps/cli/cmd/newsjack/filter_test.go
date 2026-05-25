package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
)

func TestFilterApplyFixture(t *testing.T) {
	repo := repoRootForTest(t)
	var out, err bytes.Buffer
	code := runCLI([]string{
		"filter-apply",
		"--candidates", filepath.Join(repo, "fixtures/golden/go-port/detector/mock-query.json"),
		"--decisions", filepath.Join(repo, "fixtures/golden/go-port/filter/decisions.json"),
		"--include", "keep",
	}, &out, &err)
	if code != 0 {
		t.Fatalf("filter code=%d stderr=%s", code, err.String())
	}
	var payload map[string]any
	if json.Unmarshal(out.Bytes(), &payload) != nil {
		t.Fatalf("invalid JSON: %s", out.String())
	}
	signals := signalSlice(payload["signals"])
	if len(signals) != 2 {
		t.Fatalf("selected signals=%d, want 2", len(signals))
	}
	cheap := valueOrEmptyMap(payload["cheap_filter"])
	if cheap["selected_count"].(float64) != 2 || cheap["rejected_count"].(float64) != 0 {
		t.Fatalf("unexpected cheap summary: %#v", cheap)
	}
}
