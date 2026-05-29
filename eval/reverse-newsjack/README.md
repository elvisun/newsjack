# Reverse Newsjack Eval

Reverse eval tests whether Newsjack can rediscover known current stories when given a plausible company profile with real standing.

## Files

- `RUNBOOK.md`: daily setup and execution workflow.
- `prompts/`: reusable prompts for seed harvesting, company selection, detector runs, and scoring.
- `runs/YYYY-MM-DD-source/`: committed durable eval runs with profiles, detector artifacts, results, and run notes.
- `targets-*.md`: dated target story sets.
- `company-candidates-*.md`: dated reverse-profile company candidates.

## Method

1. Select seed stories from an external trending/news surface.
2. Work backward to define a plausible company profile for each story.
3. Run the detector with that profile.
4. Grade whether the originating story appears in the high-limit detector output.
5. Separately record where it ranks in the user-facing list.

## Outcome Labels

- `candidate_pool_pass`: originating story is emitted in the uncapped or high-limit candidate pool.
- `final_hit`: originating story survives downstream LLM/coarse relevance as a hit.
- `ranking_low`: story is emitted, but below the product-visible rank threshold being tested.
- `ranking_miss`: story is scored but not emitted.
- `source_miss`: story is not scored at all.
- `invalid_seed`: seed story is not a legitimate newsjack target because of brand safety, weak standing, or poor story identity.

## Bias Notes

- If the seed came from a source the detector directly ingests, tag the run as `same_source_assisted`.
- If the seed source is not configured for the tested profile/source setup, tag the run as `independent_seed`.
- Track miss type separately from pass rate. Source misses imply retrieval/source coverage problems; ranking misses imply scoring/filtering problems.

## Metrics

- Primary recall: share of valid reverse profiles whose originating story appears in the uncapped or high-limit emitted candidate pool.
- Downstream recall: share whose originating story survives the LLM/coarse relevance pass as `keep` or `monitor_only`.
- Ranking quality: rank distribution for matching stories: top 3, top 5, top 10, below 10.

## Launch Gate Draft

- Strong recall pass: 80% or more valid reverse profiles get the originating story into the uncapped/high-limit hit list.
- Strong ranking pass: 70% or more valid reverse profiles place the originating story in top 5 after downstream scoring.
- Needs tuning: high recall with weak ranking, especially if failures are mostly `ranking_low`.
- Fail: under 65% primary recall, or source misses cluster around a source/vertical we claim to support.
