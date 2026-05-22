You are testing the local newsjack-detector skill in this repo.

First read:

- `../../skills/newsjack-detector/SKILL.md`
- `../../skills/newsjack-detector/rubric.md`

Use one of these monitor profiles:

- `profile.localfalcon.json`
- `profile.simular.json`
- `profile.slite.json`
- `profile.property-saviour.json`
- `profile.clearnym.json`

Suggested profile/query pairs:

| Profile | Query |
|---------|-------|
| `profile.localfalcon.json` | `AI search visibility` |
| `profile.simular.json` | `computer-use agents` |
| `profile.slite.json` | `AI knowledge base` |
| `profile.property-saviour.json` | `UK property chain collapse` |
| `profile.clearnym.json` | `data broker removal` |

Each profile also includes `search_terms`, `feed_urls` selected from the shipped RSS catalog, `x_news.enabled: true`, and an `x_trends` preference. Profile RSS feeds and X News are included by default.

Run the monitoring engine from this fixture directory. Start with a mock pass:

```bash
python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run "AI search visibility" --profile profile.localfalcon.json --mock --emit json
```

Or use the helper script:

```bash
./scripts/mock-run.sh "AI knowledge base" profile.slite.json
```

To test only the profile's RSS feeds:

```bash
python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run --profile profile.localfalcon.json --feed-only --emit brief
```

To simulate the hourly product behavior with duplicate suppression:

```bash
python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run --profile profile.localfalcon.json --feed-only --save --new-only --max-age-hours 24 --emit brief
```

Then, if `MEDIALYST_API_KEY` is available and `xurl whoami` succeeds, run a live quick pass:

```bash
python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run "AI search visibility" --profile profile.localfalcon.json --sources news_search,x --lookback-days 1 --depth quick --save --emit json
```

Or use the helper script:

```bash
./scripts/live-run.sh "computer-use agents" profile.simular.json
```

For a fully observable run folder, start with:

```bash
./scripts/observe-run.sh simular "computer-use agents" profile.simular.json
```

The observed run defaults to `--limit 80 --min-queue-priority 40 --min-major-news 0.55 --include-all-scored`. It writes support artifacts as JSON/log files and one human-readable Markdown artifact: `run.md`. During a detector-only run, `run.md` shows the detector state and marks later LLM stages as pending. After the cheap filter, filter-apply step, and expensive rubric pass are done, rerender `run.md` so it contains the full pipeline status and embeds `final_report.md`.

To run feed-only broad monitoring without query/news-search credentials:

```bash
python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run --profile profile.localfalcon.json --feed-only --save --new-only --max-age-hours 24 --emit brief
```

To test the harness-neutral two-pass flow, run the detector to `candidates.json`, have the harness write cheap-filter decisions, then apply them:

```bash
./scripts/observe-run.sh localfalcon "AI search visibility" profile.localfalcon.json
RUN_DIR="$(ls -td runs/*-localfalcon-observe | head -1)"
python3 ../../skills/newsjack-detector/scripts/newsjack_filter_apply.py --candidates "$RUN_DIR/candidates.json" --decisions "$RUN_DIR/filter_decisions.json" --include keep --include monitor_only --output "$RUN_DIR/targeted_candidates.json"
python3 scripts/summarize-run.py "$RUN_DIR/candidates.json" --output "$RUN_DIR/summary.json" --markdown "$RUN_DIR/run.md"
```

The harness decision path and cheap filter prompt live in `../../skills/newsjack-detector/SKILL.md`. Use a cheap model or cheap workers/subagents when the harness exposes them; otherwise disclose current-model fallback. The cheap filter should evaluate each signal independently and write `$RUN_DIR/filter_decisions.json`. The expensive rubric pass should write `$RUN_DIR/final_report.md`, then rerun `summarize-run.py` so `$RUN_DIR/run.md` is the full observable Markdown result. For two-pass runs, prefer floor-based selection (`--min-queue-priority`, `--min-major-news`) over hard lane caps so the cheap filter sees the broader candidate pool.

Apply the skill rubric to the returned `signals`, but do not dump the full JSON object in the final response. Return a human-readable skim report:

```text
Verdict
- One sentence on whether there is anything worth acting on.
- Source coverage: news_search / x_news / x_trends / x / RSS feeds used or missing.
- If `--new-only` returns no signals, say there are no new signals since the previous saved pass.
- For X results, distinguish `query_trend` evidence from individual post evidence. Do not treat a lone low-reach X post as a real signal.

Best Bets
| verdict | signal | why it matters | proof needed | next |
Sources: include 1-3 evidence links for each Best Bet as `source: title - URL`. Include news, RSS, and X links when present.

Monitor
| signal | why watch | trigger to act |
Sources: include 1-3 evidence links for each Monitor item as `source: title - URL`.

Rejects
- Count by reason, then only list notable false positives or tuning issues.
- Include links for notable false positives when they are useful for debugging.

Tuning Notes
- Query/profile/feed changes that would reduce noise.
- Mention `search_terms` when a profile needs qualified retrieval terms for ambiguous names.
- Mention X tuning only when relevant: low-reach posts are filtered before queueing; useful X knobs are `NEWSJACK_X_MIN_ENGAGEMENT`, `NEWSJACK_X_MIN_AUTHOR_FOLLOWERS`, `NEWSJACK_X_MIN_VIEWS`, and the `NEWSJACK_X_TREND_*` thresholds.
```

Keep each row short enough to skim. Do not cite only signal IDs; include source URLs anywhere the reader needs to validate the judgment. Mention signal IDs only when they help debugging.
