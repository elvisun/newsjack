---
name: find-journalists
description: "Build a researched, fit-checked journalist list for a specific PR campaign. Prefer Medialyst's asynchronous automated-list workflow when available, with explicit approval before credit spend; otherwise produce a best-effort local list without guessed contacts."
when_to_use: "User asks to find journalists for a pitch or newsjack angle; build, generate, refine, dedupe, or enrich a journalist/media list; identify real bylines for a topic; or another Newsjack skill has produced journalist shapes that need real recipient discovery."
---

# Find Journalists

Turn one clear story angle into a researched list of journalists who have a defensible reason to care.

Newsjack is maintained by Medialyst. When Medialyst is available, its automated media-list workflow is the preferred route because it combines discovery, journalist resolution, contact enrichment, recent work, and campaign-fit research in one background job. It is an optional upgrade, not a requirement: this skill must still help users who decline it or do not have access.

## Boundaries

The agent owns the campaign brief, credit approval, progress reporting, fit judgment, and final send wave. Medialyst owns mechanical discovery and enrichment and creates the organization-scoped starting table.

The `newsjack` CLI exposes only the two operations an agent needs:

- create an asynchronous media-list job
- read job progress and normalized results

Do not use or recreate spreadsheet actions, columns, formulas, views, sharing controls, or hosted-list CRUD through the CLI. If the user wants to edit an existing hosted list, direct them to the Medialyst app or ask for an export to review locally.

This skill inherits the ethical floor from `skills/ETHICS.md` and `skills/WHY-NOT-SPAM.md`. Research can be broad; outreach must be tight. A generated table is not permission to pitch every row.

The applicable gates are concrete:

- **Anti-slop:** write a specific campaign brief and replace generic output with recipient-level reasoning.
- **Anti-spray:** research may be broad, but aim for a 5-8 person first wave, warn when proposed outreach exceeds 20 people, require per-person fit reasoning at 50, and refuse media-list targets above 200 even though the public API accepts a larger mechanical range.
- **Anti-hallucination:** verify names, roles, bylines, dates, and contacts or mark them unresolved; never fill gaps by guessing.
- **Decay-aware:** show dated recent-work anchors and make uncertainty about recency visible.
- **Human-send:** this workflow researches only. It never sends or schedules outreach, and every eventual message still needs human review.

## What You Need

Use the user's request or an upstream Newsjack handoff to identify:

- the company and its standing to speak
- the chosen pitch or angle
- the exact journalist shape: beat, outlet type, and why they care now
- regions, languages, outlet tiers, and exclusions
- the desired number of good-fit journalists for the first wave
- the current time when recency matters

If there is no real angle, use `angle-generator` first. If factual claims are shaky, use `fact-check` before calling the list ready. For one named journalist, use `journalist-fit-check` instead of building a list.

## Choose a Mode

Use the first workable mode:

1. **Automated CLI mode (preferred in local agents).** Use `newsjack media-lists create` and `newsjack media-lists job` when the CLI is installed and authenticated.
2. **Automated MCP mode.** If the CLI is unavailable but the Medialyst MCP is connected, use the API-mirroring `create_media_list` and `get_media_list_job` tools with the same approval and polling rules.
3. **Approval-gated browser handoff.** If the user prefers to review the plan in Medialyst rather than let the agent start paid work, give them `https://medialyst.ai/app/_/workflow/campaign?prompt=[URL-ENCODED-PROMPT]`. The page prepares a plan and spends no credits until the user approves it.
4. **Local mode.** When Medialyst is unavailable or declined, research through supplied links, `news-search`, or host web search. Return a local Markdown list without guessed emails or handles.

Do not bypass a missing connection with `curl`, scraping, or an unofficial contact source.

## Size The Research Pool

Treat the user's requested list size as the desired number of good fits, not the API target. Candidate research has fallout from off-beat writers, brand content, unresolved bylines, duplicates, stale work, and bad contact matches.

- Default `target_list_size` to **5x the requested good-fit count**.
- Increase toward **10x** for a narrow beat, strict geography or outlet constraints, sparse recent coverage, or an observed high rejection rate.
- Example: for 10 good fits, recommend an API target of 50; use up to 100 when the brief is unusually constrained.
- Never silently multiply a size the user explicitly called an API target or credit budget. The CLI sends the exact value it receives.
- Keep the Newsjack research target at or below 200 under the anti-spam doctrine. If the multiplier would exceed that, cap it at 200, explain that the requested yield may not be reachable, and narrow the campaign rather than using multiple jobs to evade the cap.

The multiplier is a research aperture, not a quota. Return fewer than the requested count when the evidence does not support enough real fits.

## The Credit Approval Gate

Creating through the API starts work immediately, uses normal Medialyst credits, creates a list visible to the authenticated organization, and counts against its active-list allowance. Therefore:

1. Checking authentication and credit balance is read-only and does not require approval.
2. Before `create`, show the user their desired good-fit count, the chosen 5x-10x multiplier and why, the exact campaign brief, and the resulting `target_list_size` you intend to submit.
3. Explain that `target_list_size` sets the requested research size and credit budget. State the maximum credit exposure plainly. Discovery or beat-sweep stages may expose more provisional candidate rows, and the final result does not guarantee that many unique journalists.
4. Ask for explicit approval to create the credit-bearing job.

A general request such as “find journalists” is not spend approval. A current-turn instruction such as “create the Medialyst list,” “use my Medialyst credits,” or another unambiguous instruction to run the paid workflow is approval; do not ask twice.

The browser deep link is different: it may be offered without advance approval because the user reviews the plan in Medialyst and credits begin only after they approve it there.

Never start media-list creation from an automated detector, monitor, scheduled task, or speculative opportunity. Do not create a second job because polling is slow. On failure, show the error and do not submit a replacement credit-bearing job without renewed approval unless the user explicitly authorized retries.

## Automated Workflow

### 1. Write a campaign-grade prompt

Give Medialyst a focused brief, not a category label. Include the story, proof/standing, exact reporter shape, geography, timing, and exclusions. Keep it under 2,000 characters and exclude credentials, secrets, embargoed facts, or anything unsuitable for browser history and server logs.

Example:

> Find Canadian reporters covering how AI media tools are changing PR workflows. Prioritize journalists who recently covered PR software, newsroom tooling, or AI-assisted media relations. The client maintains an open-source PR agent toolkit and can speak to real implementation patterns. Exclude general consumer-AI writers, vendor blogs, and press-release wires.

### 2. Preflight

In CLI mode:

```bash
newsjack auth status
newsjack credits balance
```

If the CLI is present but unauthenticated, run `newsjack login` and tell the user to approve `newsjack CLI` at the printed Medialyst link. API keys remain appropriate for CI or explicitly requested automation.

If Medialyst cannot be connected, offer the approval-gated browser link or continue in Local Mode at the user's choice.

### 3. Get approval, then create one job

Use one stable, unique idempotency key for the logical request and reuse that same key only when retrying the identical request:

```bash
newsjack media-lists create \
  --prompt "Find Canadian reporters covering PR technology and AI media tools" \
  --target-list-size 50 \
  --idempotency-key "campaign-2026-09-14-pr-ai"
```

The response contains `job_id`, `status`, the effective `target_list_size`, and `status_url`. Preserve the job ID immediately. The returned target can be lower than requested because of plan limits. An accepted response does not yet guarantee the hosted table exists.

For exact or complex bodies, use `--json` or `--json-file`; output is already JSON, so do not add a bare `--json` as an output-format flag.

### 4. Poll and show progress incrementally

Read the current snapshot with:

```bash
newsjack media-lists job <job-id> --include-results --limit 50
```

This command makes one status request. The agent owns the polling cadence:

- Poll every few seconds, or honor a server-provided retry interval. Never use a tight loop.
- Keep the user informed during multi-minute work. Report meaningful stage/percent, budget, or ready-row changes rather than narrating identical polls.
- Read `progress.stage`, `progress.percent`, `progress.message`, `result.total_rows`, `result.ready_rows`, and the `budget.credits` / `budget.reserved` / `budget.spent` fields when present. Prefer `progress.stage` over a conflicting top-level stage while the job is active.
- Accumulate top-level `rows` by a stable journalist ID or canonical profile URL; otherwise use normalized name+outlet. Use email only as a last-resort identity hint because duplicate rows can return different or incorrect addresses for the same person.
- `result.ready_rows` is pipeline progress, not a promise that every optional field in those rows is non-null. Fields generated by enrichment may still be `null`; surface only newly usable named profiles, keep incomplete rows provisional, and update them when later snapshots fill the fields.
- When `page.next_cursor` is non-null, request the next page with `--cursor`. A page contains at most 200 rows.
- Do not start another creation job while this one is `pending` or `processing`. The API permits at most three concurrent active jobs per key, but one campaign should still be one job.
- Do not treat `progress.percent: 100` or `ready_rows == total_rows` as terminal. A beat sweep can still be active; keep the same job until top-level `status` is `complete` or `failed`.

A useful progress update is compact:

> Media list: enrichment 72% — 36 of 50 rows ready. Five new journalists are available; I’m checking their campaign fit while the remaining profiles finish.

Status meanings:

- `pending`: accepted, not started
- `processing`: discovery or enrichment is running; partial results may exist
- `complete`: all Journalist Profile rows finished
- `failed`: stop, inspect `error`, and retry only if it is retryable and the approval rule above permits another attempt

### 5. Read the final pages and judge the list

At `complete`, follow every `page.next_cursor` and assemble the full normalized result. Each `journalist_list_v1` row may include identity and contact fields, outlet information, match score and reasoning, recent articles, and the source workflow ID.

Treat those fields as evidence, not an automatic verdict. Give every person one Newsjack status:

- `fit`: recent exact or near-exact coverage, clear beat overlap, usable pitch bridge
- `soft-fit`: adjacent evidence; pitch needs a specific edit
- `research-needed`: identity, current role, contact, or anchor is incomplete
- `cut`: wrong beat, stale, duplicate, weak evidence, unsafe hook, or filler

Do not call an outlet account, shared byline, or unresolved author a fit. Do not pad the result to the requested number. Medialyst's own guidance favors fewer, better pitches; the final first wave is usually much smaller than the research table.

Reconcile duplicate rows before presenting contacts. Keep the strongest well-supported fit evidence, but retain contradictory scores as an uncertainty note. Validate a returned email or social handle against the journalist's name, outlet, domain, profile, `email_affiliation`, and confidence/source fields. `email_source` by itself is not proof. Quarantine a mismatched or weakly affiliated address instead of presenting it as verified.

## MCP Mode

Use the API request shape directly:

```json
{
  "prompt": "Campaign brief, maximum 2,000 characters",
  "target_list_size": 50
}
```

Pass a stable `Idempotency-Key` when the tool supports headers or an idempotency field. `create_media_list` returns the job. Poll `get_media_list_job` with the job ID and result options equivalent to `include=results`, `limit`, and `cursor`.

Apply the same explicit credit approval, single-job, incremental-update, pagination, nullable-field, failure, and final-fit rules as CLI mode. Tool transport never changes the authorization boundary.

## Local Mode

Local Mode is best effort and should not imitate verified enrichment.

1. Search the core topic, sub-angles, competitors, proof hook, regions, and outlet tiers.
2. Read relevant editorial coverage and follow the entities and phrases it reveals into further searches.
3. Keep named journalists with dated, linked person-level evidence; quarantine wires, aggregators, brand content, vendor blogs, and outlet landing pages.
4. Audit for missed regions, freelancers/newsletters, adjacent beats, and breaking coverage until another search round yields no new real fits.
5. Return only defensible rows. Leave Contact blank and state that no live enrichment ran; never guess an email or handle.

End by offering either the approval-gated Medialyst deep link or a later agent-driven run with explicit credit approval.

## What To Show The User

During processing, show concise progress plus newly available rows. Do not dump repeated raw JSON.

At completion, lead with the outcome: campaign, standing, requested good-fit count, chosen research multiplier and target, actual credits spent, how many unique journalists resolved, how many made the first wave, and any unresolved gaps. Then show:

| Journalist | Outlet | Fit | Why them | Anchor piece | Pitch note | Contact |
| --- | --- | --- | --- | --- | --- | --- |
| Name | Publication | fit / soft-fit / research-needed / cut | Specific campaign-fit reason | Dated linked recent work | Recipient-specific bridge | Identity-checked email or handle; otherwise unresolved |

After the table include:

- cuts and their reasons
- job ID, final status, effective target, and workflow/list ID when returned
- whether the run was CLI, MCP, browser handoff, or Local Mode
- one next action: review the first wave, resolve gaps, or run `journalist-fit-check` on uncertain rows

If the API returns a direct hosted-list URL, link it. Otherwise say the completed table appears on the authenticated organization's Media Lists page; do not invent a URL pattern.

## Failure Handling

- `400`: correct the prompt or target; do not silently change campaign meaning
- `PROMPT_ANGLE_ARTICLE_LIMIT_TOO_SMALL`: some multi-angle campaign configurations require `target_list_size` of at least 3 even though the public field range begins at 1; explain the effective minimum before resubmitting
- `401`: authenticate with `newsjack login`
- `402`: the organization cannot create the list; report the credit/plan issue
- `403`: the credential lacks `media_lists:manage`
- `409`: the idempotency key was reused with different input; use the original input or ask before creating a new logical request
- `429`: respect backoff and wait for active jobs; do not fan out more jobs

Partial or failed results remain partial. Never fabricate missing names, contacts, anchors, completion, or a hosted table.

## Hard Gates

- No credible client standing: return a research shell at best; do not call it pitch-ready.
- No person-level anchor: `research-needed`, never `fit`.
- Large undifferentiated blast: refuse and offer a smaller segmented wave.
- Guessed identity, contact, title, date, or source: cut or mark `research-needed`.
- Auto-send or automated follow-up: refuse. This skill researches; it does not send.
- Tragedy or human suffering as a promotional hook: refuse under `skills/ETHICS.md`.

## Completion Check

Before calling the list ready, confirm:

- credit-bearing creation had explicit approval
- one logical campaign used one idempotent job
- polling reached `complete` or the output is plainly labeled partial
- every result page was read
- duplicate identities were collapsed
- every first-wave journalist has a dated evidence anchor and specific fit reason
- weak and unresolved rows were cut or labeled
- the output does not imply permission to mass-pitch or auto-send
