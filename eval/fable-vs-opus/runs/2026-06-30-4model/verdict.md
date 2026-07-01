# Four-model verdict — 2026-06-30 (50 brands, round-robin)

**Setup.** The same 50 brands and constructed company updates as the 2026-06-09
study, extended from a 2-model duel to a **4-model round-robin**:

| slug | model | how its angle sets were produced |
|---|---|---|
| `opus` | **Opus 4.8** | reused from 2026-06-09 (angle-generator subagent) |
| `fable` | **Fable 5** | reused from 2026-06-09 (Fable is disabled; cannot be regenerated) |
| `s5` | **Sonnet 5** | generated 2026-06-30, headless `claude -p` (tools off, skill inlined) |
| `s46` | **Sonnet 4.6** | generated 2026-06-30, headless `claude -p` (tools off, skill inlined) |

Every model ran the **identical** `angle-generator` skill on identical input;
the only intended variable is the model. All C(4,2)=6 pairs were judged blind by
**GPT-5.5** (`codex exec`, `meanest-editor` persona) in **both orderings** — the
opus↔fable pair (100 judgments) reused from 2026-06-09, plus **5 new pairs × 50
brands × 2 orderings = 500 new judgments**, for **600 total**. `aggregate-nmodel.py`
re-anchors every judgment to model identity so position bias cancels.

> **On "Sonnet 4.7".** There is no accessible Sonnet 4.7 in this account/timeline
> (`claude-sonnet-4-7` does not resolve; base Sonnet 4 retired 2026-06-15). The
> latest pre-5 Sonnet is **Sonnet 4.6** (`claude-sonnet-4-6`), used here as the
> previous-generation Sonnet — matching `eval/newsworthiness-check`.

## Headline

**Fable 5 produced the strongest angle set, and the order is unambiguous:
Fable 5 › Opus 4.8 › Sonnet 5 › Sonnet 4.6.** Fable led **all 7 dimensions**,
won every head-to-head, and took **249/300 `publishable` verdicts**. Sonnet 5 is
a real but modest step up from Sonnet 4.6 (beats it 13–5 in robust wins), yet
**both Sonnets trail Fable and Opus clearly** — most of all on `grounding`, where
the judge repeatedly caught them decorating facts the update didn't supply.

The reused opus↔fable pair reproduced the 2026-06-09 result **exactly** (Fable
24–7 robust over Opus; position bias 0.687 vs 0.69 before) — a clean internal
consistency check that the harness and judge are stable.

## Per-model mean (1–5, ranked)

| dimension | Fable 5 | Opus 4.8 | Sonnet 5 | Sonnet 4.6 |
|---|---|---|---|---|
| news_value | **4.44** | 4.15 | 3.98 | 3.96 |
| distinctness | **4.80** | 4.49 | 4.20 | 4.22 |
| journalist_shape | **4.89** | 4.69 | 4.44 | 4.44 |
| grounding | **3.98** | 3.83 | 3.33 | 2.99 |
| anti_slop | **4.79** | 4.62 | 4.23 | 3.96 |
| proof_rigor | **4.88** | 4.86 | 4.45 | 4.53 |
| usefulness | **4.77** | 4.58 | 4.29 | 4.21 |
| **overall** | **4.65** | **4.46** | **4.13** | **4.04** |

Meanest-editor verdicts (`publishable` / `workshopable`, of 300 each):
**Fable** 249/51 · **Opus** 205/95 · **Sonnet 5** 114/186 · **Sonnet 4.6** 74/226.

## Head-to-head win matrix (row's decisive win-rate vs column)

|  | vs Fable | vs Opus | vs Sonnet 5 | vs Sonnet 4.6 |
|---|---|---|---|---|
| **Fable 5** | — | **0.67** | **0.80** | **0.89** |
| **Opus 4.8** | 0.33 | — | **0.74** | **0.78** |
| **Sonnet 5** | 0.20 | 0.26 | — | **0.58** |
| **Sonnet 4.6** | 0.11 | 0.22 | 0.42 | — |

Fable beats the field; Opus beats both Sonnets comfortably; Sonnet 5 edges
Sonnet 4.6.

## Robust wins (won both orderings of a brand)

| pair | winner tally | splits (position-bias) |
|---|---|---|
| Fable vs Opus | **Fable 24 – 7** | 19 |
| Fable vs Sonnet 5 | **Fable 33 – 3** | 14 |
| Fable vs Sonnet 4.6 | **Fable 40 – 1** | 9 |
| Opus vs Sonnet 5 | **Opus 26 – 2** | 22 |
| Opus vs Sonnet 4.6 | **Opus 31 – 3** | 16 |
| Sonnet 5 vs Sonnet 4.6 | **Sonnet 5 13 – 5** | 32 |

The Sonnet-5-vs-4.6 pair is the closest on the board: **32 of 50 brands were
position-bias splits**, so the two Sonnet generations are hard to separate — where
the judge had an order-independent preference, it favoured Sonnet 5 about 2.6:1.

## Why (from the judge's own words)

The gap is not about polish — it's about **staying inside the supplied facts** and
**finding genuinely separate beats.**

- **Fable's discipline (Airbnb).** *"A wins because it separates the story into
  real beats and keeps the proof gates honest"* — Fable's *"Expense software is
  becoming a travel distribution channel — Airbnb just plugged into Concur and
  Ramp"* found a clean adjacent B2B story where Sonnet 5 padded.
- **Fable's restraint (Mercury).** *"A wins because it keeps the story interesting
  without inflating the facts past recognition… 'Mercury wants to hold consumers'
  money without being a bank' is usefully uncomfortable and immediately names the
  proof required."*
- **Where the Sonnets lost points — invented specifics.** The judge repeatedly
  flagged ungrounded moves: Sonnet 5's *"'Airbnb wants corporate travel managers
  to stop chasing employees for receipts' invents a vivid scene"*; Sonnet 4.6's
  *"'Targeting the Market Navan Built' overstates the update"* and *"'Many Won't
  Qualify' is an invented conclusion. The update gives badge requirements, not
  host qualification rates."* This is exactly why both Sonnets score lowest on
  `grounding` (3.33 / 2.99).
- **Sonnet 5 > Sonnet 4.6 (Airbnb).** *"A wins because it keeps its hands on the
  evidence instead of grabbing shinier furniture from the wider market"* — Sonnet
  5's *"Airbnb calls business travel its fastest-growing segment. It won't say how
  big that segment actually is"* turns a caveat into a real reporter question.

So the trade is consistent with 2026-06-09: **Fable is the most generative and the
most disciplined at once; Opus is close behind and marginally more conservative;
the Sonnets are more prone to reaching past the facts, and Sonnet 5's main gain
over 4.6 is tighter grounding and less slop.**

## Caveats

- **One judge model (GPT-5.5).** A second judge or a human PR panel would harden
  the conclusion against single-judge idiosyncrasy.
- **Mixed generation harness.** `opus`/`fable` sets came from the original
  Workflow-subagent apparatus (2026-06-09); the two Sonnet sets came from headless
  `claude -p` with the same generator role + `angle-generator` skill + ETHICS +
  WHY-NOT-SPAM inlined and all tools disabled. The dominant variable is the model,
  and the ~0.5-point overall gap is far larger than a harness difference would
  plausibly create — but the harness is not identical (Fable cannot be
  regenerated), so the Sonnet numbers should be read as "this generation of Sonnet,
  run this way," not a perfectly controlled model-only delta.
- **Position bias.** The judge favoured slot A in **0.687** of decisive judgments;
  both orderings cancel it in aggregation, but near-ties (notably Sonnet 5 vs 4.6,
  32 splits) still resolve to order.
- **Constructed scenarios.** These measure angle *craft* on controlled facts, not
  live newsjacking.
- **Apparatus fixed, not tuned.** The skills are frozen; this measures the models
  and does not edit any output or cherry-pick brands. Skill SHA-256s are in
  `MANIFEST.json`.

## Reproduce

```bash
cd eval/fable-vs-opus
bash scripts/run-4model.sh runs/2026-06-30-4model all 8      # generate + judge (resumable)
python3 scripts/collect.py runs/2026-06-30-4model            # rebuild results.json
python3 aggregate-nmodel.py runs/2026-06-30-4model/results.json \
  --summary runs/2026-06-30-4model/summary.json
python3 make_figures.py runs/2026-06-30-4model/summary.json
cd ../design-system/scripts && node validate.mjs \
  ../../fable-vs-opus/runs/2026-06-30-4model/figures/four-model.html \
  --out ../../fable-vs-opus/runs/2026-06-30-4model/figures/png
```
