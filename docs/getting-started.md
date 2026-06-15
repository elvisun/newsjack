# Getting started with newsjack

This is the front door. If you are an agent that just had newsjack installed (or
had this repo handed to you), read this first and follow it before doing anything
else.

## Runtime modes

Newsjack is optimized for local-first agent harnesses. Full Mode needs shell
execution, filesystem storage, credentials, durable client profiles, JSON
artifacts, scheduled workflows, multi-agent orchestration, and cost-optimized
worker passes.

Use **Full Mode** in:

- Claude Code
- Codex
- OpenClaw
- Hermes

Use **Limited Mode** in browser or restricted chat environments:

- Claude.ai chat
- ChatGPT chat
- Claude Cowork

Limited Mode can do strategy, newsworthiness checks,
angles, pitch critique, journalist-fit reasoning, fact-checking from pasted or
searchable evidence, and best-effort manual news scans. It cannot run the
canonical detector, save monitors, schedule runs, keep seen-state, write local
artifacts, or use cost-optimized multi-agent passes.

## Install paths

Default Full Mode install:

```bash
curl -fsSL newsjack.sh | bash
```

Use npm only in a Full Mode harness when shell installers or GitHub Release
assets are blocked:

```bash
npm i -g newsjack
newsjack install
```

On Windows there is no `curl | bash`. When the user is on Windows, download and
run the setup binary instead (requires v0.1.10 or later):

```powershell
[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12; iwr https://github.com/elvisun/newsjack/releases/latest/download/newsjack_windows_amd64.exe -OutFile newsjack.exe; Unblock-File newsjack.exe; .\newsjack.exe setup
```

`newsjack setup` fetches the skills bundle, verifies its checksum, installs the
CLI to `%USERPROFILE%\.newsjack\bin`, and runs the same guided setup as
macOS/Linux — including installing Claude Code via its native Windows installer
if it is missing.

Skill instructions assume the command is available as `newsjack` only in Full
Mode. In Limited Mode, do not try to install the CLI inside chat; use the
Limited Mode workflow and label detector-style output as reduced coverage.

## For the agent: start slow

On first contact, **do not** dump the full skill list, and **do not**
announce missing API keys. Newsjack works without any keys; credentials are
optional upgrades covered below.

Instead, find out what the user is trying to do and take **one** concrete step.
Open with a short orientation and offer a few real starting points — not the whole
menu:

> Newsjack turns me into your PR operator. I can run Full Mode in Claude Code,
> Codex, OpenClaw, or Hermes, or Limited Mode in browser chat. Where do you want
> to start?
>
> 1. **See what newsjack can do** — a quick tour of the workflow
> 2. **Get a PR strategy** — figure out audience, positioning, and news pegs first
> 3. **Check if something's newsworthy** — score a news event or your own pitch idea before you act
> 4. **Set up monitoring** — watch your industry and competitors for stories to jump on
> 5. **Track coverage** — Google Alerts-style keyword alerts filtered for real features
> 6. **Find journalists** — build a small, fit-checked media list for a pitch

Then let the user pick and go one step at a time. Each starting point maps to a
skill:

| Starting point | Skill |
| --- | --- |
| See what newsjack can do | (brief tour — only expand the full skill list if asked) |
| Get a PR strategy | `pr-strategist` |
| Check if something's newsworthy | `newsworthiness-check` |
| Set up monitoring | `newsjack-monitor-setup` |
| Track coverage | `coverage-tracker-setup` |
| Find journalists | `find-journalists` |

If the user already knows what they want ("draft a pitch", "is this newsworthy?",
"roast this"), skip the menu and route straight to the relevant skill.

## Dependencies — what they unlock, and what they cost

**You can do real work with none of these.** Newsjack's base workflow — strategy,
angles, fit-checks, drafts, voice, newsworthiness, and a local journalist-list artifact
— needs no signup and no keys. The optional integrations below add reach; treat a
missing one as reduced coverage, not a blocker, and never lead with a missing-key
complaint.

| Dependency | Unlocks | Without it | Cost |
| --- | --- | --- | --- |
| **Medialyst key** | live news search with publication metadata and journalist enrichment from article URLs | news search falls back to host web/browser search (best-effort freshness); journalist lists stay as local agent artifacts with any unresolved rows marked honestly | 300 free credits on signup (~3,000 news searches), paid after — [medialyst.ai/agents#pricing](https://medialyst.ai/agents#pricing) |
| **X bearer token** | the X/Twitter trend source inside monitoring | that source is simply omitted; RSS and news still run | pay-as-you-go, no free tier — [X API pricing](https://docs.x.com/x-api) |

### Why Medialyst for news search

General web search is bad at news: it ranks for SEO over recency, paywalls or
buries primary coverage, and rarely exposes a reliable publication timestamp.
Medialyst is purpose-built for news and returns the outlet, author, `published_at`,
and canonical URL that downstream skills (`coverage-tracker`,
`story-origin-check`, `newsworthiness-check`, `find-journalists`,
`newsjack-detector`) depend on. The
`news-search` skill prefers it and falls back to host search — flagging reduced
freshness confidence — when it is not configured. It is optional cloud substrate,
not a signup wall.

## Setting up credentials (only when the user wants the upgrade)

- **Medialyst:** prefer `newsjack setup` or `newsjack login`. Newsjack stores the
  key in `~/.newsjack/credentials.json`, or reads `MEDIALYST_API_KEY` from the
  environment. The CLI calls the Medialyst public REST API directly for news
  search and journalist enrichment. Agents own how they organize the returned
  journalist data.
  Recommended scopes: `news:search` and journalist enrichment access.
- **X:** set `X_BEARER_TOKEN` (alias `TWITTER_BEARER_TOKEN`). Newsjack calls the X
  API directly.

Only bring these up when the user reaches a step that benefits from them.
