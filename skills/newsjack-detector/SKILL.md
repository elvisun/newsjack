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
- coarse-relevance application via `newsjack filter-apply`, plus a recall guard that upgrades `reject/no_profile_bridge` to `monitor_only` when detector/profile evidence already matched the client, a competitor, or a profile term
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

Use `--emit brief` for a human scan, default JSON for skill judgment, `--mock` for local verification without credentials. Full flag/source/env reference: `references/engine-cli.md`.

For each queued signal, inspect title, sources, evidence URLs, age, `routing.lane`, `mechanical_scores` (`major_news`, `novelty`, `source_agreement`), profile matches, and safety flags. For `x` evidence inspect `x_signal_type`, `x_social_signals`, `x_author_followers`, `x_query_counts`; treat lone low-reach posts as noise. A high `major_news` means the story is broadly important, **not** that the client has standing. Treat engine age/decay as provisional until `story-origin-check` verifies the first-public clock. Then apply `rubric.md` and the **Output Format**.

## Canonical Pipeline

The artifact contract is the source of truth. Write all artifacts to a timestamped run folder:

```text
RUN_DIR/
  candidates.json              # 1. detector output
  coarse_relevance_decisions.json   # 2. coarse pass
  relevant_candidates.json     # 3. filter-apply
  origin_findings.json         # 4. story-origin pass
  targeted_candidates.json     # 5. origin-apply (freshness authority)
  final_report.md              # 7. compiled report
  run.md                       # 8. rerendered — THE human-facing artifact
  detector.stderr.log  commands.log  summary.json
```

Only `run.md` is human-facing; the rest are provenance.

1. **Run the detector and save candidates:**

   ```bash
   ~/.newsjack/bin/newsjack detector run "QUERY" --profile profile.json --sources news_search,x --lookback-days 1 --depth quick --limit 80 --min-queue-priority 40 --min-major-news 0.55 --emit json > candidates.json
   ```

2. **Coarse relevance pass** → `coarse_relevance_decisions.json`. High-recall junk removal only — no ranking, angles, dates, or pitch decisions. Each worker loads `skills/relevance-coarse-filter/SKILL.md` and applies it to its assigned signals; merge every worker's output into one `decisions` array. For model/worker routing and chunking, see `references/harness-routing.md`.

3. **Apply coarse decisions:**

   ```bash
   ~/.newsjack/bin/newsjack filter-apply --candidates candidates.json --decisions coarse_relevance_decisions.json --include keep --include monitor_only --output relevant_candidates.json
   ```

4. **Story-origin pass** on `relevant_candidates.json` → `origin_findings.json`. Each worker loads `skills/story-origin-check/SKILL.md` and applies it per signal: decide same-story vs material-new-development, recover `first_public_at`, `original_url`, and canonical major coverage. It must **not** compute `fresh`/`stale`. Merge the per-signal results into one `findings` array, keyed by `signal_id`. The story-origin pass needs retrieval — see `references/harness-routing.md` for what to do when a low-cost worker cannot search or open pages.

5. **Apply the deterministic freshness gate:**

   ```bash
   ~/.newsjack/bin/newsjack origin-apply --candidates relevant_candidates.json --origins origin_findings.json --window-hours 24 --output targeted_candidates.json
   ```

   The Go CLI is the freshness authority — it computes `freshness_gate.computed_status` from the run timestamp and cutoff. If an LLM labels May 8 fresh for a May 25 run, `origin-apply` marks it stale.

6. **Angle generation** on the high-priority fresh candidates in `targeted_candidates.json`. `angle-generator` is the atomic fit step: a candidate is useful only if it yields at least one honest, journalist-shaped angle. Reject/downgrade candidates that return zero viable angles, duplicate/slop angles, or no specific journalist shape.

7. **Compile `final_report.md`** — story-first and skimmable:
   - `## Top News Today`: current stories in priority order — story size, link, fit status, and three suggested `angle-generator` angles for each non-rejected story.
   - `## Top Positioning Angles`: the strongest ways the client can enter, each anchored to specific news item(s) and story size.
   - `## Watch / Not A Fit`: relevant-but-not-ready and rejected items with a plain reason.

   Links must be clickable Markdown, not backticked or bare URLs. Do not present mechanical rank as a final fit verdict; do not mix story headlines and angle headlines without labeling which is which.

8. **Rerender the run report:**

   ```bash
   ~/.newsjack/bin/newsjack summarize-run candidates.json --output summary.json --markdown run.md
   ```

   Inside a timestamped folder, pass full paths for `candidates.json`, `summary.json`, and `run.md`. `run.md` renders `final_report.md` plus a compact candidate scan.

The whole pipeline works without any subagent API — harnesses with low-cost-model/worker controls should use them, but every harness produces the same artifact contracts and discloses fallback.

## Freshness Gate

For recurring scheduled output, a signal is not surfaceable until its Go-computed `freshness_gate.computed_status` is verified. News-search `published_at` values are good article-publication evidence for recovering originals, but they alone never decide same-story status or first publication — that is the `story-origin-check` atom's job.

Recurring output rules:

- Surface only `fresh` or `fresh_new_development`. Reject `stale`. Reject when the first-public timestamp cannot be verified inside the last 24 hours, with reason `freshness_unverified`.
- Do **not** reset the clock for AOL, Yahoo, MSN, Apple News, partner syndication, wire pickup, SEO rewrites, or "published today" pages whose canonical/source story is older.
- A newer article restarts the clock only if it adds a concrete new public fact: official action, filing, statement, data/report publication, material company update, new local impact, or another independently coverable development.
- Prefer `story_origin.canonical_coverage_url` as the report's main link — the major/most authoritative same-story coverage, not the random pickup that triggered retrieval.

`origin-apply` attaches `story_origin` and the deterministic `freshness_gate` to selected and rejected signals. If the first-public timestamp can't be verified, write `first_public_at: null` and explain the gap; `origin-apply` computes `freshness_unverified`.

## Handoff

- Breaking / same-day sourced comment → `reactive-comment`
- Needs story framing → `angle-generator`
- Named journalist check → `journalist-fit-check`
- Draft critique → `meanest-editor`

## Completion Checklist

Before reporting the run complete:

- `coarse_relevance_decisions.json` has exactly one decision per emitted candidate (unless `--allow-missing`).
- `origin_findings.json` has exactly one finding per relevant candidate (unless `--allow-missing`).
- `targeted_candidates.json` was produced by `origin-apply`.
- `final_report.md` was written from `targeted_candidates.json`, not raw `candidates.json`.
- `run.md` was rerendered after `final_report.md` existed.
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
        "status": "stale | freshness_unverified | null",
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
- Allowed rejection reasons: `stale`, `freshness_unverified`, `single_source`, `no_client_standing`, `no_journalist_shape`, `off_beat`, `already_seen`, `weak_signal`, `no_viable_angle`.
- Allowed brand-safety block reasons: `tragedy_or_human_suffering`, `client_exclusion`, `regulated_claim_risk`, `fabrication_risk`.
