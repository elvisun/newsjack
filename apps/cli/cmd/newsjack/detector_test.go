package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
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
		code := runCLI([]string{"detector", "run", "AI customer support", "--mock", "--include-all-scored"}, &out, &err)
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

func TestDetectorRunWritesScoredOutputArtifact(t *testing.T) {
	repo := repoRootForTest(t)
	dir := t.TempDir()
	scored := filepath.Join(dir, "scored_candidates.json")
	withTempEnv(t, map[string]string{
		"HOME":          t.TempDir(),
		"NEWSJACK_ROOT": repo,
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"detector", "run", "AI customer support", "--mock", "--scored-output", scored}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("detector code=%d stderr=%s", code, errBuf.String())
		}
		payload, err := readJSONMap(scored)
		if err != nil {
			t.Fatalf("read scored output: %v", err)
		}
		if got := len(signalSlice(payload["signals"])); got != 2 {
			t.Fatalf("scored signals=%d, want 2", got)
		}
	})
}

func TestXSourceAvailabilityUsesBearerTokenOnly(t *testing.T) {
	withoutBearer := configFromEnv()
	withoutBearer["X_BEARER_TOKEN"] = ""
	withoutBearer["TWITTER_BEARER_TOKEN"] = ""
	withoutBearer["X_API_BEARER_TOKEN"] = ""
	withoutBearer["TWITTER_API_BEARER_TOKEN"] = ""
	if got := availableSources(withoutBearer, []string{"x", "x_news", "x_trends"}); len(got) != 0 {
		t.Fatalf("x sources without bearer=%v, want none", got)
	}

	withBearer := configFromEnv()
	withBearer["X_BEARER_TOKEN"] = "token"
	got := availableSources(withBearer, []string{"x", "x_news", "x_trends"})
	if len(got) != 3 || got[0] != "x" || got[1] != "x_news" || got[2] != "x_trends" {
		t.Fatalf("x sources with bearer=%v, want all x sources", got)
	}
}

func TestWOEIDStringFormatsJSONNumbers(t *testing.T) {
	if got := woeidString(float64(2459115)); got != "2459115" {
		t.Fatalf("woeidString(float64)=%q, want 2459115", got)
	}
	if got := woeidString("23424975"); got != "23424975" {
		t.Fatalf("woeidString(string)=%q, want 23424975", got)
	}
}

func TestPersonalizedXTrendsRequiresUserContextOAuth(t *testing.T) {
	items, errText := collectXTrendsRaw(map[string]any{"mode": "personalized"}, "quick", "bearer-token")
	if len(items) != 0 {
		t.Fatalf("personalized trend items=%d, want none", len(items))
	}
	if !strings.Contains(errText, "user-context OAuth") {
		t.Fatalf("err=%q, want user-context OAuth guidance", errText)
	}
}

func TestSourceStatusReportsNoResultsForAttemptedEmptySource(t *testing.T) {
	status := sourceStatus(
		[]string{"x_news"},
		[]string{"x_news"},
		false,
		false,
		map[string]int{},
		map[string]any{},
		map[string]string{"X_BEARER_TOKEN": "token"},
		false,
	)
	xNews := valueOrEmptyMap(status["x_news"])
	if xNews["status"] != "no_results" || xNews["attempted"] != true {
		t.Fatalf("x_news status=%#v, want attempted no_results", xNews)
	}
}

func TestSourceStatusTreatsMockSourcesAsAvailable(t *testing.T) {
	status := sourceStatus(
		[]string{"x_news"},
		[]string{"x_news"},
		false,
		false,
		map[string]int{},
		map[string]any{},
		map[string]string{},
		true,
	)
	xNews := valueOrEmptyMap(status["x_news"])
	if xNews["status"] != "no_results" || xNews["available"] != true {
		t.Fatalf("x_news status=%#v, want mock no_results available", xNews)
	}
}

func TestDetectorRunAllowsXTrendsOnly(t *testing.T) {
	repo := repoRootForTest(t)
	dir := t.TempDir()
	profilePath := filepath.Join(dir, "profile.json")
	profile := `{
	  "company": "Nofar Method",
	  "topics": ["reformer Pilates"],
	  "x_news": {"enabled": false},
	  "x_trends": {"mode": "location", "woeids": ["2459115"], "locations": ["New York"]}
	}`
	if err := os.WriteFile(profilePath, []byte(profile), 0o644); err != nil {
		t.Fatalf("write profile: %v", err)
	}
	withTempEnv(t, map[string]string{
		"HOME":          t.TempDir(),
		"NEWSJACK_ROOT": repo,
	}, func() {
		var out, errBuf bytes.Buffer
		code := runCLI([]string{"detector", "run", "--profile", profilePath, "--sources", "x_trends", "--no-x-news", "--no-profile-feeds", "--mock", "--min-queue-priority", "0"}, &out, &errBuf)
		if code != 0 {
			t.Fatalf("detector code=%d stderr=%s", code, errBuf.String())
		}
		var payload map[string]any
		if err := json.Unmarshal(out.Bytes(), &payload); err != nil {
			t.Fatalf("invalid JSON: %s", out.String())
		}
		if got := len(signalSlice(payload["signals"])); got != 1 {
			t.Fatalf("signals=%d, want mocked x trend; payload=%s", got, out.String())
		}
		status := valueOrEmptyMap(valueOrEmptyMap(payload["diagnostics"])["source_status"])
		xTrends := valueOrEmptyMap(status["x_trends"])
		if xTrends["status"] != "used" {
			t.Fatalf("x_trends status=%#v, want used", xTrends)
		}
	})
}

func TestProfileIgnoresLegacyAssetInventory(t *testing.T) {
	legacyAssetKey := "pro" + "of_assets"
	profile := profileFromMap(map[string]any{
		"company":      "ExampleCo",
		"topics":       []any{"workforce planning"},
		legacyAssetKey: []any{"legacy phrase should not match"},
	})
	if _, ok := profile.publicDict()[legacyAssetKey]; ok {
		t.Fatalf("public profile should not expose legacy asset inventory: %#v", profile.publicDict())
	}
	if strings.Contains(profile.matchText(), "legacy phrase") {
		t.Fatalf("match text should not include legacy asset inventory: %q", profile.matchText())
	}
	if got := profileMatches(profile, "legacy phrase should not match"); len(got) != 0 {
		t.Fatalf("profileMatches used legacy asset inventory: %v", got)
	}
}

func TestSearchXNewsCallsXAPIDirectly(t *testing.T) {
	var gotPath, gotAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotAuth = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"id":"story-1","name":"Example X Story","summary":"A timely X story cluster.","updated_at":"2026-05-27T12:00:00Z"}]}`))
	}))
	defer server.Close()

	withTempEnv(t, map[string]string{"NEWSJACK_X_API_BASE": server.URL}, func() {
		response := searchXNews("Example", "quick", 24, "x-token")
		if response["error"] != nil {
			t.Fatalf("searchXNews error=%v", response["error"])
		}
		if gotPath != "/2/news/search" {
			t.Fatalf("path=%s, want /2/news/search", gotPath)
		}
		if gotAuth != "Bearer x-token" {
			t.Fatalf("authorization=%q, want bearer token", gotAuth)
		}
		items := parseXNewsResponse(response, "Example")
		if len(items) != 1 || items[0]["title"] != "Example X Story" {
			t.Fatalf("items=%#v, want parsed X story", items)
		}
	})
}

func TestWeakDetectorLanesSelectionFloors(t *testing.T) {
	now := time.Date(2026, 5, 25, 13, 0, 0, 0, time.UTC)
	profile := monitorProfile{
		Company:  "Clearnym",
		Topics:   []string{"data broker removal"},
		Standing: []string{"identity theft prevention", "privacy operations"},
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
			if tt.source == "x_news" {
				if priority := queuePriority(signal); priority < defaultMinQueuePriority {
					t.Fatalf("x_news queue priority=%v, want surfaced for review", priority)
				}
				if !passesSelectionFloor(signal, defaultMinQueuePriority, defaultMinMajorNews) {
					t.Fatalf("x_news did not pass selection floor: %#v", signal["routing"])
				}
				return
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

func TestXNewsProfileMatchUsesContextMetadata(t *testing.T) {
	now := time.Date(2026, 5, 25, 13, 0, 0, 0, time.UTC)
	profile := monitorProfile{
		Topics:      []string{"computer-use agents"},
		Competitors: []string{"Anthropic computer use"},
		Standing:    []string{"AI agent benchmarks"},
	}
	item := evidenceItem{
		Source:      "x_news",
		Title:       "Un cluster en français sur la cybersécurité",
		URL:         "https://x.com/search?q=anthropic",
		PublishedAt: now.Format(time.RFC3339),
		Metadata: map[string]any{
			"x_news_contexts": map[string]any{
				"entities": map[string]any{
					"organizations": []any{"Anthropic"},
					"products":      []any{"computer use"},
				},
				"topics": []any{"Technology", "AI"},
			},
		},
	}
	score := profileMatchScore(profile, item.text())
	if score <= 0 {
		t.Fatalf("profile match score=%v, want metadata-driven overlap", score)
	}
}

func TestXTrendLocationMetadataDoesNotClusterUnrelatedTrends(t *testing.T) {
	now := time.Date(2026, 5, 25, 13, 0, 0, 0, time.UTC).Format(time.RFC3339)
	items := []evidenceItem{
		{
			Source:      "x_trends",
			Title:       "Smotrich",
			URL:         "https://x.com/search?q=Smotrich&f=live",
			PublishedAt: now,
			Metadata:    map[string]any{"x_signal_type": "trend", "x_trend_location": "New York", "x_trend_woeid": "2459115"},
		},
		{
			Source:      "x_trends",
			Title:       "Zendaya",
			URL:         "https://x.com/search?q=Zendaya&f=live",
			PublishedAt: now,
			Metadata:    map[string]any{"x_signal_type": "trend", "x_trend_location": "New York", "x_trend_woeid": "2459115"},
		},
	}
	if clusters := clusterItems(items); len(clusters) != 2 {
		t.Fatalf("clusters=%d, want unrelated trends split: %#v", len(clusters), clusters)
	}
}

func TestXTrendLocationConfigDoesNotCreateProfileMatch(t *testing.T) {
	profile := monitorProfile{
		Company: "Nofar Method",
		Topics:  []string{"reformer Pilates"},
		XTrends: map[string]any{"mode": "location", "locations": []any{"New York", "Miami"}},
	}
	item := evidenceItem{
		Source:    "x_trends",
		Title:     "Saylor",
		Excerpt:   "Saylor. New York",
		Author:    "x-trends",
		Container: "x.com/trends",
		Metadata:  map[string]any{"x_signal_type": "trend", "x_trend_location": "New York", "x_trend_woeid": "2459115"},
	}
	if score := profileMatchScore(profile, item.text()); score >= 0.05 {
		t.Fatalf("profile match score=%v, want local trend below x_trends relevance floor", score)
	}
}

func TestXTrendLocationOnlyOverlapIsDemoted(t *testing.T) {
	now := time.Date(2026, 5, 25, 13, 0, 0, 0, time.UTC)
	profile := monitorProfile{
		Company:     "Nofar Method",
		Description: "Boutique Pilates studio with classes in New York and Miami.",
		Topics:      []string{"reformer Pilates"},
		Standing:    []string{"NYC and Miami wellness trends"},
	}
	signal := scoreSignal(signalCluster{Evidence: []evidenceItem{{
		Source:      "x_trends",
		Title:       "Saylor",
		URL:         "https://x.com/search?q=Saylor&f=live",
		Excerpt:     "Saylor. New York",
		Container:   "x.com/trends",
		PublishedAt: now.Format(time.RFC3339),
		Engagement:  map[string]any{},
		Metadata:    map[string]any{"x_signal_type": "trend", "x_trend_location": "New York", "x_trend_woeid": "2459115"},
	}}}, profile, map[string]map[string]any{}, now, detectorOptions{LookbackDays: 1, XTrendsMinProfileMatch: 0.05})
	if lane := signalLaneValue(signal); lane != "x_trends_unmatched" {
		t.Fatalf("lane=%s, want x_trends_unmatched", lane)
	}
	if match := floatValue(valueOrEmptyMap(signal["mechanical_scores"])["profile_match"]); match != 0 {
		t.Fatalf("profile_match=%v, want location-only trend match zeroed", match)
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

func TestLargeRemoteRelevantStoryPassesSelectionFloor(t *testing.T) {
	signal := map[string]any{
		"routing": map[string]any{
			"lane":           "profile_relevance_weak",
			"queue_priority": 39.9,
		},
		"mechanical_scores": map[string]any{
			"profile_match": 0.014,
			"story_size":    0.58,
		},
	}
	if !passesSelectionFloor(signal, defaultMinQueuePriority, defaultMinMajorNews) {
		t.Fatalf("large remote-relevant story did not pass selection floor: %#v", signal)
	}
}

func TestSmallWeakStoryStillDoesNotPassSelectionFloor(t *testing.T) {
	signal := map[string]any{
		"routing": map[string]any{
			"lane":           "profile_relevance_weak",
			"queue_priority": 39.9,
		},
		"mechanical_scores": map[string]any{
			"profile_match": 0.014,
			"story_size":    0.08,
		},
	}
	if passesSelectionFloor(signal, defaultMinQueuePriority, defaultMinMajorNews) {
		t.Fatalf("small weak story passed selection floor: %#v", signal)
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

func TestParseNewsResponseFiltersSyndicationPublications(t *testing.T) {
	items := parseNewsResponse(map[string]any{
		"news": []any{
			map[string]any{
				"title":  "Republished privacy explainer",
				"link":   "https://www.aol.com/articles/republished-privacy-explainer",
				"source": "AOL",
				"date":   "2 hours ago",
				"metadata": map[string]any{
					"publication_type":                  "syndication_platform",
					"domain_authority":                  92,
					"estimated_monthly_organic_traffic": 12000000,
				},
			},
			map[string]any{
				"title":  "Original privacy enforcement story",
				"link":   "https://example.com/original-privacy-enforcement",
				"source": "Example",
				"date":   "1 hour ago",
				"metadata": map[string]any{
					"publication_type":                  "editorial",
					"domain_authority":                  73,
					"estimated_monthly_organic_traffic": 460000,
				},
			},
		},
	})
	if len(items) != 1 {
		t.Fatalf("items=%d, want only non-syndication item: %#v", len(items), items)
	}
	if items[0]["title"] != "Original privacy enforcement story" {
		t.Fatalf("kept title=%v, want original editorial item", items[0]["title"])
	}
	metadata := valueOrEmptyMap(items[0]["metadata"])
	if metadata["publication_type"] != "editorial" {
		t.Fatalf("metadata=%#v, want editorial item preserved", metadata)
	}
}
