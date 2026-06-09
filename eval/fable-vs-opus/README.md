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
  Position bias is then averaged out in aggregation. (In the pilot it was severe
  — see Results — which is exactly why this matters.)

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

**Pilot (one brand, what's checked in here):** generators run as two subagents
(`model: opus`, `model: fable`); `judge.sh` runs twice (both orderings). See
`runs/2026-06-09-pilot/`.

**Scale (up to all 10, or 50+ once `brands.json` is extended):** the workflow
`scripts/run.js` (run via the Workflow tool) does the whole loop —
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

### `runs/2026-06-09-pilot/` — Notion, n=2 judgments (pilot)

Pilot purpose: validate the full pipeline (subagent generators + `codex exec`
GPT-5.5 judge + counterbalancing + aggregation) before scaling. It works
end-to-end. Headline numbers at n=2 are not conclusions — they size the apparatus.

- **Head-to-head: 1–1** after de-biasing (a tie on this brand).
- **Position bias was total: slot A won 2/2** (1.00 vs 0.50 unbiased). The judge
  picked whichever set was shown *first*, both times. This is the single most
  important pilot finding: a one-ordering design would have reported a spurious
  winner. Both orderings are mandatory.
- **Dimension signal (suggestive only):** Fable led `distinctness` (+1.50) — the
  judge, in both orderings, praised Fable's offline-editing angle as a genuinely
  distinct fourth press lane and dinged Opus for refusing offline as "just a
  feature note" and for three angles orbiting the same bank-pilot spine. Opus
  edged `proof_rigor` (+0.50). Both sets were judged `publishable` once and
  `workshopable` once.

Run `python3 aggregate.py runs/2026-06-09-pilot/results.json` to reproduce.

## Discipline

This study **measures the models; it does not tune the skill to move a number.**
The `angle-generator` and `meanest-editor` skills are fixed apparatus. If a
result looks wrong, fix the *measurement* (more brands, better counterbalancing,
a sharper judge prompt) — never edit a model's output or cherry-pick brands.
Scale to enough brands (the plan is 50) that the position-bias-averaged
head-to-head and dimension deltas are stable before drawing any conclusion.
