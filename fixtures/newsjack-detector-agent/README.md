# Newsjack Detector Fixture

This fixture runs the local Newsjack detector against sample company profiles. It is useful for smoke testing candidate retrieval, report rendering, and agent orchestration.

## Quick Start

Run every fixture profile:

```bash
fixtures/newsjack-detector-agent/scripts/run-all-profiles.sh
```

Run one profile and write a complete run folder:

```bash
fixtures/newsjack-detector-agent/scripts/run-one-profile.sh localfalcon "AI search visibility" profile.localfalcon.json
```

Run a deterministic mock detector call without live API credentials:

```bash
fixtures/newsjack-detector-agent/scripts/run-mock-detector.sh
```

Open Claude Code inside the fixture environment:

```bash
fixtures/newsjack-detector-agent/scripts/open-claude.sh
```

Ask Claude Code to run the fixture:

```bash
fixtures/newsjack-detector-agent/scripts/open-claude-bypass.sh -p "Read and follow PROMPT.md. Run scripts/run-all-profiles.sh. Inspect the generated index.md and each run.md, then summarize what surfaced, what looks wrong, and the run folder path."
```

## Scripts

| Script | Use |
|---|---|
| `scripts/run-all-profiles.sh` | Runs all configured fixture profiles and writes `runs/<timestamp>/index.md`. |
| `scripts/run-one-profile.sh` | Runs one profile, writes `candidates.json`, `summary.json`, `run.md`, logs, and includes the full scored pool by default. |
| `scripts/run-detector-json.sh` | Runs one live detector call and prints raw JSON to stdout. |
| `scripts/run-mock-detector.sh` | Runs one mock detector call and prints raw JSON to stdout. |
| `scripts/with-fixture-env.sh` | Loads root `.env`, then fixture `.env`, changes into this fixture directory, and runs a command. |
| `scripts/open-claude.sh` | Opens `claude` with fixture env loaded. |
| `scripts/open-claude-bypass.sh` | Opens `claude` with fixture env loaded and permission bypass flags. |
| `scripts/open-codex.sh` | Opens `codex` with fixture env loaded. |

## Output

`run-all-profiles.sh` writes:

```text
runs/<timestamp>/
  index.md
  <profile>/
    candidates.json
    detector.stderr.log
    summary.json
    run.md
```

`run-one-profile.sh` writes:

```text
runs/<timestamp>-<slug>-profile/
  candidates.json
  scored_candidates.json
  detector.stderr.log
  commands.log
  summary.json
  run.md
```

The fixture reports are detector previews unless an agent also runs the full semantic pipeline from `skills/newsjack-detector/SKILL.md`: coarse relevance, story-origin check, freshness gate, final report, and rerender.

Fixture-specific script usage lives here so the public skill can stay runtime-agnostic.

## Configuration

Create `fixtures/newsjack-detector-agent/.env` for fixture-local overrides. The helper loads the repo root `.env` first, then this fixture `.env`.

Common environment overrides:

```bash
NEWSJACK_SOURCES=news_search,x
NEWSJACK_LOOKBACK_DAYS=1
NEWSJACK_DEPTH=quick
NEWSJACK_LIMIT=80
NEWSJACK_MAX_AGE_HOURS=24
NEWSJACK_SAVE=1
NEWSJACK_NEW_ONLY=0
NEWSJACK_RUN_DIR=fixtures/newsjack-detector-agent/runs/manual-test
NEWSJACK_BIN=/absolute/path/to/newsjack
```

`run-all-profiles.sh` defaults to `NEWSJACK_NEW_ONLY=0` so repeated smoke runs still show current candidates. Set `NEWSJACK_NEW_ONLY=1` for hourly-style monitoring that suppresses already-seen URLs from the fixture store.

For `run-one-profile.sh`, these can replace positional args:

```bash
NEWSJACK_PROFILE_SLUG=simular
NEWSJACK_PROFILE_QUERY="computer-use agents"
NEWSJACK_PROFILE_FILE=profile.simular.json
```

Extra CLI flags are forwarded to `newsjack detector run`.
