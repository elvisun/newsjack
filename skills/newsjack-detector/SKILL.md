---
name: newsjack-detector
description: "Monitor current news and reaction signals, then decide which are credible newsjacking opportunities for a client. Uses the local monitoring engine for evidence, but the skill owns PR judgment, brand safety, standing, decay, angle fit, and handoff."
when_to_use: "User wants to monitor news for pitchable hooks, find newsjacking opportunities, react to breaking industry news, watch competitors/topics, or decide whether a current signal is worth turning into an angle or reactive comment."
---

# Newsjack Detector

Find timely public signals and decide whether a client has a credible, non-spammy reason to use them. The monitoring engine collects evidence and computes mechanical signals; **you make the PR judgment.**

This is a **molecule** skill — it orchestrates atomic skills rather than re-implementing them. Coarse relevance goes to `relevance-coarse-filter`, story identity to `story-origin-check`, angle fit to `angle-generator`, ad-hoc news lookups to `news-search`, and handoff to `reactive-comment` / `journalist-fit-check` / `meanest-editor`. Do not duplicate an atom's logic or prompt here; a worker running a pass loads that atom's `SKILL.md` directly, so the atom stays the single source of truth.

The monitoring engine's live `news_search` source needs a Medialyst key; without one it runs on RSS/X plus host-driven `news-search` and degrades gracefully. Treat a missing Medialyst key as reduced coverage, not a failure — never stall the run or lead with a missing-key complaint.

## Runtime Mode

Newsjack Detector has two runtime modes:

- **Full Mode:** Use this in Claude Code, Codex, OpenClaw, Hermes, or another capable agent harness with shell, filesystem, network, and local CLI access. Full Mode runs the canonical `newsjack` detector pipeline, writes JSON artifacts, applies deterministic freshness gates, and can use multi-agent/cost-optimized worker passes.
- **Limited Mode:** Use this in Claude.ai chat, ChatGPT chat, Claude Cowork, or any restricted runtime without shell/filesystem/CLI access. Do not attempt `curl`, `npm`, or on-demand CLI installation. Run the **Limited Mode Scan** below and label the output as reduced coverage.

**Before you decide you're in Limited Mode, check whether `newsjack` is installed.** It ships as a prebuilt, bundled binary — you do **not** need Go, a compiler, or any build/install step to run it. Never look for a Go toolchain, and never declare the CLI "missing" or tell the user they need a "Go environment" without running this check first:

1. Run `newsjack --version`. If it prints a version, you're in Full Mode — use plain `newsjack ...` for every command.
2. If `newsjack` isn't on `PATH`, try the bundled location `~/.newsjack/bin/newsjack --version`. If that prints a version, use that full path in place of `newsjack` everywhere below.
3. Only if **both** fail (and you genuinely have no shell) are you in Limited Mode.

The bundled binary is almost always already installed — assume Full Mode and verify, don't assume it's missing.

## Required Workflow (follow in order)

**Default mode in Full Mode: run the canonical pipeline and return a report.** This skill exists to produce a freshness-gated newsjack report, including for scheduled/cron runs. Execute by default — only drop into discussion/planning when Step 2 is blocked. In Limited Mode, run a disclosed reduced-coverage scan instead.

1. **CHECK DOCTRINE.** If `skills/ETHICS.md` or `skills/WHY-NOT-SPAM.md` exist, follow them. This skill refuses tragedy hooks, fabricated standing, fake urgency, and spray-and-pray output. These blocks are absolute and override every later step.

2. **ANCHOR THE CLIENT — ASK FIRST ONLY IF BLOCKED.** Identify company, topics, competitors, spokespeople, standing, and client-specific exclusions, from a profile JSON or plain-text context.
   - No profile **and** no usable client context → ask for it before running. Never invent profile facts.
   - Genuinely ambiguous (which client? which topic? one-off vs recurring?) → ask one clarifying question, then proceed. Otherwise do not stall the run.
   - Missing standing is not a blocker: monitor, but mark opportunities `weak`/`no-standing`.
   - **Load the client brief.** Read the monitor's `brief.md` (its path is surfaced as `brief_path` by `monitor run`/`monitor status`, or it sits next to the profile). It is the **source of truth** for what this client will and won't pitch and how to present the scan — see **Client Brief** below. An empty/template brief carries no rules.

3. **PICK THE RUN SHAPE.**
   - Restricted chat / no CLI / no filesystem → **Limited Mode Scan** below.
   - One-off / "what's moving on X" → **Quick Run** below.
   - Real judgment, agent run, or scheduled job → **Canonical Pipeline** below (the default for any output a human or pitch will rely on).
   - Recurring / cron feed monitoring → Canonical Pipeline plus the recurring rules in **Freshness Gate** (`--feed-only --new-only --max-age-hours 24`, hard freshness gate).

4. **RUN THE PIPELINE.** Execute the chosen path end to end. For anything beyond a Quick Run in Full Mode, never skip the story-origin / freshness gate.

5. **JUDGE — NEVER TRUST MECHANICS AS PERMISSION.** `routing.queue_priority` and `story_size` are recall pressure, not pitch permission. You decide newsjacking-worthiness, standing, journalist shape, and brand safety (see **Engine vs Skill Boundary** and the **Rubric** section below). Gate angle fit through `angle-generator`.

6. **VERIFY & CONCLUDE.** In Full Mode, run the **Completion Checklist**, then report: the `run.md` path, whether coarse passes were cost-optimized or fallback, whether every surfaced signal has verified ≤24h first-public freshness, and top findings. In Limited Mode, state that no local artifacts, saved monitor state, or deterministic freshness gate were available.

## Engine vs Skill Boundary

The Go CLI owns (mechanical, deterministic):

- ingestion, dedupe, clustering, novelty tracking
- mechanical scores only: freshness, source agreement, novelty, profile match, source quality, momentum, major-news weight
- deterministic story-size scoring from news-search metadata: log-scaled estimated monthly traffic + domain authority, with coverage spread across independently surfaced domains. When authority metadata is missing for a recognized major outlet, the engine may use a low-confidence known-outlet fallback. When publication metadata is otherwise sparse, the engine may attach a low-confidence `story_size.attention_hint` from deterministic source signals such as X News clusters, major public actors, and high-stakes event terms; this is recall pressure, not proof of magnitude.
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

## Client Brief

Each monitor may carry a `brief.md` — a prose, user-owned statement of what this client will and won't pitch and how they want the scan presented. It is the **source of truth** for client pitch/output policy; the profile JSON governs *collection*, the brief governs *what gets pitched and shown*. The CLI only creates and surfaces the file (`monitor init` scaffolds it; `brief_path` is reported by `monitor run`/`monitor status`); it never parses it — reading and applying it is yours.

- **Where it binds:** triage and report rendering — never collection. Keep retrieval and the coarse pass brief-agnostic so nothing is dropped before judgment; the brief only decides what an already-collected, already-fresh item is allowed to *be* and how it's *shown*.
- **Never pitch rules** are hard: an item matching one can never be `pitch_ready`. A non-big item drops to `watch` (`client_policy_exclusion`); a fresh `high`/`major` item stays `big_story` with `off_policy: true` (the never-drop doctrine still holds — surface it, don't hide it). `newsjack-triage` enforces this.
- **Audience / We pitch** set the standing *altitude*: topical overlap is not pitchability. A story can be on-topic and still off-brief.
- **How to surface** is presentation only: collapse a section to a disclosed count, never silence it. Disclose what the brief held back (count + reason) so nothing is hidden.
- **Feedback updates the brief.** When the user reacts to a run — "too policy-heavy," "stop showing me X," "this is exactly right" — propose an edit to `brief.md` (a new *We never pitch* rule, a *How to surface* line, or a dated *Example*) so the policy is captured durably, not just for this run. Confirm the edit. An empty/template brief means run with defaults.

## Profile Setup File

The monitor profile JSON is the source of truth for collection setup: a focused set of short broad beat topics, search terms, competitors, feeds, standing, spokespeople, and exclusions. Prefer 6-8 core 2-3 word topics, with one-word topics allowed when natural. If the user wants to change what the monitor looks for, edit the profile JSON rather than generating one-off retrieval terms during a detector run.

- Installed monitors keep the setup file at `~/.newsjack/monitors/<slug>/profile.json`; `brief.md` sits next to it.
- Direct detector runs use the file passed with `--profile`.
- Fixture profiles live under `fixtures/newsjack-detector-agent/profile.<slug>.json`.

Use `newsjack-monitor-setup` when the user wants to create or materially revise a profile. Collection feedback such as "watch broader accounting firm news" belongs in `profile.json` (`topics` / `search_terms` / `feed_urls`): put 6-8 core broad beats in `topics`, and put broad retrieval terms plus named platforms/products/regulators/competitors in `search_terms`. Pitch-policy feedback such as "don't pitch policy stories" belongs in `brief.md`. After editing `profile.json`, rerun a mock or fixture smoke before trusting the next live run.

## Limited Mode Scan

Use this path when running in Claude.ai chat, ChatGPT chat, Claude Cowork, or any runtime without shell/filesystem/CLI access.

Limited Mode is useful for PR judgment, not canonical monitoring. It does not create saved monitors, write JSON artifacts, keep seen-state, run source ingestion, apply the Go freshness gate, or use cost-optimized worker passes.

1. **Anchor the client.** Use a pasted profile, user context, website summary, or plain-text description. If there is no usable client context, ask for it.
2. **Collect a small evidence set.** Use pasted links first. If the runtime has web/search tools, search recent news for the profile topics, competitors, named regulators/platforms, and any explicit user topic. Keep the query list short and disclose it.
3. **Build candidates manually.** For each candidate, keep title, source, URL, apparent publication time, why it matched the client, and any safety concerns. Do not invent publication dates, outlet names, source counts, or traffic/authority scores.
4. **Verify freshness where possible.** Prefer primary/source-of-record pages and independent coverage. Treat unverified dates as `freshness_unverified`; do not pitch them as time-sensitive.
5. **Apply PR judgment.** Use this skill's doctrine, `story-origin-check` reasoning where possible, `newsjack-triage` for standing/routing, and `angle-generator` for any pitchable item.
6. **Return an inline report.** Use the same sections as Full Mode: `Pitch-Ready`, `Big Stories Worth a Look`, `Watch / Context`, plus a short `Limited Mode Caveat` that names missing capabilities and searches/evidence used.

Never call this a canonical detector run. If the user wants saved monitors, scheduled scans, deterministic freshness gates, local artifacts, or recurring seen-state, recommend Full Mode in Claude Code, Codex, OpenClaw, or Hermes.

## Quick Run

One-off discovery and scans:

```bash
newsjack detector run --profile profile.json --save
```

The detector emits JSON only; render any human scan yourself from the artifact facts. Use `--topic "explicit user topic"` only when the user deliberately asks to add a one-off retrieval topic. Routine profile runs should rely on the profile's durable `topics` and `search_terms`, not ad hoc generated retrieval terms. Use `--mock` for local verification without credentials. Full flag/source/env reference: `references/engine-cli.md`.

For each queued signal, inspect title, sources, evidence URLs, age, `routing.lane`, `mechanical_scores` (`major_news`, `novelty`, `source_agreement`), profile matches, and safety flags. For `x` evidence inspect `x_signal_type`, `x_social_signals`, `x_author_followers`, `x_query_counts`; treat lone low-reach posts as noise. A high `major_news` means the story is broadly important, **not** that the client has standing. Treat engine age/decay as provisional until `story-origin-check` verifies the first-public clock. Then apply the **Rubric** section below and the **Output Format**.

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
   newsjack detector run --profile profile.json --sources news_search,x --lookback-days 1 --depth quick --limit 80 --min-queue-priority 40 --min-major-news 0.55 > candidates.json
   ```

   The floors `--min-queue-priority 40` and `--min-major-news 0.55` are the engine defaults; they define the emitted pool. **Do not lower them and do not pass `--include-all-scored` or `--no-hygiene-filter`** (debug-only) for a real run — they change which signals reach the report and make two runs of the same profile incomparable. Profile terms own durable retrieval; do not hand-tune the query per run unless the user explicitly asked for a one-off `--topic`. For recurring/cron precision add `--demote-unmatched-x` (see **Freshness Gate**); that is the only flag the canonical command grows.

2. **Coarse relevance pass** → `coarse_relevance_decisions.json`. High-recall junk removal only — no ranking, angles, dates, or pitch decisions. Each worker loads `skills/relevance-coarse-filter/SKILL.md` and applies it to its assigned signals; merge every worker's output into one `decisions` array. For model/worker routing and chunking, see `references/harness-routing.md`.

3. **Apply coarse decisions:**

   ```bash
   newsjack filter-apply --candidates candidates.json --decisions coarse_relevance_decisions.json --include keep --include monitor_only --output relevant_candidates.json
   ```

3b. **Cluster same-story signals before the expensive retrieval pass:**

   ```bash
   newsjack cluster --candidates relevant_candidates.json --drop-stale --window-hours 24 --output clustered_candidates.json
   ```

   The Go CLI collapses syndicated pickups / near-duplicate headlines of the **same public event** into one representative (it shares findings, so 15 NVIDIA-GTC copies cost one story-origin retrieval, not 15) and records the rest in `clustered_duplicates`. `--drop-stale` deterministically pre-gates low-story-size signals whose detector decay is clearly outside the window (`week`/`month`) into `pre_gated_stale`, so they skip retrieval entirely; large stories (`high`/`major`) are always researched regardless of age. Run story-origin on `clustered_candidates.json` (representatives only). Disclose how many duplicates and stale items were collapsed.

4. **Story-origin pass** on `clustered_candidates.json` (representatives) → `origin_findings.json`. Each worker loads `skills/story-origin-check/SKILL.md` and applies it per signal: decide same-story vs material-new-development, recover `first_public_at`, `original_url`, and canonical major coverage. It must **not** compute `fresh`/`stale`, must return **one finding per signal (never skip)**, and must cite **≥2 independent corroborating sources** to support a fresh clock. Merge the per-signal results into one `findings` array, keyed by `signal_id`. Validate the count against the input and re-run any gaps. The story-origin pass needs retrieval — see `references/harness-routing.md`.

5. **Apply the deterministic freshness gate:**

   ```bash
   newsjack origin-apply --candidates clustered_candidates.json --origins origin_findings.json --window-hours 24 --output targeted_candidates.json
   ```

   The Go CLI is the freshness authority — it computes `freshness_gate.computed_status` from the run timestamp and cutoff. If an LLM labels May 8 fresh for a May 25 run, `origin-apply` marks it stale. Non-fresh signals carry a specific reason: `stale`, `unverified_no_corroboration` (worker cited <2 independent sources — a pipeline/worker-quality miss), `unverified_boundary` (date-only clock straddling the cutoff), or `unverified_no_timestamp` (no clock recovered). Distinguish these in the report and in metrics: `unverified_no_corroboration` means *we* didn't verify, not that the story is old.

5b. **Standing triage** on the selected fresh signals in `targeted_candidates.json` → `triaged_candidates.json`. Load `skills/newsjack-triage/SKILL.md` and pass it the **client brief** when present: re-consolidate any same-story representatives that slipped through, apply the brief's **never-pitch** rules (off-policy items can never be `pitch_ready`; fresh big ones stay `big_story` with `off_policy: true`, small ones drop to `watch`/`client_policy_exclusion`), assign `strong`/`partial`/`none` standing at the brief's audience altitude with a journalist-shape sanity check, and **route each story to a tier**: `pitch_ready` (strong, or partial with a sharp shape), `big_story` (a fresh `high`/`major` story that lacks standing — **never dropped**, always surfaced as a suggestion with a `bridge_note` + `relevance_confidence`), or `watch` (small/non-big with no standing, off-beat, duplicate). This is the standing gate the engine cannot make — it replaces ad-hoc orchestrator judgment so the decision is auditable. Only `watch` withholds a story, and only for items that are neither pitchable nor big.

6. **Angle generation** on the **routed** candidates in `triaged_candidates.json`. Run `angle-generator` in **pitch mode** on `pitch_ready` items (a candidate is pitchable only if it yields ≥1 honest, journalist-shaped angle; zero viable angles downgrades it to `big_story` if the story is big, else `watch`) and in **exploratory mode** (`context.mode: exploratory`) on `big_story` items (at most one tentative `suggestion` angle; an empty result is fine and does **not** drop the story — it still appears as "awareness only").

7. **Compile `final_report.md`** — a 3-bucket scan, story-first and skimmable. The fixture's `scripts/build_report.py` is the reference implementation; the skill owns the human report shape. Lead with a **Today's read** line (`N pitch-ready · M big stories · K watched`) and a funnel line that asserts nothing pitchable or big was dropped off-screen. Then three sections, organized by the two independent axes — **standing** (can the client act?) and **magnitude** (how big is the story?):
   - `## ✅ Pitch-Ready` (`pitch_ready` tier): each story shows freshness (with **both** the first-public date *and* the new-development date for `fresh_new_development`), standing, the angle-generator angles, and its link provenance.
   - `## 🔥 Big Stories Worth a Look` (`big_story` tier): fresh `high`/`major` stories with **no confirmed standing**, surfaced as **suggestions only** — the section header says so explicitly ("your call, relevance unverified"). **Sorted by coverage spread (distinct surfaced outlet count) desc**, no cap. Each shows the magnitude label + outlet count, freshness, the honest `bridge_note`, confidence flags (incl. the coarse `weakness_flag` → e.g. `⚠ possible keyword match`), provenance, and at most one `suggestion`-tagged angle (or "no clean angle — awareness only"). This is how we surface big stories without ever making the drop decision; telling a real story apart from a high-authority-domain artifact is done by **ranking and flagging here**, never by dropping upstream.
   - `## 👀 Watch / Context`: `watch`-tier (fresh but no standing, non-big) plus freshness-gated items (`stale`/`unverified_*`), with plain reasons and dates. Big-but-stale items are marked.
   - **Link provenance (all sections):** **One main source = the source of record** — the article the detector actually surfaced, real `published_at`, flagged when thin (`⚠ single source`, `⚠ source of record is an aggregator`). **Related coverage** underneath: clustered duplicate pickups (tagged `surfaced duplicate`) plus any `canonical_coverage_url`/`original_url` the worker *proposed*, shown with date marked **unverified** and tagged `proposed by research — UNVERIFIED`. **Never promote a worker-proposed link into the main-source position** — the anti-laundering rule. Every link carries a date.

   Links must be clickable Markdown, not backticked or bare URLs. Do not present mechanical rank as a final fit verdict.

   **Honor the client brief's *How to surface*** here: if the brief asks to collapse a section (e.g. the big-stories/awareness section), render it as a one-line **disclosed count with reasons**, never silence it. Lead with whatever the brief prioritizes. State plainly when the brief moved an item out of `pitch_ready` or collapsed a section, and quote the rule (`policy_rule`). `scripts/build_report.py` is the brief-agnostic mechanical reference (`final_report.md`); the brief is honored in the skill-rendered `run.md`.

8. **Write `run.md` yourself from the artifacts.** The CLI does not render reports. It only emits deterministic JSON. Use `final_report.md` plus the artifact facts to write a human-facing `run.md` in the run folder.

   The report must be rendered from the **gated/fresh/triaged artifacts**, never raw `candidates.json` alone. Do not resurface coarse-rejected or hard-safety-flagged signals in the ✅/🔥 sections. The only hard drops are mechanical (URL-pattern hygiene) and hard-safety flags; disclose their counts from the JSON artifacts so nothing is hidden — never silently truncate. If you need a machine-readable artifact index, run:

   ```bash
   newsjack run-summary targeted_candidates.json --output summary.json
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

Before reporting a Full Mode run complete:

- `coarse_relevance_decisions.json` has exactly one decision per emitted candidate (unless `--allow-missing`).
- `clustered_candidates.json` was produced by `cluster`; story-origin ran on its representatives, and the run disclosed how many duplicates/stale items were collapsed.
- `origin_findings.json` has exactly one finding per clustered representative (unless `--allow-missing`) — count validated, gaps re-run.
- `targeted_candidates.json` was produced by `origin-apply`; `triaged_candidates.json` was produced by `newsjack-triage` with a `tier` per signal; `pitch_ready` went to `angle-generator` in pitch mode and `big_story` in exploratory mode.
- No fresh `high`/`major` story was routed to `watch` — every fresh big story appears in **🔥 Big Stories Worth a Look** (or **✅ Pitch-Ready** if it earned standing). A brief never-pitch rule may move a big story *out of pitch-ready*, but it stays a surfaced `big_story` (`off_policy: true`), never dropped.
- If a client brief is present, the report applied it: off-policy items are out of `pitch_ready`, any collapsed section shows a disclosed count + reason, and feedback this turn that changes policy was offered as a `brief.md` edit.
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

## Rubric

Use this rubric after the engine returns queued evidence. The engine exposes mechanical scores and `routing.queue_priority`; neither is a PR judgment.

The engine has two discovery lanes:

- `profile_relevance` - profile/topic/competitor queries. These catch highly relevant but sometimes minor stories.
- `major_news` - curated RSS/Atom feed items. These catch broader major news first, then require a stricter client-relevance judgment.

Do not treat a `major_news` item as pitchable because it is big. The client still needs standing and a journalist shape.

### Story Size

Use `story_size` to calibrate effort, not to approve a pitch. It is a deterministic media-attention proxy based on news-search publication metadata:

- log-scaled estimated monthly traffic
- domain authority
- coverage spread across independently surfaced domains

When authority metadata is missing for a recognized major outlet, the engine may use a low-confidence known-outlet fallback. When publication metadata is otherwise sparse, it may attach `story_size.attention_hint`. Treat the hint as a low-confidence keep-alive signal: it can justify review in the big-stories section, but it does **not** prove the story is widely covered. Label the uncertainty plainly.

`major` or `high` story size, or a `high`/`major` attention hint, means the opportunity may justify faster review. It does not compensate for stale timing, weak standing, or a bad journalist shape.

### Freshness

For recurring scheduled output, the LLM `story-origin-check` recovers the first-public timestamp and canonical coverage, then the Go CLI `origin-apply` computes the freshness gate. News-search `published_at` values are reliable evidence for article timestamps and should be used to find candidate originals, but they are not alone a same-story or first-publication judgment.

Before assigning `pitch_now`, `develop_angle`, or `monitor`, inspect `freshness_gate.computed_status`:

- `fresh` - eligible for normal judgment.
- `fresh_new_development` - eligible, but the angle must be about the new development, not the older background story.
- `stale` - reject as stale.
- `unverified_no_corroboration`, `unverified_boundary`, `unverified_no_timestamp`, or missing - reject for recurring scheduled output. Track the reason: `unverified_no_corroboration` is a worker/pipeline miss (the clock may be fine, just under-sourced — re-runnable), while `unverified_boundary`/`unverified_no_timestamp` reflect genuinely thin evidence.

Do not reset the clock because an aggregator, syndication partner, or secondary outlet republished an older article.

When citing the story, prefer `story_origin.canonical_coverage_url` when present. It should be the major or most authoritative same-story coverage, such as a primary source, wire, major publisher, or recognized trade, instead of the small pickup that triggered retrieval.

### Verdict Ladder

#### pitch_now

Use only when all are true:

- Evidence is fresh: usually `30min`, `4hr`, or `24hr`.
- The first public story clock is verified as inside the last 24 hours, or the new development is inside the last 24 hours.
- At least one credible news source exists, preferably `news_search`.
- The client has direct standing to comment.
- The client has a real spokesperson or direct domain authority.
- A specific reporter shape is obvious.
- No hard brand-safety block applies.

#### develop_angle

Use when the signal is real but needs framing:

- Fresh or still within the week.
- Client standing is plausible but not yet sharp.
- A journalist shape exists, but the angle needs work.
- Major-news lane items often belong here when they are important but the client angle is indirect.

Handoff: `angle-generator`.

#### monitor

Use when the signal is interesting but not pitch-ready:

- Single-source or weak cross-source confirmation.
- Early chatter without enough news confirmation.
- The client might have standing, but the angle is not clear yet.
- The signal may matter if it gains traction.

#### reject

Use when any core gate fails:

- stale
- freshness unverified in recurring scheduled output
- no client standing
- no plausible journalist shape
- off-beat
- already seen with no new development
- weak source quality

### Decay

Decay uses the verified first-public timestamp from `story-origin-check`. Engine `features.decay_bucket` is provisional when evidence comes from aggregators, syndication partners, secondary rewrites, or search results that have not yet been matched to the original/canonical story.

- `30min` - live/breaking. Only use for immediate comment if the client can respond now.
- `4hr` - same-cycle. Good for reactive comment.
- `24hr` - still fresh. Good for angle generation or same-day response.
- `week` - trend/context only. Do not call it breaking.
- `month` - usually not a newsjack unless paired with a new data point or fresh hook.
- `unknown` - do not pitch as timely without independent timestamp verification.

### Standing

Strong standing:

- The client operates directly in the affected market.
- The client has direct market exposure, technical expertise, or a named executive who can speak concretely.
- The signal names the client's category, customers, regulators, technology, or competitors.

Partial standing:

- The client has adjacent expertise but needs a narrower angle.
- The client can explain impact but not the core event.

Weak standing:

- The client merely sells into the broad category.
- The client wants to comment because the topic is popular.
- The client only has generic thought leadership.

For `major_news` lane signals, standing must explain the bridge from the public story to the client:

- same buyer being affected
- same regulator or policy surface
- named competitor or platform move
- client can explain a non-obvious operational effect

If the bridge is "this is about AI and the client uses AI," reject or monitor.

### Journalist Shape

A useful journalist shape names:

- exact beat
- outlet archetype
- why the beat cares now
- who should not receive it

Bad shapes:

- "business reporter"
- "AI journalist"
- "tech media"
- "industry press"

Good shapes:

- "enterprise AI reporter covering vendor compliance claims after regulator action"
- "cybersecurity trade reporter covering identity-risk fallout from new enforcement"
- "retail operations reporter covering labor-cost impact of a same-day policy change"

### Hard Blocks

Block signals built on:

- death
- violence
- disaster
- war
- abuse
- sexual violence
- missing people
- humanitarian crisis
- hate crime
- terror
- suicide

The only acceptable work around these topics is restrained expert commentary with direct public-interest standing. Promotional hooks are refused.

## Examples

### Pitch Now

Engine signal:

```json
{
  "id": "s1",
  "title": "FTC opens inquiry into AI compliance claims",
  "sources": ["news_search", "x"],
  "features": {
    "decay_bucket": "4hr",
    "source_count": 2,
    "seen_before": false,
    "profile_matches": ["AI compliance", "enterprise governance"],
    "safety_flags": []
  },
  "routing": {
    "lane": "profile_relevance",
    "queue_priority": 86.2,
    "demoted": false
  },
  "mechanical_scores": {
    "freshness": 1.0,
    "source_agreement": 0.78,
    "novelty": 1.0,
    "profile_match": 0.44,
    "source_quality": 0.825,
    "momentum": 0.21,
    "major_news": 0.0
  }
}
```

Skill output:

```json
{
  "signal_id": "s1",
  "signal_title": "FTC opens inquiry into AI compliance claims",
  "verdict": "pitch_now",
  "decay": {
    "stage": "4hr",
    "rationale": "The signal is same-cycle by verified first-public clock, not just the search-result timestamp."
  },
  "first_publication": {
    "status": "fresh",
    "surfaced_article_published_at": "2026-05-25T13:14:00Z",
    "first_public_at": "2026-05-25T13:10:00Z",
    "original_url": "https://www.ftc.gov/news-events/news/press-releases/example",
    "canonical_coverage_url": "https://www.reuters.com/legal/government/ftc-opens-inquiry-ai-compliance-claims-2026-05-25/",
    "canonical_coverage_source": "Reuters",
    "rationale": "The official FTC press release is the earliest verified public source and is inside the 24-hour cron window."
  },
  "why_newsjacking_worthy": "Regulator action creates a live need for explainers on AI compliance claims.",
  "client_standing": {
    "assessment": "strong",
    "rationale": "The client works directly in enterprise AI governance and can explain claim substantiation."
  },
  "journalist_shape": {
    "beat_description": "Enterprise AI reporter covering compliance and regulator scrutiny",
    "why_they_care_now": "They need sourced reaction while the inquiry is fresh.",
    "do_not_target": "General startup roundups or consumer AI reviewers"
  },
  "evidence_used": [
    {
      "source": "Reuters",
      "title": "FTC opens inquiry into AI compliance claims",
      "url": "https://www.reuters.com/legal/government/ftc-opens-inquiry-ai-compliance-claims-2026-05-25/"
    },
    {
      "source": "FTC",
      "title": "FTC opens inquiry into AI compliance claims",
      "url": "https://www.ftc.gov/news-events/news/press-releases/example"
    }
  ],
  "next_skill": "reactive-comment"
}
```

### Reject (stale syndication)

Engine signal: AOL article published today, canonical URL points to a BBC story from May 4 with no new development.

Verdict: `reject`

Reason: `stale`

`first_publication.status`: `stale`
