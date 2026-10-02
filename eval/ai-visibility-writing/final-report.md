# AI Visibility Writing pilot — final report

## Decision

**INCONCLUSIVE. No supported efficiency winner. Do not claim or scale efficacy
from this pilot.** Ship the public skill only as an evidence-limited,
fact-preserving editorial audit and use a controlled publishing experiment to
test live impact.

L4, descriptive coherent structure, was the most efficient qualifying *atomic
post-retrieval simulation candidate*: +25 percentage points for citation
selection and accurate use, +0.125 human-quality delta, zero judged fidelity
delta, 15.6% median token change, and efficiency index 15.98. It is not the
overall winner because the pilot gates failed and fresh observational evidence
did not confirm it.

## Scores and gates

| Measure | Result | Bar | Outcome |
|---|---:|---:|---|
| Composite | 59.79 | 65 | Fail |
| Audit factor F1 | 0.521 | 0.70 | Fail |
| Blind quality win rate | 0.688 (95% CI 0.547–0.801; n=48) | 0.60 | Pass |
| Combined citation-selection win rate | 0.438 (95% CI 0.174–0.741; n=8) | 0.55 | Fail |
| Readability pass rate | 1.000 (95% CI 0.926–1.000; n=48) | 0.90 | Pass |
| Zero invented facts / judged fidelity | False | True | Fail |
| Schema complete | 8/12 combined pairs | 12/12 | Fail |
| Tests, linter, ledger | Green after collection | Green | Pass |

There is no fresh skill composite. The frozen apparatus had only nine model-call
slots left after development and no valid fresh simulation design that could
re-estimate all components. The one frozen fresh query draw therefore tested
observational corpus stability only; manufacturing a fresh composite from the
development judgments would have been invalid.

## Data and spend reconciliation

| Phase | Paired units | Exact spend |
|---|---:|---:|
| Calibration | 12 | $0.23975200 |
| Main | 120 | $4.80755575 |
| Repeats | 49 | $1.95635300 |
| Fresh | 49 | $1.97585875 |
| Reserve recovery | 1 Google replacement | $0.00120000 |
| **Total** | **230 paired draws including calibration/repeats/fresh** | **$8.98071950** |

All 98 fresh observations succeeded and contained citations. The development
main set contained 120 ChatGPT and 120 Google responses; Google returned an AI
Overview with citations for 113. Feature extraction retained explicit robots,
fetch, non-HTML, and size exclusions. No unknown cost remains. The DataForSEO
account was user-confirmed free-credit-only with no payment method; local phase
and total hard gates remained active despite the provider API reporting a
$1,000 daily limit.

The simulation ledger used 441/450 allowed invocations: 437 successes and four
errors. All 48 L1–L6 atomic pairs completed. Combined-skill exclusions were
`energy-01`, `energy-03`, `travel-01`, and `logistics-01`; their rewrites either
strengthened scope, added unsupported detail/advice, or failed the frozen
protected-item guard. No downstream citation trials were run for them.

## Observational evidence

None of the six preregistered Google writing features passed q<0.10 in either
draw. Development → fresh estimates in percentage points per standard deviation
were: F1 +1.4 → +0.7; F2 −1.4 → −0.8; F3 +4.6 → −1.6; F5 −7.3 → −21.9; F8
−1.7 → −0.7; F10 +2.8 → +4.9. All preregistered fresh q-values were at least
0.120. These are associations among returned organic pages, not edit effects.

The exploratory intent-fit format proxy F6 was +16.5 pp/SD in development but
−11.7 fresh. That reversal rejects a universal table/list/steps rule. ChatGPT
factor effects remain unestimable without a retrieved-but-rejected risk set.

Cross-platform source overlap stayed low: URL Jaccard 0.023 development and
0.035 fresh; domain Jaccard 0.045 and 0.066. Repeat URL overlap was also modest:
0.326 ChatGPT and 0.298 Google. One response is not a stable ranking.

## Full lever ranking

1. **L4 Descriptive coherent structure** — atomic leader; +25/+25 pp; efficiency
   15.98; provider-heterogeneous and observationally negative; not a winner.
2. **L5 Neutral qualified specificity** — +12.5/+12.5 pp; efficiency 12.50;
   provider-heterogeneous; not supported overall.
3. **L6 Remove repetition and padding** — 0/+12.5 pp; inconsistent citation
   direction; insufficient.
4. **L1 Direct scoped answer** — 0/0 pp; improved human quality; insufficient for
   citation selection.
5. **L2 Evidence context and attribution** — 0/0 pp with −0.125 fidelity delta;
   insufficient.
6. **L3 Unique verifiable information** — 0/0 pp; the fact-bounded fixtures
   could not supply real new evidence; insufficient test.
7. **L7 Clarify entities** — not atomically tested; observationally null.
8. **L8 Intent-fit format** — not atomically tested; exploratory direction
   reversed fresh.
9. **L9 Expose a legitimate date** — not atomically tested; no supported effect.
10. **L10 Preserve specialist precision** — human-quality guard, not an efficacy
    intervention.

## Thesis disposition

Supported as product safeguards: H10's outcome separation, H11's cross-platform
instability, fact fidelity, precision preservation, anti-stuffing, and explicit
causal boundaries.

Promising only for a larger causal test: H5/L4 coherent structure and H8/L5
qualified specificity because their atomic simulations improved without mean
quality harm. They are not live-effect findings.

Insufficient: H1 direct answers, H2 attribution, H3 unique information under a
fact-bounded fixture, and H15 repetition removal as a citation lever.

Rejected as universal advice: the exploratory format/table signal and any claim
that one engine's source behavior transfers directly to another. H0—the view
that retrieval/rank/domain/topic effects dominate and writing adds little—was
not rejected.

## Causal claims

The pilot may claim that the frozen skill often improved judged editorial
quality and that L4 led a constructed post-retrieval source-selection
simulation. It may not claim that editing a live page with L4, any other lever,
or the full skill raises AI rankings, retrieval, mentions, citations, traffic,
or coverage.

## Highest-leverage next steps

1. Run a preregistered, randomized publishing experiment on eligible owned pages
   using L4 versus unchanged controls, stratified by document type and query.
2. Hold URL, crawl/index state, internal links, evidence, promotion, and
   technical SEO constant; verify indexing before measuring.
3. Repeat matched Google and ChatGPT queries over multiple weeks and measure
   retrieval, mention, citation, accurate use, and human engagement separately.
4. Use enough pages to estimate provider and document-type interactions; do not
   reuse this pilot's queries as a tuning set.
