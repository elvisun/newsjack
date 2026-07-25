---
name: prompt-set-qa
description: "Gate a prompt universe for schema and provenance completeness, target or campaign contamination, evidence entailment, naturalness, one-concept clarity, architecture consistency, aided status, answer leakage, and semantic duplicates. Use after realistic prompt generation and before human panel selection."
---

# Prompt Set QA

Decide `pass`, `revise`, `quarantine`, or `reject`. Do not quietly repair upstream work.

This skill inherits the ethical floor from `skills/ETHICS.md`. It enforces anti-hallucination, provenance, blinding, and decay-aware review. Anti-spray and human-send are not applicable.

## Inputs

Require:

- `prompt_universe.json`;
- `prompt_architecture.json`;
- evidence excerpts and source metadata;
- the versioned contamination register;
- deterministic normalization, lexical scan, exact-hash, and similarity-pair results;
- optional blind human decisions.

Reject inputs that include baseline visibility, current rankings, target performance, answer-derived target pages, or selectors' preferred outcomes.

## Run deterministic checks first

Check:

1. schema and provenance completeness;
2. unique IDs and resolved references;
3. target, product, domain, people, slogan, proprietary-category, campaign, flattering-claim, and competitor terms;
4. forbidden answer-derived fields;
5. Unicode normalization, language, length, and one-concept shape;
6. exact normalized hashes;
7. lexical similarity candidate pairs;
8. embedding pairs when a fixed model/version is available;
9. architecture coverage, budget, and aided/lane consistency.

Deterministic target or campaign matches are hard failures in unaided core prompts. B0 target aliases pass only through a declared allowed exception. Embedding similarity may nominate a pair; it may never auto-delete.

## Review semantically

For each candidate, judge:

- evidence-to-prompt entailment;
- naturalness and role/locale authenticity;
- whether both variants preserve one canonical intent;
- proximity, journey, act, expected-answer, and aided-status consistency;
- commercial leading or recommendation forcing;
- semantic slogan or flattering-claim leakage;
- whether answer-derived language entered core;
- whether a similar prompt changes a material constraint.

Protect differences in locale, persona, material constraint, competitor-aided status, information act, journey, and expected answer. Merge only when the job, journey, constraints, and answer kind are materially the same.

Archive every removed variant with evidence and reason.

For health, legal, financial, safety, or other high-stakes domains, keep the prompt inside the measured navigation or information boundary. Quarantine prompts that ask the model to diagnose, determine personal suitability, prescribe, or make another professional judgment unless that judgment is explicitly in scope with qualified review. If a response may contain such advice, code it only under the declared answer-framing rubric; never score the advice as professionally correct without a separate validated protocol.

## Decision rules

- `pass`: supported, natural, correctly classified, uncontaminated, and not redundant.
- `revise`: wording can change without altering upstream facts or cell meaning; route the request to `realistic-prompt-generation`.
- `quarantine`: potentially useful but grade D, answer-derived, semantically suspicious, locale-unreviewed, or awaiting evidence/human decision.
- `reject`: contaminated, unsupported, materially leading, wrong cell, permission failure, or irreparable duplicate.

QA must not:

- invent a source, job, ICP, prevalence, or weight;
- change the architecture to rescue a prompt;
- rewrite a problem into a product recommendation;
- select based on target performance;
- resolve a disputed core/locale/high-weight decision without a human.

## Contamination register

Use:

```yaml
target_terms:
  brands: []
  products: []
  domains: []
  people: []
  slogans: []
  proprietary_categories: []
  campaign_terms: []
  flattering_claims: []
competitor_terms: []
allowed_exceptions:
  - band: B0_direct_brand_product
    term_classes: [brands, products]
```

Run normalized, token, fuzzy, and semantic checks. Ordinary shared words may be legitimate; semantic flags require a recorded human or high-confidence review decision.

## Output

Give the human a Markdown QA report first:

- accepted/revise/quarantine/reject counts;
- contamination failures by class;
- duplicate merges/splits and protected differences;
- coverage gaps created by rejection;
- every disputed core or locale decision for Gate 3.

Then write `prompt_qa.json`:

```json
{
  "schema_version": "1.0.0",
  "artifact_id": "qa-<stable-slug>",
  "created_at": "RFC3339",
  "created_by": "declared agent or human",
  "source_manifest_hash": null,
  "warnings": ["hash_not_computed: compute source_manifest_hash before freeze"],
  "baseline_fields_blinded": true,
  "decisions": [
    {
      "candidate_id": "prompt-001a",
      "status": "pass | revise | quarantine | reject",
      "rule_results": [
        {
          "rule_id": "no-target-term-unaided",
          "status": "pass | fail | review",
          "evidence": []
        }
      ],
      "duplicate_decision": {
        "canonical_cell_id": "cell-001",
        "action": "retain_variant | merge_exact | merge_semantic | split"
      },
      "route_to": "realistic-prompt-generation | prompt-proximity-architecture | buyer-job-intent-analysis | human_gate_3 | null",
      "reason": "Specific evidence-bound reason",
      "review_confidence": "high | medium | low"
    }
  ],
  "accepted_candidate_ids": ["prompt-001a"],
  "counts": {
    "total_candidates": 1,
    "pass": 1,
    "revise": 0,
    "quarantine": 0,
    "reject": 0,
    "accepted": 1
  },
  "gate_status": "ready_for_human_review | needs_revision | stop_permission_failure"
}
```

Keep every candidate, including rejected drafts, in `prompt_universe.json`. Emit exactly one decision for every universe candidate and no unknown decision IDs. Derive `accepted_candidate_ids` and `counts` from the final decision array in one pass: accepted IDs equal exactly the `pass` IDs, and every status count plus `total_candidates` must reconcile. If they do not, the output is invalid; recording the discrepancy in a warning or change ledger does not make it handoff-ready.

## Handoff

Human Gate 3 approves core/aided/campaign partitions and disputed decisions while still blind to baseline visibility. Pass only approved candidate IDs and the full rejection ledger to `ai-visibility-panel-design`.
