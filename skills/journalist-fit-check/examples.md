# Journalist Fit Check — Worked Examples

Each example shows the before input, the verdict object, and the after action. The point is not to flatter the pitch. The point is to stop bad targeting before it becomes email.

---

## Example 1: Clean Fit, Recent Anchor

### Before

```json
{
  "journalist": { "name": "Maxwell Zeff", "outlet": "TechCrunch" },
  "pitch": "Subject: Open-source eval framework for agentic AI - built after the Anthropic constitutional AI launch\n\nHi Maxwell,\n\nFollowing your Apr. 28 piece on developer adoption of Claude's agent SDK, we're launching an open-source eval harness that benchmarks agent tool-use against your suggested 'real workflow' tests - not just synthetic benchmarks. Repo opens Tuesday. Happy to give you first look + access to two pilot customers (one in legal, one in healthcare) who have measurable workflow-time data.\n\nFour-line technical summary attached. Embargo through 9am ET Tuesday if useful.\n\nJordan",
  "context": {
    "client_or_subject": "AgentEval, open-source agent benchmarking harness",
    "current_time_iso": "2026-05-18T14:00:00Z"
  }
}
```

### After

```json
{
  "verdict": "fit",
  "confidence": 0.86,
  "reasoning": "Fit. Maxwell Zeff's Apr. 28 TechCrunch piece argued that agent tool-use benchmarks fail against real workflows; your pitch ships the exact thing he flagged as missing. The embargo, repo opening, and pilot customers give him a usable scoop shape.",
  "anchor_pieces": [
    {
      "title": "Developers say Claude's agent SDK is the easiest to deploy, hardest to evaluate",
      "url": "https://techcrunch.com/2026/04/28/claude-agent-sdk-eval-gap/",
      "published_at": "2026-04-28",
      "hours_since_publish": 480,
      "relevance_note": "The piece names workflow-time benchmarks as the missing piece; the pitch launches that exact thing."
    }
  ],
  "suggested_changes": [],
  "refusal": {
    "refused": false,
    "reason": null,
    "remediation": null
  },
  "decay": {
    "last_verified_byline_at": "2026-05-16",
    "days_since_last_byline": 2,
    "decay_warning": null
  },
  "retrieval_surface": "host-agent-search",
  "retrieval_notes": "Checked TechCrunch author page and recent search results. Anchor URL: https://techcrunch.com/2026/04/28/claude-agent-sdk-eval-gap/. Recent bylines also found May 3, May 11, and May 16."
}
```

### Action

Send the pitch after normal copy polish. The target is right.

---

## Example 2: Soft-Fit, Needs A Reframe

### Before

```json
{
  "journalist": { "name": "Casey Newton", "outlet": "Platformer" },
  "pitch": "Hi Casey,\n\nWe just shipped AgentEval, an open-source benchmarking harness for AI agents. It's the first tool to measure agent tool-use against real workflows. Would love to set up a briefing.\n\nThanks,\nJordan",
  "context": {
    "client_or_subject": "AgentEval, open-source agent benchmarking harness",
    "current_time_iso": "2026-05-18T14:00:00Z"
  }
}
```

### After

```json
{
  "verdict": "soft-fit",
  "confidence": 0.62,
  "reasoning": "Soft-fit. Casey Newton's May 9 Platformer piece on Anthropic's external-evaluation push is adjacent, but he does not cover developer-tool launches as product news. The pitch needs to become about evals as a trust and policy lever, not about a new harness.",
  "anchor_pieces": [
    {
      "title": "Anthropic's bet on external evaluation as a trust signal",
      "url": "https://www.platformer.news/p/anthropic-external-eval-trust",
      "published_at": "2026-05-09",
      "hours_since_publish": 216,
      "relevance_note": "He frames eval tools as a trust signal. The pitch can fit that frame, but not as a generic developer-tool launch."
    }
  ],
  "suggested_changes": [
    "Cut the opener. Replace it with: 'Your May 9 piece on Anthropic's external-eval push named the open-source gap. We're shipping that gap on Tuesday.'",
    "Drop 'first tool to measure.' Replace it with the specific measurement: 'workflow-completion latency on a fixed task suite, replicable by any agent.'",
    "Add one sentence on why this matters for policy and trust, not only engineering."
  ],
  "refusal": {
    "refused": false,
    "reason": null,
    "remediation": null
  },
  "decay": {
    "last_verified_byline_at": "2026-05-16",
    "days_since_last_byline": 2,
    "decay_warning": null
  },
  "retrieval_surface": "host-agent-search",
  "retrieval_notes": "Checked six recent Platformer posts from Apr. 22 through May 16 plus recent public X results. Anchor URL: https://www.platformer.news/p/anthropic-external-eval-trust."
}
```

### Revised Pitch

```text
Hi Casey,

Your May 9 piece on Anthropic's external-eval push named the open-source gap. We're shipping that gap on Tuesday: AgentEval, an open-source harness that measures workflow-completion latency on a fixed task suite, replicable by any agent.

Two pilot customers are using it to decide which agent workflows are safe enough for regulated teams. If useful, I can send the repo under embargo and set up a 20-minute call with the maintainer.

Jordan
```

---

## Example 3: No-Fit, Do Not Rescue It

### Before

```json
{
  "journalist": { "url": "https://www.bloomberg.com/authors/AS0E8x9aH-0w/emily-chang" },
  "pitch": "Subject: New B2B accounting SaaS for freelancers\n\nHi Emily,\n\nWe just launched an accounting platform for freelancers and would love your coverage in Bloomberg.",
  "context": {
    "current_time_iso": "2026-05-18T14:00:00Z"
  }
}
```

### After

```json
{
  "verdict": "no-fit",
  "confidence": 0.41,
  "reasoning": "No-fit. Emily Chang's recent Bloomberg work is executive interviews, venture capital, and AI company strategy, not B2B SaaS launches or freelancer tooling. A freelancer-accounting product is outside her lane and will not become relevant through copy edits.",
  "anchor_pieces": [
    {
      "title": "Inside Anduril's pitch to the Pentagon",
      "url": "https://www.bloomberg.com/news/articles/2026/05/11/anduril-pitch",
      "published_at": "2026-05-11",
      "hours_since_publish": 168,
      "relevance_note": "Shows the company and executive-strategy level she covers; it contradicts the freelancer-SaaS ask."
    }
  ],
  "suggested_changes": [],
  "refusal": {
    "refused": false,
    "reason": null,
    "remediation": null
  },
  "decay": {
    "last_verified_byline_at": "2026-05-15",
    "days_since_last_byline": 3,
    "decay_warning": null
  },
  "retrieval_surface": "host-agent-search",
  "retrieval_notes": "Reviewed Bloomberg author page and last five visible bylines. None touched SMB accounting, freelancer tooling, fintech for individuals, or product reviews. Anchor URL: https://www.bloomberg.com/news/articles/2026/05/11/anduril-pitch."
}
```

### Action

Drop the contact. Do not rewrite this for her.

---

## Example 4: Refusal, Stale Data

### Before

```json
{
  "journalist": { "name": "Olivia Solon", "outlet": "NBC News" },
  "pitch": "Subject: New privacy-preserving analytics platform\n\nHi Olivia,\n\nWe are launching a privacy-preserving analytics platform and thought this would be relevant to your privacy reporting.",
  "context": {
    "current_time_iso": "2026-05-18T14:00:00Z"
  }
}
```

### After

```json
{
  "verdict": "unknown",
  "confidence": 0.10,
  "reasoning": "Refused on stale data. The most recent verifiable NBC News byline I found for Olivia Solon is from Mar. 22, 2024, which is 788 days old at current_time_iso. Do not pitch this contact until the role is reverified.",
  "anchor_pieces": [],
  "suggested_changes": [],
  "refusal": {
    "refused": true,
    "reason": "stale_data",
    "remediation": "Verify Olivia Solon's current role through a recent byline search, LinkedIn, Twitter/X, or her current outlet page. If she moved, re-call this skill with the updated identifier. If she left journalism, drop her from the list."
  },
  "decay": {
    "last_verified_byline_at": "2024-03-22",
    "days_since_last_byline": 788,
    "decay_warning": "Last verified byline is 788 days old. This contact is not safe to use without manual verification."
  },
  "retrieval_surface": "host-agent-search",
  "retrieval_notes": "Checked NBC News author search, site search, and public profile snippets. No NBC News byline newer than Mar. 22, 2024 was found."
}
```

### Action

Verify the journalist's current role before doing anything else. The pitch is not the blocker; the contact record is.

---

## Example 5: Independent Newsletter, Narrow Fit

### Before

```json
{
  "journalist": { "url": "https://thewenerdweekly.substack.com" },
  "pitch": "Subject: Press dinner invite - small group, AI infra founders\n\nHi Brandon,\n\nWe're hosting a small, off-the-record press dinner in NYC on June 12: eight people, all founders building AI infrastructure that real engineers actually use. No decks, no pitches, just a conversation. Would love to have you. Happy to share the guest list privately if useful.\n\nJordan, founder, AgentEval",
  "context": {
    "current_time_iso": "2026-05-18T14:00:00Z"
  }
}
```

### After

```json
{
  "verdict": "fit",
  "confidence": 0.82,
  "reasoning": "Fit. Brandon Wenerd's May 7 Substack post told PR people to stop sending mailers and go back to small, curated press dinners or low-stakes hangouts. Your pitch offers that exact format, and the private guest-list offer respects the newsletter-persona model.",
  "anchor_pieces": [
    {
      "title": "A PR Person Asked Me How to Pitch Substackers",
      "url": "https://thewenerdweekly.substack.com/p/a-pr-person-asked-me-how-to-pitch",
      "published_at": "2026-05-07",
      "hours_since_publish": 264,
      "relevance_note": "The post explicitly asks for small curated events instead of mailers; the pitch follows that instruction."
    }
  ],
  "suggested_changes": [
    "Optional: name one guest in the email itself. The guest list is the product; one credible name does more than another adjective."
  ],
  "refusal": {
    "refused": false,
    "reason": null,
    "remediation": null
  },
  "decay": {
    "last_verified_byline_at": "2026-05-16",
    "days_since_last_byline": 2,
    "decay_warning": null
  },
  "retrieval_surface": "host-agent-search",
  "retrieval_notes": "Checked Substack archive directly. Anchor URL: https://thewenerdweekly.substack.com/p/a-pr-person-asked-me-how-to-pitch. Four recent posts found; latest post was May 16."
}
```

### Action

Send after adding one credible guest name. Do not turn it into a product pitch.
