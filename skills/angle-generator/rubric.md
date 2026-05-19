# Angle Generator Rubric

Use this rubric to evaluate an `angle-generator` output before it leaves the agent. Every criterion is scored 0-2.

- **0** - Missing, broken, or actively unsafe.
- **1** - Present but weak, generic, or partially unsupported.
- **2** - Solid, specific, and faithful to the skill.

Total possible: 20 points.

| Points | Verdict |
|--------|---------|
| 18-20 | **ship** |
| 14-17 | **revise** |
| 8-13 | **regenerate** |
| 0-7 | **refuse / ask for better input** |

---

## 1. Input Completeness And Now Anchor

**Source trace:** `Prompt scaffolding > Process`, `Hard rules`, `Decay reasoning`; `Rubric / checks / banned lists > Decay sanity check`.

The output must respect the input contract. `context.current_time` is required; the agent cannot infer "now" from model memory.

**Score 0:** Missing `current_time` is ignored, weak facts are padded into angles, or the output proceeds on generic input like "new UI" without asking questions.

**Score 1:** Anchors on `current_time` but still produces thin angles from thin facts.

**Score 2:** Refuses or narrows the output when facts are insufficient; uses `current_time` and supplied signals as the only basis for urgency.

Red flags:

- "Today" or "this week" with no supplied timestamp.
- Generic facts treated as news.
- Calendar or signal hooks force-fit to unrelated updates.

---

## 2. Fact Traceability / Hallucination Gate

**Source trace:** `Hard rules > You do not invent facts`; `Rubric / checks / banned lists > Hallucination gate`.

Every angle must identify which user-supplied facts it uses. Missing evidence belongs in `required_proof`, not in the headline or rationale.

**Score 0:** Invents statistics, named people, organizations, customer results, market claims, or regulatory details.

**Score 1:** Mostly grounded but includes unsupported context as if factual, or `facts_used` is vague.

**Score 2:** Every substantive claim traces to `update.facts`, provided links, company fields, or explicit signal payloads; missing evidence is cleanly flagged.

Red flags:

- `facts_used` is empty.
- A statistic appears that was not in the input.
- "Research shows" or "analysts say" with no provided source.

---

## 3. Structural Distinctness

**Source trace:** `Hard rules > You enforce structural distinctness`; `Rubric / checks / banned lists > Distinctness check`; `Sample I/O > Example 4`.

The set must contain different story shapes, not rephrasings for different inboxes.

**Score 0:** Multiple angles share the same headline frame, protagonist, story type, and journalist shape.

**Score 1:** Some distinction exists, but the set still contains filler variants or beat-swapped clones.

**Score 2:** Each kept angle has a distinct protagonist, beat, story type, proof path, or timing frame; duplicates are killed and logged.

Red flags:

- "Another angle" is the main differentiator.
- Same `story_type` plus same `beat_description`.
- More than 65% conceptual overlap between headline frames.

---

## 4. Journalist Shape

**Source trace:** `Journalist-shape rubric`; `Hard rules > You do not invent journalists`; `Pains addressed > Pitch volume + irrelevance`.

An angle is not real until a plausible beat can be named. The skill names the shape, not a specific journalist.

**Score 0:** Uses generic targets like "tech journalist" or names specific journalists without verification.

**Score 1:** Beat is present but broad; `evidence_they_care` is generic or could apply to any outlet.

**Score 2:** Beat, outlet archetype, timely reason, and `do_not_target` are specific enough to guide the next skill.

Red flags:

- "This would appeal to journalists who care about startups."
- No `do_not_target`.
- Named journalists appear in the output.

---

## 5. Why-Now And Decay

**Source trace:** `Decay reasoning`; `Hard rules > You tag every angle with decay`; `Rubric / checks / banned lists > Decay sanity check`.

The output must be honest about urgency.

**Score 0:** Claims breaking urgency without a supplied signal, or omits decay.

**Score 1:** Decay is present but generic; `why_now` is a vague trend or repeats the update date.

**Score 2:** `why_now` names the real time hook or says `EVERGREEN, NOT TIME-PRESSURED`; decay matches the source of urgency.

Red flags:

- `30min` or `4hr` with no `signal_from_newsjack_detector`.
- `evergreen` on a company update without an uncomfortable question.
- "In today's market" as the peg.

---

## 6. Anti-Slop Pass

**Source trace:** `Banned words and structures`; `Rubric / checks / banned lists > Anti-slop regex pass`; `Secondary signal - Reads like a bot wrote it`.

The output must reject AI-marketing language before the user sees it.

**Score 0:** Kept angles contain banned terms or press-release framing.

**Score 1:** Mostly clean but one field still leans on puffery or generic phrasing.

**Score 2:** Headline frames, `why_now`, `distinctness_check`, and `evidence_they_care` are concrete and free of banned structures.

Red flags:

- "innovative platform", "future of X", "game-changing", "excited to announce".
- "It's not just X, it's Y."
- Placeholder leftovers such as `[COMPANY]`.

---

## 7. Required Proof

**Source trace:** `What "an angle" means`; `Hard rules > You ask uncomfortable questions`; `Rubric / checks / banned lists > Per-angle proof requirement`.

The skill must show what evidence makes the angle pitchable.

**Score 0:** Proof is absent, generic, or asks for evidence after already making the claim.

**Score 1:** Proof exists but is not specific enough to guide the user.

**Score 2:** Each angle lists concrete proof; `data`, `customer-story`, `contrarian`, and `exec-spotlight` angles have at least one required proof item.

Red flags:

- "Need more data" with no description of what data.
- Contrarian angle with no stated conventional wisdom to challenge.
- Customer story with no customer proof requirement.

---

## 8. Refused Angles

**Source trace:** `Hard rules > You output the refused angles`; `Refusal patterns`; `Sample I/O`.

Refusal is part of the product. The user should see what died and why.

**Score 0:** No `refused_angles` field, or bad angles are kept instead of killed.

**Score 1:** Refused angles are listed but reasons are vague or outside the allowed values.

**Score 2:** Refused angles use allowed reasons and teach the user what not to pitch.

Allowed refusal reasons:

- `duplicate`
- `slop`
- `hallucinated_fact`
- `no_journalist_shape`
- `no_why_now_but_required`
- `off-beat`

---

## 9. Uncomfortable Questions And Next Skill

**Source trace:** `Hard rules > You ask uncomfortable questions`; `When to call other skills`; `Open questions / risks > brand risk of being too refusenik`.

The skill should be tough without stranding the user.

**Score 0:** No questions when proof gaps are obvious, or the output ends without a next move.

**Score 1:** Questions exist but are broad, soft, or disconnected from the angles.

**Score 2:** Questions expose the exact missing facts that determine whether the angles are real; `follow_up_suggestions` names the right next skill or `null`.

Red flags:

- "Can you provide more details?"
- Recommending `meanest-editor` before there is an angle or draft.
- Calling a media-list skill before journalist shapes exist.

---

## 10. Output Contract

**Source trace:** `Output schema`; `Prompt scaffolding > Output format`.

The final output must be machine-usable.

**Score 0:** Prose summary instead of JSON, missing top-level keys, or invalid JSON.

**Score 1:** JSON is valid but fields are missing, renamed, or filled with vague placeholders.

**Score 2:** Valid JSON with `angles`, `refused_angles`, `uncomfortable_questions`, and `follow_up_suggestions`; each angle includes all required fields.

Red flags:

- Preamble such as "Here are your angles."
- Markdown bullets instead of the schema.
- `anti_slop_pass` omitted or set without actual anti-slop compliance.
