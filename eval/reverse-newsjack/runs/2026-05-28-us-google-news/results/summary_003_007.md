# Reverse Eval Summary: rev-20260528-003 to rev-20260528-007

Run date: 2026-05-28.

Detector mode: high-recall, `--limit 0`, `--include-all-scored`, `news_search,x`, local `go run ./cmd/newsjack`.

## Results

| Target | Company | Primary Recall | Rank Bucket | Best Match |
|---|---|---:|---|---|
| rev-20260528-003 | Orgvue | pass | top_3 | `Wix CEO confirms layoffs, blames currency fluctuations, AI (WIX:NASDAQ)` |
| rev-20260528-004 | TRM Labs | pass | top_3 | `Google Employee Accused Of Making $1 Million From Insider Trading On Polymarket` |
| rev-20260528-005 | Profound | pass | below_10 | `ChatGPT Ads Conversion-Optimized Campaigns Coming June 5th` |
| rev-20260528-006 | CloudZero | pass | top_3 | `Snowflake signs US$6bn AWS deal for enterprise AI workloads` |
| rev-20260528-007 | Chainguard | pass | top_3 | `IBM and Red Hat invest $5 billion in the future of open source` |

## Metrics

- Valid targets: 5
- Primary recall rate: 5/5
- Top-3 ranking rate: 4/5
- Top-5 ranking rate: 4/5
- Top-10 ranking rate: 4/5
- Below-10 but emitted: 1/5
- Ranking misses: 0
- Source misses: 0

## Notes

`rev-20260528-005` is the useful failure signal. The exact Business Insider seed (`OpenAI has vast ad ambitions...`) was present in `debug.all_scored_signals` but not emitted because `profile_match=0`, despite `story_size=0.526`. A same-story or tightly adjacent ChatGPT ads variant was emitted at rank 17 through the large-story recall guard.

That means the new methodology should count it as a primary recall pass, but ranking quality is weak. This is exactly why top-N should be a secondary metric instead of the recall gate.
