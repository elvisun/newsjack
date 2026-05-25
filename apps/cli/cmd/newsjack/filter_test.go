package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
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
	coarse := valueOrEmptyMap(payload["coarse_relevance"])
	if coarse["selected_count"].(float64) != 2 || coarse["rejected_count"].(float64) != 0 {
		t.Fatalf("unexpected coarse summary: %#v", coarse)
	}
}

func TestFilterApplyIgnoresOriginFields(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{},
		"signals": []any{
			map[string]any{"id": "fresh", "title": "Fresh story", "evidence": []any{map[string]any{"url": "https://example.com/fresh"}}},
		},
	}
	decisions := map[string]any{
		"decisions": []any{
			map[string]any{
				"signal_id": "fresh",
				"decision":  "keep",
				"reason":    "relevant_news",
				"first_publication": map[string]any{
					"status":                          "fresh",
					"surfaced_article_published_at":   "2026-05-25T12:10:00Z",
					"first_public_at":                 "2026-05-25T12:00:00Z",
					"original_url":                    "https://example.com/original",
					"canonical_coverage_url":          "https://example.com/major-coverage",
					"canonical_coverage_source":       "Example Major",
					"canonical_coverage_published_at": "2026-05-25T12:05:00Z",
					"confidence":                      "high",
					"rationale":                       "Original source is within the cron window.",
					"evidence_urls":                   []any{"https://example.com/original", "https://example.com/major-coverage"},
					"same_story_basis":                "same official action and same named actors",
				},
			},
		},
	}
	payload, err := applyDecisions(candidates, decisions, map[string]bool{"keep": true}, false, false)
	if err != nil {
		t.Fatalf("applyDecisions error=%v", err)
	}
	signals, _ := payload["signals"].([]map[string]any)
	if len(signals) != 1 {
		t.Fatalf("selected signals=%d, want 1", len(signals))
	}
	coarse := valueOrEmptyMap(signals[0]["coarse_relevance"])
	if _, ok := coarse["first_publication"]; ok {
		t.Fatalf("coarse relevance preserved story-origin data: %#v", coarse)
	}
}

func TestOriginApplyComputesStaleDespiteWorkerFreshStatus(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{"generated_at": "2026-05-25T18:00:00Z"},
		"signals": []any{
			map[string]any{"id": "gm", "title": "General Motors settles lawsuit over selling customer driving data"},
		},
	}
	origins := map[string]any{
		"findings": []any{
			map[string]any{
				"signal_id":              "gm",
				"status":                 "fresh",
				"first_public_at":        "2026-05-08",
				"original_url":           "https://example.com/gm-original",
				"canonical_coverage_url": "https://example.com/gm-major",
				"same_story_basis":       "Same GM settlement and same lawsuit.",
			},
		},
	}
	payload, err := applyOriginFindings(candidates, origins, originApplyOptions{
		RunTime:     mustParseTestTime(t, "2026-05-25T18:00:00Z"),
		WindowHours: 24,
	})
	if err != nil {
		t.Fatalf("applyOriginFindings error=%v", err)
	}
	if got := len(signalSlice(payload["signals"])); got != 0 {
		t.Fatalf("selected signals=%d, want 0", got)
	}
	rejected, _ := valueOrEmptyMap(payload["freshness_gate"])["rejected_signals"].([]map[string]any)
	if len(rejected) != 1 {
		t.Fatalf("rejected=%d, want 1", len(rejected))
	}
	gate := valueOrEmptyMap(rejected[0]["freshness_gate"])
	if gate["computed_status"] != "stale" {
		t.Fatalf("computed_status=%v, want stale; gate=%#v", gate["computed_status"], gate)
	}
	if gate["worker_status"] != "fresh" {
		t.Fatalf("worker_status=%v, want fresh", gate["worker_status"])
	}
}

func TestOriginApplySelectsFreshNewDevelopment(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{"generated_at": "2026-05-25T18:00:00Z"},
		"signals": []any{
			map[string]any{"id": "case", "title": "Regulator files new complaint"},
		},
	}
	origins := map[string]any{
		"findings": []any{
			map[string]any{
				"signal_id":          "case",
				"first_public_at":    "2026-05-08",
				"new_development_at": "2026-05-25T12:30:00Z",
				"new_development":    "New settlement was filed.",
			},
		},
	}
	payload, err := applyOriginFindings(candidates, origins, originApplyOptions{
		RunTime:     mustParseTestTime(t, "2026-05-25T18:00:00Z"),
		WindowHours: 24,
	})
	if err != nil {
		t.Fatalf("applyOriginFindings error=%v", err)
	}
	signals := signalSlice(payload["signals"])
	if len(signals) != 1 {
		t.Fatalf("selected signals=%d, want 1; freshness_gate=%#v", len(signals), payload["freshness_gate"])
	}
	gate := valueOrEmptyMap(signals[0]["freshness_gate"])
	if gate["computed_status"] != "fresh_new_development" {
		t.Fatalf("computed_status=%v, want fresh_new_development; gate=%#v", gate["computed_status"], gate)
	}
}

func mustParseTestTime(t *testing.T, raw string) time.Time {
	t.Helper()
	parsed, ok := parseTime(raw)
	if !ok {
		t.Fatalf("could not parse %s", raw)
	}
	return parsed
}
