package main

import "testing"

func sig(id, title, url string, queue float64, band, decay string) map[string]any {
	return map[string]any{
		"id":    id,
		"title": title,
		"evidence": []any{
			map[string]any{"source": "news_search", "title": title, "url": url, "excerpt": title},
		},
		"routing":    map[string]any{"queue_priority": queue},
		"story_size": map[string]any{"band": band, "score": 50.0},
		"features":   map[string]any{"decay_bucket": decay},
	}
}

func TestClusterCollapsesSameStoryAndKeepsDistinct(t *testing.T) {
	candidates := map[string]any{
		"signals": []any{
			sig("a", "Nvidia unveils RTX Spark superchip for Windows PCs", "https://nyt.com/a", 60, "high", "24hr"),
			sig("b", "Nvidia unveils RTX Spark superchip for Windows PCs today", "https://bbc.com/b", 55, "high", "24hr"),
			sig("c", "Connecticut enacts new data privacy delete act", "https://ct.gov/c", 50, "moderate", "24hr"),
		},
	}
	out := clusterCandidates(candidates, clusterOptions{TitleOverlap: 0.6})
	reps := signalSlice(out["signals"])
	if len(reps) != 2 {
		t.Fatalf("representatives=%d, want 2 (one NVIDIA cluster + the distinct CT story); reps=%#v", len(reps), reps)
	}
	dups, _ := out["clustered_duplicates"].([]map[string]any)
	if len(dups) != 1 {
		t.Fatalf("duplicates=%d, want 1", len(dups))
	}
	// The NVIDIA representative must be the higher-queue signal "a".
	for _, r := range reps {
		cl := valueOrEmptyMap(r["cluster"])
		if numberOr(cl["cluster_size"]) == 2 && stringValue(r["id"]) != "a" {
			t.Fatalf("NVIDIA representative=%s, want a (highest queue priority)", stringValue(r["id"]))
		}
	}
}

func TestClusterSharedURLForcesSameCluster(t *testing.T) {
	shared := "https://wire.example.com/story"
	candidates := map[string]any{
		"signals": []any{
			sig("a", "Totally different headline one", shared, 60, "moderate", "24hr"),
			sig("b", "Completely unrelated wording two", shared, 40, "moderate", "24hr"),
		},
	}
	out := clusterCandidates(candidates, clusterOptions{TitleOverlap: 0.9})
	if reps := signalSlice(out["signals"]); len(reps) != 1 {
		t.Fatalf("representatives=%d, want 1 (shared evidence URL collapses them)", len(reps))
	}
}

func TestClusterStalePreGateDropsOldLowValueOnly(t *testing.T) {
	candidates := map[string]any{
		"signals": []any{
			sig("old_small", "Old minor trend piece", "https://x.com/1", 45, "moderate", "week"),
			sig("old_big", "Old but major story", "https://x.com/2", 45, "major", "month"),
			sig("fresh_small", "Fresh minor item", "https://x.com/3", 45, "low", "24hr"),
		},
	}
	out := clusterCandidates(candidates, clusterOptions{TitleOverlap: 0.6, DropStale: true, WindowHours: 24, StaleMaxBand: "moderate"})
	reps := signalSlice(out["signals"])
	pre, _ := out["pre_gated_stale"].([]map[string]any)
	if len(pre) != 1 || stringValue(pre[0]["signal_id"]) != "old_small" {
		t.Fatalf("pre_gated=%#v, want exactly old_small", pre)
	}
	repIDs := map[string]bool{}
	for _, r := range reps {
		repIDs[stringValue(r["id"])] = true
	}
	if !repIDs["old_big"] {
		t.Fatalf("old_big (major story) must be researched, not pre-gated; reps=%v", repIDs)
	}
	if !repIDs["fresh_small"] {
		t.Fatalf("fresh_small must be researched (recent decay); reps=%v", repIDs)
	}
}

func numberOr(v any) float64 {
	f, _ := numberValue(v)
	return f
}
