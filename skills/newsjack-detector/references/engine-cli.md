# Engine CLI Reference

The Go monitoring engine collects evidence, computes mechanical scores, and orders the queue. The skill owns PR judgment. Use the engine whenever the user asks to monitor, discover, or scan current hooks.

**The CLI is the source of truth for commands, flags, and defaults.** Do not trust a hardcoded flag list — discover the current surface at runtime. This file only covers what `--help` cannot tell you: what the source lanes *mean*, how credentials are supplied, and what profile fields do.

## Discover the current surface

```bash
newsjack help               # all commands
newsjack help detector      # detector subcommands (run, recent, ...)
newsjack detector run --help   # authoritative flags + defaults for a run
newsjack doctor             # health check: runtimes, auth, what's wired
```

The binary: in this repo, prefer the `./bin/newsjack` source shim; `~/.newsjack/bin/newsjack` is the end-user install path that public skills reference. A typical run:

```bash
newsjack detector run --profile profile.json --save
```

Read `--help` for the exact flags before composing a non-trivial run. Notable behaviors worth knowing (confirm specifics via `--help`): detector output is JSON; `--mock` verifies locally without credentials; `--topic` adds an explicit one-off retrieval topic when the user asks for one; `--feed-only`, `--new-only`, and `--max-age-hours` shape recurring runs; `--min-queue-priority` / `--min-major-news` are emission floors (defaults `40` / `0.55` — the canonical run uses the defaults; lowering them changes the emitted pool and breaks run-to-run comparability); `--lane-caps` narrows a skim-only run; `--include-all-scored` and `--no-hygiene-filter` are debug-only. **`--include-all-scored` does not widen the emitted `signals[]`** — it only attaches a `debug.all_scored_signals` block — so never reach for it to "get more candidates"; raise `--limit` or adjust the floors deliberately instead. The canonical, harness-independent invocation lives in `skills/newsjack-detector/SKILL.md` step 1; follow it verbatim rather than re-deriving flags per harness.

## Source lanes (what `--sources` selects)

`--help` lists the flag but not what each lane is for:

- `news_search` — primary news-search layer. Default endpoint `POST https://medialyst.ai/api/v1/news/search`, Serper News request/response shape.
- `x` — direct X API integration. Filters out low-reach single posts by default; may emit a query-volume signal when X recent counts show a topic is moving.
- `x_news` — preferred X discovery shape: returns story clusters, not random posts. Enable by default in profiles once wired.
- `x_trends` — profile config `personalized`, `location`, or `none`. Use location trends only when geography matters.
- `major_feed` — RSS/Atom lane for curated major-news feeds. Profile `feed_urls` are included automatically.
- `reddit`, `hackernews` — optional v0 sources.

## Credentials & tuning (environment, not flags)

The engine reads credentials from the process environment or repo-root `.env`; check status with `newsjack auth status` and `newsjack doctor`. If credentials are missing, run `newsjack detector diagnose` and report which source is unavailable.

- `MEDIALYST_API_KEY`, `MEDIALYST_API_BASE`, `MEDIALYST_NEWS_PATH`, and optional X bearer-token aliases.
- X source tuning knobs (env vars, not run flags): `NEWSJACK_X_MIN_ENGAGEMENT`, `NEWSJACK_X_MIN_AUTHOR_FOLLOWERS`, `NEWSJACK_X_MIN_VIEWS`, `NEWSJACK_X_TREND_MIN_24H`, `NEWSJACK_X_TREND_MIN_6H`, `NEWSJACK_X_TREND_MIN_VELOCITY`. These raise/lower the reach and momentum bars for the `x` lane; leave at defaults unless deliberately tuning noise.

## Recurring local workflow

```bash
newsjack detector run --profile profile.json --feed-only --save --new-only --max-age-hours 24
```

A compromise for local/agent runtimes that can only run hourly. The RSS lane catches major stories first, then tests client relevance. `--new-only` uses the local monitor store to avoid re-alerting the same feed URLs; `--max-age-hours` keeps the first run from dumping a historical backlog. It is not a promise to win the first 15 minutes of a breaking story.

If `--new-only` returns no signals, report "no new signals since the last saved pass" — not source failure.

## Profiles

- `feed_urls` — used by default. The shipped catalog at `references/rss-feeds.json` is the setup/onboarding starting point.
- `search_terms` — when present, the engine retrieves with these instead of raw `topics + competitors`. Keep `topics`, `competitors`, and `standing` as canonical matching/judgment context; use `search_terms` for qualified strings like `Ada customer service`, `Aura identity theft`, `Good Move cash house buyer`.
- `search_terms` must be static, explicit, and provenance-safe: user input, client materials, named competitors/products/regulators, or current coverage. Do not use model-remembered sector trends as durable profile terms.
- If no profile file exists, accept plain-text company/client context and create a temporary JSON profile outside the repo. Do not invent profile facts.
