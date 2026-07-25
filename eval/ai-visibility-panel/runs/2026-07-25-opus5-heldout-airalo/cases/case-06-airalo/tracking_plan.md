# Tracking plan — travel eSIM marketplace AI visibility panel

**Panel:** `panel-travel-esim-marketplace` v0.1.0
**Status:** `provisional_directional` — not frozen, not representative, not significant, not causal
**Built:** 2026-07-25T20:00:00Z from one public URL plus a one-line description
**Next review:** 2026-10-25

This plan explains how the panel would be run. Every number in it is a design parameter, not a result. Nothing has been measured yet.

---

## 1. Estimands and denominators

Five estimands are declared. Each has its own denominator, and denominators never mix aided statuses or lanes.

| Estimand | Numerator | Denominator | Explicitly does not prove |
| --- | --- | --- | --- |
| `unaided_brand_presence` | Valid answers mentioning the target brand or a target-owned product name at least once | Valid observations where partition ∈ {core, sentinel} **and** `aided_status = unaided`, within one lane × surface × locale × wave | Market share, awareness, reach, revenue |
| `competitive_mention_share` | Answers naming the target, counted against all supplier brands named in those same answers | Valid observations where partition = aided **and** `aided_status = category_aided`, within one lane × surface × locale × wave | Market share or commercial performance |
| `citation_presence` | Retrieval-lane answers exposing ≥1 citation resolving to a target-owned domain | Valid retrieval-lane observations where `retrieval_used = true`, within one partition × aided_status × surface × locale × wave | Referral traffic, clicks, influence on the text |
| `aided_brand_knowledge` | Answers stating target attributes consistent with dated evidence on destination, device and plan constraints | Valid observations where partition = aided **and** `aided_status = target_aided` | That buyers believe the same, or that the stated facts are currently true |
| `answer_framing` | Answers whose structure matches the cell's `expected_answer_kind` | Valid observations within one canonical cell × lane × surface × locale × wave | Buyer preference or persuasion |

**Separation rules that are not negotiable.**

- `unaided`, `category_aided`, `competitor_aided` and `target_aided` observations live in four different denominators. The two competitor-aided cells (`cell-029`, `cell-030`) are reported on their own, never folded into `competitive_mention_share`.
- `closed_model`, `retrieval` and `consumer_surface` are three different denominators. API and consumer-surface results are never rolled up together.
- The off-category control cell (`cell-025`) has its own denominator and never enters `unaided_brand_presence`.
- Invalid runs (refusal, empty, truncated, parse failure) are excluded from both numerator and denominator, and the invalid count is published next to every rate.

`campaign_response` is **not** an estimand in this version. No campaign was supplied, so no campaign lane, treatment set, matched control or pre-registration exists.

---

## 2. Selected partitions

32 canonical cells and 61 prompt variants are selected. Two further cells are held out.

| Partition | Cells | Aided status | Candidates | Purpose |
| --- | ---: | --- | ---: | --- |
| `core` | 16 (`cell-001`…`cell-016`) | unaided | 32 | Primary unaided measurement across B1, B3 and B4 |
| `sentinel` | 4 (`cell-017`…`cell-020`) | unaided | 8 | Stable evergreen wording; carried unchanged across versions as the overlap bridge; variance pilot backbone |
| `rotating` | 2 (`cell-021`, `cell-023`) | unaided | 2 | Grade-C discovery: the B5 trend cell and the identity-registration cell |
| `control` | 1 (`cell-025`) | unaided | 1 | Off-category matched control for unprompted category/brand injection |
| `aided` | 9 (`cell-026`…`cell-034`) | 3 target-aided, 2 competitor-aided, 4 category-aided | 18 | Aided knowledge, competitor framing and category landscape |

**Held out of the panel.**

- `cell-022` (B5, multi-host-country event travel) — the only candidate was quarantined because the anchoring evidence describes an event that concluded on 2026-07-19, six days before this run. A B5 cell needs fresh dated evidence, so this is recorded as a gap rather than presented as a live trend.
- `cell-024` (en-CA supplier selection) — both candidates quarantined: the locale has no market-competent review and only company-asserted CAD price rendering distinguishes it. en-CA is therefore **unmeasured**.

---

## 3. Surfaces and lanes

| Surface ID | Lane | Configuration |
| --- | --- | --- |
| `surface-closed-a` | `closed_model` | API, no tools/retrieval/history, fixed minimal system prompt, fixed sampling, fresh session per observation |
| `surface-closed-b` | `closed_model` | Second provider, same policy, so provider effects are separable from panel effects |
| `surface-retrieval-a` | `retrieval` | Retrieval allowed; record `retrieval_used`, generated queries when exposed, live/cached state and all exposed citation metadata |
| `surface-consumer-clean` | `consumer_surface` | Logged-out or declared clean archetype, declared device, history cleared; **sentinel cells only** in this version because of manual collection cost (waiver W-006) |

Concrete providers are deliberately unassigned. Naming them is a Gate 4 decision (`pend-004`), because the choice determines the configuration hash and the drift policy.

`cell-021` (B5) is **retrieval-only**. A closed model cannot hold evidence that is fresh by construction, so pooling it with retrieval would measure staleness rather than visibility.

---

## 4. Repetitions

| Partition | Repeats per candidate per surface |
| --- | ---: |
| core | 3 |
| sentinel | 6 |
| rotating | 3 |
| control | 3 |
| aided | 3 |

**Planned volume per wave: 639 observations** — `prompt-021a` retrieval-only (1 surface × 3 = 3), 8 sentinel candidates (3 API surfaces × 6 = 144, plus consumer surface × 3 = 24), remaining 52 candidates (3 API surfaces × 3 = 468). Charter budget is set at 700 to leave retry headroom. Three waves per quarter.

These repeat counts are tier defaults. They are **not** derived from a variance pilot and must be revisited once one runs.

---

## 5. Fresh-session and retrieval-state controls

- Every observation opens a **fresh session**. No conversation carry-over, no memory, no stored personalisation.
- The only permitted continuation is between the two scripted turns of a multi-turn candidate (`prompt-013a`, `prompt-013b`, `prompt-028a`, `prompt-028b`), which run inside one session and are recorded as two rows sharing an observation group ID.
- Multi-turn text is stored as `Turn 1: … || Turn 2: …` and must be split on ` || ` by the runner.
- **Retrieval policy and retrieval fact are recorded separately.** `search_policy` records what was permitted; `retrieval_used` records what actually happened. A retrieval-lane observation where retrieval did not run is not silently treated as a closed-model observation.
- Citation metadata, generated queries and live/cached state are captured whenever the surface exposes them, and hashed into `citation_payload_hash`.
- Model and version strings are recorded verbatim on every row. A version change opens a new configuration and is flagged as drift, not absorbed into the trend.

---

## 6. Weights

Two components, stored separately, never combined into one score.

**Exposure — `equal_within_declared_strata`.** No credible audience, intent, locale or surface prevalence evidence exists for AI-assistant travel-connectivity questions. Equal weighting is a stated convention, normalised within `partition × aided_status × lane × surface × locale × wave`. It is **not** a prevalence estimate, and no equal-weighted result may be called market share, audience reach, awareness or share of users.

**Priority — `withheld_pending_human_approval`.** Strategic importance is a human judgement. No priority weight is applied and no priority-weighted number may be published until Gate 4.

Neither component may ever depend on baseline visibility or campaign performance.

---

## 7. Uncertainty

- **Wilson intervals** for simple unweighted binary strata.
- **Stratified cluster bootstrap by `canonical_cell_id`** for any aggregate across cells. Variants and repeats are nested observations inside a cell, not independent buyers.
- **Cluster unit is the canonical cell**, always.
- **Small-subgroup rule:** any stratum with fewer than 20 distinct canonical cells is reported as counts and example responses, never as a percentage leaderboard. In this version, *every partition except core* is below that threshold — and core itself is at 16 cells, so even core-level percentages should carry the counts beside them.
- Intervals quantify conditional run and sampling uncertainty. They do not repair coverage bias, and this panel's coverage bias is substantial (see §12).
- Wave-over-wave comparisons pair unchanged cells. A version change reports overlap-only change plus both full-version levels.

---

## 8. Variance pilot

**Status: pending. It has not run.**

- **Design:** 12 diverse cells — the 4 sentinels plus `cell-001`, `cell-003`, `cell-008`, `cell-010`, `cell-014`, `cell-016`, `cell-031`, `cell-033` — at 6 repeats per candidate, across at least 2 time blocks, on 2 surfaces.
- **Components to estimate:** between-cell, within-cell run, variant wording, day/time block, model/surface, and invalid/parser variance.
- **Decision rule:** high within-cell correlation favours adding unique cells; high run variance favours adding repeats.
- **Why it matters here more than usual:** only 2 of 61 selected candidates are observed language. The other 59 are evidence-grounded paraphrases written by a model. Variant-wording variance is therefore the single most important component to measure before anyone trusts a rate.

Until the pilot runs, report counts and examples, not confidence intervals.

---

## 9. Randomization

- Seed `20260725`, recorded in the panel and the run manifest.
- Candidate order is randomised within each wave.
- Surfaces are interleaved so time-of-day effects do not align with a partition.
- Two time blocks per wave.

---

## 10. Cadence

| Activity | Frequency |
| --- | --- |
| Measurement waves | 3 per quarter |
| Evidence intake | Monthly |
| Panel review | Quarterly, targeting ~70–80% core, 15–25% rotating, 5–10% sentinel + control |
| Charter approval | Annual |

**Event-triggered review** on any of: target product or plan-structure change; device-compatibility policy change; destination coverage or network change; SIM/eSIM registration regulation change; model or surface version change; competitor entry or exit visible in the answer set.

---

## 11. Refresh and version policy

Dated and mutable claims each carry an explicit review-by rule. Undated company copy is **not** treated as timeless.

| Claim class | Sources | Review by | Rule |
| --- | --- | --- | --- |
| B5 story/trend (`cell-021`) | source-010 (2026-02-10) | 2026-10-25 | Re-verify against evidence published within 90 days of the wave, or demote the cell |
| Price and package facts | source-001, source-003 | 2026-10-25 | Prices rendered in CAD at access time and are geo-dependent. Never embed a price in a prompt; re-verify before any scoring rule references one |
| Fair-use throttling threshold | source-004 (2026-07-02) | 2026-10-25 | Single independent observation (~3 GB/day, ~1 Mbps). Re-verify per package before treating as fact |
| Coverage counts / country lists | source-005 (2025-09-12) | 2026-09-25 | The 137-country figure is already stale for scoring; re-verify before use |
| Device compatibility / carrier lock | source-001, source-004, source-005 | 2026-10-25 | Device lists change with handset releases; re-verify before scoring `aided_brand_knowledge` |
| Identity-registration rules | source-008 (undated) | 2026-08-25 | Replace with a dated jurisdictional source or drop `cell-023` |

**Versioning.** Changing core cells, weights, metric definitions or the surface mix creates a new semantic version plus an overlap bridge built on the four sentinel cells. History is append-only in `panel_change_ledger.json`. A frozen version is never overwritten.

**Freeze blockers, all currently open:** `source_manifest_hash`, every `content_hash`, every `prompt_hash`, `blind_brief_hash` and every configuration/response hash are `null` because this runtime has no hashing support. No placeholder digest has been written. Real SHA-256 values are required before freeze.

---

## 12. Approvals

All four human gates are **pending**. No human has approved anything in this panel.

| Gate | Blocks |
| --- | --- |
| Gate 1 — ICPs | Promotion of `icp-006` (partner/reseller) or any `hypothesis_only` ICP into core |
| Gate 2 — Jobs | Promotion of grade-C jobs `job-005`, `job-008`, `job-009` |
| Gate 3 — Partitions | Core membership of `cell-007`, `cell-012`, `cell-013`; the quarantine decisions on `prompt-022a`, `prompt-024a`, `prompt-024b` |
| Gate 4 — Panel | Weights, limitations wording, cadence, claim language, concrete surfaces, pilot funding, and any frozen version |

Every gate is resumable from the artifacts: the pending questions are enumerated in `panel_change_ledger.json` under `pending_decisions`.

---

## 13. Limitations

1. Estimates are **conditional on this panel** and this configuration. They are not a probability sample of AI users or of travellers.
2. Only **one** first-party buyer-behaviour source was reachable (FlyerTalk thread titles). Trustpilot (403), Reddit, travel.stackexchange, both app-store listings (404), a GSMA resource page (403) and one buying guide (membership wall) were all unavailable at access time. Buyer language is thinner than the design intends.
3. Only **2 of 61** selected candidates are observed language. The rest are evidence-grounded paraphrases. Wording effects are a live risk.
4. Equal exposure weighting is a convention, not evidence of prevalence.
5. No priority weights exist, so no strategic-importance-weighted number exists.
6. 25 unaided cells against a 30–48 diagnostic default (waiver W-001). Coverage is deliberately short rather than padded.
7. No non-English locale is measured; en-CA is quarantined. Any claim about non-English AI visibility is unsupported.
8. No campaign lane, so no attribution, no before/after causal claim, and no `campaign_response`.
9. No variance pilot, so repetition counts, interval widths and effective sample size are unvalidated.
10. B5 coverage is a single retrieval-only cell on five-month-old evidence.
11. Company-asserted facts about plans, prices, coverage and device support were captured at one moment from one geography (CAD pricing) and are mutable.
12. Reported differences between waves are conditional and non-causal unless a pre-registered experiment justifies more.

---

## 14. Waivers

| ID | Subject | Evidence needed to close |
| --- | --- | --- |
| W-001 | Unaided cell count below diagnostic default | Authorised customer conversations, on-site search/query data, or reachable review and community corpora |
| W-002 | Partner/reseller area excluded | Independent partner-side sources or authorised customer evidence |
| W-003 | Loyalty/referral/stored-value area excluded | Buyer-language evidence that these programmes drive an information need |
| W-004 | Non-English locales excluded | Locale-specific evidence plus a named reviewer per locale |
| W-005 | Grade-C cells inside core (`cell-007`, `cell-012`, `cell-013`) | A behavioural or independent source per cell, or demotion at Gate 3 |
| W-006 | Consumer surface limited to sentinels | An approved collection budget |
| W-007 | Global-package area only partially covered | Forum or review language from travellers selecting a global plan |

---

## 15. Claims discipline

The evidence ladder is reported in separate rungs and never collapsed:

1. prompt-panel mention, framing and citation — *this panel measures rung 1 only*;
2. source or referral traffic;
3. self-reported discovery;
4. qualified lead or conversion;
5. incremental outcome from an experiment or counterfactual.

A rung-1 movement is never renamed revenue attribution, and a before/after increase alone is never called causal.
