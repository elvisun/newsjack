---
name: ai-visibility-panel-design
description: "Select QA-approved canonical intent cells into a versioned AI-visibility tracking panel with partitions, variants, lanes, surfaces, locales, repetitions, separate exposure and priority weights, randomization, uncertainty, refresh rules, and campaign controls. Use after prompt QA or when revising an existing panel."
---

# AI Visibility Panel Design

Turn accepted cells into a defensible measurement plan. Do not generate prompts or invent precision.

This skill inherits the ethical floor from `skills/ETHICS.md`. It enforces anti-hallucination, explicit denominators, and decay-aware versioning. Anti-spray and human-send are not applicable.

## Inputs

Require:

- measurement charter;
- `prompt_architecture.json`;
- QA-approved candidates and complete rejection ledger;
- evidence-backed weight inputs, if any;
- run and review budget;
- variance-pilot observations, when available;
- prior panel version and campaign registry, when applicable.

Never inspect target baseline performance during selection.

## Select by strata

Use the canonical intent cell as the sampling unit. Variants and repeated runs are nested observations, not extra buyers.

Allocate across:

- proximity band;
- job, journey, and information act;
- ICP/role and locale/language;
- evidence grade/source type;
- measurement lane and surface;
- `core`, `rotating`, `sentinel`, `control`, and `aided` partitions.

Select within a stratum by evidence strength, language authenticity, decision relevance, and diversity. Preserve declared minimums or emit a waiver. Do not select by current target strength, weakness, gap size, or campaign desirability.

## Separate lanes

- `closed_model`: no external search/tools/files/RAG/history; fixed system, model/version, and sampling; fresh session.
- `retrieval`: record `required`, `allowed`, or `unavailable`, whether retrieval ran, queries when exposed, live/cached state, and citation metadata.
- `consumer_surface`: explicit clean or account archetype, device, locale, history/personalization state; never merge with API rollups.
- `campaign_experiment`: pre-registered frozen evergreen, unaided resonance, aided association, and matched unaffected controls.

Never mix aided statuses or lanes in a denominator.

For multi-sided products or marketplaces, also stratify estimands by `persona_id` or declared market side. Never silently pool buyer, provider, operator, partner, or other materially different populations into one denominator.

## Weight honestly

Store two separate components:

- `exposure`: best available audience, intent, locale, and surface prevalence evidence;
- `priority`: human-approved strategic importance.

Every factor needs confidence and version plus provenance: a source ID for exposure evidence, or a human-decision artifact ID and approver for priority judgment. Neither may depend on baseline visibility or campaign performance.

If credible exposure weights do not exist, use equal weights within declared strata. Do not label priority-weighted results market share, audience reach, consumer awareness, or share of users.

Normalize weights within their declared rollup. Warn when one weight dominates or effective sample size collapses.

## Allocate cells and repeats

Use these starting points, then adapt after the variance pilot:

| Tier | Unaided cells | Variants | Repeats |
| --- | ---: | ---: | ---: |
| diagnostic | 30–48 | 2 | 3 plus deeper sentinels |
| standard | 60–120 | 2 | 3; 5–8 unstable cells |
| research | 200–400 | 1–2 | pilot-determined |
| campaign add-on | 24–40 treatment plus 24–40 control | 1–2 | pilot-determined |

Pilot 12–20 diverse sentinels with 6–8 repeats over at least two time blocks. Estimate between-cell, within-cell, variant, day/time, model/surface, and invalid/parser variance. High within-cell correlation favors more unique cells; high run variance favors repeats.

If a subgroup has fewer than 20–30 distinct cells, show counts and responses rather than a percentage leaderboard.

Compute the wave budget from each selected prompt's actual lane and surface eligibility, variants, and repeats. Do not estimate cost as every prompt multiplied by every configured surface when some combinations are ineligible or waived.

## Uncertainty and reporting

Publish:

- unique cell count, variants, repeats, eligible and invalid runs;
- dates, models/versions, surfaces, locale, and lane;
- raw and weighted numerator/denominator;
- weight source/version and effective sample size;
- interval method, overlap with prior panel, and configuration drift;
- “conditional on this panel” and non-probability coverage limits.

Use Wilson intervals only when a simple unweighted stratum has one independent binary observation per canonical cell. With variants or repeated observations, use a cell-cluster bootstrap or a validated hierarchical method; use a stratified cluster bootstrap by canonical cell for weighted aggregates. Pair unchanged cells across periods. A panel-version comparison shows overlap-only change and both full-version levels.

Intervals quantify conditional run/sampling uncertainty; they do not repair coverage bias.

## Version and refresh

Freeze:

- immutable `panel_id`, semantic version, content hashes, randomization seed;
- exact metric definitions and denominators;
- partitions, configuration, weights, cadence, and campaign linkage;
- change ledger and next review date.

Default refresh, unless evidence says otherwise:

- monthly evidence intake;
- quarterly review of disjoint unaided partitions totaling exactly 100%: roughly 70–80% core, 5–10% sentinel/control, and the remainder—normally 10–25%—rotating; aided cells have a separate allocation;
- event-triggered review for product, category, locale, regulatory, model, or surface change;
- annual charter approval.

Attach explicit review-by or refresh rules not only to B5 stories but also to mutable plan, price, eligibility/availability, regulation, service-status, and feature claims. Undated company copy is not evidence that a fact is timeless.

Changing core, weights, metrics, or surface mix creates a new version and overlap bridge. Never overwrite history.

## Campaign claims

A before/after increase alone is not attribution. Require treatment/control definitions, pre-registration, and a credible experimental or counterfactual design before causal language.

Report the evidence ladder separately:

1. prompt-panel mention/framing/citation;
2. source or referral traffic;
3. self-reported discovery;
4. qualified lead/conversion;
5. incremental outcome from experiment/counterfactual.

Never rename rung 1 revenue attribution.

## Output

Give the human `tracking_plan.md` first: charter, coverage, exact prompts, lanes, weights, cadence, uncertainty, limitations, waivers, and Gate 4 decisions.

Write `panel.yaml` as the machine handoff using the contract in `../build-ai-visibility-panel/references/artifact-contracts.md`. Also emit `run_manifest_template.json` and `panel_change_ledger.json`.

Use the contract's exact top-level keys. In particular:

- `partitions.<partition>.canonical_cell_ids` contains cell IDs;
- `selected_candidate_ids` contains every and only selected QA-pass prompt ID;
- `weight.exposure` and `weight.priority` are separate mappings;
- `statistics.cluster_unit` is `canonical_cell_id`;
- `approvals`, `waivers`, and append-only `changes` are arrays;
- the run-manifest observation template uses every exact field name listed in the contract.

Do not replace arrays with prose pointers, rename `changes`, or bury the required observation fields inside descriptions. Reconcile all selected IDs and counts before writing the tracking plan.

Human Gate 4 approves weights, limitations, cadence, campaign claims, and frozen version. If approvals or pilot data are missing, label the result `provisional_directional`; still return the comprehensive evidence-supported candidate prompt list.
