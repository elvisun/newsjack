# newsjack

> **The open-source skills that turn your agent into a PR operator.**
>
> Install once. Your agent — Claude, Codex, Hermes, OpenClaw — becomes a PR team.

```bash
curl -fsSL newsjack.sh | bash
```

**New here?** Check out **[Getting started](docs/getting-started.md)**

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

```bash
curl -fsSL newsjack.sh | bash
```

Or review the script before running it:

```bash
curl -fsSL https://newsjack.sh
```

The installer detects your agent runtime and installs the skills automatically. Supported:

| Runtime                 | Auto-detected                     |
| ----------------------- | --------------------------------- |
| Claude Code             | ✅                                 |
| Codex                   | ✅                                 |
| OpenClaw                | ✅                                 |
| Hermes                  | ✅                                 |
| ChatGPT / Claude.ai web | manual (load `skills/*/SKILL.md`) |

### Claude Code plugin

If you're on Claude Code, you can install Newsjack as a plugin instead of running the script:

```text
/plugin marketplace add elvisun/newsjack
/plugin install newsjack@newsjack
```

This registers the repo as a marketplace and installs the `newsjack` plugin — all 16 skills are auto-discovered. The skills run instruction-only out of the box; steps that use the optional `newsjack` CLI ask before installing it.

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

Newsjack auto-updates from the latest GitHub Release before each run. Set `NEWSJACK_AUTO_UPDATE=0` to disable.
