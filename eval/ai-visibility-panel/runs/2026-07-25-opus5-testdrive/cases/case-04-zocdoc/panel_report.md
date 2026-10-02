# AI Visibility Panel — US healthcare booking marketplace (zocdoc.com)

**Status: `provisional_directional`. Not frozen, not representative, not statistically significant, not causal.**
Built 2026-07-25 from one public URL plus a one-line description. No human has approved any gate.

---

## 1. Decision and limits

**Provisional business decision (agent-assumed).** Decide where evidence-backed content and answer-surface work should go so that assistant answers to real care-seeker and practice-owner questions about finding, verifying and booking in-network care surface this marketplace accurately — and so that practice-owner cost questions get correct fee mechanics rather than competitor-authored framing.

You did not supply a decision, estimands, competitors, surfaces, budget or approver. Everything in this section is an assumption stated openly, not a preference recovered from you.

**Estimands in scope** (each with its own denominator, never pooled):

| ID | Estimand | Numerator | Denominator | Does not prove |
|---|---|---|---|---|
| est-001 | `unaided_brand_presence` | Observations naming the target | Valid **unaided** observations (28 cells / 56 strings), split by lane × surface × wave | Awareness, market share, reach |
| est-002 | `competitive_mention_share` | Target mentions among discovery/acquisition brands named | All such brand mentions in the same unaided observations | Category market share |
| est-003 | `citation_presence` | Observations citing a target-owned domain | Valid **retrieval-lane** observations where retrieval ran and citations were exposed | Traffic, clicks, ranking |
| est-004 | `aided_brand_knowledge` | B0 answers materially accurate on 4 declared checkable facts | Valid **B0 target-aided** observations only (4 cells / 8 strings) | Demand or preference |
| est-005 | `answer_framing` | Target-mentioning observations coded into declared frames | Observations mentioning the target, split by aided status and lane | Sentiment prevalence among real users |

**Out of scope:** `campaign_response`. No campaign, campaign terms or pre-registration were supplied, so the control partition is deliberately empty and no causal claim is available.

**Target population.** US insured care-seekers and US independent/small-practice owners and operations staff who put healthcare-access questions to an AI assistant. **The panel does not sample that population.** It samples 35 evidence-supported question cells built from public buyer language.

**Everything below is conditional on this panel:** these 35 cells, these 70 exact strings, six surfaces, en-US only, this configuration, this wave. It generalises to nothing else.

**Hard limits you should not read past:**

- 28 unaided cells is **below** the 30–48 diagnostic floor (two cells were quarantined at QA).
- No variance pilot has run. No precision, interval width or effective sample size can be stated.
- Every subgroup here has fewer than 20 distinct cells, so results must be published as **counts and response excerpts, not percentage leaderboards**.
- Weights are equal by necessity, not by evidence.

---

## 2. Evidence base

14 sources. Class mix: **3 company-asserted, 3 buyer-behaviour, 7 independent, 1 search-proxy, 0 llm-hypothesis.** Grades: **1×A, 6×B(independent/buyer), 7×C.**

| Requirement from the method | Met? | Sources |
|---|---|---|
| Target product/capability pages | Partially | source-001 (fee mechanics), source-002 (provider product) |
| Target pricing/technical page | Yes | source-001 |
| ≥2 independent sources testing claims/category fit | Yes | source-003 (Wikipedia), source-006 (practising clinician), source-011 (US Senate Finance secret-shopper study), source-009 (news) |
| ≥3 buyer-language sources | Yes | source-004 (complaints), source-005 (BBB reviews), source-012 (forum, verbatim), source-013 (search proxy), source-007/008 (procurement-style decision framing) |
| Competitor/category sources | Yes | source-007, source-008 |
| Fresh dated evidence per B5 cell | Yes, thinly | source-009 (2026-04-21), source-010 (2026-03-24) |

**Permissions.** All 14 sources are public. No authorised private data was supplied or used. No personal data and no long copyrighted excerpts are stored anywhere in this run.

**Conflicts kept, not resolved:**

- source-005 (BBB) reports 4.89/5 across 3,042 reviews with uniformly positive recent entries; source-004 (ComplaintsBoard) records in-network mismatches, last-minute cancellations and provider fee disputes. Both are **self-selected review populations**. Neither is prevalence. Both are retained.
- source-003 records that provider counts reportedly rose ~50% after the per-booking transition, which sits against the owner-cost complaints in source-006/007/008. Unresolved.

**Evidence weaknesses you should weigh before funding anything:**

1. **The supplied URL was unreachable.** `zocdoc.com` root, `/about/whatwedo/`, `/about/news/*` all returned HTTP 403 to the research agent. Target factual standing rests on two reachable target-owned pages plus Wikipedia. The patient-side product surface and its specialty taxonomy were never directly captured.
2. **The two richest provider-side sources are competitor-published** (source-007 DentalVitals, source-008 Emitrr). The `$35–$110` per-new-patient range is a *claim from commercially motivated publishers*, corroborated only loosely by a practising clinician's "close to $70". It is graded C and never used alone.
3. **The strongest care-seeker language is five years old.** source-012 is verbatim with provenance (grade A) but dated 2021-07-19. Cells derived from it carry a decay flag.
4. **The only AI-use prevalence figures are target-commissioned** (source-010, Censuswide, n=1,186 adults + 1,000 providers, fielded Feb 2026). Usable as a dated market-event anchor; **not** independent prevalence.
5. **source-014 (2021-12-03) fails the B5 freshness test** and is explicitly barred from anchoring any B5 cell. It is retained as category context only.

---

## 3. ICPs and buyer jobs

Five ICPs. **Three supported, two `hypothesis_only`.**

| ICP | Label | Status | Trigger | Key constraint | Confidence |
|---|---|---|---|---|---|
| icp-001 | Insured care-seeker needing a genuinely in-network, reachable, soon-available provider | supported | Relocation/plan change; listing contradicted by the front desk | Coverage is a hard requirement; directory can't be trusted without a second check | medium |
| icp-002 | Care-seeker working a stale mental-health directory with effectively closed panels | supported | 33% of listings inaccurate/unreachable; >80% effectively unavailable (source-011) | Cannot make repeated business-hours calls | medium |
| icp-003 | Independent/small-practice owner buying new-patient volume, judged on cost per acquisition | supported | Unfilled capacity; a fee lands for a no-show; a returning patient billed as new | Fee charged at booking, not attendance | medium |
| icp-004 | Multi-location group / health system / DSO evaluating patient-access infrastructure | **hypothesis_only** | Calls exceed staff capacity | EHR integration | low |
| icp-005 | Self-pay / uninsured care-seeker needing a price before booking | **hypothesis_only** | Needs near-term care without usable coverage | Price transparency before booking | low |

**Why icp-004 and icp-005 stayed hypotheses.** icp-004 comes entirely from the target's own provider page — partner logos and feature copy. Company copy establishes standing, not demand. icp-005 comes only from search-result *composition*, which is a proxy for what publishers target, not for what care-seekers ask. Promoting either would be persona fiction.

**Roles kept separate:** care-seeker (role-001), self-pay care-seeker (role-002), owner-clinician (role-003), practice operations manager (role-004). **No caregiver/proxy-booker role was created** — plausible and probably material, but zero evidence was found. That is a recorded gap, not an omission.

**Eight buyer jobs. Six supported, two hypothesis_only.**

| Job | Statement (abbreviated) | Acts | Journey | Grade | Status |
|---|---|---|---|---|---|
| job-001 | Get an appointment booked with someone my plan actually covers, near enough, soon enough | plan, navigate, recommend, troubleshoot | exploration → supplier_selection → post_purchase | B | supported |
| job-002 | Confirm the listing is true *before* the visit so I don't get an out-of-network bill | verify, explain, plan | requirements_building, supplier_selection | B | supported |
| job-003 | Actually reach a mental-health provider when the list is dead ends | diagnose, plan, troubleshoot, compare, recommend | problem_identification → supplier_selection | **A** | supported |
| job-004 | Be seen quickly without a referral and without an unknown price | troubleshoot, plan, recommend, navigate | problem_identification, supplier_selection | B | supported |
| job-005 | Decide which acquisition channels are worth paying for, and what a new patient can cost | compare, verify, plan, recommend, explain | problem_identification → supplier_selection | B | supported |
| job-006 | Get wrong acquisition charges reversed and cap the monthly total | troubleshoot, explain, implement, diagnose | adoption, post_purchase | B | supported |
| job-007 | Stop losing bookings to hold times and voicemail | plan, implement, compare | requirements_building, adoption | C | **hypothesis_only** |
| job-008 | Understand what patients' AI use changes about how care is found and booked | explain | problem_identification, exploration | C | **hypothesis_only** |

**Authentic language preserved** (short spans only): *"either full or have these weird ass times"*, *"portals just dont have updated contact details"*, *"don't return the call and have full voicemails"* (source-012); *"should not charge a fee for their mistake"*, *"cancelled, giving no reason"* (source-004); *"lowest blended cost per new patient"*, *"how much capacity do you actually have to fill"* (source-007); *"whether or not the client arrives or doesn't"* (source-006).

**Negatives and post-purchase cases kept:** cancellation recourse (cell-006), fee disputes (cell-014, cell-028), miscategorised bookings (cell-028).

---

## 4. Comprehensive prompt list

All 80 generated strings, including the 10 that did not make the panel. **Multi-turn scripts show both turns separated by `||`.**

**Legend.** Aided: `U`=unaided, `CAT`=category-aided, `COMP`=competitor-aided, `TGT`=target-aided. Campaign-exposed is `N` on every row (no campaign registered). Acts: `dx`=diagnose, `ver`=verify, `pln`=plan, `tsh`=troubleshoot, `cmp`=compare, `rec`=recommend, `nav`=navigate, `imp`=implement, `exp`=explain. Journey: `PI`=problem_identification, `EX`=exploration, `RB`=requirements_building, `SS`=supplier_selection, `AD`=adoption, `PP`=post_purchase. Roles: `r1`=insured care-seeker, `r2`=self-pay care-seeker, `r3`=owner-clinician, `r4`=practice ops manager. Turn: `1T`=single, `MT`=scripted multi-turn. Lanes: `CM`=closed_model, `RET`=retrieval, `CON`=consumer_surface. Variant: `OBS`=observed-language, `PAR`=natural paraphrase, `SEN`=sensitivity. Weight: `eq/eq*` = equal exposure weight (no exposure evidence) / equal priority weight (provisional, pending Gate 4) — **two separate components, never one score**. Locale is `en-US / en` on every row.

| Prompt ID | Exact prompt | Var | Partition | Band | Aided | Camp | Job | Act | Jrny | Funnel | Role | Constraints | Expected answer | Turn | Lanes | Evidence | Transformation | Weight | QA |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| prompt-001a | every therapist i call is either full or doesn't call back. what do i actually do to get seen | OBS | core | B3 | U | N | job-003 reach a therapist | dx | PI | TOFU | r1 | limited daytime calling | diagnosis_and_options | 1T | CM/RET/CON | A · s012 | lightly_normalized | eq/eq* | pass |
| prompt-001b | I've left messages with a bunch of therapists and nobody has gotten back to me. How do people actually end up getting an appointment? | PAR | core | B3 | U | N | job-003 | dx | PI | TOFU | r1 | limited daytime calling | diagnosis_and_options | 1T | CM/RET/CON | A · s012,s011 | human_written | eq/eq* | pass |
| prompt-002a | the listing said this doctor was in my network but the office says they're not. why does that keep happening and what do i do | OBS | core | B3 | U | N | job-002 confirm coverage | ver | RB | MOFU | r1 | plan coverage required | verification_and_caveats | 1T | CM/RET/CON | B · s004 | lightly_normalized | eq/eq* | pass |
| prompt-002b | A directory showed the provider as in-network and the front desk told me otherwise. How do I settle this before my visit? | PAR | core | B3 | U | N | job-002 | ver | RB | MOFU | r1 | plan coverage required | verification_and_caveats | 1T | CM/RET/CON | B · s004,s011 | human_written | eq/eq* | pass |
| prompt-003a | just moved to a new state and need to find a primary care doctor that takes my plan. where do i even start | OBS | core | B3 | U | N | job-001 book covered care | pln | EX | TOFU | r1 | plan coverage required | step_plan | 1T | CM/RET/CON | B · s013 | search_query_expanded | eq/eq* | pass |
| prompt-003b | I relocated for work and need to set up a new primary care doctor that my insurance covers. What's the process? | PAR | core | B3 | U | N | job-001 | pln | EX | TOFU | r1 | plan coverage required | step_plan | 1T | CM/RET/CON | B · s013,s005 | human_written | eq/eq* | pass |
| prompt-004a | tooth pain got bad overnight and my regular dentist has nothing until next month. how do i get seen today | OBS | core | B3 | U | N | job-004 be seen fast | tsh | PI | null | r1 | same-day need | troubleshooting_steps | 1T | CM/RET/CON | B · s013 | search_query_expanded | eq/eq* | pass — sole variant on this cell |
| prompt-004b | Severe tooth pain and my dentist is booked out for weeks. What app should I use to get an appointment today? | PAR | — | B3 | U | N | job-004 | tsh | PI | null | r1 | same-day need | — | 1T | — | D | llm_expanded | — | **reject** — recommendation forcing |
| prompt-005a | TURN 1: half the numbers on my plan's therapist list are disconnected or the mailbox is full. is there a better list somewhere \|\| TURN 2: tried that, still nothing. what else is there | OBS | core | B3 | U | N | job-003 | tsh | EX | TOFU | r1 | limited daytime calling | troubleshooting_steps | MT | CM/RET/CON | A · s012 | lightly_normalized | eq/eq* | pass |
| prompt-005b | TURN 1: The phone numbers in my plan's provider directory don't work. How do I find a mental health provider who's actually taking patients? \|\| TURN 2: I've called about ten of them now. What's the next thing to try? | PAR | core | B3 | U | N | job-003 | tsh | EX | TOFU | r1 | limited daytime calling | troubleshooting_steps | MT | CM/RET/CON | A · s012,s011 | human_written | eq/eq* | pass |
| prompt-006a | my appointment got cancelled the day of, giving no reason. do i have any recourse here | OBS | core | B3 | U | N | job-001 | tsh | PP | null | r1 | — | troubleshooting_steps | 1T | CM/RET/CON | B · s004 | lightly_normalized | eq/eq* | pass |
| prompt-006b | The office cancelled my confirmed booking a few hours before the visit and didn't explain why. What are my options now? | PAR | core | B3 | U | N | job-001 | tsh | PP | null | r1 | — | troubleshooting_steps | 1T | CM/RET/CON | B · s004 | human_written | eq/eq* | pass |
| prompt-007a | i want to get a cleaning and checkup on the books this month using my dental plan. how do i make that happen | OBS | core | B4 | U | N | job-001 | pln | EX | TOFU | r1 | plan coverage required | step_plan | 1T | CM/RET/CON | B · s005 | lightly_normalized | eq/eq* | pass |
| prompt-007b | Trying to get a dental visit scheduled before the end of the month that my plan will actually cover. What's the fastest way to line one up? | PAR | core | B4 | U | N | job-001 | pln | EX | TOFU | r1 | plan coverage required | step_plan | 1T | CM/RET/CON | B · s005,s013 | human_written | eq/eq* | pass |
| prompt-008a | how do i make sure i don't get a surprise out of network bill from a first visit | OBS | core | B4 | U | N | job-002 | pln | RB | MOFU | r1 | plan coverage required | step_plan | 1T | CM/RET/CON | B · s004 | search_query_expanded | eq/eq* | pass |
| prompt-008b | What should I check before a first appointment so I'm not hit with an out-of-network charge afterwards? | PAR | core | B4 | U | N | job-002 | pln | RB | MOFU | r1 | plan coverage required | step_plan | 1T | CM/RET/CON | B · s004,s011 | human_written | eq/eq* | pass — lexical pair retained, framing differs |
| prompt-009a | i want to actually be in therapy in the next few weeks, not still searching. how do i get there | OBS | core | B4 | U | N | job-003 | pln | EX | TOFU | r1 | — | step_plan | 1T | CM/RET/CON | B · s012 | lightly_normalized | eq/eq* | pass |
| prompt-009b | What's a realistic way to go from looking for a therapist to having a first session within a month? | PAR | core | B4 | U | N | job-003 | pln | EX | TOFU | r1 | — | step_plan | 1T | CM/RET/CON | B · s012,s011 | human_written | eq/eq* | pass |
| prompt-010a | best way to find and book a doctor who takes my insurance | OBS | core | B1 | U | N | job-001 | rec | SS | BOFU | r1 | plan coverage required | ranked_shortlist | 1T | CM/RET/CON | B · s013 | search_query_expanded | eq/eq* | pass — canonical survivor of exact merge; sole variant |
| prompt-010b | Best way to find and book a doctor who takes my insurance? | PAR | — | B1 | U | N | job-001 | rec | SS | BOFU | r1 | — | — | 1T | — | D | llm_expanded | — | **reject** — exact normalized duplicate (merge_exact) |
| prompt-011a | TURN 1: what are the different ways to find a therapist who takes insurance, and which ones actually work \|\| TURN 2: which of those is fastest if i need someone in the next two weeks | OBS | core | B1 | U | N | job-003 | cmp | SS | BOFU | r1 | plan coverage required | comparison_of_routes | MT | CM/RET/CON | B · s012,s011 | lightly_normalized | eq/eq* | pass |
| prompt-011b | TURN 1: Compare the realistic options for finding an in-network therapist. \|\| TURN 2: Now narrow it to whichever one gets me seen soonest. | PAR | core | B1 | U | N | job-003 | cmp | SS | BOFU | r1 | plan coverage required | comparison_of_routes | MT | CM/RET/CON | B · s011 | human_written | eq/eq* | pass |
| prompt-012a | what's the fastest way to get a same day appointment with a doctor | OBS | core | B1 | U | N | job-004 | rec | SS | BOFU | r1 | same-day need | ranked_shortlist | 1T | CM/RET/CON | B · s013 | search_query_expanded | eq/eq* | pass |
| prompt-012b | If I need to see someone today, what are my actual options and which one usually works? | PAR | core | B1 | U | N | job-004 | rec | SS | BOFU | r1 | same-day need | ranked_shortlist | 1T | CM/RET/CON | B · s013 | human_written | eq/eq* | pass |
| prompt-013a | our schedule has holes every week and new patient calls are down. what's actually causing that | OBS | core | B3 | U | N | job-005 choose channels | dx | PI | TOFU | r3 | — | diagnosis_and_options | 1T | CM/RET/CON | B · s007 | lightly_normalized | eq/eq* | pass |
| prompt-013b | I run a small practice and we're not filling the schedule with new patients anymore. How do I work out where the drop-off is? | PAR | core | B3 | U | N | job-005 | dx | PI | TOFU | r3 | — | diagnosis_and_options | 1T | CM/RET/CON | B · s007,s008 | human_written | eq/eq* | pass — pair with 015a retained (different role/journey) |
| prompt-014a | TURN 1: i'm paying an acquisition fee for every new patient booking whether or not the patient arrives or doesn't. is that normal \|\| TURN 2: what can i actually do about it | OBS | core | B3 | U | N | job-006 fix wrong charges | tsh | PP | null | r3 | new vs existing patient billing | troubleshooting_steps | MT | CM/RET/CON | B · s006 | lightly_normalized | eq/eq* | pass |
| prompt-014b | TURN 1: We get charged for each new patient booking regardless of whether they show up. Is that standard for this kind of channel? \|\| TURN 2: How do practices usually handle that? | PAR | core | B3 | U | N | job-006 | tsh | PP | null | r3 | new vs existing patient billing | troubleshooting_steps | MT | CM/RET/CON | B · s006,s004 | human_written | eq/eq* | pass |
| prompt-015a | our new patient acquisition bill is different every month and i can't forecast it. how do i get control of that | OBS | core | B3 | U | N | job-006 | dx | PP | null | r4 | monthly spend must be capped | diagnosis_and_options | 1T | CM/RET/CON | B · s008 | lightly_normalized | eq/eq* | pass — sole variant on this cell |
| prompt-015b | The per-booking charges swing wildly month to month and the rate card isn't published anywhere. How do office managers budget for this? | PAR | — | B3 | U | N | job-006 | dx | PP | null | r4 | — | — | 1T | — | C | llm_expanded | — | **revise** — embeds a contested single-source claim as premise |
| prompt-016a | how do i get to the lowest blended cost per new patient across all the channels we use | OBS | core | B4 | U | N | job-005 | pln | RB | MOFU | r3 | — | step_plan | 1T | CM/RET/CON | B · s007 | lightly_normalized | eq/eq* | pass — sole variant on this cell |
| prompt-016b | Is pay-per-booking the cheapest way to get new patients, or is there something better? | PAR | — | B4 | U | N | job-005 | pln | RB | MOFU | r3 | — | — | 1T | — | D | llm_expanded | — | **reject** — target proprietary-category term in unaided core (hard contamination failure) |
| prompt-017a | what is my average new patient lifetime value and how do i actually work it out | OBS | core | B4 | U | N | job-005 | ver | RB | MOFU | r3 | — | calculation_method_and_caveats | 1T | CM/RET/CON | B · s007 | lightly_normalized | eq/eq* | pass |
| prompt-017b | What's a sensible way to calculate what a new patient is worth so I know how much I can afford to spend acquiring one? | PAR | core | B4 | U | N | job-005 | ver | RB | MOFU | r3 | — | calculation_method_and_caveats | 1T | CM/RET/CON | B · s007 | human_written | eq/eq* | pass |
| prompt-018a | what are the alternatives to paying a fee for every new patient booking | OBS | core | B1 | U | N | job-005 | cmp | SS | BOFU | r3 | — | comparison_of_routes | 1T | CM/RET/CON | B · s007,s008 | lightly_normalized | eq/eq* | pass |
| prompt-018b | Other than channels that bill per new patient appointment, what else reliably fills a practice schedule? | PAR | core | B1 | U | N | job-005 | cmp | SS | BOFU | r3 | — | comparison_of_routes | 1T | CM/RET/CON | B · s007 | human_written | eq/eq* | pass |
| prompt-019a | solo therapy practice - is it worth listing on a booking marketplace or should i build referrals myself | OBS | core | B1 | U | N | job-005 | rec | SS | BOFU | r3 | solo independent practice | recommendation_with_tradeoffs | 1T | CM/RET/CON | B · s006 | lightly_normalized | eq/eq* | pass |
| prompt-019b | I'm a solo clinician deciding between paying for marketplace listings and building my own referral pipeline. Which makes more sense? | PAR | core | B1 | U | N | job-005 | rec | SS | BOFU | r3 | solo independent practice | recommendation_with_tradeoffs | 1T | CM/RET/CON | B · s006 | human_written | eq/eq* | pass |
| prompt-020a | more of my patients are showing up having already asked an AI assistant about their symptoms. how should i handle that in the visit | OBS | rotating | B5 | U | N | job-008 AI-shifted discovery | exp | PI | TOFU | r3 | — | trend_explanation | 1T | RET/CON | C · s010 (2026-03-24, review by 2026-09-24) | lightly_normalized | eq/eq* | pass |
| prompt-020b | Patients are arriving with AI-generated theories about their condition. What's the current thinking on how clinicians should respond? | PAR | rotating | B5 | U | N | job-008 | exp | PI | TOFU | r3 | — | trend_explanation | 1T | RET/CON | C · s010 (review by 2026-09-24) | human_written | eq/eq* | pass |
| prompt-021a | patients are starting to book appointments straight from review sites and map listings. what does that change for a practice | OBS | rotating | B5 | U | N | job-008 | exp | EX | TOFU | r4 | — | trend_explanation | 1T | RET/CON | B · s009 (2026-04-21, review by 2026-10-21) | lightly_normalized | eq/eq* | pass |
| prompt-021b | Appointment booking is moving into search, maps and review platforms this year. What should a practice be doing about it? | PAR | rotating | B5 | U | N | job-008 | exp | EX | TOFU | r4 | — | trend_explanation | 1T | RET/CON | B · s009 (review by 2026-10-21) | human_written | eq/eq* | pass |
| prompt-022a | how do i put a hard ceiling on what my practice spends on new patient acquisition each month | OBS | rotating | B4 | U | N | job-006 | imp | AD | null | r3 | monthly spend must be capped | implementation_steps | 1T | CM/RET/CON | C · s008 | lightly_normalized | eq/eq* | pass — grade C, rotating only |
| prompt-022b | Is there a way to cap monthly patient acquisition spend so it can't run away from me? | PAR | rotating | B4 | U | N | job-006 | imp | AD | null | r3 | monthly spend must be capped | implementation_steps | 1T | CM/RET/CON | C · s001 | human_written | eq/eq* | pass — grade C, rotating only |
| prompt-023a | no insurance right now - how do i find out what a visit will cost before i book it | OBS | rotating | B4 | U | N | job-004 | pln | SS | null | r2 | paying out of pocket | cost_estimate_and_caveats | 1T | — | C · s013 | search_query_expanded | — | **quarantine** — icp-005 is hypothesis_only |
| prompt-023b | I'm paying out of pocket. How do I get a price for an appointment up front instead of after the visit? | PAR | rotating | B4 | U | N | job-004 | pln | SS | null | r2 | paying out of pocket | cost_estimate_and_caveats | 1T | — | C · s013,s004 | human_written | — | **quarantine** — same; cell contributes nothing this wave |
| prompt-024a | we're losing new patient calls because nobody can pick up the phone after hours. what do practices do about that | OBS | rotating | B3 | U | N | job-007 call capacity | dx | PI | TOFU | r4 | — | diagnosis_and_options | 1T | — | C · s002 | llm_expanded | — | **quarantine** — job derives only from a target capability page |
| prompt-024b | How do practices handle inbound scheduling calls they can't answer during and after business hours? | PAR | rotating | B3 | U | N | job-007 | dx | PI | TOFU | r4 | — | diagnosis_and_options | 1T | — | C · s002 | llm_expanded | — | **quarantine** — same |
| prompt-025a | i have medicaid and every office i call says they're not taking new medicaid patients. how do i find one that is | OBS | sentinel | B3 | U | N | job-002 | ver | SS | BOFU | r1 | Medicaid coverage | verification_and_caveats | 1T | CM/RET/CON | B · s011 | lightly_normalized | eq/eq* | pass — constraint protected from merge with cell-002 |
| prompt-025b | Finding a provider who actually accepts Medicaid and has openings has been impossible. What actually works? | PAR | sentinel | B3 | U | N | job-002 | ver | SS | BOFU | r1 | Medicaid coverage | verification_and_caveats | 1T | CM/RET/CON | B · s011,s004 | human_written | eq/eq* | pass |
| prompt-025c | medicaid, need a provider taking new patients. whats the move | SEN | sentinel | B3 | U | N | job-002 | ver | SS | BOFU | r1 | Medicaid coverage | verification_and_caveats | 1T | CM/RET/CON | B · s011 | human_written | eq/eq* | pass — terse-wording sensitivity axis |
| prompt-026a | small town practice - how do i get new patients when there isn't much local search volume | OBS | sentinel | B4 | U | N | job-005 | pln | RB | MOFU | r3 | rural / small population | step_plan | 1T | CM/RET/CON | B · s007 | lightly_normalized | eq/eq* | pass |
| prompt-026b | I run a practice in a rural market. What actually brings in new patients when the local population is small? | PAR | sentinel | B4 | U | N | job-005 | pln | RB | MOFU | r3 | rural / small population | step_plan | 1T | CM/RET/CON | B · s007,s001 | human_written | eq/eq* | pass |
| prompt-026c | rural practice, low population, need new patients. which channels are worth it | SEN | sentinel | B4 | U | N | job-005 | pln | RB | MOFU | r3 | rural / small population | step_plan | 1T | CM/RET/CON | B · s007 | human_written | eq/eq* | pass — pair with 026a retained as intended sensitivity axis |
| prompt-027a | i can only do video appointments. what's the best way to find a provider who takes my insurance and does telehealth | OBS | sentinel | B1 | U | N | job-003 | rec | SS | BOFU | r1 | telehealth only + coverage | ranked_shortlist | 1T | CM/RET/CON | B · s013 | search_query_expanded | eq/eq* | pass |
| prompt-027b | I need an in-network provider who sees patients over video. How should I go about finding one? | PAR | sentinel | B1 | U | N | job-003 | rec | SS | BOFU | r1 | telehealth only + coverage | ranked_shortlist | 1T | CM/RET/CON | B · s013,s012 | human_written | eq/eq* | pass |
| prompt-027c | telehealth only, in network. how do i book | SEN | sentinel | B1 | U | N | job-003 | rec | SS | BOFU | r1 | telehealth only + coverage | ranked_shortlist | 1T | CM/RET/CON | B · s013 | human_written | eq/eq* | pass |
| prompt-028a | we got billed a new patient fee for someone who's been our patient for years. they should not charge a fee for their mistake - how do we get it reversed | OBS | sentinel | B3 | U | N | job-006 | tsh | PP | null | r4 | existing billed as new + short dispute window | troubleshooting_steps | 1T | CM/RET/CON | B · s004 | lightly_normalized | eq/eq* | pass |
| prompt-028b | A returning patient was billed to us as a new patient acquisition. What's the process to dispute that, and how long do we have? | PAR | sentinel | B3 | U | N | job-006 | tsh | PP | null | r4 | existing billed as new + short dispute window | troubleshooting_steps | 1T | CM/RET/CON | B · s001,s004 | human_written | eq/eq* | pass |
| prompt-028c | existing patient billed as new. how to dispute the fee | SEN | sentinel | B3 | U | N | job-006 | tsh | PP | null | r4 | existing billed as new + short dispute window | troubleshooting_steps | 1T | CM/RET/CON | B · s001 | human_written | eq/eq* | pass |
| prompt-029a | can i see a specialist without a referral, and how do i set that up | OBS | sentinel | B4 | U | N | job-001 | nav | RB | MOFU | r1 | no referral in hand | navigation_instructions | 1T | CM/RET/CON | B · s013 | search_query_expanded | eq/eq* | pass |
| prompt-029b | How do I book a dermatologist directly without getting a referral first, and will insurance still cover it? | PAR | — | B4 | U | N | job-001 | nav | RB | MOFU | r1 | — | — | 1T | — | D | llm_expanded | — | **revise** — introduces an unevidenced specialty (split, not merged) |
| prompt-029c | specialist appointment without a referral - possible? | SEN | sentinel | B4 | U | N | job-001 | nav | RB | MOFU | r1 | no referral in hand | navigation_instructions | 1T | CM/RET/CON | B · s013 | human_written | eq/eq* | pass |
| prompt-030a | is it a good idea to ask an ai assistant about symptoms before deciding whether to book a doctor | OBS | sentinel | B5 | U | N | job-008 | exp | EX | TOFU | r1 | — | trend_explanation | 1T | RET/CON | C · s010 (review by 2026-09-24) | lightly_normalized | eq/eq* | pass |
| prompt-030b | I looked my symptoms up with an AI first. Should I still book an appointment, and should I mention that I checked? | PAR | sentinel | B5 | U | N | job-008 | exp | EX | TOFU | r1 | — | trend_explanation | 1T | RET/CON | C · s010 (review by 2026-09-24) | human_written | eq/eq* | pass |
| prompt-030c | should i tell my doctor i used ChatGPT to look up my symptoms | SEN | sentinel | B5 | U | N | job-008 | exp | EX | TOFU | r1 | — | trend_explanation | 1T | — | C · s010 | human_written | — | **quarantine** — names a third-party assistant inside the instrument; Gate 3 decision |
| prompt-031a | which online doctor appointment booking sites are actually worth using | OBS | aided | B2 | CAT | N | job-001 | cmp | EX | TOFU | r1 | plan coverage required | comparison_of_routes | 1T | CM/RET/CON | B · s013 | search_query_expanded | eq/eq* | pass — separate aided denominator |
| prompt-031b | Are appointment booking apps reliable for checking whether a doctor is in-network and has real availability? | PAR | aided | B2 | CAT | N | job-001 | cmp | EX | TOFU | r1 | plan coverage required | comparison_of_routes | 1T | CM/RET/CON | B · s005,s004 | human_written | eq/eq* | pass |
| prompt-032a | which patient acquisition marketplaces do dental practices use, and how do they compare on cost | OBS | aided | B2 | CAT | N | job-005 | cmp | RB | MOFU | r3 | — | comparison_table | 1T | CM/RET/CON | C · s007 | lightly_normalized | eq/eq* | pass |
| prompt-032b | Compare the main online marketplaces that send new patients to dental practices. | PAR | aided | B2 | CAT | N | job-005 | cmp | RB | MOFU | r3 | — | comparison_table | 1T | CM/RET/CON | C · s007,s008 | human_written | eq/eq* | pass |
| prompt-033a | how do Opencare, Solv and NexHealth compare for bringing new patients into a practice | OBS | aided | B1 | COMP | N | job-005 | cmp | SS | BOFU | r3 | — | comparison_table | 1T | CM/RET/CON | C · s007 | lightly_normalized | eq/eq* | pass — competitor terms permitted only under declared exception |
| prompt-033b | For a dental practice choosing between Opencare, 1-800-Dentist and Solv, what are the trade-offs? | PAR | aided | B1 | COMP | N | job-005 | cmp | SS | BOFU | r3 | — | comparison_table | 1T | CM/RET/CON | C · s007 | human_written | eq/eq* | pass — distinct competitor triad, retained |
| prompt-034a | how much does Zocdoc charge a practice per new patient booking | OBS | aided | B0 | TGT | N | job-005 | ver | RB | MOFU | r3 | — | factual_answer | 1T | CM/RET/CON | C · s008 | search_query_expanded | eq/eq* | pass — B0 exception; est-004 denominator |
| prompt-034b | What does listing on Zocdoc actually cost a small practice, and what exactly triggers a charge? | PAR | aided | B0 | TGT | N | job-005 | ver | RB | MOFU | r3 | — | factual_answer | 1T | CM/RET/CON | C · s001,s006 | human_written | eq/eq* | pass |
| prompt-035a | a Zocdoc patient no-showed and we were still charged. can we get that credited | OBS | aided | B0 | TGT | N | job-006 | tsh | PP | null | r3 | short dispute window | factual_answer | 1T | CM/RET/CON | C · s004,s001 | lightly_normalized | eq/eq* | pass |
| prompt-035b | What is Zocdoc's policy on booking fees when a patient cancels or doesn't show up? | PAR | aided | B0 | TGT | N | job-006 | tsh | PP | null | r3 | short dispute window | factual_answer | 1T | CM/RET/CON | C · s001 | human_written | eq/eq* | pass |
| prompt-036a | is Zocdoc free for patients and is it legit | OBS | aided | B0 | TGT | N | job-001 | exp | EX | TOFU | r1 | — | factual_answer | 1T | CM/RET/CON | C · s003 | search_query_expanded | eq/eq* | pass |
| prompt-036b | Does Zocdoc charge patients anything to book an appointment, and if not, who pays for it? | PAR | aided | B0 | TGT | N | job-001 | exp | EX | TOFU | r1 | — | factual_answer | 1T | CM/RET/CON | C · s003,s001 | human_written | eq/eq* | pass |
| prompt-037a | is the insurance information on Zocdoc accurate, or should i still call the office to confirm | OBS | aided | B0 | TGT | N | job-002 | ver | SS | MOFU | r1 | plan coverage required | verification_and_caveats | 1T | CM/RET/CON | B · s004 | lightly_normalized | eq/eq* | pass |
| prompt-037b | How reliable is Zocdoc's in-network filter? Do I still need to verify coverage myself before the visit? | PAR | aided | B0 | TGT | N | job-002 | ver | SS | MOFU | r1 | plan coverage required | verification_and_caveats | 1T | CM/RET/CON | B · s004,s011 | human_written | eq/eq* | pass |

---

## 5. Coverage matrix

**In panel: 35 canonical cells / 70 exact strings.** Generated: 37 cells / 80 strings.

**By band** (accepted cells): B0 = 4 (aided), B1 = 6 (5 unaided + 1 competitor-aided), B2 = 2 (category-aided), B3 = 12 unaided, B4 = 8 unaided, B5 = 3 unaided.

**By aided status:** unaided 28 cells / 56 strings · category_aided 2 / 4 · competitor_aided 1 / 2 · target_aided 4 / 8. **campaign_exposed = false on all 35.**

**By job:** job-001 6 cells · job-002 4 · job-003 5 · job-004 2 · job-005 8 · job-006 5 · job-007 0 (quarantined) · job-008 3.

**By information act:** troubleshoot 6 · plan 6 · verify 5 · compare 5 · recommend 4 · diagnose 3 · explain 4 · navigate 1 · implement 1. **buy 0, generate 0 — waived, not entailed by any job.**

**By journey:** problem_identification 4 · exploration 8 · requirements_building 8 · supplier_selection 9 · adoption 1 · post_purchase 5.

**By funnel:** TOFU 11 · MOFU 8 · BOFU 10 · null 6. Funnel was never inferred from band or keyword; cell-004 is an urgent B3 close to transaction and is deliberately null.

**By role:** r1 insured care-seeker 16 · r2 self-pay 0 (quarantined) · r3 owner-clinician 12 · r4 ops manager 7.

**By locale:** en-US 35. **es-US 0.**

**By lane:** closed_model 32 cells (B5 excluded) · retrieval 35 · consumer_surface 35 · campaign_experiment 0.

**By partition:** core 19 · rotating 3 · sentinel 6 · control 0 · aided 7.

**By evidence grade (accepted strings):** A 6 · B 44 · C 20 · **D 0**. No grade-D string is in core or anywhere in the panel.

**By turn form:** single_turn 32 cells · scripted_multi_turn 3 cells.

**By variant role:** observed_language 33 · natural_paraphrase 31 · sensitivity 6.

### Required waivers and gaps

| Waiver | Dimension | Reason | What would close it |
|---|---|---|---|
| waiver-001 | `control` partition empty | No registered campaign; nothing to control for | Register a campaign, then add 24–40 matched unaffected controls |
| waiver-002 | Sentinel partition is 6, below the 12–20 pilot recommendation | Diagnostic budget | Pilot borrows 6 core cells to reach 12; documented deviation |
| waiver-002b | **28 unaided cells, below the 30–48 diagnostic floor** | Two rotating cells quarantined at QA | Promote icp-005 / job-007 with real evidence, or add 2–4 new evidence-backed unaided cells |
| waiver-003 | `es-US` absent | No Spanish-language evidence, no native reviewer; machine translation prohibited | Spanish-language buyer sources + a market-competent reviewer |
| waiver-004 | Acts `buy` and `generate` absent | Not entailed by any job; booking is free to the care-seeker | Evidence of transactional or artefact-generation query patterns — likely permanently out of scope |
| waiver-005 | Four core cells carry one variant | QA rejections/revisions | Regenerate paraphrases for cell-004, cell-010, cell-015, cell-016 |
| gap | Caregiver / proxy-booker role | Zero evidence found | Forum or support evidence of booking on another's behalf |
| gap | Enterprise economic buyer and approver roles | Only target-asserted partner logos | Independent case study, RFP or procurement evidence |
| gap | B5 rests on two 2026 sources, one target-commissioned | Thin fresh evidence | Independent dated coverage of healthcare-discovery shifts |
| gap | `competitor_aided` is one cell from a competitor-published list | Weak competitor sourcing | An unaffiliated category source |

---

## 6. QA ledger

**80 candidates → 70 pass · 2 revise · 5 quarantine · 3 reject.**

**Contamination failures by class:** proprietary_categories **1** (prompt-016b, `pay-per-booking` in an unaided core candidate — hard failure, rejected). Brands 0 · products 0 · domains 0 · people 0 · slogans 0 · campaign terms 0 · flattering claims 0 · competitor terms in unaided 0. **Unaided core leakage after rejection: zero.**

**Rejections:** prompt-004b (recommendation forcing) · prompt-010b (exact normalized duplicate) · prompt-016b (contamination).
**Revisions routed back, not silently repaired:** prompt-015b (contested claim as premise) · prompt-029b (unevidenced specialty).
**Quarantines held for Gate 3:** prompt-023a/b (icp-005 hypothesis_only) · prompt-024a/b (feature-inverted job) · prompt-030c (names a third-party assistant inside the instrument).

**Duplicates:** 1 exact merge, 0 semantic merges, 1 split, 78 variants retained. **Semantic near-duplicate detection was not machine-run** — no fixed embedding model/version was declared, so `sem-dup-review` is `review`, not `pass`, on every candidate. Nothing was auto-deleted.

**Protected differences (not merged):** commercial plan vs Medicaid (cell-002 / cell-025) · no-show fee vs existing-patient miscategorisation (cell-014 / cell-028) · solo practice vs rural market (cell-019 / cell-026) · owner vs ops manager (cell-013 / cell-015).

**Coverage damage from QA:** four core cells now carry one variant; the self-pay persona and the call-capacity job are entirely absent; unaided cells fell from 30 to 28.

**Blinding.** The unaided pass worked only from `blind_design_brief.json` plus the architecture, but it ran **in the same context as the research**, not in a fresh subagent. One leak reached a core candidate and was caught. Residual risk is **medium**. Recommended Gate 3 action: regenerate the 19 core unaided cells in a fresh context and diff.

**No visibility result influenced any selection.** None exists for this panel.

**Known artifact defects blocking freeze:** (1) `prompt_architecture.json` spec-001 carries `job_id: "job-001"` with a disambiguating `job_id_authoritative: "job-003"`; job-003 is correct and is what the universe uses. (2) All SHA-256 hashes are `null` — no hashing runtime was available in this session.

---

## 7. Tracking plan

Full detail is in `tracking_plan.md`. In brief: 2 variants per cell (3 on sentinels), 3 repeats per candidate per wave (6 on sentinels), six surfaces across three lanes, fresh session per observation, retrieval state recorded rather than assumed, randomised presentation order under seed 20260725, equal exposure and equal priority weights reported **separately**, Wilson intervals for simple strata and a cell-clustered bootstrap for aggregates, monthly waves, quarterly review, next review 2026-10-25, B5 cells reviewed 2026-09-24 and 2026-10-21.

---

## 8. Human gates

| Gate | Status | Blocking question |
|---|---|---|
| 1 — ICPs | **pending** | Promote or reject icp-004 (enterprise) and icp-005 (self-pay)? Confirm exclusions. |
| 2 — Jobs | **pending** | Confirm 8 jobs. May job-007 enter on target-asserted evidence alone? |
| 3 — Partitions | **pending** | Approve core/aided partitions, 3 rejections, 2 revisions, 5 quarantines — while blind to baseline visibility. Four disputed items are listed in `prompt_qa.json`. |
| 4 — Weights | **pending** | Approve equal weights as a stated limitation, set real priority weights, approve cadence, limitations and version. |

**What would change this panel most:**

1. **A stated business decision.** The priority weight component is a placeholder until someone says whether care-seeker or practice-owner visibility matters more.
2. **Access to the target's own patient-side pages.** The homepage 403 means the specialty taxonomy and patient-facing claims were never captured; that could add or remove whole cells.
3. **Any non-vendor provider-side source.** The fee range driving eight practice-owner cells rests on competitor-published claims.
4. **Fresh care-seeker language.** The strongest verbatim source is from 2021.
5. **Running the 12-cell variance pilot.** Until then repeats are guesses and no precision claim is legitimate.
6. **Promoting or killing icp-005 and job-007**, which would restore the unaided count to the diagnostic floor or confirm the waiver.

Until Gates 1–4 are approved, the hashes are computed and the pilot has run, this stays a candidate panel: a source-cited, evidence-supported prompt list you can start tracking, not a frozen measurement instrument.
