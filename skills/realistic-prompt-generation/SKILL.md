---
name: realistic-prompt-generation
description: "Generate natural, controlled prompt variants from a target-blind design brief and prompt architecture while preserving approved jobs, acts, journeys, constraints, roles, locales, proximity bands, and evidence language. Use after architecture design and before contamination or semantic QA."
---

# Realistic Prompt Generation

Write authentic prompts without manufacturing recommendation opportunities.

This skill inherits the ethical floor from `skills/ETHICS.md`. It enforces anti-hallucination and evidence-bound language. Anti-spray and human-send are not applicable.

## Hard context boundary

For unaided generation, accept only:

- anonymized segment and role labels;
- approved jobs, constraints, information acts, journey states, and locales;
- short source-language fragments safe under the permitted-data rules;
- evidence IDs and grades;
- `prompt_architecture.json`;
- style and turn-form requirements.

Do not accept or inspect:

- target brands, products, domains, people, slogans, campaign terms, proprietary categories, or flattering claims;
- current AI answers, rankings, mentions, citations, gaps, or target pages;
- the contamination register itself.

If those fields appear, stop unaided generation and request a sanitized `blind_design_brief.json`. When subagents or fresh sessions are available, generate in a fresh context that receives only the blind brief and architecture.

The exception is an explicitly separate B0 aided pass. It may receive only the target aliases needed by approved B0 cells.

## Preserve the canonical intent cell

For each architecture cell, hold constant:

- underlying job;
- journey state and information act;
- material constraints;
- persona/role and locale;
- proximity band;
- expected answer kind.

Create two core variants by default:

1. the closest natural rendering of observed language;
2. a natural paraphrase that preserves the same intent.

Use additional variants only for a wording-sensitivity pilot or rotating discovery. Do not create a full style × persona × locale grid.

Use `variant_role: observed_language` only when the candidate is verbatim or lightly normalized from a cited behavioral/query language sample. `search_query_expanded`, `human_written`, and `llm_expanded` candidates are `natural_paraphrase` or `sensitivity`; a source ID does not by itself make generated wording observed.

## Write realistic prompts

Reflect evidence-supported styles:

- concise;
- contextual;
- imperfect but intelligible;
- natural follow-up;
- separately scripted multi-turn only when evidence supports a journey.

Avoid polished persona exposition such as “As a forward-thinking CFO at a 120-person professional-services firm...” Use only context a real person needs to get a useful answer.

Every candidate must record one transformation:

- `verbatim`;
- `lightly_normalized`;
- `search_query_expanded`;
- `human_written`;
- `llm_expanded`;
- `translated`;
- `locale_transcreated`.

Do not assign observed frequency to a generated prompt. `llm_expanded` remains evidence grade `D` until independently validated or explicitly promoted.

## Band rules

- Brand (`B0`) names the supplied target alias and stays `target_aided`.
- Shortlist (`B1`) may ask compare/recommend/buy only when the job evidence supports that act.
- Category (`B2`) may name only an accepted evidence-supported category.
- Problem (`B3`) supplies the problem/need, not the category.
- Goal (`B4`) supplies the outcome/job, not a product or category.
- Market (`B5`) needs a fresh evidence ID and a review-by date.

Use the codes in `prompt_universe.json` and the names in the Markdown generation summary.

Never:

- append “what tools should I use?” merely to force brands;
- convert “how do I solve this?” into “which platform should I buy?”;
- add a product/category not entailed by evidence;
- use answer-derived wording in core;
- machine-translate a core locale and call it transcreated;
- copy prompts from public conversational corpora.

Core non-default locales require native or market-competent review. Until reviewed, set `locale_review_status: pending` and keep the candidate outside core.

## Output

Present a short Markdown generation summary first: cell coverage, variant counts, style mix, grade-D share, locale-review gaps, and any architecture cells that could not be rendered without guessing.

Then write `prompt_universe.json`:

```json
{
  "schema_version": "1.0.0",
  "artifact_id": "universe-<stable-slug>",
  "created_at": "RFC3339",
  "created_by": "declared agent or human",
  "source_manifest_hash": null,
  "warnings": ["hash_not_computed: compute all hashes before freeze"],
  "blind_brief_hash": null,
  "canonical_cells": [
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
          "variant_role": "observed_language | natural_paraphrase | sensitivity",
          "text": "Natural user prompt",
          "language": "en",
          "locale": "en-CA",
          "transformation": "lightly_normalized",
          "source_ids": ["source-001"],
          "evidence_grade": "A",
          "locale_review_status": "not_required | pending | approved",
          "generation_provenance": {
            "model": "declared model",
            "prompt_hash": null
          }
        }
      ]
    }
  ]
}
```

Every canonical cell repeats every flat dimension shown above, copied from its architecture cell. Never shorten the record, move dimensions into a nested object, or add compensating aliases. Do not repair missing jobs or architecture fields. Return unresolved cells to their owning atom.

## Handoff

Pass the prompt universe, architecture, safe evidence excerpts, and separately held contamination register to `prompt-set-qa`. Do not expose baseline visibility.
