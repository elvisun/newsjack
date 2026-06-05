# newsworthiness-check benchmark — 2026-06-04

Three models × {with skill, without skill} over the 16-case v2 dataset.
The skill was **not modified.** With-skill runs are graded deterministically by
`grade.py`; the gold reference for cross-model deltas is **Opus + skill**, which
passes every committed assertion (67/67).

Models (full IDs):
- **claude-opus-4-8**
- **claude-sonnet-4-6**
- **claude-haiku-4-5-20251001**

## Matrix

| Model | Cond | Cases pass | Assertions | Schema complete | mean \|Δscore\| vs gold | over-rated (≥+2) | decision-class mismatch |
|-------|------|-----------|-----------|-----------------|----------------------|------------------|-------------------------|
| claude-opus-4-8 | **+skill** | **16/16** | **67/67** | 16/16 | 0.00 | 0 | 0 |
| claude-opus-4-8 | −skill | — | — | — | 0.93 | 1 | 0 |
| claude-sonnet-4-6 | **+skill** | **16/16** | **67/67** | 16/16 | 0.27 | 0 | 0 |
| claude-sonnet-4-6 | −skill | — | — | — | 1.53 | 5 | 1 |
| claude-haiku-4-5-20251001 | **+skill** | 11/16 | 62/67 | **9/16** | 0.13 | 0 | 1 |
| claude-haiku-4-5-20251001 | −skill | — | — | — | 1.53 | 6 | 2 |

- *mean |Δscore| vs gold* — average absolute difference from Opus+skill's score, over scorable cases (case 9 is null/AVOID and excluded).
- *over-rated (≥+2)* — cases scored at least 2 points above gold (the inflation the skill exists to stop).
- *decision-class mismatch* — action disagrees with gold after mapping {PROCEED,RIDE→GO; REVISE,WAIT→MAYBE; HOLD,SKIP,AVOID→NO}.
- Baselines emit only score+free-text action (no assertion schema), so they have no Cases/Assertions grade; they are measured by the score/decision deltas.

## What it shows

1. **The skill works, and it works across the model tier.** With the skill, all
   three models converge tightly on the same calibrated answer (mean |Δscore|
   ≤ 0.27, **zero** over-rated cases). Without it, scoring error is **5-12×
   higher** (0.93-1.53) and each bare model over-rates 1-6 cases.

2. **The over-rating bias is real and consistent — strongest on launches and
   positives.** The clearest example is the real NVIDIA "Vera" launch (case 5):
   bare models score it **9 / 10 / 9** ("full press push", "pitch immediately"),
   while every model *with* the skill lands a disciplined **7-8 / PROCEED**.
   Positives generally drift to 7-8 "pitch aggressively" without the skill; the
   skill returns standing-capped RIDE/WAIT with narrow-angle guidance.

3. **Bigger model ≠ calibrated model.** Opus's bare baseline is better than the
   smaller models' (0.93 vs 1.53) but still over-rates and lacks the skill's
   standing/restraint structure. The skill closes that gap for everyone — and
   notably pulls **Haiku+skill (0.13)** to within a rounding error of gold on the
   judgments it does emit.

4. **Schema reliability is the one place model tier matters with the skill on.**
   Opus and Sonnet emit all 16 complete outputs. **Haiku truncates 7 of 16**
   (cases 4, 11-16), dropping required fields (`honest_assessment`,
   `kill_switch_triggered`, `evidence_gaps`) — its *judgments* are right (0.13
   delta) but the *output contract* breaks. Since the detector pipeline consumes
   those fields, prefer Sonnet/Opus for this skill, or add a schema-completeness
   gate before trusting a Haiku worker's output.

## Per-case scores

`H/S/O` = Haiku/Sonnet/Opus **with skill**; `h/s/o` = same models **without skill**.
Gold = Opus+skill.

```
id industry           gold   H  S  O  | h  s  o
 1 ai-saas (PH)        2       2  2  2  |  4  4  3
 2 ai-saas (PH)        2       2  2  2  |  5  3  2
 3 developer-tools(PH) 1       1  1  1  |  3  3  2
 4 fintech (balanced)  4       5  4  4  |  2  5  5
 5 semiconductors      7       7  8  7  |  9 10  9   <- NVIDIA launch: bare models inflate hard
 6 ai-enterprise       6       6  6  6  |  8  7  6
 7 food-service        3       3  4  3  |  1  1  1
 8 marketing-services  3       3  3  3  |  2  1  1
 9 food-cpg (killsw)   AVOID   AVOID    |  1  1  1   <- all 3 baselines also avoid (good)
10 ai-tech (stale)     4       5  4  4  |  3  2  2
11 pharma-biotech      7       7  7  7  |  8  8  8
12 food-safety         6       6  7  6  |  7  8  7
13 energy-datacenter   7       7  7  7  |  8  8  7
14 banking-finance     2       2  2  2  |  2  2  2
15 finance-macro       6       6  6  6  |  8  8  7
16 automotive-ev       6       6  7  6  |  7  7  6
```

Note the baselines are also **erratic the other way**: on the no-standing cases
(7, 8) they collapse to 1 (over-harsh), and on the nuanced fintech counterexample
(4) they scatter 2-5. The skill produces the calibrated, consistent middle in
both directions — that consistency, not any single number, is the product.

## Reproduce

```
# with-skill (per model): blind executor writes outputs, then
python3 grade.py runs/2026-06-04-benchmark/<model-id>/with-skill/outputs \
  --out runs/2026-06-04-benchmark/<model-id>/with-skill/grading.json

# without-skill baseline (per model): generic PR-assistant prompt over the
# answer-key-free runs/2026-06-04-benchmark/baseline-prompts.json
```

Layout: `runs/2026-06-04-benchmark/<full-model-id>/{with-skill/outputs+grading.json, without-skill/baseline.json}`,
plus the shared `baseline-prompts.json`.
