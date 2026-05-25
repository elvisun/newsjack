package main

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"strings"
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
	coarse := valueOrEmptyMap(payload["coarse_filter"])
	if coarse["selected_count"].(float64) != 2 || coarse["rejected_count"].(float64) != 0 {
		t.Fatalf("unexpected coarse summary: %#v", coarse)
	}
}

func TestFilterApplyPreservesFirstPublicationJudgment(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{},
		"signals": []any{
			map[string]any{"id": "fresh", "title": "Fresh story", "evidence": []any{map[string]any{"url": "https://example.com/fresh"}}},
			map[string]any{"id": "unknown", "title": "Unverified story", "evidence": []any{map[string]any{"url": "https://example.com/unknown"}}},
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
			map[string]any{
				"signal_id": "unknown",
				"decision":  "reject",
				"reason":    "freshness_unverified",
				"first_publication": map[string]any{
					"status":     "freshness_unverified",
					"confidence": "low",
					"rationale":  "No original source could be verified.",
				},
			},
		},
	}
	payload, err := applyDecisions(candidates, decisions, map[string]bool{"keep": true}, false, false, true)
	if err != nil {
		t.Fatalf("applyDecisions error=%v", err)
	}
	signals, _ := payload["signals"].([]map[string]any)
	if len(signals) != 1 {
		t.Fatalf("selected signals=%d, want 1", len(signals))
	}
	coarse := valueOrEmptyMap(signals[0]["coarse_filter"])
	firstPublication := valueOrEmptyMap(coarse["first_publication"])
	if firstPublication["status"] != "fresh" || firstPublication["canonical_coverage_url"] != "https://example.com/major-coverage" {
		t.Fatalf("first_publication not preserved: %#v", firstPublication)
	}
	rejected, _ := valueOrEmptyMap(payload["coarse_filter"])["rejected_signals"].([]map[string]any)
	if len(rejected) != 1 {
		t.Fatalf("rejected=%d, want 1", len(rejected))
	}
	rejectedFP := valueOrEmptyMap(rejected[0]["first_publication"])
	if rejectedFP["status"] != "freshness_unverified" {
		t.Fatalf("rejected first_publication not preserved: %#v", rejectedFP)
	}
}

func TestFilterApplyRequiresFreshFirstPublicationForIncludedSignals(t *testing.T) {
	candidates := map[string]any{
		"signals": []any{
			map[string]any{"id": "stale", "title": "Syndicated story"},
		},
	}
	decisions := map[string]any{
		"decisions": []any{
			map[string]any{
				"signal_id": "stale",
				"decision":  "keep",
				"reason":    "relevant_news",
				"first_publication": map[string]any{
					"status":          "stale",
					"first_public_at": "2026-05-04",
				},
			},
		},
	}
	_, err := applyDecisions(candidates, decisions, map[string]bool{"keep": true}, false, false, true)
	if err == nil || !strings.Contains(err.Error(), "requires first_publication.status=fresh or fresh_new_development") {
		t.Fatalf("expected freshness gate error, got %v", err)
	}
}
