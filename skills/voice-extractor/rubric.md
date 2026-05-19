# Voice Extractor Rubric

Every extraction, check, and enforcement pass is evaluated against this rubric. Each criterion is scored 0-2:

- **0** - Missing, unsafe, or contradicted by the source doc.
- **1** - Present but weak, incomplete, or too generic.
- **2** - Solid, specific, and grounded in observed user samples.

Total possible: 32 points.

| Points | Verdict | Meaning |
|--------|---------|---------|
| 28-32 | **ship** | Fingerprint/check is usable with normal enforcement. |
| 20-27 | **ship with warnings** | Usable, but confidence or sample quality limits enforcement. |
| 10-19 | **rework** | Too much is generic, unconfirmed, or weakly supported. |
| 0-9 | **refuse / restart** | Unsafe, under-sampled, AI-heavy, or not a real user voice. |

---

## 1. Sample Sufficiency

**Trace:** "voice init - Inputs"; "Mode: extract / Phase 1 - Sample intake"; "Refusal 1 - Insufficient samples."

**Score 0:** Fewer than 5 samples, or sample metadata is absent.
**Score 1:** 5+ samples but under 800 total words, stale samples dominate, or source/audience/date metadata is thin.
**Score 2:** 5-20 samples with source, date, audience, and enough recent native writing to support a fingerprint.

Hard rule: fewer than 5 samples is a refusal.

---

## 2. AI-Heavy Corpus Triage

**Trace:** "It refuses to"; "Phase 2 - Sample triage"; "Refusal 2 - AI-heavy sample set."

**Score 0:** Extracts from a corpus where more than 30% of samples look AI-generated without stopping.
**Score 1:** Flags AI-heavy samples but does not clearly lower confidence or ask for replacement samples.
**Score 2:** Stops above the 30% threshold and asks for non-AI samples, or proceeds only with explicit low-confidence consent.

Signals: em-dash saturation, "not just X, it's Y", "in today's [adjective] world", tricolons, banned-word density, no typos or fragments.

---

## 3. Register Integrity

**Trace:** "It refuses to"; "Trigger"; "Phase 2 - Sample triage"; "Refusal 5 - Cross-register dump."

**Score 0:** Averages incompatible voices into one fingerprint.
**Score 1:** Notices register splits but leaves the profile ambiguous.
**Score 2:** Captures one clear register or creates separate profiles after user confirmation.

Examples: Slack DMs, founder tweets, journalist pitches, board updates, and earnings boilerplate are not automatically the same voice.

---

## 4. Consent And Identity Boundary

**Trace:** "What it is"; "Refusal 4 - Voice-stealing"; "Risks."

**Score 0:** Builds a fingerprint of a non-participating person from public writing.
**Score 1:** Warns but still proceeds without consent.
**Score 2:** Refuses non-consensual third-party fingerprints and allows ghostwriting only when the person is in the loop.

Voice is a signature. Treat it like one.

---

## 5. Local Storage And Privacy

**Trace:** "What it is"; "Layer + rationale"; "`voice init` - Outputs"; "Junior-engineer implementation hints."

**Score 0:** Sends the fingerprint to Medialyst by default or stores raw sample text inside `voice.yaml`.
**Score 1:** Stores locally but leaks sample text or misses audit hashes.
**Score 2:** Writes `~/.newsjack/voice/<profile_id>.yaml`, keeps raw text in sample files, stores hashes and metadata, and points `active.yaml` at the active profile.

Required fields: `schema_version`, `profile_id`, timestamps, sample stats, intent, register, cadence, mechanics, openers, closers, sentence initials, idioms, banned words, banned structures, topic signatures, sample index, extraction metadata.

---

## 6. Cadence Extraction

**Trace:** "`voice init` - Outputs / Cadence"; "Mode: extract / Phase 3 - Extraction."

**Score 0:** Uses subjective labels only.
**Score 1:** Captures some rhythm notes but misses comparable numeric metrics.
**Score 2:** Computes sentence length mean, median, p10, p90, stdev; 1-3-word sentence frequency; 35+ word sentence frequency; paragraph length; one-sentence paragraph rate; rhythm signature.

Cadence must be computed from samples, not inferred from the user's job title or industry.

---

## 7. Mechanics Extraction

**Trace:** "`voice init` - Outputs / Mechanics"; "Phase 3 - Extraction"; "Soft pushback - Em-dash override."

**Score 0:** Ignores mechanics or normalizes them to generic grammar.
**Score 1:** Captures obvious mechanics but misses rates or quirks.
**Score 2:** Captures contractions, contraction rate, em-dash usage, Oxford comma, ellipses, exclamations, questions, parenthetical asides, capitalization quirks, and smart quotes.

Em-dash classification is a high-risk field. Confirm it with the user before saving.

---

## 8. Sentence-Initial And Transition Habits

**Trace:** "`voice init` - Outputs / Sentence-initial"; "Prompt scaffolding / Phase 3."

**Score 0:** Allows stock AI transitions by default.
**Score 1:** Captures conjunction starts but does not ban absent AI transitions.
**Score 2:** Measures starts with But/And/So/Or/Yet/Because and bans absent `however`, `furthermore`, `moreover`, `in conclusion`, `in summary`, `imagine if`, and `picture this`.

If the samples do not show the transition, do not let the model borrow it from generic LLM voice.

---

## 9. Idiom And Vocabulary Fingerprint

**Trace:** "`voice init` - Outputs / Idiom + signature phrases"; "Phase 3 - Extraction."

**Score 0:** Produces generic descriptors like "warm, professional, concise."
**Score 1:** Captures a few phrases but does not separate signature words, real hedges, and banned hedges.
**Score 2:** Extracts repeated literal phrases, unusual signature words, hedges the user actually uses, and hedges the user never uses.

Signature phrases are allowed sparingly. They are not a phrasebook to overfit.

---

## 10. Openers And Closers

**Trace:** "`voice init` - Outputs / Openers + closers"; "Phase 4 - User confirmation."

**Score 0:** Leaves stock openers and closers available.
**Score 1:** Lists observed openers or closers but misses banned stock phrases.
**Score 2:** Captures observed clusters and bans stock phrases such as "I hope this email finds you well," "I wanted to reach out," "Looking forward to hearing from you," and "Please don't hesitate to reach out."

Opening and closing patterns must be confirmed with the user.

---

## 11. Banned Words And Structures

**Trace:** "Rubric / checks / banned lists"; "Global banned-word list (v0)"; "Structural patterns."

**Score 0:** Lets global anti-slop words or AI scaffolds pass.
**Score 1:** Applies global bans but does not handle user-specific exceptions from real samples.
**Score 2:** Applies global bans, adds user-specific bans, flags globally banned words that appear in real samples for review, and stores banned structures with ids, patterns, severity, and rationale.

Global bans are a category-level rule. Individual words can be allowed only when real user samples prove they are genuinely theirs.

---

## 12. Topic And Perspective Anchors

**Trace:** "`voice init` - Outputs / Topic + perspective anchors"; "Phase 3 - Extraction."

**Score 0:** Ignores what the user actually writes about and who they write as.
**Score 1:** Names topics but misses perspective rates.
**Score 2:** Captures 3-5 recurring themes plus first-person singular, first-person plural, second-person, and third-person rates.

This keeps the fingerprint from preserving surface style while drifting away from the user's normal point of view.

---

## 13. User Confirmation

**Trace:** "Phase 4 - User confirmation"; "Pushback (soft - argue, but defer)."

**Score 0:** Saves without showing a summary.
**Score 1:** Shows a summary but does not ask about high-risk fields.
**Score 2:** Shows a one-page summary and asks for overrides on em-dashes, openers/closers, idioms, banned words, and register.

Argue when an override makes drafts sound AI-written. Defer when the user confirms.

---

## 14. Decay And Refresh

**Trace:** "Trigger / refresh"; "Phase 5 - Stamp + tell the user about decay."

**Score 0:** No extraction timestamp or refresh signal.
**Score 1:** Stores timestamps but does not surface staleness.
**Score 2:** Stores `last_extracted_at`, sample age stats, and flags refresh at 90 days.

Voice drifts. The skill must name the drift.

---

## 15. Check Mode Precision

**Trace:** "`voice check <file>` - Output"; "Mode: check"; "Hard refusals"; "Cadence drift"; "Vocabulary drift."

**Score 0:** Returns vague critique instead of machine-usable violations.
**Score 1:** Returns rule names but misses spans, severity, or fix hints.
**Score 2:** Returns verdict, pass rate, fingerprint id, violations with rule/match/span/severity/fix hint, stats, and regenerate flag.

Check mode does not editorialize. It reports tells.

---

## 16. Enforce Mode Contract

**Trace:** "`voice enforce` - Contract"; "Mode: enforce"; "Engine-sharing with `meanest-editor`."

**Score 0:** Drafting skills ignore the fingerprint or pass failing drafts silently.
**Score 1:** Drafting skills load the fingerprint but do not retry or warn cleanly.
**Score 2:** Drafting skills inject `<voice_fingerprint>`, run `voice check`, retry block failures up to 2 times, then return with a visible warning if still failing.

`voice-extractor` is the mechanical lint layer. `meanest-editor` is the editorial roast. Keep the boundary clean.

---

## Hard Block Rules

These always block unless a rule explicitly says fingerprint confidence changes severity.

| Rule ID | Pattern / Trigger | Severity | Trace |
|---|---|---:|---|
| `stray-placeholder` | `\{[a-z _]+\}|\[[A-Z_ ]+\]|<<[A-Z_ ]+>>` | block | Hard refusals; check output |
| `banned-word-global` | Exact match against global list | block | Global banned-word list |
| `banned-word-user-specific` | Exact match against profile list | block | `voice.yaml` schema |
| `em_dash_against_fingerprint` | `—` when `em_dash_usage: never` | block | Mechanics; structural patterns |
| `banned-opener` | Banned phrase used as opener | block | Openers |
| `banned-closer` | Banned phrase used as closer | block | Closers |
| `not-just-x-its-y` | `(?i)\bit'?s not just .*?,? it'?s\b` | block | Banned structures |
| `imagine-if-opener` | `^(Imagine if|Picture this|What if I told you)` | block | Banned structures |
| `in-todays-adjective-world` | `(?i)\bin today'?s [a-z-]+ world\b` | block | Banned structures |
| `now-more-than-ever` | `(?i)\bnow more than ever\b` | block | Structural patterns |
| `ever-evolving-landscape` | `(?i)\bever[- ](evolving|changing) (landscape|world|industry)\b` | block | Structural patterns |
| `sentence-starts-with-however` | `(?<=[.!?]\s)However[,\s]` when absent from fingerprint | block | Sentence-initial habits |
| `furthermore-moreover-additionally` | `\b(Furthermore|Moreover|Additionally)\b` when absent from fingerprint | block | Sentence-initial habits |

## Warn Rules

| Rule ID | Trigger | Severity | Trace |
|---|---|---:|---|
| `cadence_mean_drift` | Sentence length mean drifts more than 40% | warn | Cadence drift |
| `cadence_p90_drift` | Sentence length p90 drifts more than 50% | warn | Cadence drift |
| `paragraph_rate_drift` | One-sentence paragraph rate below 50% or above 200% of fingerprint | warn | Cadence drift |
| `first_person_drop` | First-person singular rate drops more than 50% in pitches/social | warn | Cadence drift |
| `contraction_rate_drop` | Contraction rate falls below 50% of fingerprint | warn | Cadence drift |
| `tricolon-three-past-verbs` | More than 1 per 200 words | warn | Structural patterns |
| `three-adjective-noun-stack` | Three adjective stack before a noun | warn | Structural patterns |
| `title-case-mid-sentence` | `[a-z]\s+([A-Z][a-z]+\s+){2,}` excluding proper nouns | warn | Structural patterns |
| `excessive-hedging` | More than 3 of might/could/may/perhaps/possibly/arguably per 200 words | warn | Structural patterns |
| `signature_absence` | Fewer than 2 signature words or phrases in text over 150 words | warn | Vocabulary drift |

Low-confidence fingerprints downgrade warn rules to informational. Hard blocks stay hard.

## Global Banned-Word List

`delve`, `tapestry`, `leverage`, `leveraging`, `robust`, `comprehensive`, `holistic`, `synergy`, `paradigm`, `paradigm-shift`, `ecosystem` when metaphorical, `solutions` as a noun for products, `unlock`, `unleash`, `empower`, `empowering`, `revolutionize`, `transform` when paired with industry, `disrupt` in the consultant sense, `seamless`, `seamlessly`, `frictionless`, `supercharge`, `turbocharge`, `game-changing`, `game-changer`, `world-class`, `best-in-class`, `cutting-edge`, `state-of-the-art`, `next-generation`, `next-gen`, `leading`, `innovative`, `innovation` as a buzzword, `revolutionary`, `groundbreaking`, `paradigm-shifting`, `transformative`, `mission-critical`, `core competency`, `value-add`, `value proposition`, `deliverables`, `actionable insights`, `low-hanging fruit`, `move the needle`, `circle back`, `touch base`, `synergize`, `ideate`, `operationalize`, `we are committed to`, `we pride ourselves on`, `we believe that`, `in today's [X] world`, `in this ever-changing landscape`, `now more than ever`.
