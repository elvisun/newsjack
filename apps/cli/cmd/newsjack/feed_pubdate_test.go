package main

import "testing"

// RSS 2.0 dates the item with <pubDate>. parseFeed lowercases every element
// name before storing it, so the value lands under "pubdate" while
// feedItemDict reads m["pubDate"] — the date is dropped and the item arrives
// undated. Atom's <published>/<updated> are unaffected because they are
// already lowercase.
//
// Both timezone spellings are covered because both occur in the wild: Google
// News and Techmeme emit the "GMT" form, WordPress-backed feeds emit "+0000".
func TestParseFeedReadsRSSPubDate(t *testing.T) {
	for _, tc := range []struct {
		name    string
		pubDate string
	}{
		{"literal GMT zone", "Wed, 16 Sep 2026 14:05:00 GMT"},
		{"numeric zone", "Wed, 16 Sep 2026 14:05:00 +0000"},
		{"offset zone", "Wed, 16 Sep 2026 16:05:00 +0200"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rss := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Example Feed</title>
    <item>
      <title>Regulators approve statewide AI rules for schools</title>
      <link>https://example.com/a</link>
      <source>Example News</source>
      <pubDate>` + tc.pubDate + `</pubDate>
      <description>The board voted unanimously.</description>
    </item>
  </channel>
</rss>`

			items, err := parseFeed(rss, "https://example.com/feed.xml", 10)
			if err != nil {
				t.Fatalf("parseFeed returned an error: %v", err)
			}
			if len(items) != 1 {
				t.Fatalf("expected 1 item, got %d", len(items))
			}

			got := stringValue(items[0]["published_at"])
			if got == "" {
				t.Fatal("published_at is empty: RSS <pubDate> was dropped, so every RSS " +
					"item arrives undated and the detector's freshness gate cannot act on it")
			}
			if want := "2026-09-16T14:05:00Z"; got != want {
				t.Fatalf("published_at = %q, want %q (an un-normalised date is still "+
					"unusable downstream)", got, want)
			}
		})
	}
}

// Atom feeds date items with <published>, which already matches the lowercased
// key. This guards the fix against a regression that would swap one form for
// the other instead of supporting both.
func TestParseFeedReadsAtomPublished(t *testing.T) {
	const atom = `<?xml version="1.0" encoding="UTF-8"?>
<feed xmlns="http://www.w3.org/2005/Atom">
  <title>Example Atom</title>
  <entry>
    <title>Districts must offer a non-AI alternative</title>
    <link href="https://example.com/b"/>
    <published>2026-09-17T09:30:00Z</published>
    <summary>Parental opt-out requires a comparable path.</summary>
  </entry>
</feed>`

	items, err := parseFeed(atom, "https://example.com/atom.xml", 10)
	if err != nil {
		t.Fatalf("parseFeed returned an error: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	if got, want := stringValue(items[0]["published_at"]), "2026-09-17T09:30:00Z"; got != want {
		t.Fatalf("published_at = %q, want %q", got, want)
	}
}
