# Agent-native media lists: drop the MCP, go CLI + local-first

_Research note + v1 design. Written after a live test run of the Medialyst MCP through
the `find-journalists` skill (2026-06-12)._

## TL;DR

The Medialyst MCP exposes a **spreadsheet engine** (columns, rows, cells, formulas,
views, workflow types, opaque IDs). The agent's actual job is **"give me a short,
fit-checked list of journalists to pitch for _this_ story."** Those shapes are far apart,
and the agent pays a heavy translation tax in both directions.

**v1 direction:**
- **Drop the MCP protocol entirely.**
- **Expose one simple remote API: journalist contact + intel.** That's the only piece an
  agent can't do for itself (email discovery, deliverability, profile/byline scraping).
- **Wrap it in the `newsjack` CLI.** The CLI owns orchestration; the agent owns judgment
  (search relevance, fit against the angle).
- **The list lives on the user's machine** as a local file — diffable, editable, the user
  prunes it by hand.
- **Sending is out of scope for v1.** Later, an optional upload endpoint can take a local
  list and help send. The v1 file format just needs to be stable enough to upload later.

The whole `apply_table_action` / formula / view / `workflowType` surface — the source of
every pain point below — doesn't get ported. It disappears from the agent's contract.

---

## The test run

Task: "find journalists covering AI coding agents and what they mean for junior developer
jobs." Driven exactly as `find-journalists/SKILL.md` instructs, via the recommended
`source.type=prompt` path.

- `search_news` — fine. Dated, attributed articles. Already agent-shaped.
- `create_media_list` (`limit: 8`, `run_initial_enrichment: true`):
  - **`limit: 8` was ignored.** The prompt was expanded into **17 keyword queries** and
    **60 rows** imported; 180 credits held to enrich all of them.
  - Most rows were **junk for this brief** — aggregators/content farms (PRESS Insider,
    Newsline, News-Medical, Nearshore Americas) and off-angle pieces.
  - **Every `author` was `null`.** The list is article-shaped until enrichment runs.
- Reading results back (`inspect_table`, `cellDetail: full`) — see pain points.

(I deleted the test list afterward; it was still burning credits on off-angle rows.)

---

## Why the table model fights the agent

Every point is from the run.

1. **The agent loses control of size and quality.** Asked for 8, got 60, mostly
   off-brief, all auto-enriched before any fit judgment. The skill's "keep it to 5–15"
   rule is unreachable through this path.
2. **Article-shaped, not journalist-shaped.** `author` is `null`; the people only appear
   three async stages deep, after a chained pipeline you have to poll.
3. **Scoring answers the wrong question.** The recipe scores
   journalist-vs-article-topic. There's nowhere to put _my pitch / angle / standing_, so
   "Pitch Angle" and "Why They Fit" came back as generic slop ("our company's
   solution…"). Both top rows scored **95**; neither matched the brief. The agent must
   discard Score/Pitch/Why and redo fit by hand.
4. **Read-back is huge, redundant, truncated.** Each cell is an 8-field envelope
   (`data`/`text`/`searchText`/`sortValue`/`truncated`/…) around a few bytes of signal;
   substantive fields arrive `"truncated": true`, forcing a _second_ `read_full_values`
   call. One article shipped a full inline `data:image/jpeg;base64,…` thumbnail into
   context. ~100 KB+ of JSON for ~15 facts that matter.
5. **Opaque IDs + BI vocabulary** for every edit (`workflowType`, `fieldMapping`,
   `formula`/`mathjs`, `manage_views`, ids like `snad0srlai7cz457xgpvrsa6`).
6. **The agent must reverse-engineer the column mapping** (email lives at
   `JournalistProfile.data.meta[0]`; outlet/beat nested in the article record).

**Through-line:** the MCP makes the agent program a spreadsheet. The table is the right
durable substrate for the _web app_; it's the wrong _interface_ for an agent whose mental
model is "a short list of people for a pitch."

---

## v1 architecture

```
agent / skill (judgment: relevance, fit vs angle)
        │
        ▼
   newsjack CLI  ──────────────►  local list file  (the user's machine)
        │   (orchestration, file I/O)
        ▼
   Medialyst REST API
   POST /v1/journalists/enrich   ◄── the ONLY remote dependency: contact + intel
   POST /v1/news/search          ◄── optional; host web search also works
```

Two principles:
- **The remote API does only what's irreplaceable** — find the journalist behind a
  byline, their email + deliverability, their recent work. No tables, no storage, no fit
  scoring, no "recipe." Stateless request/response.
- **Everything else is local.** Search relevance and fit-against-the-angle are agent
  judgment. The list is a file the CLI reads and writes and the user can edit.

### The remote API (the moat)

One endpoint matters:

```http
POST /v1/journalists/enrich
Authorization: Bearer <key>     # same auth the CLI already loads

{
  "from": [
    { "article_url": "https://fortune.com/2026/05/22/microsoft-ai-cost-problem-tokens-agents/" },
    { "name": "Anthony Ha", "outlet": "TechCrunch" }     // either shape accepted
  ],
  "include_recent": 5            // recent articles per journalist (0 to skip)
}
```

Response — a **flat array of journalist objects**, contact + intel only:

```jsonc
{
  "journalists": [
    {
      "name": "Jake Angelo",
      "outlet": "Fortune",
      "beat": "AI, business, future of work",
      "email": "jake.angelo@fortune.com",
      "email_status": "deliverable",          // deliverable | risky | unknown | not_found
      "socials": { "x": "@...", "linkedin": "..." },
      "region": "New York, US",
      "source_article": {
        "title": "Microsoft reports are exposing AI's real cost problem",
        "url": "https://fortune.com/2026/05/22/microsoft-ai-cost-problem-tokens-agents/",
        "published": "2026-05-22"
      },
      "recent": [
        { "title": "Peter Thiel warns AI is a bigger threat to technical roles…",
          "url": "https://fortune.com/article/peter-thiel-ai-skills…", "published": "2026-05-31" }
      ]
    }
  ],
  "unresolved": [
    { "article_url": "https://newsline.com/…", "reason": "no byline / aggregator" }
  ],
  "credits": { "charged": 1.8 }
}
```

Notes:
- **No `fit`, no `score`, no `pitch_angle`.** Those depend on the user's angle, which the
  agent holds. Scoring stays local. The API never sees the pitch.
- **No cell envelopes, no IDs, no truncation, no base64.** Short fields, full values.
- **`unresolved` is honest** — aggregators / no-byline / dead links come back flagged,
  not faked. (This alone kills pain points #2 and the junk-import problem: the agent
  feeds only the URLs it already judged on-topic, so there's no 60-row dragnet.)
- Optional companion `POST /v1/news/search` (Medialyst's existing search, cheap and
  clean) — but the CLI can also fall back to host web search, so it's not required.

### The CLI

```bash
# optional: search (wraps /v1/news/search, or host web search)
newsjack news "AI coding agents junior developer jobs" --recency 30d        # -> JSON articles

# core: resolve bylines -> contact + intel (wraps /v1/journalists/enrich)
newsjack enrich https://a.com/x https://b.com/y                              # -> JSON journalists
echo '{"from":[…]}' | newsjack enrich                                        # stdin JSON, no shell-quoting

# orchestrator the skill drives: search -> let agent pick -> enrich -> write local file
newsjack find-journalists --out ai-dev-jobs.journalists.jsonl
```

- **Structured input via stdin JSON** removes the one ergonomic edge MCP had (typed args
  without shell-escaping). Agents are great at "write a JSON blob, pipe it, read stdout."
- **List management is just files + standard tools.** Pruning = the agent (or user)
  deletes lines / edits the JSONL. No `delete_rows`, no row-IDs. `jq` filters views.
- `--help` replaces `get_usage_guide` / `get_tool_reference` — self-documenting, always
  current, zero round-trips.

### The local list

The media list is a file on the user's machine. Proposed: **JSONL, one journalist per
line** — append-only friendly, diffable, line-deletable, `jq`-able. The CLI writes the
remote `contact + intel` fields; the **agent adds the judgment fields locally**:

```jsonl
{"name":"Jake Angelo","outlet":"Fortune","email":"jake.angelo@fortune.com","email_status":"deliverable","beat":"AI, future of work","anchor":{"title":"Microsoft reports are exposing AI's real cost problem","url":"https://fortune.com/…","published":"2026-05-22"},"fit":"soft-fit","fit_reason":"Covers AI's effect on jobs; bridge via your hiring data","pitch_note":"Lead with the junior-pipeline numbers"}
```

`fit` / `fit_reason` / `pitch_note` / `status` are **agent-written, local**, using the
skill's existing `fit | soft-fit | research-needed | cut` model. Optionally render a
human-readable `ai-dev-jobs.md` table alongside it. Default location: cwd (configurable);
the file is the artifact the user reviews and owns.

### End-to-end flow (the rewritten skill)

1. Agent searches (`newsjack news` or host web search).
2. **Agent picks** the on-topic, real-byline articles — judgment it's good at.
3. `newsjack enrich <those URLs>` → contact + intel. (Only the chosen ones; no dragnet.)
4. **Agent scores fit vs the pitch**, writes the JSONL locally, drops weak rows.
5. Presents the markdown table + the cuts.

v1 stops at a reviewed local file. The skill collapses from "drive a spreadsheet engine"
to "search, pick, enrich, judge, write file."

### Latency note (a side benefit)

The slow part of the test run was the **chained recipe** (profile → recent articles → AI
analysis → formula score). v1 drops all of that — enrichment is just resolve-byline +
email + recent articles, which is fast enough to run **synchronously per batch** of ~15,
streaming progress to stderr. If a batch is ever slow, `--wait=false` returns a job id +
`newsjack job status <id>`. No `enrichment_health` polling loop.

---

## What gets deleted

- `apps/cli/cmd/newsjack/mcp_bridge.go` — the stdio↔HTTP MCP bridge.
- The `configure*MCP` wiring in `mcp.go` (codex / claude / openclaw / hermes registration).
- `.mcp.json`.
- The MCP-registration steps in `install.sh`.
- The entire "Medialyst Tools (Cloud Mode)" + `apply_table_action` section of
  `find-journalists/SKILL.md`, replaced by the 5-step flow above.

Net negative LOC, and the agent's contract shrinks to a handful of self-documenting verbs.

---

## What stays Medialyst's / the v2 boundary

- **Stays remote:** the enrich endpoint — email-finder, deliverability, byline/profile
  scraping, recent-articles index. This is the only thing worth paying credits for and
  the only thing an agent can't do from web search. (Optionally `/v1/news/search`.)
- **Deferred to v2 (not required for v1):** an **upload + send** path. A local list gets
  POSTed back to a service that handles outreach (send, deliverability, reply tracking).
  v1 just needs the local file schema stable enough to upload later — which the flat
  JSONL above already is.
- **Gone for the agent:** durable server-side tables, views, share links, live column
  editing. Those remain valuable in the **web app** for power users; they're simply no
  longer on the agent's path.

---

## Measured: v1 production test (2026-06-14)

Shipped upstream in `Solar-Flare-Ventures/medialyst` PR **#1014** (enrichment API +
collapsed research) and **#1017** (parallelized hot path, shared Trigger queue
concurrency 50). Tested live against `medialyst.ai/api`.

**Task:** build a ~20-journalist, fit-scored list for the "AI coding agents vs. the
junior-dev job market" pitch.

**Flow exercised:** 5× `POST /v1/news/search` (→ 40 article URLs, ~5s) → 3×
`POST /v1/journalists/enrich` (batches of 15/15/10, since `from` caps at 15) with
`fit_context.pitch` → poll `GET /v1/journalist-enrichment-jobs/:id`.

**Result:**
| Metric | Value |
| --- | --- |
| Unique journalists resolved | **36 of 40 URLs** (4 unresolved aggregators, flagged) |
| Wall-clock for enrichment | **~6.3 min** (jobs ran concurrently; per-job 338s / 360s / 380s) |
| Search + orchestration | ~5s |
| Deliverable emails | yes — real addresses with `email_status: "deliverable"` |
| Fit-score spread | **30, 30, 65, 65, 75, 85×19, 95×15** — genuinely pitch-aware |
| Credits | 1 per resolved journalist (36) |

**The old pain points are gone:**
- **Pitch-aware scoring** — the score now ranges 30→95 with specific reasoning that cites
  the journalist's actual recent articles _and_ the pitch (e.g. Hugh Langley/Business
  Insider 95: "actively covers the impact of AI on software engineering jobs…"). The old
  recipe gave everyone 95. A non-journalist (OX Security) was correctly scored **30** —
  "a company, not a journalist" — i.e. an automatic `cut` signal.
- **Clean flat output** — `journalist_intel_v1` (name, outlet, email, `email_status`,
  socials, region, author_page, `source_article`, `recent[]`) + `journalist_research_v1`
  (`fit: { score, why_they_fit, personalized_angle }`). No cell envelopes, no opaque IDs,
  no base64, no second `read_full_values` round-trip.
- **No junk dragnet** — the agent feeds only URLs it already judged on-topic; 36/40
  resolved vs. the old 60-row prompt-expansion that was mostly off-brief.
- **Near-zero agent cost** — the agent does one search + 3 POSTs + a poll loop, then reads
  flat JSON. The ~100 KB of table-envelope parsing is gone.

**The one honest caveat — wall-clock.** ~6 min is the *backend enrichment* (per-source
byline resolution + email-finding + recent-article scraping), not the protocol. The
agent-side cost collapsed, but the scrape is the long pole: ~380s for 15 sources ≈ ~25s
per source effective, even with PR #1017's parallelization. Knobs exist if we want it
faster: `JOURNALIST_ENRICHMENT_SOURCE_CONCURRENCY` (default 8, max 25),
`JOURNALIST_ENRICHMENT_FIT_CONCURRENCY` (default 8), and the shared queue's concurrency
(50). For 20 journalists the wall-clock is the same ~6 min because batches run in
parallel — it doesn't scale with list size until you saturate the queue.

**Contract notes worth remembering:**
- `from` is **max 15 sources/request** → a 20-list means 2+ jobs fanned out.
- `options.wait` returned `202 pending` immediately in practice; treat it as poll-based.
  Jobs are durable (7-day expiry) and idempotent via `Idempotency-Key`.
- `include_recent` must be 0 or 3–20; `timeout_ms` ≤ 30000.
- The API WAF blocks the default `Python-urllib` User-Agent — set a normal UA.
