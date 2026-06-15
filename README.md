# Newsjack.sh

**The open-source skills that turn your agent into a full PR team.**

Install once. Your agent — Claude Code, Codex, Hermes, OpenClaw — becomes a PR team.

```bash
curl -fsSL newsjack.sh | bash
```

**Are you an agent?** Check out **[Getting started](docs/getting-started.md)**

**Are you a human?** Jump to **[platform-specific setup](#install)** below.

---

## What your agent can do once newsjack is installed

Three problems, separate lanes.

### 🛰️ Detect — surface what matters in your space

- 📡 **Monitor your industry** — find newsjacking opportunities: fresh stories you have the standing to jump on before the wave breaks ([see a sample run](docs/example-run.md))
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

Newsjack is a set of **open skills** — plain-Markdown instructions your agent
reads — plus a small open-source CLI. Most skills run anywhere your agent runs.
A few reach for a live news index, a media database, real journalist contacts,
or locally-saved monitoring state — those work best in a local agent.

### What runs where

| Skill | [Claude.ai](https://claude.ai) | [ChatGPT](https://chatgpt.com) | [Cowork](https://claude.com/product/cowork) | [Claude Code](https://claude.com/claude-code) | [Codex](https://openai.com/codex) | [Hermes](https://hermes-agent.nousresearch.com) | [OpenClaw](https://openclaw.ai) | [Medialyst](https://medialyst.ai) |
| --- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| **Strategize** | | | | | | | | |
| pr-strategist | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| newsworthiness-check | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Act** | | | | | | | | |
| angle-generator | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| headline-generator | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| meanest-editor | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| crisis-holding | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| reactive-comment | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| fact-check | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| journalist-fit-check | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| voice-extractor | ⚠️ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| find-journalists | 🔧 | ⚠️ | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | ✅ |
| **Detect** | | | | | | | | |
| news-search | ✅ | ⚠️ | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | ✅ |
| story-origin-check | 🔧 | ⚠️ | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | ✅ |
| relevance-coarse-filter | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| newsjack-triage | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| newsjack-detector | ⚠️ | ⚠️ | ⚠️ | 🔧 | 🔧 | 🔧 | 🔧 | 🔜 |
| newsjack-monitor-setup | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | 🔜 |
| coverage-tracker | ⚠️ | ⚠️ | ⚠️ | ✅ | ✅ | ✅ | ✅ | 🔜 |
| coverage-tracker-setup | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | 🔜 |


Legend:

- **✅ Runs out of the box** — no setup; works anywhere your agent does.
- **🔧 May need an external connection** — works on its own, but does its best work with an external data source connected (e.g. the X API, or the Medialyst API / MCP).
- **⚠️ Limited Mode** — runs in a chat app, but as a best-effort, one-shot pass with nothing saved between sessions: no stored voice fingerprint, no scheduled monitoring, no repeat-suppression. Connect a local agent for the saved, scheduled version.
- **🔜 Coming soon** — not available here yet.
- **❌ Not supported** — the two setup skills (`newsjack-monitor-setup`, `coverage-tracker-setup`) only save a profile or config and schedule it, so they need a local agent.

**ChatGPT** is ⚠️ across the board: Skills are in beta for **ChatGPT Business
and Enterprise** only, so other tiers can't load Newsjack skills natively (you
paste them by hand), and there's no local CLI or saved state either way.

**Set up your agent:**

- **[Local agents](#local-agents-claude-code-codex-hermes-openclaw)** — Claude Code, Codex, Hermes, OpenClaw
- **[Claude.ai & Cowork](#claudeai--cowork)** — Anthropic plugin + Medialyst MCP
- **[ChatGPT](#chatgpt)** — Skills beta (ChatGPT Business / Enterprise)

### Local agents (Claude Code, Codex, Hermes, OpenClaw)

**Technical?** One line on macOS / Linux (review [`install.sh`](install.sh) first if you like):

```bash
curl -fsSL newsjack.sh | bash
```

**Not technical?** 📋 **Copy this prompt to any AI:**

```text
help me setup https://newsjack.sh
```

It reads the [guide](docs/getting-started.md) and handles the rest — npm
fallback, Windows, credentials — on any platform.

### Claude.ai & Cowork - Install as a Claude Plugin

No install, no CLI — paste your startup URL or your news and you get a real PR
operator on the spot.

**Install the Newsjack plugin:**

1. Open **[Customize](https://claude.ai/customize)**.
2. Go to **Personal plugins**.
3. Click **Create plugin**.
4. Choose **Add marketplace**.
5. Choose **Add from a repository**.
6. In the **repository URL** field, enter `elvisun/newsjack` and confirm.
7. Open the new **`elvisun/newsjack`** marketplace, find the **`newsjack@newsjack`**
   plugin, and click **Install**.
8. *(Optional)* Create a [Medialyst](https://medialyst.ai) account to unlock the
   🔧 skills — the curated news index, journalist database, and fit-checked,
   shareable lists.
9. *(Optional)* Back in Claude, click **Connect** on the Medialyst connector and
   authorize it over **OAuth**. Without it, those skills fall back to host web
   search and hand-built lists.

A community marketplace plugin, `newsjack@claude-community`, is pending review
and will be added soon.

The two always-on monitors run here in ⚠️ Limited Mode — a best-effort, one-shot
scan with nothing saved. **Newsjack monitoring** surfaces opportunities but can't
track them over time, and **coverage tracking** can't suppress repeats or flag
only what's *new since last time* without saved seen-state. For saved, scheduled
monitoring, use a local agent.

### ChatGPT

ChatGPT runs Newsjack only in a degraded ⚠️ form. Skills aren't available to most
accounts — *"Skills [are] in beta for ChatGPT Business and Enterprise"* — so on
those plans you can add Newsjack as a Skill, and on every other tier you paste a
skill's instructions into the chat by hand. Either way there's no local CLI or
saved state, so the monitors and persistence-backed skills don't apply. For the
full toolkit, use a local agent or Claude.ai.

---

## What gets installed

```
~/.newsjack/
├── bin/newsjack          # CLI
└── newsjack/             # managed bundle

skills installed to your detected runtime(s):
  newsworthiness-check    score whether your news clears the bar
  headline-generator      headlines and subject lines from raw story facts
  find-journalists        build small, fit-checked media lists
  ...and more
```

The npm package bundles the same CLI and skills, with the `newsjack` command installed on `PATH`.

Curl-installed Newsjack auto-updates from the latest GitHub Release before each run. Set `NEWSJACK_AUTO_UPDATE=0` to disable.

You can also install the CLI byitself using npm:

```bash
npm i -g newsjack@latest
```

---

Author: Elvis Sun — [LinkedIn](https://www.linkedin.com/in/elvissun) · [X](https://x.com/elvissun)
