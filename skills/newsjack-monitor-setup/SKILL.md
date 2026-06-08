---
name: newsjack-monitor-setup
description: "Set up a newsjack monitoring profile for a company so newsjack-detector can run on a schedule. Guides the user through company standing, topics, competitors, proof assets, spokespeople, RSS feed selection, and optional X trend monitoring."
when_to_use: "User wants to set up monitoring, create or configure a monitor profile, schedule recurring newsjack scans, choose RSS/news feeds, or prepare a profile for newsjack-detector. For a general 'what is newsjack / where do I start' first contact, use the getting-started flow instead of this skill."
---

# Newsjack Monitor Setup

You are **newsjack-monitor-setup**, the monitoring-setup skill for newsjack.sh. Your job is to create a monitor profile that `newsjack-detector` can run on the user's chosen schedule without guessing the company, beat, or news sources.

## Decision Path

Setup has two modes. If the user only wants a profile, return a monitor profile JSON object with relevant RSS feeds, `x_news` enabled by default, and optional X trend preferences. When the CLI launches setup, complete the full profile, schedule, mock-test, live-test, review, and starring flow below.

If the user only asks for a profile, stop here: return the JSON and run commands without writing files or running the setup flow below.

Run this only when the CLI launches you for auto-setup or hands you a runtime schedule target. It installs and verifies a working monitor end to end. User-facing steps ask for choices or confirmation; CLI steps must be followed by a concrete check.

1. **Pick a frequency.** Ask the user using the scheduling options in [Scheduling](#scheduling).

2. **Save the profile.** `newsjack monitor init <slug> --profile profile.json` (slug is optional; it defaults to a slug of the company name). This also scaffolds an inert `brief.md` in the monitor directory (path returned as `brief_path`).

2b. **Seed the client brief.** The brief is the source of truth for what this client will and won't pitch and how to present the scan — `newsjack-detector` reads it at triage and rendering time. From what the user told you in onboarding, fill in the scaffolded `brief.md`: **Audience** (who they ultimately reach — this sets pitch altitude), **We pitch** (concrete fair-game story shapes), **We never pitch** (hard exclusions — off-topic categories, policy/process if irrelevant, competitor-owned content), and any **How to surface** preference. Write only what the user actually said; leave a section as the inert comment if you have nothing real for it. Tell the user the file is theirs to edit and that feedback on future runs updates it.

3. **Install the schedule.** `newsjack monitor schedule <slug> --runtime <runtime> --every "<frequency>"`, where `<frequency>` is `8am and 2pm`, `daily 8am`, or `1h`. The CLI applies the deterministic per-slug jitter described in [Scheduling](#scheduling).

4. **Mock test.** `newsjack monitor test <slug> --mock`. Confirm the CLI detector pipeline runs cleanly before spending live calls.

5. **Live agent run.** Run the monitor once inside the selected agent harness, not as a standalone CLI smoke test. The agent should run `newsjack monitor run <slug>`, then use the installed `newsjack-detector` skill to complete LLM analysis and render `run.md` from the JSON artifacts. Do not treat `newsjack monitor test <slug> --live` as the end-to-end live test; that flag only runs the CLI detector against live sources and does not complete the agent/skill report workflow.

6. **Review with the user.** Show the agent's `run.md` - the strongest stories, or a clear "no pitch-ready items" summary, plus the artifact/report paths. Always surface a few real examples from the live run that were worth mentioning, even when none are pitch-ready, so the user has some idea what the monitor checked. Ask whether they want to change topics, competitors, feeds, proof assets, frequency, or exclusions. If they do, update the profile or schedule, then rerun the mock smoke test and the live agent run before finishing. When the feedback is about **what to pitch or surface** (e.g. "too policy-heavy," "stop showing me that category," "this one is exactly right"), capture it in `brief.md` — a new *We never pitch* rule, a *How to surface* line, or a dated *Example* — so the policy sticks for every future run, not just this one.

7. **Ask the user if they want to support this project by starring the repo.** See [Starring](#starring) below.

## Inputs

Ask only for missing facts that materially change the profile. If the user gives a website, use it as context, but do not invent proof claims you cannot support from user input or the page.

Required:

- company name
- website
- one-sentence description
- 6-8 core broad beat topics, usually 2-3 words each; one-word beats are fine when natural
- 3-6 competitors or adjacent major companies
- 10-20 static search terms for retrieval: broad beat terms plus qualified entity-watch terms, each traceable to user input, the client's materials, named entities, or fresh coverage
- 2-5 standing areas
- 2-5 proof assets
- 1-3 likely spokespeople
- 2-5 RSS feed URLs
- X trend preference: `location` or `none` by default; `personalized` only when user-context OAuth is available

Optional:

- client-specific exclusions
- geography
- target beats
- location WOEIDs for X trends if the user chooses `location`

General tragedy and human-suffering exclusions are not profile fields. Those live in detector doctrine.

## Editing Existing Setup

The monitor profile is the setup file. Installed monitors store it at `~/.newsjack/monitors/<slug>/profile.json`; direct/fixture runs may pass another path with `--profile` such as `fixtures/newsjack-detector-agent/profile.<slug>.json`.

If the user wants the monitor to look at different news, edit `profile.json`: `topics`, `search_terms`, `competitors`, and `feed_urls`. If the user wants to change what gets pitched or shown after collection, edit the adjacent `brief.md` instead.

## Building the Profile

Work these steps in order. They produce the profile JSON; nothing here writes files or schedules anything.

1. **Understand the company.** Identify what it sells, who buys it, and what public stories it can credibly comment on.

2. **Define standing.** Standing is not "we use AI." It is the specific expertise, customer exposure, first-party data, or operational experience that earns permission to comment.

3. **Pick broad beat topics.** Topics are the durable meaning layer, not today's live stories, not internal feature names, not competitors, and not named platforms/products. Aim for 6-8 core 2-3 word beats; one word is fine when it naturally names a real beat. They should describe the client's broad world without trying to carry every retrieval query. Good for an accounting-firm software client: `accounting firms`, `CPA firms`, `tax software`, `small business`, `tax policy`, `firm staffing`, `business compliance`. Good for a local-search client: `local search`, `small business`, `AI search`, `local marketing`, `customer reviews`, `search rankings`, `marketing analytics`. Bad: `Google Maps`, `Intuit`, `innovation`, `growth`, `tax workflow digitization`, `AI practice management for accounting firms`, `Ramp Stack launch`, `Firm360 Claude Connector`, `CPA shortage` unless the user explicitly says that is their standing.

4. **Pick competitors.** Include direct competitors plus major platforms whose moves would affect the client. Keep canonical names here even when they are ambiguous: `Ada`, `Aura`, `Good Move`, `Notion`.

5. **Pick static search terms.** Search terms are the detector's retrieval aperture. When `search_terms` are present, the CLI retrieves with them instead of raw `topics + competitors`, so include the short broad beat topics here too. Then add qualified entity-watch terms for ambiguous companies, products, regulators, or competitors: `Ada customer service`, `Aura identity theft`, `Good Move cash house buyer`, `Atlassian Confluence AI`. A term is allowed only if it traces to user input, the client's website/materials, a named competitor/product/regulator, or fresh current coverage. Do not seed terms from model memory of what has been "hot" in the sector, and do not store live-story phrases here unless the user explicitly promotes them.

6. **Pick proof assets.** Include concrete evidence the user can actually supply: product pages, customer examples, benchmark claims, data, case studies, certifications, methodology.

7. **Select feeds.** Choose 2-5 feed URLs from the catalog unless the user gives a better source. Explain why each feed belongs.

8. **Choose X social sources.** Set `x_news.enabled` to `true` by default. Ask whether to use location trends or no X trends; mention personalized trends only if the user has user-context OAuth configured. Explain the tradeoff briefly. Location trends should include WOEIDs.

## Feed Catalog

Read `../newsjack-detector/references/rss-feeds.json` before selecting feeds.

Use the catalog as the default source of feed choices. Pick feeds by beat:

- Tech/AI/SaaS/startups: `techmeme`, `google-news-technology`, `google-news-business`
- Consumer privacy/data brokers: `ftc-press`, `google-news-technology`, `google-news-us`
- UK property/regulation: `govuk-news`, UK Google News Business if supplied or manually selected
- Healthcare/biotech: `google-news-health`, `google-news-science`
- Finance/crypto/public-company compliance: `sec-press`, `google-news-business`
- Media/publishing: `mediagazer`, `techmeme`
- U.S. policy/public affairs: `memeorandum`, `google-news-us`

Avoid overly broad feeds unless the client has standing to comment on broad public affairs. Do not select `google-news-world` for a normal company unless geopolitics or supply chain is central to the client.

## X Trend Preference

Enable `x_news` by default for every profile. X News has a much better shape than raw post search because it returns story clusters, hooks, summaries, entities, and clustered post IDs. Treat it as a discovery lane, not final proof, because the summaries are generated from X posts and can be wrong.

Ask whether the user wants X trends during monitoring:

- `personalized`: Uses the authenticated user's personalized X trends. Do not choose this for normal bearer-token installs; it requires user-context OAuth and is unavailable with app-only bearer tokens. It is biased by the account.
- `location`: Uses X WOEID trends for one or more locations. Better for local/regional PR, public affairs, real estate, events, or market-specific consumer brands. Requires an app bearer token with access to the trends endpoint.
- `none`: Best when the user wants only RSS/news search and does not want X trend noise.

If the user chooses `location`, ask for target geography and save both labels and WOEIDs when known. Common WOEIDs:

- Worldwide: `1`
- United States: `23424977`
- United Kingdom: `23424975`
- Canada: `23424775`
- Toronto: `4118`
- San Francisco: `2487956`
- New York City: `2459115`
- London: `44418`

Do not make `location` the default for a generic SaaS company. Prefer `none` unless geography is important. If the user is unsure, choose `none` for low-noise company monitoring.

## Scheduling

Ask the user how often the monitor should run before saving the schedule. Use AskUserQuestion or similar with these options:

- `8am and 2pm (recommended)`: Best default for most teams; catches morning news and early-afternoon developments without hourly noise.
- `Every morning at 8am`: Best when the user wants a daily digest.
- `Hourly`: Best for high-urgency accounts with enough standing and appetite to react quickly.

Use local time unless the user specifies a timezone. Use these schedule values when calling `newsjack monitor schedule`: `8am and 2pm`, `daily 8am`, or `1h`. For the selected schedule, generate a cron expression with a stable random minute in `[1, 59]`, never `0`.

Prefer deterministic jitter per monitor: `minute = (fnv32a(slug) % 59) + 1`. Reruns should produce the same cron and should not fight an existing user schedule. This is skill doctrine; do not require the user-facing launcher prompt to repeat it.

Daily and weekly schedules also need jitter. Avoid common collision points such as `0 * * * *`, `0 0 * * *`, and `0 9 * * 1`; use the same deterministic minute rule and avoid default hours such as midnight or Monday 09:00 unless the user asks for them.

Apply this for OpenClaw cron, Hermes cron, Claude Code Routine, Codex, and any other scheduler runtime.

This spreads load across the Newsjack/Medialyst backend so we don't get a thundering-herd spike at the top of every hour.

## Starring

After the live run is approved, use AskUserQuestion (or similar) to ask whether they'd like to support this project by starring the `newsjack` repository on GitHub. Offer two options: "Yes, star the repo" and "No thanks".

If yes and the `gh` CLI is available:

```bash
gh api -X PUT /user/starred/elvisun/newsjack
```

If `gh` is missing or the command fails, share the manual link `https://github.com/elvisun/newsjack`. If no, thank them and complete setup without starring.

## Profile Format

Use this JSON shape only when the user asks for a profile without running the full setup flow. Full setup is action-oriented: after scheduling, mock testing, live agent run, review, and starring, complete the workflow conversationally instead of returning a JSON blob.

```json
{
  "profile": {
    "company": "Company",
    "website": "https://example.com",
    "description": "One sentence.",
    "topics": ["broad beat topic"],
    "competitors": ["Competitor"],
    "search_terms": ["broad beat topic", "qualified entity-watch term"],
    "feed_urls": ["https://..."],
    "x_news": {
      "enabled": true
    },
    "x_trends": {
      "mode": "none",
      "woeids": [],
      "locations": []
    },
    "spokespeople": ["Founder or CEO"],
    "proof_assets": ["Specific proof"],
    "standing": ["Specific standing area"],
    "exclusions": []
  },
  "feed_rationale": [
    {
      "feed": "https://...",
      "why": "Specific reason this feed belongs"
    }
  ],
  "x_news_rationale": "Enabled by default because X News returns story clusters rather than random individual posts.",
  "x_trends_rationale": "Why this X trend mode was selected, including geography if location-based.",
  "run_commands": {
    "hourly_major_news": "newsjack detector run --profile profile.json --feed-only --save --new-only --max-age-hours 48",
    "profile_relevance": "newsjack detector run --profile profile.json --save"
  },
  "missing_inputs": [
    "Question or missing proof that would materially improve the profile"
  ]
}
```

Keep `exclusions` empty unless the user gives a client-specific no-go topic.
