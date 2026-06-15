# Journalist Enrichment API — integration feedback & punch-list

**For:** Medialyst engineering
**From:** Newsjack (agent-native media-list integration)
**Date:** 2026-06-14
**Re:** the API shipped in PR **#1014** (`POST /api/v1/journalists/enrich`, scored
`research[]`) and **#1017** (parallelized hot path, shared Trigger queue).

## Summary

We integrated the new enrichment API as the agent-facing path for building media lists
(replacing the MCP table/recipe flow, which forced agents to program a spreadsheet). **The
API is the right shape and it works** — clean flat `journalist_intel_v1` /
`journalist_research_v1`, deliverable emails, pitch-aware scoring. Everything below is
refinement, not a redesign.

**Live test (prod, 2026-06-14):** build a ~20-journalist list for a pitch ("AI coding
agents vs. the junior-dev job market"). Flow: `news/search` → 3× `journalists/enrich`
(batches of 15/15/10) with `fit_context.pitch` → poll.

| Metric | Result |
| --- | --- |
| Resolved | 36 of 40 source URLs (4 unresolved aggregators, correctly flagged) |
| Wall-clock | ~6.3 min (3 jobs concurrent; per-job 338 / 360 / 380s) |
| Emails | real, `email_status: "deliverable"` |
| Fit scores | 30, 30, 65, 65, 75, 85×19, 95×15 — genuine spread, cites real recent articles + the pitch; a non-journalist (a company) correctly scored 30 |
| Credits | 1 per resolved journalist; unresolved not charged; `usage.credits_charged` returned |

The punch-list is ordered by impact on the core "build a list of N journalists" job.

---

## v1 request-shape: lock the surface down

Deliberate scope decisions to narrow the request surface for v1 — orthogonal to the
friction fixes below. Narrow now, widen later once the UX questions are answered.

**Before (#1014):**
```jsonc
{
  "from": [ { "type": "article_url" | "journalist_reference", ... } ],   // min 1, max 15
  "fit_context": { "pitch": "…", "model_key": "…?" },
  "include_recent": 5,                          // top-level, 0 or 3–20
  "options": {
    "wait": false, "timeout_ms": 0, "external_id": "…?",
    "country": "…?", "language": "…?"
  }
}
```

**After (v1):**
```jsonc
{
  "from": [ { "type": "article_url", "url": "https://…" } ],   // min 1, max 500 (see #1)
  "fit_context": { "pitch": "…" },
  "options": {
    "include_recent": 10,                       // moved here, default 10
    "wait": false, "timeout_ms": 0, "external_id": "…?"
  }
}
```

The four changes:

1. **`article_url` is the only supported source type.** Keep the `type` discriminator and
   the union so other source types (`journalist_reference` = name + URL/domain/publication)
   stay *selectable* in the schema for later — but v1 accepts only `article_url` and keeps
   returning the existing `UNSUPPORTED_SOURCE_TYPE` for anything else. (This already matches
   #1014's behavior — just make it the explicit, documented contract.) Cap moves to 500 per
   punch-list #1.
2. **Drop `fit_context.model_key`.** Model selection returns once we design how users pick
   models; until then the server chooses. Remove it from the schema rather than accept-and-
   ignore, so nobody builds against it.
3. **Move `include_recent` into `options`, default `10`.** Range unchanged (`0`, or `3–20`).
   It's a per-request tuning knob, so it belongs alongside the other `options`, not at the
   top level. Default rises from the #1014 example's 5 to **10**.
4. **Drop `options.country` and `options.language`.** Neither is useful for an agent or user
   to set on a per-enrichment basis; let the server default. Remove from the schema.

Schema touch-points: `packages/schemas/src/journalist-enrichment.ts` (request schema +
`IncludeRecentSchema` default), `packages/api/src/routes/v1/journalist-enrichment.ts`
(option plumbing).

---

## P0 — biggest impact

### 1. Remove (or massively raise) the 15-source request cap; fan out server-side

**Observed.** `from` is `min(1).max(15)`
(`packages/schemas/src/journalist-enrichment.ts`). Building 20 journalists already
requires the client to: split URLs into ≤15 batches, submit multiple jobs, respect the
per-key active-job cap, handle 429s, poll several job IDs, and merge results. "Give me 20
for this pitch" is never one call.

**Why this is cheap to fix.** The architecture **already fans out** — #1017 made
`journalist-enrichment` the durable parent and `journalist-enrichment-source` the queued
per-source unit on the shared `journalist-profile-enrichment` queue (concurrency 50). The
15 is just a request-validation ceiling, not an execution constraint.

**Proposed change.**
- Raise `from` max to **500** (or remove the hard cap and bound by a generous default).
  One job, N sources, fanned out to child source-tasks exactly as today.
- This also makes the **per-key active-job cap** mostly moot — agents stop needing
  multi-job fan-out, so cap contention disappears.
- Pair it with per-job fairness on the shared queue so a single 500-source job can't
  starve other keys (round-robin / max-in-flight-per-job), and with **#2 (progressive
  results)** so a big job is observable while it runs.

**Contract:** `from: z.array(Source).min(1).max(500)`. No response-shape change.

### 2. Progressive results during polling

**Observed.** `GET /journalist-enrichment-jobs/:id` returned `journalists: []` for the
full ~380s, then all 15 at once. Source tasks complete independently server-side, so
partial results exist — they're just not exposed. A 6-minute blind wait means the agent
can't show the user anything or start fit-review until the whole batch lands.

**Proposed change (either, ideally both):**
- **Live counts** in the poll payload: `progress: { total, resolved, unresolved,
  in_flight }`. Cheap; enables a progress bar.
- **Incremental results**: return `journalists[]` / `research[]` as each source resolves
  (append-only; the agent dedupes/accumulates). Poll response stays the same shape, just
  populated incrementally instead of all-at-once.

This is what turns "fire job, wait blind 6 min" into "watch the list fill in."

---

## P1 — correctness / safety

### 3. Server-side dedupe by `journalist_identity_id`; don't double-charge

**Observed.** Every result carries `journalist_identity_id` (e.g.
`ih5hv2vgqohfl72u9swda3q2`), but identical identities are not deduped within or across
jobs, and the client must dedupe by email/name. Because fit-jobs disable same-URL caching
(correct — pitch scoring must be fresh), the **same identity** surfaced by two different
source URLs is re-resolved **and re-charged**.

**Proposed change.**
- Dedupe by `journalist_identity_id` within a job before returning.
- Charge **once per distinct identity** per job (and ideally per job-fan-out group), not
  once per source URL. Fit scoring can still re-run per pitch; the billable unit should be
  the resolved identity.

### 4. Make retries safe by default

**Observed.** `Idempotency-Key` is opt-in. When our first poller crashed *after*
submitting 3 jobs, a naive resubmit would have created 3 new jobs and double-charged; we
only avoided it by manually polling the orphaned IDs.

**Proposed change.**
- Encourage/default an idempotency key derived from the request hash (the PR already
  computes a request hash for `journalistEnrichmentJobs`).
- On create, if an equivalent in-flight job exists, return it (200/202 with the existing
  `id`) instead of spawning a duplicate — and make the in-flight job **discoverable** so a
  client that lost the `id` can recover it rather than resubmit.

### 5. Poll-path robustness

**Observed.** Polling 3 jobs every ~2s tripped a WAF/rate-limit that returned a **non-JSON
(HTML) body**, which crashed a JSON client. Separately, the API **403s the default
`Python-urllib` User-Agent** — curl's UA worked, so any SDK/CLI must spoof a UA, a silent
onboarding trap.

**Proposed change.**
- **All** error responses (including 403/429 from the edge/WAF) return JSON with a stable
  `{ error, code, retry_after_seconds? }`. Never HTML to an authenticated API caller.
- Don't User-Agent-gate authenticated API traffic.
- Document the expected poll cadence; `retry_after_seconds` is already returned — consider
  long-poll support (hold up to ~25s, return on change) to cut poll volume.

---

## P2 — ergonomics

### 6. Credits: simple estimate-then-actual, no held/settled

**Drop the held/settled credit model entirely.** The endpoint is now simple enough that a
hold/settle lifecycle adds nothing. Two numbers, two moments:

- **On create/accept:** `usage.estimated_credits` — the max it will use. Cost is
  deterministic (1 × source count), so this is exact-or-upper-bound, and lets the agent
  confirm spend before committing.
- **On completion:** `usage.credits_used` — the actual total charged (1 per resolved
  journalist), **plus** `usage.not_enriched` — the count of sources that produced no
  journalist because the data was missing/unresolvable (i.e. the `unresolved[]` sources;
  per-source reasons already live in that array).

So the agent sees planned vs. actual and the reason for any gap ("estimated 15, used 11,
4 not enriched — missing data"). No `held`, no `settled`, no hold→settle visibility to
track. (Currently the flow returns settled `usage.credits_charged`; rename/extend to
`credits_used` + `estimated_credits` + `not_enriched`.)

### 7. `options.wait` is effectively a no-op

**Observed.** Sending `wait:true, timeout_ms:30000` returned `202 pending` immediately;
it never held the connection. It advertises a synchronous path that doesn't exist.

**Proposed change.** Either implement it as a real long-poll (block up to `timeout_ms`,
return completed-or-partial), or remove it so clients don't build a sync path that never
fires. Document that the flow is poll-based.

### 8. Don't invest in `beat` — the AI fit reasoning replaces it

**Observed.** `beat`, `role`, and `bio` came back `null` for several journalists. Our
first instinct was "backfill these" — but on reflection, **`beat` shouldn't be a relied-on
field at all** once the scored path is in play. `fit.why_they_fit` already encodes the
journalist↔pitch relationship *dynamically* and per-pitch (e.g. "actively covers AI's
impact on software engineering jobs, per his article '…'"), which is strictly better than
a static, frequently-stale `beat` tag. A coarse "enterprise AI" label adds nothing the
reasoning doesn't, and a wrong/empty one actively misleads.

**Proposed.**
- **Don't spend effort backfilling `beat`.** We're dropping it from the agent's list
  columns in favor of `why_they_fit`. Keep returning it if it's free, but it's not worth
  scraping work.
- `role` (staff vs. contributor/freelance) and `bio` *are* worth having when cheap —
  they're factual and don't duplicate the AI analysis. Lower priority than P0/P1.

### 9. Optional conveniences (nice-to-have)

- `min_score` request param so weak fits are dropped server-side (we currently post-filter;
  the smoke harness already hard-codes a `MEDIA_LIST_MIN_SCORE`).
- Return `research[]` pre-sorted by `fit.score` desc.

---

## Latency note (context for the cap change)

~6 min wall-clock is **backend enrichment** (per-source byline resolution + email-finder +
recent-article scrape), ~25s/source effective even after #1017's parallelization, not the
protocol. Relevant knobs: `JOURNALIST_ENRICHMENT_SOURCE_CONCURRENCY` (default 8, max 25),
`JOURNALIST_ENRICHMENT_FIT_CONCURRENCY` (default 8), shared queue
`journalist-profile-enrichment` (concurrency 50). **If `from` goes to 500, the queue
fairness story in #1 matters more** — a single large job must not monopolize the shared
50-slot queue. Progressive results (#2) also make large jobs tolerable by keeping them
observable.

## Non-goals — please don't "fix" these

- **Search and enrich stay separate.** The agent extracts URLs and judges topical
  relevance, then enriches. That boundary is deliberate — the agent owns judgment; the API
  owns the data work it can't do (contact, deliverability, scrape).
- **`fit_context.pitch` stays client-supplied.** The API should not infer the campaign; it
  scores against the pitch the agent provides.

## Priorities at a glance

| # | Change | Impact | Rough effort |
| --- | --- | --- | --- |
| 1 | `from` cap 15 → 500 + queue fairness | unblocks "list of N" in one call | low schema / med queue |
| 2 | Progressive results / live counts | removes 6-min blind wait | medium |
| 3 | Dedupe by identity + bill per identity | correctness + cost | low–med |
| 4 | Idempotent/discoverable jobs by default | prevents double-charge | low |
| 5 | JSON errors everywhere + drop UA gate | clients stop crashing | low |
| 6 | Credits: `estimated_credits` on create, `credits_used` + `not_enriched` on completion (drop held/settled) | budget control | low |
| 7 | Real long-poll or drop `wait` | removes dead path | low |
| 8 | Drop `beat` reliance (AI reasoning replaces it); `role`/`bio` only if cheap | avoids wasted scrape work | none / low |

---

_Validation: all observations are from a live prod run on 2026-06-14 against
`medialyst.ai/api/v1` using a real org API key (3 enrichment jobs, 40 source URLs, 36
resolved journalists). Happy to share the request/response captures._
