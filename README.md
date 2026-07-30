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

Four problems, separate lanes.

Each capability is a skill. Invoke it by name — `/angle-generator`, `/fact-check` —
or just describe what you want and your agent picks the right one.

### 🛰️ Detect — surface what matters in your space

- ⚙️ **`/newsjack-monitor-setup`** — build the monitoring profile once: your standing, beat topics, competitors, proof assets, spokespeople, feeds *(local agent only)*
- 📡 **`/newsjack-detector`** — find newsjacking opportunities: fresh stories you have the standing to jump on before the wave breaks ([see a sample run](docs/example-run.md))
- 🔎 **`/news-search`** — dated, attributed articles on a topic, company, or competitor: when they launch, raise, or stumble, you know
- 🔍 **`/story-origin-check`** — verify the story is still fresh: who broke it, who owns it, what oxygen's left
- 🧹 **`/relevance-coarse-filter`** — cheap high-recall first pass that throws out obvious junk before anything expensive runs
- 🚦 **`/newsjack-triage`** — route surviving stories by whether you actually have standing: pitch now, watch, or surface as big news
- 🔔 **`/coverage-tracker-setup`** — declare the keywords you want watched, and what each one actually means *(local agent only)*
- 🗞️ **`/coverage-tracker`** — Google Alerts-style keyword tracking with LLM filtering for real features, alerting only on genuinely new coverage

### 🚀 Act — turn signal into output

- 🎯 **`/angle-generator`** — turn one update into hooks framed for different beats
- 📝 **`/headline-generator`** — headlines and pitch subject lines built from the story's raw facts
- 🥊 **`/meanest-editor`** — roast your pitch against the rubric editors actually use
- 🚨 **`/crisis-holding`** — holding statements, journalist Q&A posture, and what *not* to say, with a hard legal-counsel gate
- 🎙️ **`/reactive-comment`** — triage inbound HARO-style source requests, draft only the real fits
- ✅ **`/fact-check`** — extract claims, verify each, flag the shaky ones before you send
- 🤝 **`/journalist-fit-check`** — will *this* reporter actually care, or are you spamming?
- 🥇 **`/same-outlet-ranker`** — rank colleagues at the same publication and send to one, instead of spraying the masthead
- 🗣️ **`/voice-extractor`** — fingerprint your real writing, kill the AI tells
- 📋 **`/find-journalists`** — build a fit-checked media list: targeted reporters, not scraped contact dumps
- 📰 **`/press-clip`** — turn a live article URL into a branded press-clip PDF: the outlet's own logo and layout kept, ads and clutter stripped, the client's mention highlighted *(local agent only)*

### 🧭 Strategize — figure out what your story even is

- 🗺️ **`/pr-strategist`** — opinionated walkthrough if you're not PR-fluent yet: audience first, positioning second, news pegs third, drumbeat over big-bang
- 📅 **`/pr-calendar`** — plan six months of source-backed upcoming hooks, with prior-year coverage patterns for angle inspiration
- 📊 **`/newsworthiness-check`** — cold read on whether it clears the bar before you act

### 🔬 AI visibility — get your facts into AI answers (AEO/GEO)

- ✍️ **`/ai-visibility-writing`** — audit or fact-preservingly revise supplied copy so AI answers can reuse it, without inventing evidence or promising rankings or citations
- 🧭 **`/build-ai-visibility-panel`** — to measure it, run this one first: it drives the six below it in order, turning any public URL and description into an evidence-bound prompt list across intent, journey, B0–B5 proximity, aided status, roles, locales, surfaces, and tracking partitions
- 👥 **`/icp-evidence-analysis`** — turn a company and market dossier into testable ideal-customer hypotheses, buying roles, triggers, disqualifiers, and named research gaps
- 🧩 **`/buyer-job-intent-analysis`** — recover the jobs, struggling moments, workarounds, and authentic buyer language from real customer, review, forum, and search evidence
- 🏗️ **`/prompt-proximity-architecture`** — blueprint which prompt cells are required, optional, and prohibited before anyone writes prompt wording
- 💬 **`/realistic-prompt-generation`** — write natural prompt variants from a target-blind brief, so the panel isn't quietly written to flatter you
- 🧪 **`/prompt-set-qa`** — gate the prompt set on provenance, contamination, answer leakage, naturalness, and semantic duplicates, quarantining anything the evidence doesn't support
- 📐 **`/ai-visibility-panel-design`** — select QA-approved prompts into a versioned tracking panel with partitions, surfaces, locales, repetitions, weights, and refresh rules, keeping aided and unaided denominators separate and gating on a human instead of inventing an “AI visibility score”

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

| Skill | [Claude.ai](https://claude.ai) | [ChatGPT](https://chatgpt.com) | [ChatGPT Work](https://openai.com/business) | [Cowork](https://claude.com/product/cowork) | [Claude Code](https://claude.com/claude-code) | [Codex](https://openai.com/codex) | [Hermes](https://hermes-agent.nousresearch.com) | [OpenClaw](https://openclaw.ai) | [Medialyst](https://medialyst.ai) |
| --- | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: | :---: |
| **Strategize** | | | | | | | | | |
| pr-strategist | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| pr-calendar | 🔧 | ⚠️ | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | ✅ |
| newsworthiness-check | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **AI visibility** | | | | | | | | | |
| ai-visibility-writing | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| build-ai-visibility-panel | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| icp-evidence-analysis | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| buyer-job-intent-analysis | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| prompt-proximity-architecture | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| realistic-prompt-generation | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| prompt-set-qa | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| ai-visibility-panel-design | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Act** | | | | | | | | | |
| angle-generator | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| headline-generator | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| meanest-editor | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| crisis-holding | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| reactive-comment | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| fact-check | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| journalist-fit-check | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| same-outlet-ranker | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| voice-extractor | ⚠️ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| find-journalists | 🔧 | ⚠️ | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | ✅ |
| press-clip | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | 🔜 |
| **Detect** | | | | | | | | | |
| news-search | ✅ | ⚠️ | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | ✅ |
| story-origin-check | 🔧 | ⚠️ | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | ✅ |
| relevance-coarse-filter | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| newsjack-triage | ✅ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| newsjack-detector | ⚠️ | ⚠️ | ⚠️ | ⚠️ | 🔧 | 🔧 | 🔧 | 🔧 | 🔜 |
| newsjack-monitor-setup | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | 🔜 |
| coverage-tracker | ⚠️ | ⚠️ | ⚠️ | ⚠️ | ✅ | ✅ | ✅ | ✅ | 🔜 |
| coverage-tracker-setup | ❌ | ❌ | ❌ | ❌ | ✅ | ✅ | ✅ | ✅ | 🔜 |


Legend:

- **✅ Runs out of the box** — no setup; works anywhere your agent does.
- **🔧 May need an external connection** — requires or does its best work with an external data source connected (e.g. the X API or the Medialyst API).
- **⚠️ Limited Mode** — runs in a chat app, but as a best-effort, one-shot pass with nothing saved between sessions: no stored voice fingerprint, no scheduled monitoring, no repeat-suppression. Connect a local agent for the saved, scheduled version.
- **🔜 Coming soon** — not available here yet.
- **❌ Not supported here** — needs a local agent. The two setup skills (`newsjack-monitor-setup`, `coverage-tracker-setup`) only save a profile or config and schedule it; `press-clip` drives a real Chrome/Edge browser via Playwright to render and clip the live page, which chat apps can't do.

`pr-calendar` is setup-required but portable: it works anywhere the agent can use the `newsjack` CLI or a connected Medialyst MCP/connector. The calendar endpoint is free; login is required only to prevent abuse. Consumer ChatGPT is the exception for now because of the same skill-loading limitation called out above.

**Set up your agent:**

- **[Local agents](#local-agents-claude-code-codex-hermes-openclaw)** — Claude Code, Codex, Hermes, OpenClaw
- **[Claude.ai & Cowork](#claudeai--cowork)** — Anthropic plugin + Medialyst connector
- **[ChatGPT](#chatgpt)** — consumer accounts, no Skills
- **[ChatGPT Work](#chatgpt-work)** — Skills beta

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

### Claude.ai & Cowork

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

ChatGPT runs Newsjack only in a degraded ⚠️ form, because Skills aren't available
on consumer accounts. For the full toolkit, use ChatGPT Work or OpenAI's Codex.

### ChatGPT Work

ChatGPT Work has the Skills beta, so Newsjack runs there at the same level as
Cowork: everything except the two setup skills and `press-clip`, which need a
local agent.

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

You can also install the CLI by itself using npm:

```bash
npm i -g newsjack@latest
```

---

## Contributors

- **Elvis Sun** — [X](https://x.com/elvissun) · [LinkedIn](https://www.linkedin.com/in/elvissun)
- **Carly Martinetti** — [X](https://x.com/prcarly) · [LinkedIn](https://www.linkedin.com/in/prcarly)
