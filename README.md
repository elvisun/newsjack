# Newsjack.sh

**The open-source skills that turn your agent into a full PR team.**

Install once. Your agent — Claude Code, Codex, Hermes, OpenClaw — becomes a PR team.

```bash
curl -fsSL newsjack.sh | bash
```

**Are you an agent?** Check out **[Getting started](docs/getting-started.md)**

**Are you a human?** 📋 Copy this prompt to any AI:

```text
help me install https://newsjack.sh from the github repo and setup a daily newsjack monitoring for my company.
```

Jump to **[platform-specific setup](#install)** below for per platform breakdown.

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
- 🥇 **Pick who to pitch first at one outlet** — rank colleagues at the same publication and send to one, instead of spraying the masthead
- 🎙️ **Respond to source queries** — triage inbound HARO-style requests, draft only the real fits
- 🥊 **Roast your pitch** — honest critique against the rubric editors actually use
- ✅ **Fact-check before you send** — extract claims, verify each, flag the shaky ones
- ✍️ **Make writing easier for AI answers to reuse** — audit or fact-preservingly revise supplied copy without inventing evidence or promising rankings or citations
- 🗣️ **Keep drafts in your voice** — fingerprint your real writing, kill the AI tells
- 📋 **Build a fit-checked media list** — targeted reporters, not scraped contact dumps
- 📰 **Clip the coverage** — turn a live article URL into a branded press-clip PDF: the outlet's own logo and layout kept, ads and clutter stripped, the client's mention highlighted *(local agent only)*

### 🧭 Strategize — figure out what your story even is

- 🗺️ **Get a PR strategy** — opinionated walkthrough if you're not PR-fluent yet: audience first, positioning second, news pegs third, drumbeat over big-bang
- 📅 **Build a PR calendar** — plan six months of source-backed upcoming hooks, with prior-year coverage patterns for angle inspiration
- 📊 **Score newsworthiness** — cold read on whether it clears the bar before you act

### 🔬 Research — build an AI-visibility measurement panel

- 🧭 **Map the buyer prompt space** — start with any public URL and description, research buyer jobs and language, and produce an evidence-bound prompt list across intent, journey, B0–B5 proximity, aided status, roles, locales, surfaces, and tracking partitions
- 🧪 **Keep the measurement honest** — separate aided and unaided denominators, quarantine unsupported prompts, preserve source provenance, and return a provisional panel with explicit human gates instead of a made-up “AI visibility score”

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
A few reach for a live news index, journalist enrichment,
or locally-saved monitoring state — those work best in a local agent.

### What runs where

| Skill | [Claude.ai](https://claude.ai) | [ChatGPT](https://chatgpt.com) | [Cowork](https://claude.com/product/cowork) | [Claude Code](https://claude.com/claude-code) | [Codex](https://openai.com/codex) | [Hermes](https://hermes-agent.nousresearch.com) | [OpenClaw](https://openclaw.ai) | [Medialyst](https://medialyst.ai) |
| --- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| **Strategize** | | | | | | | | |
| pr-strategist | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| pr-calendar | 🔧 | ⚠️ | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | ✅ |
| newsworthiness-check | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Research / Measure** | | | | | | | | |
| build-ai-visibility-panel | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| icp-evidence-analysis | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| buyer-job-intent-analysis | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| prompt-proximity-architecture | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| realistic-prompt-generation | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| prompt-set-qa | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| ai-visibility-panel-design | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Act** | | | | | | | | |
| angle-generator | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| headline-generator | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| meanest-editor | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| crisis-holding | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| reactive-comment | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| fact-check | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| ai-visibility-writing | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| journalist-fit-check | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| same-outlet-ranker | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| voice-extractor | ⚠️ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| find-journalists | 🔧 | ⚠️ | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | ✅ |
| press-clip | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | 🔜 |
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
- **🔧 May need an external connection** — requires or does its best work with an external data source connected (e.g. the X API or the Medialyst API).
- **⚠️ Limited Mode** — runs in a chat app, but as a best-effort, one-shot pass with nothing saved between sessions: no stored voice fingerprint, no scheduled monitoring, no repeat-suppression. Connect a local agent for the saved, scheduled version.
- **🔜 Coming soon** — not available here yet.
- **❌ Not supported here** — needs a local agent. The two setup skills (`newsjack-monitor-setup`, `coverage-tracker-setup`) only save a profile or config and schedule it; `press-clip` drives a real Chrome/Edge browser via Playwright to render and clip the live page, which chat apps can't do.

`pr-calendar` is setup-required but portable: it works anywhere the agent can use the `newsjack` CLI or a connected Medialyst MCP/connector. The calendar endpoint is free; login is required only to prevent abuse. ChatGPT is the exception for now because of the same skill-loading limitation called out above.

**Set up your agent:**

- **[Local agents](#local-agents-claude-code-codex-hermes-openclaw)** — Claude Code, Codex, Hermes, OpenClaw
- **[Claude.ai & Cowork](#claudeai--cowork)** — Anthropic plugin + Medialyst connector
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

Prefer a video? Watch the [Newsjack installation walkthrough](https://www.youtube.com/watch?v=1tg6E6ZYGCk).

### Claude.ai & Cowork - Install as a Claude Plugin

**Install the Newsjack plugin:**

1. Open **[Customize](https://claude.ai/customize)**.
2. Go to **Personal plugins**.
3. Click **Create plugin**.
4. Choose **Add marketplace**.
5. Choose **Add from a repository**.
6. In the **repository URL** field, enter `elvisun/newsjack` and confirm.
7. Open the new **`elvisun/newsjack`** marketplace, find the **`newsjack@newsjack`**
   plugin, and click **Install**.
8. *(Optional)* Unlock the 🔧 skills by connecting a Medialyst account so your agent can access live news search, PR calendar lookup, and journalist enrichment. In local agent harnesses, run `newsjack login` and approve the printed Medialyst link. If you don't have an account you can create a free one on the [Medialyst signup page](https://medialyst.ai/agents).
9. *(Optional)* Back in Claude, find the Medialyst connector under [the installed Newsjack plugin page](https://claude.ai/customize/plugins/newsjack%40newsjack/connectors), click **Connect** on the Medialyst connector and authorize it over **OAuth**.

Note: A community marketplace plugin, `newsjack@claude-community`, is pending review
and will be added soon.

The two always-on monitors run here in ⚠️ Limited Mode — a best-effort, one-shot
scan with nothing saved. For saved, scheduled monitoring, use a local agent.

### ChatGPT

ChatGPT runs Newsjack only in a degraded ⚠️ form. Skills aren't available to
accounts that are not on ChatGPT Business or Enterprise. For the full toolkit, use OpenAI's Codex.

---

## What gets installed

```
~/.newsjack/
├── bin/newsjack          # CLI
└── newsjack/             # managed bundle

skills installed to your detected runtime(s):
  newsworthiness-check    score whether your news clears the bar
  headline-generator      headlines and subject lines from raw story facts
  find-journalists        build small, fit-checked journalist lists
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
