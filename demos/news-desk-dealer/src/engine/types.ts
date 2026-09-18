export type Headline = {
  id: string;
  title: string;
  url: string;
  source: string;
  domain: string;
  published_at: string;
  excerpt: string;
  feed: string;
};

// Jev-shaped question definitions. The OpenRouter/Opus adapter will translate
// these into a JSON schema; the Jev adapter passes them straight through.
export type QuestionDef =
  | { type: "choice"; instructions: string; criteria: Record<string, string> }
  | { type: "score"; instructions: string; criteria: string[] }
  | { type: "noul"; instructions: string };

export type Answer = {
  type: "choice" | "score" | "noul";
  choice?: string;
  score?: number;
  noul?: number;
  probabilities?: Record<string, number>;
  confidence?: number;
};

export type Usage = { input_tokens: number; output_tokens: number };

export type JudgeResult = {
  answers: Record<string, Answer>;
  usage: Usage;
  latency_ms: number;
};

export interface Adapter {
  id: string;
  label: string;
  vendor: string;
  logo: string;
  concurrency: number;
  pricing: { input_per_mtok: number; output_per_mtok: number };
  judge(
    state: unknown,
    questions: Record<string, QuestionDef>,
    signal?: AbortSignal,
  ): Promise<JudgeResult>;
}

export type Client = {
  id: string;
  name: string;
  sector: string;
  domain: string;
  blurb: string;
  topics: string[];
  exclusions: string[];
};

export type AStamps = {
  is_news: number;
  desk: string;
  story_type: string;
  axes: Record<"magnitude" | "velocity" | "novelty" | "window" | "heat" | "risk", number>; // 0..4 each
};

export type Fit = {
  clientId: string;
  standing: number; // 0..4, see STANDING_LEVELS
  journalist_shape: number; // 0..4
  tier: "pitch_ready" | "big_story" | "watch";
  action: "ride" | "wait" | "skip" | "avoid";
  bridge_type: string;
  off_policy: number;
  confidence: number;
};
