# Tracking plan — GB home energy AI-visibility panel

**Panel:** `panel-gb-home-energy-visibility-001` v0.1.0 · **Status `provisional_directional`** · Created 2026-07-25 · Next review **2026-08-26**
Prompts live in `panel_report.md` §4 (72 exact strings). This document is the measurement design only.

---

## 1. Estimands and denominators

Five active estimands, one deferred. **Each has its own denominator and none may be blended into a single score.**

| Estimand | Numerator | Denominator | Never counted in it |
| --- | --- | --- | --- |
`unaided_brand_presence` (primary) | valid observations whose answer text matches target brand/product terms | valid observations of **unaided** cells (core, rotating, sentinel) in one lane × surface × locale × wave | aided cells of any kind, control cells, invalid/refused runs
`competitive_mention_share` | target mention instances in eligible comparison/recommendation observations | all approved-supplier-list mention instances in the same eligible observations, same lane × surface × locale × wave | pooled unaided + competitor-aided results, control cells
`citation_presence` | retrieval-lane observations citing a target-owned domain | retrieval-lane observations **where citation metadata was exposed** | observations with no exposed citations (reported separately as unmeasurable, never as zeros), closed-model lane
`aided_brand_knowledge` | valid B0 observations with ≥1 correct factual element under a fixed rubric built from source-001/002/003; material errors counted separately | valid **target-aided (B0)** observations only | every unaided and category-aided observation
`answer_framing` | observations assigned each code from a fixed codebook (cost-first, risk/caveat-first, service-first, low-carbon-first, official-guidance-first, refuses-to-recommend) | valid observations in the same stratum × lane × surface × locale × wave | control cells
`campaign_response` | — | **deferred, no denominator** | — (declared only so nobody retro-fits it to an uncontrolled before/after)

**Invalid runs** (refusal, error, truncation, filtering, rate-limit) are counted and reported by reason. A refusal is not an absence of the target and is never dropped from a denominator silently.

**What none of these prove:** market share, audience reach, consumer awareness, preference, switching behaviour, referral traffic, or revenue. Every number is conditional on this panel, this wording, these surfaces and this wave.

---

## 2. Selected partitions

| Partition | Cells | Prompts | Role in the design |
| --- | ---: | ---: | --- |
core | 21 | 42 | the stable measurement backbone; two variants each |
rotating | 8 | 8 | discovery and thin-evidence intents; never used for trend claims |
sentinel | 5 | 10 | frozen wording, repeat-heavy, used for drift and the variance pilot |
aided | 10 | 10 | 4 target-aided (B0) + 5 category-aided + 1 competitor-aided; separate denominators |
control | 2 | 2 | drift controls, excluded from every estimand denominator |

Quarantined and therefore **outside all reporting**: cell-017 (grade-D expansion) and cell-030 (unverified B5 event). Sentinel cells 041 and 042 deliberately overlap core cells 002 and 004; the overlap is recorded so job-level rollups do not double-count.

---

## 3. Surfaces and lanes

| Surface | Lane | Configuration |
| --- | --- | --- |
surf-api-a, surf-api-b | `closed_model` | two frontier models via API, no tools, no retrieval, no history, empty system prompt, temperature 0 where exposed |
surf-answer-engine, surf-assistant-search | `retrieval` | search enabled, citations captured |
surf-consumer-clean, surf-ai-overview | `consumer_surface` | clean account archetype, GB locale, desktop, no personalisation |
— | `campaign_experiment` | **inactive**: no campaign, no pre-registration, no treatment/control definition |

**API and consumer-surface observations never share a rollup.** Consumer surfaces are unstable, personalised products; their results are reported beside API lanes, never averaged with them. 34 of 46 cells are consumer-surface eligible; cell-030 is retrieval-only.

---

## 4. Repetitions

Core and aided: 3 repeats per variant per wave. Sentinel: 6 per variant across ≥2 time blocks. Rotating: 2. Control: 3.

Variants and repeats are **nested observations inside a canonical cell**, never independent buyers. The cell is the sampling and clustering unit. A full wave across six surfaces is roughly 1,300 observations plus ~180 for sentinel deepening; a cheaper first wave can run one surface per lane for about 430.

---

## 5. Fresh-session and retrieval-state controls

- Fresh session for every observation. Memory off, custom instructions empty, no prior turns except the fixed second turn of the three scripted multi-turn cells.
- Multi-turn rule: **turn 2 text is fixed** and must not be adapted to the model's turn-1 answer, or the observation stops being comparable.
- Record per observation: provider, model as requested, model version exactly as exposed (or `unexposed` — never guessed), temperature/top-p/max-tokens, system-prompt ID and hash, account and personalisation state, device, locale requested and region observed.
- Retrieval state per observation: policy (`required` / `allowed` / `disabled` / `unavailable`), whether retrieval **actually ran**, generated queries when exposed, live-or-cached state, and full citation metadata with domain classification.
- Hash the exact prompt text, the request configuration, the response, and the citation payload. Every hash field in this run's artifacts is currently `null` because the runtime had no hashing tool; the harness must compute them at run time.

---

## 6. Weights

Two components, stored and reported **separately**:

- **Exposure — equal within declared strata** (band × job × lane × surface × locale), normalised within band × lane × surface. Factor list empty, confidence low. **No credible audience, intent, locale or surface prevalence evidence exists for these jobs.** Equal weighting is a convention that makes cells comparable; it is not a prevalence estimate. Results must never be labelled market share, audience reach, awareness or share of users.
- **Priority — withheld.** Strategic priority weights are a Gate 4 human decision. No default has been substituted, and no priority-weighted figure may be published until an approver sets and signs the factors.

Neither component may depend on baseline visibility or campaign performance. Warn whenever one weight dominates a rollup or effective sample size collapses.

---

## 7. Uncertainty

- **Wilson intervals** for a single unweighted binary proportion within one lane, one aided status and one surface.
- **Stratified cluster bootstrap by canonical cell** for anything pooled or weighted.
- Pair unchanged cells across waves; a version comparison must show overlap-only change *and* both full-version levels.
- **Reporting floor:** no percentage for any subgroup with fewer than 20–30 distinct cells. With 21 core plus 5 sentinel unaided cells, only the whole unaided partition approaches that floor, so every job-level, act-level or role-level cut is reported as **counts and example answers**, not a leaderboard.
- Intervals quantify run-to-run and sampling variation conditional on this panel. They do not repair coverage bias and license no causal or population claim.

---

## 8. Variance pilot — required before any percentage is published

**Status: not run.** Design: 16 diverse cells (cell-003, 004, 007, 011, 013, 015, 019, 023, 025, 027, 033, 037, 041, 042, 043, 044) × 2 variants × 6 repeats, across ≥2 time blocks and ≥2 surfaces per active lane.

Estimate: between-cell, within-cell run-to-run, variant-wording, day/time-block, model/surface, and invalid/parser variance. Decision rule: high within-cell correlation → more unique cells; high run-to-run variance → more repeats. Re-allocate before wave 2 and record it as a new version. Until the pilot runs, report counts, answers and qualitative framing only.

---

## 9. Randomization

Seed `20260725`. Deterministic shuffle of the candidate × surface × repeat list under that seed. Time blocks assigned round-robin — weekday 09:00–12:00, weekday 19:00–22:00, weekend 12:00–15:00 Europe/London — so no cell is systematically run at one hour. Seed, order index and time block are recorded per observation.

---

## 10. Cadence, refresh and version policy

- **Monthly** evidence intake; **quarterly** panel review targeting roughly 70–80% core, 15–25% rotating, 5–10% sentinel/control; **annual** charter approval.
- **Event triggers:** quarterly price-cap publication, tariff or product change by the target or a named comparator, grant or regulatory change, model or surface version change, new locale or market.
- **B5 decay is explicit:** cell-007 must be restated when the 1 Oct 2026 cap period is published (due by 2026-08-26); cell-030 must be verified against a primary government source or deleted by 2026-08-15. A closed-model lag behind a dated event is a finding, not an error.
- **Versioning:** changing core cells, weights, metric definitions or surface mix creates a new semantic version with an overlap bridge on unchanged cells. Frozen versions are never overwritten; `panel_change_ledger.json` is append-only.
- **Campaign claims:** a before/after increase is not attribution. The evidence ladder (1 panel mention/framing/citation → 2 source or referral traffic → 3 self-reported discovery → 4 qualified lead → 5 incremental outcome from an experiment) is reported rung by rung, and rung 1 is never renamed revenue attribution.

---

## 11. Approvals

| Gate | Owns | Status |
| --- | --- | --- |
1 | facts, ICPs, exclusions, permissions | **pending** |
2 | jobs, language, roles, locales, priority | **pending** |
3 | core/aided/campaign partitions, disputed QA | **pending** — 5 open items (tracker exception; competitor choice in prompt-008a; the "energy supplier" boundary rule; verify-or-delete cell-030; paraphrase-variant wording in core) |
4 | weights, limitations, cadence, claims, version freeze | **pending** — equal exposure weighting unapproved, priority weights undefined, no pilot |

Every gate is resumable from the artifacts: approve or reject IDs in place, and the downstream atom re-runs from that decision.

---

## 12. Limitations

1. **Non-probability instrument.** It measures how selected surfaces answer 46 constructed intents. It does not measure what GB households ask, believe or do.
2. **No grade-A evidence.** No first-party corpus was supplied; all buyer language is public-forum or search proxy. A forum post is evidence from that source, not market prevalence, and no generated prompt has an observed frequency.
3. **Paraphrase wording is agent-generated** (grade-C provenance) inside grade-B and grade-C cells — a Gate 3 item.
4. **Blinding was procedural, not architectural.** No fresh subagent was available, so unaided generation ran in the same context that had read market pages, working only from the sanitised brief. Deterministic term scanning over 62 candidates found zero leaks, but the isolation guarantee is weaker than a separate process would give.
5. **One dated B5 event only**, expiring 2026-08-26. The second candidate event is unverified and quarantined rather than treated as a trend.
6. **Unresearched populations:** prepayment, arrears and affordability-support households, renters, flats, park homes, business buyers, and Welsh-language users. Their absence is unfinished research, not evidence of no need — and the affordability group is where a wrong AI answer does the most harm.
7. **Five sources were read via search-result summaries** rather than direct retrieval, and two independent sources have fieldwork or post dates 9–28 months old.
8. **Every hash is null** in this run; provenance integrity is not yet cryptographically checkable.
9. **No causal claim is supported.** No experiment, no pre-registration, no campaign lane.
10. **Nothing here was selected on performance.** No observed AI answer, ranking, mention, citation or content gap influenced any cell, wording, partition or weight — and none may, or the panel starts measuring its own reflection.
