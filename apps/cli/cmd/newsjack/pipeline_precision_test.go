package main

import (
	"testing"
	"time"
)

func TestDemoteUnmatchedXLowersBelowFloor(t *testing.T) {
	now := time.Date(2026, 5, 25, 13, 0, 0, 0, time.UTC)
	profile := monitorProfile{Company: "Clearnym", Topics: []string{"data broker removal"}}
	item := evidenceItem{
		Source:      "x_news",
		Title:       "Coffee shops face rising import prices",
		URL:         "https://example.com/coffee-xnews",
		Excerpt:     "Independent cafes are tracking higher bean costs.",
		Container:   "X News",
		PublishedAt: now.Add(-1 * time.Hour).Format(time.RFC3339),
		Metadata:    map[string]any{},
	}
	base := detectorOptions{LookbackDays: 1, XNewsMinProfileMatch: 0.05, XTrendsMinProfileMatch: 0.05, XPostsMinProfileMatch: 0.08}

	def := scoreSignal(signalCluster{Evidence: []evidenceItem{item}}, profile, map[string]map[string]any{}, now, base)
	if signalLaneValue(def) != "x_news_unmatched" {
		t.Fatalf("lane=%s, want x_news_unmatched", signalLaneValue(def))
	}
	if queuePriority(def) < defaultMinQueuePriority {
		t.Fatalf("default x_news_unmatched=%v, want surfaced (>= %v)", queuePriority(def), defaultMinQueuePriority)
	}

	demoteOpts := base
	demoteOpts.DemoteUnmatchedX = true
	dem := scoreSignal(signalCluster{Evidence: []evidenceItem{item}}, profile, map[string]map[string]any{}, now, demoteOpts)
	if queuePriority(dem) >= defaultMinQueuePriority {
		t.Fatalf("demoted x_news_unmatched=%v, want below default floor %v", queuePriority(dem), defaultMinQueuePriority)
	}
}

func TestOriginBoundaryStatus(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{"generated_at": "2026-05-27T19:00:00Z"},
		"signals": []any{map[string]any{"id": "s1", "title": "Boundary story",
			"evidence": []any{map[string]any{"source": "news_search", "url": "https://a.com/x"}}}},
	}
	origins := map[string]any{"findings": []any{map[string]any{
		"signal_id": "s1", "first_public_at": "2026-05-26", // date-only on the cutoff day
	}}}
	payload, err := applyOriginFindings(candidates, origins, originApplyOptions{
		RunTime: mustParseTestTime(t, "2026-05-27T19:00:00Z"), WindowHours: 24})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	rejected, _ := valueOrEmptyMap(payload["freshness_gate"])["rejected_signals"].([]map[string]any)
	if len(rejected) != 1 {
		t.Fatalf("rejected=%d, want 1", len(rejected))
	}
	if got := stringValue(valueOrEmptyMap(rejected[0]["freshness_gate"])["computed_status"]); got != "unverified_boundary" {
		t.Fatalf("computed_status=%s, want unverified_boundary", got)
	}
}

func TestOriginNoTimestampStatus(t *testing.T) {
	candidates := map[string]any{
		"monitor": map[string]any{"generated_at": "2026-05-27T19:00:00Z"},
		"signals": []any{map[string]any{"id": "s1", "title": "No timestamp story",
			"evidence": []any{map[string]any{"source": "news_search", "url": "https://a.com/x"}}}},
	}
	origins := map[string]any{"findings": []any{map[string]any{
		"signal_id": "s1", "first_public_at": nil, "same_story_assessment": "unclear",
	}}}
	payload, err := applyOriginFindings(candidates, origins, originApplyOptions{
		RunTime: mustParseTestTime(t, "2026-05-27T19:00:00Z"), WindowHours: 24})
	if err != nil {
		t.Fatalf("err=%v", err)
	}
	rejected, _ := valueOrEmptyMap(payload["freshness_gate"])["rejected_signals"].([]map[string]any)
	if len(rejected) != 1 {
		t.Fatalf("rejected=%d, want 1", len(rejected))
	}
	if got := stringValue(valueOrEmptyMap(rejected[0]["freshness_gate"])["computed_status"]); got != "unverified_no_timestamp" {
		t.Fatalf("computed_status=%s, want unverified_no_timestamp", got)
	}
}
