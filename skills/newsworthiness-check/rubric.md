# Newsworthiness Check Rubric

Use this rubric to calibrate news judgment. The score is not a vibe. It is a forced placement against known anchors.

## Score Bands

| Score | Band | Meaning | Anchors |
|-------|------|---------|---------|
| 10 | historic | Generational event, global impact, remembered for years | pandemic declaration, major war beginning, death of globally recognized head of state |
| 8-9 | major | National/global headline, multi-day mainstream coverage | Supreme Court ruling, systemic bank failure, mega-acquisition, major election result |
| 6-7 | significant | Industry or sector leads, some mainstream pickup | $100M+ funding, notable CEO departure, mass layoffs, major product category launch |
| 4-5 | routine | Coverable, but not memorable | standard Series A, regional policy item, expected earnings story, incremental product launch |
| 2-3 | marginal | Niche pickup or database coverage at best | seed round, VP hire, vague partnership, ordinary feature release |
| 1 | not news | Belongs on the company site, not in media | blog post, anniversary, internal update, generic thoughts on a trend |

Distribution expectation:

- 1-2: common
- 3-4: common
- 5-6: plausible but not default
- 7-8: rare
- 9-10: exceptional

If three consecutive evaluations score `7+`, run the sanity check again. That pattern usually means inflation.

## Mode A: Event Newsjackability

Question: Is this public event worth riding for this user?

Score dimensions:

| Dimension | Weight | What To Judge |
|-----------|--------|---------------|
| Magnitude | 25% | People affected, dollars at stake, geographic scope, institutional importance |
| Velocity | 25% | How fast the story is spreading now through news search, Reddit, and X |
| Novelty | 15% | Whether this is new, surprising, record-setting, or just another instance |
| Standing | 20% | Whether the user has direct expertise, product relevance, data, or affected-customer context |
| Window | 15% | Whether the newsjacking window is open, peaking, saturated, or stale |

### Event Actions

| Action | Use When |
|--------|----------|
| `RIDE` | Score `6+`, good standing, proof available, window open, no kill switch |
| `WAIT` | Score `5-6`, still developing, proof or confirmation missing |
| `SKIP` | Score `1-4`, saturated, no journalist shape, or no standing |
| `AVOID` | Brand-safety kill switch applies |

### Timing And Decay

| Stage | Meaning | Guidance |
|-------|---------|----------|
| `30min` | Live breaking | Only act if the user can respond immediately with direct expertise |
| `4hr` | Same-cycle | Strong window for reactive comment |
| `24hr` | Still fresh | Good for angle generation or same-day response |
| `week` | Context window | Only pitch if there is new data, a contrarian proof point, or a clear trend piece |
| `month` | Mostly stale | Usually not a newsjack |
| `unknown` | Timestamp unclear | Do not call it timely |

Typical lifecycle:

- 1-6 hours after break: best window for fast stories.
- 24-48 hours: usually decaying unless the story has new developments.
- After 48 hours with declining velocity: cap at 5.

### Standing Gate

| Standing | Score Effect |
|----------|--------------|
| Direct expert | Can score normally |
| Adjacent expert | Usually max 7; angle must be narrow |
| Industry observer | Usually max 6; needs data or unusual access |
| Random company | Max 4 |
| No connection | Skip, no matter how big the event is |

Strong standing requires at least one:

- The user operates directly in the affected market.
- The user has first-party data.
- The user has named technical or domain expertise.
- The story affects the user's customers, regulators, competitors, or core technology.
- The user can explain a consequence journalists are already trying to understand.

### Kill Switches

Return `AVOID` when the event is built on:

- active mass casualties
- death, violence, abuse, terrorism, hate crime, suicide, missing people, war, disaster, or humanitarian crisis
- child harm
- political violence
- active rescue or emergency response

Narrow exception: direct public-interest expertise with no promotional hook. Even then, lower confidence and require restraint.

### Saturation

| Saturation | Signals | Action |
|------------|---------|--------|
| pre-viral | low coverage, rising velocity | Ideal if standing is strong |
| rising | increasing coverage and social spread | Good with a distinct angle |
| peak | everywhere, velocity plateauing | Risky; only act with exceptional proof |
| saturated | declining velocity, takes everywhere | Skip |
| backlash | hot takes are being criticized | Skip |

Use news search, Reddit, and X where available. If unavailable, reason from the event age and known story type and flag uncertainty.

## Mode B: Pitch Newsworthiness

Question: Is the user's own pitch idea worth a journalist's time?

Score dimensions:

| Dimension | Weight | What To Judge |
|-----------|--------|---------------|
| Beat relevance | 25% | Whether a specific journalist beat plausibly covers this |
| Timeliness | 20% | Whether there is a live hook, date, trend, embargo, or external moment |
| Magnitude | 15% | Scale of money, users, customers, market impact, people affected |
| Proof points | 15% | Original data, named customer, document, quote, technical proof, exclusive access |
| Narrative quality | 15% | Human story, conflict, stakes, surprise, tension |
| Source credibility | 10% | Whether the source has standing and credentials |

### Pitch Actions

| Action | Use When |
|--------|----------|
| `PROCEED` | Score `6+`, clear beat, proof, and credible timing |
| `REVISE` | Score `3-5`, there is a salvageable story but weak proof, timing, or framing |
| `HOLD` | Score `1-2`, no beat, no timing, no proof, or purely promotional |

### Pitch Type Baselines

| Pitch Type | Baseline | Notes |
|------------|----------|-------|
| Original research/data | 6-8 | Strong when surprising, credible, and exclusive |
| Expert commentary on breaking news | 5-7 | Depends on standing and speed |
| Trend analysis with examples | 5-6 | Needs specific evidence, not vibes |
| Funding $100M+ | 6-7 | Higher only for notable company or unusual market signal |
| Funding $30M-$100M | 5-6 | Trade interest, maybe mainstream if category is hot |
| Funding $10M-$30M | 3-5 | Usually routine |
| Funding under $10M | 2-3 | Rarely covered beyond databases |
| Product launch, new category | 5-7 | Only if genuinely novel and proved |
| Product launch, incremental | 2-3 | Usually changelog material |
| Partnership | 2-3 | Usually marketing unless concrete impact exists |
| VP or exec hire | 1-3 | Higher only for independently notable person/company |
| Company milestone or anniversary | 1-2 | Self-congratulatory unless externally meaningful |
| Thought leadership | 1-2 | Opinion is not news without a hook or proof |

### Proof That Raises Scores

Good proof:

- first-party data with a clear methodology
- named customer or user story
- exclusive access, embargo, or early look
- regulator filing, court document, or technical artifact
- defensible executive or domain-expert quote
- concrete metrics: dollars, users, growth, time saved, risk reduced

Weak proof:

- market-size claims
- unnamed customer anecdotes
- investor name-dropping
- percentages without absolute numbers
- "our CEO thinks"
- generic claims about AI, innovation, disruption, or the future

## Caps

Apply these after the initial weighted score:

| Trigger | Cap |
|---------|-----|
| No client standing for event | max 4 |
| No proof for pitch | max 4 |
| No specific journalist beat | max 4 |
| Stale event, older than 48h, no new development | max 5 |
| Single-source claim, no confirmation | max 5 |
| Purely promotional language dominates | max 3 |
| Fabricated or unverifiable central claim | max 2 |

## Anti-Inflation Pitfalls

Discount these automatically:

- "first-of-its-kind" without proof
- "groundbreaking", "revolutionary", "game-changing", "industry-leading"
- market size used as a substitute for company traction
- high growth percentages from a tiny base
- strategic partnership with no numbers or operational effect
- investor roster used as the main story
- long input that repeats the same weak fact many ways

## Improvement Guidance

Tie suggestions to weak dimensions:

| Weak Dimension | Useful Fix |
|----------------|------------|
| Beat relevance | Find the narrower reporter shape or stop pitching this beat |
| Timeliness | Tie to a current event, embargo, report date, regulatory moment, or live trend |
| Magnitude | Add absolute numbers, customer count, dollars, people affected, or market effect |
| Proof points | Add original data, a named customer, a document, or a defensible quote |
| Narrative quality | Find the tension, protagonist, consequence, or surprising change |
| Source credibility | Use a qualified spokesperson or add credentials that prove standing |

## Sanity Check

Before returning:

1. Name the closest anchor.
2. Ask whether the item is truly equivalent to that anchor.
3. Apply caps.
4. State why the score is not higher.
5. State why the score is not lower.

The final score should feel almost a little harsh to a marketer and fair to a journalist.
