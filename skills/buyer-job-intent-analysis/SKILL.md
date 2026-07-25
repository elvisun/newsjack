---
name: buyer-job-intent-analysis
description: "Recover source-bound buyer jobs, struggling moments, desired progress, forces, workarounds, information acts, journey states, criteria, constraints, roles, locales, and authentic language. Use on approved ICP hypotheses plus customer, search, review, forum, procurement, support, or public-market evidence before designing an AI-visibility prompt architecture."
---

# Buyer Job Intent Analysis

Recover what people are trying to accomplish and how they express it. Do not turn product features into imagined demand.

This skill inherits the ethical floor from `skills/ETHICS.md`. It enforces anti-hallucination, permission, provenance, and decay-aware handling of current market evidence. Anti-spray and human-send are not applicable.

## Inputs

Require:

- approved `icp_hypotheses.json`;
- `source_manifest.json` with permission and provenance;
- any user-supplied transcripts, queries, reviews, support material, or market sources.

Research missing public-market language when permitted. Prefer evidence in this order:

1. lawfully collected relevant AI conversations with collection metadata;
2. customer or prospect interviews and calls;
3. on-site search, support, chat, sales, and win/loss evidence;
4. paid-search, Search Console, marketplace, and site-search queries;
5. public reviews, forums, communities, RFPs, procurement guides, and competitor reviews;
6. broad search and People Also Ask proxies;
7. company copy;
8. LLM expansion.

Public conversational corpora may inform style, turn count, and multilingual naturalness. They must not supply category prevalence or copied prompts.

## Grade evidence

| Grade | Meaning | Eligible use |
| --- | --- | --- |
| `A` | Direct relevant behavior or verbatim customer/prospect language with provenance | Wording and exposure-weight inputs |
| `B` | Credible public-market behavior or search proxy with provenance | Wording with an explicit proxy label |
| `C` | Company assertion or expert hypothesis | Research hypothesis; approval required |
| `D` | LLM-generated expansion without independent support | Rotating discovery only |

A count is a count within the supplied corpus. Never relabel it market frequency.

## Extract jobs and language

For each supported ICP, extract:

- struggling moment or trigger;
- desired progress or outcome;
- current workaround;
- push, pull, anxiety, and habit forces;
- requested action or information need;
- decision criteria, constraints, and proof sought;
- exploration/evaluation state;
- authentic source-language samples;
- role/persona and locale when known;
- contradictory, negative, or post-purchase evidence.

Keep these axes independent:

- `buyer_job`;
- `information_act`: `explain`, `diagnose`, `plan`, `generate`, `compare`, `recommend`, `verify`, `navigate`, `buy`, `implement`, `troubleshoot`;
- `journey_state`: `problem_identification`, `exploration`, `requirements_building`, `supplier_selection`, `adoption`, `post_purchase`;
- optional `funnel`: `TOFU`, `MOFU`, `BOFU`, or `null`.

Do not infer funnel from a keyword. A direct-brand support question can be post-purchase; an urgent problem can be close to transaction.

Split a source that asks two materially different things into two candidate jobs or flag it for review. Do not collapse user, champion, buyer, approver, blocker, and post-purchase user.

## Build jobs

A job statement should identify the situation, progress, constraints, and affected actor without naming the target product:

> When [situation], help [actor] make [progress] while [material constraints].

Link every component to source IDs. If the model adds a plausible force, criterion, or act, mark it grade `D`, confidence `low`, and `hypothesis_only`.

## Output

Give the human a readable Markdown report first:

- strongest jobs by ICP;
- evidence and exact language worth preserving;
- conflicts, negative cases, and post-purchase jobs;
- missing roles/locales/acts;
- Gate 2 decision: `ready_for_human_review`, `needs_research`, or `stop_permission_failure`.

Then write `buyer_jobs.json`:

```json
{
  "schema_version": "1.0.0",
  "artifact_id": "jobs-<stable-slug>",
  "created_at": "RFC3339",
  "created_by": "declared agent or human",
  "source_manifest_hash": null,
  "warnings": ["hash_not_computed: compute source_manifest_hash before freeze"],
  "gate_status": "ready_for_human_review",
  "jobs": [
    {
      "job_id": "stable-id",
      "icp_ids": ["icp-001"],
      "statement": "When ..., help ...",
      "struggling_moment": "Evidence-bound trigger",
      "desired_progress": "Evidence-bound outcome",
      "workarounds": [],
      "forces": {"push": [], "pull": [], "anxiety": [], "habit": []},
      "information_acts": ["diagnose", "compare"],
      "journey_states": ["problem_identification", "requirements_building"],
      "criteria": [],
      "constraints": [],
      "roles": [],
      "language_samples": [
        {
          "text": "Short source language",
          "source_id": "source-001",
          "span": "locatable span",
          "locale": "en-CA",
          "evidence_grade": "A"
        }
      ],
      "supporting_source_ids": ["source-001"],
      "counterevidence_source_ids": [],
      "evidence_grade": "A",
      "confidence": "high",
      "status": "supported | hypothesis_only"
    }
  ]
}
```

## Handoff

After human Gate 2, pass only approved jobs, source-language fragments, roles, locales, constraints, grades, and evidence IDs into the blind-design step. Do not pass target names, slogans, desired target pages, current AI answers, or visibility scores.
