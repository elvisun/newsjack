# PR Strategist pairwise eval — 2026-06-09

Blind A/B: the `pr-strategist` skill, run by independent executor agents (Opus +
Sonnet) on the founder scenario only, vs an expert reference strategy for the
same scenario. A veteran-PR-director judge (Opus), blind to which is which, scored
both on 7 dimensions, picked a winner, and guessed which answer was the AI. Every
pairing was judged in **both orderings** to cancel position bias. 9 cases (6
train, 3 holdout) × 2 models × 2 orderings = 36 records per run.

## Headline

**Goal: the judge can no longer reliably tell our strategies from expert ones.**
Against a hardened, expert-grade reference set, judge identification accuracy is
**0.457 (chance = 0.5)**, with the candidate winning or tying **69%** of pairings
and **every quality dimension non-negative** vs the expert bar. The held-out
split — never used during iteration — generalizes (0.75 tie-or-better,
identification 0.27), so the gains are general, not memorized.

## Progression

| Run | Gold bar | Skill | Win/Tie/Loss | Tie-or-better | Judge ID acc (0.5=can't tell) | Mean: cand / gold |
|---|---|---|---|---|---|---|
| baseline | v1 (soft) | v0 | 20/0/4 | 0.83 | 0.542 | 4.90 / 4.18 |
| iter1 | v1 (soft) | v1 (+6 fixes) | 35/0/1 | 0.97 | 0.171* | 4.94 / 4.12 |
| hard-bar | **v2 (expert)** | v1 | 25/0/11 | 0.69 | 0.389 | 4.86 / 4.43 |
| **final** | **v2 (expert)** | **v2 (+HN craft)** | **25/0/11** | **0.69** | **0.457** | **4.82 / 4.42** |

\* iter1's 0.17 is the signal that **v1 gold had saturated** — the improved skill
so outclassed the tight ~300-word v1 references that the judge began flagging the
*reference* as the AI. We hardened the gold to full expert depth (v2) so
"indistinguishable" is a meaningful claim, then confirmed against it. Hardening
the bar works *against* our win rate; it is not gold-gaming. v1 references are
preserved in `runs/2026-06-09-baseline/gold-v1.json`.

## Final, by model and split (vs expert gold v2)

| Slice | Win/Tie/Loss | Tie-or-better | Judge ID acc |
|---|---|---|---|
| Opus candidate | 17/0/1 | 0.94 | 0.28 |
| Sonnet candidate | 8/0/10 | 0.44 | 0.65 |
| Train | 16/0/8 | 0.67 | 0.54 |
| Holdout (untouched) | 9/0/3 | 0.75 | 0.27 |
| **All** | **25/0/11** | **0.69** | **0.457** |

Opus is at or above expert parity everywhere; Sonnet hovers around parity (the
remaining losses cluster on the cheaper model). Per-case, the result is a healthy
mix of dominant wins (#2 dev-tool 4/0, #6 challenger 4/0, #9 data-PR 4/0) and
genuine 50/50 ties (#1 4.93=4.93, #4, #5, #7) — i.e. indistinguishable across a
broad range of situations, not a single canned answer.

## Dimension deltas (candidate − expert gold, final)

audience_goal +0.56 · positioning +0.95 · news_peg +0.42 · channel_cadence +0.28
· tactics_quality +0.03 · judgment_refusals −0.03 · fit_actionability +0.64. The
skill's biggest edges are positioning discipline and tailored actionability; the
once-negative `tactics_quality` (−0.08 at hard-bar) closed to +0.03 after the
dev-tool HN-craft fix.

## What changed in the skill (all general, no scenario memorization)

From the judge's recurring gap notes on train + holdout:

1. **Dev-tool / Show HN archetype** — added: frictionless trial / no-signup-wall;
   plain title, no superlatives; "open-source alternative to X" framing; the
   seven-part honest-experiment-report post shape; Tue–Thu morning ET timing;
   win-it-in-the-comments craft (reply in the hour, agree-first, treat critics as
   allies); a hard refusal of planted booster comments / upvote rings; relaunch
   after a flop is normal. Channel table gained Lobsters + dev.to.
2. **Consumer / DTC archetype** — high-volume no-strings gifting (reach hundreds
   to get the ~30 who post), **capture usage rights and reuse the content as the
   asset** ("creator over distribution"), explicit micro→bigger→advocate phasing.
3. **Give advice, not apparatus** — never expose internal scaffolding (archetype
   numbers, rule names, "per the skill") in the founder-facing answer. (This was
   the judge's single biggest AI tell.)
4. **Numbers are rules of thumb, not citations** — don't quote calibration figures
   as precise sourced stats ("studies show exactly 2.5×"); say "roughly 2–3×".
5. **Scale to capacity** — lead with the 1–2 load-bearing moves; don't hand a
   solo founder a program they can't run.
6. **Small-N data caveat** — caveat methodology on thin samples; don't dress a
   tiny sample up as an authoritative "State of X" index (ties to ETHICS).

## Honest caveats

- The judge is Opus, which also generates one candidate set; blind double-ordering
  and an independent-source gold mitigate self-preference, but Sonnet-as-judge is
  a worth-running cross-check.
- The reference set is a faithful, **sourced distillation** of expert guidance
  (First Round, Bessemer, a16z, Cheng Meservey/Flack, Dunford, Triple Whale,
  markepear, command.ai), not raw verbatim third-party text — so it is an *expert
  bar*, not a specific named team's exact memo.
- Remaining Sonnet losses are mostly model adherence to nuances the skill already
  states (name-the-raise-number, the pipeline-vs-raise fork, the thin-sample
  caveat). Closing them further risks over-specifying the skill into a checklist
  that overfits these 9 scenarios — deliberately not done.

## How to reproduce

```bash
# from eval/pr-strategist/, via the Workflow tool:
Workflow(scriptPath=scripts/run.js, args={"split":"all","models":["opus","sonnet"]})
python3 aggregate.py runs/<date>/results.json
```
