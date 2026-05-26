package main

import (
	"encoding/json"
	"path/filepath"
	"testing"
	"time"
)

const googleTrendsRSSFixture = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:ht="https://trends.google.com/trending/rss">
  <channel>
    <title>Daily Search Trends</title>
    <item>
      <title>AI outage</title>
      <link>https://trends.google.com/trending?geo=US</link>
      <pubDate>Tue, 26 May 2026 13:00:00 +0000</pubDate>
      <ht:approx_traffic>100K+</ht:approx_traffic>
      <ht:news_item>
        <ht:news_item_title>Cloud providers recover from AI outage</ht:news_item_title>
        <ht:news_item_url>https://example.com/ai-outage</ht:news_item_url>
        <ht:news_item_source>Example News</ht:news_item_source>
        <ht:news_item_snippet>AI tools were disrupted for enterprise customers.</ht:news_item_snippet>
      </ht:news_item>
      <ht:news_item>
        <ht:news_item_title>Developers look for workarounds</ht:news_item_title>
        <ht:news_item_url>https://example.org/developer-workarounds</ht:news_item_url>
        <ht:news_item_source>Example Tech</ht:news_item_source>
        <ht:news_item_snippet>Teams switched to backup workflows.</ht:news_item_snippet>
      </ht:news_item>
    </item>
  </channel>
</rss>`

func TestParseGoogleTrendsRSSExtractsNewsItems(t *testing.T) {
	trends, err := parseGoogleTrendsRSS(googleTrendsRSSFixture)
	if err != nil {
		t.Fatal(err)
	}
	if len(trends) != 1 {
		t.Fatalf("trends=%d, want 1", len(trends))
	}
	trend := trends[0]
	if trend.Title != "AI outage" || trend.ApproxTraffic != "100K+" {
		t.Fatalf("unexpected trend: %#v", trend)
	}
	if len(trend.NewsItems) != 2 {
		t.Fatalf("news items=%d, want 2", len(trend.NewsItems))
	}
	if trend.NewsItems[0].Title != "Cloud providers recover from AI outage" {
		t.Fatalf("news title=%q", trend.NewsItems[0].Title)
	}
	if trend.NewsItems[0].URL != "https://example.com/ai-outage" {
		t.Fatalf("news url=%q", trend.NewsItems[0].URL)
	}
	if trend.NewsItems[0].Source != "Example News" {
		t.Fatalf("news source=%q", trend.NewsItems[0].Source)
	}
	if trend.NewsItems[0].Snippet != "AI tools were disrupted for enterprise customers." {
		t.Fatalf("news snippet=%q", trend.NewsItems[0].Snippet)
	}

	now := time.Date(2026, 5, 26, 14, 0, 0, 0, time.UTC)
	items := googleTrendsEvidenceMaps(trends, "US", 24, googleTrendsRSSURL("US", 24), now)
	if len(items) != 2 {
		t.Fatalf("mapped items=%d, want 2", len(items))
	}
	if items[0]["source"] != "google_trends" || items[0]["title"] != "AI outage" {
		t.Fatalf("unexpected mapped item: %#v", items[0])
	}
	if items[0]["url"] != "https://example.com/ai-outage" {
		t.Fatalf("mapped url=%q", items[0]["url"])
	}
	metadata := valueOrEmptyMap(items[0]["metadata"])
	if metadata["google_trends_approx_traffic"] != "100K+" || metadata["google_trends_geo"] != "US" {
		t.Fatalf("metadata missing trend fields: %#v", metadata)
	}
	if metadata["google_trends_news_item_title"] != "Cloud providers recover from AI outage" {
		t.Fatalf("metadata news title=%v", metadata["google_trends_news_item_title"])
	}
}

func TestProfileRoundTripPreservesGoogleTrends(t *testing.T) {
	repo := repoRootForTest(t)
	profile, err := profileFromFile(filepath.Join(repo, "fixtures/newsjack-detector-agent/profile.clearnym.json"))
	if err != nil {
		t.Fatal(err)
	}
	if geo := googleTrendsGeo(profile.GoogleTrends); geo != "US" {
		t.Fatalf("geo=%q, want US", geo)
	}
	if hours := googleTrendsHours(profile.GoogleTrends); hours != 24 {
		t.Fatalf("hours=%d, want 24", hours)
	}

	data, err := json.Marshal(profile.publicDict())
	if err != nil {
		t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatal(err)
	}
	roundTripped := profileFromMap(payload)
	if geo := googleTrendsGeo(roundTripped.GoogleTrends); geo != "US" {
		t.Fatalf("round-trip geo=%q, want US", geo)
	}
	if hours := googleTrendsHours(roundTripped.GoogleTrends); hours != 24 {
		t.Fatalf("round-trip hours=%d, want 24", hours)
	}
}
