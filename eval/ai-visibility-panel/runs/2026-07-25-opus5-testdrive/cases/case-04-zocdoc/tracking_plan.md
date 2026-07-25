# Tracking plan — panel-zocdoc-us-booking v0.1.0-provisional

Status `provisional_directional`. No gate approved. Read alongside `panel_report.md`.

## 1. Estimands and denominators

Five estimands, five denominators. **They never share a denominator, and no single "AI visibility score" is produced.**

| Estimand | Numerator | Denominator | Partition scope | Lane scope |
|---|---|---|---|---|
| est-001 `unaided_brand_presence` | Valid observations naming the target anywhere in the answer | Valid **unaided** observations only — 28 cells / 56 strings — split by lane × surface × wave | core, rotating, sentinel | all three |
| est-002 `competitive_mention_share` | Target mentions among discovery/acquisition brands named | All such brand mentions in the *same* unaided observations, using a brand list frozen at panel version | core, rotating, sentinel | all three |
| est-003 `citation_presence` | Observations whose exposed citations include a target-owned domain | Valid **retrieval-lane** observations where retrieval ran **and** citations were exposed | all | retrieval only |
| est-004 `aided_brand_knowledge` | B0 answers materially accurate on 4 declared facts | Valid **B0 target-aided** observations only — 4 cells / 8 strings | aided | all three |
| est-005 `answer_framing` | Target-mentioning observations coded into declared frames | Observations mentioning the target, split by aided status and lane | all | all three |

The four checkable facts for est-004: (a) booking is free to the patient; (b) the provider is charged per new-patient booking; (c) the amount varies by specialty and location; (d) a cancellation or no-show does not automatically void the fee.

**Denominator separation is a hard rule.** Unaided, category-aided, competitor-aided and target-aided observations are counted separately. Closed-model, retrieval and consumer-surface observations are counted separately. API and consumer-surface results never roll up together, silently or otherwise.

`campaign_response` is out of scope: no campaign, no pre-registration, no control cells.

## 2. Selected partitions

| Partition | Cells | Strings | Repeats/wave | Purpose |
|---|---|---|---|---|
| core | 19 | 34 | 3 | Stable, evidence-strong unaided cells; the backbone of est-001/002 |
| rotating | 3 | 6 | 3 | Grade-C and B5 discovery; refreshed or retired at review |
| sentinel | 6 | 16 | 6 | Higher-repeat cells for variance estimation and drift detection |
| control | 0 | 0 | — | Empty by design: no campaign to control for (waiver-001) |
| aided | 7 | 14 | 3 | B0/B2/B1-competitor cells with their own denominators |

Within the unaided panel the mix is core 68% / rotating 11% / sentinel 21%, close to the 70–80 / 15–25 / 5–10 guideline except that sentinels are over-weighted deliberately to support the pilot.

Quarantined and **not** in the panel: cell-023 (self-pay care-seeker), cell-024 (call capacity).

## 3. Surfaces

Six surfaces across three lanes. **The user named no surfaces; these are agent assumptions and a human must fix exact provider/model identities before wave 1.**

| Surface | Lane | Configuration |
|---|---|---|
| srf-api-a | closed_model | Frontier assistant A, API, tools off |
| srf-api-b | closed_model | Frontier assistant B, API, tools off |
| srf-ret-a | retrieval | Frontier assistant A, search enabled |
| srf-ret-b | retrieval | Answer engine with citations enabled |
| srf-con-a | consumer_surface | Consumer assistant app, clean logged-out archetype, en-US, desktop |
| srf-con-b | consumer_surface | Search AI answer surface, clean archetype, en-US, mobile |

**Closed-model lane excludes the three B5 cells** (cell-020, cell-021, cell-030): their estimand depends on knowledge of 2026 events that may postdate training, so a closed-model miss would be uninterpretable.

## 4. Repetitions and volume

Two variants per cell (three on sentinels). Variants and repeats are **nested observations under the canonical cell**, never counted as extra buyers.

- closed_model: 64 eligible strings → 234 observations per surface → **468**
- retrieval: 70 strings → 258 per surface → **516**
- consumer_surface: 70 strings → 258 per surface → **516**
- **Wave total ≈ 1,500 valid observations**, matching the declared (assumed) budget.

Repeat counts are starting points from the diagnostic tier. They are **provisional until the variance pilot reports** and should be re-derived from it.

## 5. Fresh-session and retrieval-state controls

- One new session per observation. No history carry-over, no memory, no personalization on consumer surfaces in wave 1 (clean logged-out archetype only).
- Multi-turn scripts (prompt-005a/b, prompt-011a/b, prompt-014a/b) run as two sequential messages **inside one session**; the whole script is hashed as one unit.
- Fixed system/developer prompt, fixed model and version as exposed, fixed sampling parameters, all captured in `configuration_hash`.
- Retrieval is **recorded, never assumed**: policy (`required` / `allowed` / `disabled` / `unavailable`), whether retrieval actually ran, generated queries where exposed, live vs cached state, and full citation metadata.
- Invalid observations (refusal, error, truncation, off-topic, parser failure) are excluded from both numerator and denominator and reported as a separate invalid rate per lane and surface. They are never silently dropped.

## 6. Weights

Two components, stored and reported **separately**. Never combined into one number.

**Exposure weight — equal within declared strata.** No credible exposure evidence exists: no query-volume data, no AI-conversation prevalence data, no locale or surface prevalence data for this category in this run. Equal weighting is a **stated convention conditional on this panel**. It is not prevalence, market share, audience reach, awareness, or share of users. Any report that presents a weighted figure must carry that sentence.

**Priority weight — equal, provisional, pending Gate 4.** No human has assigned strategic importance. Equal priority is a placeholder, not a decision. It cannot become real until someone states the business decision.

Neither component may depend on baseline visibility or campaign performance. Both are normalized within their declared rollup. Effective-sample-size and weight-dominance warnings are not computable until the pilot runs.

## 7. Uncertainty

- **Simple unweighted binary strata within one lane and aided status:** Wilson intervals.
- **Weighted aggregates:** stratified cluster bootstrap with the **canonical cell** as the cluster unit.
- **Period comparisons:** pair unchanged cells; a panel-version comparison reports overlap-only change *and* both full-version levels.

**Reporting rule that binds this panel today:** any subgroup with fewer than 20–30 distinct canonical cells is published as **counts and response excerpts, not a percentage leaderboard**. Under this panel that means *every* subgroup — care-seeker 16 unaided cells, practice 12, B0 aided 4, and every per-band split. Even the 28-cell unaided total sits at the edge.

Intervals quantify conditional run and sampling variability under this exact configuration. **They do not repair coverage bias.** This is a non-probability panel and generalises to nothing beyond itself.

## 8. Variance pilot

**Status: not run.** No precision target, interval width, or effective sample size can be honestly stated until it is.

Design: **12 cells × 8 repeats × 2 time blocks.** The 6 sentinel-partition cells (cell-025 … cell-030) plus 6 borrowed core cells (cell-002, cell-005, cell-010, cell-013, cell-016, cell-018) — a documented deviation from the 12–20 dedicated-sentinel recommendation, recorded as waiver-002, not a claim that 6 sentinels suffice.

Variance components to estimate: between-cell, within-cell, variant, day/time, model/surface, invalid/parser.

Decision rule: high within-cell correlation → shift budget toward more unique cells; high run-to-run variance → shift toward more repeats.

## 9. Randomization

Seed **20260725**. Candidate presentation order randomised within surface × wave; surface order randomised within wave; time-of-day blocks balanced across at least two blocks per wave. The seed and per-observation presentation index are recorded in the run manifest so any wave is reproducible.

## 10. Cadence, refresh and versioning

- Evidence intake: **monthly**
- Panel review: **quarterly** — next **2026-10-25**
- Charter approval: **annual** — next **2027-07-25**
- B5 review-by dates: cell-020 **2026-09-24**, cell-021 **2026-10-21**, cell-030 **2026-09-24**. A B5 cell whose anchor has expired is retired or re-anchored; it is not quietly carried forward.
- Event triggers for an off-cycle review: product change, pricing change, category shift, locale addition, regulatory change, model version change, surface change.

Changing core cells, weights, metric definitions or surface mix creates a **new semantic version plus an overlap bridge**. History is append-only in `panel_change_ledger.json`; no frozen version is ever overwritten.

## 11. Approvals

All four gates are **pending** with no approver named. Every gate is resumable from the artifacts: Gate 1 from `icp_hypotheses.json`, Gate 2 from `buyer_jobs.json`, Gate 3 from `prompt_qa.json` (`disputed_decisions_for_gate_3`), Gate 4 from `panel.yaml`. No gate was self-approved.

**Must be done before this can be called frozen:**

1. Compute and back-fill every SHA-256 hash (all are currently `null`).
2. Fix `prompt_architecture.json` spec-001 `job_id` → `job-003`.
3. Regenerate second variants for cell-004, cell-010, cell-015, cell-016.
4. Resolve the four Gate 3 disputed decisions.
5. Restore ≥30 unaided cells, or formally accept waiver-002b.
6. Fix exact model, version and surface identities in the charter.
7. Run the 12-cell variance pilot across two time blocks.
8. Obtain named human approval on Gates 1–4.

## 12. Limitations

1. **Non-probability panel.** Conditional on 35 cells, 70 strings, six surfaces, en-US, this configuration. Generalises to nothing else.
2. **No causal claim available.** No campaign, treatment, control or counterfactual exists. A before/after change is not attribution. If a campaign is later registered, the evidence ladder is reported in separate rungs — panel mention/framing/citation, then referral traffic, then self-reported discovery, then qualified conversion, then incremental outcome — and rung 1 is never renamed revenue attribution.
3. **The supplied URL was unreachable** (HTTP 403). Target standing rests on two reachable target-owned pages plus a tertiary source.
4. **The richest provider-side sources are competitor-published.** The `$35–$110` fee range is a claim, not a fact, and drives eight practice-owner cells.
5. **The strongest care-seeker language is from 2021** and has decayed.
6. **The only AI-use prevalence figures are target-commissioned.**
7. **Review evidence conflicts and is unresolved** — a 4.89/5 review page and a complaint aggregator describe the same service in opposite terms. Both are self-selected; neither is prevalence.
8. **Blinding was same-context, not fresh-subagent.** One contamination leak was caught and rejected; residual risk is medium.
9. **Semantic duplicate detection was not machine-run** — no fixed embedding model was declared.
10. **Equal weights are a convention, not evidence.**
11. **28 unaided cells is below the diagnostic floor**, and every subgroup is below leaderboard thresholds.
12. **es-US and every non-English locale are absent**, with machine translation prohibited.
13. **No caregiver role and no enterprise buying committee exist** in the panel, because no evidence supported them.

## 13. Claims permitted and prohibited

**Permitted:** "Under this panel and configuration, the target was named in N of M valid unaided observations on surface X in wave W." · "In B0 observations, the answer was materially accurate on K of 4 declared facts."

**Prohibited:** market share · audience reach · consumer awareness · share of users · any single "AI visibility score" · any causal or attribution claim from a before/after change · any statement that this panel is frozen, representative, or statistically significant.
