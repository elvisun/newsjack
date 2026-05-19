# Reactive Comment Rubric

Every query gets one score and then runs through the refusal gates. The
score decides whether drafting is even eligible. The gates decide whether a
draft is allowed to leave the model.

Trace: all criteria are compressed from the source design's sections on
Topic fit, Anti-spray rubric, Anti-hallucination rubric, Decay rubric,
Banned-word list, Em-dash check, Sentence-shape blocks, Placeholder tells,
and the three output modes.

## Score Bands

| Fit score | Verdict |
|-----------|---------|
| 85-100 | `draft`, if every gate passes |
| 65-84 | `draft`, if every gate passes and any weak spots are named |
| 40-64 | `kill`, unless one missing fact could push the score over 65, then `ask` |
| 0-39 | `kill` |

Drafting requires **65+ and clean gates**. A high score never overrides a
hard veto.

## Weighted Fit Score

### 1. Specific Expertise Match - 40 Points

> Does the query's exact ask intersect the profile's exact expertise?

- **0 points:** No real intersection, only broad adjacency.
- **10 points:** Same general industry, wrong problem.
- **20 points:** Same problem family, but the query asks for a narrower
  experience the profile does not show.
- **30 points:** Strong topical fit, with one missing scope detail such as
  company stage, geography, title, or customer type.
- **40 points:** Direct fit. The query asks for something the profile
  explicitly names.

Red flags:

- "software" treated as expertise
- "founder" treated as permission to comment on every business topic
- query asks for a title the user does not hold
- query asks for first-hand experience and the profile only supports
  analysis

### 2. Proof-Point Support - 25 Points

> Can the response make a substantive claim backed by the profile?

- **0 points:** No proof point backs the answer.
- **8 points:** A proof point exists, but only supports a generic bio line.
- **16 points:** A proof point supports the topic, but not the specific
  angle the query requests.
- **25 points:** At least one proof point directly backs the claim the
  draft would make.

For `draft`, 100% of concrete claims need provenance. Allowed sources:

- `profile.proof_points[i]`
- `recent_context.journalist_bylines[i]`
- `"USER MUST CONFIRM"`

No general-knowledge claims. No invented stats. No fabricated expert
credentials.

### 3. Do-Not-Comment and Identity Fit - 15 Points

> Is the query inside the user's declared lane?

- **0 points:** The query touches `profile.do_not_comment_on`, asks for a
  credential or title not in profile, or would force a competitor/product
  comparison the profile forbids. This is a hard veto.
- **8 points:** The query is adjacent to a do-not topic but can be answered
  cleanly without entering it.
- **15 points:** No conflict with do-not topics, title, employer, stage, or
  identity.

Hard-veto examples:

- profile says no cryptocurrency; query asks for on-chain auth
- profile says no competitor comparisons; query asks Cursor vs. Copilot
- query asks for CISOs; profile only shows CTO
- query asks for Series A-B; profile has no funding-stage proof

### 4. Journalist Context and Real Personalization - 10 Points

> Is there a real anchor to the journalist's beat?

- **0 points:** The draft would need to fake familiarity with the
  journalist's work.
- **5 points:** Recent context is unavailable; neutral score. Open with the
  user's credential instead of pretending.
- **7 points:** Recent context shows the journalist covers the broad beat.
- **10 points:** A recent byline or post matches the query's topic and can
  be referenced specifically by URL, title, or topic.

Never praise generically. "Saw your recent piece" requires a supplied
recent-context title or URL.

### 5. Outlet, Source, and Requirement Hygiene - 10 Points

> Does the response respect the outlet, source platform, and query terms?

- **0 points:** Outlet matches `profile.outlets_to_skip`, the query says
  "no vendors" and the only angle is vendor-coded, or a source-platform
  rule blocks the response. Outlet skip is a hard veto.
- **4 points:** The outlet is acceptable, but the draft would brush against
  a requirement or source-platform norm.
- **7 points:** Requirements are mostly satisfied, with one minor caveat
  that should be stamped.
- **10 points:** Outlet passes filters, query requirements are met, and the
  response is visibly human-reviewed rather than automated spray.

Examples:

- "No vendor pitches" means no product pitch.
- "Real anecdotes only" means analysis is not enough.
- "Academics only" means operators do not qualify.
- Featured/HARO should share one cap when the profile treats them as one
  source family.

## Hard Gates

Any hard-gate failure overrides the score.

### Decay Gate

- `hours_until_deadline > 24`: fresh.
- `12 < hours_until_deadline <= 24`: proceed with warning.
- `2 < hours_until_deadline <= 12`: proceed with tight-window warning.
- `0 <= hours_until_deadline <= 2`: `ask` unless recent context is already
  loaded and every other gate is clean.
- `hours_until_deadline < 0`: `kill`.

If `received_at_iso` is older than 48 hours, warn that the user is late in
the source queue.

### Anti-Spray Gate

Default cap: 5 responses per source per rolling week.

- If `profile.response_cap_per_week` is missing, use 5.
- If it is above 10, `ask` for justification before drafting.
- If `responses_to_this_source_this_week >= cap`, `kill`.
- If `responses_to_this_source_this_week >= cap * 0.8`, proceed only with a
  warning.

Always stamp:

```yaml
anti_spray:
  responses_to_this_source_this_week: 0
  cap: 5
  passed: true
```

### Anti-Hallucination Gate

Refuse or downgrade if any concrete claim lacks provenance.

Concrete claims include:

- people
- companies
- products
- publications
- employers
- customer anecdotes
- statistics
- prior bylines
- conference talks
- funding stage
- headcount
- geography

If the claim could be true but is not in profile or recent context, source
it as `"USER MUST CONFIRM"` and flag it in `next_action`. If that would make
the draft misleading, return `ask` instead.

### Slop Gate

Block the draft if the body contains any of these case-insensitive
substrings:

```text
world-class
innovative
leading
revolutionary
best-in-class
we are committed to
cutting-edge
synergy
synergies
game-changing
thought leader
thought leadership
next-generation
robust
seamless
leverage
leveraging
unparalleled
industry-leading
groundbreaking
disrupting
disruptive
we are excited to
thrilled to announce
proud to announce
transform the way
redefining
reimagine
unlock
empower
empowering
comprehensive solution
end-to-end solution
holistic
turn-key
turnkey
mission-critical
move the needle
deliver value
add value
```

The word `leading` can appear only as plain grammar inside supplied profile
material, never as self-awarded praise in the draft.

### AI-Tell Gate

Block the draft if the body contains:

- an em dash character
- "it's not just X, it's Y"
- "this isn't just X, it's Y"
- "X isn't just Y. It's Z."
- "in a world where"
- "whether you're X or Y"
- "not only X but also Y"
- "at the intersection of X and Y"
- a triple dash list
- title case marketing phrases in the middle of a sentence

### Placeholder Gate

Block placeholder leakage:

- `{...}`
- `[...]` when used for fields such as company, name, title, topic, outlet,
  date, or "your X"
- `<<...>>`
- `__...__`
- `{{first_name}}` or similar mail-merge fields

### Identity Gate

The response comes from the user, not from the agent.

Refuse if:

- the draft mentions being an AI or assistant
- the draft uses a name other than `profile.name`
- the contact block is rewritten instead of appended verbatim
- the bio invents a title, employer, credential, or publication

## Output Gate

Before returning:

- exactly one YAML block
- no surrounding prose unless requested
- one query only
- one verdict only
- `anti_spray` always present
- `slop_check` present on drafts and any slop downgrade
- `provenance` present on drafts
- `next_action` says manual review and manual send on drafts
