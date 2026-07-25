---
name: icp-evidence-analysis
description: "Turn a source-bound company and market dossier into testable ideal-customer-profile hypotheses, buying roles, triggers, constraints, disqualifiers, standing, counterevidence, and research gaps. Use when a URL, company description, monitor profile, or evidence manifest needs to become defensible ICP inputs for buyer research or an AI-visibility prompt panel."
---

# ICP Evidence Analysis

Build hypotheses from evidence. Do not write persona fiction.

This skill inherits the ethical floor from `skills/ETHICS.md`. It enforces anti-hallucination and permission checks. Anti-spray and human-send are not applicable because this skill produces research artifacts, not outreach.

## Inputs

Accept any combination of:

- a company URL and user-supplied description;
- `source_manifest.json`;
- public product, pricing, integration, security, support, certification, filing, review, or coverage pages;
- an existing Newsjack monitor profile as unverified leads;
- target markets, exclusions, permitted-data rules, or fact-check results.

If only a URL and description are supplied, research enough independent public evidence to test the website's claims. Never treat a target's own positioning as buyer behavior.

## Build the evidence perimeter

For every material claim, retain:

- source ID, canonical URL or authorized file;
- title/publisher, access or publication date, and source type;
- exact span or a faithful short paraphrase;
- whether it is `company_asserted`, `buyer_behavior`, or `independent`;
- fact type, confidence, and permission status.

Keep negative and conflicting evidence. Stop if a source is private, unlawfully obtained, outside the user's permission, or contains personal data that is not needed.

Separate:

- products, capabilities, limitations, integrations, proof assets, and geographies;
- declared segments from independently supported demand;
- exact brand, product, domain, people, slogan, proprietary-category, campaign, and flattering-claim terms for the later contamination register;
- competitors named by evidence from alternatives merely guessed by the model.

Classify by publisher provenance, not by the claim's tone. A target-authored postmortem, benchmark, customer story, or technical article remains `company_asserted`; split it from any third-party analysis instead of blending both into one `independent` record.

Use `fact-check` when a material external claim lacks primary or independent support.

## Form ICP hypotheses

Create testable contexts, not demographic biographies. Each hypothesis must include:

- organization, team, household, or user context;
- triggering condition or struggling moment;
- likely user, champion, economic buyer, approver, blocker, and post-purchase user when evidence supports them;
- constraints, disqualifiers, and geography;
- company capability that establishes standing;
- supporting evidence, counterevidence, confidence, and open questions.

A core-panel ICP needs at least one independent or behavioral source. A website-only ICP stays `low` confidence and `hypothesis_only` unless a human explicitly promotes it.

Do not:

- invent age, income, title, company size, maturity, or buying committee;
- assume current customers define the target population;
- infer market size or prevalence from source counts;
- invert product features into buyer jobs;
- erase evidence that contradicts positioning;
- decide prompt wording, weights, or funnel stages.

## Confidence

- `high`: multiple relevant sources, including direct behavior or strong independent evidence, agree.
- `medium`: one strong source or several consistent weaker sources; material gaps remain.
- `low`: company assertion, sparse proxy evidence, or conflict dominates.

Use `null` for unsupported fields. Never fill a schema slot with a plausible guess.

## Output

Give the human a concise Markdown summary first:

1. factual perimeter and permission status;
2. supported ICPs and buying roles;
3. counterevidence and exclusions;
4. open research questions;
5. Gate 1 decision: `ready_for_human_review`, `needs_research`, or `stop_permission_failure`.

Then write `icp_hypotheses.json` with the shared envelope:

```json
{
  "schema_version": "1.0.0",
  "artifact_id": "icp-<stable-slug>",
  "created_at": "RFC3339",
  "created_by": "declared agent or human",
  "source_manifest_hash": null,
  "warnings": ["hash_not_computed: compute source_manifest_hash before freeze"],
  "gate_status": "ready_for_human_review",
  "icps": [
    {
      "icp_id": "stable-id",
      "label": "Evidence-bound context label",
      "context": {},
      "triggers": [{"text": "Observed trigger", "source_ids": ["source-001"]}],
      "roles": [{"role": "champion", "label": "controller", "source_ids": ["source-001"]}],
      "constraints": [],
      "disqualifiers": [],
      "standing_claim_ids": ["claim-001"],
      "supporting_source_ids": ["source-001"],
      "counterevidence_source_ids": [],
      "confidence": "high | medium | low",
      "status": "supported | hypothesis_only | excluded",
      "open_questions": []
    }
  ]
}
```

Every material field must trace to source IDs or be `null`. Preserve source spans in `source_manifest.json`; do not duplicate long excerpts here.

## Handoff

After human Gate 1, pass approved ICP IDs, rejected IDs, permitted source IDs, target markets, and unresolved questions to `buyer-job-intent-analysis`. Do not silently promote hypotheses.
