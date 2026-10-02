# Jev as a Coarse-Filter Engine

Plan for wiring Jev (TypeSafe AI's typed-decision model) into the detector pipeline as an optional engine for the `relevance-coarse-filter` pass. The News Desk Dealer demo proved the shape: 384 headlines judged in 24.9 s for $0.19, against 4 of 384 for $0.77 on Opus 5. This plan moves that from a demo into the real pipeline without changing any artifact contract.

Status: implemented in `feat/jev-coarse-filter`: CLI command, credentials, doctor, tests, skill and routing updates, and a first live agreement eval (results below).

## Goal

Today the coarse pass is an LLM worker pass: each worker loads `skills/relevance-coarse-filter/SKILL.md`, judges a chunk of 8–12 signals, and returns a `decisions` array that `newsjack filter-apply` consumes. It is the cheapest stage we have, but it is still the widest stage by volume, and it is bounded by harness worker fanout (max ~8 workers, so runs over 80 signals have to tighten `--limit`).

With Jev the coarse pass becomes:

```bash
newsjack coarse-filter --engine jev --candidates candidates.json --output coarse_relevance_decisions.json
newsjack filter-apply --candidates candidates.json --decisions coarse_relevance_decisions.json --include keep --include monitor_only --output relevant_candidates.json
```

Everything downstream (`filter-apply` guards, `cluster`, story-origin, `origin-apply`, triage, report) is untouched. The LLM worker path stays the default and the fallback; Jev is opt-in when a key is present.

## Where it lives

In the Go CLI, as a new `coarse-filter` command. This follows the CLI Unix Principle in `AGENTS.md`:

- It is the same testable operation for every user regardless of client, voice, or outlet: JSON in, JSON out, one external API call per signal.
- It is plumbing (credentials, HTTP, concurrency, retries, cost accounting), which the CLI already owns for Medialyst and X.
- Jev cannot read the skill prompt. It needs the rubric decomposed into typed questions, and that decomposition has to be code somewhere. The CLI is the only runtime-agnostic place; a Node or Python script inside the skill folder would break "public skills are runtime-agnostic" and would not be installed by `newsjack install`.

What stays in the skill: the rubric wording and the judgment doctrine. `relevance-coarse-filter/SKILL.md` remains the source of truth for what "junk" means. The Jev question set is a translation of that rubric, and it is shipped as a JSON file the CLI embeds, so wording can be reviewed as data and overridden per run with `--questions <path>` for experiments without a release.

## Command surface

```text
newsjack coarse-filter --engine jev --candidates F [--output F] [--profile F]
                       [--concurrency 8] [--questions F] [--model jev-latest]
                       [--print-questions] [--dry-run]
```

- `--engine` accepts `jev` only for now. The flag exists so an OpenRouter/Haiku engine can be added later using the same question set and JSON-schema translation the demo already has in `openrouter.ts`.
- `--candidates` is the detector `candidates.json`. The profile is read from `monitor.profile` inside it, same as `filter-apply`; `--profile` overrides.
- `--output` defaults to stdout, matching `filter-apply`.
- `--print-questions` dumps the embedded question set so a skill or a human can audit it.
- `--dry-run` builds every request and prints token estimates and projected cost without calling the API.
- Exit code is non-zero if more than 20% of signals fail after retries, so the orchestrating skill can fall back to LLM workers instead of running the expensive pass on a half-judged pool.

## Credentials and doctor

- Env var `TYPESAFE_API_KEY`, resolved the same way as `X_BEARER_TOKEN`: process env, then `.env` walk-up from cwd, then `~/.newsjack/.env`.
- `newsjack auth set-typesafe --key <key>` writes it to `~/.newsjack/.env` via the existing `writeNewsjackEnv`. No new credential file format.
- `newsjack doctor` gains a `TypeSafe (Jev)` row. Not configured is a note, not a warning: the pipeline works without it.
- Base URL override `NEWSJACK_TYPESAFE_BASE_URL` for tests and for a self-hosted proxy, defaulting to `https://api.typesafe.ai`.

## Request shape

One call per signal, six questions. This mirrors the skill's "judge one signal at a time" instruction and keeps every answer keyed to one signal id, so a malformed answer only costs one signal. Batching several signals per call is a later optimization; at $0.042 per million input tokens, repeating the profile block per signal costs about five cents per 400-signal run.

`state`:

```json
{
  "client": {
    "company": "...", "description": "...", "website": "...",
    "topics": [...], "competitors": [...], "standing": [...],
    "search_terms": [...]
  },
  "signal": {
    "id": "...", "title": "...", "query": "...", "lane": "...",
    "sources": [...], "profile_matches": [...], "safety_flags": [...],
    "story_size": {"band": "...", "attention_hint": {...}},
    "evidence": [{"title", "url", "container", "excerpt", "published_at", "publication_type", "x_author_followers"}]
  }
}
```

Evidence is capped at the first five items and excerpts at 600 characters, matching what `coarseRecallText` already treats as the judgable surface.

`questions` (embedded default, criteria text lifted from the skill rubric):

| key | type | purpose |
| --- | --- | --- |
| `decision` | choice: `keep`, `monitor_only`, `reject` | The decision itself. Criteria restate the skill's three definitions plus "when in doubt, keep". |
| `reason` | choice over the 13 allowed reasons | Reason label. Jev's per-option probabilities let us pick the best reason that is consistent with the decision. |
| `is_news` | noul | "This is a reported news item, not a docs, product, SEO, or evergreen page." Backs `not_news` / `owned_docs_or_product_page` / `seo_landing_page`. |
| `profile_bridge` | noul | "The client, a named competitor, a profile topic, a standing term, or a direct synonym appears in the title, excerpt, or evidence." Backs the profile-match rule. |
| `promotional` | noul | "This is a press release, newswire item, or vendor-authored promotional piece." Backs `competitor_or_promotional`. |
| `safety_risk` | noul | "Attaching a brand to this story would ride on tragedy, hate, crime, or an active crisis." Backs `safety_risk`. |

## Post-rules

Deterministic, applied in the CLI before writing each decision. They exist so the typed answers cannot contradict the skill's hard rules.

1. Reason consistency. `reject` must carry a junk reason; `keep` must carry `relevant_news` or `plausible_client_bridge`; `monitor_only` may carry any. If the top reason is inconsistent, take the highest-probability reason from the consistent subset.
2. Profile-bridge floor. If `profile_bridge` ≥ 0.5 or the signal has non-empty `profile_matches`, a `reject` becomes `monitor_only` and reason `no_profile_bridge` is not allowed. This duplicates the `filter-apply` profile-match guard on purpose: the decision file should be honest on its own, and the guard remains as a backstop.
3. Safety floor. `safety_risk` ≥ 0.7 forces reason `safety_risk` and decision no higher than `monitor_only`. Hard-safety flags are still applied downstream.
4. Promotional floor. `promotional` ≥ 0.7 and decision `reject` becomes `monitor_only` with reason `competitor_or_promotional`, per the skill's "don't reject on this basis".
5. Uncertainty floor. If the `decision` confidence is below 0.55 and the answer is `reject`, upgrade to `monitor_only` and set `confidence: low`. Low-confidence rejects are exactly the false negatives the skill says are expensive.
6. Big-story recall is left to `filter-apply` (`big_story_recall`), which already upgrades any `reject` on a high/major signal. No need to duplicate.
7. Confidence mapping: Jev `confidence` ≥ 0.8 → `high`, ≥ 0.6 → `medium`, else `low`.

## Output contract

The file is the existing `coarse_relevance_decisions.json` shape, so `filter-apply` and `run-summary` read it unchanged. New top-level `engine` block for disclosure:

```json
{
  "version": 1,
  "generated_at": "...",
  "engine": {
    "name": "jev", "model": "jev-latest", "questions_sha256": "...",
    "signals": 384, "calls": 384, "failures": 2,
    "input_tokens": 1180000, "output_tokens": 0, "est_cost_usd": 0.05,
    "elapsed_ms": 24900, "concurrency": 8
  },
  "decisions": [
    {
      "signal_id": "...",
      "decision": "reject",
      "reason": "keyword_collision",
      "rationale": "Jev: reject (0.83); keyword_collision (0.71). is_news 0.92, profile_bridge 0.08, promotional 0.10, safety 0.02.",
      "confidence": "high",
      "evidence_urls": ["..."],
      "relevance_basis": "No profile entity matched; query 'OpenAI Operator' collided with a Meta stock note.",
      "engine": "jev",
      "answers": { "decision": {"choice": "reject", "probabilities": {...}, "confidence": 0.83}, "...": "..." },
      "post_rules": ["uncertainty_floor"]
    }
  ]
}
```

- `rationale` and `relevance_basis` are synthesized deterministically from the answers. Jev returns no prose, and the report must not pretend otherwise: the `rationale` always starts with `Jev:` so a reader can tell.
- `answers` keeps the raw typed output so an eval can re-score without re-calling the API.
- A failed signal (after 4 attempts with backoff on 429/5xx) gets `decision: monitor_only`, `reason: plausible_client_bridge`, `confidence: low`, `rationale: "Jev call failed: <error>; kept for review."`, and is counted in `engine.failures`. Recall-preserving, and visible.

## Skill and doc changes

- `skills/relevance-coarse-filter/SKILL.md`: add a short "Engines" section. Default is a low-cost LLM worker applying this rubric. If `newsjack doctor` reports TypeSafe configured, run `newsjack coarse-filter --engine jev` instead; the rubric in this file is what the embedded question set encodes, and `--print-questions` shows the translation. Skill stays runtime-agnostic and does not list flags.
- `skills/newsjack-detector/references/harness-routing.md`: add path 0 above the three existing paths: "Jev engine path. If `newsjack doctor` shows TypeSafe configured, run the CLI coarse-filter and skip worker fanout for this pass. Disclose `coarse pass: jev` in the final response." Story-origin is unchanged; it needs retrieval and prose, which Jev does not do.
- `skills/newsjack-detector/SKILL.md` step 2 and step 6: mention the engine choice and require the disclosure line to name the engine (`jev`, `low-cost worker`, or `current-model fallback`).
- `apps/cli/cmd/newsjack/usage.go`: register `coarse-filter` and `auth set-typesafe`.
- `docs/example-run.md`: no change until a real run is captured with Jev.
- `demos/news-desk-dealer/README.md`: one line pointing at the CLI command, so demo readers know the real pipeline has it.

## Eval before shipping

The fixture runs under `fixtures/newsjack-detector-agent/runs/20260603T175505Z_*` already hold paired `coarse_chunk.N.json` inputs and `coarse_decisions.N.json` Haiku outputs for simular, slite, and localfalcon, plus older single-file runs for clearnym and bluebottle. That is a labeled agreement set, roughly 200 signals.

Add `eval/2026-09-jev-coarse-agreement/`:

- `run.sh` calls `newsjack coarse-filter --engine jev` on each chunk (chunks are the `{profile_context, signals}` shape, so the command accepts that as an alternative to full `candidates.json`).
- `compare.py` reports agreement on `decision`, agreement on survive-vs-drop (keep+monitor_only vs reject), and the two numbers that matter most: recall misses (LLM kept, Jev rejected) listed one per line with title, and junk let through (LLM rejected, Jev kept).
- Acceptance bar: zero recall misses on any signal the LLM marked `keep` with `confidence: high`, and survive-vs-drop agreement at or above 85%. Below that, tune criteria wording in the question file, not the post-rules.

Also record a live cost and latency line for the doc, the same way the demo README does.

## Tests

- `coarse_filter_test.go`: table tests for each post-rule using hand-built answer payloads; request construction from a fixture candidate (profile block present, evidence capped); output file passes `applyDecisions` with the fixture candidates; failure threshold exit code; `--dry-run` produces no network calls.
- HTTP path against an `httptest.Server` via `NEWSJACK_TYPESAFE_BASE_URL`, including one 429 then success, and one malformed answer that falls to the failure path.
- `doctor_test.go` and `auth_test.go` extensions for the new key.
- Run `go test ./...` from `apps/cli` and the mock fixture smoke test to confirm the LLM path is unaffected.

## Eval results, 2026-09-18

First live run of `eval/jev-coarse-agreement/` over the 2026-06-03 simular, slite, and localfalcon fixture pools, committed under `eval/jev-coarse-agreement/runs/20260918T162324Z/`. Model reported by the API: `jev-1.13.0`.

| measure | value |
| --- | --- |
| signals | 176 |
| calls / failures | 176 / 0 |
| wall clock, 8 workers | 8.3 s |
| input tokens / est. cost | 320k / $0.013 |
| latency p50 per call | about 260 ms |
| survive-vs-drop agreement with the Haiku worker | 76.7% |
| exact decision agreement | 60.2% |
| recall misses (worker kept, Jev rejected) | 3, none on a high-confidence `keep` |
| junk let through (worker rejected, Jev kept) | 38 |

Reading:

- The recall bar passes. The three misses are all worker `monitor_only` calls with low or medium confidence, and one of them (the Apple CEO succession story on the Slite profile) is a `high`-band story that `filter-apply`'s big-story guard upgrades regardless.
- The 85% survive-vs-drop bar fails, but every disagreement points in the recall-safe direction: Jev keeps or monitors things the worker rejected. Of the 38, Jev natively chose `monitor_only` or `keep` on 24; the uncertainty floor upgraded 9 rejects; the promotional and bridge floors handled the rest. Several of the "junk" items are Anthropic and Claude news on the Simular profile, where Jev's `keep` is defensible given `Anthropic computer use` is a named competitor.
- Jev's `confidence` field runs low: median 0.53 against a median top-option probability of 0.68, so about half of all decisions land below the 0.55 uncertainty threshold and most carry a `low` label. That is a calibration question to settle with TypeSafe's confidence docs before tuning the threshold; the plan's own rule was to tune question wording rather than post-rules, and nothing here argues for loosening recall.
- Cost and speed match the demo's numbers: the whole coarse pass for a 176-signal pool costs about a cent and finishes in under ten seconds, against roughly 8 Haiku worker calls and a few minutes of harness fanout.

Practical consequence: the expensive story-origin pass sees more `monitor_only` items than it would after the worker pass. That is the intended trade at this price point, and clustering plus the freshness gate still bound the cost downstream.

## Phases

1. CLI command with embedded questions, post-rules, output contract, credentials, doctor, tests. Done.
2. Agreement eval on fixture chunks; capture cost and latency. First run done; criteria tuning is open.
3. Skill and routing doc updates; disclosure line in the detector report. Done.
4. Later options, not in this plan: multi-signal batching per call; `--engine openrouter` reusing the same question set with JSON schema; a hybrid mode that sends only Jev's low-confidence decisions to an LLM worker instead of flooring them to `monitor_only`.

## Open questions

- Waitlist. TypeSafe was early-access on 2026-09-18. The command should fail with a clear "TypeSafe key not configured; see https://typesafe.ai" rather than a bare HTTP error, and the skills must keep working without it.
- Chunk input. Accepting `{profile_context, signals}` chunk files as well as `candidates.json` is convenient for evals and for harnesses that already chunk, but it means the profile block has two possible shapes. Proposal: accept both, normalize to the `state.client` block above, and document it in command help.
- Reason granularity. Thirteen options in one `choice` may spread probability thin. If the eval shows reason agreement well below decision agreement, split into a `junk_type` choice (10 options) and derive `relevant_news` vs `plausible_client_bridge` from `profile_bridge`.
