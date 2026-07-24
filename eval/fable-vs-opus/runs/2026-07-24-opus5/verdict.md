# Opus 5 vs Fable 5 vs Opus 4.8 — 2026-07-24

## Setup

This is the same fixed 50-brand story-angle benchmark as the original
Fable-5-vs-Opus-4.8 study.

| slug | model | angle-set source |
|---|---|---|
| `o5` | **Opus 5** | generated through headless Claude Code CLI with exact model ID `claude-opus-5`, tools off, skill inlined |
| `fable` | **Fable 5** | reused byte-for-byte from `2026-06-09-full` |
| `opus` | **Opus 4.8** | reused byte-for-byte from `2026-06-09-full` |

The old models were not rerun. Their 100 existing counterbalanced judgments were
also reused byte-for-byte. Only Opus 5 was generated, then the two new pairs were
judged fresh:

- Opus 5 vs Fable 5: 50 brands × 2 orderings = 100 judgments
- Opus 5 vs Opus 4.8: 50 brands × 2 orderings = 100 judgments
- Fable 5 vs Opus 4.8: 100 reused judgments

Total: **300 blind judgments** from the unchanged GPT-5.5 `meanest-editor`
judge. Every case retained its original fixed time,
`2026-06-09T16:00:00Z`. The skill and ethics-file hashes exactly match the prior
four-model study.

## Headline

**Opus 5 finishes first, but the honest result has two parts:**

1. **Opus 5 is a clear generational improvement over Opus 4.8.** It wins
   20–4 on order-independent robust brands and 66–34 across the two orderings.
2. **Opus 5 has a narrow, not decisive, edge over Fable 5.** It wins 15–9 on
   robust brands and 56–44 across orderings, but 26/50 brands split by position.
   Fable is substantially better grounded and receives more `publishable`
   ratings.

The practical ranking is:

**Opus 5 ≈ Fable 5 > Opus 4.8**

Use Opus 5 when you want the sharpest angle-finding and proof pressure. Use
Fable 5 when factual restraint matters more than an extra clever angle. Opus 5
still needs a fact-cleaning pass before its output goes near a journalist.

## Balanced round-robin means

Each model appears in 200 judgments across both opponents.

| dimension | Opus 5 | Fable 5 | Opus 4.8 |
|---|---:|---:|---:|
| news value | **4.36** | 4.28 | 4.02 |
| distinctness | **4.72** | 4.59 | 4.28 |
| journalist shape | **4.92** | 4.72 | 4.52 |
| grounding | 3.42 | **3.89** | 3.70 |
| anti-slop | 4.56 | **4.71** | 4.51 |
| proof rigor | **4.97** | 4.59 | 4.67 |
| usefulness | **4.71** | 4.53 | 4.38 |
| **overall** | **4.52** | **4.47** | **4.30** |

Opus 5 leads five of seven dimensions. Fable leads the two safety dimensions:
`grounding` and `anti_slop`.

## Head-to-head

### Judgment wins

| pair | result |
|---|---:|
| Opus 5 vs Fable 5 | **56–44** |
| Opus 5 vs Opus 4.8 | **66–34** |
| Fable 5 vs Opus 4.8 | **67–33** |

### Robust wins

A robust win means the same model won both A/B orderings for that brand.

| pair | robust result | position splits |
|---|---:|---:|
| Opus 5 vs Fable 5 | **15–9** | 26 |
| Opus 5 vs Opus 4.8 | **20–4** | 26 |
| Fable 5 vs Opus 4.8 | **24–7** | 19 |

The Fable comparison is close. Among the 24 brands with an order-independent
winner, Opus 5 takes 62.5%. An exploratory two-sided exact sign test gives
`p=0.307` with a 95% Wilson interval of 42.7%–78.8%, so this run does not
establish a decisive Opus-5 advantage over Fable.

The generational comparison is much firmer: Opus 5 takes 20/24
order-independent wins over Opus 4.8 (`p=0.0015`; 95% Wilson interval
64.2%–93.3%).

## The important contradiction: wins vs publishability

| model | `publishable` | `workshopable` |
|---|---:|---:|
| Fable 5 | **145/200** | 55/200 |
| Opus 5 | 117/200 | 83/200 |
| Opus 4.8 | 110/200 | 90/200 |

Opus 5 wins more pairwise decisions while Fable produces more sets the judge
would ship with only minor polish. That is not noise; it matches the dimension
scores. Opus 5 finds more ambitious, better-shaped material, but its factual
overreach forces more workshop passes.

In the direct Opus-5-vs-Fable pair, the contrast is even sharper:

| dimension | Opus 5 | Fable 5 | Opus 5 delta |
|---|---:|---:|---:|
| news value | 4.37 | 4.14 | +0.23 |
| distinctness | 4.65 | 4.41 | +0.24 |
| journalist shape | 4.89 | 4.54 | +0.35 |
| grounding | 3.38 | 3.88 | **−0.50** |
| anti-slop | 4.58 | 4.69 | −0.11 |
| proof rigor | 4.97 | 4.39 | **+0.58** |
| usefulness | 4.68 | 4.37 | +0.31 |
| **overall** | **4.50** | **4.35** | **+0.15** |

## Why Opus 5 wins

The recurring advantage is not polish. It is editorial pressure.

- **It turns missing evidence into the story.** On Replit, the judge preferred
  “Replit's new agent builds and ships working apps for people who can't code —
  and won't say how often they break.” Fable saw the same reliability hole but
  left it in the proof checklist.
- **It finds harder second-order tension.** On Plaid, “Plaid needs 12,000 banks
  to help route payments past the fees banks collect” gave the reporter a
  conflict to interrogate instead of another launch wrapper.
- **Its proof lists are unusually good.** On Sweetgreen, it demanded the sample,
  denominators, capex, staffing, confounders, and selective-disclosure risk
  needed to test the restaurant-margin claim.
- **Its reporter shapes are the most specific in the field.** This is the
  largest dimension lead: 4.92 overall.

## Why Fable still wins plenty

Opus 5 repeatedly spends facts it does not have.

- **Invented specifics:** Going's “$200 mistake fare,” despite no price in the
  update.
- **Wrong central fact:** calling Duolingo's Lily “the cartoon owl.”
- **Unsupported history:** Allbirds “spent a decade avoiding wholesale.”
- **Invented causal link:** Canva “bought an AI image startup, then shipped it as
  a brand-safety feature.”
- **Market theses promoted to facts:** Mercury entering “the market that broke
  the unprofitable ones.”
- **Keeping a dead angle after diagnosing it:** Wise says the customer angle
  does not exist without a named customer, then keeps it anyway.

All four brands where Opus 4.8 robustly beat Opus 5—Brex, Allbirds, Wise, and
Mercury—are also robust Fable wins over Opus 5. That overlap is the cleanest
evidence that grounding is the repeatable Opus-5 failure mode, not judge noise.

## Position bias and caveats

- **Position bias is severe:** slot A wins 221/300 decisive judgments
  (**73.7%**). Both orderings remove model-identity bias from the tallies, but
  26/50 brands split in each new Opus-5 pairing. Robust wins are the cleaner
  headline.
- **One judge model:** GPT-5.5 is independent of the contestants, but it is still
  one evaluator. A second vendor judge or human PR panel would strengthen the
  Fable-vs-Opus-5 call.
- **Mixed generation mechanism:** the reused Fable/Opus-4.8 outputs came from the
  original Workflow-subagent apparatus; Opus 5 came from `claude -p` with the
  same generator role, skill, ethics files, input, and tools disabled. This is
  disclosed rather than waved away.
- **Constructed cases:** this measures angle craft on fixed facts, not live
  newsjacking or long-horizon agent work.
- **The skills stayed frozen:** no output was edited, no cases were dropped, and
  no rubric was tuned after seeing results.

## Share graphics

The [share-card gallery](figures/share/index.html) contains four standalone
1600×900 HTML graphics. Matching 2× PNG exports are in
[`figures/share/png/`](figures/share/png/).

## Reproduce

```bash
cd eval/fable-vs-opus
bash scripts/run-opus5.sh runs/2026-07-24-opus5 all 8
python3 scripts/collect.py runs/2026-07-24-opus5
python3 aggregate-nmodel.py runs/2026-07-24-opus5/results.json \
  --summary runs/2026-07-24-opus5/summary.json
python3 make_figures.py runs/2026-07-24-opus5/summary.json \
  --out runs/2026-07-24-opus5/figures/three-model.html
python3 make_opus5_share_graphics.py runs/2026-07-24-opus5
```
