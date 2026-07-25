# Pilot findings

Decision: **INCONCLUSIVE — do not scale the efficacy study yet**.

The evidence supports shipping the skill only as a conservative, fact-preserving
editorial audit. It does not support claiming that the skill or any writing
lever improves live AI ranking, retrieval, mentions, citations, or traffic.

## What generalized

- Google and ChatGPT source sets overlapped very little. Mean URL Jaccard was
  0.023 in development and 0.035 fresh; domain Jaccard was 0.045 and 0.066.
- Single-query source selection was unstable: repeat URL Jaccard was 0.326 for
  ChatGPT and 0.298 for Google.
- The frozen skill improved blind human quality versus each generator's stronger
  baseline in 68.8% of tie-weighted comparisons (95% CI 54.7%–80.1%; n=48) and
  passed readability in 48/48 judged cases.
- Separating eligibility/authority limits from prose, preserving facts, and
  refusing guarantees remain sound product safeguards.

## What did not generalize

- None of F1, F2, F3, F5, F8, or F10 passed the preregistered FDR threshold in
  either the development or fresh Google adjusted model. ChatGPT adjusted
  effects were unestimable because it exposes no rejected retrieval set.
- The exploratory L8 format proxy moved from +16.5 pp/SD in development to
  −11.7 pp/SD fresh. It is not a reusable rule.
- L4 led the constructed atomic simulation (+25 pp citation selection and
  accurate use) but was flat for one provider stratum and negative in both
  Google observational draws. It is a candidate for a controlled publishing
  test, not an efficacy winner.
- Four of 12 combined rewrites failed the deterministic fact guard, leaving
  eight citation pairs. The combined tie-weighted citation win rate was 0.438
  (95% CI 0.174–0.741), below the 0.55 bar.
- Audit factor F1 was 0.521, below 0.70. The overall composite was 59.79, below
  65, and fidelity/completeness hard gates failed.

## Corpus and cost

The visible corpus contains 120 main paired units, 49 repeats, 49 fresh paired
units, and 12 calibration pairs. The fresh draw produced 98/98 successful
platform observations with citations. DataForSEO reconciled spend was
$8.9807195: calibration $0.239752, main $4.80755575, repeats $1.956353, fresh
$1.97585875, and reserve $0.0012. The model ledger contains 441 invocations:
437 successes and four recorded errors.

## What is warranted

Warranted: the skill can improve editorial clarity while enforcing factual and
causal boundaries; L4 is the most efficient *constructed atomic simulation*
candidate; platform behavior is heterogeneous and stochastic.

Not warranted: that L4 or the complete skill moves live AI answers; that tables,
headings, dates, citations, or keyword patterns are ranking factors; that an
offline source-selection simulation estimates publishing impact.

## Next test

Run a separate controlled publishing/indexing experiment. Randomize eligible
owned pages within query/document strata to frozen L4 edits versus unchanged
controls; preserve facts, URLs, technical eligibility, internal links, and
promotion; verify crawl/index state; repeat the same Google and ChatGPT queries
over several weeks; and measure retrieval, mention, citation, accurate use, and
human engagement separately. Predeclare sample size and analysis before edits.
