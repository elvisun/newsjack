# Tracking plan — Cloudflare AI-visibility panel v0.1.0 (provisional_directional)

Run date: 2026-07-25. Panel: `panel-cloudflare-network-platform`. Charter: `charter-cloudflare-2026-07-25`.

**Read this first.** Nothing here is frozen, representative, statistically powered, or causal. All four human gates are pending, no variance pilot has run, and no exposure-weight evidence exists. Every number this plan produces is *conditional on this panel*.

## 1. Estimands and denominators

Six estimands, six separate denominators. There is no blended "AI visibility score" and one will not be created.

| Estimand | Numerator | Denominator | Cells | What it does not prove |
| --- | --- | --- | ---: | --- |
| `unaided_brand_presence` | Answers naming the target brand or a target product | Valid obs from `aided_status=unaided` AND `partition ∈ {core, sentinel}`, within one lane × surface × locale × wave | 26 | Awareness, market share, reach, or that a human ever typed these prompts |
| `unaided_brand_presence` (category-aided variant) | Same numerator | Valid obs from `aided_status=category_aided` — **separate denominator** | 7 | Naming a category is a prompt choice, not buyer behaviour. Never pool with the row above |
| `competitive_mention_share` | Answers naming the target, per vendor | Obs in the same cell set naming ≥1 vendor, split by aided status | 26 / 7 | Not market share, revenue share, or consideration-set share |
| `citation_presence` | Obs where a target-owned domain appears in exposed citations | Valid `retrieval` + `consumer_surface` obs where retrieval ran and citations were exposed | 44 | Not traffic, clicks, or influence |
| `answer_framing` | Obs coded against a pre-registered framing codebook | Per code, within one cell set × lane × surface × wave | 41 | **Blocked**: no codebook, no rater-agreement measurement yet |
| `aided_brand_knowledge` | B0 obs returning a source-checkable correct statement about the named target | Valid obs from `aided_status=target_aided` only | 7 | Aided accuracy says nothing about unaided reach |

`campaign_response` is declared **out of scope**: no campaign, dates, treatment definition, controls or pre-registration were supplied.

Hard rule: aided statuses and lanes never share a denominator. The category-aided stratum has 7 distinct cells — below the 20–30 threshold — so it is reported as **counts plus example answers, not a percentage leaderboard**.

## 2. Selected partitions

| Partition | Cells | Accepted candidates | Role |
| --- | ---: | ---: | --- |
| core | 25 | 50 | Primary estimand carriers, paired across waves |
| sentinel | 8 | 16 | Stability probes and the variance pilot backbone |
| aided | 9 | 18 | B0 target-aided, B5 target-aided, and 2 competitor-aided; own denominators |
| control | 2 | 2 | Matched-unaffected, no target standing; drift and spurious-mention detection |
| rotating | 6 | **0** | All six are quarantined or awaiting revision — **contributes nothing in wave 1** |

That last row is a real gap, not a formatting artefact. Procurement criteria (cell-018), both edge-compute probes (cell-044/045), the unaided post-purchase support intent (cell-046), the only non-US locale (cell-047), and the outage retrospective (cell-035) all currently have zero runnable candidates.

## 3. Surfaces and lanes

| Lane | Surfaces | Policy |
| --- | --- | --- |
| `closed_model` | surf-api-a, surf-api-b | No search, tools, files, RAG or history. Fixed empty developer prompt, declared model+version, temperature 0 where settable, fresh session per observation |
| `retrieval` | surf-rag-a, surf-rag-b | Record `required`/`allowed`/`unavailable`, whether retrieval actually ran, generated queries when exposed, live vs cached, full citation metadata |
| `consumer_surface` | surf-cons-a, surf-cons-b | Clean archetype only: new account, no memory, no personalisation, declared device, en-US. **Never** rolled up with API lanes |
| `campaign_experiment` | — | Defined, not instantiated |

All six surfaces are agent assumptions pending approval. If a retrieval-lane observation returns `retrieval_ran=false`, it stays valid but is reported in its own subgroup.

## 4. Repetitions, fresh-session and retrieval-state controls

- 3 repeats per accepted candidate per surface-lane combination in a normal wave.
- 6 repeats for the 12 pilot cells, across ≥2 time blocks.
- 5–8 repeats for cells the pilot flags as unstable — the exact number is **not** set yet.
- Fresh session is required for every observation. For the two scripted multi-turn cells (cell-003, cell-021), the session is fresh at TURN 1 and both turns run in that same session; turns are recorded as separate observation rows with `turn_index`.
- Session reuse across cells is prohibited, because carry-over would contaminate the unaided condition.
- Retrieval state, generated queries and citation payloads are captured per observation via `run_manifest_template.json`. Live-versus-cached state is recorded because a cached retrieval and a live retrieval are not the same measurement.

Planned wave volume: **1,590 observations** against a 2,000 ceiling.

## 5. Weights

Two components, stored and reported separately, never blended.

- **Exposure weight**: `equal_within_stratum`. Status: `no_evidence`. No audience, intent, locale or surface prevalence data exists anywhere in `source_manifest.json`. Equal weighting is a *convention chosen because nothing better is available*, not a claim that these intents occur equally often in the world. Exposure-weighted figures may not be called market share, audience reach, awareness, or share of users.
- **Priority weight**: `equal_within_stratum`. Status: `pending_human_gate_4`. Strategic priority is a human judgement; until an approver sets it, no priority-weighted rollup should be published at all.

Normalisation happens within the declared rollup. Warn when any single weight dominates or when effective sample size collapses.

## 6. Uncertainty

- Wilson intervals for simple unweighted binary strata.
- Stratified cluster bootstrap for weighted aggregates, clustered on `canonical_cell_id`. Variants and repeats are nested observations, not extra buyers.
- Report per stratum: unique cells, variants, repeats, eligible runs, invalid runs and invalid reasons, dates, models/versions, surfaces, locale, lane, raw and weighted numerator/denominator, weight source/version, effective sample size.
- Cross-wave comparison pairs unchanged cells. A version comparison shows overlap-only change **and** both full-version levels.
- Intervals quantify run and sampling variability conditional on this panel. They do not repair coverage bias and do not license a population claim.

## 7. Variance pilot

**Status: not started.** Until it completes, `precision_status` stays `directional_only` and no minimum detectable difference or power figure may be quoted.

Design: 12 diverse cells (cell-003, 014, 020, 032, 034, 036, 037, 038, 039, 040, 041, 042) × 6 repeats × 2 time blocks × 2 lanes ≈ 576 observations. Estimate between-cell, within-cell, variant, day/time-block, model/surface and invalid/parser variance. High within-cell correlation → add unique cells; high run-to-run variance → add repeats. Reallocate, then reissue as v0.2.0 with an overlap bridge.

## 8. Randomization

- Cell order randomised within each surface-lane block.
- Variant order randomised within each cell.
- **Seed: `null`.** No runtime was available to generate one. A seed must be generated and frozen before wave 1; this is waiver-05.

## 9. Cadence, refresh and version policy

| Activity | Cadence |
| --- | --- |
| Evidence intake | Monthly |
| Panel review | Quarterly, targeting 70–80% core / 15–25% rotating / 5–10% sentinel+control |
| Charter approval | Annual (`2027-07-25`) |
| Next review | **2026-08-25** |

Event triggers: product change, category shift, locale addition, regulatory change, model or surface change, B5 evidence expiry.

B5 decay schedule (these are the only time-bound cells):

| Cell | Evidence | Review by | Rule |
| --- | --- | --- | --- |
| cell-032, cell-033, cell-051 | src-004, published 2026-07-01 | **2026-09-15** | Retire or re-render after the effective date |
| cell-034 | src-003, published 2025-12-08 | 2026-09-30 | Refresh the counts or retire |
| cell-035 | src-013, 2025-11-19 | 2026-08-31 | Aging story — demote or retire |

Changing core membership, weights, metric definitions or surface mix creates a new version with an overlap bridge. History is append-only in `panel_change_ledger.json`; nothing is overwritten.

## 10. Approvals

| Gate | Owns | Status | Resume from |
| --- | --- | --- | --- |
| 1 | ICPs, exclusions, permissions, standing facts | **pending** | `icp_hypotheses.json` |
| 2 | Jobs, language, roles, locales, priority | **pending** | `buyer_jobs.json` |
| 3 | Core/aided/campaign partitions, disputed QA | **pending** | `prompt_qa.json` → `summary.gate_3_disputed_decisions` |
| 4 | Weights, limitations, cadence, claims, freeze | **pending** | `panel.yaml` |

Five decisions are explicitly queued for Gate 3, including whether turn form alone justifies splitting cell-003 from cell-036, and whether a *paraphrase* of vendor policy framing is acceptable inside an unaided B5 cell.

## 11. Limitations and waivers

1. Provisional and directional. Not frozen, not representative, not powered, not causal.
2. **No grade-A evidence exists in this run.** No interviews, call transcripts, on-site search logs, support tickets or permitted AI-conversation corpus were supplied. The strongest evidence is dated independent reporting plus public forum comments with permalinks.
3. Blinding was **procedural, not architectural**. No fresh subagent or separate session was available, so a single context enforced the boundary by working only from `blind_design_brief.json` during the unaided pass. This is weaker than a fresh-context generator and is stated rather than glossed.
4. Five sources were reached only via search-result summaries; two intended buyer-language sources returned HTTP 403 and contributed nothing.
5. Embedding-based duplicate detection was not run — no fixed embedding model/version available. Duplicate detection used normalized-string comparison, lexical overlap and manual semantic review.
6. No content hashes and no seed: no hashing runtime. Every hash field reads `sha256:unavailable`. **The panel cannot be frozen until these are computed.**
7. Equal weights are a convention, not prevalence.
8. Rotating partition contributes zero observations in wave 1.
9. Locale coverage is en-US only; the single en-GB probe is quarantined.
10. Edge compute appears in the user's own description but is supported only by company assertion, so both its cells are quarantined rather than promoted to match the description.

Waivers waiver-01 through waiver-06 are recorded in `panel.yaml`, each with the specific evidence that would retire it.

## 12. Claims discipline

The evidence ladder is reported separately, never collapsed:

1. prompt-panel mention / framing / citation ← **this panel measures rung 1 only**
2. source or referral traffic
3. self-reported discovery
4. qualified lead or conversion
5. incremental outcome from an experiment or counterfactual

A before/after movement in rung 1 is not attribution. Rung 1 is never renamed revenue attribution.
