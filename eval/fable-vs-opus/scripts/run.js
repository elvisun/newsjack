export const meta = {
  name: 'fable-vs-opus-angle-study',
  description: 'Blind angle-quality study: Opus 4.8 vs Fable 5 generate story angles with the angle-generator skill; GPT-5.5 (codex exec, meanest-editor persona) judges each pair blind in both orderings',
  phases: [
    { title: 'Load', detail: 'load brands.json' },
    { title: 'Prep', detail: 'write per-brand update.txt' },
    { title: 'Generate', detail: 'one clean-context subagent per (brand x model) runs angle-generator and writes its .md' },
    { title: 'Judge', detail: 'GPT-5.5 via codex exec judges A vs B blind in both orderings' },
  ],
}

// args = { run: "2026-06-13" (required, the runs/<run> folder),
//          brand_ids: [1,2,...] (optional subset),
//          models: ["opus","fable"] (optional) }
let ARGS = args
if (typeof ARGS === 'string') { try { ARGS = JSON.parse(ARGS) } catch (e) { ARGS = {} } }
if (!ARGS || typeof ARGS !== 'object') ARGS = {}
const RUN = ARGS.run
if (!RUN) throw new Error('args.run is required, e.g. {"run":"2026-06-13"}')
const MODELS = (Array.isArray(ARGS.models) && ARGS.models.length === 2) ? ARGS.models : ['opus', 'fable']
const [MODEL_A, MODEL_B] = MODELS  // canonical slots for the on-disk filenames

const EVAL = 'eval/fable-vs-opus'
const RUNDIR = `${EVAL}/runs/${RUN}`

const slug = (s) => s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '')
const pad = (n) => String(n).padStart(2, '0')

const BRANDS_SCHEMA = {
  type: 'object', additionalProperties: false,
  required: ['current_time', 'brands'],
  properties: {
    current_time: { type: 'string' },
    brands: {
      type: 'array',
      items: {
        type: 'object', additionalProperties: false,
        required: ['id', 'name', 'sector', 'size', 'fact'],
        properties: {
          id: { type: 'integer' }, name: { type: 'string' }, sector: { type: 'string' },
          size: { type: 'string' }, fact: { type: 'string' },
        },
      },
    },
  },
}

const VERDICT_SCHEMA = {
  type: 'object', additionalProperties: false,
  required: ['scores', 'verdict_A', 'verdict_B', 'winner', 'margin', 'rationale', 'gaps_in_A', 'gaps_in_B'],
  properties: {
    scores: {
      type: 'object', additionalProperties: false, required: ['A', 'B'],
      properties: { A: { $ref: '#/$defs/dim' }, B: { $ref: '#/$defs/dim' } },
    },
    verdict_A: { type: 'string' }, verdict_B: { type: 'string' },
    winner: { type: 'string' }, margin: { type: 'string' },
    rationale: { type: 'string' },
    gaps_in_A: { type: 'array', items: { type: 'string' } },
    gaps_in_B: { type: 'array', items: { type: 'string' } },
  },
  $defs: {
    dim: {
      type: 'object', additionalProperties: false,
      required: ['news_value', 'distinctness', 'journalist_shape', 'grounding', 'anti_slop', 'proof_rigor', 'usefulness'],
      properties: {
        news_value: { type: 'integer' }, distinctness: { type: 'integer' },
        journalist_shape: { type: 'integer' }, grounding: { type: 'integer' },
        anti_slop: { type: 'integer' }, proof_rigor: { type: 'integer' }, usefulness: { type: 'integer' },
      },
    },
  },
}

phase('Load')
const loaded = await agent(
  `Read the file ${EVAL}/brands.json and return its contents as structured JSON. Copy current_time and every brand's id, name, sector, size, and fact EXACTLY character-for-character — do not summarize, paraphrase, or fix anything.`,
  { label: 'load-brands', phase: 'Load', schema: BRANDS_SCHEMA, model: 'opus' }
)
const CURRENT_TIME = loaded.current_time
let brands = loaded.brands
if (Array.isArray(ARGS.brand_ids) && ARGS.brand_ids.length) {
  const want = new Set(ARGS.brand_ids)
  brands = brands.filter((b) => want.has(b.id))
}
if (!brands.length) throw new Error('no brands selected')
log(`${brands.length} brand(s), models ${MODEL_A} vs ${MODEL_B}, run ${RUN}`)

// Generator prompt builder — the subagent reads the harness + skill, writes its
// angle set to a deterministic path, and returns only that path.
function genPrompt(brand, model, outPath) {
  return `You are running an evaluation generator. Follow ${EVAL}/harness/generator.md exactly and apply skills/angle-generator/SKILL.md faithfully. Read both files before producing anything.

Treat the current time as ground truth for "now". Treat ONLY these facts as known — do not add outside knowledge about ${brand.name}, and do not invent customers, metrics, quotes, journalists, or news pegs not present here.

Company: ${brand.name}
Current time: ${CURRENT_TIME}
Update (the fact block):
"${brand.fact}"

Produce the readable markdown angle list exactly as the skill's Output Format specifies (angles first, then Refused angles, Uncomfortable questions, Next step). Then WRITE that markdown — and nothing else — to the file ${outPath} using the Write tool. Do not print the angles in your reply. Return ONLY the literal string: ok:${outPath}`
}

function judgePrompt(updatePath, aPath, bPath, outPath) {
  return `Run this exact command from the repo root and do nothing else first:

bash ${EVAL}/scripts/judge.sh "${updatePath}" "${aPath}" "${bPath}" "${outPath}"

It invokes codex exec (GPT-5.5) as a blind meanest-editor judge and writes schema-valid JSON to ${outPath}. After it finishes, read ${outPath} and return its exact JSON contents as your structured output. Do not judge anything yourself, do not edit the files, and do not add commentary — you are only running the command and relaying its JSON result.`
}

const records = await pipeline(
  brands,
  // Stage 1 — prep: write the per-brand update.txt (single writer, no race).
  async (brand) => {
    const dir = `${RUNDIR}/brand-${pad(brand.id)}-${slug(brand.name)}`
    const updatePath = `${dir}/update.txt`
    await agent(
      `Write a file to ${updatePath} with EXACTLY this content (no extra text):\n\nCompany: ${brand.name}\nCurrent time: ${CURRENT_TIME}\n\n${brand.fact}\n`,
      { label: `prep:${slug(brand.name)}`, phase: 'Prep' }
    )
    return { dir, updatePath }
  },
  // Stage 2 — generate both models' angle sets in parallel, each writing its file.
  async (prev, brand) => {
    const aPath = `${prev.dir}/${MODEL_A}.md`
    const bPath = `${prev.dir}/${MODEL_B}.md`
    await parallel([
      () => agent(genPrompt(brand, MODEL_A, aPath), { label: `gen:${MODEL_A}:${slug(brand.name)}`, phase: 'Generate', model: MODEL_A }),
      () => agent(genPrompt(brand, MODEL_B, bPath), { label: `gen:${MODEL_B}:${slug(brand.name)}`, phase: 'Generate', model: MODEL_B }),
    ])
    return { ...prev, aPath, bPath }
  },
  // Stage 3 — judge both orderings in parallel (cancels position bias).
  async (prev, brand) => {
    const { dir, updatePath, aPath, bPath } = prev
    // ord1: slot A = MODEL_A, slot B = MODEL_B ; ord2: swapped.
    const o1Out = `${dir}/verdict-ord1-A${MODEL_A}-B${MODEL_B}.json`
    const o2Out = `${dir}/verdict-ord2-A${MODEL_B}-B${MODEL_A}.json`
    const [v1, v2] = await parallel([
      () => agent(judgePrompt(updatePath, aPath, bPath, o1Out), { label: `judge:ord1:${slug(brand.name)}`, phase: 'Judge', schema: VERDICT_SCHEMA }),
      () => agent(judgePrompt(updatePath, bPath, aPath, o2Out), { label: `judge:ord2:${slug(brand.name)}`, phase: 'Judge', schema: VERDICT_SCHEMA }),
    ])
    const recs = []
    if (v1) recs.push({ brand_id: brand.id, brand: brand.name, ordering: 'ord1', A_model: MODEL_A, B_model: MODEL_B, verdict_file: o1Out.replace(`${RUNDIR}/`, ''), verdict: v1 })
    if (v2) recs.push({ brand_id: brand.id, brand: brand.name, ordering: 'ord2', A_model: MODEL_B, B_model: MODEL_A, verdict_file: o2Out.replace(`${RUNDIR}/`, ''), verdict: v2 })
    return recs
  }
)

const flat = records.filter(Boolean).flat()
log(`collected ${flat.length} judgments across ${brands.length} brand(s)`)
return {
  run: RUN,
  study: 'Fable 5 vs Opus 4.8 — story-angle quality, GPT-5.5 (meanest-editor) blind judge',
  judge_model: 'gpt-5.5',
  generator_skill: 'skills/angle-generator',
  judge_skill: 'skills/meanest-editor',
  note: 'One record per judgment. ordering encodes which model sat in slot A vs B; aggregate.py re-anchors to model identity to cancel position bias.',
  records: flat,
}
