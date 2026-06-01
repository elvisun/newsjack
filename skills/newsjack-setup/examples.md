# Newsjack Setup Examples

## Local Falcon-Style Profile

```json
{
  "profile": {
    "company": "Local Falcon",
    "website": "https://www.localfalcon.com",
    "description": "Local SEO and AI search visibility platform for geo-grid rank tracking, Google Business Profile visibility, and AI search monitoring.",
    "topics": [
      "local rank tracking",
      "AI search visibility",
      "Google Business Profile optimization",
      "geo-grid rank tracking",
      "local SEO analytics"
    ],
    "competitors": [
      "BrightLocal",
      "Whitespark",
      "Semrush Local",
      "Yext",
      "Local Viking"
    ],
    "search_terms": [
      "local rank tracking",
      "local SEO rank tracker",
      "geo-grid rank tracking",
      "AI search visibility tracking",
      "Google Business Profile rank tracking",
      "Google AI Overviews local visibility",
      "ChatGPT local visibility",
      "BrightLocal local rank tracking",
      "Whitespark local rank tracker",
      "Semrush Local",
      "Yext local SEO"
    ],
    "feed_urls": [
      "https://www.techmeme.com/feed.xml",
      "https://news.google.com/rss/headlines/section/topic/TECHNOLOGY?hl=en-US&gl=US&ceid=US:en",
      "https://news.google.com/rss/headlines/section/topic/BUSINESS?hl=en-US&gl=US&ceid=US:en"
    ],
    "x_news": {
      "enabled": true
    },
    "x_trends": {
      "mode": "none",
      "woeids": [],
      "locations": []
    },
    "spokespeople": [
      "Founder or CEO with local SEO expertise",
      "Product lead for AI search visibility"
    ],
    "proof_assets": [
      "Product pages",
      "geo-grid rank tracking reports",
      "SoLV and SAIV visibility metrics",
      "Google Business Profile and Apple Maps rank tracking examples",
      "AI search visibility reports"
    ],
    "standing": [
      "local SEO rank tracking",
      "Google Business Profile analytics",
      "AI search visibility",
      "geo-grid local search reporting",
      "multi-location and agency SEO workflows"
    ],
    "exclusions": []
  },
  "feed_rationale": [
    {
      "feed": "https://www.techmeme.com/feed.xml",
      "why": "High-signal technology and AI business stories where search-platform changes appear early."
    },
    {
      "feed": "https://news.google.com/rss/headlines/section/topic/TECHNOLOGY?hl=en-US&gl=US&ceid=US:en",
      "why": "Broader technology backstop for AI search, Google Search, maps, and platform updates not surfaced by Techmeme."
    },
    {
      "feed": "https://news.google.com/rss/headlines/section/topic/BUSINESS?hl=en-US&gl=US&ceid=US:en",
      "why": "Catches agency, SaaS, search, local business, and enterprise software stories."
    }
  ],
  "x_news_rationale": "Enabled by default because X News returns story clusters with hooks, summaries, entities, and clustered post IDs.",
  "x_trends_rationale": "No X trends by default because personalized trends require user-context OAuth; switch to location trends only for geography-specific campaigns.",
  "run_commands": {
    "hourly_major_news": "~/.newsjack/bin/newsjack detector run --profile profile.json --feed-only --save --new-only --max-age-hours 48",
    "profile_relevance": "~/.newsjack/bin/newsjack detector run \"AI search visibility\" --profile profile.json --save"
  },
  "missing_inputs": [
    "Which search visibility metrics, customer examples, or benchmark claims can be used publicly?"
  ]
}
```
