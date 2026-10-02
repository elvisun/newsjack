# AI visibility panel — Cloudflare (network platform)

**Status: `provisional_directional` — a candidate panel, not a frozen measurement panel.**
Run: 2026-07-25T16:00:00-04:00 · Panel `panel-cloudflare-network-platform` v0.1.0 · Input: `https://www.cloudflare.com/` plus a one-line description.

---

## 1. Decision and limits

**Assumed business decision** (you did not supply one): decide where to invest answer-engine content, documentation and third-party evidence across the five product areas, by first establishing a repeatable baseline of how assistants answer the buyer questions enumerated below.

**Estimands** (six, six separate denominators): `unaided_brand_presence` (26 cells), a separately-denominated category-aided presence rate (7 cells), `competitive_mention_share`, `citation_presence` (retrieval + consumer surfaces only), `answer_framing` (blocked — no codebook), `aided_brand_knowledge` (7 target-aided cells). `campaign_response` is out of scope: no campaign exists.

**Population.** The population of interest is a *prompt-space* population — the 50 evidence-supported canonical intent cells below. It is **not** a population of AI users, and no probability sample of AI users exists here.

**Lanes.** `closed_model`, `retrieval`, `consumer_surface` instantiated; `campaign_experiment` defined but empty.

**What this panel is not.** Not frozen. Not representative. Not statistically powered — no variance pilot has run, so `precision_status` is `directional_only`. Not causal. Every result is *conditional on this panel*. All four human gates are **pending**.

Three limits deserve to be stated bluntly rather than buried:

- **There is no grade-A evidence in this run.** No interviews, call transcripts, on-site search logs, support tickets or permitted AI-conversation corpus were supplied. The best available evidence is dated independent reporting and public forum comments with permalinks.
- **Blinding was procedural, not architectural.** No fresh subagent or separate session was available, so a single context enforced the unaided boundary by working only from `blind_design_brief.json`. That is weaker than a fresh-context generator.
- **No hashing runtime.** Every hash field reads `sha256:unavailable` and the randomization seed is `null`. The panel *cannot be frozen* until these are computed.

---

## 2. Evidence base

16 sources recorded, 2 blocked.

| Class | Count | IDs |
| --- | ---: | --- |
| `company_asserted` | 3 | src-001 (homepage), src-002 (plans), src-012 (analyst-recognition LPs) |
| `independent` | 4 | src-003 (The Register, 2025-12-08), src-004 (TechCrunch, 2026-07-01), src-010 (CDN/edge buyer guide), src-013 (2025-11-18 outage analyses) |
| `buyer_behavior` | 3 | src-005 (HN, CDN selection), src-006 (HN, ZTNA/VPN), src-011 (peer review complaints) |
| `search_proxy` | 5 | src-007, src-008, src-009 (SASE RFP criteria), src-014, src-015 |
| `llm_hypothesis` | 1 | src-016 (edge-compute expansion — declared, not disguised) |

**Grades:** A = 0, B = 8, C = 7, D = 1.

**Conflicts and counterevidence kept, not erased.** src-013 (a four-hour network-wide outage on 2025-11-18, caused by an oversized bot-management feature file propagating network-wide) directly qualifies the availability standing claimed on src-001. src-006 records a practitioner reporting better uptime from a hyperscaler WAF over ~1.5 years, and another arguing mid-market firms may not need a dedicated edge vendor at all. src-011 records recurring complaints about multi-day support responses on paid plans and plan-gated features. These shaped cells 006, 011, 027, 029, 035 and 046 rather than being dropped.

**Permissions.** All sources public. No personal data stored beyond three publicly known co-founder names in the contamination register, held solely so leakage into an unaided prompt is detectable.

**Gaps and failures.**
- `https://www.trustradius.com/...cloudflare-zero-trust-services-vs-zscaler-private-access` → HTTP 403. Not read; nothing from it is used.
- `https://community.cloudflare.com/c/security/...` → HTTP 403. Not read.
- src-010, 011, 012, 013, 015 were reached only as search-result summaries. They are graded C/low-confidence and must be re-verified directly before any core promotion.
- **Company copy establishes standing, not demand.** src-002 declares audiences in Cloudflare's own words ("small businesses operating online", "Teams under 50 users"); those declarations were used to bound standing and never as evidence that buyers exist in those shapes.
- **A forum post is evidence from that source, not prevalence.** Nothing in this panel claims any intent occurs at any rate in the market.

---

## 3. ICPs and buyer jobs

| ICP | Context | Trigger evidence | Confidence / status |
| --- | --- | --- | --- |
| icp-001 | Operator of an ad- or content-supported website losing capacity and content control to crawlers | src-003 (GPTBot disallows 3.3M→5.6M Jul–Dec 2025; ClaudeBot 3.2M→5.8M), src-004, src-007 | medium / supported |
| icp-002 | Small business / ecommerce owner-operator, no security staff, site downtime = revenue event | src-014, src-008, src-002 | medium / supported |
| icp-003 | Platform / SRE team at a mid-market software company selecting or re-evaluating edge delivery + protection | src-005, src-006, src-008, src-010, src-013 | medium / supported |
| icp-004 | IT / security engineer replacing legacy VPN for hybrid workforce **and non-human clients** | src-006, src-002, src-009, src-015 | medium / supported |
| icp-005 | Developer/small team on serverless edge compute | src-001 + src-016 only | **low / hypothesis_only** |
| icp-006 | Consumer DNS-resolver user | — | **excluded** (no buyer job) |

icp-005 is the honest one. Edge compute appears in *your own description*, and Cloudflare plainly sells the capability (src-001) — but not one buyer-behaviour, review, forum, procurement or search source in this manifest shows anyone expressing that job. Inverting a feature into a buyer is exactly the failure this method forbids, so it stays `hypothesis_only` and its two cells are quarantined.

**Roles asserted:** site/content operator, small-business owner-operator, platform/SRE engineer, IT/security engineer, engineering budget owner, security/compliance reviewer. **No** company size, revenue, seniority, tenure or buying committee is inferred anywhere. No evidence exists for a finance approver, a distinct procurement persona, or a reseller.

**Jobs (11).** job-001 crawler access control · job-002 urgent availability restoration · job-003 delivery/egress cost reduction · job-004 provider selection under reliability/billing/compliance constraints · job-005 single-provider dependency reduction (post-purchase) · job-006 VPN replacement incl. agentless/ephemeral clients · job-007 SASE/SSE evaluation criteria *(grade C, hypothesis_only)* · job-008 support delay & plan-gating *(grade C, hypothesis_only)* · job-009 the dated 2026-09-15 crawler default change · job-010 edge compute *(grade C, null struggling moment — deliberately)* · job-011 control construct, not a demand finding.

**Language worth preserving** (short spans): "added GPTBot to the disallow list in their robots.txt file" (src-003); "discoverable via search and through AI services … but protections against having their intellectual property given away for free" (src-004); "doesn't support WebSockets, can't proxy gRPC" (src-005); "configuring an access client on ephemeral CI/CD compute is impractical" (src-006); "state for each service whether it is built natively or integrated from a third-party engine" (src-009).

**Negatives / constraints that shaped cells:** robots.txt is ignored by a rising share of AI-bot requests (13.26% in Q2 2025 vs 3.3% in Q4 2024, per Tollbit via src-003); flat-rate billing preferred over usage-based; cannot double operational cost for redundancy; ISO 27017/27018-type cloud certifications may block procurement.

---

## 4. Comprehensive prompt list

**92 candidates across 50 canonical cells.** These are exact trackable strings. Scripted multi-turn cells encode turns as `TURN 1: … || TURN 2: …`, both issued in one fresh session.

Legend — **Var**: OL = observed_language, NP = natural_paraphrase, SEN = sensitivity. **Aided**: U = unaided, CAT = category_aided, TGT = target_aided, COMP = competitor_aided. **Part**: C = core, S = sentinel, R = rotating, K = control, A = aided. **Journey**: PI = problem_identification, EX = exploration, RB = requirements_building, SS = supplier_selection, AD = adoption, PP = post_purchase. **Turn**: 1T = single, MT = scripted multi-turn. **Lanes**: cm = closed_model, rt = retrieval, cs = consumer_surface. **Tf** (transformation): hw = human_written, sqe = search_query_expanded, llm = llm_expanded. **Weight**: E/P = exposure/priority; `eq*` = equal weight, no prevalence evidence; `eq†` = equal exposure, priority pending Gate 4. Campaign-exposed is `false` for **every** cell in this panel (no campaign exists), so the column is stated once here rather than repeated 92 times. Funnel is `null` for every cell — see §5.

| Prompt ID | Exact prompt | Var | Part | Band | Aided | Job | Act | Journey | Role | Locale | Constraints | Expected answer | Turn | Lanes | Evidence | Tf | Weight | QA |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| prompt-001a | traffic to my site tripled but my pageviews didn't move — how do I tell how much of this is bots? | OL | C | B3 | U | job-001 crawler control | diagnose | PI | site operator | en-US | robots.txt unreliable | diagnosis_and_options | 1T | cm/rt/cs | B · src-003,007 | sqe | eq*/eq† | pass |
| prompt-001b | Our server load keeps spiking with almost no extra readers. How can I work out what share of the requests are automated? | NP | C | B3 | U | job-001 | diagnose | PI | site operator | en-US | robots.txt unreliable | diagnosis_and_options | 1T | cm/rt/cs | B · src-003,007 | hw | eq*/eq† | pass |
| prompt-002a | why is my hosting bill going up when reader numbers are flat | OL | C | B3 | U | job-001 | explain | PI | site operator | en-US | — | concept_explanation | 1T | cm/rt/cs | B · src-007 | sqe | eq*/eq† | pass |
| prompt-002b | What's driving all this extra server traffic if nobody new is actually reading the articles? | NP | C | B3 | U | job-001 | explain | PI | site operator | en-US | — | concept_explanation | 1T | cm/rt/cs | B · src-003,007 | hw | eq*/eq† | pass |
| prompt-003a | TURN 1: my website is down and my host says its a flood of traffic. what do i do right now \|\| TURN 2: it's back but really slow. how do I stop this happening again tonight? | OL | C | B3 | U | job-002 availability | troubleshoot | PI | SMB owner | en-US | no security staff | stepwise_remediation | MT | cm/rt/cs | B · src-014 | sqe | eq*/eq† | pass (Gate 3: split vs cell-036) |
| prompt-003b | TURN 1: Our site keeps timing out and the hosting company says there are too many requests. How do I get it back online? \|\| TURN 2: It's up again but unstable. What should I put in place before tonight? | NP | C | B3 | U | job-002 | troubleshoot | PI | SMB owner | en-US | no security staff | stepwise_remediation | MT | cm/rt/cs | B · src-014 | hw | eq*/eq† | pass (Gate 3) |
| prompt-004a | how do I tell the difference between a traffic spike from a promotion and an actual attack? | OL | C | B3 | U | job-002 | diagnose | PI | SMB owner | en-US | no security staff | diagnosis_and_options | 1T | cm/rt/cs | B · src-014 | sqe | eq*/eq† | pass |
| prompt-004b | We got about 40x normal traffic in ten minutes and the store went down. Was that an attack or did we just go viral? | NP | C | B3 | U | job-002 | diagnose | PI | SMB owner | en-US | no security staff | diagnosis_and_options | 1T | cm/rt/cs | B · src-014 | hw | eq*/eq† | pass |
| prompt-005a | why is data transfer out the biggest line item on our cloud bill | OL | C | B3 | U | job-003 cost | explain | PI | platform eng | en-US | usage billing | concept_explanation | 1T | cm/rt/cs | B · src-008 | sqe | eq*/eq† | pass |
| prompt-005b | Most of our infrastructure spend turns out to be egress. What actually drives that number? | NP | C | B3 | U | job-003 | explain | PI | platform eng | en-US | usage billing | concept_explanation | 1T | cm/rt/cs | B · src-005,008 | hw | eq*/eq† | pass |
| prompt-006a | our whole stack went down because one upstream provider had an incident. how do we find where else we have that dependency? | OL | C | B3 | U | job-005 dependency | diagnose | PP | platform eng | en-US | no double cost | diagnosis_and_options | 1T | cm/rt/cs | B · src-013 | hw | eq*/eq† | pass |
| prompt-006b | How do I map which parts of our platform stop working if a single external provider fails? | NP | C | B3 | U | job-005 | diagnose | PP | platform eng | en-US | no double cost | diagnosis_and_options | 1T | cm/rt/cs | B · src-006,013 | hw | eq*/eq† | pass |
| prompt-007a | our vpn is slow and gives everyone way more access than they need. how are teams handling this now? | OL | C | B3 | U | job-006 access | explain | EX | IT/sec eng | en-US | agentless clients | concept_explanation | 1T | cm/rt/cs | B · src-006 | hw | eq*/eq† | pass |
| prompt-007b | Remote staff complain the VPN is slow, and one compromised machine would expose the whole network. What's the current thinking on this? | NP | C | B3 | U | job-006 | explain | EX | IT/sec eng | en-US | agentless clients | concept_explanation | 1T | cm/rt/cs | B · src-006 | hw | eq*/eq† | pass |
| prompt-008a | I want to block AI training bots but not lose search traffic. how do people separate the two? | OL | C | B3 | U | job-001 | plan | RB | site operator | en-US | keep search indexing | options_and_tradeoffs | 1T | cm/rt/cs | B · src-003,004 | sqe | eq*/eq† | pass (split vs prompt-030b) |
| prompt-008b | How can a site stay indexed by search engines while refusing crawlers that feed model training? | NP | C | B3 | U | job-001 | plan | RB | site operator | en-US | keep search indexing | options_and_tradeoffs | 1T | cm/rt/cs | B · src-004 | hw | eq*/eq† | pass |
| prompt-009a | I need a way to decide which automated clients get to read our articles, without making it harder for actual readers | OL | C | B4 | U | job-001 | plan | EX | site operator | en-US | keep search indexing | options_and_tradeoffs | 1T | cm/rt/cs | B · src-003,007 | hw | eq*/eq† | pass (split vs cell-037) |
| prompt-009b | What's a sensible approach to governing automated access to a content site without affecting normal visitors? | NP | C | B4 | U | job-001 | plan | EX | site operator | en-US | keep search indexing | options_and_tradeoffs | 1T | cm/rt/cs | B · src-003 | hw | eq*/eq† | pass |
| prompt-010a | need to cut delivery costs on a high traffic app without breaking websockets and grpc | OL | C | B4 | U | job-003 | plan | RB | platform eng | en-US | dynamic/API traffic | cost_reduction_options | 1T | cm/rt/cs | B · src-005,008 | sqe | eq*/eq† | pass |
| prompt-010b | How do we bring down content delivery spend when a lot of our traffic is dynamic and API-based? | NP | C | B4 | U | job-003 | plan | RB | platform eng | en-US | dynamic/API traffic | cost_reduction_options | 1T | cm/rt/cs | B · src-005,008 | hw | eq*/eq† | pass |
| prompt-011a | how do teams stay online when the provider sitting in front of their site has an outage, without paying for everything twice? | OL | C | B4 | U | job-005 | plan | RB | platform eng | en-US | no double cost | tradeoff_analysis | 1T | cm/rt/cs | B · src-006,013 | hw | eq*/eq† | pass |
| prompt-011b | We want to survive an upstream provider failure but we can't afford to run two of everything. What's realistic? | NP | C | B4 | U | job-005 | plan | RB | platform eng | en-US | no double cost | tradeoff_analysis | 1T | cm/rt/cs | B · src-006,013 | hw | eq*/eq† | pass |
| prompt-012a | we need internal apps reachable by staff and by CI jobs without installing a client on every machine | OL | C | B4 | U | job-006 | plan | EX | IT/sec eng | en-US | ephemeral CI clients | options_and_tradeoffs | 1T | cm/rt/cs | B · src-006 | hw | eq*/eq† | pass |
| prompt-012b | How do you give short-lived build agents access to an internal service when they can't run a persistent VPN client? | NP | C | B4 | U | job-006 | plan | EX | IT/sec eng | en-US | ephemeral CI clients | options_and_tradeoffs | 1T | cm/rt/cs | B · src-006 | hw | eq*/eq† | pass |
| prompt-013a | we're doing a big launch next month and I really don't want the site to fall over | OL | C | B4 | U | job-002 | plan | PI | SMB owner | en-US | no security staff | readiness_checklist | 1T | cm/rt/cs | B · src-014 | hw | eq*/eq† | pass |
| prompt-013b | What should a small team do ahead of a launch so the site stays up under a big traffic spike? | NP | C | B4 | U | job-002 | plan | PI | SMB owner | en-US | no security staff | readiness_checklist | 1T | cm/rt/cs | B · src-008,014 | hw | eq*/eq† | pass |
| prompt-014a | what's the actual difference between a CDN and a WAF, and do we need both? | OL | C | B2 | CAT | job-004 selection | compare | EX | platform eng | en-US | — | comparison_table | 1T | cm/rt/cs | B · src-008 | sqe | eq*/eq† | pass (ambiguous-term review cleared) |
| prompt-014b | Do a CDN and a web application firewall overlap enough that one of them covers the other? | NP | C | B2 | CAT | job-004 | compare | EX | platform eng | en-US | — | comparison_table | 1T | cm/rt/cs | B · src-008,010 | hw | eq*/eq† | pass |
| prompt-015a | does a small ecommerce site actually need a WAF or is hosting security enough | OL | C | B2 | CAT | job-004 | explain | EX | SMB owner | en-US | no security staff | concept_explanation | 1T | cm/rt/cs | B · src-008 | sqe | eq*/eq† | pass |
| prompt-015b | Is a web application firewall worth it for a small online store, or is that overkill at our size? | NP | C | B2 | CAT | job-004 | explain | EX | SMB owner | en-US | no security staff | concept_explanation | 1T | cm/rt/cs | B · src-008 | hw | eq*/eq† | pass |
| prompt-016a | ZTNA vs VPN for a 200 person company — what actually changes day to day? | OL | C | B2 | CAT | job-006 | compare | EX | IT/sec eng | en-US | 50-user threshold | comparison_table | 1T | cm/rt/cs | B · src-006,009 | sqe | eq*/eq† | pass |
| prompt-016b | How is zero trust network access different from keeping our VPN and tightening the firewall rules? | NP | C | B2 | CAT | job-006 | compare | EX | IT/sec eng | en-US | 50-user threshold | comparison_table | 1T | cm/rt/cs | B · src-006 | hw | eq*/eq† | pass |
| prompt-017a | which CDN caching strategies actually reduce origin bandwidth when most pages are dynamic? | OL | C | B2 | CAT | job-003 | recommend | RB | platform eng | en-US | dynamic/API traffic | configuration_recommendation | 1T | cm/rt/cs | B · src-008 | sqe | eq*/eq† | pass |
| prompt-017b | For a mostly dynamic site, what CDN configuration gives the biggest origin offload without stale content? | NP | C | B2 | CAT | job-003 | recommend | RB | platform eng | en-US | dynamic/API traffic | configuration_recommendation | 1T | cm/rt/cs | B · src-005,008 | hw | eq*/eq† | pass |
| prompt-018a | what should be in a SASE RFP so I can tell native capabilities from ones that are just OEM'd in? | OL | R | B2 | CAT | job-007 procurement | plan | RB | compliance rev | en-US | certification/residency | criteria_checklist | 1T | cm/rt | **C** · src-009 | sqe | eq*/eq† | **quarantine** (grade C on hypothesis_only job) |
| prompt-019a | how do I set up bot management so scrapers get blocked but search crawlers still get through | OL | C | B2 | CAT | job-001 | implement | AD | site operator | en-US | keep search indexing | configuration_steps | 1T | cm/rt/cs | B · src-003,007 | sqe | eq*/eq† | pass (Gate 3: "bot management" ambiguity) |
| prompt-019b | What's the right way to configure bot mitigation rules so legitimate indexing bots stay allowed? | NP | C | B2 | CAT | job-001 | implement | AD | site operator | en-US | keep search indexing | configuration_steps | 1T | cm/rt/cs | B · src-003,004 | hw | eq*/eq† | pass |
| prompt-020a | what are the main options for putting caching and application protection in front of a SaaS app, and how do they actually differ? | OL | C | B1 | U | job-004 | compare | SS | platform eng | en-US | dynamic/API traffic | ranked_shortlist | 1T | cm/rt/cs | B · src-005,010 | hw | eq*/eq† | pass (split vs cell-039) |
| prompt-020b | We're shortlisting providers to sit in front of our product. What are the real differentiators between them? | NP | C | B1 | U | job-004 | compare | SS | platform eng | en-US | dynamic/API traffic | ranked_shortlist | 1T | cm/rt/cs | B · src-006,010 | hw | eq*/eq† | pass |
| prompt-021a | TURN 1: small online store, need ddos protection and faster image loading. what should I look at? \|\| TURN 2: nobody on the team is technical. which of those is least work to run? | OL | C | B1 | U | job-002 | recommend | SS | SMB owner | en-US | no security staff | ranked_shortlist | MT | cm/rt/cs | B · src-008,014 | sqe | eq*/eq† | pass |
| prompt-021b | TURN 1: Which providers should a small online shop consider for DDoS protection and image performance? \|\| TURN 2: Assume nobody here is technical. Which one is easiest to operate? | NP | C | B1 | U | job-002 | recommend | SS | SMB owner | en-US | no security staff | ranked_shortlist | MT | cm/rt/cs | B · src-008,014 | hw | eq*/eq† | pass |
| prompt-022a | Akamai vs Fastly for a global web app — which suits a small platform team? | OL | A | B1 | **COMP** | job-004 | compare | SS | platform eng | en-US | flat billing | comparison_table | 1T | cm/rt/cs | B · src-005,010 | sqe | eq*/eq† | pass (declared competitor exception) |
| prompt-022b | We're down to Akamai, Fastly and Amazon CloudFront. What are the practical differences for a team of five? | NP | A | B1 | **COMP** | job-004 | compare | SS | platform eng | en-US | flat billing | comparison_table | 1T | cm/rt/cs | B · src-006,010 | hw | eq*/eq† | pass |
| prompt-023a | Zscaler vs Netskope for SSE — what are the real differences? | OL | A | B1 | **COMP** | job-006 | compare | SS | IT/sec eng | en-US | — | comparison_table | 1T | cm/rt/cs | B · src-015 | sqe | eq*/eq† | pass (entailment flagged: snippet-only source) |
| prompt-023b | Choosing between Zscaler and Netskope for secure service edge. What should actually drive that decision? | NP | A | B1 | **COMP** | job-006 | compare | SS | IT/sec eng | en-US | — | comparison_table | 1T | cm/rt/cs | B · src-009,015 | hw | eq*/eq† | pass |
| prompt-024a | which edge and security vendors bill on a flat annual contract instead of per request? | OL | C | B1 | U | job-004 | buy | SS | budget owner | en-US | flat billing | pricing_model_options | 1T | cm/rt/cs | B · src-006 | hw | eq*/eq† | pass |
| prompt-024b | We keep getting surprised by usage-based invoices. Who prices this kind of service predictably? | NP | C | B1 | U | job-004 | buy | SS | budget owner | en-US | flat billing | pricing_model_options | 1T | cm/rt/cs | B · src-006 | hw | eq*/eq† | pass |
| prompt-025a | how do I verify a security vendor's certification and data residency claims before we sign? | OL | C | B1 | U | job-004 | verify | SS | compliance rev | en-US | certification/residency | verification_answer | 1T | cm/rt/cs | B · src-009,010 | sqe | eq*/eq† | pass |
| prompt-025b | What evidence should we ask for on certifications, log retention and where traffic is processed? | NP | C | B1 | U | job-004 | verify | SS | compliance rev | en-US | certification/residency | verification_answer | 1T | cm/rt/cs | B · src-009 | hw | eq*/eq† | pass |
| prompt-026a | what does Cloudflare actually do, in plain terms? | OL | A | **B0** | TGT | job-004 | explain | EX | platform eng | en-US | — | concept_explanation | 1T | cm/rt/cs | B · src-001 | hw | eq*/eq† | pass (B0 exception) |
| prompt-026b | Can you explain Cloudflare's main product areas and what each one is for? | NP | A | **B0** | TGT | job-004 | explain | EX | platform eng | en-US | — | concept_explanation | 1T | cm/rt/cs | B · src-001 | hw | eq*/eq† | pass |
| prompt-027a | how reliable has Cloudflare been over the last couple of years — any major outages? | OL | A | **B0** | TGT | job-004 | verify | SS | budget owner | en-US | no double cost | verification_answer | 1T | cm/rt/cs | B · src-013 | hw | eq*/eq† | pass |
| prompt-027b | What's Cloudflare's track record on availability, and how did they handle their largest incidents? | NP | A | **B0** | TGT | job-004 | verify | SS | budget owner | en-US | no double cost | verification_answer | 1T | cm/rt/cs | B · src-013 | hw | eq*/eq† | pass |
| prompt-028a | how do I replace our VPN with Cloudflare Access for about 200 users? | OL | A | **B0** | TGT | job-006 | implement | AD | IT/sec eng | en-US | 50-user threshold | configuration_steps | 1T | cm/rt/cs | B · src-002,006 | hw | eq*/eq† | pass |
| prompt-028b | What are the steps to move internal app access from a VPN to Cloudflare Zero Trust? | NP | A | **B0** | TGT | job-006 | implement | AD | IT/sec eng | en-US | 50-user threshold | configuration_steps | 1T | cm/rt/cs | B · src-002 | hw | eq*/eq† | pass |
| prompt-029a | my site shows a Cloudflare error page but my server is up. what's going on? | OL | A | **B0** | TGT | job-008 support | troubleshoot | **PP** | SMB owner | en-US | no security staff | stepwise_remediation | 1T | cm/rt/cs | **C** · src-011 | hw | eq*/eq† | pass (grade C, aided only) |
| prompt-029b | Getting a Cloudflare error screen while the origin responds fine. How do I debug that, and how long does support usually take? | NP | A | **B0** | TGT | job-008 | troubleshoot | **PP** | SMB owner | en-US | no security staff | stepwise_remediation | 1T | cm/rt/cs | **C** · src-011 | hw | eq*/eq† | pass (Gate 3: one-concept borderline) |
| prompt-030a | how do I block AI crawlers on Cloudflare while keeping search engine bots? | OL | A | **B0** | TGT | job-001 | implement | AD | site operator | en-US | keep search indexing | configuration_steps | 1T | cm/rt/cs | B · src-001,004 | sqe | eq*/eq† | pass |
| prompt-030b | Using Cloudflare, what's the way to allow search indexing but block crawlers used for model training? | NP | A | **B0** | TGT | job-001 | implement | AD | site operator | en-US | keep search indexing | configuration_steps | 1T | cm/rt/cs | B · src-004 | hw | eq*/eq† | pass (split vs prompt-008a) |
| prompt-031a | is Cloudflare Pro or Business worth it for a small online store? | OL | A | **B0** | TGT | job-004 | buy | SS | SMB owner | en-US | no security staff | pricing_model_options | 1T | cm/rt/cs | B · src-002 | sqe | eq*/eq† | pass |
| prompt-031b | Which Cloudflare plan makes sense for a small business site that takes payments? | NP | A | **B0** | TGT | job-004 | buy | SS | SMB owner | en-US | no security staff | pricing_model_options | 1T | cm/rt/cs | B · src-002 | hw | eq*/eq† | pass |
| prompt-032a | I heard that from September some providers will block AI crawlers by default on pages that carry ads. what does that mean for my site? | OL | C | **B5** | U | job-009 dated change | explain | PI | site operator | en-US | keep search indexing | policy_change_briefing | 1T | cm/rt/cs | B · src-004 (2026-07-01) | hw | eq*/eq† | pass · review_by 2026-09-15 |
| prompt-032b | There's a change coming in September 2026 where crawlers that mix search indexing with AI training get blocked by default on ad-supported pages. How would that affect a publisher? | NP | C | **B5** | U | job-009 | explain | PI | site operator | en-US | keep search indexing | policy_change_briefing | 1T | cm/rt/cs | B · src-004 | hw | eq*/eq† | pass (Gate 3: vendor-framing paraphrase) |
| prompt-033a | what should a publisher change before september if the AI crawler defaults are switching? | OL | C | **B5** | U | job-009 | plan | RB | site operator | en-US | keep search indexing | criteria_checklist | 1T | cm/rt/cs | B · src-004 | hw | eq*/eq† | pass · review_by 2026-09-15 |
| prompt-033b | The deadline is mid-September 2026 for the new crawler defaults. What's the checklist for a site owner between now and then? | NP | C | **B5** | U | job-009 | plan | RB | site operator | en-US | keep search indexing | criteria_checklist | 1T | cm/rt/cs | B · src-004 | hw | eq*/eq† | pass |
| prompt-034a | how many sites are actually blocking AI crawlers now, and is it working? | OL | S | **B5** | U | job-001 | explain | EX | site operator | en-US | robots.txt unreliable | trend_summary_with_evidence | 1T | cm/rt/cs | B · src-003 (2025-12-08) | hw | eq*/eq† | pass · review_by 2026-09-30 |
| prompt-034b | Has adding AI crawlers to robots.txt actually reduced scraping, or do the bots ignore it? | NP | S | **B5** | U | job-001 | explain | EX | site operator | en-US | robots.txt unreliable | trend_summary_with_evidence | 1T | cm/rt/cs | B · src-003 | hw | eq*/eq† | pass |
| prompt-035a | after the big edge provider outage last year, what did engineering teams actually change? | OL | R | **B5** | U | job-005 | explain | PP | budget owner | en-US | no double cost | trend_summary_with_evidence | 1T | cm/rt | B · src-013 | hw | eq*/eq† | **revise** — "last year" drifts across waves; re-render with an absolute date |
| prompt-036a | website suddenly unreachable, host says traffic flood. first steps? | OL | S | B3 | U | job-002 | troubleshoot | PI | SMB owner | en-US | no security staff | stepwise_remediation | 1T | cm/rt/cs | B · src-014 | sqe | eq*/eq† | pass (Gate 3: split vs cell-003) |
| prompt-036b | Our site went offline under a traffic flood. What are the first things to check? | NP | S | B3 | U | job-002 | troubleshoot | PI | SMB owner | en-US | no security staff | stepwise_remediation | 1T | cm/rt/cs | B · src-014 | hw | eq*/eq† | pass |
| prompt-037a | how do I know if the measures I put in place to control automated access are actually working? | OL | S | B4 | U | job-001 | verify | AD | site operator | en-US | robots.txt unreliable | verification_answer | 1T | cm/rt/cs | B · src-003 | hw | eq*/eq† | pass (split vs cell-009) |
| prompt-037b | What signals tell me whether my controls on automated content access are effective? | NP | S | B4 | U | job-001 | verify | AD | site operator | en-US | robots.txt unreliable | verification_answer | 1T | cm/rt/cs | B · src-003 | hw | eq*/eq† | pass |
| prompt-038a | compare the main content delivery network options for a global web app | OL | S | B2 | CAT | job-004 | compare | EX | platform eng | en-US | — | ranked_shortlist | 1T | cm/rt/cs | B · src-005,010 | sqe | eq*/eq† | pass |
| prompt-038b | What are the leading CDN choices for a globally distributed web application? | NP | S | B2 | CAT | job-004 | compare | EX | platform eng | en-US | — | ranked_shortlist | 1T | cm/rt/cs | B · src-010 | hw | eq*/eq† | pass |
| prompt-039a | recommend a provider for DDoS protection and application firewall for a mid sized SaaS | OL | S | B1 | U | job-004 | recommend | SS | platform eng | en-US | — | ranked_shortlist | 1T | cm/rt/cs | B · src-010,014 | sqe | eq*/eq† | pass |
| prompt-039b | Who should a mid-market SaaS company use for DDoS mitigation and application protection? | NP | S | B1 | U | job-004 | recommend | SS | platform eng | en-US | — | ranked_shortlist | 1T | cm/rt/cs | B · src-010 | hw | eq*/eq† | pass |
| prompt-040a | our vpn keeps breaking and everyone has too much access. whats the modern approach? | OL | S | B3 | U | job-006 | explain | EX | IT/sec eng | en-US | — | concept_explanation | 1T | cm/rt/cs | B · src-006 | sqe | eq*/eq† | pass |
| prompt-040b | Our legacy VPN is unreliable and over-permissive. What's replacing that model? | NP | S | B3 | U | job-006 | explain | EX | IT/sec eng | en-US | — | concept_explanation | 1T | cm/rt/cs | B · src-006 | hw | eq*/eq† | pass |
| prompt-041a | how do we cut content delivery and egress costs roughly in half without a rewrite | OL | S | B4 | U | job-003 | plan | RB | platform eng | en-US | dynamic/API traffic | cost_reduction_options | 1T | cm/rt/cs | B · src-008 | hw | eq*/eq† | pass (entailment note: "in half" is a goal, not a claim) |
| prompt-041b | What gets us the largest reduction in delivery and egress spend without re-architecting the app? | NP | S | B4 | U | job-003 | plan | RB | platform eng | en-US | dynamic/API traffic | cost_reduction_options | 1T | cm/rt/cs | B · src-005,008 | hw | eq*/eq† | pass |
| prompt-042a | what is SASE and do we need the whole thing or just ZTNA? | OL | S | B2 | CAT | job-006 | explain | EX | IT/sec eng | en-US | — | concept_explanation | 1T | cm/rt/cs | B · src-009 | sqe | eq*/eq† | pass |
| prompt-042b | Explain SASE. Is it worth adopting fully, or should we start with zero trust access only? | NP | S | B2 | CAT | job-006 | explain | EX | IT/sec eng | en-US | — | concept_explanation | 1T | cm/rt/cs | B · src-009 | hw | eq*/eq† | pass |
| prompt-044a | we want to run code close to users without managing servers — what are the tradeoffs? | OL | R | B4 | U | job-010 edge compute | plan | EX | platform eng | en-US | — | tradeoff_analysis | 1T | cm/rt | **C** · src-001,016 | llm | eq*/eq† | **quarantine** — job has null struggling moment |
| prompt-045a | edge compute vs regional serverless — when does running at the edge actually help? | SEN | R | B2 | CAT | job-010 | compare | EX | platform eng | en-US | — | tradeoff_analysis | 1T | cm/rt | **D** · src-016 | llm | eq*/eq† | **quarantine** — sole grade-D candidate |
| prompt-046a | vendor support hasn't replied in a week and the fix I need is only on a higher plan. what are my options? | OL | R | B3 | U | job-008 | troubleshoot | **PP** | SMB owner | en-US | — | options_and_tradeoffs | 1T | cm/rt | **C** · src-011 | hw | eq*/eq† | **quarantine** — snippet-only source |
| prompt-047a | how do I stop AI companies scraping our articles without losing search traffic? | OL | R | B4 | U | job-001 | plan | RB | site operator | **en-GB** | keep search indexing | options_and_tradeoffs | 1T | cm/rt | B · src-003,004 | hw | eq*/eq† | **quarantine** — locale review pending |
| prompt-048a | our card terminal in the shop keeps dropping offline during busy periods. what should I check? | OL | **K** | B3 | U | job-011 control | troubleshoot | PI | SMB owner | en-US | no security staff | stepwise_remediation | 1T | cm/rt/cs | **C** · src-014 | hw | eq*/eq† | pass — matched-unaffected control |
| prompt-049a | what should a small shop look for when choosing a payment terminal provider? | OL | **K** | B4 | U | job-011 control | recommend | SS | SMB owner | en-US | no security staff | criteria_checklist | 1T | cm/rt/cs | **C** · src-014 | hw | eq*/eq† | pass — matched-unaffected control |
| prompt-050a | bandwidth overage charges every month on a small shop site. whats normally causing that? | OL | C | B3 | U | job-003 | diagnose | PI | SMB owner | en-US | no security staff | diagnosis_and_options | 1T | cm/rt/cs | B · src-008 | sqe | eq*/eq† | pass |
| prompt-050b | Our small store keeps blowing past its bandwidth allowance. Where does that usually come from? | NP | C | B3 | U | job-003 | diagnose | PI | SMB owner | en-US | no security staff | diagnosis_and_options | 1T | cm/rt/cs | B · src-008 | hw | eq*/eq† | pass |
| prompt-051a | what is Cloudflare changing for AI crawlers in September 2026, and does it affect my site? | OL | A | **B5** | TGT | job-009 | explain | PI | site operator | en-US | keep search indexing | policy_change_briefing | 1T | cm/rt/cs | B · src-004 | hw | eq*/eq† | pass (split vs prompt-032a) · review_by 2026-09-15 |
| prompt-051b | Cloudflare's new default for AI crawlers on ad-supported pages — what do publishers need to know? | NP | A | **B5** | TGT | job-009 | explain | PI | site operator | en-US | keep search indexing | policy_change_briefing | 1T | cm/rt/cs | B · src-004 | hw | eq*/eq† | pass |

**Archived rejects** (kept with reasons, not deleted): `prompt-052a` "our site keeps going down from bot traffic — which security platform should we buy?" → recommendation forcing (proh-001). `prompt-053a` "what's the best connectivity cloud for a growing SaaS?" → target-coined proprietary category in an unaided core cell (proh-002). `prompt-054a` "is Akamai or Fastly better for stopping AI scrapers?" → competitor terms in an unaided cell; the intent survives at cell-022 in the aided partition (proh-003).

---

## 5. Coverage matrix

**Band** (50 cells): B0 6 · B1 7 · B2 9 · B3 13 · B4 10 · B5 5.
**Aided status:** unaided 32 · category_aided 9 · target_aided 7 · competitor_aided 2. Campaign-exposed: 0 (no campaign).
**Partition:** core 25 · sentinel 8 · rotating 6 · control 2 · aided 9.
**Information act:** explain 11 · plan 11 · compare 7 · troubleshoot 5 · diagnose 4 · recommend 4 · implement 3 · verify 3 · buy 2 · **generate 0 · navigate 0**.
**Journey:** exploration 13 · problem_identification 11 · supplier_selection 10 · requirements_building 8 · adoption 4 · post_purchase 4.
**Role:** platform eng 14 · site operator 12 · SMB owner 12 · IT/sec eng 7 · budget owner 3 · compliance reviewer 2.
**Locale:** en-US 49 · en-GB 1 (quarantined).
**Lane:** closed_model 50 · retrieval 50 · consumer_surface 44 · campaign_experiment 0.
**Turn form:** single 48 · scripted multi-turn 2.
**Evidence grade:** **A 0** · B 43 · C 6 · D 1. No grade-C or grade-D cell sits in core.

### Required waivers — what is missing and why

| Gap | Reason | Evidence needed to close it |
| --- | --- | --- |
| Acts `generate`, `navigate` | No approved job entails either. Forcing them would manufacture demand | Support tickets, on-site search, or a permitted AI-conversation corpus |
| All non-English locales; en-GB quarantined | No locale evidence, no market-competent reviewer. Machine translation ≠ locale equivalence | Locale-specific buyer evidence + native reviewer for transcreation |
| Grade A everywhere | No first-party behavioural corpus supplied | Interviews, calls, on-site search logs, win/loss, or permitted AI-conversation data |
| Edge compute / developer platform (in your description) | Company-asserted capability, **zero** buyer-behaviour evidence | Developer survey, community threads, or on-site search queries |
| Post-purchase depth | Only source (src-011) was snippet-only; review + community pages returned 403 | Direct access to peer review sites and the vendor community forum |
| Rotating partition runnable in wave 1 | All 6 cells quarantined or awaiting revision | Close the four gaps above |
| `campaign_experiment` lane | No campaign supplied | Campaign definition, dates, treatment/control assignment, pre-registration |
| Funnel rollup (`null` on all 50 cells) | No evidence maps these jobs to a funnel model — and this panel contains the counterexamples: cell-029 is a B0 **post-purchase** cell, cell-003 is a B3 problem cell closer to an urgent spend decision than several supplier-selection cells | Journey research linking these jobs to a purchase model |
| Randomization seed, content hashes | No hashing runtime available | Generate and freeze before wave 1 |

---

## 6. QA ledger

95 decisions: **86 pass · 1 revise · 5 quarantine · 3 reject.**

- **Contamination failures:** 2 — one `proprietary_categories` hit ("connectivity cloud"), one `competitor_terms` hit in an unaided cell. **Unaided-core target/campaign leaks: 0.**
- **Ambiguous-term reviews cleared:** 12 (CDN, WAF, ZTNA, SASE, "bot management", "zero trust"). These are ordinary category words that collide with product names; they were routed to review and cleared, never auto-failed. Two are queued for Gate 3.
- **Duplicates:** 0 exact merges, 0 semantic merges, **11 splits** with recorded protected differences (aided status, information act, journey, persona, material constraint, turn form). Nothing was auto-deleted; embedding similarity was not available and is recorded as a method limitation.
- **Revise:** prompt-035a — "last year" is a relative time reference that drifts across the panel's own quarterly cadence. Routed back to generation. cell-035 therefore has **zero** accepted candidates.
- **Quarantine:** prompt-018a (grade C on a hypothesis_only job), prompt-044a and prompt-045a (edge compute; 045a is the only grade-D candidate), prompt-046a (snippet-only source), prompt-047a (locale review pending).
- **Coverage lost to QA:** the procurement-criteria intent, both edge-compute probes, the unaided post-purchase support intent, the only non-US locale, and the outage retrospective currently have no runnable candidate.
- **Blinding confirmed:** `baseline_fields_blinded: true`. No visibility result, ranking, mention count or target-performance figure entered prompt selection at any point.

---

## 7. Tracking plan

Full detail in `tracking_plan.md`. Summary: 2 variants per core/sentinel/aided cell, 1 per rotating/control; 3 repeats per wave (6 for the 12 pilot cells across ≥2 time blocks); fresh session per observation, no session reuse across cells; retrieval state, generated queries and citations captured per run; **equal** exposure and priority weights held separately with an explicit non-prevalence limitation; Wilson intervals for unweighted binary strata and a cluster bootstrap on `canonical_cell_id` for weighted aggregates; randomised cell and variant order with an as-yet-**ungenerated** seed; monthly evidence intake, quarterly review, next review **2026-08-25**; B5 cells 032/033/051 must be retired or re-rendered by **2026-09-15**. Planned volume 1,590 observations/wave against a 2,000 ceiling.

---

## 8. Human gates

| Gate | Status | Approver |
| --- | --- | --- |
| 1 — ICPs, exclusions, permissions | **pending** | none |
| 2 — Jobs, language, roles, locales | **pending** | none |
| 3 — Partitions and disputed QA | **pending** | none |
| 4 — Weights, limitations, cadence, freeze | **pending** | none |

Every gate is resumable from its artifact. Nothing was silently promoted.

**Queued for Gate 3:** (1) does turn form alone justify splitting cell-003 from cell-036? (2) is a *paraphrase* of vendor policy framing acceptable in an unaided B5 cell (prompt-032b)? (3) does prompt-029b break the one-concept rule by combining debugging with support timing? (4) is "bot management" safe as generic category language when it is also a product name (prompt-019a)? (5) promote prompt-018a and prompt-046a from quarantine, or hold for direct source access?

### What would change this panel

1. **A first-party corpus** (interviews, on-site search, support tickets, permitted AI conversations) would create the panel's first grade-A cells and would replace equal exposure weights with evidence-based ones. This is the single highest-leverage input.
2. **Direct access to peer review sites and the vendor community forum** (both 403 here) would upgrade src-011 and src-015 from C to B and unlock the post-purchase and competitor-aided strata.
3. **Any credible exposure-prevalence evidence** would retire the equal-weight limitation — the largest caveat on every reported number.
4. **A named approver plus a run budget** would let Gates 1–4 close and move the panel from `provisional_directional` toward a frozen v1.0.0.
5. **The variance pilot** would replace `directional_only` with a real precision statement and reallocate cells versus repeats.
6. **Locale evidence plus a market-competent reviewer** would make anything beyond en-US measurable.
7. **On 2026-09-15** the B5 evidence expires. Cells 032, 033 and 051 must be re-verified against what actually shipped, then retired or re-rendered under a new version with an overlap bridge.
