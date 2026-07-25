---
name: prompt-proximity-architecture
description: "Turn an approved measurement charter, ICPs, and buyer jobs into a budget-aware prompt coverage blueprint across proximity bands, aided status, information acts, journey states, roles, locales, evidence grades, partitions, and measurement lanes. Use before prompt wording to define required, optional, and prohibited canonical intent cells."
---

# Prompt Proximity Architecture

Design the cells before writing the strings.

This skill inherits the ethical floor from `skills/ETHICS.md`. It enforces anti-hallucination and evidence-bound coverage. Anti-spray and human-send are not applicable.

## Inputs

Require:

- a measurement charter;
- approved `icp_hypotheses.json` and `buyer_jobs.json`;
- target run/review budget;
- required locales, surfaces, and lanes;
- any campaign partition and prior-panel constraints.

If the charter is provisional, design a provisional architecture and name the gaps. Reject a charter that says only “track AI visibility.”

## Keep dimensions independent

Each intent cell fixes:

- buyer job;
- information act;
- journey state;
- material constraints;
- persona or buying role;
- locale and language;
- prompt-proximity band;
- expected answer kind.

Add independent tags for evidence grade, partition, lane eligibility, turn form, and optional funnel. Do not make funnel the schema.

Use the supported acts `explain`, `diagnose`, `plan`, `generate`, `compare`, `recommend`, `verify`, `navigate`, `buy`, `implement`, and `troubleshoot`. Include only acts entailed by the job evidence.

Use journey states `problem_identification`, `exploration`, `requirements_building`, `supplier_selection`, `adoption`, and `post_purchase`.

## Assign proximity

| Band | Structure | Default aided status |
| --- | --- | --- |
| `B0_direct_brand_product` | Names target brand/product; asks about facts, fit, use, reputation, support, or implementation | `target_aided` |
| `B1_comparison_purchase` | Shortlist, recommendation, alternatives, pricing, requirements, or comparison | `unaided`; `competitor_aided` when only competitors are supplied; `target_aided` for a declared target-vs-competitor comparison |
| `B2_category` | Names an accepted solution category, not the target | `category_aided` |
| `B3_problem_need` | Describes pain, risk, trigger, or constraint without category/target | `unaided` |
| `B4_job_goal` | Asks for progress/outcome without supplying a solution category | `unaided` |
| `B5_broad_discovery_story` | Trend, event, regulation, practice, or narrative connected to the job | `unaided` |

`campaign_exposed` is an independent flag. Never combine target-aided, competitor-aided, category-aided, unaided, or campaign-exposed cells in one denominator.

Because `aided_status` is single-valued, a B1 prompt naming both the target and
a competitor is `target_aided`; record the competitor stimulus in the
contamination exception and keep the cell in the aided partition. A B1 prompt
naming competitors but not the target is `competitor_aided`.

Do not assign funnel mechanically:

- B0 can be post-purchase, not BOFU.
- B3 can describe an urgent funded decision.
- B1 can be early exploration.
- `funnel` may be `null`.

B5 requires fresh dated public evidence and a decay/refresh rule.

The charter's product/capability perimeter is also a coverage dimension. For every named area, create at least one evidence-supported job/cell or a `coverage_gap` with an exclusion/waiver and the evidence needed. Never silently omit a named area.

## Allocate coverage

Cover every dimension that the evidence supports, not every Cartesian combination. Never generate a full persona × locale × act × constraint grid.

For a diagnostic default, target 30–48 unique unaided cells, two variants per cell, and three repeats per wave. For a standard panel, target 60–120 cells. Adjust to the user's budget and keep aided cells outside the unaided quota.

Stratify minimums and targets across:

- B0–B5 where evidence supports each band;
- job, journey, and information act;
- ICP/role and materially distinct locale;
- evidence grade/source type;
- core, rotating, sentinel, control, and aided partitions;
- closed-model, retrieval, consumer-surface, and campaign lanes.

Protect underrepresented but decision-relevant strata. Grade-D cells are rotating discovery until independently validated or explicitly promoted.

When the budget cannot meet required coverage, return an allocation conflict. Do not silently drop a band, role, locale, or control.

## Prohibit unsupported cells

Mark a cell prohibited when it:

- changes an informational need into a recommendation solely to elicit brands;
- introduces a category not supported by the job evidence;
- treats a competitor-named prompt as unaided;
- uses target/campaign copy outside the allowed aided lane;
- assumes translation is locale equivalence;
- derives a core cell from current AI answers or target pages;
- exceeds the declared sampling budget without a waiver.

## Output

Give a Markdown coverage summary first: required dimensions, planned counts, gaps, conflicts, and why each missing band is unsupported.

Then write `prompt_architecture.json`:

```json
{
  "schema_version": "1.0.0",
  "artifact_id": "architecture-<stable-slug>",
  "created_at": "RFC3339",
  "created_by": "declared agent or human",
  "source_manifest_hash": null,
  "warnings": ["hash_not_computed: compute source_manifest_hash before freeze"],
  "cells": [
    {
      "cell_spec_id": "spec-001",
      "job_id": "job-001",
      "icp_ids": ["icp-001"],
      "proximity_band": "B3_problem_need",
      "aided_status": "unaided",
      "campaign_exposed": false,
      "information_act": "diagnose",
      "journey_state": "problem_identification",
      "funnel": null,
      "persona_id": "role-001",
      "locale": "en-CA",
      "language": "en",
      "material_constraints": [],
      "expected_answer_kind": "diagnosis_and_options",
      "turn_form": "single_turn",
      "lane_eligibility": ["closed_model", "retrieval"],
      "partition": "core",
      "evidence_grade": "A",
      "target_variants": 2,
      "required": true,
      "reason_source_ids": ["source-001"]
    }
  ],
  "prohibited_cells": [],
  "allocation": {
    "core_cells": 36,
    "rotating_cells": 8,
    "sentinel_cells": 12,
    "control_cells": 0,
    "aided_cells": 6,
    "budget_status": "within_budget | conflict | waived"
  }
}
```

Before handoff, require every `cells[].job_id` and every `cells[].icp_ids[]` to resolve exactly. There is one authoritative field for each relationship. Never leave a stale ID and add a compensating alias such as `job_id_authoritative`; rewrite the cell or stop with an allocation/reference error.

## Handoff

Pass the architecture plus a target-free blind brief to `realistic-prompt-generation`. Keep the contamination register outside the generator's context.
