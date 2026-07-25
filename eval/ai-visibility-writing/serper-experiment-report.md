# $25 Serper feasibility experiment — findings

Run date: 2026-07-22

Protocol: `ai-visibility-serper-feasibility-v1`

Frozen apparatus commit: `2277f48`

## Decision

**Serper passes as a fast, inexpensive Google organic-search collector, but it
fails this experiment's stability bar for a matched organic control and fails
as a direct AI-answer instrument. Do not use Serper Search results as a proxy
for Google AI Overview or ChatGPT citations.**

The experiment does not change the writing-lever ranking. L4, descriptive
coherent structure, remains the earlier post-retrieval simulation leader, but
there is still no supported live-effect winner among L1–L10.

## What ran

The frozen 138-query corpus was deduplicated across calibration, main, repeat,
and fresh phases. Every query was sent to Serper Search twice with fixed
US/English settings and a top-10 depth. The 276 new SERPs were joined to the
existing 224 Google AI Overview and 224 ChatGPT observations.

| Measure | Result |
|---|---:|
| Unique queries | 138 |
| Serper waves / responses | 2 / 276 |
| Serper organic result rows | 2,518 |
| Existing AI observations | 448 |
| Serper-to-Google-AIO response-wave matches | 448 |
| Serper-to-ChatGPT response-wave matches | 448 |
| Serper-to-DataForSEO-organic matches | 448 |

The raw provider responses remain in ignored, deletable run storage. The
committed analysis records hashes for all 276 responses, the request ledger,
the source manifests, and the joined observation files.

## Findings

### 1. Operationally excellent

Serper returned 276/276 valid responses with no malformed results. Responses
contained a median of nine organic results. Median latency was 878 ms and p95
latency was 1.51 seconds.

That passes the preregistered operational gate. It is useful for cheap discovery
and for collecting a broad organic candidate pool.

### 2. The top 10 was not stable enough to be the sole control

The two waves for the same query were only 39.6 seconds apart at the median,
yet their median URL Jaccard overlap was 0.385. Median domain overlap was 0.436.
Only 24.6% of queries reached the preregistered 0.70 URL-overlap threshold.

| Stability measure | Result | Preregistered bar |
|---|---:|---:|
| Within-Serper median URL Jaccard | 0.385 | 0.70 |
| Within-Serper median domain Jaccard | 0.436 | — |
| Serper vs DataForSEO median URL Jaccard | 0.333 | 0.50 |
| Serper vs DataForSEO median domain Jaccard | 0.444 | — |

The endpoint was operationally reliable, but the returned ranking was not
stable enough for a single Serper draw to stand in as the counterfactual
retrieval set. This is especially important because [Serper describes its
results as real-time and uncached](https://serper.dev/): repeated measurement
captures real result variation rather than a fixed snapshot.

### 3. Organic search has low recall for AI citations

No response contained an explicit AI Overview, AI-mode answer, or AI-citation
field. Sixteen responses contained a conventional `answerBox`; those were not
relabeled as AI answers.

| AI citation found in Serper organic top 10 | URL recall | Domain recall |
|---|---:|---:|
| Google AI Overview | 24.2% | 35.9% |
| ChatGPT | 8.6% | 16.8% |

So Serper can tell us what appeared in its ordinary Google result set. It cannot
tell us which pages an AI system retrieved and rejected, and absence from its
top 10 cannot be treated as absence from an AI system's candidate pool.

### 4. The owned-site publishing test was not runnable

The live preflight found no crawlable owned-page pool at `newsjack.sh`.
Command-line clients receive the installer, browser clients redirect to GitHub,
and `/robots.txt` and `/sitemap.xml` return 404 pages. Without eligible live
pages, there is nothing to randomize into L4 treatment versus unchanged control,
and no valid way to measure an edit moving an AI answer.

No page was edited, pushed, deployed, or submitted for indexing.

## Budget reconciliation

The approved $25 was treated as a hard ceiling, not a spending target.

| Budget item | Result |
|---|---:|
| Approved cap | $25.00 |
| Serper credits consumed | 276 |
| Serper replacement value at Starter pricing | $0.276 |
| Incremental cash charge | $0.00 |
| Incremental DataForSEO spend | $0.00 |
| Existing AI corpus cost, reused | $8.9807195 |
| Prior corpus plus Serper replacement value | $9.2567195 |

Serper's published Starter rate is [$1 per 1,000 queries, sold as a $50
top-up](https://serper.dev/). The run used existing credits and made no purchase.
Spending more would not fix the missing AI fields, organic instability, or the
absence of publishable treatment pages, so collection stopped at the
preregistered sample.

## Consequence for the skill

The safe product claim remains unchanged:

- The skill can audit and improve writing while preserving facts.
- L4 and L5 are reasonable hypotheses for a real publishing test.
- No lever is proven to increase AI retrieval, mentions, citations, or ranking.
- Serper is suitable as an auxiliary organic discovery source, not the AI
  outcome source and not a hidden-retrieval oracle.

The next dollar should be spent only after there is a crawlable set of owned
articles with stable URLs, indexing, treatment/control assignment, and enough
time for repeated post-publication measurements. Until then, the unspent budget
should remain unspent.
