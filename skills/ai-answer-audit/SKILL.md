---
name: ai-answer-audit
description: "Audit how major AI engines repeat a brand's user-supplied verified facts, score accuracy and coverage by engine, trace every wrong or outdated claim to its cited page and upstream root, and rank the pages to correct. Use when a user wants a point-in-time AI answer accuracy audit, not verification of their own facts."
when_to_use: "User supplies a brand and facts they treat as ground truth, then asks what ChatGPT, Gemini, Google AI Mode, Perplexity, Claude, or other AI answers say; wants a fact-by-engine accuracy grid; wants wrong AI claims traced to their feeding sources; or wants pages prioritized for correction. Not for checking claims in the user's own draft against outside evidence — use fact-check for that — and not for ongoing citation-share tracking."
metadata:
  category: AI visibility
---

# AI Answer Audit

Audit whether AI engines repeat a brand's known facts accurately, then trace the sources behind their errors.

The user's supplied facts are the ground truth for this audit. **Never re-verify, correct, or overrule them.** This is not `fact-check`: that skill checks claims in the user's draft against outside evidence. Here, the audited material is the engines' answers. If a supplied fact is vague, subjective, or not checkable in an answer — “we are the best,” for example — ask the user to restate it as a checkable fact or drop it. Do not decide whether it is true.

This is a molecule. It finds and grades wrong claims, checks which first-degree cited pages carry them, and then hands supported cited-page cases to `citation-foodchain`. Do not recreate that skill's upstream tracing logic.

Follow `skills/ETHICS.md`. Never plant, spoof, or fabricate a source. Never promise that correcting a page will change an AI answer.

## 1. Prepare the audit

Ask for:

- the brand name
- the verified facts, written exactly as the user wants them checked
- optional aliases, similarly named companies, competitors, and country to disambiguate the brand
- optional engines or run count; otherwise offer the five-engine set below and one run

Number the facts `F1`, `F2`, and so on. Split compound facts into atomic facts while preserving names, numbers, dates, prices, units, negatives, and qualifications verbatim. A fact such as “Pineglass Labs was founded in 2019 in Halifax” becomes two facts unless the user says the combination matters.

For each fact, write at least one natural question a buyer, journalist, or job-seeker might ask that would surface it. Do not put the answer in the question: “Was Pineglass founded in 2019?” measures agreement, not retrieval. A natural question may cover several atomic facts, but map it to every relevant fact id. Add a few broad questions such as “What is [brand]?”, “Is [brand] any good?”, and “[brand] pricing” because errors often appear unprompted. Every question must be at most 500 characters.

Before capture, show the user:

- the numbered facts
- the question list and its fact mapping
- the engines and country
- the planned repetitions
- the maximum Medialyst credit exposure, if that lane is available

## 2. Capture answers and citations

Use the first workable lane.

### Medialyst AI answers — preferred, optional

Use the REST endpoint, CLI wrapper, or already-connected MCP tool:

- REST: `POST /api/v1/ai-answers`
- CLI: `newsjack ai answers --prompt "..." --engines chatgpt,gemini,google_ai_mode,perplexity,claude --country US`
- MCP: `ask_ai_engines`

The request contract is:

```json
{
  "prompt": "question, maximum 500 characters",
  "engines": ["chatgpt", "gemini", "google_ai_mode", "perplexity", "claude"],
  "country": "US"
}
```

The response has `answers[]`, one per engine. Each answer has `engine`, `model`, `status` (`ok` or `error`), `answer` in Markdown, `sources[]` for cited pages, and `retrieved[]` for pages read but not cited. Source records include `url`, `domain`, `title`, `snippet`, and outlet data when known. The top level includes `creditsCharged`.

This lane costs **1 credit per successful engine answer; a failed engine is free**. Before calling it, state the maximum as `questions × engines × repetitions` credits and get explicit approval. A general request for an audit is not credit approval. If the current-turn instruction explicitly authorizes using Medialyst credits, do not ask twice. Do not rerun a failed request or add repetitions without renewed approval unless the user already authorized them.

This endpoint may not be enabled on every account. A `404`, `403`, or absent `ask_ai_engines` tool means “use the next lane,” not “the audit failed.”

### Citation-foodchain capture lanes

Read `citation-foodchain/SKILL.md` and use its capture lanes in order: public share link through `trace.mjs`, the user's authenticated browser, pasted answers plus their source panels, or the user's own model API keys. Keep its distinction between first-degree **cited** pages and merely **retrieved** pages. Never automate consumer AI products at volume.

Store the answer text, engine, model, time, question, status, cited sources, retrieved pages, capture lane, and run id. One complete pass over the approved question set is one audit run.

## 3. Grade each fact and engine

For each fact-engine pair in each audit run, read every successful answer to a question mapped to that fact, plus the broad answers. Assign one verdict:

- `correct` — states the supplied fact without a material change
- `wrong` — contradicts it; a hedge that still gives the wrong value is wrong
- `outdated_or_partial` — gives an older value, drops a material qualification, or states only a related part
- `silent` — the relevant successful answers never touch the fact; silence is not an error
- `not_asked_or_engine_failed` — no relevant question ran successfully for that engine

When several answers in one run disagree, use the most serious observed verdict: `wrong` before `outdated_or_partial`, then `correct`, then `silent`. Preserve every exact engine sentence that caused the verdict so the inconsistency remains visible. Flag `wrong_brand: true` when the engine describes a different company with a similar name.

Do not grade claims outside the supplied fact list. Put notable extras in **Unverified, not graded**, with the exact engine sentence, so the user can decide whether to add them to a future audit.

Use the deterministic scorer after grading:

```bash
node score.mjs graded-audit.json
```

`score.mjs` sits beside this file in a source checkout and in the full Newsjack install at `~/.newsjack/newsjack/skills/ai-answer-audit/`. The hosted plugin holds instructions only. If the script is unavailable, apply the same formula and say the scores were calculated manually.

For each engine and overall:

- accuracy = `correct / (correct + wrong + outdated_or_partial)`
- coverage = non-silent graded observations divided by `(graded + silent)`
- report silent and failed observations separately

With repeated audit runs, score the fact-engine observation from each run and show verdict counts in the grid instead of hiding variance.

## 4. Find the source of every wrong or outdated claim

For each exact wrong or outdated sentence, inspect only the pages cited for the answer that contained it. Retrieved-but-not-cited pages are context, never first-degree citations.

Use `citation-foodchain/trace.mjs` to open each cited page and locate candidate passages:

```bash
node trace.mjs fetch <cited-url> --claim "<exact wrong sentence>"
```

Read the passage yourself; `verdict_hint` is only a locator. Record one of three outcomes:

1. **The cited page carries the claim.** Quote its exact words and URL. Hand the wrong claim, supplied correct fact, question, engine, answer sentence, cited URL, and run date to `citation-foodchain`; state that the supplied fact is audit ground truth, not a fact-check target. Use that skill's returned hops, root, carrier, drift, and confidence directly. If several cited pages carry the claim, trace each.
2. **The cited pages do not carry the claim.** Quote any nearby, partial, or contradictory passage and say whether a source misread or an answer from memory is more likely, with a short reason. This is an inference, not a trace. A nearby number, dropped qualifier, or citation attached to the wrong sentence favors “misread”; no matching passage and no claim-level citation favors “memory.” Use `unknown` when neither is defensible.
3. **The answer has no citations.** Say so. There is no first-degree source to trace; do not invent one through search.

Do not repeat the upstream-hop rubric, root taxonomy, or carrier logic from `citation-foodchain`. Invoke it and preserve its artifact.

## 5. Rank realistic fixes

Group errors by the cited page that carries them and by the upstream root returned by `citation-foodchain`. For each group calculate reach as:

`distinct engines × distinct wrong-or-outdated claim ids`

Sort highest reach first. Show the page, root, claims, engines, reach, ownership, and one realistic action:

- the user's own page: correct or clarify it; use `ai-visibility-writing` when the prose needs an AI-discoverability pass
- an outlet: request a factual correction with the exact sentence and evidence
- a directory or wiki: use its proper correction or editing process and disclose conflicts where required
- an unfixable or unknown root: publish a better sourced page only when the user has real evidence; never plant or spoof corroboration

Do not call an unsupported engine claim a “page to fix.” Do not promise rankings, citations, traffic, or changed AI output.

## Variance and stopping point

AI answers vary across runs. One run is one sample. Recommend repeating every wrong or borderline result before acting; three runs is a useful default when the user approves the added cost or can provide captures. Label every score with the audit date, engines, models, capture lanes, and number of runs.

Stop when every wrong or outdated claim has one of the three source outcomes, every cited-page carrier has a `citation-foodchain` result or an honest limitation, and the fix list contains no action unsupported by the evidence.

## Output

Lead with readable Markdown:

1. **Score line** — overall accuracy, coverage, silent count, date, engines, models, runs
2. **Engine scores**
3. **Fact-by-engine grid** — facts as rows; engines as columns; show counts when runs disagree
4. **Wrong-claim cards** — supplied fact, exact engine sentence, engine/model/run, cited-page outcome and quote, upstream chain/root, wrong-brand flag
5. **Fix list** — ranked by reach
6. **Unverified, not graded**
7. **Caveats** — variance, failed engines, missing citations, unreachable pages, inferred causes

Then emit a clearly labeled JSON artifact. Keep this shape so `score.mjs` and later runs can reuse it:

```json
{
  "skill": "ai-answer-audit",
  "version": 1,
  "brand": "Brand name",
  "aliases": [],
  "country": "US",
  "audited_at": "ISO timestamp",
  "facts": [{ "id": "F1", "text": "user-supplied fact, verbatim" }],
  "questions": [{ "id": "Q1", "text": "non-leading question", "fact_ids": ["F1"] }],
  "engines": [{ "id": "chatgpt", "label": "ChatGPT", "models": ["model name"] }],
  "runs": [
    {
      "id": "R1",
      "run_at": "ISO timestamp",
      "capture": "medialyst | share_link | browser | paste | model_api",
      "answers": [
        {
          "question_id": "Q1",
          "engine": "chatgpt",
          "model": "model name",
          "status": "ok | error",
          "answer": "answer markdown or null",
          "sources": [{ "url": "https://...", "domain": "example.com", "title": "title", "snippet": "snippet" }],
          "retrieved": []
        }
      ]
    }
  ],
  "grades": [
    {
      "fact_id": "F1",
      "engine": "chatgpt",
      "run_id": "R1",
      "verdict": "correct | wrong | outdated_or_partial | silent | not_asked_or_engine_failed",
      "engine_quotes": ["exact sentence behind every non-silent verdict"],
      "question_ids": ["Q1"],
      "wrong_brand": false,
      "notes": "short reason"
    }
  ],
  "wrong_claims": [
    {
      "id": "W1",
      "fact_id": "F1",
      "engine": "chatgpt",
      "model": "model name",
      "run_id": "R1",
      "question_id": "Q1",
      "sentence": "exact wrong or outdated sentence",
      "expected_fact": "F1 text",
      "source_outcome": "cited_page_carries | cited_pages_do_not_carry | no_citations",
      "sources_checked": [
        {
          "url": "https://...",
          "outcome": "carries_claim | does_not_carry | unreachable",
          "quote": "verbatim cited-page passage or null",
          "likely_cause": "source_misread | answer_from_memory | unknown | null",
          "explanation": "short evidence-based reason",
          "foodchain": "full citation-foodchain artifact or null"
        }
      ]
    }
  ],
  "unverified_claims": [{ "engine": "chatgpt", "run_id": "R1", "sentence": "not in supplied facts" }],
  "fixes": [
    {
      "feeder_url": "https://... or null",
      "root_url": "https://... or null",
      "wrong_claim_ids": ["W1"],
      "engines": ["chatgpt"],
      "reach": 1,
      "action": "realistic correction path"
    }
  ],
  "caveats": ["one run is one sample"]
}
```

## Worked example — invented brand

The user says Pineglass Labs was founded in 2019 and plans start at $29 per month. ChatGPT says, *“Pineglass Labs was founded in 2018,”* and cites a directory whose page says, *“Founded: 2018.”* It gives the price correctly. Gemini stays silent on the founding year and says plans start “around $30.”

| Fact | ChatGPT | Gemini |
| --- | --- | --- |
| F1: Founded in 2019 | wrong | silent |
| F2: Plans start at $29/month | correct | outdated/partial |

ChatGPT accuracy is 50% (1/2); Gemini accuracy is 0% (0/1), with one silent fact; overall accuracy is 33.3% (1/3). The wrong-claim card quotes both ChatGPT and the directory, then hands the directory claim to `citation-foodchain`. If the directory is the unsourced root, the fix list ranks that page by its one-engine, one-claim reach and recommends its normal correction process. The report still warns that one run is one sample.
