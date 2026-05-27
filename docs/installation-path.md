# Ideal Newsjack Installation Path

The install path should get a user from zero to a monitored Newsjack workflow
with two commands:

```bash
curl -fsSL newsjack.sh | sh
newsjack setup
```

The installer owns deterministic machine setup. The agent harness owns the
judgment-heavy onboarding conversation. The user should never have to manually
copy skill files, edit MCP JSON, invent an agent routine, or figure out which
runtime path was detected.

Target end state:

- the `newsjack` command is available in the current shell or the installer
  prints one exact current-shell command to make it available
- Newsjack skills are installed into every detected supported harness
- optional Medialyst MCP is configured when a runtime exposes a reliable
  noninteractive setup path
- a monitor profile exists at `~/.newsjack/monitors/<slug>/profile.json`
- an hourly monitor is installed through the selected agent harness scheduler
- a mock test run has passed
- when any live source is available, a live test run has produced a
  human-facing `run.md`

## Installer Contract

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

## Setup Contract

`newsjack setup` is the frictionless onboarding path after install. It should
be usable from a normal terminal and from inside an agent harness.

Required behavior:

- present supported skill runtimes and let the user choose where skills should
  be copied: Codex, Claude Code, OpenClaw, Hermes, all, or other/manual
- for other/manual runtimes, print a copyable instruction telling the agent to
  copy every `skills/*/SKILL.md` directory into its own skill path and then use
  `newsjack setup --json` for local paths
- ask separately which agent harness should own scheduled runs, with default
  recommendation order: Hermes, OpenClaw, Claude Code, Codex, other/manual
- install schedules inside the selected agent harness, not system cron
- if Claude Code is missing, ask for explicit permission and then run the
  official Claude Code native installer:
  `curl -fsSL https://claude.ai/install.sh | bash`
- verify `newsjack` is on PATH for that harness; if not, pass an absolute path
- verify skills are visible to the chosen harness
- run `newsjack doctor` and show only actionable problems
- guide the user through a monitor profile using the `newsjack-setup` skill
- write the profile to `~/.newsjack/monitors/<slug>/profile.json`
- ask for optional X and Medialyst credentials with clear explanations, links,
  and skip options
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

## Dependency And Credential Timing

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
  source or workflow that needs them. Always offer a skip path. X bearer tokens
  are saved to `~/.newsjack/.env` with user-only permissions; Medialyst keys are
  saved to `~/.newsjack/credentials.json`.
- During `newsjack monitor test <slug> --mock`: ask for no secrets. This is the
  guaranteed first success path.
- During `newsjack monitor test <slug> --live`: if no live source is available,
  offer clear choices: run RSS/public sources only, add Medialyst, set up X, or
  skip the live test.
- During agent-scheduled runs: never prompt. Use configured sources, degrade
  gracefully, and write dependency status into the run artifacts.

Credential rules:

- Medialyst is optional. Ask for `MEDIALYST_API_KEY` during setup, explain that
  it powers live news search and MCP-backed media-list workflows, link to
  `https://medialyst.ai/docs`, and allow skip. Save it with `newsjack login`
  semantics in `~/.newsjack/credentials.json` using user-only permissions.
  `MEDIALYST_API_BASE` and `MEDIALYST_NEWS_PATH` are advanced overrides, not
  onboarding prompts.
- X is optional. Ask for an X bearer token during setup, explain that it powers
  X News, X trends, and X post search, link to
  `https://docs.x.com/fundamentals/authentication/oauth-2-0/bearer-tokens`, and
  allow skip. Newsjack should call the X API directly rather than depending on
  an external X CLI.
- X bearer tokens are optional source configuration. Accept `X_BEARER_TOKEN`,
  `TWITTER_BEARER_TOKEN`, `X_API_BEARER_TOKEN`, or
  `TWITTER_API_BEARER_TOKEN` from the environment or dotenv. Do not ask for or
  save bearer tokens in the pipe installer.

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
- Optional X support: X bearer-token env vars read by Newsjack, used for X post
  search, X News, and trends through direct API calls.
- Optional MCP bridge: Node.js with `npx` because `newsjack mcp-bridge`
  launches `npx -y mcp-remote`. Missing `npx` should disable only the bridge,
  not the detector or agent scheduler.
- Optional agent harnesses: Codex, Claude Code, Cursor Agent, OpenCode,
  OpenClaw, Hermes, or other runtimes. If none are detected, install portable
  skills and keep the CLI usable.
- Optional agent scheduler: OpenClaw cron, Hermes cron, Claude Code Routine, or
  a similar supported harness scheduler. System `launchd`, user `systemd`, and
  `crontab` are not v1 scheduler targets because they cannot reliably trigger
  the agent harness to perform LLM analysis after data collection.
- Development-only dependencies: Docker for the harness, Go for local CLI
  builds, and Node/pnpm for the website and distribution build.

The default recommended setup should be no-key first: create the monitor, add
RSS feeds or public sources when possible, run the mock test, then offer
Medialyst and X as coverage upgrades before the live test.

## Monitor Commands

The public CLI should make the recurring workflow explicit:

```bash
newsjack monitor init
newsjack monitor test <slug> --mock
newsjack monitor test <slug> --live
newsjack monitor schedule <slug> --runtime <agent-runtime> --every 1h
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

## Agent Scheduler Contract

The v1 scheduler should run inside an agent harness, not through system cron.
The recurring job must be able to trigger both deterministic data collection and
agent-side LLM analysis over the collected evidence.

Preferred scheduler targets:

- OpenClaw cron
- Hermes cron
- Claude Code Routine
- similar agent-native scheduling in a supported harness

System `launchd`, user `systemd`, and `crontab` are out of scope for v1
recurring monitors. They can run the detector, but they cannot reliably trigger
the user's chosen agent harness to do the LLM analysis step and produce the
final opportunity report. A system scheduler can be revisited later for a
detector-only or fully headless mode.

Every agent-scheduled run must:

- run the deterministic detector and save raw candidates
- invoke the installed Newsjack skill in the selected harness for analysis
- use a lock file so hourly jobs cannot overlap
- write stdout/stderr to monitor-local logs
- preserve the environment needed by the detector
- avoid auto-sending or auto-scheduling journalist outreach
- produce an inspectable `run.md`, even when no opportunities are found

If multiple harness schedulers are available, `newsjack setup` should recommend
Hermes first, then OpenClaw, then Claude Code, then Codex. Users can override
with `newsjack setup --schedule-runtime openclaw`, `hermes`, `claude`, or
`codex`. If no supported agent scheduler is available, setup should still copy
skills, save optional credentials, then print one exact manual prompt for the
user to run inside their agent harness.

## Research Inputs

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
- Scheduled agent features are the v1 recurring path because Newsjack needs the
  harness to run LLM analysis, not only the local detector.
