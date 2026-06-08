# newsjack

> **The open-source skills that turn your agent into a PR operator.**
>
> Install once. Your local-first agent — Claude Code, Codex, Hermes, OpenClaw — becomes a PR team.

```bash
curl -fsSL newsjack.sh | bash
```

**New here?** Check out **[Getting started](docs/getting-started.md)**

---

## Runtime stance

Newsjack is optimized for local-first agent harnesses. The full product depends
on capabilities that browser chat products usually do not expose: shell
execution, filesystem storage, credentials, durable client profiles, JSON
artifacts, scheduled workflows, multi-agent orchestration, and cost-optimized
worker passes.

**Full Mode is the recommended path** and is available in capable agent harnesses:

- Claude Code
- Codex
- OpenClaw
- Hermes

Browser and restricted chat environments support **Limited Mode**:

- Claude.ai chat
- ChatGPT chat
- Claude Cowork

Limited Mode can still use Newsjack's instruction layer for PR strategy,
newsworthiness checks, pitch and angle generation, pitch critique,
journalist-fit reasoning, fact-checking from pasted or searchable evidence, and
best-effort manual news scans. It is not the canonical detector experience: no
saved monitors, scheduled runs, seen-state, deterministic freshness gates, full
source ingestion, local artifacts, or cost-optimized multi-agent passes.

---

## What your agent can do once newsjack is installed

Three problems, separate lanes.

### 🧭 Strategize — figure out what your story even is

- 🗺️ **Get a PR strategy** — opinionated walkthrough if you're not PR-fluent yet: audience first, positioning second, news pegs third, drumbeat over big-bang

### 🛰️ Detect — surface what matters in your space

- 📡 **Monitor your industry** — set up feeds, get fresh stories worth jumping on
- 🗞️ **Track your coverage** — Google Alerts-style keyword tracking with LLM filtering for real features
- 🔭 **Track competitors** — when they launch, raise, or stumble, you know
- 🔍 **Verify the story is still fresh** — who broke it, who owns it, what oxygen's left
- 📊 **Score newsworthiness** — cold read on whether it clears the bar before you act

### 🚀 Act — turn signal into output

- 🎯 **Generate story angles** — turn one update into hooks framed for different beats
- 🤝 **Fit-check a journalist** — will *this* reporter actually care, or are you spamming?
- 🎙️ **Respond to source queries** — triage inbound HARO-style requests, draft only the real fits
- 🥊 **Roast your pitch** — honest critique against the rubric editors actually use
- ✅ **Fact-check before you send** — extract claims, verify each, flag the shaky ones
- 🗣️ **Keep drafts in your voice** — fingerprint your real writing, kill the AI tells
- 📋 **Build a fit-checked media list** — targeted reporters, not scraped contact dumps

---

## Who this is for

- **Founders** doing their own PR because the agency quote was insane
- **PR agencies** running more accounts than humans can babysit
- **Marketers** at small companies who need leverage, not headcount
- **Anyone** whose agent is already running their day-to-day — and should be better at it

---

## Install

Pick the path for the harness you're using:

### Full Mode: Claude Code, Codex, OpenClaw, Hermes

Use the curl installer when your full agent harness can reach GitHub Release
assets:

```bash
curl -fsSL newsjack.sh | bash
```

Or review the script before running it:

```bash
curl -fsSL https://newsjack.sh
```

The installer detects your agent runtime and installs the CLI-backed skills
automatically.

### Limited Mode: Claude.ai, ChatGPT, Claude Cowork

Do not try to install the Newsjack CLI inside browser chat or restricted Cowork
surfaces. Use the marketplace/upload/manual skill path for a Limited Mode
experience.

For Claude.ai plugin-style setup:

1. Open **Customize**.
2. Go to **Personal plugins**.
3. Click **Create plugin**.
4. Choose **Add marketplace**.
5. Choose **Add from a repository**.
6. In the repository URL field, enter:

```text
elvisun/newsjack
```

Then install the Newsjack plugin from that marketplace. For full detector runs,
saved monitors, scheduled workflows, and artifacts, set up Newsjack in a Full
Mode harness instead.

### Blocked GitHub Releases / npm-only Full Mode environments

Use npm only when you are in a full agent harness but shell installers or GitHub
Release assets are blocked:

```bash
npm i -g newsjack
newsjack install
```

After install, agents and skills should call the CLI as `newsjack`.

### Supported runtimes

| Runtime             | Status      | Experience        | Install path |
| ------------------- | ----------- | ----------------- | ------------ |
| Claude Code         | Recommended | Full Mode         | curl, npm, or plugin |
| Codex               | Recommended | Full Mode         | curl or npm |
| OpenClaw            | Recommended | Full Mode         | curl or npm |
| Hermes              | Recommended | Full Mode         | curl or npm |
| Claude.ai chat      | Limited only | Limited Mode     | plugin/upload/manual skills |
| ChatGPT chat        | Limited only | Limited Mode     | manual skills |
| Claude Cowork       | Limited only | Limited Mode     | plugin/upload/manual skills |

### Claude Code plugin

If you're on Claude Code, you can install Newsjack as a plugin instead of running the script:

```text
/plugin marketplace add elvisun/newsjack
/plugin install newsjack@newsjack
```

This registers the repo as a marketplace and installs the `newsjack` plugin —
all 16 skills are auto-discovered. Use `newsjack setup` or the curl/npm path
above when you want the full CLI-backed detector, monitor, and artifact
workflow.

Once Newsjack is approved for the community marketplace, you'll also be able to install it from there:

```text
/plugin marketplace add anthropics/claude-plugins-community
/plugin install newsjack@claude-community
```

---

## What gets installed

```
~/.newsjack/
├── bin/newsjack          # CLI
└── newsjack/             # managed bundle

skills installed to your detected runtime(s):
  coverage-tracker-setup  create a keyword coverage tracker
  coverage-tracker        find real keyword features and suppress repeat alerts
  newsjack-monitor-setup  create a company monitoring profile
  newsjack-detector       find newsworthy moments before the wave breaks
  story-origin-check      verify a story is fresh, not already saturated
  newsworthiness-check    score whether your news clears the bar
  news-search             dated, attributed news search (Medialyst or web fallback)
  meanest-editor          roast your pitch against the editor rubric
  angle-generator         turn one update into ten pitchable hooks
  media-list-manager      build small, fit-checked media lists
```

The npm package bundles the same CLI and skills, with the `newsjack` command installed on `PATH`.

Curl-installed Newsjack auto-updates from the latest GitHub Release before each run. Set `NEWSJACK_AUTO_UPDATE=0` to disable.

Npm-installed Newsjack uses npm for CLI updates:

```bash
npm i -g newsjack@latest
```
