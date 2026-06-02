package main

import "testing"

// ev builds a news_search evidence item with the given domain authority, monthly
// traffic, and publication_type — the three fields story-size scoring reads.
func ev(url string, authority, traffic float64, pubType string) evidenceItem {
	return evidenceItem{
		Source: "news_search",
		Title:  "headline",
		URL:    url,
		Metadata: map[string]any{
			"domain_authority":                  authority,
			"estimated_monthly_organic_traffic": traffic,
			"publication_type":                  pubType,
		},
	}
}

func bandOf(t *testing.T, c signalCluster) (string, any) {
	t.Helper()
	out := storySizeScore(c, 0.95, 0, 0)
	return stringValue(out["band"]), out["score"]
}

// TestStorySizeIgnoresLonePressRelease pins the AibleClaw failure: a single
// brand_content/newswire release syndicated onto a high-authority aggregator
// (TradingView, DA 91, 230M traffic) must NOT inflate to the high/major band.
// With no editorial outlet to score, it falls to the unknown band.
func TestStorySizeIgnoresLonePressRelease(t *testing.T) {
	for _, pt := range []string{"brand_content", "newswire", "press_release", "sponsored"} {
		c := signalCluster{Evidence: []evidenceItem{
			ev("https://www.tradingview.com/news/acn107446/", 91, 230_000_000, pt),
		}}
		band, score := bandOf(t, c)
		if band != "unknown" {
			t.Errorf("%s release alone: band=%q score=%v, want unknown (no editorial outlet to credit)", pt, band, score)
		}
	}
}

// TestStorySizeKeepsEditorialAlongsidePromotional verifies that when a story has
// both a press release on a giant aggregator and modest genuine editorial
// coverage, only the editorial outlet drives the band — the release is ignored,
// not blended in to lift the score.
func TestStorySizeKeepsEditorialAlongsidePromotional(t *testing.T) {
	editorialOnly := signalCluster{Evidence: []evidenceItem{
		ev("https://customerthink.com/post/", 40, 44_000, "editorial"),
	}}
	withRelease := signalCluster{Evidence: []evidenceItem{
		ev("https://customerthink.com/post/", 40, 44_000, "editorial"),
		ev("https://www.tradingview.com/news/acn107446/", 91, 230_000_000, "brand_content"),
	}}
	bandA, scoreA := bandOf(t, editorialOnly)
	bandB, scoreB := bandOf(t, withRelease)
	if bandA != bandB {
		t.Errorf("band changed when a press release was added: %q -> %q (release must not count)", bandA, bandB)
	}
	if floatValue(scoreA) != floatValue(scoreB) {
		t.Errorf("score changed when a press release was added: %v -> %v (release must not count)", scoreA, scoreB)
	}
}

// TestStorySizeEditorialStillScores guards against the skip going too far: a
// normal editorial item on a high-authority domain must still reach a high band.
func TestStorySizeEditorialStillScores(t *testing.T) {
	c := signalCluster{Evidence: []evidenceItem{
		ev("https://techcrunch.com/story/", 92, 7_700_000, "editorial"),
	}}
	band, score := bandOf(t, c)
	if band != "high" && band != "major" {
		t.Errorf("editorial on a DA-92 outlet: band=%q score=%v, want high/major", band, score)
	}
}
