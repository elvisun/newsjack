import type { Adapter, Answer, JudgeResult, QuestionDef } from "./types";

// Mock adapter. Answers come from cheap keyword heuristics plus a seeded PRNG so
// the demo is reproducible and the lanes look sensible. It simulates latency and
// token usage per model so the counters behave like the real thing.

type MockOpts = {
  id: string;
  label: string;
  vendor: string;
  logo: string;
  concurrency: number;
  latency: [number, number]; // ms per set-A call, uniform
  latencyB: [number, number]; // ms per set-B fan-out call
  outputTokensPerQuestion: number; // 0 for Jev (typed answers), >0 for an LLM
  flipRate: number; // probability a categorical answer is perturbed vs the base heuristic
  pricing: { input_per_mtok: number; output_per_mtok: number };
  speed: () => number; // multiplier, 1 = realistic
};

function hash(s: string) {
  let h = 2166136261;
  for (let i = 0; i < s.length; i++) { h ^= s.charCodeAt(i); h = Math.imul(h, 16777619); }
  return h >>> 0;
}
function rng(seed: number) {
  let a = seed || 1;
  return () => { a ^= a << 13; a ^= a >>> 17; a ^= a << 5; return ((a >>> 0) % 10000) / 10000; };
}
const sleep = (ms: number, signal?: AbortSignal) =>
  new Promise<void>((res, rej) => {
    if (signal?.aborted) return rej(new DOMException("aborted", "AbortError"));
    const t = setTimeout(res, ms);
    signal?.addEventListener("abort", () => { clearTimeout(t); rej(new DOMException("aborted", "AbortError")); }, { once: true });
  });

const RX = {
  incident: /shoot|crash|dead|death|killed|attack|explosion|fire|breach|hacked|recall|scandal|arrest|murder|war crimes|strike/i,
  regulation: /court|ruling|judge|lawsuit|regulator|ftc|sec\b|cftc|antitrust|bill|law\b|ban\b|congress|senate|tariff|sanction|indict/i,
  funding: /raises|funding|acquire|acquisition|merger|ipo|valuation|\$\d+ ?[bm]illion|buyout|deal\b/i,
  data: /study|report|survey|data|index|earnings|quarter|research|found that|scientists|analysis/i,
  personnel: /ceo|hires|steps down|resigns|appoint|names .* as|fired|chief/i,
  announce: /launch|unveil|announce|rolls out|introduc|plans to|debut|reveal/i,
  opinion: /^why |^how |opinion|column|explained|what .* means/i,
  big: /trump|iran|war\b|fed\b|boj|apple|openai|anthropic|google|nvidia|tesla|amazon|microsoft|meta\b|supreme court|nasa|fda/i,
  safety: /shoot|dead|death|killed|murder|crash|war crimes|genocide|suicide|abuse|assault/i,
};
const DESK_BY_FEED: Record<string, string> = { technology: "tech", business: "business", science: "science", health: "health", us: "policy", world: "world", techmeme: "tech" };

function choose<T extends string>(r: () => number, weights: Record<T, number>): T {
  const total = Object.values<number>(weights).reduce((a, b) => a + b, 0);
  let x = r() * total;
  for (const [k, w] of Object.entries<number>(weights)) { x -= w; if (x <= 0) return k as T; }
  return Object.keys(weights)[0] as T;
}
function probs(options: string[], pick: string, conf: number, r: () => number) {
  const p: Record<string, number> = {};
  let rest = 1 - conf;
  const others = options.filter((o) => o !== pick);
  others.forEach((o, i) => { const v = i === others.length - 1 ? rest : rest * r() * 0.8; p[o] = +v.toFixed(3); rest -= v; });
  p[pick] = +conf.toFixed(3);
  return p;
}
function choiceAnswer(q: Extract<QuestionDef, { type: "choice" }>, pick: string, r: () => number, conf?: number): Answer {
  const c = conf ?? 0.55 + r() * 0.44;
  return { type: "choice", choice: pick, probabilities: probs(Object.keys(q.criteria), pick, c, r), confidence: +c.toFixed(3) };
}
function scoreAnswer(q: Extract<QuestionDef, { type: "score" }>, idx: number, r: () => number): Answer {
  const c = 0.5 + r() * 0.45;
  const opts = q.criteria.map((_, i) => String(i));
  return { type: "score", score: idx, probabilities: probs(opts, String(idx), c, r), confidence: +c.toFixed(3) };
}
const noul = (p: number): Answer => ({ type: "noul", noul: +Math.min(0.99, Math.max(0.01, p)).toFixed(3) });

function maybeFlip<T extends string>(r: () => number, rate: number, value: T, options: T[]): T {
  if (r() < rate) { const alt = options.filter((o) => o !== value); return alt[Math.floor(r() * alt.length)]; }
  return value;
}

type HeadlineState = { id: string; title: string; excerpt: string; source: string; feed: string };
type ClientState = { id: string; name: string; topics: string[]; exclusions: string[] };

const clamp = (v: number) => Math.max(0, Math.min(4, Math.round(v)));
const jitter = (r: () => number, v: number, flip: number) => (r() < flip ? clamp(v + (r() < 0.5 ? -1 : 1)) : clamp(v));

function layerA(state: HeadlineState, questions: Record<string, QuestionDef>, r: () => number, flip: number): Record<string, Answer> {
  const text = `${state.title} ${state.excerpt}`;
  const q = questions as Record<string, any>;
  let story_type =
    RX.incident.test(text) ? "incident_crisis" : RX.regulation.test(text) ? "regulation_policy" : RX.funding.test(text) ? "funding_deal" :
    RX.personnel.test(text) ? "personnel_move" : RX.data.test(text) ? "data_report" : RX.opinion.test(state.title) ? "opinion_analysis" :
    RX.announce.test(text) ? "announcement" : r() < 0.5 ? "breaking_event" : "trend_feature";
  let desk = DESK_BY_FEED[state.feed] ?? "business";
  if (/bank|rates|inflation|crypto|payment|stock|market/i.test(text)) desk = "finance";
  else if (/netflix|film|music|celebrity|royal|nfl|nba|game\b/i.test(text)) desk = "culture";
  else if (/food|grocery|travel|retail|shopping/i.test(text)) desk = "consumer";
  const opts = (k: string) => Object.keys(q[k].criteria);
  story_type = maybeFlip(r, flip, story_type, opts("story_type"));
  desk = maybeFlip(r, flip, desk, opts("desk"));

  const big = RX.big.test(text);
  const breaking = story_type === "incident_crisis" || story_type === "breaking_event";
  const slow = story_type === "trend_feature" || story_type === "opinion_analysis";
  const magnitude = jitter(r, (big ? 2 : 0) + r() * 2.6 + (/breaking|live updates|historic|biggest/i.test(text) ? 1 : 0), flip);
  const velocity = jitter(r, (breaking ? 2 : slow ? 0 : 1) + (big ? 1 : 0) + r() * 1.6, flip);
  const novelty = jitter(r, (/first|unprecedented|never|record|new/i.test(text) ? 2 : 1) + r() * 1.8 - (/again|another|latest/i.test(text) ? 1 : 0), flip);
  const window = jitter(r, (breaking ? 3 : slow ? 0.5 : 1.5) + r() * 1.2, flip);
  const heat = jitter(r, (/trump|iran|war|immigra|abortion|gun|union|strike|lawsuit|ban/i.test(text) ? 2 : 0) + r() * 1.8, flip);
  const risk = jitter(r, RX.safety.test(text) ? 3 + r() * 1.2 : /lawsuit|charged|guilty|arrest|abuse/i.test(text) ? 2 + r() : r() * 1.4, flip);
  return {
    is_news: noul(/^(how|why|the best|top \d|\d+ ways)/i.test(state.title) ? 0.25 + r() * 0.3 : 0.85 + r() * 0.14),
    desk: choiceAnswer(q.desk, desk, r),
    story_type: choiceAnswer(q.story_type, story_type, r),
    magnitude: scoreAnswer(q.magnitude, magnitude, r),
    velocity: scoreAnswer(q.velocity, velocity, r),
    novelty: scoreAnswer(q.novelty, novelty, r),
    window: scoreAnswer(q.window, window, r),
    heat: scoreAnswer(q.heat, heat, r),
    risk: scoreAnswer(q.risk, risk, r),
  };
}

function layerB(state: { headline: HeadlineState; a: Record<string, number>; clients: ClientState[] }, questions: Record<string, QuestionDef>, r: () => number, flip: number): Record<string, Answer> {
  const text = `${state.headline.title} ${state.headline.excerpt}`.toLowerCase();
  const mag = state.a.magnitude ?? 1;
  const risk = state.a.risk ?? 0;
  const heat = state.a.heat ?? 0;
  const out: Record<string, Answer> = {};
  const q = questions as Record<string, any>;
  for (const c of state.clients) {
    const word = (t: string) => new RegExp(`\\b${t.replace(/[.*+?^${}()|[\]\\]/g, "\\$&")}s?\\b`, "i").test(text);
    const hits = c.topics.filter(word).length;
    const offPolicy = c.exclusions.some((e) => e.split(" ").every(word));
    const named = word(c.id) || word(c.name.split(" ")[0]);
    const standing = jitter(r, named ? 4 : hits >= 3 ? 3 + r() * 0.6 : hits === 2 ? 2 + r() * 1.2 : hits === 1 ? 0.8 + r() * 1.1 : r() * 0.7, flip * 0.6);
    const shape = jitter(r, standing >= 3 ? 2.4 + r() * 1.6 : standing === 2 ? 1.2 + r() * 1.6 : r() * 1.4, flip * 0.6);
    let tier: "pitch_ready" | "big_story" | "watch" =
      offPolicy || risk >= 4 ? "watch" :
      standing >= 3 && shape >= 2 ? "pitch_ready" :
      standing === 2 && mag >= 3 && shape >= 3 ? "pitch_ready" :
      standing === 2 && mag >= 2 ? "big_story" :
      standing === 1 && mag >= 3 && r() < 0.3 ? "big_story" :
      standing === 0 && mag >= 4 && r() < 0.06 ? "big_story" : "watch";
    tier = maybeFlip(r, flip * 0.3, tier, ["pitch_ready", "big_story", "watch"]);
    if (standing <= 1 && tier === "pitch_ready") tier = "big_story";
    const action = risk >= 4 ? "avoid" : tier === "pitch_ready" ? (heat >= 4 ? "wait" : "ride") : tier === "big_story" ? "wait" : "skip";
    const bridge = tier === "watch" ? "none" : choose(r, { expert_reaction: 4, data_contrast: 2, contrarian: 1.5, explainer: 2, customer_proof: 1, none: 0.3 });
    const conf = hits === 0 && !named ? 0.8 + r() * 0.19 : hits === 1 ? 0.35 + r() * 0.4 : 0.6 + r() * 0.35;
    const p = (k: string) => `${c.id}.${k}`;
    out[p("standing")] = scoreAnswer(q[p("standing")], standing, r);
    out[p("journalist_shape")] = scoreAnswer(q[p("journalist_shape")], shape, r);
    out[p("bridge_type")] = choiceAnswer(q[p("bridge_type")], bridge, r);
    out[p("tier")] = choiceAnswer(q[p("tier")], tier, r, conf);
    out[p("action")] = choiceAnswer(q[p("action")], action, r);
    out[p("off_policy")] = noul(offPolicy ? 0.85 + r() * 0.14 : r() * 0.15);
  }
  return out;
}

export function makeMockAdapter(o: MockOpts): Adapter {
  return {
    id: o.id,
    label: o.label,
    vendor: o.vendor,
    logo: o.logo,
    concurrency: o.concurrency,
    pricing: o.pricing,
    async judge(state, questions, signal) {
      const isB = "clients" in (state as any);
      const seedKey = isB ? (state as any).headline.id + "B" : (state as any).id + "A";
      const r = rng(hash(seedKey + o.id));
      const [lo, hi] = isB ? o.latencyB : o.latency;
      const ms = (lo + r() * (hi - lo)) / Math.max(0.01, o.speed());
      const t0 = performance.now();
      await sleep(ms, signal);
      const answers = isB ? layerB(state as any, questions, r, o.flipRate) : layerA(state as any, questions, r, o.flipRate);
      const inputChars = JSON.stringify(state).length + JSON.stringify(questions).length;
      const nq = Object.keys(questions).length;
      return {
        answers,
        usage: { input_tokens: Math.round(inputChars / 4), output_tokens: o.outputTokensPerQuestion * nq },
        latency_ms: Math.round(performance.now() - t0),
      };
    },
  };
}
