# Newsjack.sh

**The open-source skills that turn your agent into a full PR team.**

Install once. Your agent — Claude Code, Codex, Hermes, OpenClaw — becomes a PR team.

```bash
curl -fsSL newsjack.sh | bash
```

**Are you an agent?** Check out **[Getting started](docs/getting-started.md)**

---

## What your agent can do once newsjack is installed

Three problems, separate lanes.

### 🛰️ Detect — surface what matters in your space

- 📡 **Monitor your industry** — find newsjacking opportunities: fresh stories you have the standing to jump on before the wave breaks
- 🗞️ **Track your coverage** — Google Alerts-style keyword tracking with LLM filtering for real features
- 🔭 **Track competitors** — when they launch, raise, or stumble, you know
- 🔍 **Verify the story is still fresh** — who broke it, who owns it, what oxygen's left

### 🚀 Act — turn signal into output

- 🎯 **Generate story angles** — turn one update into hooks framed for different beats
- 🤝 **Fit-check a journalist** — will *this* reporter actually care, or are you spamming?
- 🎙️ **Respond to source queries** — triage inbound HARO-style requests, draft only the real fits
- 🥊 **Roast your pitch** — honest critique against the rubric editors actually use
- ✅ **Fact-check before you send** — extract claims, verify each, flag the shaky ones
- 🗣️ **Keep drafts in your voice** — fingerprint your real writing, kill the AI tells
- 📋 **Build a fit-checked media list** — targeted reporters, not scraped contact dumps

### 🧭 Strategize — figure out what your story even is

- 🗺️ **Get a PR strategy** — opinionated walkthrough if you're not PR-fluent yet: audience first, positioning second, news pegs third, drumbeat over big-bang
- 📊 **Score newsworthiness** — cold read on whether it clears the bar before you act

---

## Who this is for

- **Founders** doing their own PR because the agency quote was insane
- **PR agencies** running more accounts than humans can babysit
- **Marketers** at small companies who need leverage, not headcount
- **Anyone** whose agent is already running their day-to-day — and should be better at it

---

## Install

Pick the path for the harness you're using.

### Supported runtimes

| Runtime             | Status       | Experience   | Install path |
| ------------------- | ------------ | ------------ | ------------ |
| Claude Code         | Recommended  | Full Mode    | curl, npm, or plugin |
| Codex               | Recommended  | Full Mode    | curl or npm |
| OpenClaw            | Recommended  | Full Mode    | curl or npm |
| Hermes              | Recommended  | Full Mode    | curl or npm |
| Claude.ai chat      | Limited only | Limited Mode | plugin/upload/manual skills |
| ChatGPT chat        | Limited only | Limited Mode | manual skills |
| Claude Cowork       | Limited only | Limited Mode | plugin/upload/manual skills |

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

If shell installers or GitHub Release assets are blocked, use npm instead:

```bash
npm i -g newsjack
newsjack install
```

After install, agents and skills should call the CLI as `newsjack`.

### Limited Mode: Claude.ai, ChatGPT, Claude Cowork

Do not try to install the Newsjack CLI inside browser chat or restricted Cowork
surfaces. Use the marketplace/upload/manual skill path instead.

Limited Mode keeps Newsjack's instruction layer for PR strategy, newsworthiness
checks, pitch and angle generation, pitch critique, journalist-fit reasoning,
fact-checking from pasted or searchable evidence, and best-effort manual news
scans. It does not include saved monitors, scheduled runs, seen-state,
deterministic freshness gates, full source ingestion, local artifacts, or
cost-optimized multi-agent passes — set up a Full Mode harness for those.

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

Then install the Newsjack plugin from that marketplace:

- Marketplace: `elvisun/newsjack`
- Plugin: `newsjack@newsjack`

A community marketplace plugin (`newsjack@claude-community`) is pending review
and will be added soon.

---

## What gets installed

```
~/.newsjack/
├── bin/newsjack          # CLI
└── newsjack/             # managed bundle

skills installed to your detected runtime(s):
  newsjack-monitor-setup  create a company monitoring profile
  newsjack-detector       find newsworthy moments before the wave breaks
  newsworthiness-check    score whether your news clears the bar
  angle-generator         turn one update into ten pitchable hooks
  find-journalists      build small, fit-checked media lists
  …and more
```

The npm package bundles the same CLI and skills, with the `newsjack` command installed on `PATH`.

Curl-installed Newsjack auto-updates from the latest GitHub Release before each run. Set `NEWSJACK_AUTO_UPDATE=0` to disable.

Npm-installed Newsjack uses npm for CLI updates:

```bash
npm i -g newsjack@latest
```
