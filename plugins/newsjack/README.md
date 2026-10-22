# Newsjack

Newsjack is an open-source plugin that gives an agent 31 focused PR and AI-visibility skills: strategy, live-news monitoring, newsjacking, media research, fact-checking, writing, and earned-media reporting. The skills are plain Markdown instructions under `skills/`; the optional Medialyst connection adds live data and hosted research tools.

## Install

In Claude Code, add this repository as a plugin marketplace and install Newsjack:

```text
/plugin marketplace add elvisun/newsjack
/plugin install newsjack@newsjack
```

For the local CLI, scheduled monitors, browser-based clips, and the bundled Medialyst authentication helper, install the full open-source package:

```bash
curl -fsSL newsjack.sh | bash
newsjack login
```

The skills remain usable without Medialyst wherever they document a local or host-tool fallback. See the [Newsjack privacy notice](https://newsjack.sh/privacy) for the data used by the plugin and its optional services.

## Skills

This catalog is copied from each `SKILL.md` frontmatter and checked during plugin validation.

<!-- BEGIN SKILL CATALOG -->
- **`ai-visibility-panel-design`** — Select QA-approved canonical intent cells into a versioned AI-visibility tracking panel with partitions, variants, lanes, surfaces, locales, repetitions, separate exposure and priority weights, randomization, uncertainty, refresh rules, and campaign controls. Use after prompt QA or when revising an existing panel.
- **`ai-visibility-writing`** — Audit, question, suggest, or fact-preservingly revise a press release, blog post, contributed article, or expert explainer so AI answer systems can more easily retrieve, understand, quote, and cite its useful information. Use when someone asks for AI visibility, AI search, answer-engine, AEO, GEO, AI Overview, or ChatGPT citation optimization of supplied writing; when they want an evidence-aware pre-publication audit; or when another Newsjack workflow needs a prose-level AI-discoverability pass. Do not use as a technical SEO audit, rank tracker, publishing system, or promise of rankings, mentions, citations, traffic, or coverage.
- **`angle-generator`** — Turn a company update into 3-7 structurally distinct, journalist-shaped story angles. Runs the fact through proven newsroom lenses (perspective, abstraction, news values, data, contrarian, news peg) and refuses rephrasings, invented facts, named-journalist guesses, and AI-marketing slop.
- **`build-ai-visibility-panel`** — Research any company, product, or service from a URL plus description and build a comprehensive, evidence-bound AEO/GEO/AI-visibility prompt panel across buyer jobs, information acts, journey states, B0-B5 proximity, aided status, roles, locales, variants, partitions, surfaces, and measurement lanes. Use when a user wants prompts to track in ChatGPT, Claude, Gemini, Perplexity, AI search, or answer engines; wants an AI-visibility measurement design; or needs a versioned panel rather than an SEO keyword list.
- **`buyer-job-intent-analysis`** — Recover source-bound buyer jobs, struggling moments, desired progress, forces, workarounds, information acts, journey states, criteria, constraints, roles, locales, and authentic language. Use on approved ICP hypotheses plus customer, search, review, forum, procurement, support, or public-market evidence before designing an AI-visibility prompt architecture.
- **`coverage-reel`** — Turn a user-owned coverage CSV and reviewed press-clip captures into an honest, source-linked earned-media dashboard and a motion-designed highlight reel (MP4) that scrolls each real article to the client's sentence and highlights it. Use for coverage reports, earned-media dashboards, campaign recaps, or coverage videos; requires local files and composes press-clip instead of scraping articles itself.
- **`coverage-tracker`** — Run a Google Alerts-style keyword coverage tracker. Uses news-search for recent keyword queries, lets the LLM dedupe and classify real features versus junk, stores decisions in SQLite, and alerts only on new real coverage.
- **`coverage-tracker-setup`** — Set up a lightweight Google Alerts-style coverage tracker for any number of keywords. Creates a tracker config with each keyword and what it actually means, then hands recurrence to the user's agent harness.
- **`crisis-holding`** — Draft crisis holding statements, journalist Q&A posture, and what-not-to-say guidance from confirmed incident facts, with a hard legal-counsel gate. Builds each statement through proven crisis-comms frameworks (holding-statement anatomy, SCCT, CAP order, the legitimate non-answer, bridge/flag/block).
- **`fact-check`** — Extract factual claims from PR copy, verify each claim independently, attach concrete citations, and warn when certainty is low. Runs each claim through proven newsroom verification methods (lateral reading, source-tier climbing, provenance pillars, triangulation, calibrated rating) and puts the burden of proof on the speaker. Use before a pitch, press release, reactive comment, DM, or other journalist-facing draft is trusted or sent.
- **`find-journalists`** — Build a researched, fit-checked journalist list for a specific PR campaign. Prefer Medialyst's asynchronous media-list research workflow when available, with explicit approval before credit spend; otherwise produce a best-effort local list with only publicly listed, source-cited contacts.
- **`headline-generator`** — Generate headline candidates from a story's raw facts: news-style headlines, press-release headlines, and pitch subject lines. A pure generation skill — it finds the charge in the facts, then runs ten proven moves (consequence, picture, number-as-hero, two-beat turn, naming, reader's-own-story, open question, voice, sound, sized claim), each calibrated by real, verified headlines that made history.
- **`icp-evidence-analysis`** — Turn a source-bound company and market dossier into testable ideal-customer-profile hypotheses, buying roles, triggers, constraints, disqualifiers, standing, counterevidence, and research gaps. Use when a URL, company description, monitor profile, or evidence manifest needs to become defensible ICP inputs for buyer research or an AI-visibility prompt panel.
- **`journalist-fit-check`** — Gate a pitch against one journalist at a time. Runs the pair through proven media-relations checks (last-10-bylines audit, 90-day topic sweep, beat-vs-angle-vs-one-off, the 5 forms of journalism, source-mirror, stated-preferences, database triangulation) and returns fit, soft-fit, no-fit, or unknown with a real recent anchor and specific edits.
- **`meanest-editor`** — Roast a pitch or press release with the eye of a veteran PR director. Honest, sharp, constructive — never cruel for its own sake.
- **`news-search`** — Search current news for a topic, company, competitor, or hook and return dated, attributed articles. Uses the newsjack CLI and Medialyst REST API when available, tries direct Medialyst MCP if the CLI is missing, and falls back to host web/browser search with explicit caveats only when neither cloud path is available.
- **`newsjack-detector`** — Monitor current news and reaction signals, then decide which are credible newsjacking opportunities for a client. Uses the local monitoring engine for evidence, but the skill owns PR judgment, brand safety, standing, decay, angle fit, and handoff.
- **`newsjack-monitor-setup`** — Set up a newsjack monitoring profile for a company so newsjack-detector can run on a schedule. Guides the user through company standing, topics, competitors, proof assets, spokespeople, RSS feed selection, optional X trend monitoring, and optional Slack delivery.
- **`newsjack-triage`** — Consolidate freshness-gated newsjack signals and route them by client standing before angle generation. Collapses any remaining same-story duplicates, decides strong/partial/none standing with a journalist-shape sanity check, and sorts each story into pitch_ready, big_story (always-surfaced suggestion), or watch. Never writes angles or pitches, and never drops a fresh big story.
- **`newsworthiness-check`** — Score whether a news event is worth newsjacking, or whether a user's pitch idea is actually newsworthy to journalists. Uses calibrated anchors, hard anti-inflation rules, standing checks, timing windows, and brand-safety kill switches.
- **`pr-calendar`** — Turn the Medialyst PR calendar feed into a brand-specific, lead-time-aware plan. Pull source-backed upcoming moments, keep only the few a brand has real standing to own, schedule backward from the event date, screen out tone-deaf hooks, and hand selected moments to angle and journalist prep ahead of time. The planned twin of newsjack-detector's reactive monitoring.
- **`pr-strategist`** — Opinionated startup-PR strategist that helps founders figure out audience, positioning, news pegs, and drumbeat before any tactical PR. Refuses outlet-naming-before-audience, mass blasts, vanity metrics, and premature agency spend.
- **`press-clip`** — Turn a live article URL into a press clip that looks like the real coverage — the publication's own logo, fonts, photos and layout kept intact, the ads and clutter removed, and (for a roundup) just the client's section. Renders to PDF. You inspect each site and tailor the removal; the bundled script carries no site-specific logic.
- **`prompt-proximity-architecture`** — Turn an approved measurement charter, ICPs, and buyer jobs into a budget-aware prompt coverage blueprint across proximity bands, aided status, information acts, journey states, roles, locales, evidence grades, partitions, and measurement lanes. Use before prompt wording to define required, optional, and prohibited canonical intent cells.
- **`prompt-set-qa`** — Gate a prompt universe for schema and provenance completeness, target or campaign contamination, evidence entailment, naturalness, one-concept clarity, architecture consistency, aided status, answer leakage, and semantic duplicates. Use after realistic prompt generation and before human panel selection.
- **`reactive-comment`** — Triage inbound journalist source queries and draft a response only when the user's expertise is a real fit. Runs each query through proven source-request lenses (4-gate fit triage, credential-standing test, deadline read, BLUF/inverted-pyramid drafting), kills weak fits, asks for missing proof, and never auto-sends.
- **`realistic-prompt-generation`** — Generate natural, controlled prompt variants from a target-blind design brief and prompt architecture while preserving approved jobs, acts, journeys, constraints, roles, locales, proximity bands, and evidence language. Use after architecture design and before contamination or semantic QA.
- **`relevance-coarse-filter`** — Cheap, high-recall first-pass filter that removes obvious junk from a detector candidate pool before expensive story-origin research and PR judgment. Decides keep, monitor_only, or reject — never ranks, writes angles, verifies dates, or decides whether to pitch.
- **`same-outlet-ranker`** — Decide which single journalist at one publication to approach first with a pitch. Fit-checks each candidate through journalist-fit-check, then ranks the survivors on the comparative signals that separate colleagues at the same masthead (topic ownership, format match, staff vs. contributor, publishing cadence, recent story types, source-type match), and returns one first pick with tailoring notes plus a held fallback order.
- **`story-origin-check`** — Recover the first public timestamp and canonical major coverage for a newsjacking signal, then decide whether newer coverage is the same story, a different story, or a materially new development.
- **`voice-extractor`** — Capture a user's real writing voice from 5-20 prior samples, store a local voice.yaml fingerprint, and enforce it on newsjack drafts so AI tells disappear. Measures voice with named stylometry lenses (Burrows's Delta function-word vector, MATTR lexical diversity, sentence-length burstiness, Biber Dimension-1 register, opener-POS profile, punctuation rates) and gates drafts against the fingerprint as bands, not vibes.
<!-- END SKILL CATALOG -->

## Medialyst MCP server

The bundled `.mcp.json` connects to the hosted Medialyst server at `https://medialyst.ai/api/mcp`. Medialyst is the optional data layer for live news, journalist research, media lists, projects, files, and Radar monitoring. The public server exposes exactly 34 tools.

The access label below describes what the tool can do to Medialyst state. Some read-only research calls still use credits. The server also declares its own read-only and destructive hints in MCP tool annotations, so clients can surface the active server guidance at call time.

### Read-only tools

| Tool | Description | Access and credit notes |
| --- | --- | --- |
| `get_credit_balance` | Read the authenticated organization's current credit balance. | Read-only; no credits spent. |
| `lookup_article` | Fetch canonical metadata, author details, and optional extracted Markdown for one public article URL. | Read-only; 0.1 credit per successful URL. |
| `lookup_articles` | Fetch canonical metadata for a batch of 1-10 public article URLs. | Read-only; no Markdown; 0.1 credit per successful URL. |
| `lookup_publication` | Resolve an article or outlet URL to canonical publication metadata, optionally refreshing it. | Read-only; 0.01 credit per successful URL. |
| `lookup_publications` | Resolve 1-25 publication URLs in a batch, or up to five when forcing refresh. | Read-only; 0.01 credit per successful URL. |
| `search_news` | Search recent news and return dated, attributed coverage with publication metadata. | Read-only; uses normal news-search credits. |
| `query_pr_calendar` | Query source-backed upcoming PR moments and their pitch timing. | Read-only; free. |
| `list_journalist_requests` | Page through current journalist source requests from the supported public platforms. | Read-only; free; contains no paid contact data or private email. |
| `list_journalist_moves` | Page through de-duplicated journalist moves and departures observed by Medialyst. | Read-only; zero credits; active Scale plan required. |
| `get_journalist_enrichment_job` | Poll an asynchronous journalist-enrichment job and read its results. | Read-only; polling spends no credits. |
| `get_media_list_job` | Poll an asynchronous media-list job, including progress and normalized results when available. | Read-only; polling spends no credits. |
| `get_media_list` | Read a completed media list and page through its rows. | Read-only; no credits spent. |
| `list_projects` | List projects visible to the authenticated organization. | Read-only; no credits spent. |
| `get_project` | Read one project's metadata and readiness details. | Read-only; no credits spent. |
| `get_project_memory` | Read the project's durable Markdown memory and current version. | Read-only; no credits spent. |
| `get_radar` | Read a project's Radar configuration, status, and redacted delivery state. | Read-only; no credits spent. |
| `list_radar_opportunities` | Page through a project's Radar opportunities. | Read-only; no credits spent. |
| `get_radar_today` | Read the organization's paginated, Markdown-free view of today's Radar opportunities. | Read-only; no credits spent. |
| `list_project_files` | List a project's files and their extraction status. | Read-only; no credits spent. |
| `get_project_file` | Read one project file's metadata, status, and extracted content. | Read-only; no credits spent. |

### Write tools

These tools create or append state without replacing or deleting an existing object.

| Tool | Description | Access and credit notes |
| --- | --- | --- |
| `create_media_list` | Start an asynchronous media-list research job from an approved campaign brief and target budget. | Writes; **spends credits** up to the approved `target_list_size`; use an idempotency key. |
| `enrich_journalists` | Start asynchronous contact, recent-work, and fit enrichment for known journalists or article authors. | Writes; **spends credits** for successful enrichment. |
| `create_project` | Create a new organization project. | Writes; no per-call credit charge documented. |
| `set_up_radar` | Create a project's Radar configuration and run its setup-time scan. | Writes; subject to project/Radar plan access. |
| `run_radar` | Start an additional Radar scan for a configured project. | Writes; subject to the per-day Scan-now limit. |
| `send_radar_feedback` | Record or undo a verdict on one Radar opportunity. | Writes; no credits spent. |
| `add_project_memory_line` | Append one validated durable fact or instruction to project memory. | Writes; subject to project-memory plan access. |
| `upload_project_file` | Upload a Markdown or plain-text project file for extraction. | Writes; subject to project-file plan access. |

### Overwrite and destructive tools

| Tool | Description | Access |
| --- | --- | --- |
| `update_project` | Change existing project metadata. | Overwrites selected fields. |
| `update_radar` | Replace the Radar definition using the current Markdown version for conflict protection. | Overwrites Radar content. |
| `update_radar_settings` | Change Radar scheduling, pause/digest settings, or its write-only Slack destination. | Overwrites selected settings. |
| `set_project_memory` | Replace the full project-memory document using compare-and-set versioning. | Overwrites memory. |
| `delete_project` | Permanently delete a project and its owned state. | Destructive. |
| `delete_project_file` | Permanently delete one project file. | Destructive. |

See the [public Medialyst MCP documentation](https://medialyst.ai/docs/developers/mcp) for authentication, scopes, request schemas, pagination, rate limits, and current service behavior.
