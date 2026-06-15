# Go CLI Port Plan

This document captures the move from the current curl installer plus shell shim to a real Go `newsjack` CLI, and the verification strategy for preserving detector behavior while moving implementation behind the CLI.

Status: implemented in this branch as a Go CLI under `apps/cli`, with `install.sh` reduced to bootstrap, runtime skill installation delegated to `newsjack install`, and public commands routed through `newsjack ...`.

## Goal

Make `newsjack` a stable product interface:

```bash
newsjack install
newsjack update
newsjack doctor
newsjack skills list
newsjack runtimes detect
newsjack login
newsjack news search ...
newsjack media-lists create ...
newsjack detector run ...
```

The user-facing interface should not expose repo-relative implementation paths.

## Current Problem

The first curl installer works, but it blurs ownership:

- `install.sh` owns runtime detection and file copying
- `bin/newsjack` is a shell dispatcher
- runtime-owned skill dirs receive full skill folders, including scripts
- skill scripts are part product interface, part implementation detail

That works for v1 bootstrap, but it is harder to explain:

```text
Go CLI installs implementation scripts into Claude/Codex/OpenClaw/Hermes skill dirs.
```

The better model is:

```text
~/.newsjack/
  bin/newsjack                  # Go binary
  newsjack/                     # managed Newsjack bundle
    skills/
    scripts/
    fixtures/
  credentials.json
  state/

~/.claude/skills/
  newsjack-detector/
    SKILL.md                    # thin wrapper or rewritten full instruction file

~/.agents/skills/
  newsjack-detector/
    SKILL.md
```

Runtime skill dirs should contain instructions. Newsjack-owned code should live under `~/.newsjack`.

## Architecture

### Go Owns

- CLI command parsing
- install/update
- runtime detection
- skill registration into runtime dirs
- Medialyst REST API client commands
- config and credential paths
- `doctor`
- downloads and checksums once release artifacts exist
- stable command UX

### Skill Engines Own

- news detection
- RSS/news/HN/Reddit/X retrieval
- scoring and filtering
- monitor store behavior
- run summarization

The end state may be all-Go, but migration should not rewrite every behavior at once without fixtures.

## Repository Shape

Proposed layout:

```text
apps/cli/
  go.mod
  cmd/newsjack/main.go
  internal/
    auth/
    cli/
    doctor/
    execx/
    install/
    runtimes/
      claude.go
      codex.go
      hermes.go
      openclaw.go
    skills/
    update/
```

Optional later layout if detector is ported:

```text
apps/cli/internal/detector/
  feeds/
  profile/
  scoring/
  sources/
  store/
```

## Command Contract

The Go CLI should support the current shim commands first:

```bash
newsjack help
newsjack path
newsjack skills
newsjack login [--key KEY]
newsjack auth status
newsjack auth headers
newsjack auth logout
newsjack detector run ...
newsjack filter-apply ...
newsjack run-summary ...        # JSON artifact summary only; report rendering lives in skills
newsjack news search ...
newsjack media-lists create ...
newsjack update
```

Then add real product commands:

```bash
newsjack version
newsjack doctor
newsjack skills list
newsjack skills install
newsjack runtimes detect
```

Compatibility aliases can keep `newsjack skills` equivalent to `newsjack skills list`.

## Installation Model

Keep `install.sh` small and POSIX.

For v1:

1. Fetch GitHub `main` tarball.
2. Install managed source bundle at `~/.newsjack/newsjack`.
3. Install or build `~/.newsjack/bin/newsjack`.
4. Run `newsjack install --source ~/.newsjack/newsjack`.

Later, once release artifacts exist:

1. Detect OS and arch.
2. Download the matching Go binary.
3. Verify checksum.
4. Fetch or sync the Newsjack bundle.
5. Run `newsjack install`.

## Skill Registration Model

Do not copy Python scripts into runtime-owned skill dirs.

Instead:

1. Keep canonical skill source in `~/.newsjack/newsjack/skills`.
2. Generate runtime skill files into:
   - `~/.agents/skills` for Codex-style agent skills
   - `~/.claude/skills` for Claude Code
   - `~/.openclaw/skills` for OpenClaw
   - `~/.hermes/skills` for Hermes
3. Installed skill files should point agents to `~/.newsjack/bin/newsjack`.
4. Generated files get a Newsjack marker so updates can safely replace them.
5. Existing non-Newsjack user skill files are never overwritten unless forced.

Two wrapper options:

- Thin wrapper: short `SKILL.md` that says where to read canonical instructions and which CLI command to call.
- Rewritten full skill: full canonical instructions copied into runtime dir with script paths rewritten to `newsjack ...`.

Default recommendation: use rewritten full skill files. Agents can read everything locally, while executable code still stays under `~/.newsjack`.

## Porting Strategy

### Phase 1: Go CLI Parity

Install a Go `newsjack` binary and keep `bin/newsjack` only as a source-checkout developer shim.

Keep behavior parity first. The goal is command parity before changing public contracts.

Deliverables:

- `apps/cli` Go module
- compiled `newsjack` binary
- command parity with the previous shell-dispatched CLI
- `version`
- basic `doctor`
- tests for command parsing and path resolution

### Phase 2: Move Install Logic Into Go

Move these out of `install.sh`:

- runtime detection
- runtime selection parsing
- skill registration
- marker-file overwrite protection
- REST-backed media-list and news commands
- doctrine file handling

`install.sh` becomes bootstrap only.

### Phase 3: Port Small Operational Scripts

Port low-risk operational paths into Go:

- auth and credential management
- REST API calls for Medialyst news, enrichment, and media lists
- coarse-relevance decision application
- story-origin freshness gate
- run summarization

Keep output shape compatible.

### Phase 4: Golden-Output Detector Port

Port the detector only behind golden fixtures.

The detector is behavior-heavy and includes:

- RSS parsing
- date normalization
- scoring
- dedupe
- local SQLite monitor store
- source-specific quirks for news, HN, Reddit, and X
- JSON output contracts used by skills and fixtures

Do not port it without fixture coverage.

## Golden-Output Verification

The detector port should use golden fixtures generated from the current Python implementation.

### Fixture Shape

Each fixture should include:

```text
fixtures/golden/detector/<case>/
  profile.json
  inputs/
    rss.xml
    news_search.json
    hackernews.json
    reddit.json
    x.json
  command.txt
  expected.json
  expected.brief.txt
```

Use local/static inputs wherever possible so tests do not depend on live APIs.

### Golden Generation

For each case, generate expected JSON and brief output from the current stable implementation, then compare the Go CLI output after normalizing volatile fields.

Normalize fields that are expected to vary:

- run timestamps
- absolute temp paths
- unordered maps
- non-deterministic evidence order, if unavoidable

Prefer making the detector deterministic over over-normalizing output.

### Golden Comparison

The Go detector test should:

1. Run the Go detector against the same fixture.
2. Normalize output.
3. Compare against `expected.json`.
4. Fail with a useful structural diff.

Compatibility gates:

- same top-level JSON shape
- same candidate IDs where IDs are deterministic
- same evidence URLs
- same verdict-relevant scores within tolerance
- same queue inclusion/exclusion decisions
- same warning/error fields

### Fixture Cases

Start with:

- profile-only query
- RSS feed-only run
- stale RSS item ignored by `--max-age-hours`
- `--new-only` suppresses previously stored URLs
- profile relevance scoring
- hygiene filter removes product/docs/SEO junk
- HN source fixture
- Reddit source fixture
- X source fixture with low-engagement filtering
- major-news weak-overlap candidate admitted by `--min-major-news`
- no credentials / missing source fallback

## Verification Layers

### Layer 1: Go Unit Tests

Cover pure logic:

- runtime selection parsing:
  - `auto`
  - `all`
  - `none`
  - `codex,claude`
  - whitespace-separated values
  - `cladue` alias to `claude`
- OS/arch mapping
- Newsjack home path resolution
- skill source discovery
- generated skill path rewriting
- marker-file overwrite rules
- runtime config payload generation
- Hermes YAML insertion and idempotency
- credential file permissions and JSON shape

### Layer 2: Fixture Filesystem Tests

Use temporary dirs:

- fake `$HOME`
- fake Newsjack source bundle
- fake runtime skill dirs
- fake runtime binaries in `$PATH`
- fake existing user skills
- fake Hermes config

Assert:

- generated skills land in the expected runtime dirs
- existing non-Newsjack files are not overwritten
- Newsjack-managed files update cleanly
- forced install overwrites only when requested
- generated wrappers call `~/.newsjack/bin/newsjack`
- install is idempotent
- uninstall leaves user-owned files alone

### Layer 3: Fake Runtime Command Tests

Put fake commands earlier in `$PATH`:

```text
codex
claude
openclaw
hermes
```

Each fake command appends argv to a log file.

Assert:

- runtime skill registration uses the expected command paths
- Hermes config is edited without shelling to an interactive prompt
- commands do not consume installer stdin
- failures are warnings unless the user requested strict mode

### Layer 4: Command Integration Tests

Run the compiled CLI against a temp install:

```bash
newsjack path
newsjack version
newsjack skills list
newsjack runtimes detect
newsjack doctor
newsjack auth status
newsjack auth headers
```

Expected behavior:

- successful commands return `0`
- missing optional dependencies are warnings in `doctor`
- missing credentials are clear, non-panicking statuses
- output is stable enough for tests and docs

### Layer 5: Installer Smoke Tests

Run:

```bash
HOME="$(mktemp -d)" \
NEWSJACK_SOURCE_DIR="$PWD" \
NEWSJACK_RUNTIMES=all \
sh ./install.sh
```

Assert:

- `~/.newsjack/bin/newsjack` exists and is executable
- `newsjack skills list` shows all expected skills
- generated runtime skill dirs exist
- canonical source bundle exists under `~/.newsjack/newsjack`
- repeated install is idempotent

### Layer 6: Site Verification

Keep current Next app checks:

```bash
pnpm lint
pnpm build
```

Local server checks:

```bash
curl -A 'curl/8.0' http://localhost:PORT/
curl -A 'Mozilla/5.0' http://localhost:PORT/
curl http://localhost:PORT/install.sh
```

Assert:

- curl user agent returns shell
- browser user agent returns HTML
- `/install.sh` returns shell

### Layer 7: Cross-Platform Build Checks

CI should build:

- `darwin/arm64`
- `darwin/amd64`
- `linux/amd64`
- `linux/arm64`

Before binary distribution, add:

- checksums
- install script checksum verification
- release artifact naming convention

## Pre-Ship Gates

Before replacing the installed shell shim with Go:

- Go CLI passes parity tests
- install smoke tests pass in temp `$HOME`
- actual local Codex setup still works
- actual local Claude Code setup still works or fails with a clear warning
- OpenClaw and Hermes paths have fixture coverage
- docs mention `newsjack`, not direct Python script paths

Before porting detector:

- golden fixtures exist for RSS, news search, HN, Reddit, X, monitor store, scoring, and output rendering
- legacy and Go outputs match normalized goldens
- fixture diffs are reviewed manually at least once
- downstream skill docs do not rely on removed fields

Before all-Go claim:

- no command depends on a second implementation runtime
- `newsjack doctor` reports missing credentials clearly without requiring Node.js
- Linux and macOS installs pass from clean temp homes
- update and uninstall paths are documented

## Recommendation

This branch executes the migration in one PR because the installer, runtime skill layout, auth bridge, detector command, filter application, and summarizer all cross the same public `newsjack` command contract.

Remaining follow-ups:

1. Decide whether npm ships platform binaries or remains a bootstrapper once curl v1 is live.
2. Add release checksums and artifact naming once binary distribution begins.
3. Expand golden coverage for live-source edge cases as source APIs evolve.
