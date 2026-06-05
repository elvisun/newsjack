# Newsworthiness Check Eval

Cross-industry calibration eval for the `skills/newsworthiness-check` skill. It
checks whether following the skill produces the harsh, anchor-disciplined scores
the skill claims to — especially that it does **not** inflate ordinary product
launches, routine corporate PR, or topical-but-unrelated events into "news".

This eval **grades the skill, it does not change it.** If a case fails, the fix
is a better case or a logged model/skill finding — not editing the rubric to make
the number come out.

## Design

Built from the references the maintainer pointed at, with two later refinements
(positives must come from real news, and the set must span industries):

- **Anthropic skill-creator eval format** — `evals.json` with `id`/`eval_name`/
  `prompt`/`expected_output`/`assertions`, graded as `{text, passed, evidence}`.
  https://github.com/anthropics/skills/blob/main/skills/skill-creator/SKILL.md
- **Anthropic "Demystifying evals for AI agents"** — build a *balanced* set; test
  where a behavior should fire and where it should not.
  https://www.anthropic.com/engineering/demystifying-evals-for-ai-agents
- **last30days** skill for case framing / "what people actually launch".
  https://github.com/mvanhorn/last30days-skill

### v2 (16 cases, 8+ industries)

- **Negatives / should-not** (the bias the skill exists to stop):
  - `1` Fundraisly, `2` Vokal, `3` Paste MCP — Product Hunt #1 / AI-buzzword /
    incremental dev tool → HOLD. A community-vote ranking is not a hook.
  - `14` bank CD-rate promo — routine corporate PR matching the market → HOLD.
  - `7` restaurant chain on Anthropic IPO, `8` SEO agency on DeepSeek —
    big/topical event, no real standing (keyword-adjacency trap) → SKIP.
  - `9` meal-kit brand on an active E. coli outbreak → AVOID (kill switch).
  - `10` week-old Alphabet raise → WAIT/SKIP (stale-window cap).
- **Balanced counterexample:** `4` Co-Invest — a genuine first-mover claim with
  *unverified* traction → REVISE (not floored, not PROCEED). Stops the set from
  teaching "Product Hunt → reject."
- **Positives / should-be-newsworthy, from REAL news + REAL press releases:**
  - `5` **NVIDIA "Vera, the CPU for Agents"** launch (real press release) →
    PROCEED. New category, dominant player, named partners, concrete spec.
  - `6` **Anthropic confidential IPO / $65B raise** + a Claude-based builder →
    RIDE (adjacent standing, capped).
  - `11` **FDA cell & gene therapy guidance** + a gene-therapy biotech → RIDE.
  - `12` **37M-lb Trader Joe's/Kroger frozen recall** + a food-safety lab → RIDE.
  - `13` **SoftBank €75B / 5GW France data centers** + a cooling/power firm → RIDE.
  - `15` **Fed rate pressure / June FOMC** + a lending fintech with data → RIDE.
  - `16` **Tesla/Volvo EV recalls** + an EV-safety engineering firm → RIDE
    (preventive recall → explainer; kill switch must *not* fire — the
    recall-vs-active-casualty distinction, contrasted with case 9).

### Industries covered

ai/saas · developer-tools · semiconductors · fintech · banking & finance-macro ·
ai-enterprise · pharma-biotech · food (cpg / safety / service) · energy &
datacenters · automotive-ev · marketing-services.

## Real-world anchors (week of 2026-06-01, now-anchor 2026-06-04)

| Cases | Real basis | Source |
|---|---|---|
| 1-4 | Fundraisly / Vokal / Paste MCP / Co-Invest PH launches | producthunt.com/leaderboard/daily/2026/6/2 |
| 5 | NVIDIA unveils "Vera, the CPU for Agents" | nvidianews.nvidia.com/news/nvidia-unveils-vera-the-cpu-for-agents |
| 6, 7 | Anthropic confidential IPO, $65B round at ~$965B | cnbc.com / fortune.com / cnn.com (2026-06-01) |
| 8 | DeepSeek raising ~$7.4B (Tencent, CATL) | techstartups.com Top Tech News 2026-06-03 |
| 10 | Alphabet $80B stock sale for AI compute | techstartups.com Top Tech News 2026-06-03 |
| 11 | FDA draft guidance to accelerate cell & gene therapies | fda.gov (2026-06-02) |
| 12 | 37M-lb Trader Joe's/Kroger frozen recall (glass) | foxbusiness.com |
| 9 | E. coli O157:H7 outbreak linked to beef kofta | efoodalert.com / FSIS |
| 13 | SoftBank up to €75B / 5GW AI data centers, France | techstartups / news reports |
| 15 | Fed holds 3.5-3.75%, 8-4 split, FOMC June 16-17 | federalreserve.gov / reporting |
| 16 | Tesla Model Y/3 propulsion + Volvo EX30 battery recalls | autoevolution.com / thestreet.com |

Event "clients" are realistic but non-fixture (e.g. "a Claude-based enterprise
builder", "a food-safety testing lab") so standing is grounded without putting
words in a specific named third party's mouth. Case 5's client is NVIDIA itself,
since its real press release is the input.

## Files

- `evals.json` — dataset. Every assertion carries a machine-checkable `check`.
- `grade.py` — deterministic grader. `python3 grade.py <outputs_dir>` →
  `grading.json` + a printed pass/fail summary. Reusable across models/versions.
- `runs/2026-06-04-benchmark/` — the 3-model × {with-skill, without-skill}
  benchmark:
  - `<full-model-id>/with-skill/outputs/` + `grading.json`
  - `<full-model-id>/without-skill/baseline.json`
  - `baseline-prompts.json` — shared, stripped, **answer-key-free** scenarios so
    baseline models can't be contaminated by `expected_output`/`assertions`.
  - `benchmark.md` + `benchmark.json` — aggregate results, the measured bias, and
    the Haiku schema-completeness finding.

  Model IDs benchmarked: `claude-opus-4-8`, `claude-sonnet-4-6`,
  `claude-haiku-4-5-20251001`.

## How to run

1. Executor (blind to assertions): an agent reads
   `skills/newsworthiness-check/{SKILL,rubric,examples}.md` and each `prompt`,
   emits the skill's strict JSON to
   `runs/2026-06-04-benchmark/<model-id>/with-skill/outputs/<id>-<name>.json`.
2. Grade: `python3 grade.py runs/2026-06-04-benchmark/<model-id>/with-skill/outputs
   --out runs/2026-06-04-benchmark/<model-id>/with-skill/grading.json`.
3. Baseline (per model): a generic PR-assistant prompt runs on the shared
   `baseline-prompts.json` to measure with-skill vs without-skill lift.

## Result (2026-06-04 benchmark)

Gold reference = Opus+skill (passes 67/67).

| Model | +skill | −skill mean \|Δ\| vs gold / over-rated | schema |
|-------|--------|----------------------------------------|--------|
| claude-opus-4-8 | 16/16 | 0.93 / 1 | 16/16 |
| claude-sonnet-4-6 | 16/16 | 1.53 / 5 | 16/16 |
| claude-haiku-4-5-20251001 | 11/16 | 1.53 / 6 | 9/16 |

With the skill, all three models converge on the calibrated answer (mean
|Δscore| ≤ 0.27, zero over-rated). Without it they inflate 5-12× more and
over-rate launches/positives (real NVIDIA "Vera" launch: bare models 9/10/9 vs
the skill's 7-8). Haiku+skill's judgments match gold (0.13 delta) but it
**truncates 7/16 outputs**, dropping required fields — prefer Sonnet/Opus for this
skill or add a schema-completeness gate. Full writeup in
`runs/2026-06-04-benchmark/benchmark.md`.
