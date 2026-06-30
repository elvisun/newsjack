---
name: pr-calendar
description: "Turn the Medialyst PR calendar feed into a brand-specific, lead-time-aware plan. Pull source-backed upcoming moments, keep only the few a brand has real standing to own, schedule backward from the event date, screen out tone-deaf hooks, and hand selected moments to angle and journalist prep ahead of time. The planned twin of newsjack-detector's reactive monitoring."
when_to_use: "User wants a PR calendar, editorial calendar, PR plan, or upcoming moments to pitch around; asks what to plan for next quarter or the next six months; asks what awareness days, holidays, seasonal hooks, industry events, conferences, anniversaries, or cultural moments are coming up; or wants to pre-build angles and media lists for a known future date. Use newsjack-detector instead when the trigger is news breaking right now."
---

# PR Calendar

A PR calendar is the planned half of PR. `newsjack-detector` watches news breaking right now; this skill looks at known upcoming moments and decides which ones this brand should plan around, when pitching should start, and what proof the brand needs before anyone contacts journalists.

The calendar feed is source-backed, but the value is still judgment. Do not dump a list of dates. Most moments should disappear because the brand has no standing, the hook is too generic, or the moment is unsafe to commercialize.

This is a MOLECULE. It reads the brand's monitor profile, queries the calendar API through the CLI, cuts the feed to real candidates, checks what kinds of stories landed around last year's version of those moments, and returns a readable plan. It does not write angles or pitches itself, and it persists nothing.

## What this skill will not do

- It will not promise a magical best day or best time to send a pitch. Use windows, not fake precision.
- It will not list every observance. A weak or generic hook is not a media plan.
- It will not invent dates. Every moment in **Act this week**, **Six-month plan**, **Watch list**, and **Avoid** must come from the calendar feed or from a date the user explicitly supplied. If an obvious anchor is missing from the feed, mention it only under **Gaps & clusters** as "missing from the feed; verify separately before planning."
- It will not run the detector pipeline, coarse relevance filter, story-origin check, or freshness gate. This is a smaller context problem. One careful LLM pass over the calendar results is enough.

## Inputs

1. **The brand profile.** Reuse the existing monitor profile from `newsjack-setup` or fixture profile JSON. Read company description, topics, search terms, competitors, standing, spokespeople, proof assets, and market. Do not run a new questionnaire unless no profile exists.
2. **The planning window.** Default to the next six months from today. Long-lead opportunities can already be opening, so do not use a shorter default.
3. **The calendar feed.** Dated opportunities come from `newsjack pr-calendar query`, which wraps Medialyst's PR calendar API.

If no profile exists, ask only for the four things needed for judgment: what the brand does, what topics it can credibly speak to, what proof or data it can attach, and which market it pitches into.

## Get The Calendar

Newsjack reaches the PR calendar two ways. Try them in this order and stop at the first one that works:

1. **CLI mode (preferred).** Use `newsjack pr-calendar query` when the CLI is installed and authenticated.
2. **MCP mode (fallback).** If the `newsjack` CLI is not installed or not on PATH but the `medialyst` MCP server is connected, use the MCP PR calendar query tool. It mirrors `POST /api/v1/pr-calendar/query`, so the request fields and response shape are the same as the CLI — only the transport differs.

The PR calendar endpoint is free to use. It requires Medialyst login only to prevent abuse. If the live path is unauthenticated, connect Medialyst and retry; do not build a fallback calendar from memory.

### CLI mode

Use the repo shim in a source checkout and the installed binary otherwise:

```bash
bin/newsjack --version
newsjack --version
```

If the command is unauthenticated and you have shell access, run:

```bash
newsjack login
```

Tell the user to open the printed Medialyst link and approve `newsjack CLI`, then retry the calendar query.

Then query one broad six-month window per profile:

```bash
newsjack pr-calendar query --from <YYYY-MM-DD> --to <YYYY-MM-DD> \
  --industry <tech|fin|health|retail|media|food> \
  --limit 100
```

Use `--query` only when it helps without over-narrowing. Keep it short and broad, usually one or two words. For a developer-tools or AI-software profile, `--industry tech` is often better than packing the whole topic list into `--query`. Over-specific queries can hide useful conferences and seasonal hooks.

Do not use `--country-code` by default. Add it only when the user explicitly wants a local calendar, because many global events do not carry a country code.

For exact API-shaped requests, use:

```bash
newsjack pr-calendar query --json '{"industries":["tech"],"start_date":"2026-06-30","end_date":"2026-12-30","limit":100}'
```

The response returns `calendar` metadata plus `opportunities`. Important opportunity fields include `title`, `type`, `industry`, `start_date`, `end_date`, `pitch_start_at`, `pitch_deadline_at`, `lead_time_summary`, `description`, `pitch_angles`, `audiences`, `coverage`, `source_urls`, `tags`, and `confidence`.

### MCP mode

Use this path only when the `newsjack` CLI is unavailable and the `medialyst` MCP server is connected. The MCP tool takes the API request body directly. In most runtimes the tool is named `mcp__medialyst__query_pr_calendar` or similar; if the runtime exposes a different PR calendar tool name, use the one whose description maps to `POST /api/v1/pr-calendar/query`.

MCP request body:

```json
{
  "industries": ["tech"],
  "start_date": "2026-06-30",
  "end_date": "2026-12-30",
  "limit": 100
}
```

Available request fields mirror the CLI/API: `q`, `industries`, `types`, `tags`, `audiences`, `country_codes`, `start_date`, `end_date`, `pitch_start_before`, `pitch_deadline_after`, `limit`, and `cursor`.

Use MCP results exactly like CLI results: `calendar` metadata plus `opportunities`. If neither CLI nor MCP is available, stop and tell the user to connect Medialyst; do not synthesize a calendar.

## Candidate Review

Review the returned opportunities once. Assign each opportunity to one of four buckets:

| Bucket | Meaning | Output treatment |
|---|---|---|
| **Pitch-ready** | The brand has strong standing and can attach proof, data, a product fact, a customer, or a credible spokesperson. | Put in the plan. |
| **Watch** | Plausible but generic. It needs proprietary proof before pitching. | Include only if it helps the user plan, and say what proof would upgrade it. |
| **Avoid** | Sensitive, solemn, political, medical, memorial, or otherwise likely to read as opportunistic for this brand. | Put in Avoid with the reason. |
| **Drop** | No legitimate media angle for this brand. | Do not list individually unless the user asks for rejects. |

Use these gates inside the one pass:

**Brand standing.** Ask: would a journalist naturally accept this brand as a source on this moment? A developer-tools company at a major developer or enterprise software conference can be strong. A fintech on World Savings Day is usually watch unless it has proprietary savings data. A B2B logistics SaaS on National Pizza Day is drop.

**Proof.** A pitch-ready moment needs something concrete: usage data, a benchmark, customer access, founder expertise, technical research, product news, or a clear spokesperson. If the profile has no proof, mark watch, not pitch-ready.

**Brand safety.** Avoid commercializing memorials, disasters, serious illness days, religious solemn observances, and charged political moments unless the brand has genuine standing and a respectful public-interest angle. Follow `skills/ETHICS.md` and `skills/WHY-NOT-SPAM.md`.

**Timing.** Use the API's `pitch_start_at`, `pitch_deadline_at`, and `lead_time_summary` first. Compare those dates with today:

- **Pitch now** means today is between the pitch start and deadline.
- **Window opens `<date>`** means the pitch start is still ahead.
- **Deadline passed** means redirect to shorter-lead outlets or skip.
- **Event imminent** means only reactive commentary, owned content, or very short-lead desks remain.

When useful, explain outlet timing as windows, not exact laws:

| Outlet tier | Usual pitch window |
|---|---|
| Long-lead magazines and trades | 3-6 months |
| National online publications | 1-2 months |
| Weekly magazines | 2 weeks-2 months |
| Daily print | 1-2 weeks |
| Online, blogs, podcasts, regional | 1-4 weeks |
| News desks | Up to the day |

**Drumbeat.** After bucketing, scan the six-month shape once. Note quiet stretches that need evergreen stories and crowded weeks that should be staggered.

## Prior-Year Coverage

After you have candidates, use `news-search` to research what actually earned coverage around last year's version of each pitch-ready and useful watch-list moment. This is inspiration and calibration, not a freshness gate.

Keep this bounded:

- Research only candidates that survived the standing and safety review.
- For a normal six-month calendar, search the top 8-12 candidates at most. If there are more, prioritize pitch-ready moments, open/closing pitch windows, and moments with the clearest brand standing.
- Use 1-2 focused searches per candidate. Do not exhaustively research every possible article.

For each candidate, derive a prior-year search window:

1. Start with the candidate's `start_date`.
2. Move it back one year to approximate last year's comparable moment. If the feed or prior coverage clearly gives the exact previous occurrence date, use that instead.
3. Strip the year from event names when searching recurring events, then search from six months before that prior-year date through six months after it.

Use `news-search` / `newsjack news search` with `after:` and `before:` terms in the query. If the CLI is missing but Medialyst MCP is connected, call `mcp__medialyst__search_news` with the same query string.

```bash
newsjack news search --query "\"<moment name without year>\" after:<prior-year-date-minus-6-months> before:<prior-year-date-plus-6-months>" --limit 10
```

Examples:

```bash
newsjack news search --query "\"Dreamforce\" \"AI agents\" after:2025-03-15 before:2026-03-15" --limit 10
newsjack news search --query "\"Pumpkin Spice\" coffee after:2025-03-01 before:2026-03-01" --limit 10
newsjack news search --query "\"Black Friday\" \"coffee subscription\" after:2025-05-27 before:2026-05-27" --limit 10
```

Read the results like a PR strategist:

- What format landed: trend piece, service journalism, data story, founder analysis, gift guide, event preview, expert reaction, local angle, product roundup, or trade recap.
- What proof made it publishable: proprietary data, expert access, customer example, research, product news, survey, price data, local presence, or cultural timing.
- Which outlets and desks covered it.
- Whether the coverage was actually about the moment or only loosely adjacent.

Use that pattern to improve the plan's angle seeds. If last year's coverage was mostly gift guides, do not pitch a generic thought-leadership essay; suggest gift-guide assets and deadlines. If last year's coverage was analyst/expert reaction, suggest a tight expert POV with the right spokesperson. If coverage was thin or SEO-heavy, lower confidence or keep the moment on watch.

## Multiple Profiles

For one profile, do one calendar query and one review pass.

For multiple profiles, do not mix them in one giant review. Run each profile independently. If your runtime supports subagents, spawn one subagent per profile with the profile JSON and the same six-month window. Each subagent returns the calendar command it ran, selected moments, prior-year coverage searches, coverage patterns, avoid items, and gaps or clusters. Merge the profile-level outputs only after those independent reviews are complete.

## Output

Write readable markdown in this order:

**Act this week.** Moments whose pitch window is open, closing, or already missed. Put the user's next action first.

**Six-month plan.** Keep this table tight:

| Date | Moment | Standing | Angle seed | Pitch window | Status | Evidence |
|---|---|---|---|---|---|---|

Use one-line angle seeds. They are prompts for `angle-generator`, not finished angles. Evidence should include one source URL or API prior-coverage anchor plus, when researched, one short prior-year coverage pattern such as "2025 coverage was mostly gift guides" or "trade coverage favored expert reaction."

**Watch list.** Optional. Include only plausible moments that need a stronger proof asset.

**Avoid.** Sensitive or unsafe moments with a one-line reason.

**Gaps & clusters.** Quiet weeks to fill and crowded weeks to stagger. If a marquee moment for the brand appears to be missing from the feed, call it out here only. Do not place it in the plan unless the user supplied the date or you run a separate live query/source check that verifies it.

**Coverage patterns.** Briefly summarize what prior-year coverage taught you for the selected moments. Keep this to bullets, not a second report.

**Command trail.** List the exact `newsjack pr-calendar query` command(s) or Medialyst MCP tool call(s), the profile file(s), whether login was already configured or had to be completed before retrying, and the `newsjack news search` / `mcp__medialyst__search_news` queries used for prior-year coverage.

## Handoff

When the user picks a moment, hand it down the existing chain. Do not duplicate that work here:

1. `angle-generator` - the calendar moment is the news peg.
2. `find-journalists` - build the journalist list for the selected angle.
3. Drafting, then `fact-check`, then `voice-extractor` before anything is sent.

Provide a compact machine handoff as a secondary block:

```json
{
  "moment": "<name>",
  "date": "<YYYY-MM-DD>",
  "news_peg": "<one-line peg>",
  "standing": "strong",
  "proof": ["<data/customer/spokesperson the brand can attach>"],
  "pitch_window": {
    "start": "<YYYY-MM-DD>",
    "deadline": "<YYYY-MM-DD>"
  },
  "source_urls": ["<url>"]
}
```

## Boundaries

- **CLI/API owns the data:** dated opportunities, source URLs, prior coverage, tags, event type, and API pitch-window fields.
- **News-search owns coverage evidence:** dated, attributed prior-year articles used to see what landed before.
- **This skill owns the judgment:** standing, proof needs, brand safety, lead-time interpretation, coverage-pattern interpretation, drumbeat shaping, and the readable plan.
- **Downstream skills own execution:** angles, journalist lists, drafts, fact-checking, and voice matching.
- **Stateless by default:** do not save a calendar, seen-store, or inferred preferences unless the user explicitly asks to promote them into a profile or brief.
