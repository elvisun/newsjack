# AI visibility panel artifact contracts

Read this file before creating or validating a panel run.

## Contents

1. Common envelope
2. Required files
3. Source manifest
4. Charter
5. Contamination and blinding
6. Canonical prompt record
7. QA and panel
8. Human prompt-table columns
9. Run manifest
10. Completion checklist

## Common envelope

Every JSON/YAML artifact carries:

```json
{
  "schema_version": "1.0.0",
  "artifact_id": "stable-id",
  "created_at": "RFC3339",
  "created_by": "human-or-agent",
  "source_manifest_hash": null,
  "warnings": ["hash_not_computed: compute source_manifest_hash before freeze"]
}
```

Use stable slug-like IDs. References must resolve. Hash canonical UTF-8 content with sorted keys and no incidental whitespace when the runtime supports it. To avoid a self-referential digest, `source_manifest_hash` is the SHA-256 of the canonicalized `sources` array only; every downstream artifact copies that value. A candidate `prompt_hash` is the SHA-256 of its exact UTF-8 `text`.

`null` is permitted only on a `provisional_directional` run when the runtime cannot compute a hash. Add a warning naming the missing hash and make it a freeze blocker. Never write a plausible digest or placeholders such as `sha256:unavailable`. Frozen artifacts require real SHA-256 values. The same provisional rule applies to `content_hash`, `blind_brief_hash`, `prompt_hash`, and response/configuration hashes.

## Required files

```text
panel_report.md
tracking_plan.md
measurement_charter.json
source_manifest.json
icp_hypotheses.json
buyer_jobs.json
contamination_register.yaml
blind_design_brief.json
prompt_architecture.json
prompt_universe.json
prompt_qa.json
panel.yaml
run_manifest_template.json
panel_change_ledger.json
```

Use a run directory chosen by the user or a new explicit directory. Never overwrite a frozen panel.

## Source manifest

`source_manifest.json` contains:

```json
{
  "schema_version": "1.0.0",
  "artifact_id": "sources-example",
  "created_at": "2026-01-01T00:00:00Z",
  "created_by": "agent",
  "source_manifest_hash": null,
  "warnings": ["hash_not_computed: compute source_manifest_hash before freeze"],
  "sources": [
    {
      "source_id": "source-001",
      "url": "https://example.com/page",
      "title": "Page title",
      "publisher": "Publisher",
      "published_at": null,
      "accessed_at": "2026-01-01T00:00:00Z",
      "source_class": "company_asserted | buyer_behavior | independent | search_proxy | llm_hypothesis",
      "source_type": "product | pricing | support | technical | review | forum | interview | query | procurement | news | other",
      "permission": "public | user_authorized | generated",
      "evidence_grade": "A | B | C | D",
      "span": "Short locatable excerpt or faithful paraphrase",
      "span_locator": "heading, paragraph, line, or query",
      "fact_type": "capability | limitation | segment | trigger | job | criterion | language | competitor | market_event | other",
      "confidence": "high | medium | low",
      "content_hash": null
    }
  ]
}
```

Do not store unnecessary personal data or long copyrighted excerpts. Grade A requires direct relevant behavior or verbatim customer/prospect language with provenance. Grade B is a credible public behavior/search proxy. Grade C is assertion/hypothesis. Grade D is unsupported model expansion.

Public and user-authorized evidence requires a non-empty URL or authorized-file locator. `generated` is allowed only with `source_class: llm_hypothesis`, `evidence_grade: D`, and `url: null`; it never counts toward the minimum evidence mix or supports a material claim.

## Charter

`measurement_charter.json` declares:

- `business_decision`;
- `estimands[]`;
- `target_population`, `exclusions[]`;
- `products[]`, `markets[]`, `locales[]`, `time_horizon`;
- `surfaces[]`, `lanes[]`, `reporting_strata[]`;
- `run_budget`, `review_budget`;
- `precision_status`;
- `approver`, `approval_status`.

Estimands:

- `unaided_brand_presence`;
- `aided_brand_knowledge`;
- `competitive_mention_share`;
- `citation_presence`;
- `answer_framing`;
- `campaign_response`.

Each estimand stores exact numerator, denominator, eligible partitions/lanes, and what it does not prove.

Machine shapes are strict:

- `lanes` and `surfaces` are arrays of enum strings, not arrays of policy objects;
- each `estimands[]` item uses `name` for the allowed estimand enum;
- put lane policies in a separate `lane_policies` mapping;
- put surface configuration in a separate `surface_policies` mapping.

Minimal shape:

```json
{
  "business_decision": "Decide what this panel can safely monitor",
  "estimands": [
    {
      "name": "unaided_brand_presence",
      "numerator": "Valid eligible answers that mention the target",
      "denominator": "All valid observations in the declared unaided partition, lane, surface, locale, and wave",
      "eligible_partitions": ["core", "sentinel"],
      "eligible_lanes": ["closed_model", "retrieval"],
      "does_not_prove": "Market share, awareness, audience reach, or revenue attribution"
    }
  ],
  "lanes": ["closed_model", "retrieval"],
  "surfaces": ["declared-surface-id"],
  "lane_policies": {
    "closed_model": "No external retrieval; fresh session",
    "retrieval": "Record whether retrieval ran and any exposed citations"
  }
}
```

## Contamination and blinding

`contamination_register.yaml` stores:

- `target_terms.brands`;
- `target_terms.products`;
- `target_terms.domains`;
- `target_terms.people`;
- `target_terms.slogans`;
- `target_terms.proprietary_categories`;
- `target_terms.campaign_terms`;
- `target_terms.flattering_claims`;
- `competitor_terms`;
- `allowed_exceptions`.

It must parse as YAML 1.2. Prefer JSON-compatible YAML: quote every free-text scalar, especially text containing `:`, `#`, parentheses, or apostrophes; represent a list plus its explanation as separate mapping fields; never place a mapping key inside a scalar list. JSON syntax is valid YAML and is acceptable in this `.yaml` file.

`blind_design_brief.json` contains only approved jobs, anonymized role labels, constraints, locales, safe language fragments, evidence IDs/grades, and required strata. It must not contain target terms, current answers/performance, cited/desired pages, or campaign copy.

## Canonical prompt record

`prompt_universe.json` nests candidates under canonical cells. Every accepted candidate must resolve to:

```json
{
  "canonical_cell_id": "cell-001",
  "cell_spec_id": "spec-001",
  "job_id": "job-001",
  "icp_ids": ["icp-001"],
  "information_act": "diagnose",
  "journey_state": "problem_identification",
  "funnel": null,
  "proximity_band": "B3_problem_need",
  "aided_status": "unaided",
  "campaign_exposed": false,
  "persona_id": "role-001",
  "locale": "en-CA",
  "language": "en",
  "material_constraints": [],
  "expected_answer_kind": "diagnosis_and_options",
  "turn_form": "single_turn",
  "lane_eligibility": ["closed_model", "retrieval"],
  "partition": "core",
  "evidence_grade": "A",
  "reason_source_ids": ["source-001"],
  "candidates": [
    {
      "candidate_id": "prompt-001a",
      "variant_role": "observed_language",
      "text": "Exact prompt text",
      "transformation": "lightly_normalized",
      "source_ids": ["source-001"],
      "evidence_grade": "A",
      "locale_review_status": "not_required",
      "generation_provenance": {
        "model": "declared-model",
        "prompt_hash": null
      }
    }
  ]
}
```

The shown canonical-cell fields are required and flat in every `canonical_cells[]` record. Copy them exactly from the referenced architecture cell, then add `canonical_cell_id` and `candidates`. Do not replace them with only `cell_spec_id`, `job_id`, and `expected_answer_kind`; do not hide them in an `intent`, `dimensions`, or metadata object. If the architecture changes, regenerate the canonical cell instead of adding aliases such as `job_id_authoritative`.

Enums:

- information act: `explain`, `diagnose`, `plan`, `generate`, `compare`, `recommend`, `verify`, `navigate`, `buy`, `implement`, `troubleshoot`;
- journey: `problem_identification`, `exploration`, `requirements_building`, `supplier_selection`, `adoption`, `post_purchase`;
- funnel: `TOFU`, `MOFU`, `BOFU`, `null`;
- band: `B0_direct_brand_product`, `B1_comparison_purchase`, `B2_category`, `B3_problem_need`, `B4_job_goal`, `B5_broad_discovery_story`;
- aided: `target_aided`, `competitor_aided`, `category_aided`, `unaided`;
- lane: `closed_model`, `retrieval`, `consumer_surface`, `campaign_experiment`;
- partition: `core`, `rotating`, `sentinel`, `control`, `aided`;
- transformation: `verbatim`, `lightly_normalized`, `search_query_expanded`, `human_written`, `llm_expanded`, `translated`, `locale_transcreated`.

## QA and panel

`prompt_qa.json` records exactly one `pass`, `revise`, `quarantine`, or `reject` decision per candidate in `prompt_universe.json`; cited rule results; duplicate action; route; reason; confidence; accepted IDs; and `baseline_fields_blinded`.

Every rejected or quarantined draft remains in `prompt_universe.json` so its decision resolves. `accepted_candidate_ids` contains every and only candidate whose decision is `pass`. A `counts` object, when present, is derived from the decision array and accepted-ID list immediately before writing; it is never an independently edited claim.

`panel.yaml` records:

- `panel_id`, semantic `version`, status, charter ID, source hash;
- estimands and exact denominators;
- canonical cell IDs by partition;
- selected candidate IDs and variants;
- lanes, surfaces, locales, model/search/session policy;
- repetitions, randomization seed, controls;
- `weight.exposure` and `weight.priority` separately with factor source IDs/confidence; human priority factors also carry a decision artifact ID and approver;
- uncertainty method, cluster unit, pilot status;
- refresh policy and next review;
- limitations, approvals, waivers, change ledger link.

Use these top-level machine shapes:

```yaml
panel_id: "panel-example"
version: "0.1.0"
status: "provisional_directional"
partitions:
  core:
    canonical_cell_ids: ["cell-001"]
selected_candidate_ids: ["prompt-001a"]
weight:
  exposure:
    method: "equal_within_declared_strata"
    factors: []
  priority:
    method: "withheld_pending_human_approval"
    factors: []
statistics:
  cluster_unit: "canonical_cell_id"
  pilot_status: "pending"
limitations:
  - "Estimates are conditional on this panel and do not represent a probability sample."
approvals: []
waivers: []
```

Do not substitute prose references such as `accepted_ids_ref` for `selected_candidate_ids`. All selected IDs must resolve to QA-pass candidates.

Use equal weights when exposure evidence is absent. State that equal weighting is conditional, not prevalence.

## Human prompt-table columns

The comprehensive table in `panel_report.md` uses:

| Column | Meaning |
| --- | --- |
| Prompt ID | Stable candidate ID |
| Exact prompt | Trackable string |
| Variant | Observed-language, paraphrase, or sensitivity |
| Partition | Core, rotating, sentinel, control, aided |
| Band | B0–B5 |
| Aided state | Target, competitor, category, unaided |
| Campaign-exposed | Boolean |
| Buyer job | Stable job ID plus short label |
| Information act | Enum |
| Journey | Enum |
| Funnel | Optional rollup or null |
| Role/persona | Evidence-supported role |
| Locale/language | Locale and language |
| Constraints | Material cell constraints |
| Expected answer | Answer-kind label |
| Turn form | Single or scripted multi-turn |
| Lanes/surfaces | Eligible measurement conditions |
| Evidence | Grade plus source IDs |
| Transformation | Provenance enum |
| Weight status | Exposure and priority basis, never one ambiguous score |
| QA | Decision and rule/review note |

## Run manifest

`run_manifest_template.json` records per observation:

- exact prompt and configuration hashes;
- provider model/version as exposed;
- surface, lane, locale, account/personalization state;
- search/tool policy and whether retrieval ran;
- generated queries and citation metadata when exposed;
- temperature/sampling controls and approved system/developer-prompt IDs and hashes;
- fresh/session-state metadata and randomized order seed;
- response/citation payload hashes;
- timestamp, retry/validity status, parser version.

API and consumer-surface observations never share a silent rollup.

Raw system/developer prompts, provider-hidden instructions, account-linked
session state, secrets, and PII are denied by default. Apply an explicit
allowlist and redaction pass before storage; use hashes and non-sensitive
metadata for unavailable or sensitive content. Any explicitly approved raw
diagnostic context is access-restricted, records its approval and purpose, and
expires within 30 days. Never commit it in eval artifacts.

The observation template uses these exact field names: `exact_prompt`, `configuration_hash`, `provider`, `model`, `surface`, `lane`, `locale`, `search_policy`, `retrieval_used`, `session_state`, `response_payload_hash`, `citation_payload_hash`, `timestamp`, `retry_status`, `validity_status`, and `parser_version`. Additional fields are allowed; aliases do not satisfy this contract.

## Completion checklist

- All IDs are unique and references resolve.
- All JSON parses as JSON; `contamination_register.yaml` and `panel.yaml` parse as YAML 1.2.
- Provisional missing hashes are `null` plus warnings; frozen hashes are real SHA-256 values.
- Every prompt traces to a job, canonical cell, and permitted evidence.
- Material claims retain source spans and assertion class.
- Every supported dimension is covered; every gap has a waiver/research need.
- Unaided core has no target/campaign leakage.
- Grade-D prompts are rotating/quarantined unless explicitly promoted.
- Exact duplicates are merged; semantic pairs retain judgment.
- QA decisions and derived counts exactly reconcile with the universe and accepted-ID list.
- Budget and allocation agree.
- Lanes and aided statuses have separate denominators.
- Exposure and priority weights are separate and sourced or explicitly equal.
- Variants/repetitions stay nested under cells.
- Conditional uncertainty and coverage limits are visible.
- Human gates are recorded and resumable.
- Frozen history is append-only.
