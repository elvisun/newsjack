# Newsjack Detector Fixture

This fixture runs the local Newsjack detector against sample company profiles. It is useful for smoke testing candidate retrieval, machine-readable summaries, and agent orchestration.

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
fixtures/newsjack-detector-agent/scripts/open-claude-bypass.sh -p "Read and follow PROMPT.md. Run scripts/run-all-profiles.sh. Inspect the generated index.md plus each summary.json/candidates.json, then summarize what surfaced, what looks wrong, and the run folder path."
```

## Scripts

| Script | Use |
|---|---|
| `scripts/run-all-profiles.sh` | Runs all configured fixture profiles and writes `runs/<timestamp>/index.md`. |
| `scripts/run-one-profile.sh` | Runs one profile, writes `candidates.json`, `summary.json`, logs, and includes the full scored pool by default. |
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
```

`run-one-profile.sh` writes:

```text
runs/<timestamp>-<slug>-profile/
  candidates.json
  scored_candidates.json
  detector.stderr.log
  commands.log
  summary.json
```

When a companion brief exists, `run-all-profiles.sh` links it from `index.md`
and `run-one-profile.sh` records it as `brief_path` in `commands.log`.

The fixture summary is a detector artifact, not a human report. A `run.md` report is created only when an agent runs the full semantic pipeline from `skills/newsjack-detector/SKILL.md`: coarse relevance, story-origin check, freshness gate, triage, angle generation, and skill-owned report rendering.

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
NEWSJACK_USE_INSTALLED=1
```

By default, the fixture uses the repo source shim when Go is available. In an
installed-user container without Go, it falls back to `~/.newsjack/bin/newsjack`
when that binary exists. Set `NEWSJACK_USE_INSTALLED=1` to require the installed
binary explicitly.

`run-all-profiles.sh` defaults to `NEWSJACK_NEW_ONLY=0` so repeated smoke runs still show current candidates. Set `NEWSJACK_NEW_ONLY=1` for hourly-style monitoring that suppresses already-seen URLs from the fixture store.

For `run-one-profile.sh`, these can replace positional args:

```bash
NEWSJACK_PROFILE_SLUG=simular
NEWSJACK_PROFILE_QUERY="computer-use agents"
NEWSJACK_PROFILE_FILE=profile.simular.json
```

Extra CLI flags are forwarded to `newsjack detector run`.

## Client briefs

A profile's optional companion `brief.<slug>.md` is the prose **source of truth** for what that client will and won't pitch and how the scan should be presented (see the **Client Brief** section in `skills/newsjack-detector/SKILL.md`). It is read at the triage and report-rendering stages — never at collection — so nothing is dropped before judgment. A "never pitch" rule keeps an off-policy item out of `pitch_ready` (a fresh big story stays surfaced as `off_policy` `big_story`, never hidden); "How to surface" can collapse a section to a disclosed count. In a real install the brief lives at `~/.newsjack/monitors/<slug>/brief.md` and is scaffolded by `newsjack monitor init`; in this fixture, where profiles are flat files, it is `brief.<slug>.md` beside the profile and is surfaced by the fixture scripts. `brief.clearnym.md` is a worked example. When the user gives feedback on what to pitch or surface, the agent updates the brief so the policy sticks across runs.
