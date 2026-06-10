# Full run verdict — 2026-06-09 (50 brands)

**Setup.** 50 brands across sector and size. Each fed to **Opus 4.8** and **Fable
5** as clean-context subagents running the `angle-generator` skill on identical
input. Each pair judged blind by **GPT-5.5** (via `codex exec`, `meanest-editor`
persona) in **both orderings** → 100 judgments. Aggregation re-anchors every
judgment to model identity so position bias cancels.

## Headline

**Fable 5 produced the stronger angle set.** Re-anchored to model identity it
won **67 of 100 judgments (Opus 33, 0 ties)** and led **6 of 7 quality
dimensions**; `proof_rigor` was a dead tie.

But the raw head-to-head understates the picture because the judge has a real
position bias (it favored whichever set it saw first in **69%** of decisive
judgments). The honest read is the **position-independent** view — who won in
*both* orderings of the same brand:

| Outcome (per brand, both orderings) | Brands |
|---|---|
| **Fable won both** (robust win) | **24 / 50** |
| **Opus won both** (robust win) | **7 / 50** |
| Split (judge defaulted to position) | 19 / 50 |

**All 19 split brands are textbook position bias** — slot A won both times — i.e.
on those brands the two sets were close enough that the judge fell back on order.
So where the judge had a genuine, order-independent preference, **Fable won 24 to
7 (≈77%)**. This is why both orderings are mandatory: a single-ordering run would
have reported a contaminated number.

## Per-dimension (mean 1–5, Fable − Opus)

| dimension | Fable 5 | Opus 4.8 | Δ |
|---|---|---|---|
| news_value | 4.41 | 4.08 | +0.33 |
| distinctness | 4.78 | 4.37 | +0.41 |
| journalist_shape | 4.90 | 4.62 | +0.28 |
| grounding | 3.90 | 3.65 | +0.25 |
| anti_slop | 4.73 | 4.51 | +0.22 |
| proof_rigor | 4.79 | 4.80 | −0.01 |
| usefulness | 4.69 | 4.49 | +0.20 |
| **overall** | **4.60** | **4.36** | **+0.24** |

Meanest-editor verdicts: **Fable** `publishable` 78 / `workshopable` 22 · **Opus**
`publishable` 54 / `workshopable` 46.

## Why Fable won (from the judge's own words)

The wins cluster on **finding an extra genuinely distinct angle and shaping it
sharper**, and — notably — on **better hallucination discipline at the headline**:

- *Canva:* Fable **refused** "Canva's recent AI acquisition is already powering
  its enterprise suite," which Opus ran with — "the difference between an angle
  and an invented connection."
- *Duolingo:* Fable's "Educators call gamified learning shallow. Duolingo's
  answer is a talking cartoon" vs Opus's "more generic markets mush."
- *Discord:* Fable separated the human server-operator story ("about to find out
  what their communities are worth") as its own reporting path; Opus left it at
  trend altitude.
- *Reddit:* Fable's "profit story now leans on $200M of data deals investors
  can't fully see" vs Opus overstating that the deals "rival ads."

## Where Opus won (the genuine counter-signal)

Opus's 7 robust wins (Airbnb, Oatly, Peloton, Substack, Allbirds, Deel, Replit)
and its tied `proof_rigor` come from one consistent strength: **evidentiary
restraint.** The judge repeatedly praised Opus for *not* pretending missing facts
exist —

- *Ramp:* "'there is no single breaking news peg supplied here' is the kind of
  restraint that keeps a founder out of fake-peg trouble"; it dinged Fable for
  turning one launch into an unsupported "category thesis."
- *Klarna / Hims & Hers / Replit:* Opus flagged the exact hole a reporter would
  probe; in these cases Fable's "stronger headline energy" went "past the
  supplied facts — how a decent angle becomes a correction waiting to happen."

So the trade is real and narrow: **Fable is more generative, more distinct, and
sharper-shaped; Opus is marginally more conservative.** Fable's edge on
`grounding` (+0.25) shows its extra reach usually stayed inside the facts, but
Opus's tied `proof_rigor` is where its caution earned its keep.

## Caveats

- One judge model (GPT-5.5). A second judge (or a human PR panel) would harden
  the conclusion against single-judge idiosyncrasy.
- Position bias is strong (0.69); we cancel it by counterbalancing, but it makes
  ~38% of brands too close to separate.
- Scenarios are constructed (not live news), so this measures angle *craft* on
  controlled facts, not real-time newsjacking.

## Reproduce

```bash
python3 scripts/collect.py runs/2026-06-09-full      # rebuild results.json from disk
python3 aggregate.py runs/2026-06-09-full/results.json
```
