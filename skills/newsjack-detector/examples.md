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
  "scores": {
    "rank": 86.2
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
    "rationale": "The signal is same-cycle and confirmed by news plus expert reaction."
  },
  "why_newsjacking_worthy": "Regulator action creates a live need for explainers on AI compliance claims.",
  "client_standing": {
    "assessment": "strong",
    "rationale": "The client has enterprise AI governance proof and can explain claim substantiation."
  },
  "required_proof": [
    "Named compliance lead quote",
    "Specific customer-safe example or first-party governance data"
  ],
  "journalist_shape": {
    "beat_description": "Enterprise AI reporter covering compliance and regulator scrutiny",
    "why_they_care_now": "They need sourced reaction while the inquiry is fresh.",
    "do_not_target": "General startup roundups or consumer AI reviewers"
  },
  "evidence_used": [],
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

## Brand-Safety Block

Engine signal: public tragedy with high social momentum.

Block reason: `tragedy_or_human_suffering`

Do not hand off to angle generation.
