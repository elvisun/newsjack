import type { Adapter, AStamps, Client, Fit, Headline, JudgeResult, QuestionDef } from "./types";
import { AXES, CLIENT_QUESTION_KEYS, LAYER_A, LAYER_B_TEMPLATE } from "../questions";

export type RunEvent =
  | { kind: "phase"; phase: "A" | "B" | "done" | "stopped" }
  | { kind: "a"; headline: Headline; result: JudgeResult; stamps: AStamps }
  | { kind: "sorting"; headline: Headline }
  | { kind: "b"; headline: Headline; result: JudgeResult; fits: Fit[] }
  | { kind: "error"; headline: Headline; message: string };

export function buildLayerB(clients: Client[]): Record<string, QuestionDef> {
  const q: Record<string, QuestionDef> = {};
  for (const c of clients) for (const k of CLIENT_QUESTION_KEYS) q[`${c.id}.${k}`] = LAYER_B_TEMPLATE[k];
  return q;
}

export function stampsFrom(result: JudgeResult): AStamps {
  const a = result.answers;
  const axes = Object.fromEntries(AXES.map((k) => [k, a[k]?.score ?? 0])) as AStamps["axes"];
  return { is_news: a.is_news?.noul ?? 0, desk: a.desk?.choice ?? "", story_type: a.story_type?.choice ?? "", axes };
}

// A story lands on at most this many desks: the best-fitting ones by standing,
// journalist shape, tier and confidence. The rest are demoted to watch.
export const MAX_DESKS = 3;

export function fitsFrom(result: JudgeResult, clients: Client[], stamps: AStamps): Fit[] {
  const a = result.answers;
  const fits = clients.map((c) => {
    const g = (k: string) => a[`${c.id}.${k}`];
    let tier = (g("tier")?.choice ?? "watch") as Fit["tier"];
    const standing = g("standing")?.score ?? 0;
    const off = g("off_policy")?.noul ?? 0;
    let action = (g("action")?.choice ?? "skip") as Fit["action"];
    // Post-rules ported from the skills.
    if (stamps.axes.risk >= 4) action = "avoid";
    if (off > 0.7) tier = "watch";
    if (standing <= 1 && tier === "pitch_ready") tier = "big_story";
    return { clientId: c.id, standing, journalist_shape: g("journalist_shape")?.score ?? 0, tier, action, bridge_type: g("bridge_type")?.choice ?? "none", off_policy: off, confidence: g("tier")?.confidence ?? 0 };
  });
  const rank = (f: Fit) => f.standing * 2 + f.journalist_shape + (f.tier === "pitch_ready" ? 2 : 0) + f.confidence;
  const keep = new Set(fits.filter((f) => f.tier !== "watch").sort((x, y) => rank(y) - rank(x)).slice(0, MAX_DESKS).map((f) => f.clientId));
  return fits.map((f) => (f.tier !== "watch" && !keep.has(f.clientId) ? { ...f, tier: "watch" as const, action: "skip" as const } : f));
}

async function pool<T>(items: T[], n: number, fn: (t: T) => Promise<void>, signal: AbortSignal) {
  let i = 0;
  const workers = Array.from({ length: n }, async () => {
    while (i < items.length && !signal.aborted) { const item = items[i++]; await fn(item); }
  });
  await Promise.all(workers);
}

export async function runModel(
  adapter: Adapter,
  headlines: Headline[],
  clients: Client[],
  opts: { concurrency: number; signal: AbortSignal; onEvent: (e: RunEvent) => void },
) {
  const { signal, onEvent } = opts;
  const stampsById = new Map<string, AStamps>();
  const layerB = buildLayerB(clients);
  const clientState = clients.map((c) => ({ id: c.id, name: c.name, blurb: c.blurb, topics: c.topics, exclusions: c.exclusions }));
  try {
    onEvent({ kind: "phase", phase: "A" });
    await pool(headlines, opts.concurrency, async (h) => {
      try {
        const state = { id: h.id, title: h.title, excerpt: h.excerpt, source: h.source, published_at: h.published_at, feed: h.feed };
        const result = await adapter.judge(state, LAYER_A, signal);
        const a = stampsFrom(result);
        stampsById.set(h.id, a);
        onEvent({ kind: "a", headline: h, result, stamps: a });
        if (a.is_news < 0.5 || signal.aborted) return;
        onEvent({ kind: "sorting", headline: h });
        const bState = { headline: { id: h.id, title: h.title, excerpt: h.excerpt, source: h.source, feed: h.feed }, a: a.axes, clients: clientState };
        const bResult = await adapter.judge(bState, layerB, signal);
        onEvent({ kind: "b", headline: h, result: bResult, fits: fitsFrom(bResult, clients, a) });
      } catch (e) {
        // One bad call (rate limit, malformed answer) should not end the run.
        if (signal.aborted || (e as Error).name === "AbortError") return;
        console.error(`[${adapter.id}] ${h.id}`, e);
        onEvent({ kind: "error", headline: h, message: (e as Error).message ?? String(e) });
      }
    }, signal);
    onEvent({ kind: "phase", phase: signal.aborted ? "stopped" : "done" });
  } catch (e) {
    if ((e as Error).name === "AbortError" || signal.aborted) onEvent({ kind: "phase", phase: "stopped" });
    else throw e;
  }
}
