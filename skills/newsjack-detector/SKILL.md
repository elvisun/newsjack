---
name: newsjack-detector
description: "Monitor current news and reaction signals, then decide which are credible newsjacking opportunities for a client. Uses the local monitoring engine for evidence, but the skill owns PR judgment, brand safety, standing, decay, angle fit, and handoff."
when_to_use: "User wants to monitor news for pitchable hooks, find newsjacking opportunities, react to breaking industry news, watch competitors/topics, or decide whether a current signal is worth turning into an angle or reactive comment."
---

# Newsjack Detector

Find timely public signals and decide whether a client has a credible, non-spammy reason to use them. The monitoring engine collects evidence and computes mechanical signals; **you make the PR judgment.**

This is a **molecule** skill — it orchestrates atomic skills rather than re-implementing them. Coarse relevance goes to `relevance-coarse-filter`, story identity to `story-origin-check`, angle fit to `angle-generator`, and handoff to `reactive-comment` / `journalist-fit-check` / `meanest-editor`. Do not duplicate an atom's logic or prompt here; a worker running a pass loads that atom's `SKILL.md` directly, so the atom stays the single source of truth.

## Required Workflow (follow in order)

**Default mode: run the canonical pipeline and return a report.** This skill exists to produce a freshness-gated newsjack report, including for scheduled/cron runs. Execute by default — only drop into discussion/planning when Step 2 is blocked.

1. **CHECK DOCTRINE.** If `skills/ETHICS.md` or `skills/WHY-NOT-SPAM.md` exist, follow them. This skill refuses tragedy hooks, fabricated standing, fake urgency, and spray-and-pray output. These blocks are absolute and override every later step.

2. **ANCHOR THE CLIENT — ASK FIRST ONLY IF BLOCKED.** Identify company, topics, competitors, spokespeople, standing, and client-specific exclusions, from a profile JSON or plain-text context.
   - No profile **and** no usable client context → ask for it before running. Never invent profile facts.
   - Genuinely ambiguous (which client? which topic? one-off vs recurring?) → ask one clarifying question, then proceed. Otherwise do not stall the run.
   - Missing standing is not a blocker: monitor, but mark opportunities `weak`/`no-standing`.

3. **PICK THE RUN SHAPE.**
   - One-off / "what's moving on X" → **Quick Run** below.
   - Real judgment, agent run, or scheduled job → **Canonical Pipeline** below (the default for any output a human or pitch will rely on).
   - Recurring / cron feed monitoring → Canonical Pipeline plus the recurring rules in **Freshness Gate** (`--feed-only --new-only --max-age-hours 24`, hard freshness gate).

4. **RUN THE PIPELINE.** Execute the chosen path end to end. For anything beyond a Quick Run, never skip the story-origin / freshness gate.

5. **JUDGE — NEVER TRUST MECHANICS AS PERMISSION.** `routing.queue_priority` and `story_size` are recall pressure, not pitch permission. You decide newsjacking-worthiness, standing, journalist shape, and brand safety (see **Engine vs Skill Boundary** and `rubric.md`). Gate angle fit through `angle-generator`.

6. **VERIFY & CONCLUDE.** Run the **Completion Checklist**, then report: the `run.md` path, whether coarse passes were cost-optimized or fallback, whether every surfaced signal has verified ≤24h first-public freshness, and top findings.

## Engine vs Skill Boundary

The Go CLI owns (mechanical, deterministic):

- ingestion, dedupe, clustering, novelty tracking
- mechanical scores only: freshness, source agreement, novelty, profile match, source quality, momentum, major-news weight
- deterministic story-size scoring from news-search metadata: log-scaled estimated monthly traffic + domain authority, with coverage spread across independently surfaced domains
- deterministic hygiene filtering for docs/help/product/SEO pages
- coarse-relevance application via `newsjack filter-apply`, plus two recall guards: a **big-story guard** that upgrades *any* `reject` of a `high`/`major` `story_size` signal to `monitor_only` (`big_story_recall`) — the cheap pass can never hard-drop a big story — and a **profile-match guard** that upgrades `reject/no_profile_bridge` to `monitor_only` when detector/profile evidence already matched the client, a competitor, or a profile term
- deterministic freshness gating via `newsjack origin-apply`
- operational routing: lane, queue priority, threshold-demotion flag
- deterministic safety flags

You own (PR judgment):

- whether the signal is newsjacking-worthy and whether the client has standing
- same-story / original-coverage judgment (via `story-origin-check`)
- final decay explanation from `freshness_gate`
- journalist shape, brand-safety judgment, and handoff to the next skill

Never treat `routing.queue_priority` as permission to pitch — it is only operational queue order.

## Quick Run

One-off discovery and scans:

```bash
~/.newsjack/bin/newsjack detector run "QUERY" --profile profile.json --save
```

The detector emits JSON only; render any human scan yourself from the artifact facts. Use `--mock` for local verification without credentials. Full flag/source/env reference: `references/engine-cli.md`.

For each queued signal, inspect title, sources, evidence URLs, age, `routing.lane`, `mechanical_scores` (`major_news`, `novelty`, `source_agreement`), profile matches, and safety flags. For `x` evidence inspect `x_signal_type`, `x_social_signals`, `x_author_followers`, `x_query_counts`; treat lone low-reach posts as noise. A high `major_news` means the story is broadly important, **not** that the client has standing. Treat engine age/decay as provisional until `story-origin-check` verifies the first-public clock. Then apply `rubric.md` and the **Output Format**.

## Canonical Pipeline

The artifact contract is the source of truth. Write all artifacts to a timestamped run folder:

```text
RUN_DIR/
  candidates.json              # 1. detector output
  coarse_relevance_decisions.json   # 2. coarse pass
  relevant_candidates.json     # 3. filter-apply
  clustered_candidates.json    # 3b. cluster — same-story dedup + stale pre-gate
  origin_findings.json         # 4. story-origin pass (representatives only)
  targeted_candidates.json     # 5. origin-apply (freshness authority)
  triaged_candidates.json      # 5b. newsjack-triage — standing + consolidation
  final_report.md              # 7. compiled 3-bucket scan (pitch-ready / big stories / watch)
  run.md                       # 8. skill-rendered — THE human-facing artifact
  detector.stderr.log  commands.log  summary.json
```

Only `run.md` is human-facing; the rest are provenance.

1. **Run the detector and save candidates.** This is the **canonical invocation** — use it verbatim for any run a human or pitch will rely on, across every harness, so runs stay comparable:

   ```bash
   ~/.newsjack/bin/newsjack detector run "QUERY" --profile profile.json --sources news_search,x --lookback-days 1 --depth quick --limit 80 --min-queue-priority 40 --min-major-news 0.55 > candidates.json
   ```

   The floors `--min-queue-priority 40` and `--min-major-news 0.55` are the engine defaults; they define the emitted pool. **Do not lower them and do not pass `--include-all-scored` or `--no-hygiene-filter`** (debug-only) for a real run — they change which signals reach the report and make two runs of the same profile incomparable. The positional `"QUERY"` is only a label/seed; the engine retrieves from the profile's `search_terms`, so keep the profile authoritative rather than hand-tuning the query per harness. For recurring/cron precision add `--demote-unmatched-x` (see **Freshness Gate**); that is the only flag the canonical command grows.

2. **Coarse relevance pass** → `coarse_relevance_decisions.json`. High-recall junk removal only — no ranking, angles, dates, or pitch decisions. Each worker loads `skills/relevance-coarse-filter/SKILL.md` and applies it to its assigned signals; merge every worker's output into one `decisions` array. For model/worker routing and chunking, see `references/harness-routing.md`.

3. **Apply coarse decisions:**

   ```bash
   ~/.newsjack/bin/newsjack filter-apply --candidates candidates.json --decisions coarse_relevance_decisions.json --include keep --include monitor_only --output relevant_candidates.json
   ```

3b. **Cluster same-story signals before the expensive retrieval pass:**

   ```bash
   ~/.newsjack/bin/newsjack cluster --candidates relevant_candidates.json --drop-stale --window-hours 24 --output clustered_candidates.json
   ```

   The Go CLI collapses syndicated pickups / near-duplicate headlines of the **same public event** into one representative (it shares findings, so 15 NVIDIA-GTC copies cost one story-origin retrieval, not 15) and records the rest in `clustered_duplicates`. `--drop-stale` deterministically pre-gates low-story-size signals whose detector decay is clearly outside the window (`week`/`month`) into `pre_gated_stale`, so they skip retrieval entirely; large stories (`high`/`major`) are always researched regardless of age. Run story-origin on `clustered_candidates.json` (representatives only). Disclose how many duplicates and stale items were collapsed.

4. **Story-origin pass** on `clustered_candidates.json` (representatives) → `origin_findings.json`. Each worker loads `skills/story-origin-check/SKILL.md` and applies it per signal: decide same-story vs material-new-development, recover `first_public_at`, `original_url`, and canonical major coverage. It must **not** compute `fresh`/`stale`, must return **one finding per signal (never skip)**, and must cite **≥2 independent corroborating sources** to support a fresh clock.

   **Retrieval pre-flight is mandatory.** Before creating or delegating story-origin work, prove the execution surface can run live `news_search` and fetch/open surfaced URLs (`WebFetch`, browser/page fetch, or equivalent). If retrieval is unavailable, stop the pipeline before `origin_findings.json`; report `story_origin_retrieval_unavailable` with the missing tools and the exact retry action (`newsjack mcp setup --runtimes <runtime>`, run story-origin in the retrieval-capable main harness, or pass extracted page/search evidence to workers). Do **not** synthesize fallback findings from detector metadata.

   Merge the per-signal results into one `findings` array, keyed by `signal_id`. Validate the count against the input and re-run any gaps. Also validate the retrieval contract: any `same_story`, `fresh_new_development`, or `different_story` finding with zero `timestamp_evidence` and zero `evidence_urls` is invalid; if a whole batch comes back with empty evidence arrays, treat it as a harness retrieval failure, not as 15 unverified stories. The story-origin pass needs retrieval — see `references/harness-routing.md`.

5. **Apply the deterministic freshness gate:**

   ```bash
   ~/.newsjack/bin/newsjack origin-apply --candidates clustered_candidates.json --origins origin_findings.json --window-hours 24 --output targeted_candidates.json
   ```

   Never run `origin-apply` on `story_origin_retrieval_unavailable` output or on fallback findings that failed the retrieval validation above. In that case the correct scheduled-run result is an actionable harness/tooling error, not a `run.md` that quietly moves every representative to `unverified_no_corroboration`.

   The Go CLI is the freshness authority — it computes `freshness_gate.computed_status` from the run timestamp and cutoff. If an LLM labels May 8 fresh for a May 25 run, `origin-apply` marks it stale. Non-fresh signals carry a specific reason: `stale`, `unverified_no_corroboration` (worker cited <2 independent sources — a pipeline/worker-quality miss), `unverified_boundary` (date-only clock straddling the cutoff), or `unverified_no_timestamp` (no clock recovered). Distinguish these in the report and in metrics: `unverified_no_corroboration` means *we* didn't verify, not that the story is old.

5b. **Standing triage** on the selected fresh signals in `targeted_candidates.json` → `triaged_candidates.json`. Load `skills/newsjack-triage/SKILL.md`: re-consolidate any same-story representatives that slipped through, assign `strong`/`partial`/`none` standing with a journalist-shape sanity check, and **route each story to a tier**: `pitch_ready` (strong, or partial with a sharp shape), `big_story` (a fresh `high`/`major` story that lacks standing — **never dropped**, always surfaced as a suggestion with a `bridge_note` + `relevance_confidence`), or `watch` (small/non-big with no standing, off-beat, duplicate). This is the standing gate the engine cannot make — it replaces ad-hoc orchestrator judgment so the decision is auditable. Only `watch` withholds a story, and only for items that are neither pitchable nor big.

6. **Angle generation** on the **routed** candidates in `triaged_candidates.json`. Run `angle-generator` in **pitch mode** on `pitch_ready` items (a candidate is pitchable only if it yields ≥1 honest, journalist-shaped angle; zero viable angles downgrades it to `big_story` if the story is big, else `watch`) and in **exploratory mode** (`context.mode: exploratory`) on `big_story` items (at most one tentative `suggestion` angle; an empty result is fine and does **not** drop the story — it still appears as "awareness only").

7. **Compile `final_report.md`** — a 3-bucket scan, story-first and skimmable. The fixture's `scripts/build_report.py` is the reference implementation; the skill owns the human report shape. Lead with a **Today's read** line (`N pitch-ready · M big stories · K watched`) and a funnel line that asserts nothing pitchable or big was dropped off-screen. Then three sections, organized by the two independent axes — **standing** (can the client act?) and **magnitude** (how big is the story?):
   - `## ✅ Pitch-Ready` (`pitch_ready` tier): each story shows freshness (with **both** the first-public date *and* the new-development date for `fresh_new_development`), standing, the angle-generator angles, and its link provenance.
   - `## 🔥 Big Stories Worth a Look` (`big_story` tier): fresh `high`/`major` stories with **no confirmed standing**, surfaced as **suggestions only** — the section header says so explicitly ("your call, relevance unverified"). **Sorted by coverage spread (distinct surfaced outlet count) desc**, no cap. Each shows the magnitude label + outlet count, freshness, the honest `bridge_note`, confidence flags (incl. the coarse `weakness_flag` → e.g. `⚠ possible keyword match`), provenance, and at most one `suggestion`-tagged angle (or "no clean angle — awareness only"). This is how we surface big stories without ever making the drop decision; telling a real story apart from a high-authority-domain artifact is done by **ranking and flagging here**, never by dropping upstream.
   - `## 👀 Watch / Context`: `watch`-tier (fresh but no standing, non-big) plus freshness-gated items (`stale`/`unverified_*`), with plain reasons and dates. Big-but-stale items are marked.
   - **Link provenance (all sections):** **One main source = the source of record** — the article the detector actually surfaced, real `published_at`, flagged when thin (`⚠ single source`, `⚠ source of record is an aggregator`). **Related coverage** underneath: clustered duplicate pickups (tagged `surfaced duplicate`) plus any `canonical_coverage_url`/`original_url` the worker *proposed*, shown with date marked **unverified** and tagged `proposed by research — UNVERIFIED`. **Never promote a worker-proposed link into the main-source position** — the anti-laundering rule. Every link carries a date.

   Links must be clickable Markdown, not backticked or bare URLs. Do not present mechanical rank as a final fit verdict.

8. **Write `run.md` yourself from the artifacts.** The CLI does not render reports. It only emits deterministic JSON. Use `final_report.md` plus the artifact facts to write a human-facing `run.md` in the run folder.

   The report must be rendered from the **gated/fresh/triaged artifacts**, never raw `candidates.json` alone. Do not resurface coarse-rejected or hard-safety-flagged signals in the ✅/🔥 sections. The only hard drops are mechanical (URL-pattern hygiene) and hard-safety flags; disclose their counts from the JSON artifacts so nothing is hidden — never silently truncate. If you need a machine-readable artifact index, run:

   ```bash
   ~/.newsjack/bin/newsjack run-summary targeted_candidates.json --output summary.json
   ```

   `run-summary` writes JSON metadata only; it does not write Markdown or make editorial decisions.

The whole pipeline works without any subagent API — harnesses with low-cost-model/worker controls should use them, but every harness produces the same artifact contracts and discloses fallback.

## Freshness Gate

For recurring scheduled output, a signal is not surfaceable until its Go-computed `freshness_gate.computed_status` is verified. News-search `published_at` values are good article-publication evidence for recovering originals, but they alone never decide same-story status or first publication — that is the `story-origin-check` atom's job.

Recurring output rules:

- Surface only `fresh` or `fresh_new_development`. Reject `stale` and every `unverified_*` status. The unverified statuses are distinct on purpose: `unverified_no_corroboration` (worker cited <2 independent sources — a *pipeline* miss, often re-runnable), `unverified_boundary` (date-only clock straddling the cutoff), `unverified_no_timestamp` (no clock recovered). Report them separately so worker-quality misses are not mistaken for genuinely old stories.
- Run with `--demote-unmatched-x` so unmatched X News/Trends clusters fall below the queue floor unless the large-story recall guard lifts them. X News surfaces for review by default; recurring precision wants it demoted unless it is a genuinely large story.
- Cluster (step 3b) before retrieval and prefer `--drop-stale` so syndicated duplicates and clearly-old low-value items never burn story-origin retrieval.
- Do **not** reset the clock for AOL, Yahoo, MSN, Apple News, partner syndication, wire pickup, SEO rewrites, or "published today" pages whose canonical/source story is older.
- A newer article restarts the clock only if it adds a concrete new public fact: official action, filing, statement, data/report publication, material company update, new local impact, or another independently coverable development.
- Prefer `story_origin.canonical_coverage_url` as the report's main link — the major/most authoritative same-story coverage, not the random pickup that triggered retrieval.

`origin-apply` attaches `story_origin` and the deterministic `freshness_gate` to selected and rejected signals. If the first-public timestamp can't be verified, write `first_public_at: null` and explain the gap; `origin-apply` computes the appropriate `unverified_*` status.

## Handoff

- Breaking / same-day sourced comment → `reactive-comment`
- Needs story framing → `angle-generator`
- Named journalist check → `journalist-fit-check`
- Draft critique → `meanest-editor`

## Completion Checklist

Before reporting the run complete:

- `coarse_relevance_decisions.json` has exactly one decision per emitted candidate (unless `--allow-missing`).
- `clustered_candidates.json` was produced by `cluster`; story-origin ran on its representatives, and the run disclosed how many duplicates/stale items were collapsed.
- Story-origin retrieval pre-flight passed in the execution surface that actually produced `origin_findings.json`; no missing-retrieval fallback was used.
- `origin_findings.json` has exactly one finding per clustered representative (unless `--allow-missing`) — count validated, gaps re-run, and no non-`unclear` finding is empty of both `timestamp_evidence` and `evidence_urls`.
- `targeted_candidates.json` was produced by `origin-apply`; `triaged_candidates.json` was produced by `newsjack-triage` with a `tier` per signal; `pitch_ready` went to `angle-generator` in pitch mode and `big_story` in exploratory mode.
- No fresh `high`/`major` story was routed to `watch` — every fresh big story appears in **🔥 Big Stories Worth a Look** (or **✅ Pitch-Ready** if it earned standing).
- `final_report.md` is the 3-bucket scan (✅ Pitch-Ready / 🔥 Big Stories Worth a Look / 👀 Watch / Context), written from `targeted_candidates.json` / `triaged_candidates.json`, not raw `candidates.json`.
- `run.md` was skill-rendered from the gated/fresh/triaged artifacts after `final_report.md` existed — never from raw `candidates.json` alone.
- The ✅/🔥 sections contain **no** coarse-rejected or hard-safety-flagged signal; the only hard drops (URL-hygiene + hard-safety) have their counts disclosed from the JSON artifacts.
- The final response names the `run.md` path, the cost-optimized-vs-fallback status, whether every surfaced signal has verified ≤24h first-public freshness, and top findings.

## Output Format

Return exactly this JSON object. No prose before or after it. Every opportunity must include source URLs in `evidence_used` — `story_origin.canonical_coverage_url` first when present, then the original/source URL and other support (usually 1–3 links across news, RSS, and X).

```json
{
  "opportunities": [
    {
      "signal_id": "engine signal id",
      "signal_title": "Observed public signal",
      "verdict": "pitch_now",
      "decay": {
        "stage": "4hr",
        "rationale": "Why this clock applies"
      },
      "story_size": {
        "band": "low | moderate | high | major",
        "score": 0,
        "rationale": "How publication traffic/domain authority and coverage spread should affect effort priority"
      },
      "first_publication": {
        "status": "fresh | fresh_new_development",
        "first_public_at": "ISO timestamp or YYYY-MM-DD",
        "original_url": "https://...",
        "canonical_coverage_url": "https://... or null",
        "canonical_coverage_source": "Outlet/source name or null",
        "rationale": "Why this first-public clock controls"
      },
      "why_newsjacking_worthy": "Specific reason this is timely and not generic trend-chasing.",
      "client_standing": {
        "assessment": "strong | partial | weak",
        "rationale": "What gives the client standing, or what is missing"
      },
      "journalist_shape": {
        "beat_description": "Specific reporter shape, not a name",
        "why_they_care_now": "Why this beat plausibly cares now",
        "do_not_target": "Who should not receive this"
      },
      "evidence_used": [
        {
          "source": "news_search",
          "title": "Evidence title",
          "url": "https://...",
          "published_at": "YYYY-MM-DD"
        }
      ],
      "next_skill": "angle-generator"
    }
  ],
  "rejected_signals": [
    {
      "signal_id": "engine signal id",
      "signal_title": "Rejected public signal",
      "reason": "no_client_standing",
      "first_publication": {
        "status": "stale | unverified_no_corroboration | unverified_boundary | unverified_no_timestamp | null",
        "first_public_at": "ISO timestamp, YYYY-MM-DD, or null",
        "original_url": "https://... or null",
        "canonical_coverage_url": "https://... or null"
      }
    }
  ],
  "brand_safety_blocks": [
    {
      "signal_id": "engine signal id",
      "signal_title": "Blocked public signal",
      "reason": "tragedy_or_human_suffering"
    }
  ],
  "monitor_notes": [
    "Operational note or missing source, if relevant"
  ]
}
```

- Allowed verdicts: `pitch_now`, `develop_angle`, `monitor`, `reject`.
- Allowed rejection reasons: `stale`, `freshness_unverified` (umbrella; or the specific `unverified_no_corroboration` / `unverified_boundary` / `unverified_no_timestamp`), `single_source`, `no_client_standing`, `no_journalist_shape`, `off_beat`, `already_seen`, `weak_signal`, `no_viable_angle`.
- Allowed brand-safety block reasons: `tragedy_or_human_suffering`, `client_exclusion`, `regulated_claim_risk`, `fabrication_risk`.
