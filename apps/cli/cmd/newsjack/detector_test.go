package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
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

func TestStorySizeUsesTrafficAuthorityAndCoverageSpread(t *testing.T) {
	now := time.Date(2026, 5, 25, 13, 0, 0, 0, time.UTC)
	opts := detectorOptions{LookbackDays: 1}
	single := scoreSignal(signalCluster{Evidence: []evidenceItem{{
		Source:      "news_search",
		Title:       "Major outlet covers AI story",
		URL:         "https://wsj.com/articles/ai-story",
		Container:   "WSJ",
		PublishedAt: now.Format(time.RFC3339),
		Metadata: map[string]any{
			"domain_authority":                  95,
			"estimated_monthly_organic_traffic": 100000000,
		},
	}}}, monitorProfile{}, map[string]map[string]any{}, now, opts)
	spread := scoreSignal(signalCluster{Evidence: []evidenceItem{
		{
			Source:      "news_search",
			Title:       "Major outlet covers AI story",
			URL:         "https://wsj.com/articles/ai-story",
			Container:   "WSJ",
			PublishedAt: now.Format(time.RFC3339),
			Metadata: map[string]any{
				"domain_authority":                  95,
				"estimated_monthly_organic_traffic": 100000000,
			},
		},
		{
			Source:      "news_search",
			Title:       "Forbes covers AI story",
			URL:         "https://forbes.com/sites/example/ai-story",
			Container:   "Forbes",
			PublishedAt: now.Format(time.RFC3339),
			Metadata: map[string]any{
				"domain_authority":                  90,
				"estimated_monthly_organic_traffic": 70000000,
			},
		},
		{
			Source:      "news_search",
			Title:       "USA Today covers AI story",
			URL:         "https://usatoday.com/story/news/ai-story",
			Container:   "USA Today",
			PublishedAt: now.Format(time.RFC3339),
			Metadata: map[string]any{
				"domain_authority":                  88,
				"estimated_monthly_organic_traffic": 60000000,
			},
		},
	}}, monitorProfile{}, map[string]map[string]any{}, now, opts)

	singleSize := valueOrEmptyMap(single["story_size"])
	spreadSize := valueOrEmptyMap(spread["story_size"])
	if singleSize["band"] != "high" {
		t.Fatalf("single band=%v, want high; story_size=%#v", singleSize["band"], singleSize)
	}
	if spreadSize["band"] != "major" {
		t.Fatalf("spread band=%v, want major; story_size=%#v", spreadSize["band"], spreadSize)
	}
	if floatValue(spreadSize["score"]) <= floatValue(singleSize["score"]) {
		t.Fatalf("spread score=%v, want above single=%v", spreadSize["score"], singleSize["score"])
	}
	if valueOrEmptyMap(spread["mechanical_scores"])["story_size"] == nil {
		t.Fatalf("mechanical story_size missing: %#v", spread["mechanical_scores"])
	}
}

func TestParseNewsResponsePreservesPublicationMetadata(t *testing.T) {
	items := parseNewsResponse(map[string]any{
		"results": []any{
			map[string]any{
				"title":  "Example story",
				"url":    "https://example.com/story",
				"source": "Example",
				"date":   "2026-05-26",
				"metadata": map[string]any{
					"publication_type":                  "editorial",
					"domain_authority":                  73,
					"estimated_monthly_organic_traffic": 460000,
				},
			},
		},
	})
	if len(items) != 1 {
		t.Fatalf("items=%d, want 1", len(items))
	}
	metadata := valueOrEmptyMap(items[0]["metadata"])
	if metadata["domain_authority"] != 73 || metadata["estimated_monthly_organic_traffic"] != 460000 {
		t.Fatalf("publication metadata not preserved: %#v", metadata)
	}
	if metadata["raw_source"] != "Example" {
		t.Fatalf("raw_source=%v, want Example", metadata["raw_source"])
	}
}

func TestRSSFeedCatalogAuditMetadata(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRootForTest(t), "skills", "newsjack-detector", "references", "rss-feeds.json"))
	if err != nil {
		t.Fatal(err)
	}
	type catalogGate struct {
		RequiredBeats []string `json:"required_beats"`
	}
	type catalogFeed struct {
		ID      string       `json:"id"`
		Name    string       `json:"name"`
		URL     string       `json:"url"`
		Tier    string       `json:"tier"`
		Beats   []string     `json:"beats"`
		UseWhen string       `json:"use_when"`
		Notes   string       `json:"notes"`
		Gate    *catalogGate `json:"gate"`
	}
	var catalog struct {
		Version   int `json:"version"`
		Changelog []struct {
			Version int    `json:"version"`
			Date    string `json:"date"`
			Summary string `json:"summary"`
		} `json:"changelog"`
		DefaultFeeds []string      `json:"default_feeds"`
		Feeds        []catalogFeed `json:"feeds"`
	}
	if err := json.Unmarshal(data, &catalog); err != nil {
		t.Fatal(err)
	}
	if catalog.Version != 2 {
		t.Fatalf("version=%d, want 2", catalog.Version)
	}
	if len(catalog.Changelog) == 0 || catalog.Changelog[0].Version != 2 || catalog.Changelog[0].Date != "2026-05-26" {
		t.Fatalf("missing v2 changelog: %#v", catalog.Changelog)
	}
	if len(catalog.DefaultFeeds) != 2 || catalog.DefaultFeeds[0] != "techmeme" || catalog.DefaultFeeds[1] != "google-news-technology" {
		t.Fatalf("default_feeds=%v, want techmeme/google-news-technology", catalog.DefaultFeeds)
	}

	feeds := map[string]catalogFeed{}
	allowedTiers := stringSet([]string{"primary", "demoted", "gated"})
	for _, feed := range catalog.Feeds {
		if !allowedTiers[feed.Tier] {
			t.Fatalf("feed %s has invalid tier %q", feed.ID, feed.Tier)
		}
		if feed.ID == "" || feed.Name == "" || feed.URL == "" || len(feed.Beats) == 0 || feed.UseWhen == "" || feed.Notes == "" {
			t.Fatalf("feed %s missing required catalog fields: %#v", feed.ID, feed)
		}
		if feed.Tier == "gated" && (feed.Gate == nil || len(feed.Gate.RequiredBeats) == 0) {
			t.Fatalf("gated feed %s missing gate.required_beats", feed.ID)
		}
		feeds[feed.ID] = feed
	}
	for _, dropped := range []string{"google-news-science", "google-news-health"} {
		if _, ok := feeds[dropped]; ok {
			t.Fatalf("%s should have been removed", dropped)
		}
	}
	for _, id := range []string{"google-news-world", "google-news-us", "google-news-business"} {
		if feeds[id].Tier != "demoted" {
			t.Fatalf("%s tier=%q, want demoted", id, feeds[id].Tier)
		}
	}
	for _, id := range []string{"memeorandum", "mediagazer"} {
		if feeds[id].Tier != "gated" || feeds[id].Gate == nil || len(feeds[id].Gate.RequiredBeats) == 0 {
			t.Fatalf("%s gate missing: %#v", id, feeds[id])
		}
	}
	wantPrimary := map[string]string{
		"stat-news":      "https://www.statnews.com/feed/",
		"fda-press":      "https://www.fda.gov/about-fda/contact-fda/stay-informed/rss-feeds/press-releases/rss.xml",
		"cms-press":      "https://www.cms.gov/newsroom/rss-feeds",
		"endpoints-news": "https://endpts.com/feed/",
		"sec-edgar-8k":   "https://www.sec.gov/cgi-bin/browse-edgar?action=getcurrent&type=8-K&output=atom",
	}
	for id, url := range wantPrimary {
		feed, ok := feeds[id]
		if !ok {
			t.Fatalf("missing feed %s", id)
		}
		if feed.Tier != "primary" {
			t.Fatalf("%s tier=%q, want primary", id, feed.Tier)
		}
		if feed.URL != url {
			t.Fatalf("%s url=%q, want %q", id, feed.URL, url)
		}
	}
}

func TestCollectFeedConditionalGETPersistsETag(t *testing.T) {
	requests := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get("User-Agent") == "" {
			t.Fatal("missing User-Agent")
		}
		switch requests {
		case 1:
			if got := r.Header.Get("If-None-Match"); got != "" {
				t.Fatalf("first request If-None-Match=%q, want empty", got)
			}
			w.Header().Set("Content-Type", "application/rss+xml")
			w.Header().Set("ETag", "abc")
			_, _ = w.Write([]byte(`<?xml version="1.0"?><rss><channel><title>Test Feed</title><item><title>First item</title><link>https://example.com/first</link><description>Body</description><pubDate>Tue, 26 May 2026 13:00:00 GMT</pubDate></item></channel></rss>`))
		case 2:
			if got := r.Header.Get("If-None-Match"); got != "abc" {
				t.Fatalf("second request If-None-Match=%q, want abc", got)
			}
			w.WriteHeader(http.StatusNotModified)
			_, _ = w.Write([]byte(`<not-rss>`))
		default:
			t.Fatalf("unexpected request %d", requests)
		}
	}))
	defer server.Close()

	store := filepath.Join(t.TempDir(), "monitor.db")
	items, debug, errText := collectFeedWithStore(server.URL, 10, store, true)
	if errText != "" {
		t.Fatalf("first fetch error=%s", errText)
	}
	if len(items) != 1 {
		t.Fatalf("first fetch items=%d, want 1", len(items))
	}
	if debug["status"] != "parsed" {
		t.Fatalf("first fetch debug=%#v, want parsed", debug)
	}
	state, err := feedHTTPStateFor(server.URL, store)
	if err != nil {
		t.Fatal(err)
	}
	if state.ETag != "abc" {
		t.Fatalf("stored etag=%q, want abc", state.ETag)
	}

	items, debug, errText = collectFeedWithStore(server.URL, 10, store, true)
	if errText != "" {
		t.Fatalf("second fetch error=%s", errText)
	}
	if len(items) != 0 {
		t.Fatalf("second fetch items=%d, want 0", len(items))
	}
	if debug["status"] != "not_modified" || debug["not_modified"] != true {
		t.Fatalf("second fetch debug=%#v, want not_modified", debug)
	}
	if requests != 2 {
		t.Fatalf("requests=%d, want 2", requests)
	}
}
