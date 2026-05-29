# Prompt: Select Reverse-Profile Companies

Use this prompt after seed stories are selected.

```text
Given this reverse-newsjack seed list, identify plausible companies for each target story.

Goal:
For each seed story, choose companies that would have credible standing to comment on or newsjack the story. Vary company scale across the batch.

Rules:
- Provide startup/growth, mid-market, and large/enterprise candidates where possible.
- Recommend one first-run company per target.
- Do not choose the company that is the subject of the story unless testing competitor/self-coverage behavior.
- Use size bands as approximate eval-design labels only, not exact headcount claims.
- Prefer narrower companies with direct standing over famous generic companies.
- If no credible company exists, mark the seed `invalid_seed`.

For each target, return:
- target_id
- originating_story
- startup_or_growth_candidate
- mid_market_candidate
- large_or_enterprise_candidate
- recommended_first_profile
- size_shape
- standing_rationale
- search_terms_to_include
- search_terms_to_avoid

Save as Markdown.
```
