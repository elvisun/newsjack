# Prompt: Score Reverse-Eval Results

Use this prompt after detector runs are complete.

```text
Score the reverse-newsjack eval results.

Inputs:
- seeds.md
- company-candidates.md
- runs/TARGET_ID/candidates.json
- runs/TARGET_ID/run.md when present

Goal:
For each target story, decide whether the originating story was rediscovered by Newsjack in the uncapped/high-limit hit list. Separately record product-ranking position.

Matching guidance:
- Use LLM judgment, not exact string matching only.
- Match on entity cluster, story action, source/date, and distinctive facts.
- Treat syndicated versions and same-story headlines as matches.
- Do not count adjacent industry stories as matches unless they cover the same originating event.

Classification labels:
- candidate_pool_pass
- final_hit
- ranking_low
- ranking_miss
- source_miss
- invalid_seed
- unclear

For each target, record:
- target_id
- recommended_company
- originating_story
- classification
- recall_pass
- rank_bucket
- best_matching_detector_title
- best_matching_detector_url
- best_matching_lane
- rank_or_position
- source_match_evidence
- miss_reason
- bias_tag
- notes

Definitions:
- candidate_pool_pass: originating story is emitted in candidates.json from an uncapped or high-limit detector run.
- final_hit: originating story survives downstream LLM/coarse relevance as keep or monitor_only.
- ranking_low: story is emitted, but below the product-visible rank threshold being tested.
- ranking_miss: story appears in debug.all_scored_signals but not emitted.
- source_miss: story does not appear in emitted candidates or debug.all_scored_signals.
- invalid_seed: story was unsuitable for positive recall testing.
- rank_bucket: one of top_3, top_5, top_10, below_10, not_emitted.

Write:
- results/overlap_matrix.csv
- results/summary.md

In summary.md include:
- total valid targets
- primary recall rate: candidate_pool_pass or final_hit
- downstream hit rate when LLM/coarse relevance has run
- rank bucket distribution
- ranking_miss count
- source_miss count
- biggest miss mode
- source/bias caveats
- recommended next tuning action
```
