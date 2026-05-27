# Newsjack Install and Agent Harness

This document has two jobs:

- define the ideal `curl -fsSL newsjack.sh | sh` installation and onboarding
  path
- keep the disposable harness commands for testing the real installer

The harness intentionally does not contain a second product implementation. The
installer should always exercise the compiled `newsjack` binary and the real
runtime skill install paths.

## Ideal Installation Path

The install path should get a user from zero to a monitored Newsjack workflow
with two commands:

```bash
curl -fsSL newsjack.sh | sh
newsjack setup
```

The installer owns deterministic machine setup. The agent harness owns the
judgment-heavy onboarding conversation. The user should never have to manually
copy skill files, edit MCP JSON, invent a cron command, or figure out which
runtime path was detected.

Target end state:

- the `newsjack` command is available in the current shell or the installer
  prints one exact current-shell command to make it available
- Newsjack skills are installed into every detected supported harness
- optional Medialyst MCP is configured when a runtime exposes a reliable
  noninteractive setup path
- a monitor profile exists at `~/.newsjack/monitors/<slug>/profile.json`
- an hourly monitor is installed through the best local scheduler for the
  platform
- a mock test run has passed
- when credentials are present, a live test run has produced a human-facing
  `run.md`

### Installer Contract

`curl -fsSL newsjack.sh | sh` should be short, deterministic, idempotent, and
safe to re-run.

Required behavior:

- detect OS, architecture, libc shape where relevant, and unsupported platforms
  with actionable errors
- resolve the selected channel or version
- download a prebuilt artifact; never require Go on the user's machine
- verify the artifact checksum, and later verify a signed manifest
- install into a managed, versioned layout, then atomically update symlinks
- keep a previous working install until the new one verifies
- install or update `~/.newsjack/bin/newsjack`
- add PATH with marked shell-config blocks, or print a single exact fallback
  command when shell config cannot be updated
- install completions when the CLI supports them
- detect existing runtime installs and configure all supported runtimes unless
  `NEWSJACK_RUNTIMES` narrows the list
- install only instruction/data files into runtime skill directories
- configure MCP as best effort; warn instead of failing the install
- run a quiet `newsjack doctor` equivalent and summarize the result
- end with one primary next step: `newsjack setup`

Installer controls should mirror current agent CLI conventions:

```bash
curl -fsSL newsjack.sh | sh
curl -fsSL newsjack.sh | NEWSJACK_RUNTIMES=codex,claude sh
curl -fsSL newsjack.sh | NEWSJACK_RUNTIMES=all sh
curl -fsSL newsjack.sh | NEWSJACK_INSTALL_MCP=0 sh
curl -fsSL newsjack.sh | NEWSJACK_VERSION=<version-or-commit> sh
curl -fsSL newsjack.sh | NEWSJACK_NO_MODIFY_PATH=1 sh
NEWSJACK_AUTO_UPDATE=0 newsjack doctor
```

The installer output should look closer to modern agent CLIs than to a package
manager log:

```text
newsjack installed

Detected harnesses:
  ok Claude Code
  ok Codex
  - Cursor Agent

Installed:
  CLI      ~/.newsjack/bin/newsjack
  Skills   ~/.claude/skills, ~/.agents/skills
  MCP      medialyst configured for Codex; Claude Code skipped, login needed

Next:
  newsjack setup
```

### Setup Contract

`newsjack setup` is the frictionless onboarding path after install. It should
be usable from a normal terminal and from inside an agent harness.

Required behavior:

- detect supported harnesses and pick the only detected harness automatically
- if multiple harnesses are detected, ask the user to choose one
- verify `newsjack` is on PATH for that harness; if not, pass an absolute path
- verify skills are visible to the chosen harness
- run `newsjack doctor` and show only actionable problems
- guide the user through a monitor profile using the `newsjack-setup` skill
- write the profile to `~/.newsjack/monitors/<slug>/profile.json`
- ask for optional credentials only after the user chooses sources that need
  them
- install a schedule using the platform's best local scheduler
- run `newsjack monitor test <slug> --mock`
- if any live no-key or credentialed source is available, run
  `newsjack monitor test <slug> --live`
- print the final `run.md` path and the schedule status

The setup flow should be agent-native but not agent-dependent. A user should be
able to say this in Claude Code, Codex, Cursor Agent, OpenCode, OpenClaw, or
Hermes:

```text
Use Newsjack to set up an hourly monitor for my company, install the schedule,
and run a test. Ask only for facts you cannot infer safely.
```

The agent should then use the installed skills and CLI instead of making up a
workflow.

### Dependency And Credential Timing

The installer should never ask for API keys. `curl | sh` should finish with no
interactive secret handling and no dependency prompts beyond the tools required
to download and unpack the binary. Optional integrations are discovered and
reported, then activated during `newsjack setup` or the first live test that
needs them.

Prompt timing:

- During `curl -fsSL newsjack.sh | sh`: ask for no secrets. Detect optional
  tools, install the CLI and skills, configure best-effort MCP, and end with
  `newsjack setup`.
- During `newsjack setup`: ask for credentials only after the user opts into a
  source or workflow that needs them. Always offer a skip path.
- During `newsjack monitor test <slug> --mock`: ask for no secrets. This is the
  guaranteed first success path.
- During `newsjack monitor test <slug> --live`: if no live source is available,
  offer clear choices: run RSS/public sources only, add Medialyst, set up X, or
  skip the live test.
- During scheduled runs: never prompt. Use configured sources, degrade
  gracefully, and write dependency status into the run artifacts.

Credential rules:

- Medialyst is optional. Ask for `MEDIALYST_API_KEY` only when the user chooses
  Medialyst news search, media-list generation, or MCP-backed media-list
  workflows. Save it with `newsjack login` in `~/.newsjack/credentials.json`
  using user-only permissions. `MEDIALYST_API_BASE` and `MEDIALYST_NEWS_PATH`
  are advanced overrides, not onboarding prompts.
- X is optional. Ask about X only when the user chooses `x`, `x_news`, or
  `x_trends` lanes. Prefer `xurl auth oauth2 login` for normal X setup because
  Newsjack can reuse xurl's OAuth state without storing an X secret.
- X bearer tokens are optional advanced configuration. Accept
  `X_BEARER_TOKEN`, `TWITTER_BEARER_TOKEN`, `X_API_BEARER_TOKEN`, or
  `TWITTER_API_BEARER_TOKEN` from the environment or dotenv when app-auth
  endpoints are needed, especially location trends. Do not ask for or save
  bearer tokens in the pipe installer.

Dependency classes:

- Required for install: POSIX `sh`, `curl` or `wget`, `tar`, `sha256sum` or
  `shasum`, and a supported macOS/Linux arm64/amd64 platform.
- Required for normal use: the prebuilt `newsjack` binary and network access to
  configured sources. Users do not need Go, Node, Docker, jq, pnpm, or a system
  SQLite binary.
- Optional live sources with no Newsjack API key: RSS/Atom feeds, public Reddit
  search, and Hacker News search.
- Optional Medialyst support: `MEDIALYST_API_KEY` for `news_search` and
  Medialyst MCP/media-list workflows.
- Optional X support: `xurl` plus its OAuth login for X post search,
  X News, and personalized trends; bearer-token env vars for direct app-auth
  calls and location trends.
- Optional MCP bridge: Node.js with `npx` because `newsjack mcp-bridge`
  launches `npx -y mcp-remote`. Missing `npx` should disable only the bridge,
  not the detector or scheduler.
- Optional agent harnesses: Codex, Claude Code, Cursor Agent, OpenCode,
  OpenClaw, Hermes, or other runtimes. If none are detected, install portable
  skills and keep the CLI usable.
- Optional scheduler tools: `launchctl`, user `systemd`, or `crontab`, needed
  only when installing the monitor schedule.
- Development-only dependencies: Docker for the harness, Go for local CLI
  builds, and Node/pnpm for the website and distribution build.

The default recommended setup should be no-key first: create the monitor, add
RSS feeds or public sources when possible, run the mock test, then offer
Medialyst and X as coverage upgrades before the live test.

### Monitor Commands

The public CLI should make the recurring workflow explicit:

```bash
newsjack monitor init
newsjack monitor test <slug> --mock
newsjack monitor test <slug> --live
newsjack monitor schedule <slug> --every 1h
newsjack monitor run <slug>
newsjack monitor status <slug>
newsjack monitor open <slug>
```

`monitor run` should write a timestamped run folder under:

```text
~/.newsjack/monitors/<slug>/runs/<timestamp>/
  candidates.json
  detector.stderr.log
  commands.log
  summary.json
  run.md
```

`run.md` is the human-facing artifact. JSON and log files are support artifacts.

### Scheduler Contract

The default scheduler should be local and boring:

- macOS: `launchd`
- Linux with user systemd: user `systemd` timer
- Linux without user systemd, WSL, and generic Unix: cron fallback

Every scheduled run must:

- use a lock file so hourly jobs cannot overlap
- write stdout/stderr to monitor-local logs
- preserve the environment needed by the detector
- avoid auto-sending or auto-scheduling journalist outreach
- produce an inspectable `run.md`, even when no opportunities are found

Native agent scheduling is an optional enhancement, not the default v1 path.
Claude Routines, Claude `/loop`, Codex Automations, Cursor Cloud Agents, and
similar features are still moving quickly and vary by plan, cloud/local file
access, and whether the local machine must be awake. Newsjack can print
runtime-specific instructions later, but the reliable base path should remain a
local scheduler plus a deterministic CLI run.

### Research Inputs

Current agent installer patterns worth copying:

- Grok CLI `https://x.ai/cli/install.sh`: channel selection, deployment-key
  config, aliases, completions, shell profile blocks, and immediate PATH help.
- Claude Code `https://claude.ai/install.sh`: checksum verification, channel
  and version arguments, native auto-update, and a strong `doctor` story.
- Codex CLI `https://chatgpt.com/codex/install.sh`: release locks, stale lock
  cleanup, conflict detection for npm/Homebrew installs, checksum verification,
  and atomic symlink switching.
- Cursor Agent `https://cursor.com/install`: concise first-run output and a
  clear next step.
- OpenCode `https://opencode.ai/install`: `--version`, `--binary`,
  `--no-modify-path`, XDG-aware config probing, and local binary install for
  testing.

Agent onboarding patterns to honor:

- Skills are the on-demand workflow layer.
- MCP is the tool/data-access layer.
- `AGENTS.md`, `CLAUDE.md`, and runtime rules orient the agent, but should stay
  short and point to deeper skill docs.
- Scheduled agent features are useful for advanced users, but local scheduling
  remains the lowest-friction cross-runtime base.

## Harness Verification

### Build The Harness Image

From the repo root:

```bash
docker build -f harness/Dockerfile -t newsjack-agent-harness:local .
```

Open a shell with the repo mounted:

```bash
docker run --rm -it \
  -v "$PWD:/repo" \
  -w /repo \
  newsjack-agent-harness:local \
  bash
```

Inside the container, isolate runtime state:

```bash
export HOME=/tmp/newsjack-home
export XDG_CONFIG_HOME="$HOME/.config"
export XDG_CACHE_HOME="$HOME/.cache"
export XDG_DATA_HOME="$HOME/.local/share"
export PATH="$HOME/.newsjack/bin:$PATH"
rm -rf "$HOME"
```

`curl | sh` runs in a child shell, so it cannot update the parent shell's
`PATH`. Keep the `PATH` export above in the interactive shell, or run the
binary by absolute path:

```bash
"$HOME/.newsjack/bin/newsjack" version
```

### Local Source Path

Use this when iterating on `install.sh`, skills, or the CLI before deploying.
The public installer no longer builds Go on user machines, so local source
installs must provide a compiled binary explicitly.

Inside the container:

```bash
mkdir -p /tmp/newsjack-build
(cd apps/cli && CGO_ENABLED=0 go build -trimpath -buildvcs=false -o /tmp/newsjack-build/newsjack ./cmd/newsjack)

NEWSJACK_SOURCE_DIR=/repo \
NEWSJACK_CLI_BINARY=/tmp/newsjack-build/newsjack \
NEWSJACK_RUNTIMES=all \
NEWSJACK_INSTALL_MCP=1 \
sh ./install.sh

hash -r
command -v newsjack
file "$(command -v newsjack)"
newsjack version
newsjack skills list
newsjack doctor | jq .
```

Expected:

- `file "$(command -v newsjack)"` reports an ELF executable, not a shell script.
- Skills land under the temp home runtime dirs, for example
  `$HOME/.agents/skills`, `$HOME/.claude/skills`, and `$HOME/.openclaw/skills`.
- MCP setup either configures detected runtimes or logs non-blocking warnings.

### Local Hosted-Dist Path

Use this to test the same shape as production before pushing: the site serves
`/install.sh` and `/dist`, and the container installs via HTTP.

On the host, from the repo root:

```bash
pnpm --dir apps/site run build
pnpm --dir apps/site exec next start --port 3010
```

In another terminal, start the harness container. On Docker Desktop for macOS,
the host is reachable as `host.docker.internal`:

```bash
docker run --rm -it \
  -v "$PWD:/repo" \
  -w /repo \
  newsjack-agent-harness:local \
  bash
```

Inside the container:

```bash
export HOME=/tmp/newsjack-home
export XDG_CONFIG_HOME="$HOME/.config"
export XDG_CACHE_HOME="$HOME/.cache"
export XDG_DATA_HOME="$HOME/.local/share"
export PATH="$HOME/.newsjack/bin:$PATH"
rm -rf "$HOME"

curl -fsSL http://host.docker.internal:3010 | \
  NEWSJACK_DIST_BASE=http://host.docker.internal:3010/dist \
  NEWSJACK_RUNTIMES=all \
  NEWSJACK_INSTALL_MCP=1 \
  sh

hash -r
file "$(command -v newsjack)"
newsjack version
newsjack doctor | jq .
```

On Linux hosts, add Docker's host gateway mapping when starting the container:

```bash
docker run --rm -it \
  --add-host=host.docker.internal:host-gateway \
  -v "$PWD:/repo" \
  -w /repo \
  newsjack-agent-harness:local \
  bash
```

### Production Path

Use this after a push/deploy to verify the live domain end to end:

```bash
export HOME=/tmp/newsjack-home
export XDG_CONFIG_HOME="$HOME/.config"
export XDG_CACHE_HOME="$HOME/.cache"
export XDG_DATA_HOME="$HOME/.local/share"
export PATH="$HOME/.newsjack/bin:$PATH"
rm -rf "$HOME"

curl -fsSL newsjack.sh | \
  NEWSJACK_RUNTIMES=all \
  NEWSJACK_INSTALL_MCP=1 \
  sh

hash -r
command -v newsjack
file "$(command -v newsjack)"
newsjack version
newsjack skills list
newsjack doctor | jq .
```

To inspect the deployed channel:

```bash
curl -fsSL https://newsjack.sh/dist/channels/main.txt
curl -fsSL https://newsjack.sh/dist/manifest.json | jq .
```

### Auto-Update Observation

Installed binaries auto-update from the hosted `main` channel before normal
user-facing commands. To force that path in the container:

```bash
printf 'stale-version\n' > "$HOME/.newsjack/newsjack/VERSION"
newsjack doctor > /tmp/newsjack-doctor.json 2> /tmp/newsjack-update.log

cat /tmp/newsjack-update.log
jq . /tmp/newsjack-doctor.json
cat "$HOME/.newsjack/newsjack/VERSION"
```

Expected:

- stderr shows `newsjack: auto-updating ...`.
- stdout remains valid JSON for `doctor`.
- `VERSION` is rewritten to the live channel commit.

Disable auto-update for deterministic debugging:

```bash
NEWSJACK_AUTO_UPDATE=0 newsjack doctor | jq .
```

### Notes

- Put installer environment variables on the `sh` side of the pipe:
  `curl -fsSL newsjack.sh | NEWSJACK_RUNTIMES=all sh`.
- Do not use `NEWSJACK_RUNTIMES=all curl ... | sh`; that only sets the
  variable for `curl`, not for the installer shell.
- The old `harness/run.py` workflow is gone. If a scripted harness returns,
  keep it as a thin shell or Go wrapper around these commands and the compiled
  `newsjack` binary.
