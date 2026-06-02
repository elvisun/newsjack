# Harness Routing & Chunking

The two coarse passes (relevance, story-origin) are low-cost tasks. They become low-cost *model* passes only when the harness supports model selection or low-cost worker/subagent routing. Story-origin is also retrieval-bound: a cheaper worker is acceptable only if it can run live `news_search` and fetch/open surfaced URLs, or if the orchestrator gives it extracted search/page evidence. The artifact contracts (`coarse_relevance_decisions.json`, `origin_findings.json`) are the same regardless of harness — only the cost story changes, and that must be disclosed.

## Execution decision path

Before each coarse pass, identify the harness and take the first available path. For story-origin, run a retrieval probe before choosing worker fanout.

1. **Direct low-cost-model path.** If the harness can choose a model for a single step, run the coarse prompt with the lowest-cost reliable model below, then run the expensive pass with the strong model.
2. **Low-cost subagent/worker path.** If the harness can't switch the current model but can spawn workers/subagents with a model hint, split candidates into chunks. Relevance workers return only `decisions`; origin workers return only `findings`. For story-origin, use this path only after proving the worker has retrieval tools or after passing extracted evidence into the worker. Merge each pass into its single JSON artifact.
3. **Current-model fallback with retrieval.** If neither cost-optimized path is possible but the current harness has retrieval, run the prompts with the current model and state explicitly in the final response: `coarse passes ran with current model; this was semantic multi-stage, not cost-optimized multi-stage`.
4. **Retrieval unavailable.** If story-origin cannot access live `news_search` and page fetch/open tools, stop before `origin_findings.json` and return `story_origin_retrieval_unavailable` with the missing tools and retry action. Do not produce same-story verdicts from detector metadata alone.

## Harness hints

- **Claude Code / Claude-style coding harnesses:** if a `Task`/subagent tool or model override exists, use it for chunks and request the latest Haiku/low-cost alias; use the latest Sonnet or Opus alias for the expensive pass. No control exposed → current-model fallback.
- **Codex:** if `spawn_agent` is available and model override is allowed, spawn workers with a low-reasoning small/fast model (`gpt-5.4-mini`, `gpt-5-nano`, or the newest GPT-5.x at low reasoning). Use `gpt-5.5` or the strongest Codex model at medium/high reasoning for the expensive pass. If override is not allowed, spawn default workers for parallelism or use fallback; disclose which.
- **OpenClaw:** relevance can use low-cost worker fanout. For story-origin, prefer inline execution on the main OpenClaw agent for normal cron batches (<20 cluster representatives), because main OpenClaw can be MCP-wired while `TaskCreate` / Claude ACP subagents may not inherit `news_search` or reliable page fetch. Use `TaskCreate` / Claude ACP subagents for story-origin only after an explicit tool probe succeeds inside that worker, or after the main orchestrator passes extracted search/page evidence into the worker. Low-cost models: Gemini 3 Flash Preview, Claude Haiku/latest alias, or GPT-5.x low-reasoning. Expensive models: Gemini 3 Pro Preview, Claude Sonnet/Opus latest, or GPT-5.5+ higher reasoning.
- **API harnesses:** call the configured low-cost model for coarse prompts, then the configured stronger model for the expensive rubric pass.
- **Unknown harness:** current-model fallback unless an explicit low-cost-model or worker mechanism is exposed.

## Preferred models

- **Coarse passes:** Gemini 3 Flash Preview, Claude Haiku/latest Haiku alias, GPT-5-nano, GPT-5.4-mini low reasoning, GPT-5.5 low reasoning, or the harness's lowest-cost fast equivalent.
- **Expensive pass:** Gemini 3 Pro Preview, Claude Sonnet/Opus latest aliases, GPT-5.5 medium/high reasoning, GPT-5.4 medium/high reasoning, or the harness's strongest reasoning model.
- Treat Gemini 2.5 Pro as stale for this pipeline — fallback only when Gemini 3 Pro is unavailable. Gemini 2.5 Flash/Flash-Lite are coarse-pass fallbacks when Gemini 3 Flash is unavailable.
- If the exact named model is unavailable, choose the closest current low-cost/fast model for pass 1 and the closest current strong reasoning model for pass 2.

## Chunking guidance

- 1–15 signals: one call per coarse pass is fine. For story-origin, keep this inline on a retrieval-capable main harness unless a worker retrieval probe has passed.
- 16–40 signals: split into chunks of 8–15. Prefer at least 2 workers/subagents when the harness exposes them and, for story-origin, after worker retrieval is proven.
- 41–80 signals: split into chunks of 8–12. Use worker/subagent fanout if available.
- More than 80 signals: do not ask one low-cost call or one subagent to process everything. Split into chunks of 8–12; if that would need more than 8 low-cost workers, tighten detector `--limit`, lane caps, or rerun by profile/source lane before filtering.
- Each chunk must include the profile context plus only its assigned signals. Each worker returns only items for its assigned signal IDs.
- The merged `coarse_relevance_decisions.json` and `origin_findings.json` must each contain exactly one item per input signal unless intentionally using `--allow-missing`.
- Do not let coarse workers perform the expensive rubric pass, compare across chunks, pick best bets, or write the final report.
- If the harness has low-cost workers but no model override, still split large sets for reliability and disclose that model cost was not optimized.
- The story-origin pass needs retrieval. If a low-cost worker cannot open pages or search the web, either give it extracted page/search evidence from the orchestrator or run that pass in the current harness with retrieval tools. If neither is possible, hard-fail with `story_origin_retrieval_unavailable` — do not let a worker return a story-identity verdict it had no evidence for.
