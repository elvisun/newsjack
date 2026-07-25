# Deriving defensible prompt panels for AEO, GEO, and AI visibility

- **Status:** implementation-ready research recommendation
- **Research cutoff:** 2026-07-25
- **Scope:** prompt derivation, panel design, measurement, and campaign attribution; no proposed skills are implemented here
- **Privacy:** all examples are synthetic and use public information only

## Executive decision

Newsjack should not build a “prompt generator.” It should build a **prompt-panel design workflow**.

The durable product artifact should be a versioned, user-owned panel that records:

- what population and decision the panel is meant to describe;
- which company, customer, market, and behavioral evidence supports each prompt;
- whether a prompt is aided, competitor-aided, or unaided;
- the buyer job, information need, journey state, prompt-proximity band, persona, locale, and measurement surface;
- which wording variants belong to the same canonical intent cell;
- exposure weights separately from business-priority weights;
- the lane in which the prompt may run: closed-model, retrieval/citation, consumer surface, or campaign experiment;
- the panel’s frozen core, rotating discovery cells, controls, version, and change history; and
- the uncertainty that comes from both stochastic answers and a necessarily incomplete, non-probability prompt panel.

The core design principle is:

> **Use company evidence to establish standing, customer evidence to establish the need and language, and model judgment only to organize and expand the evidence. Never let the target brand’s copy, its current AI answers, or a campaign slogan define the unaided panel that will later “measure” that brand.**

This is stricter than common vendor practice. The public vendor methods reviewed here are useful operational references, but most start with a brand domain, product description, SEO terms, or accepted prompt suggestions and then run those prompts daily. That is efficient, but it can produce a self-confirming measurement system. Some vendors offer behavioral prompt data or search-backed demand proxies, which improves discovery, yet their public methods generally do not disclose enough about sampling frames, coverage error, panel selection, weighting, or variance to justify calling the resulting score representative of all category buyers.

The recommended composition is six atomic skills plus one orchestration skill:

1. `icp-evidence-analysis`
2. `buyer-job-intent-analysis`
3. `prompt-proximity-architecture`
4. `realistic-prompt-generation`
5. `prompt-set-qa`
6. `ai-visibility-panel-design`
7. `build-ai-visibility-panel` as the orchestrating molecule

Newsjack already has useful evidence, provenance, staged-filtering, deduplication, freshness, local-state, and human-handoff patterns. The first implementation slice should therefore be schemas, fixtures, and deterministic validators—not a new runner or dashboard.

## 1. What is being measured

The field uses “AEO,” “GEO,” “AI visibility,” “share of model,” and “share of model mind” inconsistently. Newsjack should define the estimand before it generates a prompt.

### 1.1 Required estimands

| Estimand | Exact question | Valid numerator | Denominator | What it does not prove |
| --- | --- | --- | --- | --- |
| Unaided brand presence | When the target is not named, how often is it mentioned for this panel? | Eligible responses containing a verified target alias | Eligible prompt runs in the unaided panel | Population reach, preference, causation, or sales |
| Aided brand knowledge | When asked about the target, how often is the answer accurate, complete, or favorable? | Responses meeting a declared fact/quality rule | Eligible direct-brand runs | Unaided awareness |
| Competitive mention share | Among tracked eligible brand mentions, what fraction belongs to the target? | Target mentions | Target plus declared competitor mentions | Share of all buyers or of the whole market |
| Citation presence | When retrieval is enabled, how often is an owned or earned source cited? | Runs containing an eligible cited domain/URL | Eligible retrieval runs | That a human saw or clicked the citation |
| Answer framing | How is the target characterized on declared attributes? | Audited classifications or scores | Eligible target-containing responses | Sentiment in the market at large |
| Campaign response | Did a pre-registered intervention change a declared panel outcome relative to a credible counterfactual? | Treatment-minus-control outcome | The experimental design’s eligible units | Revenue impact unless revenue is an outcome |

“Share of model mind” should be treated as a report label, not a scientific quantity. In Newsjack artifacts its machine name should be explicit, such as `competitive_mention_share`, `weighted_unaided_presence`, or `citation_presence`. A score without its prompt panel, lane, surfaces, time window, weights, and denominator is not reproducible.

### 1.2 Two weights, never one ambiguous score

Newsjack should calculate two distinct rollups:

- **Exposure-weighted visibility** estimates performance over the panel’s best available evidence of audience, intent, locale, and surface prevalence.
- **Priority-weighted visibility** emphasizes jobs or segments the company has chosen as strategically valuable.

The second is a planning index, not audience reach. The UI and report must never label it “market share,” “consumer awareness,” or “share of users.” Neither weighting scheme may depend on the target’s baseline visibility, the size of a current gap, or whether a prompt makes a campaign look successful.

## 2. What competent teams do now

### 2.1 Vendor and practitioner method comparison

The table distinguishes **what the public method actually supports** from the claim a buyer might infer.

| Method | Publicly described prompt source and operation | Useful contribution | Evidence gap or contamination risk |
| --- | --- | --- | --- |
| Profound | Profound says Prompt Volumes licenses conversations from ChatGPT, Gemini, Claude, and Perplexity, then cleans and probabilistically models them; it exposes example prompts, intent, region, demographics, and historical topic volume ([Profound Prompt Volumes](https://help.tryprofound.com/articles/4288109168-prompt-volumes), accessed 2026-07-25). Tracking prompts can also be generated, entered manually, bulk-uploaded, and tagged ([Profound prompt management](https://help.tryprofound.com/articles/3730240593-create-manage-and-tag-prompts), accessed 2026-07-25). | Behavioral language and topic demand are stronger discovery evidence than brand-copy-only generation. The separation between research data and a tracked set is sound. | The public page does not disclose the licensed sources’ sampling frames, selection probabilities, deduplication effects, probabilistic model validation, or uncertainty. Geographic coverage is explicitly uneven. Treat volumes and demographics as proprietary modeled estimates, not census counts. |
| Semrush | Semrush describes a 289-million-prompt database across several AI surfaces, combining third-party AI interactions with machine learning; because individual prompts are often unique, it aggregates demand at topic level. Custom prompts run daily while the broader database refreshes monthly ([Semrush AI Visibility data](https://www.semrush.com/kb/1607-semrush-ai-visibility-data), accessed 2026-07-25; [Prompt Research report](https://www.semrush.com/kb/1597-prompt-research-report), accessed 2026-07-25). | Correctly treats exact-string volume as sparse and uses topic-level demand for discovery. Separates a large research corpus from daily custom tracking. | The public method does not expose the third-party data’s sampling frame, the topic-volume estimator, platform mixture, coverage error, or confidence bounds. |
| Ahrefs Brand Radar | Ahrefs derives “search-backed prompts” from its keyword database and People Also Ask, expands them semantically, runs millions of questions, refreshes monthly, and labels estimated impressions as potential rather than actual reach ([Ahrefs Brand Radar methodology](https://ahrefs.com/blog/brand-radar-methodology/), accessed 2026-07-25; [Ahrefs Brand Radar help](https://help.ahrefs.com/en/articles/11064852-what-is-brand-radar-and-how-to-use-it), accessed 2026-07-25). | Large-scale, behaviorally anchored discovery; unusually clear warning that estimated impressions are potential exposure. | Search queries and People Also Ask are proxies for AI prompts, not observed AI conversations. Public documentation does not establish how repeated answers, prompt variants, or uncertainty are handled. |
| Scrunch | Scrunch recommends defining brand, market, personas, competitors, and topics; sourcing prompts from paid search, SEO, customer calls, support, communities, hand-writing, or LLM expansion; then tagging by topic, persona, funnel, region, and campaign ([Scrunch monitoring guide](https://scrunch.com/guides/ai-search-guide/monitoring/), accessed 2026-07-25). It recommends stable prompt clusters and repeated monitoring, with directional attribution from self-report, referral traffic, and visibility ([Scrunch monitoring questions](https://scrunch.com/blog/ai-search-monitoring-questions-answered), accessed 2026-07-25). | Broad source mix, operational taxonomy, customer-language emphasis, and explicit acknowledgement that only platforms observe exact prompt volume. | Domain-inferred personas and “must-win” topics can encode brand strategy as buyer demand. Claims that agent crawler visits indicate “real consideration,” or that presence/citations imply commercial impact, are not established by the cited method. Its public accuracy claim describes binary mention detection and aggregation but does not publish a labeled accuracy study or interval ([Scrunch accuracy FAQ](https://scrunch.com/faqs/how-accurate-is-scrunch-at-tracking-brand-presence-across-llms/), accessed 2026-07-25). |
| Peec AI | Peec recommends prompts containing intent plus audience, use case, or constraints; suggestions are generated from the website, industry context, brand profile, topics, and accepted prompts. A relative 1–5 prompt-volume score uses search trends and business-specific weighting ([Peec prompt setup](https://docs.peec.ai/setting-up-your-prompts), accessed 2026-07-25). | Clear prompt anatomy, locale metadata, topic organization, and manageable review workflow. | Its documentation says exact wording does not matter much and advises adding “what tools should I use?” when an informational prompt would not elicit brands. No validation is supplied for the wording claim; the second instruction changes the user’s job to manufacture recommendation opportunities. Website- and brand-profile-conditioned suggestions need a blind unaided lane. The 1–5 score is relative and partially business-weighted, so it is not prompt frequency. |
| Otterly.AI | Otterly recommends AI-assisted prompt research, query fan-out variants, Search Console/Bing data, surveys, sales and support language, then tagging by product, funnel, country, and language ([Otterly prompt research](https://help.otterly.ai/onboarding3), accessed 2026-07-25). It says prompts run daily and distinguishes mentions, citations, and estimated intent volume ([Otterly prompt monitoring](https://help.otterly.ai/search-prompt-monitoring), accessed 2026-07-25). | Good mix of first-party and search-proxy sources; useful localization and lifecycle tags. | Its collection page says public web interfaces capture what customers see ([Otterly data collection](https://help.otterly.ai/how-otterlyai-collects-data), accessed 2026-07-25). A clean automated browser is a valuable surface, but it cannot be “identical” to all customers’ personalized, logged-in, experimental, or locale-specific experiences. |
| Evertune | Evertune describes an almost 25-million-person “EverPanel” balanced to internet demographics and three survey-like prompt families: unaided category questions, attribute-specific questions, and aided brand review/association, repeated across models ([Evertune prompt tracking](https://www.evertune.ai/resources/insights-on-ai/how-evertunes-prompt-tracking-works), accessed 2026-07-25). | The unaided/aided split and attribute-level design are the closest public analogue to brand research practice. | The article calls differences statistically significant without publishing the cell sample, selection probabilities, response protocol, variance estimator, or intervals. Its claim that the gap between brands is “lost business” is a causal leap, not a measurement result. |
| Adobe LLM Optimizer | Adobe generates an initial small set of categories, topics, and prompts from the onboarded brand/domain, after which users customize them; the documentation explicitly notes that exact prompts used by real LLM users are not publicly disclosed ([Adobe LLM Optimizer quick start](https://experienceleague.adobe.com/en/docs/llm-optimizer/using/essentials/quick-start), accessed 2026-07-25). | Useful candor about the observability gap; separates prompt presence from agentic and referral traffic. | Brand/domain-generated prompts are hypotheses, not buyer samples. The public quick start does not describe a population, weighting, contamination control, or uncertainty model. |
| HubSpot AEO | HubSpot asks for brand, competitors, products, ICPs, and prompts, runs tracked prompts daily, and reports prompt coverage and competitor presence ([HubSpot AEO documentation](https://knowledge.hubspot.com/seo/set-up-and-analyze-ai-visibility), accessed 2026-07-25). Its practitioner guide suggests calls, tickets, forums, personas, pain points, categories, 100–200 seeds, and topic/intent/region/funnel tags ([HubSpot prompt tracking guide](https://blog.hubspot.com/marketing/aeo-prompt-tracking), accessed 2026-07-25). | Strong operational ownership, CRM/customer evidence, prompt taxonomy, and daily longitudinal tracking. | “Fewer than 50 won’t give statistically meaningful citation data” is not justified without an estimand, variance, clustering, or desired precision. Mapping a target page to every prompt can contaminate selection if done before the panel is frozen. Its 1,850% leads claim is a company case result, not validation of prompt representativeness or attribution generally. |
| LLM Pulse | LLM Pulse describes its UI prompts as synthetic and explicitly says they do not represent real user journeys ([LLM Pulse methodology](https://llmpulse.ai/help-center/how-llm-pulse-works), accessed 2026-07-25). | A useful transparency standard for generated panels. | Synthetic prompts are suitable for controlled diagnostics, not population prevalence, unless validated and weighted with external evidence. |

No reviewed public method establishes a probability sample of all relevant AI-query users. That does not make vendor panels useless. It means their proper role is one of:

1. observed-language discovery with documented but incomplete coverage;
2. search-behavior proxy discovery;
3. controlled synthetic diagnostics; or
4. repeated measurement of a declared custom panel.

The category’s common error is to slide from role 4 to “market awareness” without establishing roles 1–3, weights, or uncertainty.

### 2.2 The strongest practitioner pattern

The most defensible public practitioner advice is to begin with behavioral language. Seer Interactive recommends paid-search query data as a fast floor for how people phrase needs, followed by sales-call transcripts and customer language rather than internally imagined prompts ([Seer Interactive](https://www.seerinteractive.com/insights/are-you-picking-the-prompts-for-ai-tracking-without-observing-real-customers), accessed 2026-07-25). Scrunch and HubSpot likewise name calls, tickets, communities, and search terms, although both mix these with domain-conditioned or persona-generated prompts.

The practical synthesis is:

- **Observed AI conversations**, when lawfully obtained with adequate provenance, are the closest evidence of prompt form and topic prevalence.
- **Customer and prospect language** in calls, interviews, support, on-site search, surveys, and paid/organic queries is the strongest company-specific evidence of needs and vocabulary.
- **Public market language** in reviews, forums, procurement documents, community discussions, and competitor reviews broadens coverage beyond existing customers.
- **Search queries and People Also Ask** are useful behavioral proxies, especially for retrieval-shaped prompts, but they are not AI-prompt frequency.
- **Company copy, SEO plans, and product strategy** establish what the company does and where it has standing; they do not prove that buyers ask a question.
- **LLM-generated prompts** are expansion hypotheses. They do not earn core-panel weight until validated against independent evidence or deliberately labeled synthetic.

## 3. Research foundations Newsjack should preserve

### 3.1 Information need is not the same as funnel stage

Broder’s classic web-search taxonomy separates navigational, informational, and transactional intent and warns that a short query only imperfectly reveals the underlying need ([Broder, “A taxonomy of web search,” 2002](https://sigir.hosting.acm.org/files/forum/F2002/broder.pdf), accessed 2026-07-25). Bates’s berrypicking model describes an information need and query as evolving while a person encounters new information, rather than as one fixed retrieval event ([Bates, 1989](https://pages.gseis.ucla.edu/faculty/bates/berrypicking.html), accessed 2026-07-25).

These foundations matter more in conversational systems, where one user may move from diagnosis, to requirements, to a shortlist, to implementation in a single thread. A panel therefore needs at least four independent axes:

- **buyer job:** the progress the person is trying to make;
- **speech act/information intent:** explain, diagnose, generate, compare, recommend, verify, navigate, transact, or implement;
- **journey state:** problem identification, exploration, requirements building, supplier selection, adoption, or post-purchase;
- **prompt proximity:** how much brand or category language the prompt supplies.

TOFU/MOFU/BOFU remains a useful report rollup, but it cannot be the schema.

### 3.2 Buyers move through jobs, not a clean funnel

Jobs-to-be-Done research focuses on the progress a person is trying to make in a struggling moment and the forces that support or resist a change ([Jobs to Be Done, Bob Moesta](https://jobstobedone.org/), accessed 2026-07-25; [Christensen et al., Harvard Business Review](https://hbr.org/2016/09/know-your-customers-jobs-to-be-done), accessed 2026-07-25). Google’s “messy middle” research models exploration as expansive and evaluation as reductive rather than a one-directional path ([Google, “Navigating purchase behavior”](https://business.google.com/en-all/think/consumer-insights/navigating-purchase-behavior-and-decision-making/), accessed 2026-07-25). Gartner’s B2B journey similarly describes problem identification, solution exploration, requirements building, and supplier selection as buying jobs that recur rather than line up neatly ([Gartner B2B buying journey](https://www.gartner.com/en/sales/insights/b2b-buying-journey), accessed 2026-07-25).

Newsjack should use those models as question-generating lenses, not universal truth. Each ICP and job hypothesis still needs company-specific evidence.

### 3.3 Prompt form and answer variance are empirical questions

Public conversational corpora show why keyword-style templates alone are inadequate. LMSYS-Chat-1M contains one million real-world conversations across 25 models and 210,000 IP addresses ([Zheng et al., ICLR 2024](https://proceedings.iclr.cc/paper_files/paper/2024/hash/5f9bfdfe3685e4ccdbc0e7fb29cccf2a-Abstract-Conference.html), accessed 2026-07-25). WildChat reports roughly one million conversations, an average 2.52 turns, about 41% multi-turn conversations, and material multilingual usage ([Zhao et al., 2024](https://arxiv.org/abs/2405.01470), accessed 2026-07-25). OpenAI’s privacy-preserving study of 1.5 million consumer conversations found broad “Asking” and “Doing” work, not merely shopping or search ([OpenAI, “How people are using ChatGPT”](https://openai.com/index/how-people-are-using-chatgpt/), accessed 2026-07-25).

These corpora can help test naturalness, length, task form, multilingual handling, and multi-turn coverage. They cannot provide category-specific market weights: their users, collection surfaces, periods, geographies, and tasks do not match an arbitrary client’s buyers.

Output variance also changes the sample design. NIST recommends decomposing variance between items and within repeated trials: when between-item variance dominates, more unique items are more useful; when within-item variance is material, repeated trials are more useful, and treating correlated repetitions as independent understates uncertainty ([NIST AI 800-3, 2026](https://nvlpubs.nist.gov/nistpubs/ai/NIST.AI.800-3.pdf), accessed 2026-07-25). A 2026 preprint measuring four Swiss-German verticals found substantial within-day answer, source, and brand variation and recommends repeated measurements; its exact run-count findings should not be generalized because the study used only eight prompts per vertical and a particular prompt-generation and engine setup ([Schulte, Bleeker, and Kaufmann, “Don’t Measure Once,” preprint](https://arxiv.org/abs/2604.07585), accessed 2026-07-25).

The implication is not “always run each prompt ten times.” It is “run a variance pilot, then allocate budget between unique cells and repetitions.”

### 3.4 Wording and order are treatment variables

Survey-method practice supplies useful safeguards. Pew notes that small wording changes and question order can change responses, and recommends avoiding order effects and testing questions ([Pew Research Center, writing survey questions](https://www.pewresearch.org/writing-survey-questions/), accessed 2026-07-25). AAPOR recommends simple target-population language, one concept at a time, avoidance of leading wording, cognitive testing, and piloting alternatives ([AAPOR best practices](https://aapor.org/standards-and-ethics/best-practices/), accessed 2026-07-25).

An LLM is not a survey respondent, but the design lesson transfers: wording and preceding context are part of the experimental condition. They must be controlled or intentionally varied, not dismissed.

### 3.5 Content-optimization evidence is not prompt-sampling evidence

The original GEO research framed content optimization as a black-box problem and reported improvements on its benchmark visibility metric ([Aggarwal et al., KDD 2024](https://arxiv.org/abs/2311.09735), accessed 2026-07-25). It did not estimate the distribution of real buyer prompts, demonstrate population-level brand awareness, or show revenue causality. Newsjack should not use content-optimization studies as evidence that a prompt panel is representative.

## 4. Recommended end-to-end process

Every stage produces a reviewable artifact. Later stages may reject or reclassify earlier hypotheses, but they must not silently rewrite their evidence.

### Stage 0: write the measurement charter

Before researching prompts, record:

- the business decision;
- the estimand or estimands from section 1;
- the target population and exclusions;
- products, markets, locales, and time horizon;
- surfaces and lanes;
- desired reporting strata;
- maximum run and review budget;
- acceptable precision or explicit “directional only” status; and
- the owner who can approve ICP, job, and panel choices.

**Gate 0:** reject a charter that says only “track our AI visibility.” It must state which kind, among whom, where, and for what decision.

### Stage 1: build a company-evidence dossier

Collect source-bound facts from:

- the company’s current public product, pricing, integration, security, and support pages;
- supplied strategy, positioning, and customer material the user is authorized to use;
- regulatory filings, certifications, or public technical documentation where relevant;
- credible third-party reviews and coverage;
- public competitor category language; and
- existing Newsjack monitor profiles as leads, never as unquestioned truth.

For each claim record the source URL or file, excerpt or span, date, fact type, confidence, and whether it is company-asserted or independently supported.

Outputs include:

- products and capabilities;
- served segments and geographies;
- declared alternatives and competitors;
- proof assets and limitations;
- exact brand/product/domain aliases;
- slogans, proprietary category terms, campaign vocabulary, and flattering adjectives for the contamination lexicon.

**Gate 1:** a human confirms the factual perimeter and redacts material that may not be used in prompt design.

### Stage 2: form ICP hypotheses

An ICP hypothesis is a testable description of an organization or consumer context in which the product may fit. It is not a generated persona biography.

Each hypothesis should include:

- organization or household context;
- triggering condition;
- likely user, champion, economic buyer, approver, blocker, and post-purchase user where applicable;
- constraints and disqualifiers;
- evidence for and against;
- confidence;
- unresolved research questions; and
- the company capability that gives it standing.

Company positioning may nominate a segment, but at least one independent source should support a core-panel ICP: customer/prospect evidence, observed behavioral data, credible public demand, or a human promotion of the hypothesis with an explicit low-confidence label.

### Stage 3: recover buyer jobs, intents, and language

Build a source ledger in descending order of evidentiary value:

1. lawfully collected, relevant AI prompt or conversation samples with collection metadata;
2. customer/prospect interviews and call transcripts;
3. on-site search, support tickets, chat, sales questions, and win/loss research;
4. paid-search terms, Search Console queries, marketplace search, and internal site search;
5. public reviews, forums, communities, RFPs, procurement guides, and competitor reviews;
6. broad search and People Also Ask proxies;
7. company copy; and
8. LLM-generated hypotheses.

Extract:

- struggling moment or trigger;
- desired progress or outcome;
- current workaround;
- anxieties, habits, switching forces, and constraints;
- information need or requested action;
- exploration/evaluation state;
- decision criteria and proof sought;
- exact source language;
- persona/role and locale when known; and
- contradictions or negative evidence.

Use an evidence grade:

| Grade | Meaning | Core-panel use |
| --- | --- | --- |
| A | Direct relevant behavior or verbatim customer/prospect language with provenance | Eligible for weight and wording |
| B | Credible public market behavior or search proxy with provenance | Eligible with proxy label |
| C | Company assertion or expert hypothesis | Hypothesis; requires approval or validation |
| D | LLM-generated expansion without independent support | Discovery pool only |

**Gate 2:** a human approves the jobs and chooses which are important enough to represent. No prompt selection has occurred yet.

### Stage 4: create a blind design brief

Build the generation brief from approved jobs, constraints, role labels, language examples, and evidence IDs, then remove:

- target brand and product names;
- target domains and founder/executive names;
- slogans and campaign language;
- proprietary category phrases used only by the target;
- target-favorable adjectives and claims;
- current tracked answers, target mentions, rankings, and cited pages; and
- “target pages” the marketing team wants cited.

The generator sees this blind brief. The QA stage separately receives the contamination lexicon so it can detect leakage.

### Stage 5: construct the prompt universe

Create candidates independently within evidence-source strata so a large synthetic batch cannot crowd out a small set of observed prompts.

For every approved job, generate or extract candidates across:

- prompt-proximity bands in section 5;
- information acts: explain, diagnose, plan, compare, recommend, verify, navigate, buy, implement, and troubleshoot;
- exploration and evaluation;
- relevant journey states, including post-purchase;
- authentic persona constraints;
- locale and language;
- single-turn and, only when evidence supports it, separately scripted multi-turn forms; and
- concise, contextual, imperfect, and follow-up wording styles observed in evidence.

Every candidate records its source evidence and transformation:

- `verbatim`
- `lightly_normalized`
- `search_query_expanded`
- `human_written`
- `llm_expanded`
- `translated`
- `locale_transcreated`

Generated candidates are never assigned observed frequency merely because they resemble a high-volume topic.

### Stage 6: apply prompt-proximity architecture

Assign one primary proximity band and independent tags for job, intent, journey state, and funnel rollup. Use the rules in section 5. A candidate that cannot be classified without guessing returns to the job stage or stays in discovery.

### Stage 7: generate controlled variants

The canonical unit is an **intent cell**, not an exact string. An intent cell fixes:

- the underlying job;
- journey state;
- information act;
- material constraints;
- persona and locale;
- proximity band; and
- expected kind of answer.

Create two core wording variants by default:

- one closest to observed language;
- one natural paraphrase that preserves the cell.

Use more variants only in a wording-sensitivity pilot or rotating panel. Do not generate a combinatorial persona × locale × constraint × style grid. Select combinations supported by the sampling plan.

### Stage 8: run contamination and panel QA

Run deterministic checks first, then semantic review:

1. schema and provenance completeness;
2. target/campaign/alias lexical scan;
3. forbidden answer-derived field scan;
4. prompt length, language, and one-concept checks;
5. exact normalization and duplicate hashes;
6. lexical and embedding near-duplicate candidate generation;
7. semantic same-intent review;
8. evidence-to-prompt entailment;
9. naturalness and persona/locale authenticity;
10. proximity, journey, and aided-status consistency;
11. commercial-leading and recommendation-forcing detection; and
12. blind human review without baseline visibility scores.

Failures return to the owning atom. A QA skill must not “fix” an ICP or invent evidence.

### Stage 9: select and weight the panel

Selection is stratified, not top-N by a vendor volume score. Define minimum and target allocations across:

- proximity band;
- buyer job and journey state;
- information act;
- ICP/persona;
- locale/language;
- evidence grade/source type; and
- surface/lane.

Preserve underrepresented but strategically necessary strata. Within a stratum, select by evidence strength, language authenticity, decision relevance, and diversity. Select blind to current brand performance.

Store weights as components with provenance:

```yaml
weight:
  exposure:
    value: 0.0125
    basis:
      - factor: intent_prevalence
        source_id: source-search-042
        confidence: medium
      - factor: locale_share
        source_id: source-crm-aggregate-003
        confidence: high
  priority:
    value: 0.025
    multiplier: 2.0
    rationale: "Human-approved expansion segment"
```

If credible exposure weights do not exist, report equal-weighted panel results by stratum. Equal weighting is more honest than invented precision.

### Stage 10: pilot and validate

Before freezing:

- conduct cognitive review with 3–5 people who know the buyer language, ideally including customer-facing staff and at least one target-role participant;
- run 12–20 deliberately diverse sentinel intent cells repeatedly to estimate within-cell variance;
- inspect wording sensitivity, model/surface differences, citation variance, invalid/refusal rates, and brand-alias detection;
- calculate between-cell and within-cell variance;
- revise repetitions, cell allocation, and strata;
- label the panel directional if precision or coverage remains weak; and
- archive rejected prompts and reasons.

Validation must not select only prompts on which the target performs well or poorly. Baseline results remain hidden from prompt selectors until the panel is frozen.

### Stage 11: freeze, run, and track longitudinally

A panel release receives:

- immutable `panel_id` and semantic `version`;
- content hashes for prompts and configuration;
- frozen core, rotating, sentinel, and control partitions;
- model/surface configuration;
- run randomization seed and session policy;
- weight version;
- metric definitions;
- campaign registry linkage, if any; and
- a change ledger.

Every observation records the exact prompt, model and version as exposed by the provider, surface, locale, search/tool policy and invocation, temperature or sampling controls when available, system/developer prompt, session state, response/citation payload hashes, timestamps, retry status, and parser version.

### Stage 12: refresh without destroying the time series

Recommended default:

- **monthly intake:** collect new evidence and candidates without changing the core;
- **quarterly panel review:** retain roughly 70–80% core, rotate 15–25% discovery cells, and keep 5–10% sentinels/controls;
- **event-triggered review:** product/category change, new locale, material model or surface change, regulatory event, or demonstrated buyer-language shift;
- **weekly closed-model and retrieval waves** for normal tracking;
- **daily or 2–3-times-weekly retrieval waves** only for active fast-moving campaigns or news;
- **annual charter review:** re-approve population, estimands, and weighting sources.

The percentages and cadences are starting defaults, not empirical laws. A user may tune them, but changing the core, weights, metrics, or surface mix creates a new panel version. Reports should show an overlap bridge so longitudinal comparisons use the unchanged cells before and after the transition.

## 5. Prompt-proximity bands and journey mapping

Prompt proximity measures how much of the answer space the prompt supplies. It is deliberately separate from purchase stage.

| Band | Prompt structure | Aided status | Common journey/funnel mapping | Important exception |
| --- | --- | --- | --- | --- |
| `B0_direct_brand_product` | Names the target brand/product and asks about facts, fit, use, reputation, support, or implementation | Aided | Often BOFU, adoption, or post-purchase | “How do I export from Brand X?” is not purchase intent |
| `B1_comparison_purchase` | Requests shortlist, recommendation, alternatives, pricing, requirements, or comparison; unaided core omits target | Unaided or `competitor_aided` | BOFU; evaluation, requirements, supplier selection | A first-time explorer can ask “best” without being purchase-ready |
| `B2_category` | Names the accepted solution category but not target | Unaided | MOFU; solution exploration | A knowledgeable urgent buyer may be near transaction |
| `B3_problem_need` | Describes pain, risk, trigger, or constraint without category or target | Unaided | TOFU/MOFU; problem identification and requirements | An acute problem with budget can be BOFU |
| `B4_job_goal` | Asks for an outcome or progress without supplying the solution category | Unaided | Pre-funnel/TOFU; exploration | Existing users may ask a job-level implementation question |
| `B5_broad_discovery_story` | Asks about a broader trend, event, regulation, practice, or narrative connected to the job | Unaided | Pre-trigger/TOFU; discovery | Breaking events can create immediate purchase urgency |

Funnel tags should be derived from evidence and journey context, not mechanically from the band. The schema should allow `funnel: null` when the label would mislead.

### Required contamination distinctions

- `target_aided`: target name, product, domain, executive, or unmistakable slogan appears.
- `competitor_aided`: a named competitor anchors the answer set.
- `category_aided`: the category is supplied, but no brand is.
- `unaided`: neither target nor competitor is supplied.
- `campaign_exposed`: campaign wording or creative concept appears, whether or not the brand does.

Never combine these into one “visibility” denominator.

## 6. Anti-leading and contamination controls

### 6.1 The contamination register

Build a versioned register before prompt generation:

```yaml
target_terms:
  brands: []
  products: []
  domains: []
  people: []
  slogans: []
  proprietary_categories: []
  campaign_terms: []
  flattering_claims: []
competitor_terms: []
allowed_exceptions:
  - band: B0_direct_brand_product
    term_classes: [brands, products]
```

Use normalized, token, fuzzy, and semantic checks. Deterministic matches are hard failures in unaided core prompts. Semantic flags require review because ordinary words may overlap a slogan.

### 6.2 Two-pass blinding

The evidence/ICP/job atoms may see company facts. The realistic prompt generator receives:

- anonymized segment and role labels;
- approved jobs, constraints, and source-language fragments;
- evidence IDs and grades;
- required prompt strata; and
- the forbidden instruction “do not introduce products, categories, or recommendations not entailed by the evidence.”

It does not receive the target name, current answers, rank gaps, desired target pages, or campaign copy. The QA atom receives the generated set plus the contamination register.

### 6.3 Prohibited shortcuts

Do not:

- add “what tools should I use?” solely because the original problem prompt did not produce brands;
- rewrite an informational need into “best platform” to increase mention opportunity;
- generate new core prompts from answers that already mention the target;
- select prompts because the target is almost visible;
- let current cited URLs or a content backlog determine the core panel;
- treat a competitor-named alternative query as unaided awareness;
- machine-translate every prompt and assume it represents the locale;
- run sequential core prompts in one conversation, allowing earlier answers to prime later ones;
- feed campaign slogans into the frozen core after launch; or
- drop persistent zero-mention prompts merely because they hurt the score.

Answer-derived terms can enter a quarantined discovery pool. They require independent validation in customer, behavioral, or public-market evidence before the next panel version.

### 6.4 Session and surface controls

For core single-turn runs:

- use fresh stateless sessions;
- disable memory and prior-message context where possible;
- randomize prompt order;
- keep system/developer instructions minimal, fixed, and recorded;
- keep sampling controls fixed and recorded;
- separate logged-out automated surfaces from logged-in consumer surfaces;
- record locale, account state, personalization state, and experiments where observable; and
- never describe an API result as the consumer UI result.

Google documents that AI Mode may use history or connected personal context when personalization is enabled ([Google AI Search personalization](https://support.google.com/websearch/answer/17212611?hl=en), accessed 2026-07-25). Consumer-surface panels therefore need explicit account archetypes or a clean unpersonalized condition, not an unqualified “what users see.”

Multi-turn journeys belong in a separate scripted panel. They should preserve realistic follow-up dependence while holding the starting condition, turn order, and stopping rule constant.

## 7. Separate measurement lanes

### 7.1 Lane A: closed-model behavior

Purpose: measure what a declared model/API returns without external tools, supplied files, RAG, browsing, or conversation history.

Controls:

- no search or external tool;
- empty/fixed system prompt;
- new session for each single-turn run;
- fixed model/version and sampling settings;
- explicit refusal/knowledge-cutoff handling.

Call this `closed_model`, not “pure model memory.” Providers may change hidden system instructions, safety layers, routing, and model weights. The lane isolates observable no-external-retrieval behavior; it does not reveal the source of internal knowledge.

### 7.2 Lane B: retrieval and citation behavior

Purpose: measure recommendations, sources, and citations when external retrieval is required or allowed.

Store:

- whether retrieval was `required`, `allowed`, or unavailable;
- whether it actually ran;
- generated search queries where exposed;
- live versus cached/index-only mode;
- citation URLs, domains, positions, snippets, and retrieval timestamps; and
- answer-level and citation-level presence separately.

OpenAI’s API distinguishes enabling web search from requiring it and supports a cached/index-only mode via `external_web_access: false` ([OpenAI web search tool](https://developers.openai.com/api/docs/guides/tools-web-search), accessed 2026-07-25). Anthropic’s web-search tool can be available while Claude decides whether to search ([Anthropic web search tool](https://platform.claude.com/docs/en/agents-and-tools/tool-use/web-search-tool), accessed 2026-07-25). Gemini’s Google Search grounding may issue multiple searches and returns grounding/citation metadata ([Gemini Google Search grounding](https://ai.google.dev/gemini-api/docs/google-search), accessed 2026-07-25). Those states are not interchangeable.

Google also documents “query fan-out,” in which AI search issues related searches across subtopics ([Google Search AI features](https://developers.google.com/search/docs/appearance/ai-features), accessed 2026-07-25). Fan-out queries are valuable retrieval diagnostics, but they are model-generated retrieval actions—not evidence that humans use those exact prompts.

### 7.3 Lane C: consumer-surface behavior

Purpose: audit the actual product surface when API and UI behavior differ.

Use explicit cells such as:

- clean logged-out browser;
- fresh account with no history;
- declared recurring-user archetype;
- mobile versus desktop;
- country/language/location.

Keep these cells out of API rollups unless a report shows each surface separately. Automation must follow provider terms and should store only the minimum response evidence permitted.

### 7.4 Lane D: campaign experiment

Campaign prompts are not simply tags on the evergreen panel. They are a pre-registered experiment with three partitions:

1. **Frozen evergreen core:** selected before campaign language exists or held unchanged.
2. **Campaign-resonance panel:** asks about the underlying need/theme without the target; exact slogan language is allowed only when independent evidence shows people have adopted it.
3. **Aided campaign panel:** explicitly tests recognition, association, message accuracy, or branded claims.

Add matched controls covering similar jobs, stages, locales, and baseline visibility that the campaign does not target. Freeze prompt selection, metrics, models/surfaces, and analysis rules before launch. Do not choose prompts post-treatment based on where mentions moved.

A before/after increase alone is not attribution: model updates, retrieval-index changes, competitor activity, seasonality, and answer variance are alternative explanations. When randomization is possible, use time, geography, audience, or content holdouts. When it is not, use a declared comparison series and a counterfactual method with its assumptions. Bayesian structural time-series methods such as CausalImpact estimate a counterfactual from controls when experiments are unavailable, but depend on stable relationships and unaffected controls ([Brodersen et al., 2015](https://arxiv.org/abs/1506.00356), accessed 2026-07-25). Geo experiments can randomize treatment at a regional level when spillover and scale permit ([Google Research, geo experiments](https://research.google/pubs/measuring-ad-effectiveness-using-geo-experiments/), accessed 2026-07-25).

Campaign reporting must form an evidence ladder:

1. prompt-panel mention/framing/citation outcome;
2. retrieved-source and agent/referral traffic;
3. self-reported discovery;
4. qualified lead or conversion outcome; and
5. incremental business outcome from an experiment or credible counterfactual.

Never rename rung 1 “revenue attribution.”

## 8. Sampling, panel size, weighting, and uncertainty

### 8.1 The sampling unit

The primary sampling unit is the **canonical intent cell**. Wording variants and repeat runs are nested observations, not extra independent buyers.

Recommended partitions:

- `core`: stable cells for longitudinal comparison;
- `rotating`: new or lower-confidence discovery;
- `sentinel`: deliberately diverse cells used to estimate drift and run variance;
- `control`: campaign or measurement controls;
- `aided`: direct-brand and aided-recall diagnostics.

### 8.2 Starting panel sizes

These are practical starting points to be revised after the variance and coverage pilot:

| Tier | Canonical unaided cells | Variants | Repetitions per wave | Legitimate claim |
| --- | ---: | ---: | ---: | --- |
| Diagnostic pilot | 30–48 | 2 | 3, plus deeper sentinel repeats | Directional gaps within named strata |
| Standard longitudinal | 60–120 | 2 | 3, adaptively 5–8 for unstable cells | Conditional trend for this declared panel |
| Research-grade panel | 200–400 | 1–2 | Pilot-determined | More stable aggregate and subgroup estimates, still conditional on a non-probability panel |
| Campaign add-on | 24–40 target plus 24–40 matched control cells | 1–2 | Pilot-determined | Directional campaign response unless powered for a causal design |

At a worst-case 50% binary rate, a simple independent sample has an approximate 95% margin of error of ±17.9 percentage points at 30 observations, ±13.9 at 50, ±9.8 at 100, ±6.9 at 200, and ±4.9 at 400. These are orientation bounds, not the interval Newsjack should publish: weights, repeated variants, clustering, strata, and non-probability selection reduce the effective sample and limit generalization.

If a report needs a subgroup, allocate at least 20–30 distinct cells to that subgroup for even a coarse directional rate; otherwise show counts and responses, not a percentage leaderboard.

### 8.3 Allocate unique prompts versus repetitions empirically

Pilot 12–20 sentinels with 6–8 repetitions over at least two time blocks. Estimate:

- between-cell variance;
- between-variant variance within cell;
- within-variant run variance;
- day/time variance;
- model/surface variance; and
- invalid/refusal/parser variance.

If within-cell intraclass correlation is high, repetitions are redundant and budget should buy more unique cells. If within-cell answer or citation variance is high, add repetitions to unstable strata. This implements NIST’s variance-allocation principle rather than copying one vendor’s or one preprint’s fixed run count.

### 8.4 Persona, locale, and surface coverage

Do not build a full factorial. Create a sampling matrix:

1. choose the ICP/job cells that materially differ by role;
2. choose locales with evidence of demand or strategic commitment;
3. transcreate prompts with a native or market-competent reviewer;
4. choose surfaces used by the target population or needed for a diagnostic;
5. oversample small strategic strata if necessary; and
6. weight back only with credible, versioned prevalence data.

Locale is more than translation. Currency, regulation, category names, procurement practices, units, and available products may change the intent cell. A translated prompt that changes the buyer job receives a new canonical cell.

### 8.5 Published uncertainty

Every report should include:

- unique intent-cell count;
- variants and repetitions per cell;
- eligible/invalid run count;
- dates, models, versions, surfaces, locale, and lane;
- raw and weighted numerator/denominator;
- weight source/version and effective sample size;
- interval method;
- core-panel overlap with the prior period;
- model/surface changes;
- prompt-panel limitations; and
- a “conditional on this panel” label.

For simple unweighted binary strata, use a Wilson interval rather than a naive Wald interval. For weighted aggregates, use a stratified cluster bootstrap that resamples canonical intent cells and keeps variants/repetitions nested. When comparing periods, resample paired unchanged cells. When comparing a new panel version, show both:

- the overlap-only change; and
- the level under each full version.

Intervals quantify conditional sampling/run uncertainty. They do not repair coverage bias or turn an expert-curated prompt panel into a probability sample of all AI users.

### 8.6 Deduplication

Use a staged deterministic-plus-judgment process:

1. Unicode normalization, case/whitespace/punctuation normalization, and exact hash;
2. lexical similarity such as token Jaccard or MinHash to create candidate pairs;
3. embedding similarity to create additional candidate pairs;
4. rule protection for locale, persona, constraint, competitor-aided status, and information act;
5. semantic judgment: do the prompts express the same job, journey state, material constraints, and expected answer?

Merge exact duplicates automatically. Never auto-delete from an embedding threshold alone. Similarity thresholds are fixture-calibrated candidate generators. Keep one canonical intent cell with up to two representative core variants; archive other variants with their evidence and rejection reason.

## 9. Worked synthetic examples

### 9.1 LedgerLift: B2B expense operations

**Synthetic company.** LedgerLift sells expense-policy and receipt-collection software to 50–500-person professional-services firms. The hypothetical evidence dossier contains:

- eight finance-leader call excerpts about chasing receipts and missing close deadlines;
- paid-search terms around expense policy enforcement and NetSuite receipt workflows;
- support evidence that approvers care about mobile capture;
- company evidence that LedgerLift integrates with NetSuite; and
- campaign copy using the slogan “Close by Friday.”

Approved job:

> When project staff submit expenses late, help the finance team close client and company books within five working days without adding headcount or damaging employee trust.

Candidate architecture:

| Band | Prompt | Status and reasoning |
| --- | --- | --- |
| B0 | “Does LedgerLift integrate with NetSuite, and how does it handle missing receipts?” | `target_aided`; product knowledge/adoption, not awareness |
| B1 | “What expense management tools work well for a 120-person consulting firm using NetSuite?” | Unaided comparison; supported category and constraint |
| B2 | “How should a consulting firm evaluate expense management software for policy enforcement?” | Unaided category/requirements |
| B3 | “How can finance stop chasing consultants for receipts at month end?” | Unaided problem; closest to verbatim call language |
| B4 | “Help me close employee expenses in five working days without adding finance headcount.” | Unaided job/goal |
| B5 | “How are finance teams changing expense controls as client work becomes more distributed?” | Broad discovery hypothesis; rotating until public-market evidence supports it |

Rejected or quarantined:

- “What platform helps finance teams Close by Friday?” — target campaign language contaminates the evergreen panel and forces a product answer.
- “Which tools are best for eliminating receipt chaos?” — “receipt chaos” appears only in company copy; no buyer evidence.
- “How can I collect receipts?” rewritten as “Which expense platform should I buy?” — changes an implementation/problem act into a recommendation act.

Campaign design:

- keep the pre-campaign B1–B4 core frozen;
- add resonance cells such as “How can a consulting finance team shorten its expense close to five days?” only if the five-day goal predates the campaign in customer evidence;
- keep “Close by Friday” in an aided campaign-association panel;
- match controls on other finance-operations jobs not addressed by the campaign;
- treat mention lift as an intermediate outcome; require a holdout/counterfactual plus referral, self-report, or pipeline evidence for stronger attribution.

### 9.2 HarborHeat: local consumer decision

**Synthetic company.** HarborHeat is a heat-pump installer operating in Ontario. Its hypothetical evidence includes public rebate pages, calls from homeowners with cold upstairs rooms, search terms about replacing gas furnaces, and reviews concerned about winter performance and electrical upgrades.

One job has materially different locale cells:

- Ontario English: “Can a cold-climate heat pump replace my gas furnace in a 1970s Toronto semi, and what electrical upgrades might I need?”
- Ontario French: a native-reviewed transcreation using Canadian equipment and rebate terminology;
- United States: rejected from the Ontario panel because incentives, codes, currency, and contractor availability change the decision.

Proximity examples:

| Band | Prompt | Journey interpretation |
| --- | --- | --- |
| B0 | “Is HarborHeat licensed to install cold-climate heat pumps in Toronto?” | Aided verification |
| B1 | “Which Toronto installers should I compare for a cold-climate heat pump in a 1970s semi?” | Supplier selection/BOFU |
| B2 | “What should I compare when choosing a cold-climate heat-pump installer?” | Requirements/MOFU |
| B3 | “My upstairs stays cold and my gas furnace is near end of life. What should I investigate?” | Problem identification; may be high urgency despite TOFU-like wording |
| B4 | “How can I heat an older Toronto home reliably in winter with lower household emissions?” | Job/goal/exploration |
| B5 | “What are Ontario homeowners changing as heating incentives and gas costs evolve?” | Story/discovery; must refresh with current public evidence |

This example shows why proximity is not a funnel proxy and why locale cannot be a translation tag alone.

## 10. Existing Newsjack audit

The repository has no AEO/GEO prompt-panel skill or CLI command. It does have composable patterns that should be reused.

| Existing primitive | What to reuse | What not to make it own |
| --- | --- | --- |
| [`pr-strategist`](../skills/pr-strategist/SKILL.md) | Audience-before-outlet discipline, positioning evidence, and refusal of vanity metrics | Detailed ICP/job extraction or prompt sampling |
| [`newsjack-monitor-setup`](../skills/newsjack-monitor-setup/SKILL.md) | Explicit user-owned company profile; topics, competitors, search terms, standing, exclusions | AEO prompt panel. Monitor topics/search terms are discovery aperture, not buyer-query samples |
| [`newsjack-detector`](../skills/newsjack-detector/SKILL.md) | Canonical engine/skill boundary, deterministic evidence normalization, staged judgment, freshness, provenance, and seen-state | AEO report wording or prompt-selection judgment in Go |
| [`relevance-coarse-filter`](../skills/relevance-coarse-filter/SKILL.md) | Cheap high-recall pass before expensive semantic review | Final prompt selection, weighting, or newsworthiness |
| [`newsjack-triage`](../skills/newsjack-triage/SKILL.md) | Bounded responsibility, explicit non-responsibilities, human-readable result plus machine handoff | Prompt generation or visibility scoring |
| [`story-origin-check`](../skills/story-origin-check/SKILL.md) | Evidence ledger, first-source reasoning, confidence, and machine contract | Buyer intent or prompt freshness |
| [`fact-check`](../skills/fact-check/SKILL.md) | Claim extraction, source-tier climbing, citations, calibrated confidence | ICP desirability or prompt realism |
| [`coverage-tracker-setup`](../skills/coverage-tracker-setup/SKILL.md) and [`coverage-tracker`](../skills/coverage-tracker/SKILL.md) | Explicit semantic meaning per tracked term, validation, recurrence, local seen state, suppression hygiene | Reusing keyword trackers as prompt panels |
| [`voice-extractor`](../skills/voice-extractor/SKILL.md) | Versioned local artifact, deterministic measurements, check/enforce modes, fixture-calibrated bands | Applying a company’s writing voice to buyer prompts |
| [`angle-generator`](../skills/angle-generator/SKILL.md), [`news-search`](../skills/news-search/SKILL.md), and [`pr-calendar`](../skills/pr-calendar/SKILL.md) | Fresh story evidence for B5 discovery and campaign context | Evergreen buyer-job derivation |
| CLI [`cluster`](../apps/cli/cmd/newsjack/cluster.go), [`filter`](../apps/cli/cmd/newsjack/filter.go), [`origin`](../apps/cli/cmd/newsjack/origin.go), [`run-summary`](../apps/cli/cmd/newsjack/run_summary.go), profile, and stores | Deterministic JSON transforms, normalized IDs, candidate-pair generation, decision application, provenance, run manifests, limited inspectable state | Prompt-shaped taxonomies, PR judgment, human report prose, or inferred durable buyer topics |

The current monitor profile fields—company, description, topics, competitors, search terms, feeds, spokespeople, standing, and exclusions—can seed the company-evidence stage. They must not be copied directly into prompt weights. This preserves the repository’s discovery doctrine: setup owns explicit discovery aperture, runtime evidence owns judgment, and model-inferred topics do not become hidden durable state.

The other current public skills—`crisis-holding`, `find-journalists`, `headline-generator`, `journalist-fit-check`, `meanest-editor`, `newsworthiness-check`, `press-clip`, `reactive-comment`, and `same-outlet-ranker`—were also reviewed. They are downstream crisis, editorial, media-selection, outreach, or presentation capabilities and should not become dependencies of prompt-panel derivation. This explicit exclusion prevents the new workflow from inheriting journalist or pitch heuristics that do not describe buyer-query prevalence.

### Gap summary

Newsjack needs new contracts for:

- source-bound ICP hypotheses;
- source-bound buyer jobs and language;
- canonical prompt intent cells and proximity;
- contamination and semantic QA decisions;
- sampling/weighting/uncertainty plans; and
- lane-specific run manifests and observations.

It does not need:

- a second fact checker;
- a second generic search skill;
- a durable inferred-topic database;
- an AEO prose renderer in the CLI; or
- vendor-specific metric names in core schemas.

## 11. Proposed architecture

```text
company/public evidence ──> icp-evidence-analysis ──> human Gate 1
customer/market language ─> buyer-job-intent-analysis ─> human Gate 2
                                      │
                                      v
                         prompt-proximity-architecture
                                      │
                                      v
                           realistic-prompt-generation
                                      │
                   contamination register + evidence ledger
                                      │
                                      v
                               prompt-set-qa
                                      │
                 deterministic normalize/dedupe/validate
                                      │
                                      v
                              human panel Gate 3
                                      │
                                      v
                          ai-visibility-panel-design
                                      │
                                      v
                  versioned panel.yaml + immutable run manifest
                       /              |              \
              closed-model      retrieval/UI      campaign experiment
                       \              |              /
                      deterministic observations + uncertainty
```

The orchestration skill owns the sequence and handoffs. Each atom can run independently on an existing compatible artifact.

## 12. Proposed atomic skills and contracts

All contracts should include:

```json
{
  "schema_version": "1.0.0",
  "artifact_id": "stable-id",
  "created_at": "RFC3339",
  "created_by": "human-or-agent",
  "source_manifest_hash": "sha256:...",
  "warnings": []
}
```

### 12.1 `icp-evidence-analysis`

| Field | Specification |
| --- | --- |
| Responsibility | Turn a source-bound company/market dossier into ranked ICP hypotheses, buying roles, triggers, constraints, disqualifiers, standing, counterevidence, and research gaps. |
| Non-responsibilities | Buyer-job decomposition, prompt generation, prompt weighting, market-size claims, or promoting unsupported personas as facts. |
| Inputs | Company evidence manifest; optional monitor profile; target markets; exclusions; fact-check results. |
| Outputs | `icp_hypotheses.json` plus readable evidence/gap summary. |
| Dependencies | Reuse `fact-check` when material company claims lack independent verification. May consume `newsjack-monitor-setup` profiles as unverified leads. |
| Evidence requirements | Every attribute links to source IDs and spans; confidence is calibrated; company assertions are distinguished from behavior and independent evidence; negative evidence is retained. |
| Failure modes | Brand-copy mirroring; invented demographics; persona storytelling; confusing a current customer with the target population; ignoring buying committees; leaking private material. |
| Eval fixtures | Synthetic B2B committee with conflicting evidence; consumer/local service; multi-product company; unsupported website-only ICP; regulated exclusion; evidence that contradicts positioning. |

Suggested core contract:

```json
{
  "icps": [{
    "icp_id": "icp-finance-prof-services-50-500",
    "label": "Finance teams at 50–500 person professional-services firms",
    "context": {},
    "triggers": [{"text": "close delays from missing receipts", "source_ids": ["call-01"]}],
    "roles": [{"role": "champion", "label": "controller", "source_ids": ["call-01"]}],
    "constraints": [],
    "disqualifiers": [],
    "standing_claim_ids": ["claim-integration-02"],
    "supporting_source_ids": ["call-01", "search-08"],
    "counterevidence_source_ids": [],
    "confidence": "medium",
    "open_questions": []
  }]
}
```

### 12.2 `buyer-job-intent-analysis`

| Field | Specification |
| --- | --- |
| Responsibility | Extract jobs, struggling moments, desired progress, forces, workarounds, information acts, journey states, decision criteria, constraints, exact language, persona/locale context, and evidence grade. |
| Non-responsibilities | Deciding product fit, producing prompts, assigning funnel stage from keywords alone, or inventing prevalence. |
| Inputs | Approved ICP hypotheses; transcript/query/review/public-market source manifest; permitted-data rules. |
| Outputs | `buyer_jobs.json`, verbatim language bank, evidence gaps, and conflicts. |
| Dependencies | `fact-check` for external factual claims; `news-search` only when current market/story evidence is needed. |
| Evidence requirements | Every job has at least one source or is explicitly a grade-D hypothesis; excerpts retain context; frequency is a count within the supplied corpus, not market prevalence. |
| Failure modes | Feature-to-job inversion; treating internal copy as buyer voice; overgeneralizing one quote; losing negative cases; collapsing user, buyer, and approver; forcing TOFU/MOFU/BOFU. |
| Eval fixtures | Calls plus search terms that agree; sources that conflict; post-purchase job; ambiguous two-concept query; multilingual excerpt; only grade-D hypotheses. |

Suggested core contract:

```json
{
  "jobs": [{
    "job_id": "job-close-expenses-five-days",
    "icp_ids": ["icp-finance-prof-services-50-500"],
    "struggling_moment": "late receipts delay close",
    "desired_progress": "close expenses within five working days",
    "workarounds": ["spreadsheet reminders"],
    "forces": {"push": [], "pull": [], "anxiety": [], "habit": []},
    "information_acts": ["diagnose", "compare", "implement"],
    "journey_states": ["problem_identification", "requirements_building"],
    "criteria": ["policy enforcement", "NetSuite compatibility"],
    "language_samples": [{
      "text": "I spend the last week chasing receipts",
      "source_id": "call-01",
      "locale": "en-CA"
    }],
    "evidence_grade": "A",
    "confidence": "high"
  }]
}
```

### 12.3 `prompt-proximity-architecture`

| Field | Specification |
| --- | --- |
| Responsibility | Turn approved jobs into a coverage blueprint across proximity, aided status, information act, journey state, funnel rollup, persona, locale, evidence grade, and lane. Identify missing/overfull cells without writing prompts. |
| Non-responsibilities | Prompt wording, product standing, weight estimation, deduplication, or visibility measurement. |
| Inputs | Measurement charter; approved ICP/job artifacts; run budget; required locales/surfaces. |
| Outputs | `prompt_architecture.json` containing required, optional, and prohibited cells and allocation targets. |
| Dependencies | None beyond upstream artifacts. Uses the band definitions in this document. |
| Evidence requirements | Every required cell points to a job and reason; B5 cells need fresh public/story evidence; funnel may be null; aided and unaided allocations remain separate. |
| Failure modes | Full-factorial explosion; funnel-as-schema; one band dominating; requiring a recommendation act unsupported by evidence; treating each locale as translation; budget exceeding limits. |
| Eval fixtures | Same job across multiple bands; direct-brand post-purchase; urgent problem near purchase; sparse locale; campaign partition; allocation-budget conflict. |

Suggested core contract:

```json
{
  "cells": [{
    "cell_spec_id": "spec-001",
    "job_id": "job-close-expenses-five-days",
    "proximity_band": "B3_problem_need",
    "aided_status": "unaided",
    "information_act": "diagnose",
    "journey_state": "problem_identification",
    "funnel": "TOFU",
    "persona_id": "role-controller",
    "locale": "en-CA",
    "lane_eligibility": ["closed_model", "retrieval"],
    "target_variants": 2,
    "required": true,
    "reason_source_ids": ["call-01"]
  }],
  "allocation": {"core_cells": 72, "rotating_cells": 18, "aided_cells": 12}
}
```

### 12.4 `realistic-prompt-generation`

| Field | Specification |
| --- | --- |
| Responsibility | Create natural prompt candidates and controlled variants from the blind design brief, preserving the approved job, information act, constraints, persona, locale, and evidence language. |
| Non-responsibilities | Seeing target/campaign terms for unaided generation; selecting the final panel; asserting real-world frequency; changing architecture; evaluating brand visibility. |
| Inputs | Blind job/language brief; prompt architecture; style/locale requirements; source IDs without target-bearing content. |
| Outputs | `prompt_universe.json` with canonical-cell candidates, transformation type, source support, and generator provenance. |
| Dependencies | May use public conversational corpora only as style-validation references, not weights. Native/market-competent human review for core non-default locales. |
| Evidence requirements | Every candidate cites the job and evidence; synthetic expansions are grade D until validated; transformations are explicit; variants preserve the same intent cell. |
| Failure modes | Brand/category leakage; polished marketing questions; forcing products into problems; implausible persona exposition; copying public corpus text; overlong prompts; false localization; answer leakage. |
| Eval fixtures | Verbatim-to-light-normalization; concise and contextual variants; category forbidden in B4; slogan trap; competitor trap; locale transcreation; multi-turn script; deliberately awkward source language. |

Suggested core contract:

```json
{
  "canonical_cells": [{
    "canonical_cell_id": "cell-job-close-b3-diagnose-en-ca-controller",
    "cell_spec_id": "spec-001",
    "expected_answer_kind": "diagnosis_and_options",
    "candidates": [{
      "candidate_id": "prompt-001a",
      "text": "How can finance stop chasing consultants for receipts at month end?",
      "language": "en",
      "locale": "en-CA",
      "transformation": "lightly_normalized",
      "source_ids": ["call-01"],
      "evidence_grade": "A",
      "generation_provenance": {"model": "declared-model", "prompt_hash": "sha256:..."}
    }]
  }]
}
```

### 12.5 `prompt-set-qa`

| Field | Specification |
| --- | --- |
| Responsibility | Gate schema, provenance, contamination, evidence entailment, naturalness, one-concept clarity, architecture consistency, aided status, answer leakage, and semantic duplicate decisions. Return pass, revise, quarantine, or reject with reasons. |
| Non-responsibilities | Quietly rewriting prompts, inventing evidence, changing ICP/jobs, assigning business weights, or selecting based on baseline performance. |
| Inputs | Prompt universe; architecture; contamination register; evidence excerpts; deterministic scan/candidate-pair output; optional blind human decisions. |
| Outputs | `prompt_qa.json`, accepted candidate IDs, revisions requested, quarantine/rejection ledger, canonical merge/split decisions. |
| Dependencies | Deterministic schema/lexicon/hash/similarity scripts first; LLM semantic review second; human gate for core, high-weight, locale, and disputed decisions. |
| Evidence requirements | Each decision cites rule IDs and evidence; baseline response/performance fields are forbidden inputs; reviewers are told which fields were blinded. |
| Failure modes | QA “repair” that changes intent; embedding auto-deletion; false slogan positives; leniency toward flattering prompts; inconsistent merge decisions; selectors seeing results; locale review by unqualified model only. |
| Eval fixtures | Gold set of leaked aliases, semantic slogan leakage, legitimate shared word, recommendation forcing, answer-derived prompt, exact duplicate, paraphrase duplicate, near-but-materially-different constraint, aided mislabel, adversarial flattering prompt. |

Suggested core contract:

```json
{
  "decisions": [{
    "candidate_id": "prompt-001a",
    "status": "pass",
    "rule_results": [
      {"rule_id": "no-target-term-unaided", "status": "pass", "evidence": []},
      {"rule_id": "job-entailment", "status": "pass", "evidence": ["call-01"]}
    ],
    "duplicate_decision": {
      "canonical_cell_id": "cell-job-close-b3-diagnose-en-ca-controller",
      "action": "retain_variant"
    },
    "review_confidence": "high"
  }]
}
```

### 12.6 `ai-visibility-panel-design`

| Field | Specification |
| --- | --- |
| Responsibility | Select accepted canonical cells into panel partitions; set allocation, variants, surfaces, locales, lanes, repetitions, weights, randomization, uncertainty method, refresh policy, versioning, and campaign controls. |
| Non-responsibilities | Generating prompts, estimating unsupported population frequency, running models, interpreting PR angles, or claiming campaign causality without a design. |
| Inputs | Charter; architecture; QA-approved universe; evidence-backed weight inputs; variance-pilot observations; run budget; prior panel version. |
| Outputs | Human tracking plan plus `panel.yaml` and deterministic validation requirements. |
| Dependencies | Deterministic weight/coverage/effective-N calculator, variance decomposition, Wilson interval, stratified cluster bootstrap, schema validator, and version diff. |
| Evidence requirements | Every weight has provenance/confidence; allocation meets declared minima or emits a waiver; interval limits are stated; campaign design identifies treatment/control and pre-registration; core changes have reasons. |
| Failure modes | Repetitions counted as independent cells; invented precision; full-factorial budget; mixing lanes/aided status; performance-based selection; weights not summing to one; panel refresh overwriting history; causal language from before/after. |
| Eval fixtures | Equal-weight sparse panel; evidence-weighted panel; dominant weight/effective-N warning; high within-cell variance; high between-cell variance; panel-version overlap; missing subgroup cells; campaign with contaminated controls; parser failure rates. |

Suggested core contract:

```yaml
schema_version: 1.0.0
panel_id: panel-ledgerlift-en-ca
version: 1.0.0
charter_id: charter-001
estimands:
  - weighted_unaided_presence
partitions:
  core:
    canonical_cell_ids: [cell-job-close-b3-diagnose-en-ca-controller]
  rotating:
    canonical_cell_ids: []
  aided:
    canonical_cell_ids: []
lanes:
  closed_model:
    surfaces:
      - provider: example
        model: fixed-model-id
        search_policy: disabled
    repetitions_per_wave: 3
statistics:
  binary_strata_interval: wilson
  aggregate_interval: stratified_cluster_bootstrap
  cluster_unit: canonical_cell_id
  bootstrap_replicates: 2000
refresh:
  evidence_intake: monthly
  panel_review: quarterly
  core_target_share: 0.75
limitations:
  - "Non-probability prompt panel; estimates are conditional on declared coverage."
```

## 13. Orchestration workflow

The molecule should be named `build-ai-visibility-panel`.

### Ordered flow

1. Read repository/user context and measurement charter.
2. Build or ingest the source manifest.
3. Run `icp-evidence-analysis`.
4. **Human Gate 1:** confirm factual perimeter, ICPs, exclusions, and permitted sources.
5. Run `buyer-job-intent-analysis`.
6. **Human Gate 2:** confirm buyer jobs, language, locales, and strategic priority.
7. Create the contamination register and blind brief.
8. Run `prompt-proximity-architecture`.
9. Run `realistic-prompt-generation`.
10. Run deterministic schema, lexicon, hash, and similarity checks.
11. Run `prompt-set-qa`.
12. **Human Gate 3:** approve core/aided/campaign partitions and disputed QA decisions while still blind to baseline visibility.
13. Run a variance pilot.
14. Run `ai-visibility-panel-design`.
15. **Human Gate 4:** approve weights, limitations, cadence, campaign claims, and frozen version.
16. Emit the panel, run manifest template, readable methodology, and next-review date.

### Deterministic scripts

The following operations are universal, mechanical, and testable enough for scripts or narrowly scoped CLI JSON transforms:

- schema validation and migrations;
- source/hash/provenance completeness;
- Unicode/exact normalization;
- target/campaign lexical and fuzzy scan;
- exact duplicate hashes;
- lexical/embedding candidate-pair generation;
- coverage-grid and allocation-budget checks;
- weight normalization, dominance, and effective sample size;
- seeded prompt order and run manifests;
- response/citation payload hashing;
- audited alias/domain detection;
- run validity and retry policy;
- Wilson intervals, variance decomposition, paired/cluster bootstrap;
- panel-version diff and overlap bridge; and
- inspectable local recurrence/seen state with a clear lifecycle.

Embedding candidate generation can be deterministic for fixed model/version/input, but the merge decision remains judgment.

### LLM judgment

LLMs may:

- interpret mixed evidence into explicit ICP hypotheses;
- extract jobs and buyer language with cited spans;
- propose architecture cells;
- create blinded natural variants;
- assess semantic contamination, naturalness, entailment, and same-intent duplicates;
- explain evidence gaps and failure modes; and
- render human-readable reports from validated artifacts.

LLMs may not silently persist inferred topics, weights, personas, or preferences. A human must promote them into a versioned user-owned artifact.

## 14. Implementation sequence and acceptance criteria

### Slice 0: contracts, ethics, and fixtures

Deliver:

- measurement charter schema;
- shared source/evidence manifest;
- ICP, job, prompt-universe, QA, and panel schemas;
- contamination taxonomy;
- synthetic LedgerLift and HarborHeat fixtures;
- gold QA/dedup adversarial fixtures;
- artifact lifecycle and privacy notes.

Acceptance:

- every example artifact validates;
- every prompt traces to evidence and a canonical cell;
- no real/private client identifiers appear;
- aided, competitor-aided, unaided, retrieval, and campaign states cannot be conflated;
- schema versions and IDs are stable;
- machine artifacts are secondary to readable Markdown outputs.

### Slice 1: deterministic validator and QA substrate

Deliver:

- schema/provenance validator;
- contamination lexicon scanner;
- normalization and exact hashes;
- similarity candidate generator;
- coverage/weight/effective-N checks;
- panel diff/overlap report.

Acceptance:

- deterministic output is byte-stable for fixed inputs and tool versions;
- gold lexical leaks have 100% recall;
- legitimate allowed exceptions pass;
- embedding similarity never auto-deletes;
- invalid weights, missing provenance, repeated IDs, lane mixing, and campaign/core leakage fail clearly;
- unrelated repository behavior is unchanged.

### Slice 2: ICP and buyer-job atoms

Deliver `icp-evidence-analysis` and `buyer-job-intent-analysis`.

Acceptance:

- unsupported fields are null/hypothesis, never invented;
- every material conclusion has source IDs/spans and calibrated confidence;
- company assertion, behavior, and independent evidence remain distinguishable;
- negative evidence and conflicting sources survive;
- role, job, journey, and funnel are not collapsed;
- privacy/permission failures stop the workflow.

### Slice 3: proximity and realistic-generation atoms

Deliver `prompt-proximity-architecture` and `realistic-prompt-generation`.

Acceptance:

- fixtures cover B0–B5 without mechanically assigning funnel;
- architecture stays inside the run budget;
- the generator cannot access target/campaign lexicons in unaided mode;
- prompt variants preserve canonical-cell meaning;
- grade-D generations cannot enter core without explicit validation/promotion;
- locale fixtures require transcreation review;
- public conversational corpora affect style tests, not weights.

### Slice 4: semantic QA atom and human gate

Deliver `prompt-set-qa`.

Acceptance:

- gold same-intent and materially-different pairs meet fixture-set precision/recall thresholds chosen before tuning;
- all rule decisions are explainable and evidence-bound;
- QA never rewrites upstream facts;
- selectors cannot see baseline visibility fields;
- disputed or low-confidence core decisions require human resolution;
- rejected and quarantined prompts remain inspectable.

### Slice 5: panel design and statistics

Deliver `ai-visibility-panel-design` and deterministic statistical helpers.

Acceptance:

- repeat observations are nested under canonical cells;
- variance pilots alter allocation in the expected direction on synthetic data;
- Wilson and cluster-bootstrap calculations match trusted reference fixtures;
- dominant weights emit effective-N warnings;
- reports include conditional-panel language and complete denominators;
- panel version changes produce overlap-only and full-version comparisons;
- causal wording is blocked for unregistered before/after campaign designs.

### Slice 6: orchestration

Deliver `build-ai-visibility-panel`.

Acceptance:

- all four human gates are resumable from artifacts;
- each atom can be rerun independently;
- unchanged inputs do not create hidden durable guesses;
- the workflow produces readable Markdown plus validated JSON/YAML;
- failures route to the owning atom;
- no model runner or vendor integration is required to build a valid panel.

### Slice 7: lane-specific runners and longitudinal tracking

Only after the panel methodology passes fixtures:

- implement closed-model, retrieval, and permitted consumer-surface adapters;
- emit immutable run manifests and observations;
- add recurrence/seen state with inspectable lifecycle;
- add campaign experiment registry and outcome ladder;
- keep human reports in skills, not hard-coded CLI prose.

Acceptance:

- search allowed/required/used is observable;
- API and UI results never silently merge;
- model/surface/config drift is surfaced;
- responses and citations are reproducible to the extent provider interfaces allow;
- retries and invalid runs do not inflate denominators;
- campaign panels cannot overwrite evergreen history.

## 15. Risks and open research

1. **Coverage remains the hardest problem.** Even licensed conversation panels have unknown or incomplete selection mechanisms. Newsjack should expose source mixtures and panel limits rather than manufacture a universal prompt-volume number.
2. **Provider behavior changes.** Hidden routing, personalization, safety layers, and model updates can break a time series. Sentinels and run manifests reveal some drift, not all.
3. **Semantic dedup is value-laden.** Two prompts can look similar while changing urgency, constraint, or expected answer. Gold fixtures must include these boundary cases.
4. **Weights can disguise strategy as demand.** The exposure/priority split and provenance are mandatory.
5. **Multi-turn behavior is underrepresented.** Public corpora establish that multi-turn use is material, but company-specific journey scripts require new evidence and separate methods.
6. **Citation is not consumption.** A cited source, bot visit, human referral, remembered brand, and purchase are different outcomes.
7. **Campaign spillover complicates controls.** Public content can enter all retrieval conditions. Geo/time/content holdouts need feasibility and contamination checks.
8. **The field lacks shared benchmarks.** Newsjack should publish synthetic fixtures and metric definitions first, then consider a privacy-safe public benchmark of panel design and output variance.

## 16. Research conclusion

Competent teams derive useful prompt sets by combining customer language, search behavior, market evidence, company standing, structured buyer research, and controlled synthetic expansion. The leading platforms have converged on repeated cross-model monitoring, prompt taxonomies, tagging, and a mix of observed, proxy, and generated discovery. They have not converged on a transparent sampling frame, a definition of “share of model,” a defensible minimum panel size, or campaign causality.

Newsjack’s opportunity is not to claim a better magic score. It is to make the measurement design inspectable:

- evidence before generation;
- jobs and information needs before funnel labels;
- proximity before “high intent” shortcuts;
- unaided before aided;
- panel cells before prompt strings;
- unique-cell coverage before brute-force repetition;
- retrieval state before citation comparison;
- frozen controls before campaign claims;
- conditional uncertainty before leaderboards; and
- versioned user-owned artifacts before hidden memory.

That process composes with Newsjack’s existing evidence-first, strict-surfacing, deterministic-data-layer doctrine and creates small, testable implementation slices without turning the CLI into a marketing-opinion engine.
