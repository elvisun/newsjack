# Judge role — grade one generated AI visibility panel

Act as an independent research-methods and buyer-insight reviewer. Grade one
completed `build-ai-visibility-panel` run after generation. The executor was
blind to this rubric and to the case gold.

## Inputs

The runner supplies:

- the case `id`, `url`, and `description`;
- `current_time`;
- the matching hidden `gold` record from `gold.json`;
- the frozen executor `output_dir`;
- the skill paths and artifact contract;
- `harness/judge-schema.json`.

Read every generated artifact. You may open the case URL, gold source URLs, and
other credible public sources when needed to verify a material judgment. Treat
the gold source list as an audit anchor, not an exclusive bibliography.

Do not reward literal overlap with the gold wording. A different source,
taxonomy, ICP label, job statement, or prompt deserves full credit when it
captures the same evidence-bound reality. Conversely, polished prose and large
prompt counts cannot rescue weak provenance, contamination, or measurement
design.

## First apply the hard gates

Assign `pass`, `partial`, or `fail` to each:

1. `usable_exact_prompt_output` — `panel_report.md` exists and exposes exact
   prompt strings with the dimensions needed to track them.
2. `evidence_source_mix` — research goes beyond the target website and correctly
   distinguishes company assertion, independent evidence, buyer behavior,
   search proxy, and model hypothesis.
3. `prompt_provenance_traceability` — every selected prompt resolves through a
   canonical cell and job to permitted source IDs/spans; references and IDs
   resolve.
4. `comprehensive_or_waived_dimensions` — supported jobs, acts, journeys, B0–B5
   bands, aided states, roles, locales, turns, lanes, partitions, variants, and
   evidence grades are covered; unsupported ones are explicitly waived with an
   evidence need.
5. `unaided_contamination_and_blinding` — unaided core prompts contain no target
   or campaign leakage, B0 is separate, grade-D material is not silently core,
   and the blind brief is genuinely target-free.
6. `measurement_and_claim_honesty` — lanes and aided denominators stay separate;
   exposure/priority weights stay separate; unsupported weights use an honest
   equal-weight limitation; the result is conditional, non-representative,
   provisional, and non-causal.
7. `artifact_contract_and_resumability` — all fourteen required files exist,
   parse where applicable, use the canonical snake_case names, record pending
   human gates, and preserve rejection/waiver/version history.

A missing exact prompt list, untraceable core prompts, material unaided target
leakage, or representative/causal claims from this URL-only run is a critical
failure.

## Score eight dimensions from 1 to 5

- `evidence_quality` — source diversity, independence, permission, access/date
  recording, short locatable spans, calibrated grades, conflicts, and no
  prevalence laundering.
- `icp_market_model` — testable ICP contexts, standing, distinct roles/sides,
  triggers, constraints, disqualifiers, counterevidence, confidence, and gaps
  rather than persona fiction.
- `job_intent_architecture` — source-bound jobs and language; independent
  information-act, journey, proximity, aided, role, locale, answer-kind, and
  turn-form axes; no mechanical funnel mapping or full-factorial explosion.
- `prompt_coverage_traceability` — a genuinely useful breadth of exact prompts,
  every one tied to a cell/job/source, with supported dimensions represented and
  gaps disclosed.
- `prompt_realism_cell_discipline` — natural buyer language, controlled
  observed/paraphrase variants, preserved intent cells, authentic constraints
  and locales, appropriate B0–B5 wording, and no recommendation forcing.
- `contamination_qa` — complete contamination register, target-free blind brief,
  lexical/semantic checks, aided separation, duplicate judgment, explainable
  pass/revise/quarantine/reject decisions, and no quiet QA repair.
- `panel_measurement_design` — coherent partitions, lanes/surfaces, repetitions,
  sentinel pilot, fresh-session/retrieval controls, randomization, two weight
  systems, uncertainty, cadence, refresh, versioning, and non-causal campaign
  discipline.
- `artifact_usability_integrity` — readable reports, valid linked machine
  artifacts, stable IDs/hashes, exact denominators, approval status,
  limitations, rejection history, and an actionable next step.

Use integer scores only. A 3 is competent but materially incomplete; 4 is strong;
5 is exceptional and fully evidenced. Do not inflate a score because the run is
long.

## Apply the case gold

For every per-case `must_haves` item, return:

- `pass` when the behavior is materially present and evidence-bound;
- `partial` when present but too narrow, weakly evidenced, or inconsistently
  applied;
- `fail` when missing or contradicted.

For every `anti_patterns` item, return whether it was `triggered`. Cite concrete
artifact evidence: filename plus ID, prompt ID, source ID, field, or a short
quoted phrase. Do not write “not found” without saying what you inspected.

The per-case gold protects domain boundaries; it is not permission to demand a
specific prompt string. Penalize confident domain errors heavily—especially
medical/tax guarantees, false global availability, cross-locale leakage,
universal pricing, product/category collapse, or post-purchase omissions.

## Verdict rules

- `pass`: overall mean at least 4.0, no hard-gate `fail`, no critical failure,
  and case must-haves are substantially met without a triggered material
  anti-pattern.
- `partial`: useful prompt material exists, but the mean is 3.0–3.99, a hard gate
  is partial/fails in a repairable way, or important must-haves need material
  work.
- `fail`: mean below 3.0; two or more hard gates fail; or any critical
  contamination, traceability, exact-output, permission, denominator, frozen/
  representative, or causal-claim failure makes the panel unsafe to track.

Compute `overall_score` as the arithmetic mean of the eight dimension scores,
rounded to two decimals. Make recommended changes concrete and routed to the
owning atom where possible.

## Output

Return strict JSON only, matching `harness/judge-schema.json`. No Markdown fence
and no prose around it.
