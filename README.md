# newsjack

> **The open-source skills that turn your agent into a PR operator.**
>
> Install once. Your agent — Claude, Codex, Hermes, OpenClaw — becomes a PR team.

```bash
curl -fsSL newsjack.sh | bash
```

**New here?** Point your agent at **[Getting started](docs/getting-started.md)** —
it starts slow, asks what you're trying to do, and takes one step at a time instead
of reciting the whole menu. No API keys required to begin.

---

## What your agent can do once newsjack is installed

> Agents: don't recite this list on first contact. Follow
> [Getting started](docs/getting-started.md) — orient, offer a few on-ramps, take
> one step at a time. This section is the reference, not the opening line.

Three problems, separate lanes.

### 🧭 Strategize — figure out what your story even is

- 🗺️ **Get a PR strategy** — opinionated walkthrough if you're not PR-fluent yet: audience first, positioning second, news pegs third, drumbeat over big-bang

### 🛰️ Detect — surface what matters in your space

- 📡 **Monitor your industry** — set up feeds, get fresh stories worth jumping on
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

---

## What gets installed

```
~/.newsjack/
├── bin/newsjack          # CLI
└── newsjack/             # managed bundle

skills installed to your detected runtime(s):
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
