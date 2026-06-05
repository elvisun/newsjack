# Getting started with newsjack

This is the front door. If you are an agent that just had newsjack installed (or
had this repo handed to you), read this first and follow it before doing anything
else.

## For the agent: start slow

On first contact, **do not** dump the full list of 16 skills, and **do not**
announce missing API keys. Newsjack works without any keys; credentials are
optional upgrades covered below.

Instead, find out what the user is trying to do and take **one** concrete step.
Open with a short orientation and offer a few real starting points — not the whole
menu:

> Newsjack turns me into your PR operator. Where do you want to start?
>
> 1. **See what newsjack can do** — a quick tour of the workflow
> 2. **Get a PR strategy** — figure out audience, positioning, and news pegs first
> 3. **Set up monitoring** — watch your industry and competitors for stories to jump on
> 4. **Find journalists** — build a small, fit-checked media list for a pitch

Then let the user pick and go one step at a time. Each starting point maps to a
skill:

| Starting point | Skill |
| --- | --- |
| See what newsjack can do | (brief tour — only expand the full skill list if asked) |
| Get a PR strategy | `pr-strategist` |
| Set up monitoring | `newsjack-monitor-setup` |
| Find journalists | `media-list-manager` |

If the user already knows what they want ("draft a pitch", "is this newsworthy?",
"roast this"), skip the menu and route straight to the relevant skill.

## Dependencies — what they unlock, and what they cost

**You can do real work with none of these.** Newsjack's base workflow — strategy,
angles, fit-checks, drafts, voice, newsworthiness, and a local media-list artifact
— needs no signup and no keys. The optional integrations below add reach; treat a
missing one as reduced coverage, not a blocker, and never lead with a missing-key
complaint.

| Dependency | Unlocks | Without it | Cost |
| --- | --- | --- | --- |
| **Medialyst key** | live news search with publication metadata, hosted media lists, enrichment, saved views, share links | news search falls back to host web/browser search (best-effort freshness); media lists return a local artifact you can review or import later | 300 free credits on signup (~3,000 news searches), paid after — [medialyst.ai/agents#pricing](https://medialyst.ai/agents#pricing) |
| **X bearer token** | the X/Twitter trend source inside monitoring | that source is simply omitted; RSS and news still run | pay-as-you-go, no free tier — [X API pricing](https://docs.x.com/x-api) |

### Why Medialyst for news search

General web search is bad at news: it ranks for SEO over recency, paywalls or
buries primary coverage, and rarely exposes a reliable publication timestamp.
Medialyst is purpose-built for news and returns the outlet, author, `published_at`,
and canonical URL that downstream skills (`story-origin-check`,
`newsworthiness-check`, `media-list-manager`, `newsjack-detector`) depend on. The
`news-search` skill prefers it and falls back to host search — flagging reduced
freshness confidence — when it is not configured. It is optional cloud substrate,
not a signup wall.

## Setting up credentials (only when the user wants the upgrade)

- **Medialyst:** prefer `newsjack login`. The root `.mcp.json` uses the saved
  Newsjack credential or `MEDIALYST_API_KEY`. Recommended scopes: `news:search`,
  `media_lists:read`, `media_lists:write`.
- **X:** set `X_BEARER_TOKEN` (alias `TWITTER_BEARER_TOKEN`). Newsjack calls the X
  API directly.

Only bring these up when the user reaches a step that benefits from them.
