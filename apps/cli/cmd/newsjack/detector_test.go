package main

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"
)

func TestDetectorMockRunAcceptsFlagsAfterQuery(t *testing.T) {
	repo := repoRootForTest(t)
	withTempEnv(t, map[string]string{
		"HOME":          t.TempDir(),
		"NEWSJACK_ROOT": repo,
	}, func() {
		var out, err bytes.Buffer
		code := runCLI([]string{"detector", "run", "AI customer support", "--mock", "--include-all-scored", "--emit", "json"}, &out, &err)
		if code != 0 {
			t.Fatalf("detector code=%d stderr=%s", code, err.String())
		}
		var payload map[string]any
		if json.Unmarshal(out.Bytes(), &payload) != nil {
			t.Fatalf("invalid JSON: %s", out.String())
		}
		monitor := valueOrEmptyMap(payload["monitor"])
		if monitor["mock"] != true {
			t.Fatalf("mock=false payload=%s", out.String())
		}
		signals := signalSlice(payload["signals"])
		if len(signals) != 2 {
			t.Fatalf("signals=%d, want 2", len(signals))
		}
		if signals[0]["id"] != "e32ebc6ac34ee9d2" || signals[1]["id"] != "578832fabe7e6a64" {
			t.Fatalf("unexpected signal ids: %#v %#v", signals[0]["id"], signals[1]["id"])
		}
		debug := valueOrEmptyMap(payload["debug"])
		if dropped := anySlice(debug["dropped_signal_ids"]); len(dropped) != 0 {
			t.Fatalf("dropped=%v, want empty", dropped)
		}
	})
}

func TestDemotedDetectorLanesStayBelowDefaultQueueFloor(t *testing.T) {
	now := time.Date(2026, 5, 25, 13, 0, 0, 0, time.UTC)
	profile := monitorProfile{
		Company:     "Clearnym",
		Topics:      []string{"data broker removal"},
		Standing:    []string{"identity theft prevention"},
		ProofAssets: []string{"privacy operations"},
	}
	opts := detectorOptions{
		LookbackDays:                    1,
		XNewsMinProfileMatch:            0.05,
		XPostsMinProfileMatch:           0.08,
		ProfileRelevanceMinProfileMatch: 0.05,
		MajorNewsMinProfileMatch:        0.05,
		XTrendsMinProfileMatch:          0.05,
	}

	tests := []struct {
		name       string
		source     string
		wantLane   string
		sourceName string
	}{
		{name: "profile query", source: "news_search", wantLane: "profile_relevance_weak", sourceName: "Reuters"},
		{name: "x news", source: "x_news", wantLane: "x_news_unmatched", sourceName: "X News"},
		{name: "x posts", source: "x", wantLane: "x_posts_weak", sourceName: "X"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			signal := scoreSignal(signalCluster{Evidence: []evidenceItem{{
				Source:      tt.source,
				Title:       "Coffee shops face rising import prices",
				URL:         "https://example.com/coffee-" + tt.source,
				Excerpt:     "Independent cafes are tracking higher bean costs.",
				Container:   tt.sourceName,
				PublishedAt: now.Add(-1 * time.Hour).Format(time.RFC3339),
				Engagement:  map[string]any{},
				Metadata:    map[string]any{},
			}}}, profile, map[string]map[string]any{}, now, opts)
			if lane := signalLaneValue(signal); lane != tt.wantLane {
				t.Fatalf("lane=%s, want %s", lane, tt.wantLane)
			}
			if priority := queuePriority(signal); priority >= defaultMinQueuePriority {
				t.Fatalf("queue priority=%v, want below default floor %v", priority, defaultMinQueuePriority)
			}
			if passesSelectionFloor(signal, defaultMinQueuePriority, defaultMinMajorNews) {
				t.Fatalf("demoted lane passed default selection floor: %#v", signal["routing"])
			}
		})
	}
}

func TestMajorNewsFallbackDoesNotSelectUnmatchedMajorNews(t *testing.T) {
	unmatched := map[string]any{
		"routing":           map[string]any{"lane": "major_news_unmatched", "queue_priority": 39.9},
		"mechanical_scores": map[string]any{"major_news": 0.8},
	}
	if passesSelectionFloor(unmatched, defaultMinQueuePriority, defaultMinMajorNews) {
		t.Fatal("major_news_unmatched passed the broad major-news fallback")
	}

	matched := map[string]any{
		"routing":           map[string]any{"lane": "major_news", "queue_priority": 39.9},
		"mechanical_scores": map[string]any{"major_news": 0.8},
	}
	if !passesSelectionFloor(matched, defaultMinQueuePriority, defaultMinMajorNews) {
		t.Fatal("matched major_news did not pass the broad major-news fallback")
	}
}

func TestProfileMatchesRequireTokenBoundaryForSingleTerms(t *testing.T) {
	profile := monitorProfile{Competitors: []string{"Aura"}}
	if matches := profileMatches(profile, "Une dermatologue poursuit une action en justice."); len(matches) != 0 {
		t.Fatalf("matches=%v, want no substring-only match", matches)
	}
	if matches := profileMatches(profile, "Aura announced a new identity protection feature."); len(matches) != 1 || matches[0] != "Aura" {
		t.Fatalf("matches=%v, want exact Aura token match", matches)
	}
}
