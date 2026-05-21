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
python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run --profile profile.localfalcon.json --feed-only --save --new-only --max-age-hours 48 --emit brief
```

Then, if `MEDIALYST_API_KEY` is available and `xurl whoami` succeeds, run a live quick pass:

```bash
python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run "AI search visibility" --profile profile.localfalcon.json --sources news_search,x --lookback-days 7 --depth quick --save --emit json
```

Or use the helper script:

```bash
./scripts/live-run.sh "computer-use agents" profile.simular.json
```

To run feed-only broad monitoring without query/news-search credentials:

```bash
python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run --profile profile.localfalcon.json --feed-only --save --new-only --max-age-hours 48 --emit brief
```

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
