# AI Visibility Panel — US healthcare booking marketplace (Zocdoc)

**Panel:** `panel-zocdoc-booking` v0.1.0 · **Status:** `provisional_directional` · **Built:** 2026-07-25 · **Next review:** 2026-10-25
**Inputs used:** one public URL plus a one-line description. No charter choice, approver, customer corpus, locale reviewer, exposure weights, or variance pilot was supplied.

---

## 1. Decision and limits

**Business decision.** Decide where AI assistants surface this marketplace when (a) patients try to find an in-network, near-term appointment and (b) independent practices evaluate paid new-patient acquisition — and therefore which evidence-supported buyer situations to invest in and re-measure.

**Estimands** (exact numerators/denominators in `measurement_charter.json`): `unaided_brand_presence`, `aided_brand_knowledge`, `competitive_mention_share`, `citation_presence`, `answer_framing`, plus a separate control false-positive rate. `campaign_response` is **excluded** — no campaign was supplied.

**Population.** US English-speaking users of general assistants and AI search who are either booking care for themselves/a dependent or running a small US practice. Excluded: non-US markets, non-English prompts (one rotating es-US probe held out), enterprise/EHR procurement, clinical accuracy, and any prompt touching a real person's health data.

**Lanes/surfaces.** `closed_model`, `retrieval`, `consumer_surface` across three declared surfaces. No campaign lane.

**Hard limits on what this can say.**
- Results are **conditional on this panel**. It is a judgment sample of intent cells, not a probability sample of AI users.
- Equal exposure weighting is an assumption. **No source measures prompt prevalence** in this category.
- Aided, unaided, control, and lane observations **never share a denominator**. There is no single "AI visibility score."
- The panel is **not frozen**: hashes are null, all four human gates are pending, and no variance pilot has run.
- Nothing here supports causal or attribution claims.

---

## 2. Evidence base

17 sources: **5 company-asserted**, **4 buyer-behavior**, **5 independent**, **2 search-proxy**, **1 llm-hypothesis** (grade D, used only to mark a discovery cell as unsupported).
Grades: A ×1, B ×7, C ×8, D ×1.

| Class | IDs | What they carry |
| --- | --- | --- |
| company_asserted | 002, 003, 004, 005, 016 | Fee model and spend controls; insurance-matching limits; two dated 2026 announcements; Spanish language filter |
| buyer_behavior | 007, 008, 010, 014 | Dated 2026 BBB complaints (patients and practices), aggregated complaint themes, a physician's first-person economics, and a zero-review directory listing |
| independent | 001, 006, 009, 012, 013 | Encyclopedic perimeter; payer-deal trade coverage; a competitor-published pricing critique; May 2026 ghost-network editorial; Second Circuit / OIG record |
| search_proxy | 011, 015 | Patient-question FAQ headings and a query-shaped landing slug |
| llm_hypothesis | 017 | Declared model expansion, no URL, grade D |

**Conflicts and counterevidence kept.** The target's own page states insurance verification **does not guarantee in-network coverage** (source-003), which qualifies the marketplace's core promise and is corroborated by complaint language (007, 008). A practitioner account reports ~3× no-shows and a **6% one-year repeat rate** and discontinuation (010). A major review directory carries **zero reviews** (014), so no aggregated satisfaction corpus exists there. The pricing critique (009) is published by a competing vendor and is graded C with the bias stated.

**Permissions.** All sources are public. No personal data beyond short public complaint phrasings was retained; no long excerpts.

**Evidence gaps that matter.**
1. `www.zocdoc.com` root, `/about/`, `/languages`, and `/about/news/` returned **HTTP 403**, so first-party positioning is under-sampled and the contamination lexicon may be incomplete.
2. No first-party corpus (interviews, chat logs, site search, AI conversations) was supplied — all buyer language is public proxy.
3. No query-volume or AI-prompt-frequency data exists anywhere in the manifest → **no sourced exposure weights**.
4. Spanish-language demand is asserted only by an indexed page title; **no Spanish behavioral evidence**.
5. Two of three B5 cells rest partly on vendor-published dated material.

---

## 3. ICPs and buyer jobs

**Supported ICPs.** `icp-001` insured adult blocked by inaccurate directories and unanswered phones (medium); `icp-002` patient with an acute 24–72h need (medium); `icp-003` caregiver for a dependent, sharpest in pediatric behavioral health (medium); `icp-004` owner/office manager of an independent practice paying per booking (**high** — the best-evidenced segment).
**Hypothesis-only.** `icp-005` front-desk phone bottleneck (low; the product page was unreachable and the trigger is inferred from the patient side); `icp-006` health-plan digital lead (low; one deal report is not a segment).

**Jobs** (12; `buyer_jobs.json`): find someone genuinely in-network and open (`job-001`); get seen in 24–72h (`job-002`); pick a route to an appointment (`job-003`); book without phoning (`job-004`); fix a bad booking outcome (`job-005`); care for a dependent (`job-006`); value a new patient and choose an acquisition model (`job-007`); dispute fees and cap spend (`job-008`); judge listing/ranking trust (`job-009`); phone-bottleneck (`job-010`, hypothesis); member-portal booking (`job-011`, hypothesis); post-visit billing (`job-012`, **out-of-perimeter control only**).

**Language worth preserving** (verbatim-ish, provenance retained): "charged me 3 times for clients I couldn't see"; "fee jump of over 200%"; "a completely different person than I expected"; "10 to 20 hours calling 60 clinicians"; "only 6% became repeat patients over a year"; "How do I find doctors that take my insurance?"; "Can I book a doctor online the same day?".

**Negatives / disqualifiers.** Uninsured cash-pay has no supporting evidence here. Enterprise health systems have no recovered language. Post-purchase jobs exist on **both** sides (patient billing disputes, practice fee disputes) and are covered rather than collapsed into pre-purchase intent.

---

## 4. Comprehensive prompt list

44 canonical cells, 86 candidates. Every candidate below is exact and trackable. Weight status is identical across the panel — **exposure: equal-within-strata (unsourced); priority: withheld pending human approval** — and is stated once here rather than repeated per row; per-cell deviations would appear in the Weight column if any existed.

Abbreviations: bands `B0`–`B5`; aided `U`=unaided, `T`=target-aided, `Cmp`=competitor-aided, `Cat`=category-aided; lanes `CM`=closed_model, `R`=retrieval, `CS`=consumer_surface; all rows are `campaign_exposed: false`, `turn_form: single_turn`, `funnel: null`, and locale `en-US` unless noted.

| Prompt ID | Exact prompt | Variant | Partition | Band | Aided | Job | Act | Journey | Role | Locale | Constraints | Expected answer | Lanes | Evidence | Transformation | QA |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| prompt-001a | My insurance website lists a bunch of in-network doctors but every office I call says they're not taking new patients or don't take my plan. What's going on and what do I do? | observed | core | B3 | U | job-001 in-network access | diagnose | problem_identification | patient | en-US | wrong directory; in-network | diagnosis_and_options | CM,R,CS | B / 012,008 | lightly_normalized | pass |
| prompt-001b | I've called eight offices from my insurer's directory and none of them can actually see me. Why are these lists so wrong and how do I find someone who's really available? | paraphrase | core | B3 | U | job-001 | diagnose | problem_identification | patient | en-US | wrong directory | diagnosis_and_options | CM,R,CS | B / 012 | human_written | pass |
| prompt-002a | I just switched jobs and need a new primary care doctor who takes my new insurance. What's the fastest way to get set up without spending a whole day on the phone? | paraphrase | core | B4 | U | job-001 | plan | exploration | patient | en-US | new plan; avoid phone | step_by_step_plan | CM,R,CS | B / 011,015 | human_written | pass |
| prompt-002b | Moving to a new city next month and I need to line up a regular doctor that's in network. How do people usually do this now? | paraphrase | core | B4 | U | job-001 | plan | exploration | patient | en-US | relocation | step_by_step_plan | CM,R,CS | B / 011 | human_written | pass |
| prompt-003a | How do I confirm a doctor is actually in-network before I show up, so I don't get a surprise bill? | paraphrase | core | B4 | U | job-001 | verify | requirements_building | patient | en-US | plan variant; no surprise bill | verification_checklist | CM,R,CS | B / 003,008 | human_written | pass |
| prompt-003b | What exactly should I ask the office and my insurer before an appointment to make sure the visit is covered? | paraphrase | core | B4 | U | job-001 | verify | requirements_building | patient | en-US | coverage confirmation | verification_checklist | CM,R,CS | B / 003 | human_written | pass |
| prompt-004a | How do I find doctors that take my insurance and book online? | observed | core | B1 | U | job-001 | recommend | supplier_selection | patient | en-US | in-network; bookable | ranked_recommendations | CM,R,CS | B / 011 | lightly_normalized | pass |
| prompt-004b | What's the best way to search for and book an in-network doctor online right now? | paraphrase | core | B1 | U | job-001 | recommend | supplier_selection | patient | en-US | in-network | ranked_recommendations | CM,R,CS | B / 011 | human_written | pass |
| prompt-004c | Which doctor-booking platform should I buy a subscription to? | sensitivity | core | B1 | U | job-001 | recommend | supplier_selection | patient | en-US | — | ranked_recommendations | — | D / none | llm_expanded | **reject** — recommendation forcing |
| prompt-005a | I've had a sore throat and fever for three days and my doctor's next opening is six weeks out. Should I go to urgent care, do a video visit, or wait? | paraphrase | core | B3 | U | job-002 near-term care | diagnose | problem_identification | patient | en-US | acute; no opening | care_setting_triage_options | CM,R,CS | B / 011 | human_written | pass |
| prompt-005b | UTI symptoms started last night and my regular office can't fit me in. Where should I actually go today? | paraphrase | core | B3 | U | job-002 | diagnose | problem_identification | patient | en-US | acute; today | care_setting_triage_options | CM,R,CS | B / 011 | human_written | pass |
| prompt-006a | Can I book a doctor online for the same day? | observed | core | B4 | U | job-002 | plan | problem_identification | patient | en-US | 24–72h | step_by_step_plan | CM,R,CS | B / 011 | lightly_normalized | pass |
| prompt-006b | I need to see someone in the next day or two, not in three weeks. How do I find an appointment that soon? | paraphrase | core | B4 | U | job-002 | plan | problem_identification | patient | en-US | 24–72h | step_by_step_plan | CM,R,CS | B / 011 | human_written | pass |
| prompt-007a | What are my options for getting seen today: urgent care, a walk-in clinic, or booking online somewhere? | paraphrase | core | B1 | U | job-002 | compare | supplier_selection | patient | en-US | today; cost unknown | option_comparison | CM,R,CS | B / 011 | human_written | pass |
| prompt-007b | Compare the ways to get an appointment today when your own doctor has no openings. | paraphrase | core | B1 | U | job-002 | compare | supplier_selection | patient | en-US | today | option_comparison | CM,R,CS | B / 011 | human_written | pass |
| prompt-008a | Is it better to search my insurance company's provider directory or use an appointment booking site to find a doctor? | paraphrase | core | B1 | U | job-003 route choice | compare | exploration | patient | en-US | must end in a booking | option_comparison | CM,R,CS | B / 011,006 | human_written | pass |
| prompt-008b | Insurer directory, an online booking site, or just calling the office - which one actually gets you an appointment? | paraphrase | core | B1 | U | job-003 | compare | exploration | patient | en-US | must end in a booking | option_comparison | CM,R,CS | B / 011 | human_written | pass |
| prompt-009a | Do I need an account to book a doctor's appointment online, and what information will they ask for? | observed | core | B4 | U | job-004 book without calling | implement | adoption | patient | en-US | first booking; card details | preparation_checklist | CM,R,CS | B / 011 | lightly_normalized | pass |
| prompt-009b | First time booking a specialist online. What do I need ready, and what happens after I book? | paraphrase | core | B4 | U | job-004 | implement | adoption | patient | en-US | first booking | preparation_checklist | CM,R,CS | B / 003 | human_written | pass |
| prompt-010a | I've spent hours on hold calling clinics and nobody picks up or calls back. Is there a way to get an appointment without phoning? | observed | core | B3 | U | job-004 | troubleshoot | exploration | patient | en-US | no callback | workaround_options | CM,R,CS | B / 012 | lightly_normalized | pass |
| prompt-010b | Every office sends me to voicemail. How can I get scheduled without playing phone tag for a week? | paraphrase | core | B3 | U | job-004 | troubleshoot | exploration | patient | en-US | no callback | workaround_options | CM,R,CS | B / 012 | human_written | pass |
| prompt-011a | The listing said the doctor was in-network but I got billed out-of-network. What are my options now? | observed | core | B3 | U | job-005 bad booking | troubleshoot | post_purchase | patient | en-US | already billed | escalation_path | CM,R,CS | B / 008,003 | lightly_normalized | pass |
| prompt-011b | I booked based on an online listing that showed my plan, then the office said they don't take it and I owe the full amount. What can I do? | paraphrase | core | B3 | U | job-005 | troubleshoot | post_purchase | patient | en-US | already billed | escalation_path | CM,R,CS | B / 008,003 | human_written | pass |
| prompt-011c | Zocdoc said the doctor took my insurance but I got billed out of network - what do I do? | sensitivity | core | B3 | U | job-005 | troubleshoot | post_purchase | patient | en-US | — | escalation_path | — | D / 008 | llm_expanded | **reject** — target term in unaided cell |
| prompt-012a | I showed up for my appointment and was seen by a completely different provider than the one I booked. Is that allowed? | observed | core | B3 | U | job-005 | verify | post_purchase | patient | en-US | provider swapped | rights_and_next_steps | CM,R,CS | A / 007 | lightly_normalized | pass |
| prompt-012b | My appointment was cancelled an hour before, then the same slot showed as available again. What should I do next? | paraphrase | core | B3 | U | job-005 | verify | post_purchase | patient | en-US | late cancellation | rights_and_next_steps | CM,R,CS | A / 007 | human_written | pass |
| prompt-013a | I've called dozens of therapists from my kid's insurance list and almost none are taking new patients or even call back. How do other parents get their child into care? | observed | core | B3 | U | job-006 dependent care | diagnose | problem_identification | caregiver | en-US | child; behavioral health | diagnosis_and_options | CM,R,CS | B / 012 | lightly_normalized | pass |
| prompt-013b | Trying to find a child therapist that takes our plan and I keep hitting dead ends. What's the realistic path here? | paraphrase | core | B3 | U | job-006 | diagnose | problem_identification | caregiver | en-US | child; in-network | diagnosis_and_options | CM,R,CS | B / 012 | human_written | pass |
| prompt-014a | I need a pediatrician for my son who takes our plan and has appointments after school hours. How should I go about finding one? | paraphrase | core | B4 | U | job-006 | plan | requirements_building | caregiver | en-US | after-school; child plan | step_by_step_plan | CM,R,CS | B / 012,011 | human_written | pass |
| prompt-014b | As a busy working parent of two in a major metro area, what is the optimal strategy for identifying a pediatric provider aligned with our insurance coverage and after-school availability requirements? | sensitivity | core | B4 | U | job-006 | plan | requirements_building | caregiver | en-US | — | step_by_step_plan | — | D / 012 | llm_expanded | **revise** — persona exposition |
| prompt-014c | Need a pediatrician who takes our insurance and has late afternoon slots. Where do I even start? | paraphrase | core | B4 | U | job-006 | plan | requirements_building | caregiver | en-US | after-school; child plan | step_by_step_plan | CM,R,CS | B / 012 | human_written | pass (revision of 014b) |
| prompt-015a | Are the doctor ratings and reviews on health directory sites trustworthy, or do providers pay to show up higher? | paraphrase | core | B3 | U | job-009 listing trust | verify | exploration | patient | en-US | suspects paid placement | trust_assessment | CM,R,CS | B / 007,013 | human_written | pass |
| prompt-015b | How much should I trust online reviews when picking a new doctor? | paraphrase | core | B3 | U | job-009 | verify | exploration | patient | en-US | review integrity | trust_assessment | CM,R,CS | B / 007 | human_written | pass |
| prompt-016a | My schedule has holes every week and the new patients I do get from paid channels barely come back. How do I figure out what's actually working? | observed | core | B3 | U | job-007 acquisition value | diagnose | problem_identification | owner-clinician | en-US | unfilled slots; low repeat | diagnosis_and_options | CM,R,CS | B / 010 | lightly_normalized | pass |
| prompt-016b | Small practice with empty slots most weeks. I can't tell which of my marketing spend is producing real new patients. How do I measure this? | paraphrase | core | B3 | U | job-007 | diagnose | problem_identification | owner-clinician | en-US | attribution unclear | diagnosis_and_options | CM,R,CS | B / 010 | human_written | pass |
| prompt-017a | What should a practice actually be willing to pay to acquire one new patient? | paraphrase | core | B4 | U | job-007 | plan | requirements_building | owner-clinician | en-US | acquisition ceiling | cost_model_explanation | CM,R,CS | B / 010,009 | human_written | pass |
| prompt-017b | How do I work out what a new patient is worth over time so I know what an acquisition fee should cap out at? | paraphrase | core | B4 | U | job-007 | plan | requirements_building | owner-clinician | en-US | lifetime value | cost_model_explanation | CM,R,CS | B / 010 | human_written | pass |
| prompt-017c | What is the industry-standard patient acquisition cost benchmark by specialty in 2026? | sensitivity | core | B4 | U | job-007 | plan | requirements_building | owner-clinician | en-US | — | cost_model_explanation | — | D / none | llm_expanded | **quarantine** — unsupported premise → Gate 3 |
| prompt-018a | Is paying per new patient booking better than a flat monthly scheduling subscription for a small practice? | paraphrase | core | B1 | U | job-007 | compare | supplier_selection | office manager | en-US | small practice; predictable cost | option_comparison | CM,R,CS | B / 009,002 | human_written | pass |
| prompt-018b | We're weighing per-booking fees against a fixed monthly fee for online scheduling. Which makes more sense at low volume? | paraphrase | core | B1 | U | job-007 | compare | supplier_selection | office manager | en-US | low volume | option_comparison | CM,R,CS | B / 009 | human_written | pass |
| prompt-019a | We see Medicare and Medicaid patients. Is there a compliance risk if we pay a marketing fee for every new patient who books with us? | paraphrase | core | B3 | U | job-007 | verify | requirements_building | owner-clinician | en-US | federal program patients | compliance_explanation | CM,R,CS | B / 013 | human_written | pass |
| prompt-019b | Can a practice legally pay a per-new-patient fee for referrals when federal health programs are involved? | paraphrase | core | B3 | U | job-007 | verify | requirements_building | owner-clinician | en-US | AKS exposure | compliance_explanation | CM,R,CS | B / 013 | human_written | pass |
| prompt-020a | We keep getting charged booking fees for patients we couldn't actually see because they weren't in network. How do practices handle this? | observed | core | B3 | U | job-008 fee disputes | troubleshoot | post_purchase | office manager | en-US | charged for unusable bookings | escalation_path | CM,R,CS | A / 007 | lightly_normalized | pass |
| prompt-020b | Getting billed for no-shows and cancellations we had no control over. What recourse does a practice usually have? | paraphrase | core | B3 | U | job-008 | troubleshoot | post_purchase | office manager | en-US | no-shows | escalation_path | CM,R,CS | A / 007,008 | human_written | pass |
| prompt-021a | How can we cap what we spend per month on new-patient bookings without turning off online scheduling completely? | paraphrase | core | B4 | U | job-008 | implement | post_purchase | office manager | en-US | spend caps | configuration_guidance | CM,R,CS | B / 002 | human_written | pass |
| prompt-021b | Is there a way to pause new patient bookings for one provider while keeping the rest of the practice listed? | paraphrase | core | B4 | U | job-008 | implement | post_purchase | office manager | en-US | per-provider pause | configuration_guidance | CM,R,CS | B / 002 | human_written | pass |
| prompt-022a | I keep reading about ghost networks in insurance directories. What is that and how does it affect me trying to find a therapist? | paraphrase | core | B5 | U | job-001 | explain | exploration | patient | en-US | story; review-by 2026-10-25 | trend_explanation | R,CS | B / 012 | human_written | pass |
| prompt-022b | There was a Boston Globe editorial in May about insurers listing therapists who aren't reachable. Is anything actually being done about it? | paraphrase | core | B5 | U | job-001 | explain | exploration | patient | en-US | story; review-by 2026-10-25 | trend_explanation | R,CS | B / 012 | human_written | pass |
| prompt-023a | A lot of people seem to be asking AI chatbots about symptoms before seeing a doctor. Is that changing how appointments go? | paraphrase | core | B5 | U | job-009 | explain | exploration | patient | en-US | story; vendor-survey source | trend_explanation | R,CS | C / 005 | human_written | pass (flagged → Gate 3) |
| prompt-023b | Should I tell my doctor that I looked up my symptoms with an AI first? | paraphrase | core | B5 | U | job-009 | explain | exploration | patient | en-US | story; review-by 2026-10-25 | trend_explanation | R,CS | C / 005 | human_written | pass (flagged → Gate 3) |
| prompt-023c | What did the 2026 federal telehealth rule change about booking appointments? | sensitivity | core | B5 | U | job-009 | explain | exploration | patient | en-US | — | trend_explanation | — | D / none | llm_expanded | **reject** — undated/invented event |
| prompt-024a | I noticed review apps now let you book a doctor's appointment right from the listing. Is that reliable? | paraphrase | core | B5 | U | job-003 | explain | exploration | patient | en-US | story; review-by 2026-10-25 | trend_explanation | R,CS | C / 004 | human_written | pass |
| prompt-024b | Can you actually book medical appointments straight from a search app or a health plan's website now, or is it just a lead form? | paraphrase | core | B5 | U | job-003 | explain | exploration | patient | en-US | story; review-by 2026-10-25 | trend_explanation | R,CS | C / 004,006 | human_written | pass |
| prompt-025a | Does Zocdoc actually verify my insurance, or can I still end up out of network? | paraphrase | aided | B0 | T | job-005 | explain | post_purchase | patient | en-US | coverage guarantee | capability_and_limits_explanation | CM,R | C / 003 | human_written | pass |
| prompt-025b | If Zocdoc shows my plan for a doctor, is the visit guaranteed to be covered? | paraphrase | aided | B0 | T | job-005 | explain | post_purchase | patient | en-US | coverage guarantee | capability_and_limits_explanation | CM,R | C / 003 | human_written | pass |
| prompt-026a | Is Zocdoc legit? Do doctors pay to show up higher in the results? | paraphrase | aided | B0 | T | job-009 | verify | exploration | patient | en-US | paid ranking | trust_assessment | CM,R | B / 013 | human_written | pass |
| prompt-026b | Are Zocdoc reviews real, and how does it decide which doctors to show me? | paraphrase | aided | B0 | T | job-009 | verify | exploration | patient | en-US | review integrity | trust_assessment | CM,R | B / 007 | human_written | pass |
| prompt-027a | How much does Zocdoc charge a practice per new patient booking? | paraphrase | aided | B0 | T | job-007 | explain | requirements_building | owner-clinician | en-US | mutable price; review-by 2026-10-25 | pricing_explanation | CM,R | B / 009,002 | human_written | pass |
| prompt-027b | What does Zocdoc cost for a dental practice compared with a primary care office? | paraphrase | aided | B0 | T | job-007 | explain | requirements_building | owner-clinician | en-US | specialty spread | pricing_explanation | CM,R | B / 009 | human_written | pass |
| prompt-028a | Our Zocdoc booking fee jumped a lot this year. Can we dispute charges or cap our monthly spend? | observed | aided | B0 | T | job-008 | troubleshoot | post_purchase | office manager | en-US | fee increase | escalation_path | CM,R | A / 007,002 | lightly_normalized | pass |
| prompt-028b | How do we remove our providers from Zocdoc and stop being charged for bookings we can't honor? | paraphrase | aided | B0 | T | job-008 | troubleshoot | post_purchase | office manager | en-US | exit path | escalation_path | CM,R | A / 007 | human_written | pass |
| prompt-029a | Which is better, Zocdoc or Healthgrades? | observed | aided | B1 | T | job-003 | compare | supplier_selection | patient | en-US | booking vs research | head_to_head_comparison | CM,R | B / 011 | verbatim | pass |
| prompt-029b | For actually booking an appointment rather than researching a doctor, is Zocdoc or Healthgrades the better choice? | paraphrase | aided | B1 | T | job-003 | compare | supplier_selection | patient | en-US | booking vs research | head_to_head_comparison | CM,R | B / 011 | human_written | pass |
| prompt-030a | For a small practice, how does Healthgrades compare with Emitrr for getting and scheduling new patients? | paraphrase | aided | B1 | Cmp | job-007 | compare | supplier_selection | office manager | en-US | small practice | head_to_head_comparison | CM,R | C / 009,011 | human_written | pass |
| prompt-030b | We're looking at Healthgrades profile tools and Emitrr scheduling. Which is a better fit for a two-provider clinic? | paraphrase | aided | B1 | Cmp | job-007 | compare | supplier_selection | office manager | en-US | two providers | head_to_head_comparison | CM,R | C / 009 | human_written | pass |
| prompt-031a | What are the best online doctor appointment booking sites for finding an in-network provider? | paraphrase | aided | B2 | Cat | job-003 | recommend | supplier_selection | patient | en-US | insurance filter | ranked_recommendations | CM,R,CS | B / 011 | human_written | pass |
| prompt-031b | Which appointment booking sites let you filter by insurance and book instantly? | paraphrase | aided | B2 | Cat | job-003 | recommend | supplier_selection | patient | en-US | instant booking | ranked_recommendations | CM,R,CS | B / 011 | human_written | pass |
| prompt-032a | What should a practice look for in an online patient scheduling marketplace? | paraphrase | aided | B2 | Cat | job-007 | compare | requirements_building | office manager | en-US | evaluation criteria | criteria_and_tradeoffs | CM,R,CS | C / 009 | human_written | pass |
| prompt-032b | Comparing patient acquisition marketplaces for medical practices - what are the main tradeoffs? | paraphrase | aided | B2 | Cat | job-007 | compare | requirements_building | office manager | en-US | tradeoffs | criteria_and_tradeoffs | CM,R,CS | C / 009,014 | human_written | pass |
| prompt-033a | El directorio de mi seguro tiene muchos médicos, pero cuando llamo me dicen que no aceptan mi plan o que no reciben pacientes nuevos. ¿Qué puedo hacer? | paraphrase | rotating | B3 | U | job-001 | diagnose | problem_identification | patient | **es-US / es** | Spanish; locale review pending | diagnosis_and_options | R,CS | C / 016 | translated | **quarantine** — locale review pending |
| prompt-034a | Necesito una cita esta semana con un médico que hable español y acepte mi seguro. ¿Cómo la consigo? | paraphrase | rotating | B4 | U | job-002 | plan | problem_identification | patient | **es-US / es** | Spanish; this week | step_by_step_plan | R,CS | C / 016 | translated | **quarantine** — locale review pending |
| prompt-035a | Our front desk can't keep up with the phones and we're losing appointment calls to voicemail. What do practices do about this? | paraphrase | rotating | B3 | U | job-010 phone bottleneck | diagnose | problem_identification | office manager | en-US | small front desk | diagnosis_and_options | CM,R | C / 012,007 | human_written | pass (rotating discovery only) |
| prompt-036a | How can a health plan let members book in-network appointments directly from the member portal instead of just listing phone numbers? | paraphrase | rotating | B4 | U | job-011 payer portal | plan | requirements_building | payer digital lead | en-US | member portal | step_by_step_plan | CM,R | C / 006 | human_written | pass (rotating discovery only) |
| prompt-037a | I don't have insurance right now. What's the cheapest way to see a doctor about something non-urgent? | paraphrase | rotating | B4 | U | job-002 | plan | exploration | patient | en-US | uninsured | step_by_step_plan | R | D / 017 | llm_expanded | **quarantine** — grade D, no evidence |
| prompt-038a | How do I find a doctor who takes my insurance and is actually accepting new patients? | observed | sentinel | B3 | U | job-001 | diagnose | problem_identification | patient | en-US | in-network; accepting | diagnosis_and_options | CM,R,CS | B / 011,015 | lightly_normalized | pass |
| prompt-038b | how do i find a doctor that takes my insurance and is taking new patients | sensitivity | sentinel | B3 | U | job-001 | diagnose | problem_identification | patient | en-US | imperfect casing variant | diagnosis_and_options | CM,R,CS | B / 015 | lightly_normalized | pass |
| prompt-039a | I need to see a dermatologist within the next week. How do I find one with an opening? | paraphrase | sentinel | B4 | U | job-002 | plan | problem_identification | patient | en-US | within a week | step_by_step_plan | CM,R,CS | B / 011 | human_written | pass |
| prompt-039b | need derm appt this week, how do i find someone with openings | sensitivity | sentinel | B4 | U | job-002 | plan | problem_identification | patient | en-US | abbreviated variant | step_by_step_plan | CM,R,CS | B / 011 | human_written | pass |
| prompt-040a | What's the best way to book a doctor's appointment online? | paraphrase | sentinel | B1 | U | job-003 | recommend | supplier_selection | patient | en-US | no phone call | ranked_recommendations | CM,R,CS | B / 011 | human_written | pass |
| prompt-040b | Recommend a reliable way to book a medical appointment online without calling. | paraphrase | sentinel | B1 | U | job-003 | recommend | supplier_selection | patient | en-US | no phone call | ranked_recommendations | CM,R,CS | B / 011 | human_written | pass |
| prompt-041a | How do independent practices get new patients these days without overpaying for each one? | paraphrase | sentinel | B3 | U | job-007 | diagnose | problem_identification | owner-clinician | en-US | cost discipline | diagnosis_and_options | CM,R,CS | B / 010,009 | human_written | pass |
| prompt-041b | What are the most cost-effective ways for a small practice to bring in new patients? | paraphrase | sentinel | B3 | U | job-007 | diagnose | problem_identification | owner-clinician | en-US | cost discipline | diagnosis_and_options | CM,R,CS | B / 010 | human_written | pass |
| prompt-042a | How do I find a specialist for my elderly mother who takes her Medicare plan and has an opening soon? | paraphrase | sentinel | B4 | U | job-006 | plan | requirements_building | caregiver | en-US | Medicare dependent | step_by_step_plan | CM,R,CS | C / 006,012 | human_written | pass (thin evidence) |
| prompt-042b | Looking for a specialist for my mom on Medicare with an appointment in the next couple of weeks. What's the best approach? | paraphrase | sentinel | B4 | U | job-006 | plan | requirements_building | caregiver | en-US | Medicare dependent | step_by_step_plan | CM,R,CS | C / 006 | human_written | pass (thin evidence) |
| prompt-043a | My health insurer denied a claim for a visit I already had. How do I appeal it? | paraphrase | control | B3 | U | job-012 (out-of-perimeter) | plan | post_purchase | patient | en-US | matched control | escalation_path | CM,R,CS | B / 008 | human_written | pass |
| prompt-044a | I got an explanation of benefits I don't understand and a bill that doesn't match it. How do I sort this out? | paraphrase | control | B3 | U | job-012 (out-of-perimeter) | explain | post_purchase | patient | en-US | matched control | explanation | CM,R,CS | B / 008 | human_written | pass |

---

## 5. Coverage matrix

**Cells by partition:** core 24 · aided 8 · rotating 5 · sentinel 5 · control 2 = **44** (41 selected into the panel).
**Unaided cells:** 31 (core 24 + sentinel 5 + control 2) — inside the 30–48 diagnostic band.

| Band | Cells | Aided status |
| --- | ---: | --- |
| B0 direct brand/product | 4 | target_aided (aided partition) |
| B1 comparison/purchase | 8 | 5 unaided, 1 target_aided, 2 competitor_aided |
| B2 category | 2 | category_aided |
| B3 problem/need | 17 | unaided |
| B4 job/goal | 12 | unaided |
| B5 discovery/story | 3 | unaided, all with dated evidence + review-by 2026-10-25 |

**Acts:** diagnose 11 · plan 11 · verify 6 · compare 7 · troubleshoot 5 · explain 6 · recommend 4 · implement 3 · navigate 0 · buy 0 · generate 0.
**Journeys:** problem_identification 13 · exploration 10 · requirements_building 9 · supplier_selection 8 · post_purchase 7 · adoption 1.
**Roles:** patient 26 · caregiver 4 · owner-clinician 8 · office manager 8 · payer lead 1 (persona counts are per cell; some cells serve two ICPs).
**Locales:** en-US 42 · es-US 2 (both excluded from the panel).
**Grades:** A 4 cells · B 26 · C 13 · D 1.
**Lanes:** closed_model 39 cells · retrieval 44 · consumer_surface 33.

**Required waivers / declared gaps** (also in `prompt_architecture.json` and `panel.yaml`):

| Gap | Why | Evidence needed |
| --- | --- | --- |
| `generate` and `buy` acts absent | No evidence anyone asks an assistant to produce an artifact or transact in this perimeter | Query or conversation evidence |
| `navigate` act absent as a standalone cell | Navigation intent is folded into verify/plan cells; no standalone evidence | Site-search or query-log evidence |
| Scripted multi-turn absent | No lawful conversational corpus | Consented multi-turn logs or interviews |
| es-US excluded from panel | Translated, unreviewed; no Spanish behavioral evidence | Native review + Spanish buyer/query evidence |
| Uninsured/cash-pay excluded | Grade-D model expansion only | Independent behavioral evidence |
| AI phone-assistant product area | Product page HTTP 403; no independent coverage | Retrievable page or trade coverage + practice-side language |
| Payer segment thin | One deal report, no buyer language | Payer RFP/procurement or interview evidence |
| Enterprise/EHR persona absent | No recovered language | Procurement evidence |
| Campaign lane absent | No campaign supplied | Campaign terms, window, pre-registration, matched controls |
| Practice-side `adoption` journey | Only company copy | Practice onboarding language |

No Cartesian grid was generated. Every persona × locale × act combination not listed above is deliberately absent.

---

## 6. QA ledger

**86 candidates → 78 pass · 1 revise · 4 quarantine · 3 reject.** Accepted IDs equal exactly the pass IDs; all counts are derived from the decision array.

- **Rejected (3):** `prompt-004c` recommendation forcing (invents a purchase decision to elicit vendors); `prompt-011c` **target brand token inside an unaided B3 cell** — the only contamination hard failure in the set; `prompt-023c` B5 prompt presupposing a 2026 regulation with no dated source.
- **Quarantined (4):** `prompt-017c` unsupported "industry-standard benchmark" premise; `prompt-033a` and `prompt-034a` es-US translations pending native review; `prompt-037a` grade-D uninsured expansion.
- **Revised (1):** `prompt-014b` persona exposition → rewritten as `prompt-014c`, which passed.
- **Contamination scan:** brands 1 failure, all other classes 0. B0/B1 aided cells passed via declared allowed exceptions. Two observed FAQ fragments containing brand terms were withheld from the blind brief and logged.
- **Duplicates:** no exact normalized duplicates. Semantically close pairs were reviewed and **retained** where they protect a material difference (`prompt-021a` spend cap vs `prompt-021b` per-provider pause; `prompt-008a` two-way vs `prompt-008b` three-way route framing). Nothing was auto-deleted.
- **Grade-D discipline:** all six grade-D candidates are rejected or quarantined; none reached core.
- **Disputed decisions for Gate 3:** `prompt-017c`, `prompt-023a`, `prompt-023b`, `prompt-033a`, `prompt-034a`, `prompt-037a`.
- **Blinding:** QA ran with `baseline_fields_blinded: true`. No visibility result, ranking, or target performance influenced any selection.

---

## 7. Tracking plan (summary)

Full detail in `tracking_plan.md`. In brief: 41 cells / 78 prompts × 3 repeats across three declared surfaces ≈ **650 scored observations per monthly wave**; fresh session per observation; retrieval state recorded per run; randomized order under seed 20260801; equal exposure weights with priority weights withheld; Wilson intervals for unweighted strata and a cluster bootstrap by canonical cell for weighted aggregates; a 15-cell × 8-repeat variance pilot across two time blocks is **pending**; quarterly review with monthly evidence intake and event triggers; next review 2026-10-25.

---

## 8. Human gates

| Gate | Status | Decision needed |
| --- | --- | --- |
| 1 — ICPs | **pending** | Promote or reject `icp-005`/`icp-006`; confirm exclusions |
| 2 — Jobs | **pending** | Confirm `job-010`/`job-011` as rotating hypotheses and `job-012` as control-only |
| 3 — Partitions & disputed QA | **pending** | Approve core/aided/control split; resolve the six disputed candidates; approve or reject es-US after native review |
| 4 — Weights, limits, cadence, version | **pending** | Approve equal exposure weighting, set priority weights, approve limitations, cadence, and v0.1.0 |

**What would change this panel most:**
1. **Any real exposure evidence** (query volumes, assistant-usage data, or first-party logs) would replace equal weighting and could re-rank the whole cell allocation.
2. **A first-party customer corpus** would upgrade several C-grade cells to A/B and would likely add jobs this public-only manifest cannot see.
3. **Restored access to blocked first-party pages** would complete the contamination lexicon and may add B0 cells for the phone-assistant and telehealth areas.
4. **A native Spanish reviewer** would move two rotating cells into the panel.
5. **Running the variance pilot** would replace default repeat counts and give honest interval widths.

*Nothing in this report is frozen, representative, statistically significant, or causal. It is a candidate panel with its assumptions stated.*
