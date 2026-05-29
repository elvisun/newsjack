# Reverse Newsjack Eval Runbook

Use this runbook to recreate the reverse-eval setup on any day.

The eval answers two separate questions:

1. Recall: if we start from a known current story and give Newsjack a plausible company profile with real standing, does the detector surface that originating story anywhere in the uncapped or high-limit hit list?
2. Ranking: if the story is found, does the downstream pipeline prioritize it high enough for a user-facing brief?

## Daily Workflow

1. **Create a dated scratch folder.**

   ```bash
   mkdir -p .tmp/reverse-newsjack/YYYY-MM-DD
   ```

2. **Collect independent seed stories.**

   Use `prompts/01-harvest-seeds.md`.

   Save output as:

   ```text
   .tmp/reverse-newsjack/YYYY-MM-DD/seeds.md
   ```

3. **Select reverse-profile companies.**

   Use `prompts/02-select-companies.md`.

   Save output as:

   ```text
   .tmp/reverse-newsjack/YYYY-MM-DD/company-candidates.md
   ```

4. **Create temporary profile JSON files.**

   Profiles should live under the run folder, not in production fixtures:

   ```text
   .tmp/reverse-newsjack/YYYY-MM-DD/profiles/
   ```

   Keep profiles factual and minimal. Do not invent company facts.

5. **Run the detector batch.**

   Use `prompts/03-run-detector-batch.md`.

   Save output under:

   ```text
   .tmp/reverse-newsjack/YYYY-MM-DD/runs/
   ```

6. **Score story recall and ranking separately.**

   Use `prompts/04-score-results.md`.

   Save:

   ```text
   .tmp/reverse-newsjack/YYYY-MM-DD/results/overlap_matrix.csv
   .tmp/reverse-newsjack/YYYY-MM-DD/results/summary.md
   ```

7. **Promote durable target sets.**

   If the daily target set is worth keeping, copy the cleaned target/company files into this folder with a date, for example:

   ```text
   eval/reverse-newsjack/targets-YYYY-MM-DD-source.md
   eval/reverse-newsjack/company-candidates-YYYY-MM-DD.md
   ```

8. **Promote the whole run into a committed dated folder.**

   `.tmp/` is ignored by Git. Before committing, copy the run into:

   ```text
   eval/reverse-newsjack/runs/YYYY-MM-DD-source/
   ```

   Use this durable shape:

   ```text
   eval/reverse-newsjack/runs/YYYY-MM-DD-source/
     README.md
     targets.md
     company-candidates.md
     profiles/
     detector-runs/
     full-pipeline-smoke/
     results/
     progress.md
   ```

   Keep the scratch layout under `.tmp/` while iterating, but only commit the durable folder.

## Required Artifacts

Each scratch eval run should contain:

```text
.tmp/reverse-newsjack/YYYY-MM-DD/
  seeds.md
  company-candidates.md
  profiles/
  runs/
  results/
    overlap_matrix.csv
    summary.md
  progress.md
```

Each committed eval run should contain the same material under `eval/reverse-newsjack/runs/YYYY-MM-DD-source/`, with scratch `runs/` renamed to `detector-runs/` to avoid nested `runs/runs` ambiguity.

## Scoring Labels

- `candidate_pool_pass`: originating story is emitted in the uncapped or high-limit candidate pool.
- `final_hit`: originating story survives downstream LLM/coarse relevance as a hit.
- `ranking_low`: story is emitted, but below the product-visible rank threshold being tested.
- `ranking_miss`: story is present in `debug.all_scored_signals` but not emitted.
- `source_miss`: story is not scored at all.
- `invalid_seed`: story should not be in the positive recall set.

## Bias Tags

- `independent_seed`: seed source was not directly ingested during the detector run.
- `same_source_assisted`: detector ingested the same source family as the seed source.
- `profile_search_assisted`: profile `search_terms` included direct exact-story terms that make retrieval easier than a real client profile would be.

For launch-gate numbers, report all three views:

- all runs,
- excluding `invalid_seed`,
- excluding `same_source_assisted` and `profile_search_assisted`.

## Profile Design Rules

- The company must have credible standing on the story.
- Do not use the company that is the subject of the story unless the test is explicitly about competitor/self-coverage.
- Vary size and market shape across the batch.
- Keep `search_terms` realistic for a client profile. Do not include the exact article headline.
- Include broad category terms, competitor names, and specific product/category phrases.

## Detector Run Defaults

Use these defaults for reverse eval unless testing a specific source behavior:

```bash
~/.newsjack/bin/newsjack detector run "QUERY" \
  --profile PROFILE.json \
  --sources news_search,x \
  --lookback-days 1 \
  --depth quick \
  --limit 0 \
  --include-all-scored \
  --no-x-trends \
  --emit json
```

Use `--limit 0` for the primary recall gate so the eval does not confuse candidate-pool recall with ranking quality. If you also want a product-visible ranking check, rerun or post-process with the product limit, for example top 20, and record rank buckets separately.

Do not use `--save` or `--new-only` for eval runs unless the eval is specifically about monitor-store behavior.

## Launch Gate Draft

- Strong recall pass: 80% or more valid reverse profiles get the originating story into the uncapped/high-limit hit list.
- Strong ranking pass: 70% or more valid reverse profiles place the originating story in top 5 after downstream scoring.
- Needs tuning: high recall with weak ranking, especially if failures are mostly `ranking_low`.
- Fail: under 65% primary recall, or source misses cluster around a source/vertical we claim to support.
