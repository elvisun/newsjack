# AI visibility panel — GB home energy (target: octopus.energy)

**Run date:** 2026-07-25 16:00 -04:00 · **Panel:** `panel-gb-home-energy-visibility-001` v0.1.0 · **Status: `provisional_directional`**

Built from one public URL (`https://octopus.energy/`) plus a one-line description. No approver, no customer data, no variance pilot, no campaign. 18 sources · 6 buyer jobs · 46 canonical intent cells · 72 exact prompts · 70 accepted · 2 quarantined · 2 rejected.

---

## 1. Decision and limits

**Assumed decision (agent-set, needs confirmation):** where should GB home-energy answer-engine effort go over the next two quarters — which buyer jobs, journey states and proximity bands does the brand fail to appear in when a household describes its situation *without* naming a supplier, and which of those gaps justify content, data or PR work?

**Estimands, each with its own denominator** (full numerators in `measurement_charter.json`):

| Estimand | Denominator | Explicitly does not prove |
| --- | --- | --- |
| `unaided_brand_presence` (primary) | valid observations of **unaided** cells within one lane × surface × locale × wave | market share, reach, awareness, preference, revenue |
| `competitive_mention_share` | all approved-supplier-list mention instances in eligible comparison/recommendation observations | share of customers or switchers |
| `citation_presence` | retrieval-lane observations **where citations were exposed** | referral traffic or influence |
| `aided_brand_knowledge` | valid **B0 target-aided** observations only | sentiment, favourability, unaided visibility |
| `answer_framing` | valid observations in the same stratum | that framing changed behaviour |
| `campaign_response` | **deferred — no denominator** | nothing; declared so nobody retro-fits it to an uncontrolled before/after |

**Population:** GB households (England, Scotland, Wales) with a decision matching one of six evidenced jobs. What is actually measured is a **non-probability set of 46 constructed intents** on selected surfaces. Excluded: Northern Ireland and non-GB markets, business/SME and salary-sacrifice buyers, prepayment/arrears/affordability contexts, renters and flats for the heating job, all non-English locales.

**Lanes:** `closed_model`, `retrieval`, `consumer_surface` active; `campaign_experiment` declared **inactive**.

**Status:** provisional and directional. Not frozen, not representative, not statistically significant, not causal. Every finding is **conditional on this panel**, this wording, these surfaces and this wave. No composite "AI visibility score" is defined, and one should not be created.

---

## 2. Evidence base

| Class | n | What it was used for |
| --- | ---: | --- |
| `company_asserted` | 3 | factual standing and the contamination lexicon only (source-001 home page, source-002 export tariffs, source-003 heat pumps) |
| `independent` | 5 | market events and third-party assessment (source-004/005 Ofgem cap, source-006 Citizens Advice Q1 2026, source-007 Which? Jan 2026, source-012 MoneySavingExpert 24 Jul 2026) |
| `buyer_behavior` | 5 | authentic buyer language (source-008 myenergi, source-009/010/011 MoneySavingExpert, source-014 Green Building Forum) |
| `search_proxy` | 4 | question shape and category context (source-013, source-015, source-016, source-017) |
| `llm_hypothesis` | 1 | one rotating discovery cell only (source-018, grade D) |

**Grade ceiling: B.** No grade-A source exists because no first-party customer, interview, support or AI-conversation corpus was supplied. Public forum text and search-result shape are proxies, and a forum post is evidence *from that source*, not evidence of market prevalence. No count in this run is a frequency.

**Conflicts, resolved in favour of primary publishers and retained as counterevidence in source-016:**

- 1 July 2026 price cap: **GBP 1,663** typical dual-fuel direct debit, +13% (Ofgem, 27 May 2026) — several aggregators say GBP 1,862.
- Which? customer score: **79%**, ninth consecutive Recommended Provider year (Which?, 19 Jan 2026) — aggregators say 74%.
- Citizens Advice Q1 2026: **rank 4, 3.67/5** (E.ON Next 3.71, OVO 3.30, Scottish Power 2.94, EDF 2.90, British Gas 2.61) — aggregators say the target "ranks first".

**Permissions and omissions.** All sources are public, unauthenticated pages. The target home page displays first names and some full names of individual customers and support staff; these were deliberately **not** recorded anywhere, including the contamination register. No long copyrighted excerpts were stored.

**Decay.** source-007 fieldwork ran Sept–Oct 2025; source-008 is dated March 2024. Both inform direction and language shape, not current price or current service quality. Five sources (source-011, 014, 015, 016, 017) were read through search-result summaries rather than direct retrieval and carry lower confidence, recorded in their `span_locator`.

**Gaps in the evidence base:** no first-party corpus; no primary confirmation of the claimed heat-pump grant change; no researched evidence for prepayment/affordability, renters, business buyers or Welsh-language demand; only one dated primary event available for B5.

---

## 3. ICPs and buyer jobs

Seven ICP hypotheses; **icp-007 (business / salary-sacrifice) is website-only and produced zero cells** — company copy establishes standing, not demand. icp-004 (heat pump) is `hypothesis_only` at low confidence because its behavioural source was read via summary.

| Job | Statement (abbreviated) | Trigger | Roles | Key constraints | Criteria | Language anchor |
| --- | --- | --- | --- | --- | --- | --- |
| **job-001** fix or stay | default rate moved; choose between locking a price and staying flexible | +13% cap change 1 Jul 2026 | bill payer (r1) | can't absorb volatility; regional rates; exit fees | annual cost, risk, service | "is it time to fix your energy or stay on the Price Cap?" (s012) |
| **job-002** cheapest home EV charging | charging is now the biggest slice of the bill | new EV / rising bill | driver + charge-point owner (r2) | charger/car must be controllable; export deal must survive; battery losses | effective p/kWh, compatibility | "anyone with similar installation seeing advantage" (s008) |
| **job-003** export payments | suspects legacy export payments are below market | legacy scheme statement | generation owner (r3) | half-hourly export meter; export MPAN; would forfeit a guaranteed payment | p/kWh, reversibility | "could I do better though by switching…?" (s009) |
| **job-004** heat pump suitability & running cost | boiler near end of life, grant available | failing boiler | owner-occupier (r4) | property fabric, radiators, cylinder space, grant conditions | suitability, 10-yr cost, disruption | "should I delay and hope costs fall" (s014) |
| **job-005** smart meter decision | wants time-of-use or metered export but lacks the meter | tariff blocked by metering | household (r5) | connectivity risk, meter generation | what it unlocks, reliability | "I'm still skeptical about them" (s010) |
| **job-006** payment / direct debit | monthly payment rises while account is in credit | payment increase notice | account holder (r6) | seasonal credit is normal supplier behaviour | forecast accuracy, unit cost by method | "Direct Debit Increase BUT in credit" (s011) |

**Negative and post-purchase cases kept:** already-fixed customers have no job at all (source-005); households without export-capable metering cannot act until metering changes (source-002); job-006 and parts of job-002/005 are post-purchase or adoption jobs — a brand-named support question is post-purchase, **not** bottom-of-funnel.

**Missing:** landlord, installer, SME procurement and employer roles; `generate` and `buy` acts; prepayment and affordability contexts; cy-GB. All are recorded as gaps, never filled with plausible guesses.

---

## 4. Comprehensive prompt list — 72 exact, trackable strings

**Legend.** *Var:* obs = observed-language rendering, par = natural paraphrase, sens = sensitivity/control. *Part:* core / rot(ating) / sent(inel) / ctrl / aided. *Aided:* un = unaided, cat = category-aided, tgt = target-aided, cmp = competitor-aided. *Camp:* campaign-exposed (false for every cell — no campaign registered). *Act:* exp explain, dia diagnose, cmpr compare, pln plan, rec recommend, ver verify, tsh troubleshoot, nav navigate. *Journey:* PI problem-identification, EX exploration, RB requirements-building, SS supplier-selection, AD adoption, PP post-purchase. *Locale:* all en-GB / en. *Turn:* 1T single turn, MT scripted multi-turn (turn 2 fixed, shown after ‖). *Lanes:* CM closed-model, RT retrieval, CS consumer-surface. *Transf:* LN lightly-normalised, SQE search-query-expanded, LLM llm-expanded, HW human-written. *Wt:* exposure weight / priority weight — `eq` = equal within stratum (no prevalence evidence), `pend` = awaiting Gate 4. Never one blended score.

| Prompt ID | Exact prompt | Var | Part | Band | Aided | Camp | Buyer job | Act | Journey | Funnel | Role | Locale | Constraints | Expected answer | Turn | Lanes | Evidence | Transf | Weight (exp/pri) | QA |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| prompt-001a | our gas and electricity bill has jumped this month and we haven't changed anything at home. what's going on? | obs | core | B3 | un | no | job-001 fix-or-stay | dia | PI | — | r1 bill payer | en-GB/en | nothing changed; regional rates | diagnosis_and_options | 1T | CM,RT,CS | B · s004,s005 | LN | eq / pend | pass |
| prompt-001b | why would my energy costs suddenly be higher this quarter when nothing about our usage has changed? | par | core | B3 | un | no | job-001 | dia | PI | — | r1 | en-GB/en | as above | diagnosis_and_options | 1T | CM,RT,CS | C · s004,s005 | LLM | eq / pend | pass |
| prompt-002a | what's the best way to keep our energy costs under control for the next year without gambling on prices going down? | obs | core | B4 | un | no | job-001 | pln | EX | — | r1 | en-GB/en | can't absorb volatility | options_with_tradeoffs | 1T | CM,RT,CS | B · s013,s012 | SQE | eq / pend | pass |
| prompt-002b | I want a predictable energy bill for the next 12 months without overpaying. how should I approach it? | par | core | B4 | un | no | job-001 | pln | EX | — | r1 | en-GB/en | as above | options_with_tradeoffs | 1T | CM,RT,CS | C · s012 | LLM | eq / pend | pass |
| prompt-003a | is it worth fixing my energy now or staying on the price cap? | obs | core | B1 | un | no | job-001 | cmpr | RB | — | r1 | en-GB/en | 12-month horizon; exit fees | comparison_with_tradeoffs | 1T | CM,RT,CS | B · s012 | LN | eq / pend | pass · pilot cell |
| prompt-003b | should I sign a 12 month fixed energy deal or stay on the standard variable rate? what are the trade-offs, including exit fees? | par | core | B1 | un | no | job-001 | cmpr | RB | — | r1 | en-GB/en | as above | comparison_with_tradeoffs | 1T | CM,RT,CS | C · s012,s005 | LLM | eq / pend | pass |
| prompt-004a | which energy deal should I go for in the UK for a 3 bed house if I want the lowest total cost over a year? | obs | core | B1 | un | no | job-001 | rec | SS | BOFU | r1 | en-GB/en | typical 3-bed usage | ranked_recommendation | 1T | CM,RT,CS | B · s013,s012 | SQE | eq / pend | pass · pilot cell |
| prompt-004b | we're a family of four in a three bedroom semi in the UK. who should we get our gas and electricity from for the cheapest annual cost? | par | core | B1 | un | no | job-001 | rec | SS | BOFU | r1 | en-GB/en | as above | ranked_recommendation | 1T | CM,RT,CS | C · s012 | LLM | eq / pend | pass |
| prompt-005a | how do tracker tariffs compare with fixed and standard variable tariffs for UK households? | obs | aided | B2 | cat | no | job-001 | cmpr | EX | — | r1 | en-GB/en | category named, no supplier | category_explanation_with_tradeoffs | 1T | CM,RT,CS | B · s012 | LN | eq / pend | pass · **Gate 3**: "tracker" generic-word exception |
| prompt-006a | Is Octopus Energy's Flexible Octopus tariff actually cheaper than the Ofgem price cap right now, and what does it charge per kWh? | obs | aided | B0 | tgt | no | job-001 | ver | SS | — | r1 | en-GB/en | target named | factual_verification | 1T | CM,RT,CS | C · s001,s005 | HW | eq / pend | pass · B0 exception |
| prompt-007a | what did the energy price cap change on 1 July 2026 actually do to a typical UK household bill? | obs | sent | B5 | un | no | job-001 | exp | PI | TOFU | r1 | en-GB/en | tied to 1 Jul 2026 reset | event_explanation_with_implications | 1T | CM,RT,CS | B · s004,s005 | LN | eq / pend | pass · expires 2026-08-26 |
| prompt-007b | the price cap went up on 1 July 2026 - how much more will a typical dual fuel household pay, and what should they do about it? | par | sent | B5 | un | no | job-001 | exp | PI | TOFU | r1 | en-GB/en | as above | event_explanation_with_implications | 1T | CM,RT,CS | C · s004 | LLM | eq / pend | pass · expires 2026-08-26 |
| prompt-008a | is E.ON Next or British Gas better for a 12 month fixed energy deal in the UK right now? | obs | aided | B1 | cmp | no | job-001 | cmpr | SS | — | r1 | en-GB/en | two incumbents named | comparison_with_tradeoffs | 1T | CM,RT,CS | B · s006,s012 | HW | eq / pend | pass · **Gate 3**: competitor choice |
| prompt-009a | how do I check whether the fixed energy deal I just signed is actually cheaper than the cap for my usage? | obs | rot | B1 | un | no | job-001 | ver | PP | — | r1 | en-GB/en | already fixed | verification_method | 1T | CM,RT | B · s012 | LN | eq / pend | pass |
| prompt-010a | is it worth fixing my energy or staying on the price cap at the moment? ‖ ok, based on that, which specific UK deals should I be looking at? | obs | core | B1 | un | no | job-001 | cmpr | RB | — | r1 | en-GB/en | fixed turn-2 follow-up | comparison_then_shortlist | MT | CM,RT,CS | B · s012,s013 | LN | eq / pend | pass |
| prompt-010b | should I lock in an energy price for a year or leave it variable? ‖ which suppliers and tariffs would you actually shortlist for that? | par | core | B1 | un | no | job-001 | cmpr | RB | — | r1 | en-GB/en | as above | comparison_then_shortlist | MT | CM,RT,CS | C · s012 | LLM | eq / pend | pass |
| prompt-011a | what's the cheapest way to charge an electric car at home overnight in the UK? | obs | core | B4 | un | no | job-002 EV charging | pln | EX | — | r2 EV driver | en-GB/en | home charging only | options_with_tradeoffs | 1T | CM,RT,CS | B · s017 | SQE | eq / pend | pass · pilot cell |
| prompt-011b | we're getting an EV. how do we set things up so charging it at home costs as little as possible? | par | core | B4 | un | no | job-002 | pln | EX | — | r2 | en-GB/en | as above | options_with_tradeoffs | 1T | CM,RT,CS | C · s008 | LLM | eq / pend | pass |
| prompt-012a | which UK electricity tariff is best for charging an EV overnight if my charger can be controlled remotely by the supplier? | obs | core | B1 | un | no | job-002 | cmpr | RB | — | r2 | en-GB/en | controllable charge point | comparison_with_tradeoffs | 1T | CM,RT,CS | B · s017,s008 | LN | eq / pend | pass |
| prompt-012b | my home charger can be controlled by the energy company. what overnight charging tariff options should I compare, and how do they differ? | par | core | B1 | un | no | job-002 | cmpr | RB | — | r2 | en-GB/en | as above | comparison_with_tradeoffs | 1T | CM,RT,CS | C · s017 | LLM | eq / pend | pass |
| prompt-013a | charging our new electric car at home has pushed our electricity bill up a lot. what are our options? | obs | core | B3 | un | no | job-002 | dia | PI | — | r2 | en-GB/en | new EV in household | diagnosis_and_options | 1T | CM,RT,CS | B · s017 | LN | eq / pend | pass · pilot cell |
| prompt-013b | since we got the EV our electricity usage has nearly doubled and the bill is painful. what should I look at first? | par | core | B3 | un | no | job-002 | dia | PI | — | r2 | en-GB/en | as above | diagnosis_and_options | 1T | CM,RT,CS | C · s008 | LLM | eq / pend | pass |
| prompt-014a | how do EV time-of-use electricity tariffs work in the UK, and what do you need to qualify for one? | obs | aided | B2 | cat | no | job-002 | exp | EX | — | r2 | en-GB/en | category named, no supplier | category_explanation | 1T | CM,RT,CS | B · s017,s004 | LN | eq / pend | pass |
| prompt-015a | I've got solar, a house battery and an EV on a fixed cheap overnight rate. would a half-hourly wholesale tariff actually be better, and is it compatible with my export payments? | obs | core | B1 | un | no | job-002 (+icp-003) | cmpr | RB | — | r2 | en-GB/en | solar+battery+EV; export must stay compatible; round-trip losses | comparison_with_tradeoffs | 1T | CM,RT | B · s008 | LN | eq / pend | pass · pilot cell |
| prompt-015b | with solar, battery storage and an electric car, is it worth moving from a fixed night rate to prices that change every half hour? I don't want to lose my export deal. | par | core | B1 | un | no | job-002 | cmpr | RB | — | r2 | en-GB/en | as above | comparison_with_tradeoffs | 1T | CM,RT | C · s008 | LLM | eq / pend | pass |
| prompt-016a | Does Intelligent Octopus Go work with a Zappi charger and a home battery, and can I keep an export tariff alongside it? | obs | aided | B0 | tgt | no | job-002 | ver | AD | — | r2 | en-GB/en | target product + hardware named | factual_verification | 1T | CM,RT,CS | C · s001,s008 | HW | eq / pend | pass · B0 exception |
| prompt-017a | my smart charging schedule didn't give me the cheap overnight rate last night. what should I check? | obs | rot | B3 | un | no | job-002 | tsh | AD | — | r2 | en-GB/en | schedule already set up | troubleshooting_checklist | 1T | CM,RT | **D** · s018 | LLM | eq / pend | **quarantine** — no behavioural source; routed back to job analysis |
| prompt-018a | I get a fixed payment for the solar electricity I export under an old scheme. am I losing money by staying on it? | obs | core | B3 | un | no | job-003 export | dia | PI | — | r3 solar owner | en-GB/en | legacy guaranteed payment; no export meter | diagnosis_and_options | 1T | CM,RT,CS | B · s009 | LN | eq / pend | pass |
| prompt-018b | my solar export payment is estimated rather than metered. how do I tell whether I'd be better off changing? | par | core | B3 | un | no | job-003 | dia | PI | — | r3 | en-GB/en | as above | diagnosis_and_options | 1T | CM,RT,CS | C · s009 | LLM | eq / pend | pass |
| prompt-019a | is there a way I can calculate the difference between my deemed export payments and a metered export tariff? | obs | core | B1 | un | no | job-003 | cmpr | RB | — | r3 | en-GB/en | wants a method; would forfeit a guaranteed payment | comparison_method_and_numbers | 1T | CM,RT,CS | B · s009 | LN | eq / pend | pass · pilot cell |
| prompt-019b | how do I work out, in pounds, whether metered export would pay me more than the guaranteed export payment I get now? | par | core | B1 | un | no | job-003 | cmpr | RB | — | r3 | en-GB/en | as above | comparison_method_and_numbers | 1T | CM,RT,CS | C · s009 | LLM | eq / pend | pass |
| prompt-020a | how do I get the most money for the solar electricity we don't use ourselves? | obs | core | B4 | un | no | job-003 | pln | EX | — | r3 | en-GB/en | surplus currently underpaid | options_with_tradeoffs | 1T | CM,RT,CS | B · s009 | LN | eq / pend | pass |
| prompt-020b | we generate more than we use in summer. what's the best thing to do with the surplus? | par | core | B4 | un | no | job-003 | pln | EX | — | r3 | en-GB/en | as above | options_with_tradeoffs | 1T | CM,RT,CS | C · s009 | LLM | eq / pend | pass |
| prompt-021a | how does the Smart Export Guarantee work, and is there a difference in SEG payments between providers? | obs | aided | B2 | cat | no | job-003 | cmpr | EX | — | r3 | en-GB/en | public scheme named | category_explanation_with_rates | 1T | CM,RT,CS | B · s009 | LN | eq / pend | pass |
| prompt-022a | What do I need in place to get Outgoing Octopus export payments if I'm already on a legacy feed-in tariff? | obs | aided | B0 | tgt | no | job-003 | exp | RB | — | r3 | en-GB/en | target product named; legacy scheme | eligibility_and_requirements | 1T | CM,RT,CS | C · s002 | HW | eq / pend | pass · B0 exception |
| prompt-023a | what do I need before I can be paid per unit for the electricity I export - do I need a particular meter and a second MPAN? | obs | core | B3 | un | no | job-003 | ver | RB | — | r3 | en-GB/en | half-hourly export data; export MPAN | prerequisite_checklist | 1T | CM,RT,CS | B · s002,s009 | LN | eq / pend | pass · pilot cell |
| prompt-023b | what has to be in place, meter wise, before a supplier will pay me for exported electricity? | par | core | B3 | un | no | job-003 | ver | RB | — | r3 | en-GB/en | as above | prerequisite_checklist | 1T | CM,RT,CS | C · s002 | LLM | eq / pend | pass |
| prompt-024a | how do metered export payments compare with the estimated payments I get now? ‖ which UK suppliers pay the most per unit for exported electricity at the moment? | obs | rot | B1 | un | no | job-003 | cmpr | SS | — | r3 | en-GB/en | fixed turn-2 follow-up | comparison_then_shortlist | MT | CM,RT | B · s009 | LN | eq / pend | pass |
| prompt-025a | our gas boiler is on its last legs and I don't know if a heat pump would work in a 1930s house with the radiators we've got. | obs | core | B3 | un | no | job-004 heat pump | dia | PI | — | r4 owner-occupier | en-GB/en | older fabric; existing radiators | diagnosis_and_options | 1T | CM,RT,CS | C · s014 | LN | eq / pend | pass · pilot cell |
| prompt-025b | we need to replace the boiler soon. how do I tell whether our older house is actually suitable for a heat pump? | par | core | B3 | un | no | job-004 | dia | PI | — | r4 | en-GB/en | as above | diagnosis_and_options | 1T | CM,RT,CS | C · s014 | LLM | eq / pend | pass |
| prompt-026a | how can we cut our heating bills and carbon without a huge upfront cost? | obs | core | B4 | un | no | job-004 | pln | EX | — | r4 | en-GB/en | limited upfront budget | options_with_tradeoffs | 1T | CM,RT,CS | C · s014 | LN | eq / pend | pass |
| prompt-026b | what's the most sensible way to make an old house cheaper to heat if we can't spend a fortune all at once? | par | core | B4 | un | no | job-004 | pln | EX | — | r4 | en-GB/en | as above | options_with_tradeoffs | 1T | CM,RT,CS | C · s014 | LLM | eq / pend | pass |
| prompt-027a | new gas boiler versus an air source heat pump - what's the honest cost comparison over ten years, including the grant? | obs | core | B1 | un | no | job-004 | cmpr | RB | — | r4 | en-GB/en | 10-year view; grant assumed | comparison_with_numbers | 1T | CM,RT,CS | C · s014,s015 | LN | eq / pend | pass · pilot cell |
| prompt-027b | if I compare replacing our boiler with putting in an air source heat pump, what does each cost to install and to run over a decade? | par | core | B1 | un | no | job-004 | cmpr | RB | — | r4 | en-GB/en | as above | comparison_with_numbers | 1T | CM,RT,CS | C · s014 | LLM | eq / pend | pass |
| prompt-028a | what does an air source heat pump installation actually involve in a UK home? | obs | aided | B2 | cat | no | job-004 | exp | EX | — | r4 | en-GB/en | category named, no installer | process_explanation | 1T | CM,RT,CS | B · s014,s015 | LN | eq / pend | pass |
| prompt-029a | will I need bigger radiators and a hot water cylinder if we put in a heat pump, and where does the cylinder normally go? | obs | rot | B3 | un | no | job-004 | ver | RB | — | r4 | en-GB/en | existing radiators; no cylinder space | prerequisite_checklist | 1T | CM,RT | C · s014 | LN | eq / pend | pass |
| prompt-030a | have the UK heat pump grants changed this year for homes on oil heating? | obs | rot | B5 | un | no | job-004 | exp | EX | TOFU | r4 | en-GB/en | tied to a claimed grant change | event_explanation_with_implications | 1T | RT | C · s015 | SQE | eq / pend | **quarantine** — B5 needs dated primary evidence; verify or delete by 2026-08-15 |
| prompt-031a | How does getting a Cosy heat pump installed by Octopus Energy compare with using a local MCS-certified installer? | obs | aided | B0 | tgt | no | job-004 | cmpr | SS | — | r4 | en-GB/en | target offer named | comparison_with_tradeoffs | 1T | CM,RT,CS | C · s003 | HW | eq / pend | pass · B0 exception |
| prompt-032a | our boiler is 18 years old and I'm not sure a heat pump would suit our house. ‖ assuming it might work, what should I do next and who should assess it? | obs | core | B3 | un | no | job-004 | dia | RB | — | r4 | en-GB/en | fixed turn-2 follow-up | diagnosis_then_next_steps | MT | CM,RT,CS | C · s014 | LN | eq / pend | pass |
| prompt-032b | I've been told our house isn't right for a heat pump but I don't know if that's true. ‖ how do I get a proper assessment, and what should the survey cover? | par | core | B3 | un | no | job-004 | dia | RB | — | r4 | en-GB/en | as above | diagnosis_then_next_steps | MT | CM,RT,CS | C · s014 | LLM | eq / pend | pass |
| prompt-033a | is a smart meter actually worth it? I'm still sceptical about them. | obs | core | B3 | un | no | job-005 metering | ver | PI | — | r5 legacy meter | en-GB/en | distrust; legacy meter | balanced_pros_cons | 1T | CM,RT,CS | B · s010 | LN | eq / pend | pass · pilot cell |
| prompt-033b | what do I actually gain, and what could go wrong, if I let them fit a smart meter? | par | core | B3 | un | no | job-005 | ver | PI | — | r5 | en-GB/en | as above | balanced_pros_cons | 1T | CM,RT,CS | C · s010 | LLM | eq / pend | pass |
| prompt-034a | what do I need to change to get onto a cheaper time of use electricity rate? | obs | core | B4 | un | no | job-005 | pln | EX | — | r5 | en-GB/en | meter can't report half-hourly | prerequisite_checklist | 1T | CM,RT,CS | B · s004,s010 | LN | eq / pend | pass |
| prompt-034b | we'd like to pay less by using electricity at off peak times. what has to happen first? | par | core | B4 | un | no | job-005 | pln | EX | — | r5 | en-GB/en | as above | prerequisite_checklist | 1T | CM,RT,CS | C · s002 | LLM | eq / pend | pass |
| prompt-035a | my smart meter has stopped sending readings. does that affect my tariff or my export payments? | obs | rot | B3 | un | no | job-005 | tsh | AD | — | r5 | en-GB/en | readings failing post-install | troubleshooting_checklist | 1T | CM,RT | B · s010 | LN | eq / pend | pass |
| prompt-036a | what's the difference between SMETS1 and SMETS2 smart meters, and why does it matter for time of use tariffs? | obs | aided | B2 | cat | no | job-005 | exp | RB | — | r5 | en-GB/en | metering standard named | category_explanation | 1T | CM,RT,CS | B · s002 | LN | eq / pend | pass |
| prompt-037a | why has my energy direct debit gone up when my account is in credit? | obs | core | B3 | un | no | job-006 payments | dia | PI | — | r6 account holder | en-GB/en | account in credit | diagnosis_and_options | 1T | CM,RT,CS | B · s011 | LN | eq / pend | pass · pilot cell |
| prompt-037b | my supplier wants to increase my monthly payment even though the account has been in credit all summer. is that justified? | par | core | B3 | un | no | job-006 | dia | PI | — | r6 | en-GB/en | seasonal credit build-up | diagnosis_and_options | 1T | CM,RT,CS | C · s011 | LLM | eq / pend | pass |
| prompt-038a | how do I stop paying more than I need to for energy each month? | obs | core | B4 | un | no | job-006 | pln | PP | — | r6 | en-GB/en | payment should match usage | options_with_tradeoffs | 1T | CM,RT,CS | B · s011 | LN | eq / pend | pass |
| prompt-038b | what's the right way to set a monthly energy payment so it matches what we actually use? | par | core | B4 | un | no | job-006 | pln | PP | — | r6 | en-GB/en | as above | options_with_tradeoffs | 1T | CM,RT,CS | C · s011 | LLM | eq / pend | pass |
| prompt-039a | who do I complain to if my energy supplier won't reduce my direct debit? | obs | rot | B3 | un | no | job-006 | nav | PP | — | r6 | en-GB/en | supplier has refused | escalation_route | 1T | CM,RT | B · s011,s004 | LN | eq / pend | pass |
| prompt-040a | is paying by monthly direct debit actually cheaper than paying each bill when it arrives? | obs | rot | B1 | un | no | job-006 | cmpr | RB | — | r6 | en-GB/en | payment method is the variable | comparison_with_numbers | 1T | CM,RT | B · s004 | LN | eq / pend | pass |
| prompt-041a | what's the cheapest way to run a home in the UK at the moment? | obs | sent | B4 | un | no | job-001 | pln | EX | — | r1 | en-GB/en | frozen broad wording | options_with_tradeoffs | 1T | CM,RT,CS | B · s013 | SQE | eq / pend | pass · overlaps cell-002, not double-counted |
| prompt-041b | how do I cut the cost of running my home this year? | par | sent | B4 | un | no | job-001 | pln | EX | — | r1 | en-GB/en | as above | options_with_tradeoffs | 1T | CM,RT,CS | C · s013 | LLM | eq / pend | pass |
| prompt-042a | who should I switch my gas and electricity to? | obs | sent | B1 | un | no | job-001 | rec | SS | — | r1 | en-GB/en | unconstrained; max recommendation pressure | ranked_recommendation | 1T | CM,RT,CS | B · s013 | SQE | eq / pend | pass |
| prompt-042b | which energy supplier should I use in the UK? | par | sent | B1 | un | no | job-001 | rec | SS | — | r1 | en-GB/en | as above | ranked_recommendation | 1T | CM,RT,CS | C · s013 | LLM | eq / pend | pass · **Gate 3**: confirm "energy supplier" is not category-aiding |
| prompt-043a | energy bills too high what do i do | obs | sent | B3 | un | no | job-006 | dia | PI | — | r1 | en-GB/en | deliberately low-context, imperfect | diagnosis_and_options | 1T | CM,RT,CS | B · s013 | SQE | eq / pend | pass |
| prompt-043b | my energy bill is too much, help | par | sent | B3 | un | no | job-006 | dia | PI | — | r1 | en-GB/en | as above | diagnosis_and_options | 1T | CM,RT,CS | C · s013 | LLM | eq / pend | pass |
| prompt-044a | we want a renewable energy tariff but service matters more to us than saving a couple of pounds. which UK supplier should we go with? | obs | sent | B2 | cat | no | job-001 | rec | SS | — | r1 | en-GB/en | green-tariff category; service criterion | ranked_recommendation | 1T | CM,RT,CS | B · s007,s012 | LN | eq / pend | pass · category-aided denominator |
| prompt-044b | if I'm choosing a green electricity tariff and I care most about how the company treats customers, who should I pick? | par | sent | B2 | cat | no | job-001 | rec | SS | — | r1 | en-GB/en | as above | ranked_recommendation | 1T | CM,RT,CS | C · s006,s007 | LLM | eq / pend | pass |
| prompt-045a | which broadband provider should I go with in the UK? | sens | ctrl | off-domain | un | no | — | rec | SS | — | — | en-GB/en | off-domain by design | ranked_recommendation | 1T | CM,RT,CS | C · none | HW | excluded from all denominators | pass · drift control |
| prompt-046a | what's the best way to insulate a loft in an older UK house? | sens | ctrl | adjacent | un | no | — | pln | EX | — | — | en-GB/en | adjacent need, no target standing | options_with_tradeoffs | 1T | CM,RT,CS | C · none | HW | excluded from all denominators | pass · drift control |

**Two prompts were drafted and rejected**, kept in the ledger rather than deleted:

- *"my energy bills are too high — which supplier should I switch to?"* — rejected for welding a recommendation request onto a B3 problem statement purely to elicit brands. The halves now live as cell-043 (problem) and cell-042 (recommendation) with separate denominators.
- *"is an agile tariff cheaper than a fixed one?"* — rejected because the adjective collides with a target product name; the surviving comparison is prompt-005a with neutral wording.

---

## 5. Coverage matrix

**Bands** (cells): B0 4 · B1 11 · B2 5 · B3 13 · B4 8 · B5 2 · control 2 → 46 cells, 72 prompts.

**Aided status** (cells, separate denominators): unaided 34 · category-aided 5 · target-aided 4 · competitor-aided 1 · control 2 (in no denominator). Campaign-exposed: 0.

**Partitions:** core 21 cells / 42 prompts · rotating 8 / 8 · sentinel 5 / 10 · aided 10 / 10 · control 2 / 2.

**Acts:** compare 12 · diagnose 7 · explain 7 · plan 7 · verify 6 · recommend 4 · troubleshoot 2 · navigate 1 · **implement 0 · generate 0 · buy 0**.

**Journeys:** problem-identification 12 · exploration 12 · requirements-building 12 · supplier-selection 6 · post-purchase 3 · adoption 2.

**Jobs:** job-001 10 · job-002 7 · job-003 7 · job-004 8 · job-005 4 · job-006 4 · cross-job sentinel/control 6.

**Roles:** r1 12 · r2 7 · r3 7 · r4 8 · r5 4 · r6 4 · none (controls) 2. **Locales:** en-GB 46. **Turn form:** single 43 · scripted multi-turn 3.

**Evidence grade (cells):** B 28 · C 17 · D 1 · **A 0**. Prompt-level: B 30 · C 41 · D 1.

**Lanes:** closed-model 45 cells · retrieval 46 · consumer-surface 34 · campaign-experiment 0 (inactive).

### Gaps and required waivers

| Missing stratum | Why it is missing | Evidence needed to add it |
| --- | --- | --- |
| Evidence grade A | no first-party corpus supplied | lawfully collected customer/prospect conversations with collection metadata |
| Acts `generate`, `buy` | no evidence buyers use assistants to draft or transact here | behavioural evidence of that use |
| Act `implement` | adoption cells exist but none asks for execution | same |
| B5 breadth (only 1 dated event) | only one dated primary event found | dated primary coverage of a second relevant event |
| Locale cy-GB | no demand evidence, no native reviewer | locale query/community evidence + native reviewer |
| Prepayment, arrears, affordability | **not researched in this run** — the group where a wrong answer does most harm | regulator/charity guidance + behaviourally anchored language |
| Renters, flats, park homes (heating) | unresearched | forum or survey evidence from those tenures |
| Business / SME / salary sacrifice | website-only standing (company copy ≠ demand) | procurement/RFP language, SME reviews, independent SME source |
| Landlord, installer, employer roles | no buyer-side evidence | interviews or community evidence |
| Campaign lane | no campaign, no pre-registration | treatment/control definitions and pre-registration |
| Per-region cells (14 cap regions) | region is a constraint inside cells, not a locale; 14 cells would be a padded grid | evidence that answers differ materially by region |

No cell was created to hit a number. The 33 unaided cells sit inside the 30–48 diagnostic band because the evidence reached that far, not because the band asked for it.

---

## 6. QA ledger

**Outcome (authoritative per-candidate records in `prompt_qa.json`):** 72 candidates → **70 pass · 0 revise · 2 quarantine · 2 rejected drafts**. Note: the aggregate `counts` block inside `prompt_qa.json` v1.0.0 is wrong (it reads pass 65 / quarantine 5); the decision records and `accepted_candidate_ids` (70 IDs) govern, and the discrepancy is logged as `chg-0002` in `panel_change_ledger.json` for correction before freeze.

**Contamination:** 62 unaided and category-aided candidates scanned (normalised, token, edit-distance-1, plus semantic review of slogans and flattering claims). **Zero target-term hits, zero competitor hits.** One controlled exception: "tracker" in prompt-005a. Five candidates carry declared exceptions (four B0 target-aided, one competitor-aided).

**Quarantined:** prompt-017a (grade-D model expansion, no behavioural source — routed back to job analysis); prompt-030a (B5 resting on unverified commercial guidance — verify against a primary government source or delete by 2026-08-15). Neither appears in any reported result.

**Duplicates:** no exact normalised duplicates. Eleven similarity pairs nominated and **all retained** — nine as protected differences (act, journey, band, or a material constraint differs) and two as deliberate sentinel/core overlaps flagged so job-level rollups do not double-count. Nothing was auto-deleted; no embedding model of fixed version was available, so pairs were nominated by review.

**Coverage created by rejection/quarantine:** B5 is one event deep and expires 2026-08-26; the EV adoption troubleshooting intent is uncovered in core.

**Five open Gate 3 decisions:** (1) confirm the "tracker" generic-word exception; (2) approve or replace the two named incumbents in prompt-008a; (3) confirm panel-wide that "energy supplier" is a market-role noun and not category-aiding; (4) verify or delete cell-030; (5) confirm that agent-written paraphrase variants (grade-C wording) may run inside grade-B/C core cells.

---

## 7. Tracking plan (summary — full version in `tracking_plan.md`)

Two variants per core and sentinel cell, one per rotating, aided and control cell. Three repeats per variant per wave (six for sentinels across ≥2 time blocks, two for rotating). Six surfaces across three lanes; **API and consumer-surface observations never share a rollup**. Fresh session for every observation, memory and custom instructions off, clean account archetype, model version recorded exactly as exposed or as `unexposed`. Retrieval state recorded per observation (`required`/`allowed`/`disabled`/`unavailable`, whether retrieval actually ran, generated queries and citations when exposed). Order randomised under seed `20260725` across three time blocks. Exposure weights are **equal within strata with an explicit limitation**; priority weights are withheld pending Gate 4 — the two are never blended. Wilson intervals for single unweighted proportions within one lane and aided status; stratified cluster bootstrap by canonical cell for anything pooled or weighted; **no percentages for subgroups under 20–30 cells**. A 16-cell × 2-variant × 6-repeat variance pilot must run before any percentage is published. Cadence: monthly evidence intake, quarterly review, event triggers (cap publication, product/grant/regulatory change, model or surface change), annual charter approval. Next review **2026-08-26**.

---

## 8. Human gates

| Gate | Owns | Status |
| --- | --- | --- |
| 1 | facts, ICPs, exclusions, permissions | **pending** — 7 ICPs unapproved; icp-004 low confidence; icp-007 excluded from all cells |
| 2 | jobs, language, roles, locales, priority | **pending** — 6 jobs unapproved; job-004 `hypothesis_only` |
| 3 | core/aided/campaign partitions, disputed QA | **pending** — 5 open items listed in §6 |
| 4 | weights, limitations, cadence, claims, version | **pending** — priority weights undefined; equal exposure weighting unapproved; no pilot |

**What would change the panel:**

1. A first-party corpus (support tickets, chat logs, call notes, on-site search) would lift language to grade A, re-rank which jobs deserve core cells, and could supply the first real exposure weights.
2. Confirming the actual business decision could reshape the charter — competitive benchmarking would expand the competitor-aided partition; campaign measurement would require a pre-registered lane with matched controls before any before/after reading.
3. Verifying the heat-pump grant change either promotes cell-030 to a dated B5 sentinel or deletes it.
4. Researching prepayment and affordability contexts would likely add 4–8 cells and is the highest-priority gap.
5. The variance pilot decides whether budget goes to more unique cells or more repeats, and until it runs, only counts and example answers should be published.
6. Running any of this changes nothing about the panel's selection: no observed AI answer, ranking, mention or content gap influenced a single cell or word here, and none may.
