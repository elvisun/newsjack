# AI visibility panel — travel eSIM marketplace (local / regional / global packages)

**Target:** https://www.airalo.com/ — travel eSIM marketplace with local, regional and global packages constrained by destination, device and plan
**Panel:** `panel-travel-esim-marketplace` v0.1.0 · **status `provisional_directional`**
**Built:** 2026-07-25T20:00:00Z · **Next review:** 2026-10-25
**Inputs supplied by the user:** one public URL and one sentence. No charter, competitors, budget, customer evidence, campaign, locales or approver were supplied.

> This is a candidate panel, not a frozen measurement instrument. Nothing has been measured. All four human gates are pending. Do not describe it as frozen, representative, statistically significant, or causal.

---

## 1. Decision and limits

**Business decision this panel is designed to inform.** Decide whether assistant answers to real travel-connectivity questions surface this marketplace, in which buyer situations and constraint types it is absent, and where the evidence is too thin to measure at all — so content and evidence work can be prioritised for the next two quarters.

**Estimands** (full numerators/denominators in `tracking_plan.md` §1 and `measurement_charter.json`):

- `unaided_brand_presence` — core + sentinel, `aided_status = unaided`
- `competitive_mention_share` — aided partition, `category_aided` only
- `citation_presence` — retrieval lane, `retrieval_used = true`
- `aided_brand_knowledge` — aided partition, `target_aided` only
- `answer_framing` — within a single cell

`campaign_response` is excluded: no campaign was supplied, so there is no campaign lane, treatment set or pre-registration.

**Population.** People who ask a general-purpose AI assistant, in English, how to get mobile data while travelling outside their home country, in the situations evidenced in `buyer_jobs.json`. This is a declared analytical population, **not** a probability sample of AI users.

**Lanes.** `closed_model`, `retrieval`, `consumer_surface` (sentinels only). No `campaign_experiment`.

**Locales.** en-US only. en-CA is quarantined pending locale review. No non-English locale is measured.

**Everything below is conditional on this panel.** Equal weighting is a convention, not prevalence. A mention rate here is not market share, awareness, reach or revenue.

---

## 2. Evidence base

**14 sources.** By class: `company_asserted` 3 · `independent` 6 · `buyer_behavior` 1 · `search_proxy` 4 · `llm_hypothesis` 0.
By grade: **B** 4 (source-004, 005, 006, 007) · **C** 10 · A 0 · D 0.

| Requirement from the method | Met? | Evidence |
| --- | --- | --- |
| Target product/capability pages | Yes | source-001 (homepage: local/regional/global packages, 200+ locations, device requirements), source-003 (Japan packages, networks, top-up, install policy) |
| Target pricing / support / technical pages | Yes | source-002 (help centre popular questions and categories), source-003 (package pricing, rendered in CAD) |
| ≥2 independent sources testing the claims | Yes | source-004 (2026-07-02, fair-use throttling, carrier-lock failure, no local number, single-device), source-005 (2025-09-12, 137-country global plans, device-origin limits, cashback, top-up) |
| ≥3 buyer-language sources | Partially | source-006 (FlyerTalk thread titles — the only first-party buyer writing), source-007 (secondary characterisation of Reddit phrasings), source-002 (support-question titles), source-011/012/013 (question-shaped titles) |
| Competitor / category sources | Yes | source-005, source-007, source-013, source-014 — Holafly, Saily, Nomad, Ubigi, Yesim, eSIM Go, GigSky, Maya Mobile, Jetpac, Sakura Mobile, Orange, Visible |
| Fresh dated evidence per B5 cell | Partially | source-010 (2026-02-10) anchors `cell-021`. source-009 (2026-03-25) anchored `cell-022`, whose event concluded 2026-07-19 — **quarantined as decayed** |

**Blocked evidence — the most important limitation in this report.** Trustpilot returned HTTP 403; Reddit and travel.stackexchange could not be fetched; both app-store listings returned 404; a GSMA resource page returned 403; one buying guide sat behind a membership wall. The consequence is that **only one source (source-006) is direct buyer-written language**, and only **2 of 61** selected prompts are observed language. Everything else is an evidence-grounded paraphrase.

**Grades and what they mean here.** No grade-A source exists: there is no authorised customer corpus, no interview, no query export. Grade B is carried by two independent reviews and two community/forum proxies. Three sources (source-011, 012, 014) were observed as **search-result titles only** — their page bodies were never retrieved, and they are graded C with low confidence and labelled as such in the manifest.

**Conflicts and counterevidence, kept rather than smoothed.**

- The target's homepage frames the offer as "affordable, flexible coverage"; source-004 reports unlimited regional plans throttling to roughly 1 Mbps after about 3 GB/day, making heavy tethering impractical. Both are retained; the second is what drives `cell-008`, `cell-009`, `cell-030`, `cell-034`.
- The target states devices must be eSIM-capable and unlocked; source-004 calls carrier lock "the most common complaint from new users" and source-005 adds that handsets sold in China, Hong Kong, Macao and Taiwan may lack support. This counterevidence drives `cell-010`, `cell-011`, `cell-019`, `cell-027`.
- source-013 shows the "how much data" query space is dense with affiliate discount codes for nine providers including the target — so `competitive_mention_share` in that space partly reflects publishing economics, not merit. Recorded as a limitation on that estimand.

**Permissions and personal data.** All 14 sources are public. No named individuals were collected from any source; no reviewer or forum-poster identity appears in any artifact. Spans are short paraphrases or brief quoted fragments, not long excerpts.

**Decay.** Every mutable claim class — prices, coverage counts, fair-use thresholds, device lists, registration rules, B5 stories — carries an explicit review-by date in `panel.yaml` and `tracking_plan.md` §11. The CAD price rendering is treated as geo-dependent and time-bound, and no price is embedded in any prompt.

---

## 3. ICPs and buyer jobs

Six ICP hypotheses, five supported and one hypothesis-only. Nine jobs, seven supported and two hypothesis-only. Full records in `icp_hypotheses.json` and `buyer_jobs.json`.

| ICP | Context | Trigger | Constraints / disqualifiers | Confidence |
| --- | --- | --- | --- | --- |
| `icp-001` | Short-trip leisure traveller, single destination | Trip booked; home-carrier roaming charged per day | eSIM-capable + unlocked handset; install needs internet; plan bound to destination. Disqualifier: carrier-locked handset | medium, supported |
| `icp-002` | Multi-country traveller (regional/global) | Itinerary crosses borders; per-country buying is repetitive | Coverage lists are finite; validity time-boxed | medium, supported |
| `icp-003` | Heavy-data traveller / remote worker | Needs to work or navigate continuously; discovers throttling | ~1 Mbps after ~3 GB/day on unlimited regional plans; some plans single-device | medium, supported |
| `icp-004` | Blocked at device or identity layer | Compatibility is the most-viewed help topic; carrier lock is the most common new-user complaint | Device support, carrier lock, device origin, local registration law | medium, supported |
| `icp-005` | Post-purchase traveller in destination | "Why is my eSIM not working?"; package expiry; top-up | Line cannot move to another handset; support response reported slow | medium, supported |
| `icp-006` | Business / reseller partner | Partner programme published by the target | — | **low, hypothesis_only — excluded from the panel (W-002)** |

**Buying roles.** `role-001` first-time buyer (self-buyer, low-ticket) · `role-002` experienced multi-country traveller who also advises others · `role-003` heavy-data remote worker · `role-004` traveller in destination with a failing or expiring line. There is **no evidence** for an economic buyer, approver or procurement role distinct from the user — a household or business buying committee was not invented.

| Job | Statement (abbreviated) | Acts | Journey states | Grade |
| --- | --- | --- | --- | --- |
| `job-001` | Arrive with working data instead of paying roaming, handset must support a digital line and stay unlocked | explain, plan, compare, recommend | problem_identification → supplier_selection | B |
| `job-002` | Keep one plan working across borders instead of buying per country | explain, compare, plan, recommend | exploration, requirements_building, supplier_selection | B |
| `job-003` | Size the plan and avoid fair-use throttling while tethering | explain, plan, compare, verify | requirements_building, exploration | B |
| `job-004` | Verify device support, carrier lock and device origin before paying | verify, diagnose, explain | requirements_building, problem_identification | B |
| `job-005` | Keep home number, messaging apps and bank codes working on a data-only line | explain, verify, implement | requirements_building, adoption | C |
| `job-006` | Restore connectivity after landing with no service | troubleshoot, diagnose | post_purchase, adoption | B |
| `job-007` | Extend an existing line rather than lose it mid-trip | plan, buy, explain | post_purchase | B |
| `job-008` | Understand what actually changed about buying data abroad | explain | problem_identification, exploration | C, hypothesis_only |
| `job-009` | Find out whether a destination requires identity registration | verify, explain | requirements_building | C, hypothesis_only |

**Language worth preserving** (source language, not a measured distribution): *"which eSIM should I get for Japan?"*, *"Best e-sim for NZ?"*, *"Best eSIM for Bali in 2026? Coverage & speed experiences"*, *"Dual eSim Contortions.. should I just bite the bullet?"*, *"Is it really unlimited?"*, *"Why is my eSIM not working?"*, *"When does my eSIM data package expire?"*, *"How do I check if my iOS device supports eSIM?"*.

**Negative and post-purchase evidence kept:** no local phone number on standard plans; line cannot be moved to a friend's handset; support queue waits; rural coverage complaints; single-device restriction.

---

## 4. Comprehensive prompt list

All **68 candidates** across **34 canonical cells**, including the 7 that did not pass QA. These are the exact strings to track.

**Legend.** Bands: B0 direct brand · B1 comparison/purchase · B2 category · B3 problem/need · B4 job/goal · B5 discovery/story. Aided: UN unaided · CAT category-aided · COMP competitor-aided · TGT target-aided. Lanes: CM closed_model · RT retrieval · CS consumer_surface. Variant: OBS observed_language · PAR natural_paraphrase · SEN sensitivity. Turn: 1T single_turn · MT scripted_multi_turn. **Weight status is identical for every row — exposure `equal_within_declared_strata` (a convention, not prevalence); priority `withheld_pending_human_approval`** — and is repeated per row rather than collapsed into one score. Multi-turn text splits on ` || `.

| Prompt ID | Exact prompt | Variant | Partition | Band | Aided state | Campaign-exposed | Buyer job | Information act | Journey | Funnel | Role/persona | Locale/language | Constraints | Expected answer | Turn form | Lanes/surfaces | Evidence | Transformation | Weight status | QA |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| prompt-001a | I'm flying to Japan next month and my carrier wants $12 a day to use my phone there. What are my options? | PAR | core | B3 | UN | false | job-001 single-destination data | explain | problem_identification | null | role-001 first-time buyer | en-US / en | destination-bound | options_and_tradeoffs | 1T | CM, RT | B; source-010, source-004 | human_written | Exp equal / Pri withheld | pass |
| prompt-001b | Going overseas for a week soon. How do people usually avoid paying their carrier's daily fee just to use the internet on their phone? | PAR | core | B3 | UN | false | job-001 | explain | problem_identification | null | role-001 | en-US / en | destination-bound | options_and_tradeoffs | 1T | CM, RT | B; source-010, source-004 | human_written | Exp equal / Pri withheld | pass |
| prompt-001c | I'm going abroad for a week - which eSIM company should I buy from? | SEN | core | B3 | UN | false | job-001 | explain | problem_identification | null | role-001 | en-US / en | destination-bound | options_and_tradeoffs | 1T | CM, RT | D; none | llm_expanded | Exp equal / Pri withheld | **reject** — recommendation forcing; wrong band |
| prompt-002a | How do I make sure I have internet on my phone the whole time I'm in Japan for a week? | PAR | core | B4 | UN | false | job-001 | plan | exploration | null | role-001 | en-US / en | destination-bound; install-needs-internet | step_by_step_plan | 1T | CM, RT | B; source-007 | human_written | Exp equal / Pri withheld | pass |
| prompt-002b | Landing in Osaka in two weeks and I want maps and messaging working from the airport. What do I need to sort out before I go? | PAR | core | B4 | UN | false | job-001 | plan | exploration | null | role-001 | en-US / en | destination-bound; install-needs-internet | step_by_step_plan | 1T | CM, RT | B; source-006, source-007 | human_written | Exp equal / Pri withheld | pass |
| prompt-003a | Which eSIM should I get for Japan? | **OBS** | core | B1 | UN | false | job-001 | recommend | supplier_selection | BOFU | role-001 | en-US / en | destination-bound | ranked_shortlist | 1T | CM, RT | B; source-007 | lightly_normalized | Exp equal / Pri withheld | pass |
| prompt-003b | What's the best travel eSIM to buy for a 10 day trip to Japan? | PAR | core | B1 | UN | false | job-001 | recommend | supplier_selection | BOFU | role-001 | en-US / en | destination-bound | ranked_shortlist | 1T | CM, RT | B; source-006, source-007 | human_written | Exp equal / Pri withheld | pass — split from cell-017 (destination constraint protected) |
| prompt-004a | eSIM vs buying a local SIM when I land vs my carrier's roaming pass - which actually makes sense for two weeks? | PAR | core | B1 | UN | false | job-001 | compare | requirements_building | null | role-002 multi-country traveller | en-US / en | destination-bound; data-only | comparison_table | 1T | CM, RT | B; source-004, source-005 | human_written | Exp equal / Pri withheld | pass |
| prompt-004b | For a two week trip, is a travel eSIM cheaper than a local prepaid SIM or roaming? | PAR | core | B1 | UN | false | job-001 | compare | requirements_building | null | role-002 | en-US / en | destination-bound; data-only | comparison_table | 1T | CM, RT | B; source-004, source-010 | human_written | Exp equal / Pri withheld | pass |
| prompt-005a | Three weeks, five countries in Europe. How do I keep my phone online the whole trip without sorting something out at every border? | PAR | core | B4 | UN | false | job-002 cross-border one plan | plan | exploration | null | role-002 | en-US / en | destination-bound; validity-window | plan_and_options | 1T | CM, RT | B; source-006, source-005 | human_written | Exp equal / Pri withheld | pass |
| prompt-005b | I'm doing a multi-country Europe trip in September and want one thing that works everywhere. How should I plan the phone side? | PAR | core | B4 | UN | false | job-002 | plan | exploration | null | role-002 | en-US / en | destination-bound; validity-window | plan_and_options | 1T | CM, RT | B; source-006 | human_written | Exp equal / Pri withheld | pass |
| prompt-006a | Is there one data plan that covers several European countries instead of buying one per country? | PAR | core | B1 | UN | false | job-002 | recommend | supplier_selection | null | role-002 | en-US / en | destination-bound | ranked_shortlist | 1T | CM, RT | B; source-006, source-005 | search_query_expanded | Exp equal / Pri withheld | pass |
| prompt-006b | Best regional travel eSIM that works across Spain, France, Italy and Greece on one plan? | PAR | core | B1 | UN | false | job-002 | recommend | supplier_selection | null | role-002 | en-US / en | destination-bound | ranked_shortlist | 1T | CM, RT | B; source-005, source-006 | human_written | Exp equal / Pri withheld | pass — split from cell-005 (act differs) |
| prompt-007a | How much data do I need for a week in Europe? | PAR | core | B3 | UN | false | job-003 sizing / throttling | explain | requirements_building | null | role-001 | en-US / en | validity-window | quantity_estimate | 1T | CM, RT | C; source-012 | search_query_expanded | Exp equal / Pri withheld | pass — **Gate 3: grade-C in core** |
| prompt-007b | Two of us going to Italy for 8 days, mostly maps, WhatsApp and some Instagram. How many GB is realistic? | PAR | core | B3 | UN | false | job-003 | explain | requirements_building | null | role-001 | en-US / en | validity-window | quantity_estimate | 1T | CM, RT | C; source-012, source-013 | human_written | Exp equal / Pri withheld | pass — **Gate 3: grade-C in core** |
| prompt-007c | How much data do I need for a week in Europe and which provider is cheapest? | SEN | core | B3 | UN | false | job-003 | explain | requirements_building | null | role-001 | en-US / en | validity-window | quantity_estimate | 1T | CM, RT | D; source-012 | llm_expanded | Exp equal / Pri withheld | **revise** — two concepts; provider clause forces a supplier answer |
| prompt-008a | If a travel plan says unlimited data, does it actually slow down after a few GB a day? | PAR | core | B3 | UN | false | job-003 | verify | requirements_building | null | role-003 remote worker | en-US / en | fair-use-throttle; tethering | caveat_and_verification | 1T | CM, RT | B; source-004, source-007 | human_written | Exp equal / Pri withheld | pass — competitor brand removed from the source phrasing |
| prompt-008b | Do unlimited data plans for travellers throttle you? I need to tether my laptop most days. | PAR | core | B3 | UN | false | job-003 | verify | requirements_building | null | role-003 | en-US / en | fair-use-throttle; tethering | caveat_and_verification | 1T | CM, RT | B; source-004 | human_written | Exp equal / Pri withheld | pass |
| prompt-009a | Working remotely from Portugal for a month and I need to tether my laptop all day. How do I sort out reliable data? | PAR | core | B4 | UN | false | job-003 | plan | exploration | null | role-003 | en-US / en | tethering; fair-use-throttle; single-device | plan_and_options | 1T | CM, RT | B; source-004, source-007 | human_written | Exp equal / Pri withheld | pass |
| prompt-009b | I'm working from Southeast Asia for six weeks and hotspot my laptop most days. What's the best way to get data that won't get throttled? | PAR | core | B4 | UN | false | job-003 | plan | exploration | null | role-003 | en-US / en | tethering; fair-use-throttle; single-device | plan_and_options | 1T | CM, RT | B; source-004 | human_written | Exp equal / Pri withheld | pass |
| prompt-010a | How do I check whether my phone can use a foreign mobile line before I buy one? | PAR | core | B3 | UN | false | job-004 device eligibility | verify | requirements_building | null | role-001 | en-US / en | device-support; carrier-lock | verification_steps | 1T | CM, RT | B; source-002, source-004 | search_query_expanded | Exp equal / Pri withheld | pass |
| prompt-010b | Not sure if my Pixel will work with a travel data plan. How can I tell before I pay? | PAR | core | B3 | UN | false | job-004 | verify | requirements_building | null | role-001 | en-US / en | device-support; carrier-lock | verification_steps | 1T | CM, RT | B; source-002, source-004 | human_written | Exp equal / Pri withheld | pass — split from cell-019 (named handset protected) |
| prompt-011a | My phone might still be locked to my carrier. Does that stop me using cheaper data abroad? | PAR | core | B3 | UN | false | job-004 | diagnose | problem_identification | null | role-001 | en-US / en | carrier-lock; device-origin | diagnosis_and_options | 1T | CM, RT | B; source-004 | human_written | Exp equal / Pri withheld | pass |
| prompt-011b | I bought my phone in Hong Kong. Will that be a problem for getting data when I travel? | PAR | core | B3 | UN | false | job-004 | diagnose | problem_identification | null | role-001 | en-US / en | carrier-lock; device-origin | diagnosis_and_options | 1T | CM, RT | B; source-005 | human_written | Exp equal / Pri withheld | pass |
| prompt-012a | Can I still use WhatsApp if I switch to a data-only line abroad? | PAR | core | B3 | UN | false | job-005 keep home number | explain | requirements_building | null | role-001 | en-US / en | data-only | explanation_and_caveats | 1T | CM, RT | C; source-011 | search_query_expanded | Exp equal / Pri withheld | pass — **Gate 3: grade-C in core** |
| prompt-012b | If I use a separate data connection while travelling, will my bank still be able to text me codes? | PAR | core | B3 | UN | false | job-005 | explain | requirements_building | null | role-001 | en-US / en | data-only | explanation_and_caveats | 1T | CM, RT | C; source-011, source-004 | human_written | Exp equal / Pri withheld | pass — **Gate 3: grade-C in core** |
| prompt-013a | Turn 1: I want my normal number to keep working while I use cheaper data in Thailand. How do I set that up on an iPhone? \|\| Turn 2: Do I need to turn anything off so my own carrier doesn't charge me roaming? | PAR | core | B4 | UN | false | job-005 | implement | adoption | null | role-001 | en-US / en | data-only; device-support | step_by_step_plan | **MT** | CM, RT | C; source-011, source-005 | human_written | Exp equal / Pri withheld | pass — **Gate 3: grade-C in core** |
| prompt-013b | Turn 1: How do I set my Android up to run two lines, my normal number plus data abroad? \|\| Turn 2: Which one should be the default for texts? | PAR | core | B4 | UN | false | job-005 | implement | adoption | null | role-001 | en-US / en | data-only; device-support | step_by_step_plan | **MT** | CM, RT | C; source-011, source-005 | human_written | Exp equal / Pri withheld | pass — **Gate 3: grade-C in core** |
| prompt-014a | Just landed and my phone says no service even though I set up data before I flew. What now? | PAR | core | B3 | UN | false | job-006 no service on arrival | troubleshoot | post_purchase | null | role-004 in-destination | en-US / en | install-needs-internet; single-device | troubleshooting_steps | 1T | CM, RT | B; source-002, source-004 | human_written | Exp equal / Pri withheld | pass |
| prompt-014b | I set up a travel data line before my flight and it isn't connecting in Bangkok. How do I fix it when I have no data to look things up? | PAR | core | B3 | UN | false | job-006 | troubleshoot | post_purchase | null | role-004 | en-US / en | install-needs-internet; single-device | troubleshooting_steps | 1T | CM, RT | B; source-002, source-004 | human_written | Exp equal / Pri withheld | pass |
| prompt-014c | My Airalo eSIM isn't working after landing in Japan - what should I do? | SEN | core | B3 | UN | false | job-006 | troubleshoot | post_purchase | null | role-004 | en-US / en | install-needs-internet; single-device | troubleshooting_steps | 1T | CM, RT | D; source-002 | llm_expanded | Exp equal / Pri withheld | **reject** — target brand token in an unaided core cell |
| prompt-015a | I'm halfway through my trip and nearly out of data. Is it better to extend what I have or start a new plan? | PAR | core | B3 | UN | false | job-007 extend / top up | plan | post_purchase | null | role-004 | en-US / en | validity-window; topup | options_and_tradeoffs | 1T | CM, RT | B; source-002, source-005 | human_written | Exp equal / Pri withheld | pass |
| prompt-015b | My travel data expires in two days but I'm here another week. What are my options? | PAR | core | B3 | UN | false | job-007 | plan | post_purchase | null | role-004 | en-US / en | validity-window; topup | options_and_tradeoffs | 1T | CM, RT | B; source-002, source-003 | human_written | Exp equal / Pri withheld | pass — mutable validity facts carry a refresh rule |
| prompt-016a | Every time I cross into the next country my phone data stops working. Why does that keep happening? | PAR | core | B3 | UN | false | job-002 | diagnose | problem_identification | null | role-002 | en-US / en | destination-bound | diagnosis_and_options | 1T | CM, RT | B; source-005, source-006 | human_written | Exp equal / Pri withheld | pass |
| prompt-016b | Data worked fine in France but died as soon as I got to Switzerland. What's going on? | PAR | core | B3 | UN | false | job-002 | diagnose | problem_identification | null | role-002 | en-US / en | destination-bound | diagnosis_and_options | 1T | CM, RT | B; source-005 | human_written | Exp equal / Pri withheld | pass |
| prompt-017a | What's the best travel eSIM for an international trip? | PAR | sentinel | B1 | UN | false | job-001 | recommend | supplier_selection | null | role-001 | en-US / en | none | ranked_shortlist | 1T | CM, RT, CS | B; source-006, source-007 | human_written | Exp equal / Pri withheld | pass |
| prompt-017b | Which travel eSIM should I buy? | PAR | sentinel | B1 | UN | false | job-001 | recommend | supplier_selection | null | role-001 | en-US / en | none | ranked_shortlist | 1T | CM, RT, CS | B; source-007 | human_written | Exp equal / Pri withheld | pass — deliberate terse/verbose pair for the wording-sensitivity pilot |
| prompt-018a | How much mobile data do I need for a two week trip? | PAR | sentinel | B3 | UN | false | job-003 | explain | requirements_building | null | role-001 | en-US / en | validity-window | quantity_estimate | 1T | CM, RT, CS | C; source-012, source-013 | search_query_expanded | Exp equal / Pri withheld | pass |
| prompt-018b | Is 5GB enough for two weeks abroad? | PAR | sentinel | B3 | UN | false | job-003 | explain | requirements_building | null | role-001 | en-US / en | validity-window | quantity_estimate | 1T | CM, RT, CS | C; source-013 | human_written | Exp equal / Pri withheld | pass |
| prompt-019a | How can I tell if my phone will work with a foreign mobile line before I travel? | PAR | sentinel | B3 | UN | false | job-004 | verify | requirements_building | null | role-001 | en-US / en | device-support | verification_steps | 1T | CM, RT, CS | B; source-002, source-004 | human_written | Exp equal / Pri withheld | pass |
| prompt-019b | Does my iPhone support a second digital line? | PAR | sentinel | B3 | UN | false | job-004 | verify | requirements_building | null | role-001 | en-US / en | device-support | verification_steps | 1T | CM, RT, CS | C; source-002 | search_query_expanded | Exp equal / Pri withheld | pass |
| prompt-020a | Travel data plan installed but I have no signal after landing. | PAR | sentinel | B3 | UN | false | job-006 | troubleshoot | post_purchase | null | role-004 | en-US / en | install-needs-internet | troubleshooting_steps | 1T | CM, RT, CS | B; source-002, source-004 | human_written | Exp equal / Pri withheld | pass — deliberately terse/imperfect style |
| prompt-020b | Phone shows no service abroad even though I bought data before the trip. Why is it not working? | PAR | sentinel | B3 | UN | false | job-006 | troubleshoot | post_purchase | null | role-004 | en-US / en | install-needs-internet | troubleshooting_steps | 1T | CM, RT, CS | B; source-002 | search_query_expanded | Exp equal / Pri withheld | pass — split from cell-014 (constraints protected) |
| prompt-021a | Are travellers actually moving away from carrier roaming plans in 2026, and what's driving it? | PAR | rotating | B5 | UN | false | job-008 roaming-shift story | explain | problem_identification | TOFU | role-002 | en-US / en | none | dated_trend_summary | 1T | **RT only** | C; source-010 (2026-02-10) | human_written | Exp equal / Pri withheld | pass — review-by 2026-10-25 |
| prompt-021c | Are third-party data providers reshaping how mobile operators earn roaming revenue? | SEN | rotating | B5 | UN | false | job-008 | explain | problem_identification | TOFU | role-002 | en-US / en | none | dated_trend_summary | 1T | RT only | D; none | llm_expanded | Exp equal / Pri withheld | **quarantine** — grade-D expansion with no supporting source |
| prompt-022a | How did fans travelling between the US, Canada and Mexico for the 2026 World Cup handle phone data across all three countries? | PAR | rotating | B5 | UN | false | job-008 | explain | exploration | TOFU | role-002 | en-US / en | destination-bound | dated_trend_summary | 1T | RT only | C; source-009 (2026-03-25) | human_written | Exp equal / Pri withheld | **quarantine** — story decayed; event concluded 2026-07-19 |
| prompt-023a | Do any countries make you register your passport before you can use a mobile line as a tourist? | PAR | rotating | B3 | UN | false | job-009 ID registration | verify | requirements_building | null | role-001 | en-US / en | id-registration | verification_steps | 1T | CM, RT | C; source-008 (undated) | human_written | Exp equal / Pri withheld | pass — open phrasing because evidence names no country |
| prompt-024a | Best travel eSIM for a Canadian going to the States for a week? | PAR | rotating | B1 | UN | false | job-001 | recommend | supplier_selection | null | role-001 | **en-CA** / en | destination-bound | ranked_shortlist | 1T | CM, RT | C; source-001, source-003 | human_written | Exp equal / Pri withheld | **quarantine** — locale review pending |
| prompt-024b | What should I get for data when I'm flying from Toronto to Mexico this winter? | SEN | rotating | B1 | UN | false | job-001 | recommend | supplier_selection | null | role-001 | **en-CA** / en | destination-bound | ranked_shortlist | 1T | CM, RT | D; source-001 | llm_expanded | Exp equal / Pri withheld | **quarantine** — locale review pending + grade D |
| prompt-025a | Plan a 10 day itinerary for northern Japan in October. | PAR | **control** | B4 | UN | false | job-001 (control) | plan | exploration | null | role-001 | en-US / en | none | itinerary_plan_without_connectivity_entailment | 1T | CM, RT | C; source-006 | human_written | Exp equal / Pri withheld — **own denominator, never pooled with core** | pass |
| prompt-026a | Is Airalo reliable in Japan? | **OBS** | aided | B0 | **TGT** | false | job-001 | verify | supplier_selection | BOFU | role-001 | en-US / en | destination-bound | verdict_with_caveats | 1T | CM, RT | B; source-007 | lightly_normalized | Exp equal / Pri withheld — aided denominator | pass — allowed exception exc-001 |
| prompt-026b | Is Airalo any good for a two week trip to Japan, or should I look at something else? | PAR | aided | B0 | TGT | false | job-001 | verify | supplier_selection | BOFU | role-001 | en-US / en | destination-bound | verdict_with_caveats | 1T | CM, RT | B; source-007, source-004 | human_written | Exp equal / Pri withheld — aided denominator | pass — no slogan or flattering claim in the stimulus |
| prompt-027a | Does Airalo work on my phone? I have a Samsung I bought in Taiwan. | PAR | aided | B0 | TGT | false | job-004 | explain | requirements_building | null | role-001 | en-US / en | device-support; carrier-lock; device-origin | eligibility_answer | 1T | CM, RT | B; source-005 | human_written | Exp equal / Pri withheld — aided denominator | pass |
| prompt-027b | Which phones are compatible with Airalo eSIMs? | PAR | aided | B0 | TGT | false | job-004 | explain | requirements_building | null | role-001 | en-US / en | device-support; carrier-lock; device-origin | eligibility_answer | 1T | CM, RT | B; source-002, source-001 | search_query_expanded | Exp equal / Pri withheld — aided denominator | pass — device lists are mutable; refresh rule attached |
| prompt-028a | Turn 1: My Airalo eSIM says no service after landing. How do I fix it? \|\| Turn 2: I've restarted and turned data roaming on, still nothing. What next? | PAR | aided | B0 | TGT | false | job-006 | troubleshoot | post_purchase | null | role-004 | en-US / en | install-needs-internet; single-device | troubleshooting_steps | **MT** | CM, RT | B; source-002, source-004 | human_written | Exp equal / Pri withheld — aided denominator | pass |
| prompt-028b | Turn 1: Installed an Airalo eSIM but I'm getting no data in Italy. \|\| Turn 2: Do I need to change the APN, or should I contact support? | PAR | aided | B0 | TGT | false | job-006 | troubleshoot | post_purchase | null | role-004 | en-US / en | install-needs-internet; single-device | troubleshooting_steps | **MT** | CM, RT | B; source-002, source-004 | human_written | Exp equal / Pri withheld — aided denominator | pass |
| prompt-029a | Holafly or Saily for a trip to Japan - which is better? | PAR | aided | B1 | **COMP** | false | job-001 | compare | supplier_selection | null | role-001 | en-US / en | destination-bound | comparison_table | 1T | CM, RT | B; source-007, source-014 | human_written | Exp equal / Pri withheld — **separate competitor-aided denominator** | pass — allowed exception exc-002 |
| prompt-029b | Nomad vs Holafly for two weeks in Japan, which would you pick? | PAR | aided | B1 | COMP | false | job-001 | compare | supplier_selection | null | role-001 | en-US / en | destination-bound | comparison_table | 1T | CM, RT | B; source-005, source-007 | human_written | Exp equal / Pri withheld — separate competitor-aided denominator | pass |
| prompt-030a | Is Holafly's unlimited plan better than a fixed GB plan if I tether my laptop every day? | PAR | aided | B1 | COMP | false | job-003 | compare | requirements_building | null | role-003 | en-US / en | fair-use-throttle; tethering | comparison_table | 1T | CM, RT | B; source-004, source-007 | human_written | Exp equal / Pri withheld — separate competitor-aided denominator | pass |
| prompt-030b | Saily vs Ubigi for a month in Europe when I need to hotspot most days? | PAR | aided | B1 | COMP | false | job-003 | compare | requirements_building | null | role-003 | en-US / en | fair-use-throttle; tethering | comparison_table | 1T | CM, RT | B; source-004 | human_written | Exp equal / Pri withheld — separate competitor-aided denominator | pass — split from cell-029 (persona + constraint protected) |
| prompt-031a | How do travel eSIMs actually work? | PAR | aided | B2 | **CAT** | false | job-001 | explain | exploration | null | role-001 | en-US / en | device-support; install-needs-internet | explanation_and_caveats | 1T | CM, RT | C; source-002, source-010 | search_query_expanded | Exp equal / Pri withheld — category-aided denominator | pass |
| prompt-031b | What is a travel eSIM and what do I need to use one? | PAR | aided | B2 | CAT | false | job-001 | explain | exploration | null | role-001 | en-US / en | device-support; install-needs-internet | explanation_and_caveats | 1T | CM, RT | C; source-002, source-001 | human_written | Exp equal / Pri withheld — category-aided denominator | pass |
| prompt-032a | What should I look for when choosing a travel eSIM? | PAR | aided | B2 | CAT | false | job-001 | explain | requirements_building | null | role-002 | en-US / en | destination-bound; validity-window; data-only | criteria_checklist | 1T | CM, RT | B; source-004, source-014 | human_written | Exp equal / Pri withheld — category-aided denominator | pass — split from cell-017 (band, act and denominator differ) |
| prompt-032b | What do people get wrong when they buy a travel eSIM for the first time? | PAR | aided | B2 | CAT | false | job-001 | explain | requirements_building | null | role-002 | en-US / en | destination-bound; validity-window; data-only | criteria_checklist | 1T | CM, RT | B; source-004, source-007 | human_written | Exp equal / Pri withheld — category-aided denominator | pass |
| prompt-033a | What's the difference between a regional eSIM and a global eSIM plan? | PAR | aided | B2 | CAT | false | job-002 | compare | requirements_building | null | role-002 | en-US / en | destination-bound | comparison_table | 1T | CM, RT | B; source-001, source-005 | human_written | Exp equal / Pri withheld — category-aided denominator | pass — the only cell covering the global-package area (W-007) |
| prompt-033b | Is a global travel eSIM worth it compared with buying a regional one for each area? | PAR | aided | B2 | CAT | false | job-002 | compare | requirements_building | null | role-002 | en-US / en | destination-bound | comparison_table | 1T | CM, RT | B; source-005 | human_written | Exp equal / Pri withheld — category-aided denominator | pass |
| prompt-034a | Unlimited travel eSIM or a set number of GB - which is better value? | PAR | aided | B2 | CAT | false | job-003 | compare | requirements_building | null | role-003 | en-US / en | fair-use-throttle; validity-window | comparison_table | 1T | CM, RT | B; source-003, source-004 | human_written | Exp equal / Pri withheld — category-aided denominator | pass |
| prompt-034b | Are unlimited travel eSIM plans worth paying extra for over a 10GB plan? | PAR | aided | B2 | CAT | false | job-003 | compare | requirements_building | null | role-003 | en-US / en | fair-use-throttle; validity-window | comparison_table | 1T | CM, RT | B; source-004 | human_written | Exp equal / Pri withheld — category-aided denominator | pass — no mutable price embedded |

---

## 5. Coverage matrix

**Canonical cells: 34 · candidates: 68 · accepted: 61 · cells selected into the panel: 32.**

**By proximity band**

| Band | Cells | Accepted candidates | Notes |
| --- | ---: | ---: | --- |
| B0 direct brand/product | 3 | 6 | All target-aided, aided partition, allowed exception exc-001 |
| B1 comparison/purchase | 7 | 12 | 5 unaided (cells 003, 004, 006, 017, 024*) and 2 competitor-aided; `cell-024` contributes no accepted candidate |
| B2 category | 4 | 8 | All category-aided |
| B3 problem/need | 13 | 25 | The best-evidenced band |
| B4 job/goal | 5 | 9 | Includes the control cell and both multi-turn implementation variants |
| B5 discovery/story | 2 | 1 | **Weakest band.** One retrieval-only cell on 5-month-old evidence; the event-anchored cell is quarantined |

\* `cell-024` is counted as a cell but contributes no accepted candidate.

**By aided state**

| Aided state | Cells | Accepted candidates | Denominator |
| --- | ---: | ---: | --- |
| unaided | 25 | 43 | core + sentinel (40) feed `unaided_brand_presence`; rotating (2) and control (1) have separate denominators |
| category_aided | 4 | 8 | `competitive_mention_share` |
| competitor_aided | 2 | 4 | Reported separately; never pooled |
| target_aided | 3 | 6 | `aided_brand_knowledge` |
| campaign_exposed | 0 | 0 | **Waived — no campaign supplied** |

**By partition** — core 16 cells / 32 candidates · sentinel 4 / 8 · rotating 4 cells but only 2 accepted candidates · control 1 / 1 · aided 9 / 18.

**By buyer job** — job-001: 11 cells · job-002: 4 · job-003: 6 · job-004: 4 · job-005: 2 · job-006: 3 · job-007: 1 · job-008: 2 · job-009: 1. (Total 34.)

**By information act**

| Act | Cells | Status |
| --- | ---: | --- |
| explain | 9 | covered |
| plan | 5 | covered |
| compare | 5 | covered |
| recommend | 4 | covered |
| verify | 5 | covered |
| diagnose | 2 | covered |
| troubleshoot | 3 | covered |
| implement | 1 | covered (multi-turn) |
| buy | 0 | folded into `job-007` `plan` cells; no evidence of a distinct transactional prompt shape |
| navigate | 0 | **waived — no evidence** |
| generate | 0 | **waived — no evidence** |

**By journey state** — problem_identification 4 · exploration 6 · requirements_building 13 · supplier_selection 6 · adoption 1 · post_purchase 4. (Total 34.) All six states covered.

**By role** — role-001: 18 cells · role-002: 8 · role-003: 4 · role-004: 4. (Total 34.) Economic buyer, approver and procurement roles: **waived, no evidence**.

**By locale** — en-US 33 cells · en-CA 1 cell (quarantined, unmeasured) · all other locales **waived (W-004)**.

**By lane** — closed_model 32 cells · retrieval 34 · consumer_surface 4 (sentinels only, W-006) · campaign_experiment 0 (waived).

**By evidence grade** — B: 24 cells · C: 10 cells · A: 0 · D: 0 at cell level. Grade-D exists only at candidate level (4 drafts), and none is accepted. Three grade-C cells sit in core under waiver W-005 and are flagged for Gate 3.

**By turn form** — single_turn 32 cells · scripted_multi_turn 2 cells (`cell-013`, `cell-028`).

**By variant role** — observed_language 2 · natural_paraphrase 59 · sensitivity 0 accepted (4 sensitivity drafts, all rejected or quarantined).

### Required waivers and gaps

| ID | Missing dimension | Why | Evidence needed |
| --- | --- | --- | --- |
| gap-001 / W-004 | Non-English locales (es-MX, ja-JP, de-DE, ar-AE, fr-FR) | No non-English buyer-language evidence; no native reviewer; machine translation is not locale equivalence | Locale-specific forum/review/query evidence plus a named market-competent reviewer per locale |
| gap-002 | en-CA in core | Only company-asserted CAD price rendering distinguishes it | Canadian buyer-language evidence and a locale review of `cell-024` |
| gap-003 / W-002 | Economic-buyer, procurement and partner/reseller roles | Company-asserted evidence only for `icp-006` | Independent partner-side or authorised customer evidence |
| gap-004 | `generate` and `navigate` acts | No evidence travellers ask an assistant for either in this category | Query or conversation evidence showing the act |
| gap-005 | `campaign_experiment` lane | No campaign supplied | A pre-registered campaign with treatment/control definitions |
| gap-006 | Fresh-dated B5 | Best story evidence describes an event that concluded 2026-07-19; trend source is 5 months old | Dated market or regulatory evidence published within 90 days of the wave |
| gap-007 / W-006 | Consumer surface beyond sentinels | Manual collection cost | An approved collection budget |
| gap-008 / W-007 | Global multi-region packages as a distinct intent | No independent or behavioural language specific to global plans | Forum/review language from travellers choosing a global plan |
| W-001 | 25 unaided cells vs a 30–48 diagnostic default | Evidence does not support more distinct unaided intents | Authorised customer conversations, on-site search/query data, or reachable review corpora |
| W-005 | Grade-C cells in core (`cell-007`, `cell-012`, `cell-013`) | Supporting sources are title-only proxies or inference | A behavioural or independent source per cell, or demotion at Gate 3 |

**Perimeter check.** Every area named in the user's description resolves: local packages → cells 002, 003, 014, 026 · regional → 005, 006, 016, 033 · global → 033 (partial, W-007) · destination constraint → 002, 003, 005, 006, 014, 016, 023, 026 · device constraint → 010, 011, 019, 027 · plan constraint → 007, 008, 015, 018, 034. Partner and loyalty areas are explicitly excluded with waivers rather than silently dropped.

---

## 6. QA ledger

| Outcome | Count | IDs |
| --- | ---: | --- |
| pass | 61 | See `prompt_qa.json` `accepted_candidate_ids` |
| revise | 1 | `prompt-007c` |
| quarantine | 4 | `prompt-021c`, `prompt-022a`, `prompt-024a`, `prompt-024b` |
| reject | 2 | `prompt-001c`, `prompt-014c` |
| **total** | **68** | — |

**Contamination results.**

- Target brand/product/domain hits in unaided cells: **1** (`prompt-014c`) → hard fail, rejected.
- Competitor terms in unaided cells: **0**. Where an observed source phrasing named a competitor (*"Is Holafly really unlimited?"*), the brand was removed rather than carried into an unaided cell — see `prompt-008a`.
- Slogan / flattering-claim hits: **0**, including in the B0 aided pass. No stimulus feeds the model the target's own marketing language.
- Proprietary-category hits ("eSIM store", "eSIM marketplace"): **0**. `prompt-021c` was drafted without the phrase after review.
- Campaign-term hits: **0**. No campaign is measured, so campaign terms are prohibited everywhere.
- Allowed exceptions used: exc-001 on cells 026–028 (B0 target-aided), exc-002 on cells 029–030 (B1 competitor-aided).
- Generic vocabulary ("eSIM", "travel eSIM", "regional/global eSIM", "roaming") is deliberately **not** on the target-term list; banning it would make B1 and B2 unmeasurable. B3 and B4 prompts avoid category vocabulary by construction, which is why they say "mobile data", "internet on my phone" or "a foreign mobile line".

**Duplicate decisions — protected differences, not merges.** No candidate was auto-deleted; no embedding model with a fixed version was available, so semantic pairs were reviewed by judgment and recorded.

| Pair | Decision | Protected difference |
| --- | --- | --- |
| `prompt-003b` vs `prompt-017a` | split | Destination constraint vs destination-free sentinel |
| `prompt-006b` vs `prompt-005a` | split | Supplier selection vs outcome planning |
| `prompt-010b` vs `prompt-019a` | split | Named handset context vs generic sentinel |
| `prompt-020b` vs `prompt-014a` | split | Constraint-free sentinel vs offline + single-device constraints |
| `prompt-030b` vs `prompt-029a` | split | Persona, constraint and region all differ |
| `prompt-032a` vs `prompt-017a` | split | B2 criteria / category-aided vs B1 shortlist / unaided — different band, act and denominator |
| `prompt-017a` vs `prompt-017b` | retain both | Deliberate terse/verbose pair for the wording-sensitivity pilot |

**Coverage lost to rejection.** `cell-022` (event B5) and `cell-024` (en-CA) have no accepted candidate and are excluded from the panel; both are recorded as gaps. No other cell lost coverage — every rejected or revised draft duplicated an intent already measured elsewhere.

**Disputed decisions reserved for Gate 3:** grade-C core membership of `cell-007`, `cell-012`, `cell-013`; and the quarantines on `prompt-022a`, `prompt-024a`, `prompt-024b`.

**Blinding.** The unaided generation pass worked only from `blind_design_brief.json` and `prompt_architecture.json`. It received anonymised roles, approved job statements, constraints, safe language fragments, evidence IDs and grades — and no brand, domain, slogan, campaign wording, current answer, ranking, citation, content gap or desired landing page. B0 target-aided prompts were written in a separate, explicitly aided pass that received only the target alias. QA received the contamination register **after** generation. No baseline visibility data exists anywhere in this run, so none could have influenced selection (`baseline_fields_blinded: true`).

---

## 7. Tracking plan

Full detail in `tracking_plan.md`. Summary:

- **Variants:** 2 per core/sentinel/aided cell; 1 per rotating/control cell. Variants are nested observations inside a cell, never extra buyers.
- **Repetitions:** 3 per candidate per surface; 6 for sentinels.
- **Surfaces:** two closed-model, one retrieval, one clean consumer surface (sentinels only). Providers unassigned pending Gate 4.
- **Volume:** 639 observations per wave; 3 waves per quarter.
- **Fresh-session rule:** every observation opens a fresh session; the only continuation is between the two scripted turns of a multi-turn candidate.
- **Retrieval state:** `search_policy` (what was permitted) and `retrieval_used` (what happened) are recorded separately, with generated queries and citation metadata when exposed.
- **Weights:** exposure `equal_within_declared_strata` (a convention, explicitly not prevalence); priority `withheld_pending_human_approval`. Never combined into one score.
- **Uncertainty:** Wilson for simple binary strata; stratified cluster bootstrap by `canonical_cell_id` for aggregates. Any stratum under 20 distinct cells is reported as counts and examples, not a percentage leaderboard — which in this version is every partition except core.
- **Randomization:** seed `20260725`; order randomised within wave; surfaces interleaved across two time blocks.
- **Cadence:** monthly evidence intake, quarterly panel review, annual charter approval, plus event triggers.
- **Refresh:** explicit review-by rules on B5 stories, prices, fair-use thresholds, coverage counts, device lists and registration rules.
- **Next review:** 2026-10-25.

---

## 8. Human gates

**Approvals made: none.** Every gate is pending, and every gate is resumable from the artifacts.

| Gate | Status | What it blocks | Pending question |
| --- | --- | --- | --- |
| Gate 1 — ICPs | pending | Promotion of `icp-006` or any hypothesis-only ICP into core | Should the partner/reseller area be researched, or stay excluded under W-002? |
| Gate 2 — Jobs | pending | Promotion of grade-C jobs `job-005`, `job-008`, `job-009` | Are these three jobs worth core measurement on title-only proxy evidence? |
| Gate 3 — Partitions | pending | Core membership of `cell-007`, `cell-012`, `cell-013`; the three quarantines | Do the grade-C cells stay in core or move to rotating? Is en-CA worth a reviewer? |
| Gate 4 — Panel | pending | Weights, limitations wording, cadence, claim language, concrete surfaces, frozen version | Are equal exposure weights accepted? Which providers fill the four surfaces? Is the variance pilot funded? |

### What would change this panel

1. **An authorised customer or query corpus.** This is the single highest-leverage input. It would create the first grade-A evidence, raise observed-language coverage well above 2 of 61, and probably justify 10–20 more unaided cells — closing W-001.
2. **Reachable review and community sources.** Trustpilot, Reddit and the app stores were all blocked at access time. Their language would likely promote `job-005`, `job-008` and `job-009` out of grade C and either confirm or kill the three disputed core cells.
3. **A locale decision.** Naming one non-English market and a market-competent reviewer would open a whole locale stratum. Without that, no claim about non-English visibility is supportable.
4. **Fresh dated evidence for B5.** The current best story anchor describes an event that finished six days before this run. B5 is one retrieval-only cell until that is fixed.
5. **The variance pilot.** Until it runs, repetition counts are defaults and no interval should be published — especially given that 59 of 61 prompts are paraphrases whose wording sensitivity is unmeasured.
6. **Concrete surfaces and a run budget.** Provider selection fixes the configuration hash and the drift policy, and turns the 639-observation estimate into a real plan.

### Freeze blockers

`source_manifest_hash`, all `content_hash`, all `prompt_hash`, `blind_brief_hash` and all configuration/response hashes are `null` because this runtime cannot compute SHA-256. No placeholder digest was written. Real hashes plus all four gate approvals are required before this panel can be called frozen.
