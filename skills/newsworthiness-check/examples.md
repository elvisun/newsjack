# Newsworthiness Check Examples

## Event: Ride

Input:

```json
{
  "event_headline": "FTC opens inquiry into AI compliance claims",
  "event_timestamp": "2026-05-20T14:00:00-04:00",
  "user_company": "Enterprise AI governance vendor",
  "user_expertise_area": "AI compliance workflows and claim substantiation",
  "evidence": [
    {"source": "news_search", "title": "FTC opens inquiry into AI compliance claims", "url": "https://example.com/ftc-ai", "published_at": "2026-05-20T14:00:00-04:00"}
  ]
}
```

Output:

```json
{
  "mode": "event_newsjacking",
  "overall_score": 7,
  "band": "significant",
  "recommended_action": "RIDE",
  "will_get_coverage": "likely in trade press, possible in mainstream business press",
  "calibration_anchor": {
    "closest_anchor": "Federal agency action with sector-wide implications",
    "why_not_higher": "This is not yet a sweeping rule or court decision with national public impact",
    "why_not_lower": "Regulator action creates immediate demand for expert explanation"
  },
  "dimensions": {
    "magnitude": {"score": 7, "rationale": "Affects enterprise AI vendors and buyers making compliance claims"},
    "velocity": {"score": 7, "rationale": "Fresh same-day news with expected trade pickup"},
    "novelty": {"score": 6, "rationale": "AI compliance scrutiny is familiar, but an inquiry is a concrete escalation"},
    "standing": {"score": 8, "rationale": "The user directly works on AI governance and claim substantiation"},
    "window": {"score": 8, "rationale": "Same-cycle response window is still open"}
  },
  "caps_applied": [],
  "kill_switch_triggered": false,
  "evidence_used": [
    {
      "source": "news_search",
      "title": "FTC opens inquiry into AI compliance claims",
      "url": "https://example.com/ftc-ai",
      "published_at": "2026-05-20T14:00:00-04:00"
    }
  ],
  "evidence_gaps": [
    "Need confirmation of reporter traction or X discussion if prioritizing same-hour outreach"
  ],
  "honest_assessment": "This is a real newsjacking opportunity because the user has direct standing and the story creates a same-day need for explainers.",
  "next_move": {
    "skill": "reactive-comment",
    "rationale": "Prepare a restrained sourced comment while the window is open"
  }
}
```

## Event: Skip Despite Big News

Input: a consumer snack brand wants to comment on a major antitrust lawsuit against a cloud infrastructure company.

Output:

```json
{
  "mode": "event_newsjacking",
  "overall_score": 4,
  "band": "routine-to-significant event, weak user fit",
  "recommended_action": "SKIP",
  "will_get_coverage": "the event will get coverage, but not with this user as a source",
  "calibration_anchor": {
    "closest_anchor": "Major sector legal story",
    "why_not_higher": "The user's lack of standing caps the newsjacking score",
    "why_not_lower": "The event itself is legitimate and timely"
  },
  "dimensions": {
    "magnitude": {"score": 8, "rationale": "Large company legal action with industry implications"},
    "velocity": {"score": 7, "rationale": "Fresh event likely to spread across business press"},
    "novelty": {"score": 5, "rationale": "Antitrust action is notable but not unprecedented"},
    "standing": {"score": 1, "rationale": "The snack brand has no credible connection to cloud infrastructure or antitrust"},
    "window": {"score": 7, "rationale": "The timing window is open, but irrelevant to this user"}
  },
  "caps_applied": [
    {
      "cap": "no_client_standing_max_4",
      "applied": true,
      "rationale": "A large event does not create permission for an unrelated brand to comment"
    }
  ],
  "kill_switch_triggered": false,
  "evidence_used": [],
  "evidence_gaps": [],
  "honest_assessment": "The story is newsworthy. Your connection to it is not.",
  "next_move": {
    "skill": null,
    "rationale": "Do not pitch this angle"
  }
}
```

## Event: Avoid

Input: a company wants to use an active natural disaster as a hook for a product promotion.

Output:

```json
{
  "mode": "event_newsjacking",
  "overall_score": null,
  "band": "blocked",
  "recommended_action": "AVOID",
  "will_get_coverage": "not applicable",
  "calibration_anchor": {
    "closest_anchor": "brand-safety kill switch",
    "why_not_higher": "Active human suffering cannot be used as a promotional hook",
    "why_not_lower": "Score is not assigned because this is blocked before scoring"
  },
  "dimensions": {},
  "caps_applied": [],
  "kill_switch_triggered": true,
  "evidence_used": [],
  "evidence_gaps": [],
  "honest_assessment": "Do not newsjack this. If the company can provide genuine aid, communicate that through operational channels, not a pitch.",
  "next_move": {
    "skill": null,
    "rationale": "No PR handoff is appropriate"
  }
}
```

## Pitch: Revise

Input:

```json
{
  "pitch_headline": "Acme raises $15M Series A to transform enterprise workflow",
  "pitch_type": "funding",
  "company_stage": "series_a",
  "company_notability": "unknown startup",
  "target_journalist_beat": "enterprise software",
  "data_available": false,
  "exclusive_angle": false,
  "human_story_available": false
}
```

Output:

```json
{
  "mode": "pitch_newsworthiness",
  "overall_score": 4,
  "band": "marginally newsworthy",
  "recommended_action": "REVISE",
  "will_get_coverage": "unlikely outside narrow trade coverage",
  "calibration_anchor": {
    "closest_anchor": "Standard Series A at an unknown startup",
    "why_not_higher": "The pitch has no original data, customer proof, or timely external hook",
    "why_not_lower": "A $15M Series A can be briefly coverable for an enterprise software beat"
  },
  "dimensions": {
    "beat_relevance": {"score": 6, "rationale": "Enterprise software is the right general beat"},
    "timeliness": {"score": 3, "rationale": "The only timing is the company's announcement date"},
    "magnitude": {"score": 4, "rationale": "$15M is a normal Series A, not a market signal by itself"},
    "proof_points": {"score": 2, "rationale": "No data, customer, exclusive, or methodology is supplied"},
    "narrative_quality": {"score": 3, "rationale": "The framing is generic transformation language"},
    "source_credibility": {"score": 4, "rationale": "Unknown startup with no stated independent credentials"}
  },
  "weak_dimensions": ["timeliness", "proof_points", "narrative_quality"],
  "improvement_suggestions": [
    "Lead with a customer or usage metric if one exists",
    "Tie the pitch to a specific workflow problem reporters are already covering",
    "Replace 'transform enterprise workflow' with the concrete before-and-after"
  ],
  "evidence_used": [],
  "evidence_gaps": [
    "Customer names or anonymized customer proof",
    "Usage, revenue, retention, or efficiency metrics",
    "Reason this matters this week beyond the funding close"
  ],
  "honest_assessment": "This is a routine funding pitch. It may earn a brief if the beat is narrow and the proof improves, but it is not a strong story yet.",
  "next_move": {
    "skill": "angle-generator",
    "rationale": "Find a non-funding angle before drafting outreach"
  }
}
```

## Pitch: Hold

Input: "Our CEO has thoughts on the future of AI."

Output:

```json
{
  "mode": "pitch_newsworthiness",
  "overall_score": 1,
  "band": "not newsworthy",
  "recommended_action": "HOLD",
  "will_get_coverage": "no",
  "calibration_anchor": {
    "closest_anchor": "Generic thought leadership",
    "why_not_higher": "There is no news hook, proof, beat, data, or exclusive access",
    "why_not_lower": "This is already the floor for non-news"
  },
  "dimensions": {
    "beat_relevance": {"score": 1, "rationale": "No specific beat is identified"},
    "timeliness": {"score": 1, "rationale": "No current event or date is attached"},
    "magnitude": {"score": 1, "rationale": "No scale or affected audience is stated"},
    "proof_points": {"score": 1, "rationale": "No evidence is supplied"},
    "narrative_quality": {"score": 1, "rationale": "There is no story structure"},
    "source_credibility": {"score": 2, "rationale": "A CEO title alone is not enough standing"}
  },
  "weak_dimensions": ["beat_relevance", "timeliness", "magnitude", "proof_points", "narrative_quality", "source_credibility"],
  "improvement_suggestions": [
    "Do not pitch this as-is",
    "Bring original data, a current news hook, or a specific operational lesson before revisiting"
  ],
  "evidence_used": [],
  "evidence_gaps": [],
  "honest_assessment": "This is content marketing, not news.",
  "next_move": {
    "skill": null,
    "rationale": "Hold until there is a real story"
  }
}
```
