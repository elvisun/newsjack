# Prompt: Run Reverse-Eval Detector Batch

Use this prompt with a coding agent in the repo.

```text
Run a reverse-newsjack detector batch for the selected targets.

Do not modify production profile fixtures. Write all run artifacts under:
.tmp/reverse-newsjack/YYYY-MM-DD/

Make the run resumable:
- Write progress.md immediately.
- Append a timestamped progress line after each profile.
- If an output file already exists and is valid JSON, treat it as complete.
- Continue on failures and record them.

Inputs:
- seeds.md
- company-candidates.md

Tasks:
1. Create one temporary profile JSON per recommended first-run company under profiles/.
2. Keep each profile factual and minimal:
   - company
   - topics
   - standing
   - competitors
   - spokespeople
   - search_terms
   - feed_urls only if intentionally testing source-assisted recall
3. Do not include exact article headlines in search_terms.
4. Run the detector for each profile with:

   ~/.newsjack/bin/newsjack detector run "QUERY" \
     --profile PROFILE.json \
     --sources news_search,x \
     --lookback-days 1 \
     --depth quick \
     --limit 0 \
     --include-all-scored \
     --no-x-trends

5. Do not use --save or --new-only.
6. Use `--limit 0` for primary recall. Rank buckets are scored separately from recall.
7. Save per target:
   - command.txt
   - profile.json
   - candidates.json
   - stderr.log
   - summary.json from `newsjack run-summary`

Output folder shape:
runs/TARGET_ID/
  command.txt
  profile.json
  candidates.json
  stderr.log
  summary.json

Final response:
- run folder path
- count of profiles run
- failures
- next scoring command/prompt to use
```
