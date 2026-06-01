# Agent Runtime Harness Plan

This plan is retained as product intent, but the old checked-in runner has been
removed. Newsjack should not carry a second implementation surface beside the Go
CLI.

## Goal

Build a repeatable harness that tests the real Newsjack installer and setup flow
against real agent runtimes:

- Claude Code
- Codex
- Hermes
- OpenClaw

The harness must not install anything into the developer's host home directory.
All runtime binaries, auth state, Newsjack state, transcripts, and generated
artifacts must live inside disposable containers or explicit harness output
directories.

## Current State

The active verification surface is:

```bash
(cd apps/cli && go test ./...)
./bin/newsjack detector run "AI search visibility" --mock --limit 1 --emit json
```

The `harness/` directory may keep container assets and prompts, but there is no
checked-in harness runner. If runtime coverage returns, implement it as Go or
shell orchestration around `newsjack`.

## Non-Goals

- Do not fake runtime binaries for the main test path.
- Do not mutate host `~/.codex`, `~/.claude`, `~/.hermes`, `~/.openclaw`, or
  `~/.newsjack`.
- Do not rely on terminal scraping as the primary control plane when ACP is
  available.
- Do not require browser-based login during automated runs.
- Do not duplicate detector, auth, MCP, filtering, or summary behavior outside
  the Go CLI.

## Design Summary

Use a container-first harness:

1. Build a Linux image with common runtime dependencies.
2. Install the real agent CLIs inside the image.
3. Run each test in a fresh container with an isolated `HOME`.
4. Execute the real installer through the public curl path or a local source
   override.
5. Drive agent turns through ACP when available.
6. Fall back to each runtime's non-interactive CLI only for smoke coverage.
7. Keep tmux as a last-resort fallback for interactive flows that cannot be
   exercised through ACP or non-interactive commands.

## Container Contract

Runtime state stays inside the container:

```bash
export HOME=/tmp/newsjack-home
export XDG_CONFIG_HOME=/tmp/newsjack-home/.config
export XDG_CACHE_HOME=/tmp/newsjack-home/.cache
export XDG_DATA_HOME=/tmp/newsjack-home/.local/share
```

For local installer iteration:

```bash
NEWSJACK_SOURCE_DIR=/repo \
NEWSJACK_CLI_BINARY=/repo/.tmp/newsjack \
NEWSJACK_RUNTIMES=codex \
sh ./install.sh
```

For production-path validation:

```bash
curl -fsSL newsjack.sh | NEWSJACK_RUNTIMES=codex sh
```

## Runtime Installation

The image should install common tools:

- `bash`
- `curl`
- `git`
- `node`
- `npm`
- `jq`
- `ripgrep`
- `tmux`

Install real runtimes inside the image:

```bash
harness/scripts/build-image.sh --harness all
harness/scripts/build-image.sh --harness claude --harness openclaw --image newsjack-agent-harness:claude-openclaw
```

Pin versions once the first green run establishes a baseline.

## Failure Output

Every failure should include:

- runtime
- mode
- selected launch path
- container image digest
- runtime versions
- installer log path
- ACP event log path if applicable
- artifact directory path
- final assertion error

Do not print secrets. Redact values for variables matching:

- `*_API_KEY`
- `*_TOKEN`
- `*_SECRET`
- `*_PASSWORD`

Local token-burning tests may load `harness/.env.local` with
`harness/scripts/docker-shell.sh --with-local-env` or
`harness/scripts/run-ci-installer.sh --with-local-env`. Env files are passed at
`docker run` time only, validated as git-ignored when they live in the repo, and
excluded from Docker build context by `.dockerignore`.
