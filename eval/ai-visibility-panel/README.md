# AI visibility panel public-URL eval

This eval tests whether `build-ai-visibility-panel` can start from the minimum
realistic input—a public URL and a plain-language description—and produce a
source-cited, comprehensive, provisional prompt panel without borrowing the
target's marketing language as buyer demand.

The benchmark uses ten public companies and products across B2B software,
infrastructure, fintech, healthcare, local marketplaces, travel, wellness, tax,
energy, and automotive. The cases were checked on 2026-07-25. Live pages can
change, so the evidence manifest produced by each run must record access times
and uncertainty.

This directory contains the method, inputs, reproducible Opus 5 runners, and
dated run outputs selected for review.

## Files

- `cases.json` — executor-visible case inputs. Each case contains only the target
  URL and description. Global metadata supplies the current-time anchor, skill
  paths, and output-directory convention.
- `gold.json` — executor-hidden evaluation anchors: corroborating public URLs,
  per-case must-haves, and anti-patterns. These are not reference answers or
  required prompt strings.
- `harness/executor.md` — clean-context instructions for the agent building the
  panel.
- `harness/judge.md` — post-generation judging method.
- `harness/judge-schema.json` — strict JSON Schema for one case verdict.
- `scripts/run-opus5.sh` — clean-context exact-model executor runner.
- `scripts/judge-opus5.sh` — fresh, tools-off exact-model judge runner.
- `scripts/grade.py` — deterministic artifact and case-assertion grader.
- `scripts/aggregate.py` — suite summary and verdict renderer.
- `results/2026-07-25-opus5.md` — development, retest, and untouched held-out
  results with the observed failures and limitations.
- `runs/` — dated, hashed execution and judging artifacts.

Run directories retain panel artifacts, manifests, deterministic grades,
aggregates, and structured verdicts. Raw model streams, assembled prompts, and
logs are ignored: they are large, can reproduce fetched page text, and are not
needed to review the scored result.

## Blinding contract

The executor receives exactly:

1. one case's `url`;
2. one case's `description`;
3. the seven skill paths listed in `cases.json`;
4. `current_time`;
5. a new case-specific `output_dir`.

It must not receive `gold.json`, the judge prompt, rubric dimensions, expected
sources, another case, or a prior run. The executor prompt also forbids reading
anything under `eval/ai-visibility-panel/`; a runner should assemble the prompt
outside the agent's context and expose only the permitted values.

The judge runs only after the executor has finished and the output directory has
been frozen. It receives the case, that case's hidden gold record, the generated
artifacts, the rubric, and the current-time anchor. The gold sources are evidence
anchors, not an exclusive source list. A different credible source or a
different natural prompt can earn full credit when it supports the same buyer
reality.

## Evaluation sequence

1. Create a clean, empty output directory for one case.
2. Start a fresh agent context. Supply `harness/executor.md` plus only the five
   allowed inputs above.
3. Let the executor research the live public web and run the skill family end to
   end. Do not seed it with expected sources or prompt ideas.
4. Freeze the output directory and record the executor model/version, tool
   policy, skill hashes, start/end times, and artifact hashes outside the panel
   artifacts.
5. Start a separate judge context. Supply `harness/judge.md`, the generated
   artifacts, the public case input, the matching hidden gold record, and
   `harness/judge-schema.json`.
6. Validate the verdict against the schema before aggregation.
7. Repeat cases in fresh contexts. For the requested Claude test drive, address
   the exact model ID `claude-opus-5`; do not reuse conversation state between
   cases.

Do not let the executor and judge share scratch files, browsing history, or
conversation context. Do not tune the skill on a failed run and then count the
same run as independent evidence.

## What the judge measures

The judge scores eight dimensions:

- evidence quality and source classification;
- source-bound ICP and market modeling;
- job, intent, journey, role, locale, and proximity architecture;
- breadth and traceability of the exact prompt list;
- prompt naturalness and canonical-cell discipline;
- unaided/aided contamination control and blinding;
- lane, partition, weighting, uncertainty, and refresh design;
- artifact integrity and practical usability.

Several rules are hard gates. A polished report cannot pass if it lacks exact
prompts, cannot trace prompts to evidence, contaminates unaided core prompts,
conflates lanes or aided denominators, or presents a URL-only result as frozen,
representative, or causal.

“Comprehensive” is conditional on evidence. The executor should cover every
supported dimension and name every unsupported dimension with the evidence
needed to add it. A fabricated Cartesian grid is worse than an explicit gap.

## Interpreting results

- `pass` means the output is useful as a provisional, conditional panel and no
  hard gate failed.
- `partial` means the panel has usable material but needs material research,
  coverage, QA, or artifact repair.
- `fail` means the output is unsafe as a tracking design, usually because of
  contamination, unsupported claims, missing exact prompts/provenance, or
  denominator/measurement errors.

The joint result is authoritative: a run passes only when deterministic
contract checks pass and the semantic judge returns `pass`. A high semantic
score cannot waive malformed YAML, broken references, count drift, or another
machine-contract failure.

Scores are conditional on these ten cases. They do not establish that the panel
represents all buyers, all prompts, or all AI surfaces.
