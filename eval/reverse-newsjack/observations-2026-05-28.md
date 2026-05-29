# Observations: 2026-05-28 Reverse Eval Smoke Tests

Artifacts live under `.tmp/reverse-newsjack/2026-05-28-test/`.

## rev-20260528-001: Specright / Temu Fine

Verdict: `ranking_miss`.

The detector retrieved and scored the Temu story, but did not emit it. The run produced:

- 4 emitted signals.
- 114 scored signals.
- 35 Temu-titled items in `debug.all_scored_signals`.

Representative matching titles:

- `EU Commission fines Temu €200m under DSA over illegal product risks`
- `Chinese online retailer Temu hit with $232 million fine over unsafe toys and electronics`
- `EU Tests Limits of Platform 'Risk Assessments' with €220 Million Temu Fine`

The matching Temu items were demoted to `profile_relevance_weak` or `x_news_unmatched` because `profile_match` stayed below the `0.05` lane threshold. Profile-match values were generally `0.014-0.034`, while emitted candidates scored `0.060-0.138`.

### Sensitivity Check

A temporary profile variant moved marketplace/DSA concepts into scored profile fields (`topics` and `standing`) instead of leaving them only in `search_terms`.

Result: Temu variants emitted immediately.

- 17 emitted signals.
- 117 scored signals.
- Multiple Temu variants in the emitted set.
- Representative queue/profile scores: `queue_priority` around `61.3-61.9`, `profile_match` around `0.050-0.089`.

This suggests the miss was not retrieval coverage. It was scorer-field / profile vocabulary alignment.

### Large-Story Recall Fix Check

After adding a `large_story_remote_relevance` guard, reran the original profile without the sensitivity vocabulary changes.

Result: Temu variants emitted without relying on search-term promotion.

- 17 emitted signals.
- 82 scored signals.
- 4 Temu variants in the emitted set.
- Temu variants appeared at positions 7, 8, 9, and 17.
- Representative queue/profile/story-size scores: `queue_priority` `46`, `profile_match` `0.013-0.029`, `story_size` `0.501-0.614` for high-band variants.

This changes the failure mode from `ranking_miss` at the mechanical gate to `candidate_pool_pass`; a downstream LLM/rubric still needs to rank whether the story is a top-3 fit.

## rev-20260528-002: Form Health / CVS Zepbound Coverage

Verdict: `pass_top_3`.

The detector emitted the originating story at the top of the candidate set. The run produced:

- 10 emitted signals.
- 127 scored signals.
- Positions 1-6 were variants of the CVS/Zepbound/Foundayo story.

Representative matching titles:

- `CVS Expands Access to Eli Lilly's Obesity Medicines`
- `CVS Caremark Delivers Affordability and Access to GLP-1 Weight Management Medications with Expanded Coverage Options`
- `CVS to restore coverage of Zepbound, add Eli Lilly's obesity pill to drug plans`

This profile worked because the real client vocabulary and the story vocabulary naturally overlapped: GLP-1 access, obesity medicine, PBM/formulary coverage, and anti-obesity medications.

## rev-20260528-003 to rev-20260528-007: Parallel Recall Batch

Method update: primary recall is now measured against an uncapped/high-limit emitted hit list (`--limit 0`). Top-3/top-5 are ranking-quality metrics, not the recall gate.

Summary:

| Target | Company | Primary Recall | Rank Bucket | Note |
|---|---|---:|---|---|
| rev-20260528-003 | Orgvue | pass | top_3 | Wix AI/layoff variants at ranks 1-3. |
| rev-20260528-004 | TRM Labs | pass | top_3 | Google/Polymarket insider-trading variant at rank 2. |
| rev-20260528-005 | Profound | pass | below_10 | ChatGPT ads variant emitted at rank 17; exact BI seed scored but not emitted. |
| rev-20260528-006 | CloudZero | pass | top_3 | Snowflake/AWS $6B partnership variant at rank 2. |
| rev-20260528-007 | Chainguard | pass | top_3 | IBM/Red Hat $5B open-source story at rank 1. |

Artifacts:

- `.tmp/reverse-newsjack/2026-05-28-test/results/overlap_matrix_003_007.csv`
- `.tmp/reverse-newsjack/2026-05-28-test/results/summary_003_007.md`

Batch result: 5/5 primary recall, 4/5 top-3 ranking.

## Scoring Finding

`search_terms` currently drive retrieval, but not profile relevance scoring.

Relevant code path:

- `monitorProfile.queryTerms()` uses `SearchTerms` for retrieval.
- `monitorProfile.matchText()` excludes `SearchTerms`.
- `profileMatchScore()` scores against `matchText()`.
- `signalLane()` demotes profile-relevance signals below the configured profile-match floor.

Practical effect: a story can be retrieved by an excellent `search_terms` query, then demoted because the scorer ignores that same vocabulary.

This matters for reverse eval because realistic profiles often use qualified retrieval strings for ambiguous/current concepts. If those terms are not also represented in `topics`, `standing`, or `competitors`, recall can fail after retrieval.

## Candidate Fix Direction

Do not blindly add all `search_terms` to scoring. Some search terms are retrieval hacks or competitor-specific strings and could inflate weak matches.

Safer options:

1. Add a separate scored field such as `relevance_terms` / `profile_match_terms`.
2. Include selected `search_terms` in scoring with lower weight.
3. Preserve `search_terms` for retrieval, but require profile setup to mirror durable concepts into `topics` or `standing`.
4. Add eval instrumentation that reports `retrieved_by_query` and `profile_match_terms_hit` separately.

For launch-gate reporting, distinguish:

- `source_miss`: story not scored at all.
- `retrieval_success_ranking_miss`: story scored but demoted.
- `profile_vocab_miss`: story scored because of `search_terms`, but `profile_match` failed because durable profile fields lacked matching vocabulary.
