# Prompt: Harvest Daily Seed Stories

Use this prompt with an agent that has web access.

```text
Find today's reverse-newsjack eval seed stories.

Current date: YYYY-MM-DD
Location context: USER_LOCATION_OR_SOURCE_CONTEXT

Goal:
Create a cross-industry list of current seed stories that can be used to test whether Newsjack rediscovers known trending stories from plausible company profiles.

Seed surfaces:
- Google News Business
- Google News Technology
- Google News Health
- Google News Science
- Google Trends
- Major primary/source pages where available
- AP/Reuters/major trade coverage where available

Rules:
- Pick 12-20 stories.
- Span multiple industries: SaaS, AI, fintech, CPG/retail, climate/energy, healthtech, cybersecurity, media, consumer hardware, proptech if available.
- Prefer stories with multiple-source clusters or a primary-source anchor.
- Exclude tragedy hooks from the positive recall set.
- Avoid broad live blogs, generic stock-market updates, and personality/tabloid stories unless they are explicitly for rejection testing.
- Do not use Newsjack detector output as the source of truth for this seed list.

For each story, return:
- target_id: rev-YYYYMMDD-###
- seed_surface
- source_cluster_title
- canonical_story_summary
- source_urls, with publication/source names
- industry_batch
- expected_entities
- why_it_is_a_good_reverse_eval_seed
- brand_safety_status: positive_seed / rejection_only / exclude
- bias_notes

Save the result as Markdown with a table plus short notes.
```
