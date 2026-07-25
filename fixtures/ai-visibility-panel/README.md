# AI visibility panel fixtures

This directory tests the public AI visibility panel skills without relying on a
real client, private data, or a model runner. All companies, people, quotations,
domains, products, competitors, and source records are synthetic.

## Contents

- `inputs/ledgerlift/` is a B2B buying-committee fixture. It includes conflicting
  evidence, an ERP constraint, observed finance language, a post-purchase job,
  a campaign-slogan trap, and a low-fit segment.
- `inputs/harborheat/` is a local consumer-service fixture. It includes
  regulation and electrical constraints, English and French Canadian language,
  a material locale boundary, post-purchase troubleshooting, and a fresh B5
  public-story source.
- `gold/qa-adversarial.json` is the 18-case predeclared semantic QA gold set. It covers
  lexical and semantic contamination, legitimate shared wording, exact and
  semantic duplicates, protected constraints, recommendation forcing,
  answer-derived wording, aided-state errors, unsupported grade-D core cells,
  and locale-review failures.
- `grade_output.py` is a Python-standard-library validator for a completed
  pipeline output directory and an optional grader for the adversarial set.

The inputs are evidence dossiers, not expected final panels. A pipeline may
make different judgment calls if it cites the evidence and preserves the
required gaps. `expected-coverage.json` states invariant coverage and
guardrails, not exact prompt wording.

## Run a fixture

Give the orchestration skill the fixture's `request.json` and
`source_manifest.json`. Write its artifacts to a disposable run directory
outside this fixture tree, for example:

```text
/tmp/newsjack-ai-panel-ledgerlift/
```

Then validate the full run:

```bash
python3 fixtures/ai-visibility-panel/grade_output.py \
  /tmp/newsjack-ai-panel-ledgerlift
```

The validator requires the fourteen files in the orchestration artifact
contract. It parses JSON and the conservative YAML subset used by the panel
contracts. It returns exit code `0` only when no errors are found. Warnings
identify claims that still require human or semantic review.

## Run the adversarial QA evaluation

Run `prompt-set-qa` over the cases in `gold/qa-adversarial.json`, place the
  result at `<run-dir>/prompt_qa.json`, and grade only the semantic decisions:

```bash
python3 fixtures/ai-visibility-panel/grade_output.py \
  <run-dir> \
  --gold fixtures/ai-visibility-panel/gold/qa-adversarial.json \
  --qa-only
```

For an end-to-end adversarial run, omit `--qa-only`; structural validation and
gold scoring both run.

The gold labels were declared before tuning. A passing QA evaluation requires
all case IDs, exact statuses, required rule outcomes, duplicate actions, and
routes to match. The grader reports per-category accuracy so a strong lexical
scanner cannot hide weak semantic or locale judgment.

## What the validator proves

The deterministic checks cover:

- required files, common envelopes, stable IDs, SHA-256-shaped frozen hashes,
  and warned `null` hash blockers on provisional runs;
- permitted, classified source evidence with locatable spans;
- source, ICP, job, architecture, cell, candidate, QA, and panel references;
- enum consistency and one QA decision per candidate;
- target and campaign lexical leakage in unaided core;
- aided-state and campaign/core contradictions;
- accepted grade-D core prompts;
- accepted untranslated or unreviewed locale variants;
- accepted normalized duplicate strings;
- blind-brief lexical contamination;
- panel candidate references, separate exposure/priority weights, conditional
  limitations, and exact prompts in the human report.

The validator does not claim to determine naturalness, evidence entailment,
semantic leakage, recommendation forcing, or same-intent equivalence. Those
are tested against the gold QA cases and remain review judgments in real runs.

Do not commit generated run directories.
