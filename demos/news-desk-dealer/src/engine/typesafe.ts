import type { Adapter, Answer, QuestionDef } from "./types";
import { callApi } from "./rpc";

// Live adapter for Jev (TypeSafe AI System One). Our QuestionDef map is
// already the API's question shape, so the request is state + questions.
// Score answers come back probability-weighted (e.g. 1.6); the app works in
// whole levels, so they are rounded. The dev server's /api/typesafe proxy
// holds the key.

const sleep = (ms: number, signal?: AbortSignal) => new Promise<void>((res, rej) => {
  const t = setTimeout(res, ms);
  signal?.addEventListener("abort", () => { clearTimeout(t); rej(new DOMException("aborted", "AbortError")); }, { once: true });
});

function toAnswers(raw: Record<string, any>, questions: Record<string, QuestionDef>): Record<string, Answer> {
  const out: Record<string, Answer> = {};
  for (const [k, q] of Object.entries(questions)) {
    const a = raw[k] ?? {};
    if (q.type === "choice") out[k] = { type: "choice", choice: a.choice ?? Object.keys(q.criteria)[0], probabilities: a.probabilities, confidence: a.confidence };
    else if (q.type === "score") out[k] = { type: "score", score: Math.max(0, Math.min(q.criteria.length - 1, Math.round(Number(a.score) || 0))), probabilities: a.probabilities, confidence: a.confidence };
    else out[k] = { type: "noul", noul: Number(a.noul) || 0 };
  }
  return out;
}

export function makeTypeSafeAdapter(o: { id: string; label: string; vendor: string; logo: string; concurrency: number; model: string; pricing: Adapter["pricing"]; path?: string }): Adapter {
  const path = o.path ?? "/v1/systemone";
  return {
    id: o.id, label: o.label, vendor: o.vendor, logo: o.logo, concurrency: o.concurrency, pricing: o.pricing,
    async judge(state, questions, signal) {
      const body = { model: o.model, state, questions };
      const t0 = performance.now();
      let lastErr: unknown;
      for (let attempt = 0; attempt < 4; attempt++) {
        if (attempt) await sleep(500 * 2 ** attempt, signal); // back off on 429 / 529
        const res = await callApi("/api/typesafe", path, body, signal);
        const text = res.text;
        if (res.status === 429 || res.status >= 500 || res.status === 0) { lastErr = new Error(`TypeSafe ${res.status}: ${text.slice(0, 200)}`); continue; }
        if (res.status < 200 || res.status >= 300) throw new Error(`TypeSafe ${res.status}: ${text.slice(0, 300)}`);
        const data = JSON.parse(text);
        return {
          answers: toAnswers(data.answers ?? {}, questions),
          usage: { input_tokens: data.usage?.input_tokens ?? 0, output_tokens: data.usage?.output_tokens ?? 0 },
          latency_ms: Math.round(performance.now() - t0),
        };
      }
      throw lastErr instanceof Error ? lastErr : new Error("TypeSafe: gave up");
    },
  };
}
