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
| pr-strategist | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| newsworthiness-check | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| **Act** | | | | | | | | |
| angle-generator | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| headline-generator | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| meanest-editor | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| crisis-holding | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| reactive-comment | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| fact-check | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| journalist-fit-check | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| voice-extractor | ⚠️ | ⚠️ | ⚠️ | ✅ | ✅ | ✅ | ✅ | ✅ |
| find-journalists | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | ✅ |
| **Detect** | | | | | | | | |
| news-search | ✅ | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | ✅ |
| story-origin-check | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | 🔧 | ✅ |
| relevance-coarse-filter | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| newsjack-triage | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
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

### With a local agent (Claude Code, Codex, OpenClaw, Hermes)

Use the curl installer when your agent can reach GitHub Release assets:

```bash
curl -fsSL newsjack.sh | bash
```

Or review the script before running it:

```bash
curl -fsSL https://newsjack.sh
```

No surprises: that's the exact, unminified [`install.sh`](install.sh) in this
repo — `newsjack.sh` serves this file and every release bundles it unchanged, so
what you read here is what runs. Read it before you pipe it to a shell.

The routing is open too: [`apps/site/proxy.ts`](apps/site/proxy.ts) is the
Next.js handler behind `newsjack.sh` — it rewrites installer user-agents
(curl/wget) to `install.sh` and 308-redirects everyone else to this repo.

The installer detects your agent runtime and installs the CLI-backed skills
automatically.

If shell installers or GitHub Release assets are blocked, use npm instead:

```bash
npm i -g newsjack
newsjack install
```

After install, agents and skills should call the CLI as `newsjack`.

### Windows

No script, no git, no Node. One PowerShell line downloads the CLI and runs
setup (requires v0.1.10 or later):

```powershell
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; iwr https://github.com/elvisun/newsjack/releases/latest/download/newsjack_windows_amd64.exe -OutFile newsjack.exe; Unblock-File newsjack.exe; .\newsjack.exe setup
```

The TLS line makes the download work on stock Windows PowerShell 5.1 (it's a
no-op on PowerShell 7).

`newsjack setup` downloads the skills bundle, verifies its checksum, installs
the CLI to `%USERPROFILE%\.newsjack\bin`, and walks you through the same
guided setup as macOS/Linux — including installing Claude Code through its
own native Windows installer if you don't have it yet. Updates are automatic,
same as the curl install.

### In a chat app (Claude.ai, ChatGPT, Cowork)

No install, no CLI — paste your startup URL or your news and you get a real PR
operator on the spot.

**Nearly every skill runs right in the chat:** PR strategy, newsworthiness
scoring, story angles, pitch drafting and honest critique, journalist-fit
reasoning, media lists, reactive source-query responses, voice fingerprinting,
and fact-checking against pasted or searchable evidence. Connect the
[Medialyst](https://medialyst.ai) MCP for the live news index and fit-checked
media lists (the 🔧 skills above).

Both always-on monitors run in chat as ⚠️ Limited Mode — a best-effort, one-shot
scan you trigger by hand. **Newsjack monitoring** surfaces opportunities but
can't track them over time, and **coverage tracking** does a one-off check but
can't suppress repeats or flag only what's *new since last time* without saved
seen-state. Want them saved, scheduled, and running on autopilot? Use a local
agent (above), or [Medialyst](https://medialyst.ai) (coming soon).

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
  headline-generator      headlines and subject lines from raw story facts
  find-journalists        build small, fit-checked media lists
  …and more
```

The npm package bundles the same CLI and skills, with the `newsjack` command installed on `PATH`.

Curl-installed Newsjack auto-updates from the latest GitHub Release before each run. Set `NEWSJACK_AUTO_UPDATE=0` to disable.

Npm-installed Newsjack uses npm for CLI updates:

```bash
npm i -g newsjack@latest
```

---

Author: Elvis Sun — [LinkedIn](https://www.linkedin.com/in/elvissun) · [X](https://x.com/elvissun)
