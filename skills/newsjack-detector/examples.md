# Newsjack Detector Examples

## Pitch Now

Engine signal:

```json
{
  "id": "s1",
  "title": "FTC opens inquiry into AI compliance claims",
  "sources": ["news_search", "x"],
  "features": {
    "decay_bucket": "4hr",
    "source_count": 2,
    "seen_before": false,
    "profile_matches": ["AI compliance", "enterprise governance"],
    "safety_flags": []
  },
  "routing": {
    "lane": "profile_relevance",
    "queue_priority": 86.2,
    "demoted": false
  },
  "mechanical_scores": {
    "freshness": 1.0,
    "source_agreement": 0.78,
    "novelty": 1.0,
    "profile_match": 0.44,
    "source_quality": 0.825,
    "momentum": 0.21,
    "major_news": 0.0
  }
}
```

Skill output:

```json
{
  "signal_id": "s1",
  "signal_title": "FTC opens inquiry into AI compliance claims",
  "verdict": "pitch_now",
  "decay": {
    "stage": "4hr",
    "rationale": "The signal is same-cycle by verified first-public clock, not just the search-result timestamp."
  },
  "first_publication": {
    "status": "fresh",
    "surfaced_article_published_at": "2026-05-25T13:14:00Z",
    "first_public_at": "2026-05-25T13:10:00Z",
    "original_url": "https://www.ftc.gov/news-events/news/press-releases/example",
    "canonical_coverage_url": "https://www.reuters.com/legal/government/ftc-opens-inquiry-ai-compliance-claims-2026-05-25/",
    "canonical_coverage_source": "Reuters",
    "rationale": "The official FTC press release is the earliest verified public source and is inside the 24-hour cron window."
  },
  "why_newsjacking_worthy": "Regulator action creates a live need for explainers on AI compliance claims.",
  "client_standing": {
    "assessment": "strong",
    "rationale": "The client works directly in enterprise AI governance and can explain claim substantiation."
  },
  "journalist_shape": {
    "beat_description": "Enterprise AI reporter covering compliance and regulator scrutiny",
    "why_they_care_now": "They need sourced reaction while the inquiry is fresh.",
    "do_not_target": "General startup roundups or consumer AI reviewers"
  },
  "evidence_used": [
    {
      "source": "Reuters",
      "title": "FTC opens inquiry into AI compliance claims",
      "url": "https://www.reuters.com/legal/government/ftc-opens-inquiry-ai-compliance-claims-2026-05-25/"
    },
    {
      "source": "FTC",
      "title": "FTC opens inquiry into AI compliance claims",
      "url": "https://www.ftc.gov/news-events/news/press-releases/example"
    }
  ],
  "next_skill": "reactive-comment"
}
```

## Monitor

Engine signal: X discussion only, no news confirmation.

Verdict: `monitor`

Reason: "Single-source chatter with no confirmed news event. Watch for news search confirmation or official filing."

## Reject

Engine signal: "Influencers debate AI regulation again" with `month` decay and no new document.

Verdict: `reject`

Reason: `stale`

Engine signal: AOL article published today, canonical URL points to a BBC story from May 4 with no new development.

Verdict: `reject`

Reason: `stale`

`first_publication.status`: `stale`

Engine signal: secondary article published today, no canonical/source metadata, and searches do not verify the first public source.

Verdict: `reject`

Reason: `freshness_unverified`

## Brand-Safety Block

Engine signal: public tragedy with high social momentum.

Block reason: `tragedy_or_human_suffering`

Do not hand off to angle generation.
