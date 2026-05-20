# Newsjack Detector Rubric

Use this rubric after the engine returns ranked evidence. The engine score is only queue priority.

## Verdict Ladder

### pitch_now

Use only when all are true:

- Evidence is fresh: usually `30min`, `4hr`, or `24hr`.
- At least one credible news source exists, preferably `news_search`.
- The client has direct standing to comment.
- The client can supply proof or a real spokesperson now.
- A specific reporter shape is obvious.
- No hard brand-safety block applies.

### develop_angle

Use when the signal is real but needs framing:

- Fresh or still within the week.
- Client standing is plausible but not yet sharp.
- Proof exists or can be requested.
- A journalist shape exists, but the angle needs work.

Handoff: `angle-generator`.

### monitor

Use when the signal is interesting but not pitch-ready:

- Single-source or weak cross-source confirmation.
- Early chatter without enough news confirmation.
- The client might have standing, but proof is missing.
- The signal may matter if it gains traction.

### reject

Use when any core gate fails:

- stale
- no client standing
- missing proof that cannot be supplied quickly
- no plausible journalist shape
- off-beat
- already seen with no new development
- weak source quality

## Decay

- `30min` - live/breaking. Only use for immediate comment if the client can respond now.
- `4hr` - same-cycle. Good for reactive comment.
- `24hr` - still fresh. Good for angle generation or same-day response.
- `week` - trend/context only. Do not call it breaking.
- `month` - usually not a newsjack unless paired with a new data point or fresh hook.
- `unknown` - do not pitch as timely without independent timestamp verification.

## Standing

Strong standing:

- The client operates directly in the affected market.
- The client has first-party data, customer evidence, technical expertise, or a named executive who can speak concretely.
- The signal names the client's category, customers, regulators, technology, or competitors.

Partial standing:

- The client has adjacent expertise but needs a narrower angle.
- The client can explain impact but not the core event.

Weak standing:

- The client merely sells into the broad category.
- The client wants to comment because the topic is popular.
- The proof is generic thought leadership.

## Proof

Required proof must be specific:

- first-party data
- named customer or user example
- real executive quote
- technical artifact
- regulatory or market analysis the client can defend
- recent product or customer evidence

Do not accept generic claims like "we help companies with AI" as proof.

## Journalist Shape

A useful journalist shape names:

- exact beat
- outlet archetype
- why the beat cares now
- who should not receive it

Bad shapes:

- "business reporter"
- "AI journalist"
- "tech media"
- "industry press"

Good shapes:

- "enterprise AI reporter covering vendor compliance claims after regulator action"
- "cybersecurity trade reporter covering identity-risk fallout from new enforcement"
- "retail operations reporter covering labor-cost impact of a same-day policy change"

## Hard Blocks

Block signals built on:

- death
- violence
- disaster
- war
- abuse
- sexual violence
- missing people
- humanitarian crisis
- hate crime
- terror
- suicide

The only acceptable work around these topics is restrained expert commentary with direct public-interest standing. Promotional hooks are refused.
