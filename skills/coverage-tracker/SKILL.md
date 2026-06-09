---
name: coverage-tracker
description: "Run a Google Alerts-style keyword coverage tracker. Uses news-search for recent keyword queries, lets the LLM dedupe and classify real features versus junk, stores decisions in SQLite, and alerts only on new real coverage."
when_to_use: "User wants to run coverage alerts, check new mentions, track press coverage, monitor brand/company keyword coverage, or execute a previously configured coverage tracker."
---

# Coverage Tracker

Track whether configured keywords appeared in real coverage. Keep this simple: `news-search` collects dated article evidence, you dedupe and judge with the keyword's meaning snippet, and the CLI only stores config plus seen/alert state.

This is not `newsjack-detector`. Do not score newsjacking opportunities, generate angles, assess standing, or use monitor profiles.

## Runtime Mode

- **Full Mode:** Use this in Claude Code, Codex, OpenClaw, Hermes, or another capable agent harness with shell, filesystem, network, and local CLI access. Full Mode can read tracker config, check SQLite seen-state, record decisions, and suppress repeat alerts.
- **Limited Mode:** Use this in Claude.ai chat, ChatGPT chat, Claude Cowork, or any restricted runtime without shell/filesystem/CLI access. Do not attempt `curl`, `npm`, or on-demand CLI installation. Run a one-off manual coverage check from pasted/searchable evidence and disclose that no tracker config, SQLite state, or repeat suppression was used.

**Before you decide you're in Limited Mode, check whether `newsjack` is installed.** It ships as a prebuilt, bundled binary — you do **not** need Go, a compiler, or any build/install step to run it. Never look for a Go toolchain, and never declare the CLI "missing" or tell the user they need a "Go environment" without running this check first:

1. Run `newsjack --version`. If it prints a version, you're in Full Mode — use plain `newsjack ...` for every command.
2. If `newsjack` isn't on `PATH`, try the bundled location `~/.newsjack/bin/newsjack --version`. If that prints a version, use that full path in place of `newsjack` everywhere below.
3. Only if **both** fail (and you genuinely have no shell) are you in Limited Mode.

The bundled binary is almost always already installed — assume Full Mode and verify, don't assume it's missing.

## Workflow

In Limited Mode, skip the CLI state steps and:

1. Ask for the keyword, what it means, lookback window, and any exclusions if they are not already present.
2. Use pasted links first. If search tools are available, search the keyword plus a recency bound and keep dated, attributed article evidence when available.
3. Dedupe obvious repeats by URL/title/outlet/date.
4. Classify each item with the verdicts below.
5. Return a short inline report with new likely coverage, filtered counts, and a Limited Mode caveat that repeat suppression was unavailable.

In Full Mode, run the persistent workflow:

1. **Find the tracker.**
   - If the user gave a slug, run `newsjack coverage status <slug>`.
   - If working in a source checkout, prefer `bin/newsjack`.
   - If no slug is given, ask which tracker to run unless local context makes it obvious.
   - Read the returned `config_path`.

2. **Search each keyword.**
   - Read the current date from the system (e.g. run `date`); do not recall it from memory. Compute `since_date` as that date minus `lookback_days` from the config, defaulting to `2`.
   - For each keyword, call `news-search` with exactly:

     ```text
     "keyword" after:YYYY-MM-DD
     ```

   - If the keyword has aliases, search each alias the same way, but keep the original keyword entry attached.
   - Keep dated, attributed article fields from `news-search`: `title`, `url`, `outlet`, `author`, `published_at`, and snippet/summary when available.

3. **Dedupe and check stored decisions before using the LLM.**
   - Dedupe exact canonical URLs first, then obvious same-article duplicates by title/outlet/date.
   - Create a minimal candidate JSON file only because the CLI helper consumes a file. Put it in a temporary location, or in a timestamped run folder if your harness normally keeps run artifacts:

     ```json
     {
       "items": [
         {
           "keyword": "profound",
           "title": "Article title",
           "url": "https://...",
           "outlet": "Outlet",
           "author": "Author or null",
           "published_at": "2026-06-05T12:00:00Z",
           "snippet": "Search snippet or summary"
         }
       ]
     }
     ```

   - Run:

     ```bash
     newsjack coverage check <slug> --input candidates.json
     ```

   - Do not reclassify `known_items` unless the user explicitly asks for a fresh review. Use their `prior_decision` to count filtered/known results and suppress repeat alerts.
   - Classify only `unknown_items`.

4. **LLM classify unknown items.**
   - Use the keyword's `means` field as the authority for entity matching.
   - Reject generic-word usage and wrong entities, especially for ambiguous keywords like `profound`.
   - Do not alert from title/snippet keyword presence alone. When the snippet leaves the verdict unclear, read the article before deciding.

   Use these verdicts:

   - `real_feature`: the article is substantially about the intended entity/product/person. Alert.
   - `substantive_mention`: meaningful paragraph-level mention, but not a feature. Save, normally no alert.
   - `passing_mention`: brief mention only. Save, no alert.
   - `wrong_entity`: the keyword refers to something else. Save, no alert.
   - `junk`: SEO, scraper, duplicate landing page, job post, docs/help page, or non-news. Save, no alert.
   - `uncertain`: insufficient evidence. Save, no alert.

5. **Record only newly classified unknown items.**
   If there are no `unknown_items`, skip `coverage record` and report from the `coverage check` result.

   If you classified new items, write the minimum `decisions.json` needed by the CLI:

   ```json
   {
     "items": [
       {
         "keyword": "profound",
         "title": "Article title",
         "url": "https://...",
         "outlet": "Outlet",
         "author": "Author or null",
         "published_at": "2026-06-05T12:00:00Z",
         "verdict": "real_feature",
         "confidence": "high",
         "alert": true,
         "rationale": "Why this is about the intended keyword."
       }
     ]
   }
   ```

   Include every newly classified unknown item, not only alerts. Saving rejects lets future runs skip them through `coverage check`.

6. **Persist decisions and suppress repeat alerts.**

   ```bash
   newsjack coverage record <slug> --input decisions.json
   ```

   Add `--run-dir RUN_DIR` only when you already created a run folder for harness provenance. Read the command's JSON stdout directly. Do not write `record.json` unless the user or harness explicitly wants artifacts. Only articles in `new_alerts` are newly alertable; previously alerted URLs must not be re-alerted.

7. **Deliver the result to the user.**
   SQLite is the durable record. Do not write a separate report file; compose a short, readable message straight to the user from `coverage check` and, when run, `coverage record`:
   - If `new_alerts` is non-empty, lead with the new real coverage items.
   - If there are no new alerts, say there is no new confirmed coverage.
   - Include a short run summary: keywords searched, since date, candidate count, skipped-known count, newly judged count, new alert count.
   - Include a quiet "Filtered" line with counts by verdict, including prior decisions from known items, not a dump of every junk item.
   - Use clickable links.

   This skill is meant to run unattended on a schedule, so the message must stand alone — the user should not have to open any file to understand the run.

## Output Discipline

- Do not invent publication dates, outlets, or authors. If `news-search` cannot recover a date, classify cautiously.
- Do not use story-origin or freshness gates; this is not a newsjack opportunity scan.
- Do not call `angle-generator`, `journalist-fit-check`, or `meanest-editor`.
- Do not install native cron. Recurrence belongs to the agent harness.
