# Fable 5 vs Opus 4.8 — story-angle quality study

A blind, counterbalanced model comparison. Both **Claude Opus 4.8** and
**Claude Fable 5** generate story angles from the *same* company update using the
repo's `angle-generator` skill. A third, independent model — **GPT-5.5, driven
through `codex exec`** and grounded in the `meanest-editor` skill — judges each
pair blind. We measure which model's angles a veteran PR editor finds stronger,
and on which dimensions.

This is designed to be published as a study, so the method is deliberately
conservative about bias.

## Why this shape

- **Two Claude models, one neutral judge.** Using Opus or Fable to grade Opus-vs-Fable
  invites self-preference. The judge is a *different* vendor's model (GPT-5.5),
  so neither contestant judges itself.
- **The skills are the apparatus, not the contestant.** Both models run the exact
  same `angle-generator` skill on the exact same input; the only variable is the
  model. The judge runs the exact same `meanest-editor` rubric for every pair.
- **Everything is a clean-context subagent.** Each generator is an isolated
  subagent that sees only the brand fact + the skill — never the other model's
  output, never the rubric, never that it's being compared. Each judge call is a
  fresh `codex exec` GPT-5.5 session that sees only two anonymized angle sets.
- **Blind + counterbalanced.** The judge never learns which set is which model,
  and every pair is judged in **both orderings** (A=Opus/B=Fable and the swap).
  Position bias is then averaged out in aggregation. (In the full run it was
  severe — slot A won 69% of decisive calls — which is exactly why this matters.)

## Inputs (`brands.json`)

Ten real, well-known brands spanning sector (consumer CPG, enterprise software,
travel, fintech, dev tools, edtech, hardware, climate) and size (a 14-person
seed startup → Airbnb). Each is paired with a **constructed but plausible**
company update. These are eval scenarios, *not* claims of real events — the
update text is the only input the generator sees, and both models get it
verbatim along with a fixed `current_time`. Brand #10 (Verdano) is a deliberate
thin-facts stress test.

## Harness (`harness/`)

- `generator.md` — blind generator role: read `angle-generator/SKILL.md` + the
  brand fact + current time, produce the angle list. Never sees the rubric, the
  judge, or the other model.
- `judge.md` — blind judge role: adopt the `meanest-editor` persona/rubric (the
  full skill is appended at runtime by `judge.sh`), score two anonymized angle
  sets A and B on 7 dimensions, pick a winner, list concrete gaps.
- `judge-schema.json` — JSON Schema enforced on the judge's output via
  `codex exec --output-schema`.

### The 7 judge dimensions (1–5)

`news_value` · `distinctness` · `journalist_shape` · `grounding` · `anti_slop` ·
`proof_rigor` · `usefulness`. Each set also gets a one-word meanest-editor
verdict (`publishable` / `workshopable` / `start-over`).

## The Codex judge (`scripts/judge.sh`)

`codex exec` is the non-interactive Codex driver (Codex CLI ≥ 0.138, logged in via
ChatGPT). `judge.sh` assembles a fully self-contained prompt — judge instructions
+ the `meanest-editor` skill + the company update + both angle sets — and runs:

```
codex exec --skip-git-repo-check --sandbox read-only --model gpt-5.5 \
  --output-schema harness/judge-schema.json \
  --output-last-message <out.json> -
```

It needs no approvals (read-only sandbox, no tool calls) and writes
schema-validated JSON to the output file. Usage:

```bash
scripts/judge.sh UPDATE_FILE A_FILE B_FILE OUT_FILE
```

The caller owns the A/B→model mapping and records it next to the verdict; the
script itself is blind.

## Running it

The workflow `scripts/run.js` (run via the Workflow tool) does the whole loop —
load → write `update.txt` → generate both models in parallel → judge both
orderings via `codex exec` — every step a clean-context subagent. Invoke with:

```
Workflow({ scriptPath: "eval/fable-vs-opus/scripts/run.js",
           args: { run: "2026-06-13" } })          // all brands
Workflow({ scriptPath: "eval/fable-vs-opus/scripts/run.js",
           args: { run: "2026-06-13", brand_ids: [1] } })   // a subset
```

It returns the full results object; write it to `runs/<run>/results.json`, then:

```bash
python3 aggregate.py runs/<run>/results.json
```

## Aggregation (`aggregate.py`)

Re-anchors every judgment from slot (A/B) to model identity so position bias
cancels, then reports head-to-head win/tie/loss, per-dimension means and the
**Fable − Opus** delta, the meanest-editor verdict distribution, and a
position-bias diagnostic (how often slot A won regardless of model — 0.50 is
unbiased).

## Results

### `runs/2026-06-09-full/` — 50 brands, 100 judgments (headline study)

Full writeup: [`runs/2026-06-09-full/verdict.md`](runs/2026-06-09-full/verdict.md).

**Fable 5 produced the stronger angle set.** Re-anchored to model identity it won
**67/100 judgments** (Opus 33, 0 ties) and led **6 of 7 dimensions**
(`proof_rigor` tied). Past the judge's position bias (slot A won 69% of decisive
calls), the position-independent view is cleaner: **Fable won both orderings on
24/50 brands, Opus on 7/50**, and all 19 split brands were textbook position bias
— so where the judge had a genuine order-independent preference, **Fable won
≈77%**. Fable's edge was finding an extra distinct, sharper-shaped angle (and even
better headline-level hallucination discipline); Opus's counter-signal was
evidentiary restraint, which earned its tied `proof_rigor`.

| | Fable 5 | Opus 4.8 | Δ |
|---|---|---|---|
| overall dim mean | **4.60** | 4.36 | +0.24 |
| robust wins (both orderings) | **24/50** | 7/50 | |
| meanest-editor `publishable` | 78/100 | 54/100 | |

### `runs/2026-06-30-4model/` — 4-model round-robin (Sonnet 5 added)

Extends the study to **Opus 4.8 · Fable 5 · Sonnet 5 · Sonnet 4.6** — a full
C(4,2)=6-pair round-robin, 600 judgments (the opus↔fable pair reused; the 5
Sonnet pairs judged fresh). Uses `scripts/generate.sh` (headless `claude -p` by
exact model id, since the Workflow model alias can't address a specific prior
Sonnet), `scripts/run-4model.sh` (round-robin driver), `aggregate-nmodel.py`
(N-model per-dim means + win matrix + robust wins), and `make_figures.py`
(grouped bars + win matrix + heatmap + big-stats, one command from `summary.json`).

**Result: Fable 5 › Opus 4.8 › Sonnet 5 › Sonnet 4.6** (overall 4.65 / 4.46 /
4.13 / 4.04). Fable led all 7 dimensions and won every head-to-head; Sonnet 5
beat Sonnet 4.6 (13–5 robust) but both Sonnets trailed on `grounding`. The reused
opus↔fable pair reproduced 2026-06-09 exactly (24–7 robust; position bias 0.687).
There is **no accessible Sonnet 4.7** (retired); 4.6 is the latest pre-5 Sonnet.
Full writeup: [`runs/2026-06-30-4model/verdict.md`](runs/2026-06-30-4model/verdict.md).

### `runs/2026-07-24-opus5/` — Opus 5 added

Reuses the frozen Fable 5 and Opus 4.8 outputs and their 100 existing judgments,
generates only **Opus 5** through headless Claude Code CLI with exact model ID
`claude-opus-5`, then judges the two new pairs in both orderings.

**Result: Opus 5 ≈ Fable 5 > Opus 4.8.** Opus 5 ranks first on balanced
round-robin mean (4.52 vs Fable 4.47 and Opus 4.8 4.30), beats Fable 15–9 on
robust brands, and beats Opus 4.8 20–4. The Opus-5 edge over Fable is not
decisive, however: 26/50 brands split by position, Fable leads grounding by
0.47, and Fable earns more `publishable` ratings (145/200 vs 117/200). Full writeup:
[`runs/2026-07-24-opus5/verdict.md`](runs/2026-07-24-opus5/verdict.md).

## Discipline

This study **measures the models; it does not tune the skill to move a number.**
The `angle-generator` and `meanest-editor` skills are fixed apparatus. If a
result looks wrong, fix the *measurement* (more brands, better counterbalancing,
a sharper judge prompt) — never edit a model's output or cherry-pick brands.
Scale to enough brands (the plan is 50) that the position-bias-averaged
head-to-head and dimension deltas are stable before drawing any conclusion.
