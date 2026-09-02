# Fable 5.1 vs the field — 2026-09-01 (50 brands, 6-model round-robin)

## Setup

Same fixed 50-brand story-angle benchmark as every prior run. Fable 5.1
launched today, so it is the only fresh contestant; the other five angle sets
and all of their existing judgments are reused byte-for-byte.

| slug | model | angle-set source |
|---|---|---|
| `f51` | **Fable 5.1** | generated 2026-09-01 through headless Claude Code CLI, exact id `claude-fable-5-1`, tools off, skill inlined |
| `o5` | Opus 5 | reused from `2026-07-24-opus5` |
| `fable` | Fable 5 | reused from `2026-06-09-full` |
| `opus` | Opus 4.8 | reused from `2026-06-09-full` |
| `s5` | Sonnet 5 | reused from `2026-06-30-4model` |
| `s46` | Sonnet 4.6 | reused from `2026-06-30-4model` |

Judging is the unchanged independent **GPT-5.5 `meanest-editor` judge** via
`codex exec`, blind, both orderings, schema-validated. The round-robin is
complete: C(6,2) = 15 pairs × 50 brands × 2 orderings = **1500 judgments**.

- 800 reused: the three pairs among Opus 5 / Fable 5 / Opus 4.8 (from
  `2026-07-24-opus5`) and the five pairs involving a Sonnet (from
  `2026-06-30-4model`).
- 700 fresh: the five Fable 5.1 pairs (500) plus Opus 5 vs each Sonnet (200),
  which no earlier run had crossed.

**Apparatus note.** `skills/ETHICS.md` was edited on 2026-08-03. Fable 5.1 was
generated against the frozen 2026-05-19 version every prior run used
(`apparatus/ETHICS.md`, sha `897f4fd0…`, via the new `ETHICS_FILE` override in
`generate.sh`). The angle-generator and meanest-editor skill hashes are
unchanged. Every case kept its fixed time, `2026-06-09T16:00:00Z`.

## Headline

**Fable 5.1 finishes first, and unlike the Opus-5-vs-Fable-5 result in July,
this one is not close.**

**Fable 5.1 › Opus 5 › Fable 5 › Opus 4.8 › Sonnet 5 › Sonnet 4.6**

1. **Fable 5.1 beats Opus 5 decisively.** 19–4 on order-independent robust
   brands (two-sided exact sign test `p=0.0026`), 65–34 across orderings, and
   it leads the direct pair by +0.79 on `grounding` while conceding only
   `proof_rigor` (−0.12) and `journalist_shape` (−0.01).
2. **Fable 5.1 is a full generation over Fable 5.** 38–1 robust
   (`p<0.001`), 87–13 across orderings, up on every one of the seven
   dimensions in the direct pair (from +0.37 to +0.65).
3. **The judge would ship 451 of its 500 sets.** That is the highest
   `publishable` rate in the study by a wide margin (Fable 5: 372, Opus 5: 312).
4. **It never lost a robust brand to Opus 4.8, Sonnet 5, or Sonnet 4.6.**

The July trade-off ("Opus 5 is sharper, Fable is safer") mostly collapses:
Fable 5.1 keeps Fable's grounding discipline and adds most of Opus 5's
angle-finding and proof pressure. The one thing Opus 5 still does slightly
better is turn a missing fact into a headline accountability angle.

## Balanced round-robin means

Each model appears in 500 judgments across its five opponents.

| dimension | Fable 5.1 | Opus 5 | Fable 5 | Opus 4.8 | Sonnet 5 | Sonnet 4.6 |
|---|---:|---:|---:|---:|---:|---:|
| news value | **4.46** | 4.31 | 4.31 | 4.08 | 3.92 | 3.89 |
| distinctness | **4.82** | 4.71 | 4.64 | 4.36 | 4.08 | 4.12 |
| journalist shape | **4.91** | 4.90 | 4.76 | 4.56 | 4.32 | 4.34 |
| grounding | **4.17** | 3.50 | 3.88 | 3.74 | 3.26 | 2.91 |
| anti-slop | **4.91** | 4.61 | 4.71 | 4.53 | 4.15 | 3.90 |
| proof rigor | 4.93 | **4.95** | 4.69 | 4.72 | 4.31 | 4.41 |
| usefulness | **4.81** | 4.68 | 4.58 | 4.44 | 4.08 | 4.06 |
| **overall** | **4.72** | **4.52** | **4.51** | **4.35** | **4.02** | **3.95** |

Fable 5.1 leads six of seven dimensions; Opus 5 keeps `proof_rigor` by 0.02.
Grounding is where the new model separates from everyone: 4.17 against a
previous best of 3.88, and the single-largest jump over Fable 5 (+0.29).

Meanest-editor `publishable` / `workshopable`, of 500 each:
**Fable 5.1** 451/49 · **Fable 5** 372/128 · **Opus 5** 312/188 ·
**Opus 4.8** 299/201 · **Sonnet 5** 161/339 · **Sonnet 4.6** 105/395.

## Head-to-head win matrix (row's decisive win-rate vs column)

|  | vs Fable 5.1 | vs Opus 5 | vs Fable 5 | vs Opus 4.8 | vs Sonnet 5 | vs Sonnet 4.6 |
|---|---:|---:|---:|---:|---:|---:|
| **Fable 5.1** | — | **0.66** | **0.87** | **0.90** | **0.98** | **0.94** |
| **Opus 5** | 0.34 | — | **0.56** | **0.66** | **0.86** | **0.89** |
| **Fable 5** | 0.13 | 0.44 | — | **0.67** | **0.80** | **0.89** |
| **Opus 4.8** | 0.10 | 0.34 | 0.33 | — | **0.74** | **0.78** |
| **Sonnet 5** | 0.02 | 0.14 | 0.20 | 0.26 | — | **0.58** |
| **Sonnet 4.6** | 0.06 | 0.11 | 0.11 | 0.22 | 0.42 | — |

## Robust wins (same model won both orderings)

| pair | robust | split | sign test |
|---|---:|---:|---:|
| Fable 5.1 vs Opus 5 | **19–4** | 27 | p=0.0026 |
| Fable 5.1 vs Fable 5 | **38–1** | 11 | p<0.001 |
| Fable 5.1 vs Opus 4.8 | **40–0** | 10 | p<0.001 |
| Fable 5.1 vs Sonnet 5 | **48–0** | 2 | p<0.001 |
| Fable 5.1 vs Sonnet 4.6 | **44–0** | 6 | p<0.001 |
| Opus 5 vs Sonnet 5 (new) | **36–0** | 14 | p<0.001 |
| Opus 5 vs Sonnet 4.6 (new) | **39–0** | 11 | p<0.001 |

The reused pairs reproduce their earlier results exactly (Opus 5 vs Fable 5
15–9, Fable 5 vs Opus 4.8 24–7, Sonnet 5 vs Sonnet 4.6 13–5), because they are
the same files.

## Direct pair: Fable 5.1 vs Opus 5

| dimension | Fable 5.1 | Opus 5 | Δ |
|---|---:|---:|---:|
| news value | 4.27 | 4.15 | +0.12 |
| distinctness | 4.62 | 4.50 | +0.12 |
| journalist shape | 4.74 | 4.75 | −0.01 |
| grounding | 4.13 | 3.34 | **+0.79** |
| anti-slop | 4.90 | 4.48 | +0.42 |
| proof rigor | 4.73 | 4.85 | −0.12 |
| usefulness | 4.58 | 4.43 | +0.15 |
| **overall** | **4.57** | **4.36** | **+0.21** |

`publishable`: Fable 5.1 76/100, Opus 5 44/100. Slot A won 74% of this pair's
decisive judgments, which is why 27 brands split; the robust 19–4 is the
cleaner read.

## Direct pair: Fable 5.1 vs Fable 5

| dimension | Fable 5.1 | Fable 5 | Δ |
|---|---:|---:|---:|
| news value | 4.49 | 4.12 | +0.37 |
| distinctness | 4.81 | 4.40 | +0.41 |
| journalist shape | 4.97 | 4.60 | +0.37 |
| grounding | 4.09 | 3.56 | +0.53 |
| anti-slop | 4.93 | 4.47 | +0.46 |
| proof rigor | 4.98 | 4.41 | +0.57 |
| usefulness | 4.87 | 4.22 | **+0.65** |
| **overall** | **4.73** | **4.25** | **+0.48** |

`publishable`: Fable 5.1 92/100, Fable 5 56/100.

## Why Fable 5.1 wins

The pattern in the rationales is that it keeps Opus 5's ambition without
Opus 5's habit of spending facts it does not have.

- **Duolingo.** Opus 5 put "its cartoon owl" in the headline; the update names
  Lily. Fable 5.1's "Duolingo says beta users talked to its AI character for
  nine minutes at a stretch" is anchored to the supplied metric. Judge: "cute
  wrongness is still wrongness, and journalists do not grade on charm."
- **Going.** Opus 5 invented a "$200 mistake fare" and implied in-app booking
  the update rules out. Fable 5.1's "Going's new AI planner lets the airfare
  pick the destination" is "a real inversion from the update."
- **Whatnot.** Opus 5 produced "worth less than one times revenue it doesn't
  keep," which the judge called bad math that confuses GMV with revenue.
  Fable 5.1 refused the "live shopping isn't dead in America" premise the
  update never supplied and led with "Whatnot's sellers moved $3B in a year.
  What they kept is still a secret."
- **Calm.** Fable 5.1 added a consumer-service angle Opus 5 missed and kept
  four protagonist-distinct lanes; Opus 5 asserted "demand for care exceeds
  clinician supply" after admitting the fact was not supplied.

## Where Opus 5 still wins

All four Opus 5 robust wins (Substack, Databricks, Sweetgreen, Anker) share one
shape: Opus 5 turned the weakest fact into the story, while Fable 5.1 left it
in the proof checklist.

- **Databricks.** Opus 5: "Databricks graded its own model, and that is the
  part enterprises should check." Fable 5.1 kept the self-reported benchmark
  as a caveat, not an angle.
- **Substack.** Opus 5: "…split more than $400M last year — and won't say how
  the split works." Fable 5.1 kept a case-study angle that needed a named
  writer the update never supplied.
- **Sweetgreen.** Opus 5 found the capital-allocation story ("betting on
  machines before it has proved the business"); Fable 5.1 refused a retrofit
  operations angle as a duplicate too quickly.

The single Fable 5 robust win (Mercury) is a grounding slip: Fable 5.1 called
Mercury "a consumer bank" and "the bank startups fled to after SVB" when the
update says it uses partner banks and holds no charter.

## Caveats

- **Position bias remains severe:** slot A won 66.0% of decisive judgments
  overall and 74% in the Fable 5.1 vs Opus 5 pair. Both orderings cancel
  identity bias; robust wins are the headline for a reason.
- **One judge model.** GPT-5.5 is independent of every contestant but is still
  one evaluator.
- **Mixed generation mechanism.** Fable 5, Opus 4.8 came from the original
  Workflow-subagent apparatus; Opus 5, both Sonnets, and Fable 5.1 came from
  `claude -p` with the same role, skill, ethics text, input, and tools off.
- **Constructed cases** measure angle craft on fixed facts, not live
  newsjacking.
- **One judge call hung** (Deel, Fable 5 vs Fable 5.1, ordering 1) and was
  killed and rerun; no output was edited, no case dropped, no rubric tuned.

## Share graphics

[`figures/share/index.html`](figures/share/index.html) holds four 1600×900
cards; 2× PNGs in [`figures/share/png/`](figures/share/png/). The full
six-model figure page is [`figures/six-model.html`](figures/six-model.html).

## Reproduce

```bash
cd eval/fable-vs-opus
bash scripts/run-fable51.sh runs/2026-09-01-fable51 all 10
python3 scripts/collect.py runs/2026-09-01-fable51
python3 aggregate-nmodel.py runs/2026-09-01-fable51/results.json \
  --summary runs/2026-09-01-fable51/summary.json
python3 make_figures.py runs/2026-09-01-fable51/summary.json \
  --out runs/2026-09-01-fable51/figures/six-model.html
python3 make_fable51_share_graphics.py runs/2026-09-01-fable51
cd ../design-system/scripts
for f in ../../fable-vs-opus/runs/2026-09-01-fable51/figures/share/0*.html; do
  node validate.mjs "$f" --width 1600 --out ../../fable-vs-opus/runs/2026-09-01-fable51/figures/share/png
done
```
