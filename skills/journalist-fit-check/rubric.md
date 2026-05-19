# Journalist Fit Check Rubric

This rubric maps the source design into operational checks. Hard gates override the score. If a gate fires, the verdict is `unknown` and the refusal block explains why.

## Verdict Ladder

| Verdict | Confidence | Standard |
|---------|------------|----------|
| `fit` | `>= 0.80` | Exact or near-exact angle match, anchored to recent work. Reserve `> 0.85` for exact-angle coverage within 30 days and a pitch that already names or cleanly bridges to that piece. |
| `soft-fit` | `0.55-0.80` | Real but indirect overlap. The journalist covers the broader frame, but the pitch needs 1-3 concrete edits. |
| `no-fit` | `0.30-0.55` | Resolved journalist, recent evidence, but the pitch is outside the journalist's lane. |
| `unknown` | `< 0.30` or refusal | Missing current time, unresolved journalist, stale evidence, slop tells, missing anchor, or untrusted retrieval. |

Most real calls should land between `0.50` and `0.75`. If everything looks like `0.85`, the evaluator is flattering the pitch.

## Hard Gates

### Gate 1 - Current-time anchor

**Source trace:** Input schema; Time and decay.

Fail when `context.current_time_iso` is missing.

Result: `unknown`, `refusal.reason = "missing_current_time"`.

### Gate 2 - Journalist resolution

**Source trace:** Hard refusal conditions; unresolved pushback pattern.

Fail when the journalist cannot be tied to a public current identity: no author page, profile, recent byline, newsletter, personal site, or fetchable social footprint.

Beat strings alone fail resolution. Ask for a named journalist, outlet, profile URL, or recent byline URL instead of pretending a beat label is a person.

Result: `unknown`, `refusal.reason = "unresolved"`.

### Gate 3 - Slop tells in pitch

**Source trace:** Hard refusal conditions; slop-tells regex pack; placeholder pitch refusal pattern.

Fail when any hard slop tell appears in the pitch. Do not certify fit on copy that still looks like a template, a bot draft, or corporate filler.

Result: `unknown`, `refusal.reason = "slop_tells_in_pitch"`.

### Gate 4 - Anchor piece missing

**Source trace:** Anchor-piece check; anchoring definition; uncertainty refusal.

Fail when `fit` or `soft-fit` cannot cite a real, dated, URL-pointed piece by that journalist.

Result: `unknown`, `refusal.reason = "uncertainty_above_threshold"`.

### Gate 5 - Stale byline

**Source trace:** Decay rubric; stale-contact request.

Fail when the most recent verifiable byline is more than 90 days old at `current_time_iso`.

Result: `unknown`, `refusal.reason = "stale_data"`.

### Gate 6 - Hallucinated or unaudited anchor

**Source trace:** Hallucination guard.

Fail when an anchor title, URL, or date did not come from the retrieval surface, or when `anchor_pieces[].url` is missing from `retrieval_notes`.

Result: strip the anchor. If no anchor remains, return `unknown`.

## Scored Criteria

Score each criterion 0-2 after hard gates. Use the total to calibrate confidence, not to override judgment.

- **0** - Missing, false, stale, or generic.
- **1** - Present but weak, indirect, or under-audited.
- **2** - Specific, recent, cited, and usable.

Total possible: 20 points.

| Points | Default verdict range |
|--------|-----------------------|
| 17-20 | `fit`, if fit eligibility gates pass |
| 12-16 | `soft-fit` |
| 7-11 | `no-fit` or low `soft-fit`, depending on angle overlap |
| 0-6 | `unknown` unless a clean `no-fit` is better supported |

### 1. Retrieval audit trail

**Source trace:** Layer + rationale; output schema; retrieval notes.

**Score 0:** No retrieval surface named, or notes are vague.

**Score 1:** Surface named, but notes do not show enough of what was checked.

**Score 2:** `retrieval_surface` is one of `host-agent-search`, `medialyst`, or `cache`; `retrieval_notes` lists surfaces checked and URLs used as anchors.

### 2. Journalist identity and current role

**Source trace:** Refusal conditions; stale data pain; direct invocation paths.

**Score 0:** Identity is ambiguous, misspelled, stale, or outlet association cannot be verified.

**Score 1:** Journalist is likely identified, but current outlet or role is thinly supported.

**Score 2:** Journalist is resolved to a current outlet, newsletter, profile, or byline page.

### 3. Anchor-piece validity

**Source trace:** Anchor-piece check.

**Score 0:** No specific piece, no URL, no date, or generic "recent work" reasoning.

**Score 1:** Specific piece exists, but one field is weak: date uncertain, URL indirect, title paraphrased, or relevance note thin.

**Score 2:** Anchor has verbatim title, URL, parseable date within 90 days, and a relevance note tied to the pitch.

### 4. Decay discipline

**Source trace:** Decay rubric.

**Score 0:** Most recent byline is older than 90 days, or decay block is missing.

**Score 1:** Byline is 61-90 days old and warning is present.

**Score 2:** Byline is 60 days old or newer; decay block is complete.

### 5. Beat and angle overlap

**Source trace:** Verdict ladder; confidence floor for `fit`; no-fit examples.

**Score 0:** Recent work contradicts the pitch lane. Wrong beat, wrong outlet format, wrong audience, or wrong story type.

**Score 1:** Broad beat overlap only. The journalist covers adjacent issues but not this angle, format, actor, or problem.

**Score 2:** Direct overlap with the exact angle, named actor, problem, format, or story type in the pitch.

### 6. Pitch-to-anchor bridge

**Source trace:** What "specific changes" means; fit confidence floor.

**Score 0:** Pitch does not mention or plausibly connect to the anchor piece.

**Score 1:** Pitch can be edited into relevance with a small bridge.

**Score 2:** Pitch already names the anchor or clearly frames itself around the same gap, question, or problem.

### 7. Format fit

**Source trace:** Substack/byline-is-the-product rule; Substack edge case; breaking-news decay stage.

**Score 0:** Pitch asks for a format the journalist does not do: product launch to essayist, vendor briefing to columnist, evergreen pitch to breaking-news reporter, or listicle angle to enterprise reporter.

**Score 1:** Format could work with a reframe, but the current ask is mismatched.

**Score 2:** Pitch format matches the journalist's current mode: reported story, analysis, newsletter item, interview, embargo, data scoop, event invite, or other observed format.

### 8. Confidence calibration

**Source trace:** Confidence section.

**Score 0:** Confidence is inflated, unsupported, or outside the verdict threshold.

**Score 1:** Confidence roughly matches the verdict but does not reflect evidence quality.

**Score 2:** Confidence matches recency, directness, number of anchors, retrieval quality, and whether the pitch already bridges to the piece.

### 9. Suggested-change quality

**Source trace:** What "specific changes" means; soft-fit sample.

**Score 0:** Suggestions are vague, generic, or suggest "do more research."

**Score 1:** Suggestions name the angle but not the exact edit.

**Score 2:** For `soft-fit`, each suggestion names what to cut, replace, or add and ties the edit to a specific anchor piece. For `no-fit`, suggestions are empty.

### 10. No-fit discipline

**Source trace:** What you do not do; no-fit sample; pushback patterns.

**Score 0:** The verdict softens a clear no-fit to avoid conflict.

**Score 1:** The verdict says no-fit but hedges with unnecessary workarounds.

**Score 2:** The verdict plainly says the journalist is wrong for this pitch and does not launder the miss as a copy problem.

## Slop-Tells Pack

Run these against the pitch text. Case-insensitive unless noted. Any hard match triggers `slop_tells_in_pitch`.

```regex
# Bracketed placeholders
\{[A-Z][A-Za-z _\-]*\}
\[[A-Z][A-Z _\-]*\]
<<<[^>]+>>>

# Banned phrases
\bworld[- ]class\b
\binnovative\b
\bleading\b(?=\s+(?:provider|platform|company|firm|solution))
\bbest[- ]in[- ]class\b
\brevolutionary\b
\bwe are committed to\b
\bwe are (?:excited|thrilled) to (?:announce|share)\b
\bcutting[- ]edge\b
\bgame[- ]changer\b
\bgame[- ]changing\b
\bunlock(?:s|ing)? value\b
\bsynergy\b

# Bot sentence structures
\bIt'?s not just [^,.]+,?\s+it'?s\b
[A-Z][^—.!?]{0,80}—\s+and that'?s why\b
\bIn today'?s (?:fast[- ]paced|rapidly[- ]evolving) world\b

# Greeting voids
^(?:Hi|Hello|Hey)\s+[A-Z][a-z]+,?\s*\n?\s*Hope (?:you'?re well|this (?:finds you|email finds you) well)
\bI hope this (?:email|message) finds you well\b
```

Em-dash density is a warning, not automatic refusal: if `pitch.count("—") > 2` and `len(pitch) < 1500`, flag it. If em-dash density appears with any banned phrase, refuse.

## Generic Reasoning Rejects

These phrases are signs the evaluator failed to anchor the verdict:

```regex
\btheir recent work\b
\bthe outlet (?:often )?covers\b
\b(?:she|he|they) (?:often|tend(?:s)? to|frequently) cover(?:s)?\b
\bgiven (?:their|her|his) beat\b
\bbroadly relevant\b
\baligns with (?:their|her|his) interests\b
```

If reasoning contains these and there is no valid `anchor_pieces[]` entry, downgrade `fit` or `soft-fit` to `unknown`.

## Fit Eligibility

A `fit` verdict requires all of this:

- At least one anchor piece within the last 30 days.
- Direct topical relevance, not broad beat relevance.
- Pitch already names the piece or can be trivially edited to it.
- `days_since_last_byline <= 60`.
- No slop tells.
- No outlet-level-only reasoning.

If any item fails, downgrade to `soft-fit` or `unknown`.

## No-Fit Handling

For `no-fit`, the output should still cite recent work, but the relevance note explains why the anchor contradicts the pitch. Do not propose changes. A no-fit is a targeting problem, not a writing exercise.
