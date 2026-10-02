import { makeMockAdapter } from "./mock";
import { makeOpenRouterAdapter } from "./openrouter";
import { makeTypeSafeAdapter } from "./typesafe";
import type { Adapter } from "./types";

// List prices per million tokens. Opus matches OpenRouter's anthropic/claude-opus-5
// entry ($5 in, $25 out) as of 2026-09-18; Jev per TypeSafe's launch pricing.
export const PRICING = {
  jev: { input_per_mtok: 0.042, output_per_mtok: 0 },
  opus: { input_per_mtok: 5, output_per_mtok: 25 },
};

export type Engine = "mock" | "live";

// Mock unless VITE_ENGINE=live in .env or ?engine=live in the URL. Live mode
// needs OPENROUTER_API_KEY and TYPESAFE_API_KEY in .env for the dev proxy.
export function currentEngine(): Engine {
  const fromUrl = new URLSearchParams(location.search).get("engine");
  const e = fromUrl ?? (import.meta.env.VITE_ENGINE as string | undefined) ?? "mock";
  return e === "live" ? "live" : "mock";
}

let speed = 1;
export const setSpeed = (s: number) => { speed = s; };

const JEV = { id: "jev", label: "Jev", vendor: "TypeSafe AI · jev-latest", logo: "/logos/typesafe.png", concurrency: 8, pricing: PRICING.jev };
const OPUS = { id: "opus", label: "Claude Opus 5", vendor: "Anthropic · via OpenRouter", logo: "/logos/claude.svg", concurrency: 6, pricing: PRICING.opus };

export function makeAdapters(engine: Engine = "mock"): { jev: Adapter; opus: Adapter } {
  if (engine === "live") {
    return {
      jev: makeTypeSafeAdapter({ ...JEV, model: (import.meta.env.VITE_TYPESAFE_MODEL as string) || "jev-latest" }),
      opus: makeOpenRouterAdapter({ ...OPUS, model: (import.meta.env.VITE_OPENROUTER_MODEL as string) || "anthropic/claude-opus-5" }),
    };
  }
  return {
    jev: makeMockAdapter({ ...JEV, latency: [120, 380], latencyB: [200, 520], outputTokensPerQuestion: 0, flipRate: 0, speed: () => speed }),
    opus: makeMockAdapter({ ...OPUS, latency: [2800, 5500], latencyB: [6000, 12000], outputTokensPerQuestion: 14, flipRate: 0.12, speed: () => speed }),
  };
}
