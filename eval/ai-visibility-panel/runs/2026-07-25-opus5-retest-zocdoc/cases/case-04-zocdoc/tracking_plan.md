# Tracking plan — `panel-zocdoc-booking` v0.1.0 (`provisional_directional`)

Companion to `panel_report.md`. Machine handoff: `panel.yaml`, `run_manifest_template.json`, `panel_change_ledger.json`.

## 1. Estimands and denominators

Six quantities, each with its own denominator. **They are never pooled and there is no combined score.**

| Estimand | Numerator | Denominator | Does not prove |
| --- | --- | --- | --- |
| `unaided_brand_presence` | Valid answers mentioning the target | Valid observations in core+sentinel, `unaided`, one lane × surface × locale × wave | Market share, awareness, reach, bookings, revenue |
| `aided_brand_knowledge` | Valid target-aided answers with a materially correct, source-checkable target fact | Valid `aided` partition observations with `target_aided`, one lane × surface × locale × wave | Unaided discoverability or preference |
| `competitive_mention_share` | Mentions of one named option in valid unaided answers | All option mentions in the same unaided stratum | Category leadership or booking volume |
| `citation_presence` | Retrieval answers exposing a target-owned citation | Valid retrieval observations where citations were exposed at all; aided and unaided reported separately | Referral traffic or influence on the answer |
| `answer_framing` | Answers coded into a fixed framing category | Valid coded observations in the same partition × aided status × lane × surface × locale × wave | Behavioural effect or causation |
| `control_false_positive_rate` | Control-cell answers mentioning the target | Valid control observations, same lane × surface × wave | Anything about in-perimeter presence |

Invalid runs (refusal, provider error, truncation, off-topic) leave both numerator and denominator and are reported as a per-lane invalid rate.

## 2. Selected partitions

41 cells, 78 prompts. Core 24 cells / 48 prompts · aided 8 / 16 · rotating 2 / 2 · sentinel 5 / 10 · control 2 / 2.
Excluded with waivers: `cell-033`, `cell-034` (es-US, native review pending) and `cell-037` (grade D). Their candidates stay in `prompt_universe.json` with quarantine decisions so nothing disappears silently.

The canonical intent cell is the sampling unit. Variants and repeats are **nested observations, not extra buyers**.

## 3. Surfaces, repetitions, and session/retrieval controls

| Surface | Lane | Config |
| --- | --- | --- |
| `surface-api-chat-closed` | closed_model | Search/tools/files/history disabled; fixed model+version, fixed minimal system prompt, fixed lowest-variance sampling; fresh session per observation |
| `surface-api-chat-retrieval` | retrieval | Search allowed; record `retrieval_used`, generated queries when exposed, live/cached state, and every citation |
| `surface-consumer-assistant-clean` | consumer_surface | Clean archetype account, memory/personalization disabled where the setting exists, declared device and locale, new conversation per observation |

**Repeats:** 3 per prompt per surface per wave; 8 for pilot cells. Core/aided/sentinel/control run on all eligible lanes for their cells; the two rotating cells run on retrieval (and closed_model for `cell-035`/`cell-036`) only.
**Wave volume:** ≈650 scored observations. API and consumer-surface observations are **never** merged into one rollup.
**Every observation** records the exact prompt, configuration hash, provider/model version string, lane, surface, locale, account/personalization state, search policy, whether retrieval ran, session state, order seed and index, response/citation payload hashes, timestamp, retry and validity status, and parser version (`run_manifest_template.json`). All hash fields are `null` today — that is a freeze blocker, not a placeholder.

## 4. Weights

Two separate components, never combined into one number:

- **Exposure — `equal_within_declared_strata`, zero factors.** No source in the manifest measures query volume, assistant usage, or job prevalence. Equal weighting is a conditional analytic choice and **must not be described as prevalence, reach, or share of users.**
- **Priority — `withheld_pending_human_approval`, zero factors.** Strategic importance is a human judgment and no approver exists yet.

Exposure- and priority-weighted results are reported on separate lines with their own effective sample sizes. A warning fires if any future weight dominates or if effective sample size collapses.

## 5. Uncertainty

- Wilson score intervals for simple unweighted binary strata within one partition × aided status × lane × surface × locale × wave.
- Stratified cluster bootstrap resampling `canonical_cell_id` for weighted aggregates, with variants and repeats nested.
- Subgroups with fewer than 20 distinct cells are reported as **counts and example responses**, not as a percentage leaderboard. That currently applies to almost every subgroup here, including all B0/B1/B2 aided cells.
- Wave-over-wave comparisons pair unchanged cells and show overlap-only change plus both full-version levels.
- Intervals quantify run/sampling noise **conditional on this panel**. They do not repair coverage bias.

## 6. Variance pilot (pending)

15 cells × 8 repeats × 2 time blocks: `cell-038`–`cell-042` plus `cell-001, 005, 011, 013, 016, 020, 022, 025, 031, 043`. Estimates between-cell, within-cell run, variant, day/time-block, model/surface, and invalid-parse variance. High within-cell correlation → add unique cells; high run variance → add repeats. Until it runs, the repeat counts above are defaults, not derived values. The five dedicated sentinels fall short of the 12–20 the method wants, so ten core cells participate without changing partition (`waiver-002`).

## 7. Randomization

Seed `20260801`. Cell order and surface order are randomized within each wave; the realized `order_index` is recorded per observation so order effects can be tested rather than assumed away.

## 8. Cadence, refresh, and versioning

- Monthly evidence intake; quarterly panel review; annual charter re-approval.
- Target mix at review: ~70–80% core, 15–25% rotating, 5–10% sentinel/control.
- **Review-by 2026-10-25** for all three B5 cells and for every mutable claim: per-booking fee range, 24–72 hour availability, payer/partner integrations, spend-control features. Undated company copy is not treated as timeless.
- Event triggers: target pricing or policy change; a new payer/partner booking surface; provider-directory regulation or enforcement; model or consumer-surface version change; a category entrant or exit named by new evidence.
- Changing core cells, weights, metric definitions, or surface mix creates a **new version with an overlap bridge**. `panel_change_ledger.json` is append-only; frozen history is never overwritten.

## 9. Approvals

All four gates are **pending** and resumable from the artifacts: Gate 1 ICPs, Gate 2 jobs, Gate 3 partitions plus six disputed QA decisions, Gate 4 weights/limitations/cadence/version. Freeze blockers: null `source_manifest_hash`, null `prompt_hash` values, null `blind_brief_hash`, no pilot, pending locale review, no approver.

## 10. Limitations

1. Conditional on this panel; not a probability sample of AI users, patients, or practices.
2. Equal exposure weighting is an assumption with no supporting source.
3. All buyer language is public proxy evidence; no first-party corpus was supplied.
4. Several first-party target pages returned HTTP 403, so the factual perimeter and contamination lexicon may be incomplete.
5. No variance pilot; repeat counts and interval widths are defaults.
6. B5 cells decay, and two of three lean on vendor-published dated material.
7. Two rotating cells (`cell-035`, `cell-036`) rest on grade-C inference and are discovery only.
8. Consumer-surface results depend on an account archetype that cannot be guaranteed identical to any real user's state.
9. **No causal or attribution language is licensed by this design.** A rise in mentions after a content or PR change is rung 1 of the evidence ladder (panel mention), not incremental outcome. Rungs 2–5 (referral traffic, self-reported discovery, qualified conversion, experimental increment) require separate instrumentation and a pre-registered treatment/control design that does not exist here.
