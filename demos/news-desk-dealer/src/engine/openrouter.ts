import type { Adapter, Answer, QuestionDef } from "./types";
import { callApi } from "./rpc";

// Live adapter for an LLM on OpenRouter. The typed question set becomes a
// strict JSON schema; the model answers every question in one call. The
// browser talks to the dev server's /api/openrouter proxy, which holds the key.

const SYSTEM = `You are a wire editor's judgment engine. You receive a STATE (a headline, or a headline plus the companies on a PR desk) and a map of QUESTIONS.
Answer every question, keyed exactly as given, in the JSON shape required.
- choice: pick one option key from "criteria"; "confidence" is your probability (0-1) that this option is right.
- score: return the 0-based index into the ordered "criteria" levels; "confidence" is your probability that this level is right.
- noul: return the probability (0-1) that the proposition in "instructions" is true.
Follow the criteria text literally. Do not explain.`;

const STRICT_MAX_QUESTIONS = 30;

function schemaFor(questions: Record<string, QuestionDef>) {
  const properties: Record<string, unknown> = {};
  for (const [k, q] of Object.entries(questions)) {
    if (q.type === "choice") properties[k] = { type: "object", properties: { choice: { type: "string", enum: Object.keys(q.criteria) }, confidence: { type: "number" } }, required: ["choice", "confidence"], additionalProperties: false };
    else if (q.type === "score") properties[k] = { type: "object", properties: { score: { type: "integer", minimum: 0, maximum: q.criteria.length - 1 }, confidence: { type: "number" } }, required: ["score", "confidence"], additionalProperties: false };
    else properties[k] = { type: "object", properties: { noul: { type: "number" } }, required: ["noul"], additionalProperties: false };
  }
  return { type: "object", properties, required: Object.keys(properties), additionalProperties: false };
}

const clamp01 = (v: unknown) => Math.max(0, Math.min(1, Number(v) || 0));
function spread(options: string[], pick: string, conf: number): Record<string, number> {
  const rest = options.length > 1 ? (1 - conf) / (options.length - 1) : 0;
  return Object.fromEntries(options.map((o) => [o, +(o === pick ? conf : rest).toFixed(3)]));
}

function toAnswers(raw: Record<string, any>, questions: Record<string, QuestionDef>): Record<string, Answer> {
  const out: Record<string, Answer> = {};
  for (const [k, q] of Object.entries(questions)) {
    const a = raw[k] ?? {};
    if (q.type === "choice") {
      const opts = Object.keys(q.criteria);
      const choice = opts.includes(a.choice) ? a.choice : opts[0];
      const conf = clamp01(a.confidence ?? 0.5);
      out[k] = { type: "choice", choice, confidence: conf, probabilities: spread(opts, choice, conf) };
    } else if (q.type === "score") {
      const n = q.criteria.length;
      const score = Math.max(0, Math.min(n - 1, Math.round(Number(a.score) || 0)));
      const conf = clamp01(a.confidence ?? 0.5);
      const opts = q.criteria.map((_, i) => String(i));
      out[k] = { type: "score", score, confidence: conf, probabilities: spread(opts, String(score), conf) };
    } else {
      out[k] = { type: "noul", noul: clamp01(a.noul) };
    }
  }
  return out;
}

function parseJson(text: string): Record<string, any> {
  const t = text.trim().replace(/^```(?:json)?\s*/i, "").replace(/```\s*$/, "");
  const start = t.indexOf("{"), end = t.lastIndexOf("}");
  return JSON.parse(start >= 0 && end > start ? t.slice(start, end + 1) : t);
}

const sleep = (ms: number, signal?: AbortSignal) => new Promise<void>((res, rej) => {
  const t = setTimeout(res, ms);
  signal?.addEventListener("abort", () => { clearTimeout(t); rej(new DOMException("aborted", "AbortError")); }, { once: true });
});

export function makeOpenRouterAdapter(o: { id: string; label: string; vendor: string; logo: string; concurrency: number; model: string; pricing: Adapter["pricing"]; path?: string }): Adapter {
  const path = o.path ?? "/chat/completions";
  return {
    id: o.id, label: o.label, vendor: o.vendor, logo: o.logo, concurrency: o.concurrency, pricing: o.pricing,
    async judge(state, questions, signal) {
      // Anthropic rejects strict grammars past a few dozen properties ("compiled
      // grammar is too large"), so the 90-question set-B fan-out is asked for
      // as plain JSON described in the prompt and parsed leniently instead.
      const schema = schemaFor(questions);
      const strict = Object.keys(questions).length <= STRICT_MAX_QUESTIONS;
      const body = {
        model: o.model,
        temperature: 0,
        max_tokens: 8192,
        messages: [
          { role: "system", content: strict ? SYSTEM : `${SYSTEM}\nReply with one JSON object and nothing else. It must match this JSON schema exactly:\n${JSON.stringify(schema)}` },
          { role: "user", content: JSON.stringify({ state, questions }) },
        ],
        ...(strict ? { response_format: { type: "json_schema", json_schema: { name: "answers", strict: true, schema } } } : {}),
        usage: { include: true },
      };
      const t0 = performance.now();
      let lastErr: unknown;
      for (let attempt = 0; attempt < 4; attempt++) {
        if (attempt) await sleep(800 * 2 ** attempt, signal);
        const res = await callApi("/api/openrouter", path, body, signal);
        const text = res.text;
        if (res.status === 429 || res.status >= 500 || res.status === 0) { lastErr = new Error(`OpenRouter ${res.status}: ${text.slice(0, 200)}`); continue; }
        if (res.status < 200 || res.status >= 300) throw new Error(`OpenRouter ${res.status}: ${text.slice(0, 300)}`);
        const data = JSON.parse(text);
        const content = data.choices?.[0]?.message?.content;
        if (typeof content !== "string") { lastErr = new Error(`OpenRouter: no content in response ${text.slice(0, 200)}`); continue; }
        let raw: Record<string, any>;
        try { raw = parseJson(content); } catch (e) { lastErr = new Error(`OpenRouter: bad JSON: ${String(e)}`); continue; }
        return {
          answers: toAnswers(raw, questions),
          usage: { input_tokens: data.usage?.prompt_tokens ?? 0, output_tokens: data.usage?.completion_tokens ?? 0 },
          latency_ms: Math.round(performance.now() - t0),
        };
      }
      throw lastErr instanceof Error ? lastErr : new Error("OpenRouter: gave up");
    },
  };
}
