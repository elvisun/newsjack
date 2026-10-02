import type { QuestionDef } from "./engine/types";

// Every radar axis is a Jev `score` with five ordered levels (0–4). Criteria
// text is lifted from newsworthiness-check and relevance-coarse-filter so the
// demo is auditable against the public skills.

export const AXES = ["magnitude", "velocity", "novelty", "window", "heat", "risk"] as const;
export type Axis = (typeof AXES)[number];

export const LEVELS: Record<Axis, readonly [string, string, string, string, string]> = {
  magnitude: ["marginal", "routine", "significant", "major", "historic"],
  velocity: ["static", "slow", "building", "fast", "viral"],
  novelty: ["recurring", "familiar", "fresh twist", "new", "unprecedented"],
  window: ["month", "week", "24hr", "4hr", "30min"],
  heat: ["cold", "mild", "contested", "polarising", "backlash"],
  risk: ["clean", "delicate", "sensitive", "tragedy-adjacent", "kill switch"],
};

const score = (instructions: string, axis: Axis, notes: readonly string[]): QuestionDef => ({
  type: "score",
  instructions,
  criteria: LEVELS[axis].map((l, i) => `${l}: ${notes[i]}`),
});

export const LAYER_A: Record<string, QuestionDef> = {
  is_news: { type: "noul", instructions: "This is a reported news item, not a product page, listicle, or evergreen SEO page." },
  desk: {
    type: "choice",
    instructions: "Which newsroom desk owns this?",
    criteria: { tech: "Technology, software, AI, devices", business: "Companies, markets, deals, management", finance: "Banks, rates, payments, crypto", health: "Medicine, public health, wellness", science: "Research, space, climate science", policy: "Government, regulation, courts, elections", consumer: "Products, retail, food, travel", culture: "Media, entertainment, sport, celebrity", world: "International affairs, conflict, diplomacy" },
  },
  story_type: {
    type: "choice",
    instructions: "What kind of story is this?",
    criteria: { breaking_event: "Something happened in the last day", announcement: "A launch, plan, or decision was announced", data_report: "A study, survey, earnings, or dataset is the news", regulation_policy: "A law, rule, ruling, or enforcement action", funding_deal: "Funding, acquisition, merger, IPO, or contract", personnel_move: "Hire, departure, or leadership change", opinion_analysis: "Argument or explainer rather than a new fact", trend_feature: "Feature tying together several developments", incident_crisis: "Breach, recall, scandal, accident, or public-safety incident" },
  },
  magnitude: score("How big is this story?", "magnitude", ["niche, few outlets would run it", "ordinary daily coverage", "leads a section, several outlets", "front page, widely followed today", "defines the week, everyone is on it"]),
  velocity: score("How fast is coverage spreading right now?", "velocity", ["no pickup, one outlet", "a few pickups over days", "new outlets joining each hour", "wire-speed pickups, live blogs", "everywhere at once, social amplification"]),
  novelty: score("How new is this compared with what has run before?", "novelty", ["the same story that runs every cycle", "a known thread with a small update", "a familiar subject with a genuinely new angle", "a development nobody had reported", "a first, with no precedent"]),
  window: score("How short is the reaction window? Higher means more urgent.", "window", ["slow-burn theme, a month of runway", "follow-ups land all week", "next-day analysis still lands", "same-day reaction only", "live breaking, instant reaction only"]),
  heat: score("How contested or emotionally charged is the coverage?", "heat", ["nobody is arguing", "mild disagreement", "clear sides forming", "heated, identity-charged", "coverage has turned on the story or its people"]),
  risk: score("How dangerous is it for a brand to attach itself to this story?", "risk", ["no sensitivity", "handle with care", "touches a sensitive group or issue", "adjacent to death, disaster, or litigation", "rides directly on tragedy, hate, or an active crime"]),
};

export const STANDING_LEVELS = ["none", "observer", "adjacent", "category", "direct"] as const;
export const SHAPE_LEVELS = ["no reporter", "stretch", "plausible", "likely", "obvious"] as const;

export const LAYER_B_TEMPLATE: Record<string, QuestionDef> = {
  standing: {
    type: "score",
    instructions: "Does the company have real standing to comment? (newsworthiness-check standing gate)",
    criteria: ["none: no credible reason to be quoted", "observer: can comment as a bystander only", "adjacent: expertise in a neighbouring area", "category: expertise in exactly this category", "direct: first-party data, product, or involvement"],
  },
  journalist_shape: {
    type: "score",
    instructions: "Would a named reporter on a real beat want this company's take?",
    criteria: ["no reporter: nobody covers this pairing", "stretch: a reporter might, with effort", "plausible: fits a beat, needs a hook", "likely: reporters on this beat take such calls", "obvious: this is exactly what the beat wants"],
  },
  bridge_type: {
    type: "choice",
    instructions: "Which angle-generator lens gives the company a way in?",
    criteria: { expert_reaction: "Outside expert reaction to the event", data_contrast: "Own numbers confirm or contradict the story", contrarian: "A defensible counter-position", explainer: "Explain the mechanism behind the news", customer_proof: "A customer story that embodies the trend", none: "No honest bridge" },
  },
  tier: {
    type: "choice",
    instructions: "Newsjack-triage tier.",
    criteria: { pitch_ready: "Fresh, standing exists, a journalist would plausibly want the take", big_story: "Fresh and big, no confirmed standing; surface as a suggestion only", watch: "Not worth acting on now" },
  },
  action: {
    type: "choice",
    instructions: "Newsworthiness-check action.",
    criteria: { ride: "Act now inside the window", wait: "Hold for the next development", skip: "Not worth it", avoid: "Kill switch: touching this would harm the company" },
  },
  off_policy: { type: "noul", instructions: "Pitching this would violate one of the company's stated exclusions." },
};

export const CLIENT_QUESTION_KEYS = Object.keys(LAYER_B_TEMPLATE);

// Weighted newsworthiness (0–10) from the axes, per the newsworthiness-check
// Mode A weights. Standing is client-specific and added in set B.
export function newsworthiness(axes: Record<Axis, number>, standing?: number) {
  const base = axes.magnitude * 0.25 + axes.velocity * 0.25 + axes.novelty * 0.15 + axes.window * 0.15; // max 3.2 of 4
  const withStanding = standing == null ? base / 0.8 : base + standing * 0.2;
  return +Math.min(10, (withStanding / 4) * 10).toFixed(1);
}
