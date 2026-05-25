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
NEWSJACK_RUNTIMES=codex \
curl -fsSL https://newsjack.sh/install.sh | sh
```

For production-path validation:

```bash
NEWSJACK_RUNTIMES=codex \
curl -fsSL https://newsjack.sh/install.sh | sh
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
npm install -g @openai/codex
npm install -g @anthropic-ai/claude-code
npm install -g openclaw@latest
npm install -g acpx@latest
npm install -g @zed-industries/codex-acp
npm install -g @zed-industries/claude-agent-acp
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
