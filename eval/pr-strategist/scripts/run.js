export const meta = {
  name: 'pr-strategist-pairwise-eval',
  description: 'Blind pairwise eval: skill-generated PR strategy vs expert reference, judged by a veteran-PR-director LLM in both orderings',
  phases: [
    { title: 'Load', detail: 'load gold.json cases for the requested split' },
    { title: 'Generate', detail: 'one independent agent per (case x model) runs the pr-strategist skill on the scenario only' },
    { title: 'Judge', detail: 'blind veteran-PR-director judge scores candidate vs gold in both orderings' },
  ],
}

// args = { split: "train"|"holdout"|"all" (default train), models: ["opus","sonnet"] }
let ARGS = args
if (typeof ARGS === 'string') { try { ARGS = JSON.parse(ARGS) } catch (e) { ARGS = {} } }
if (!ARGS || typeof ARGS !== 'object') ARGS = {}
const SPLIT = (typeof ARGS.split === 'string' && ARGS.split) || 'train'
const models = (Array.isArray(ARGS.models) && ARGS.models) || ['opus', 'sonnet']

const GOLD_SCHEMA = {
  type: 'object', additionalProperties: false,
  required: ['rubric_dimensions', 'cases'],
  properties: {
    rubric_dimensions: {
      type: 'array',
      items: { type: 'object', additionalProperties: true, required: ['key', 'desc'], properties: { key: { type: 'string' }, desc: { type: 'string' } } },
    },
    cases: {
      type: 'array',
      items: {
        type: 'object', additionalProperties: false,
        required: ['id', 'name', 'split', 'scenario', 'reference_strategy', 'must_haves', 'anti_patterns'],
        properties: {
          id: { type: 'integer' }, name: { type: 'string' }, split: { type: 'string' },
          scenario: { type: 'string' }, context_notes: { type: 'string' },
          reference_strategy: { type: 'string' },
          must_haves: { type: 'array', items: { type: 'string' } },
          anti_patterns: { type: 'array', items: { type: 'string' } },
        },
      },
    },
  },
}

phase('Load')
const gold = await agent(
  'Read the file eval/pr-strategist/gold.json and return its contents as structured JSON. Include EVERY case verbatim — copy scenario, reference_strategy, must_haves, anti_patterns EXACTLY character-for-character, do not summarize, paraphrase, shorten, or fix anything. Include the top-level rubric_dimensions (key + desc each). Omit only the per-case "industry", "archetype", and "sources" fields.',
  { label: 'load-gold', phase: 'Load', schema: GOLD_SCHEMA, model: 'opus' }
)
const allCases = gold.cases
const cases = SPLIT === 'all' ? allCases : allCases.filter((c) => c.split === SPLIT)
const rubricDims = JSON.stringify(gold.rubric_dimensions || [], null, 2)
if (!cases.length) throw new Error('no cases for split ' + SPLIT)

const JUDGE_SCHEMA = {
  type: 'object',
  additionalProperties: false,
  required: ['scores', 'winner', 'margin', 'which_is_ai', 'ai_tells', 'gaps_in_A', 'gaps_in_B'],
  properties: {
    scores: {
      type: 'object', additionalProperties: false, required: ['A', 'B'],
      properties: {
        A: dimScore(), B: dimScore(),
      },
    },
    winner: { type: 'string', enum: ['A', 'B', 'tie'] },
    margin: { type: 'string', enum: ['clear', 'slight'] },
    which_is_ai: { type: 'string', enum: ['A', 'B', 'unsure'] },
    ai_tells: { type: 'string' },
    gaps_in_A: { type: 'array', items: { type: 'string' } },
    gaps_in_B: { type: 'array', items: { type: 'string' } },
  },
}
function dimScore() {
  const dims = ['audience_goal', 'positioning', 'news_peg', 'channel_cadence', 'tactics_quality', 'judgment_refusals', 'fit_actionability']
  const props = {}
  for (const d of dims) props[d] = { type: 'integer', minimum: 1, maximum: 5 }
  return { type: 'object', additionalProperties: false, required: dims, properties: props }
}

function execPrompt(c) {
  return [
    'You are a founder\'s PR strategist. A founder has pasted their situation. Produce the PR strategy you would give them.',
    '',
    'METHOD (must follow):',
    '1. Read the file skills/pr-strategist/SKILL.md IN FULL and apply it faithfully — its operating loop, decision tree, gates, archetypes, guardrails, and voice. This skill is the behavior under test; do NOT substitute generic PR advice.',
    '2. Also respect skills/ETHICS.md and skills/WHY-NOT-SPAM.md.',
    '3. You have ONLY the founder text below. You have NOT seen any reference answer, rubric, or grading criteria, and must not ask for them.',
    '4. Respond exactly as to a real founder in ONE turn: opinionated, tailored, actionable. Open on their asset, walk the gates, route to an archetype, offer a menu, give a sequenced plan, refuse the dumb defaults the situation invites.',
    '',
    'OUTPUT: the human-facing markdown strategy only — no meta-commentary, do not mention the skill or that you are being evaluated. Length is whatever the advice needs (~250-600 words typical); do not pad.',
    '',
    '--- FOUNDER SITUATION ---',
    c.scenario,
    c.context_notes ? ('\nAdditional context: ' + c.context_notes) : '',
  ].join('\n')
}

function judgePrompt(c, A, B) {
  return [
    'You are a veteran startup-PR director with 20 years of in-house and agency experience (the bar set by First Round\'s comms guides, Lulu Cheng Meservey\'s "go direct" school, April Dunford on positioning, and a16z\'s startup-PR doctrine).',
    'Two strategies, A and B, were written for the SAME founder situation. One was written by a professional PR strategist / expert team; the other by an AI assistant. You do NOT know which is which; order is randomized.',
    '',
    'Judge ONLY strategic substance — correctness, fit to THIS founder, and whether it is what a top operator would actually advise. IGNORE surface style, length, formatting, tone. A longer or more polished answer is not better; generic best-practices are worse than a sharp, correctly-sequenced, tailored plan. Penalize confident WRONG advice heavily.',
    '',
    'Use the rubric to ground scoring, but do NOT require either answer to match must_haves verbatim — there is a broad range of equally-good plans, and achieving the same strategic outcome a different SOUND way gets full credit. Reward only PR-sound divergence. An anti_pattern is always a real defect.',
    '',
    'RUBRIC DIMENSIONS:',
    rubricDims,
    '',
    'STRATEGIC MUST-HAVES (a strong answer hits these outcomes, not necessarily verbatim):',
    JSON.stringify(c.must_haves, null, 2),
    'ANTI-PATTERNS (a weak/wrong answer does these):',
    JSON.stringify(c.anti_patterns, null, 2),
    '',
    '--- FOUNDER SITUATION ---',
    c.scenario,
    '',
    '--- STRATEGY A ---',
    A,
    '',
    '--- STRATEGY B ---',
    B,
    '',
    'Score each of A and B on all 7 dimensions (1=absent/wrong, 3=competent, 5=expert-grade). Then pick winner (A/B/tie; tie only if neither is meaningfully better advice), margin, your best guess which_is_ai (A/B/unsure, substance-based only), ai_tells (concrete substance tells, or "none"), and gaps_in_A / gaps_in_B (1-4 most important concrete things each got wrong or missed vs expert advice for THIS scenario; name any anti_pattern hit). Output strict JSON per the schema.',
  ].join('\n')
}

const tasks = []
for (const c of cases) for (const m of models) tasks.push({ c, m })

log(`generating ${tasks.length} candidates (${cases.length} cases x ${models.length} models), judging each in 2 orderings`)

const out = await pipeline(
  tasks,
  async (t) => {
    const strat = await agent(execPrompt(t.c), { label: `gen:${t.c.name}:${t.m}`, phase: 'Generate', model: t.m })
    return { ...t, candidate: strat }
  },
  async (s) => {
    if (!s.candidate) return null
    const orderings = [
      { ordering: 'cand_first', candidate_label: 'A', A: s.candidate, B: s.c.reference_strategy },
      { ordering: 'gold_first', candidate_label: 'B', A: s.c.reference_strategy, B: s.candidate },
    ]
    const judged = await parallel(orderings.map((o) => () =>
      agent(judgePrompt(s.c, o.A, o.B), { label: `judge:${s.c.name}:${s.m}:${o.ordering}`, phase: 'Judge', schema: JUDGE_SCHEMA, model: 'opus' })
        .then((j) => ({
          case_id: s.c.id, case_name: s.c.name, split: s.c.split,
          model: s.m, ordering: o.ordering, candidate_label: o.candidate_label,
          candidate_text: s.candidate, judge: j,
        }))
        .catch(() => null)
    ))
    return judged.filter(Boolean)
  }
)

const records = out.filter(Boolean).flat()
log(`done: ${records.length} judgment records`)
return { records }
