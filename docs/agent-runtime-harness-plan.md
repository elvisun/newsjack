# Agent Runtime Harness Plan

## Goal

Build a repeatable harness that tests the real Newsjack installer and setup flow against real agent runtimes:

- Claude Code
- Codex
- Hermes
- OpenClaw

The harness must not install anything into the developer's host home directory. All runtime binaries, auth state, Newsjack state, transcripts, and generated artifacts must live inside disposable containers or explicit harness output directories.

## Non-goals

- Do not fake runtime binaries for the main test path.
- Do not mutate host `~/.codex`, `~/.claude`, `~/.hermes`, `~/.openclaw`, or `~/.newsjack`.
- Do not rely on terminal scraping as the primary control plane when ACP is available.
- Do not require browser-based login during automated runs.

## Design Summary

Use a container-first harness:

1. Build a Linux image with common runtime dependencies.
2. Install the real agent CLIs inside the image.
3. Run each test in a fresh container with an isolated `HOME`.
4. Execute the real installer through the public curl path or a local source override.
5. Drive agent turns through ACP when available.
6. Fall back to each runtime's non-interactive CLI only for smoke coverage.
7. Keep tmux as a last-resort fallback for interactive flows that cannot be exercised through ACP or non-interactive commands.

## Repository Layout

Proposed files:

```text
harness/
  README.md
  Dockerfile
  .env.example
  .env.local        # ignored; local token-burning integration credentials
  run.py
  acp_client.py
  runtime_probe.py
  assertions.py
  config.example.yaml
  adapters/
    __init__.py
    codex.py
    claude.py
    hermes.py
    openclaw.py
  prompts/
    install-smoke.md
    setup-flow.md
    detector-mock.md
  scripts/
    install-runtimes.sh
    run-in-container.sh
```

Output:

```text
harness/runs/
  2026-05-25T120000Z/
    codex/
      env.json
      installer.log
      acp-events.jsonl
      final-response.md
      artifacts/
    claude/
    hermes/
    openclaw/
```

`harness/runs/` should be gitignored.

Local token-burning integration credentials should live at:

```text
harness/.env.local
```

That file must never be committed. Start from `harness/.env.example` and pass it to Docker with:

```bash
docker run --rm --env-file harness/.env.local ...
```

## Container Contract

The host command should look like:

```bash
docker run --rm \
  --name newsjack-harness-codex \
  -v "$PWD:/repo:ro" \
  -v "$PWD/harness/runs:/runs" \
  -e OPENAI_API_KEY \
  -e ANTHROPIC_API_KEY \
  -e MEDIALYST_API_KEY \
  newsjack-agent-harness \
  python3 /harness/run.py --runtime codex --mode local-source
```

Runtime state stays inside the container:

```bash
export HOME=/tmp/newsjack-home
export XDG_CONFIG_HOME=/tmp/newsjack-home/.config
export XDG_CACHE_HOME=/tmp/newsjack-home/.cache
export XDG_DATA_HOME=/tmp/newsjack-home/.local/share
```

The repo is mounted read-only. Test artifacts go to `/runs`.

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

Use local-source mode in PR tests and production-path mode in scheduled tests.

## Runtime Installation

The image should install common tools:

- `bash`
- `curl`
- `git`
- `python3`
- `pipx`
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
pipx install hermes-agent
npm install -g acpx@latest
npm install -g @zed-industries/codex-acp
npm install -g @zed-industries/claude-agent-acp
```

Pin versions once the first green run establishes a baseline. Start unpinned only while discovering currently compatible versions.

## ACP Launch Strategy

Runtime ACP support changes quickly. The harness should probe supported launch commands instead of assuming one shape.

Probe order:

```json
{
  "codex": [
    ["codex", "--acp", "--stdio"],
    ["codex-acp"],
    ["acpx", "codex"]
  ],
  "claude": [
    ["claude", "--acp", "--stdio"],
    ["claude-agent-acp"],
    ["acpx", "claude"]
  ],
  "hermes": [
    ["hermes", "--acp", "--stdio"],
    ["hermes", "acp", "--accept-hooks"]
  ],
  "openclaw": [
    ["openclaw", "--acp", "--stdio"],
    ["openclaw", "acp"]
  ]
}
```

The chosen launch command must be recorded in `env.json`.

ACP probing should verify more than process startup:

1. Spawn the candidate command with clean stdio.
2. Send ACP `initialize`.
3. Create a new session with the test workspace cwd.
4. Send a trivial prompt.
5. Confirm at least one valid ACP response or update.

If every ACP candidate fails, mark ACP unavailable and run only the runtime's native CLI smoke test for that runtime.

## Native CLI Smoke Strategy

Native smoke tests protect against adapter drift. They should be short and cheap.

Commands:

```bash
codex exec --cd /workspace --sandbox danger-full-access --ask-for-approval never "Say READY."
claude --bare -p "Say READY."
hermes chat -Q -q "Say READY."
openclaw agent --local --message "Say READY." --json
```

These do not replace ACP setup-flow tests. They only prove the installed runtime binary can run in the container with the supplied credentials.

## Newsjack Installer Assertions

After `curl newsjack.sh | sh`, assert:

- `~/.newsjack/newsjack` exists.
- `~/.newsjack/bin/newsjack` exists and is executable.
- `~/.newsjack/bin/newsjack skills` lists the expected skills.
- Runtime-specific skill directories exist:
  - Codex: `~/.agents/skills`
  - Claude Code: `~/.claude/skills`
  - Hermes: `~/.hermes/skills`
  - OpenClaw: `~/.openclaw/skills`
- The selected runtime has `newsjack-setup`, `newsjack-detector`, and `media-list-manager`.
- MCP config path was attempted or written:
  - Codex: inspect `codex mcp list` when available.
  - Claude: inspect `claude mcp list` when available.
  - Hermes: inspect `~/.hermes/config.yaml`.
  - OpenClaw: inspect `openclaw mcp list` when available.

If `MEDIALYST_API_KEY` is present:

```bash
~/.newsjack/bin/newsjack login --key "$MEDIALYST_API_KEY"
~/.newsjack/bin/newsjack auth status
```

If no key is present, skip Medialyst-backed tests but still run skills-only and detector mock tests.

## Agent Setup-Flow Test

Prompt file: `harness/prompts/setup-flow.md`

Expected behavior:

1. Agent discovers or uses the installed `newsjack-setup` skill.
2. Agent creates a monitor profile for a fixture company.
3. Agent saves it to `/runs/<run>/<runtime>/artifacts/profile.json`.
4. Agent runs detector mock mode with that profile.
5. Agent writes a short result summary to `/runs/<run>/<runtime>/artifacts/result.md`.

Suggested prompt:

```text
You are testing the installed Newsjack skills.

Create a Newsjack monitor profile for Fixture Coffee, a specialty coffee company.
Save it exactly at ARTIFACT_DIR/profile.json.

Then run the local newsjack detector in mock mode with that profile.
Save a short Markdown result at ARTIFACT_DIR/result.md.

Do not ask follow-up questions. Make reasonable assumptions.
```

The harness replaces `ARTIFACT_DIR` before sending.

Assertions:

- `profile.json` exists.
- `profile.json` parses as JSON.
- Required profile fields exist.
- Detector mock command exits successfully.
- `result.md` exists and mentions whether the run succeeded.

## Detector Direct Test

The harness should also directly exercise the installed CLI without an agent:

```bash
~/.newsjack/bin/newsjack detector run \
  "specialty coffee" \
  --profile /repo/fixtures/newsjack-detector-agent/profile.bluebottle.json \
  --mock \
  --emit json
```

This separates installer/CLI problems from agent orchestration problems.

## Runtime-Specific Notes

### Codex

Primary path should be ACP if one of the candidate launchers works. Keep a native `codex exec` smoke test because Codex skill loading and MCP config are part of the Newsjack install surface.

Auth:

- Prefer `OPENAI_API_KEY`.

### Claude Code

Claude's ACP surface may be adapter-based depending on installed version. Keep a native `claude --bare -p` smoke test because the installer targets `~/.claude/skills`, and an ACP adapter may not exercise the same skill discovery path.

Auth:

- Prefer `ANTHROPIC_API_KEY`.
- Run with `--bare` for deterministic automation when using native smoke.

### Hermes

Hermes exposes `hermes acp`. It uses the same Hermes configuration and credentials as normal CLI mode, so isolated `HOME` is important.

Auth:

- Prefer provider configuration via env vars for automation.
- Use `--accept-hooks` in ACP mode to avoid headless hook prompts.

### OpenClaw

OpenClaw exposes `openclaw acp`, but it is Gateway-backed. The harness must either:

- start a local OpenClaw Gateway inside the container before ACP tests, or
- configure `openclaw acp` against a test Gateway URL.

Start with local Gateway mode. If Gateway setup is too heavy for PR tests, keep OpenClaw in native smoke mode for PRs and run full ACP setup-flow in scheduled CI.

## Test Modes

### `installer`

Runs only:

- container setup
- real runtime install
- `curl newsjack.sh | sh`
- installer assertions
- direct detector mock

No agent API spend unless runtime install/auth checks require it.

### `native-smoke`

Runs:

- installer mode
- one short native prompt per runtime

### `acp-smoke`

Runs:

- installer mode
- ACP launch probe
- trivial ACP prompt

### `setup-flow`

Runs:

- installer mode
- ACP launch probe
- full Newsjack setup-flow prompt
- artifact assertions

### `production-path`

Same as selected mode, but does not use `NEWSJACK_SOURCE_DIR=/repo`. It validates the public `newsjack.sh` path.

## CI Plan

PR checks:

```bash
harness/run.py --runtime codex --mode installer --local-source
harness/run.py --runtime claude --mode installer --local-source
harness/run.py --runtime hermes --mode installer --local-source
harness/run.py --runtime openclaw --mode installer --local-source
```

Nightly checks:

```bash
harness/run.py --runtime all --mode native-smoke --production-path
harness/run.py --runtime all --mode acp-smoke --production-path
```

Manual/release checks:

```bash
harness/run.py --runtime all --mode setup-flow --production-path
```

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

## Implementation Phases

### Phase 1: Container and Installer Assertions

Build `harness/Dockerfile`, `harness/run.py`, and installer assertions for one runtime at a time. Start with Codex because its native non-interactive mode is straightforward.

Done when:

- host remains clean
- container installs real Codex
- installer succeeds with `NEWSJACK_SOURCE_DIR=/repo`
- skill and CLI assertions pass

### Phase 2: All Runtime Installer Coverage

Add Claude, Hermes, and OpenClaw runtime installation and installer assertions.

Done when:

- all four runtimes pass installer mode in fresh containers
- each runtime records version and skill installation paths

### Phase 3: ACP Probe

Implement minimal ACP JSON-RPC client and capability probes.

Done when:

- ACP probe records the first working candidate per runtime
- failures are structured and readable
- native smoke fallback works when ACP is unavailable

### Phase 4: Agent Setup Flow

Add `setup-flow.md`, artifact assertions, and direct detector separation.

Done when:

- at least one runtime completes the full setup flow through ACP
- all generated artifacts are saved under `/runs`

### Phase 5: CI and Hardening

Add GitHub Actions workflow, version pins, log redaction, timeouts, and scheduled production-path checks.

Done when:

- PR checks do not require expensive agent calls
- scheduled checks exercise real public installer path
- manual workflow can run full setup flow with secrets

## Open Questions

- Which runtime versions should we pin for the first stable baseline?
- Should full OpenClaw ACP run in PR checks or only scheduled checks because of Gateway setup cost?
- Should `newsjack.sh` production-path checks run against `main` only, or also against a preview deployment for pull requests?
- Do we need a small JSON schema for monitor profiles to make setup-flow validation stricter?
- Should Medialyst MCP connection be tested with a real API call, or is `auth status` plus MCP config inspection enough for PR checks?
