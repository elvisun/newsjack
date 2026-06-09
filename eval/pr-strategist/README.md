# PR Strategist Eval — blind pairwise vs expert strategies

Quality eval for the `skills/pr-strategist` skill. It measures whether the skill,
run by an independent agent, produces PR strategies that a veteran PR director
**cannot reliably tell apart from — or prefer over — real expert strategies**,
across a broad range of founder situations.

This is the launch-blocking skill: the plan is for founders to comment a startup
URL and get a PR strategy back, so the bar is "expert-team quality, every time,
across many situations" — not "passes a fixed checklist."

## Why pairwise, not a static rubric

`newsworthiness-check` and `onboarding` grade against per-case assertions. A PR
*strategy* has no single right answer — many plans are equally good, and the
failure mode we most fear is **overfitting** (memorizing one canned answer). So
the headline metric is a **blind A/B**: our skill's strategy vs an expert
reference for the same situation, judged by an LLM playing a veteran PR director
who does not know which is which. The rubric (`gold.json > rubric_dimensions`,
and per-case `must_haves`/`anti_patterns`) only *grounds* the judge; it never
requires a verbatim match, and PR-sound divergence gets full credit.

**Success = parity, not victory.** The goal (from the maintainer) is that the
judge cannot distinguish ours from the expert's and does not systematically
prefer the expert's — i.e. identification accuracy ≈ 0.5 and candidate
tie-or-better rate high, with **no systematic negative dimension delta**. A
held-out split confirms the gains generalize rather than overfit.

## The gold set (`gold.json`)

Nine cases spanning the skill's archetypes and many industries, split
`train` (6) / `holdout` (3). Each case has:

- `scenario` — the concrete founder message (the only input the executor sees).
- `reference_strategy` — the expert bar: a faithful distillation of what a named
  real source actually recommends for that situation (sources cited per case).
  It is *an* expert answer, not *the* answer.
- `must_haves` / `anti_patterns` — strategic substance a strong answer should hit
  / avoid, used to ground the judge (not for exact-match).
- `sources` — real, auditable URLs (First Round, Bessemer, a16z, Lulu Cheng
  Meservey/Flack, April Dunford, Triple Whale, markepear HN guide, command.ai
  Linear teardown, etc.).

Coverage: b2b-saas drumbeat · dev-tool/Show HN · consumer-DTC creator seeding ·
enterprise analyst-led · post-fundraise (bundle trap) · category challenger ·
solo founder go-direct · pre-product redirect · weak-peg data PR.

### Anti-overfit safeguards

1. The skill must **never name these companies/scenarios**; advice is judged on
   general strategic quality, not recall.
2. The **holdout split is never inspected during iteration** — only used to
   confirm gains generalize.
3. The judge rewards a broad range of sound plans and penalizes only real
   `anti_patterns`, so the skill can't win by parroting one template.
4. Two models generate candidates (Opus + Sonnet); the skill has to lift both.

## Harness (`harness/`)

- `executor.md` — blind executor: reads `skills/pr-strategist/SKILL.md` and the
  scenario *only*, returns the strategy a founder would receive. Never sees the
  reference, rubric, or grading.
- `judge.md` — blind judge: a veteran PR director scores A and B on 7 dimensions,
  picks a winner, guesses which is the AI, and lists concrete gaps. Judges
  substance only; ignores style/length.

Executor and judge are **separate agents** so the executor can't see the answer
key, and each pairing is judged in **both orderings** (candidate-first and
gold-first) to cancel position bias.

## Running it

The loop is orchestrated by a workflow (`scripts/run.js`, invoked via the
Workflow tool): for each train case × {Opus, Sonnet} it generates a candidate,
then judges candidate-vs-gold in both orderings, emitting one record per
judgment to `runs/<date>/results.json`. Then:

```bash
python3 aggregate.py runs/<date>/results.json
```

`aggregate.py` re-anchors every judgment to candidate-vs-gold and reports
win/tie/loss, dimension deltas, judge identification accuracy, position-bias
sanity, and the most common candidate gaps (the iteration signal).

## Iteration discipline

This eval **grades the skill; iteration edits the skill, never the gold to make
the number move.** Fixes must be *general* (a better gate, a sharper default, a
missing archetype move) — never "memorize case N." After each skill edit, re-run
train; only at the end, run holdout once to confirm no overfit. Findings and
per-run results live in `runs/<date>/`.

## Results

Latest: [`runs/2026-06-09-final/verdict.md`](runs/2026-06-09-final/verdict.md).

Against a hardened, expert-grade reference set, the skill reaches the goal —
**judge identification accuracy 0.457 (chance = 0.5)**, candidate win-or-tie
**69%**, **every quality dimension non-negative**, and the held-out split (never
used in iteration) generalizes. Opus+skill is at/above expert parity; Sonnet+skill
hovers around parity.

Run progression (see verdict for the full table):

| Run | Gold bar | Skill | W/T/L | Tie+ | Judge ID acc |
|---|---|---|---|---|---|
| baseline | v1 soft | v0 | 20/0/4 | 0.83 | 0.54 |
| iter1 | v1 soft | v1 | 35/0/1 | 0.97 | 0.17 (gold saturated) |
| hard-bar | v2 expert | v1 | 25/0/11 | 0.69 | 0.39 |
| **final** | v2 expert | v2 | 25/0/11 | 0.69 | **0.457** |

`runs/2026-06-09-baseline/` keeps the original soft gold (`gold-v1.json`) and its
results; `runs/2026-06-09-final/` keeps the expert gold (`gold-v2.json`),
per-record results, and the verdict. The six general skill changes that produced
the lift are listed in the verdict's "What changed in the skill" section.
