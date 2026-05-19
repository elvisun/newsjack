# Crisis Holding Rubric

Every crisis-holding output is evaluated against this rubric before it is returned. Hard gates block output. Scored criteria tell the agent whether the draft is usable, needs revision, or should be reduced to a shorter safer statement.

Each criterion maps to a section of the source design doc: Intake, Legal counsel gate, Drafting rules, Q&A scaffold, What-not-to-say, Decay, Pushback/refusal patterns, and Sample I/O.

## Hard gates

Block the output and ask for correction when any of these fail.

| Gate | Source section | Fail condition | Required behavior |
|---|---|---|---|
| Missing intake | Inputs / Intake | Any required field is missing. | Ask for missing fields one question at a time. Do not draft. |
| Counsel required | Legal counsel gate | Any auto-fire trigger is present and `legal_status == no_counsel_yet`. | Return the STOP block only, unless `--counsel-review-mode` is set. |
| Unconfirmed fact | Drafting rules | Statement asserts a fact absent from `known_facts`. | Remove the sentence or ask the user to confirm. |
| Unknown asserted | Drafting rules | Statement asserts anything from `unknown_or_unverified`. | Remove it from statements; handle it in Q&A as unknown. |
| Invented commitment | Drafting rules | Statement promises an action, owner, deadline, refund, donation, investigation, notification, or deliverable absent from `actions_taken_so_far` or `actions_committed_to`. | Remove it. |
| Unconsented name | Inputs / Legal counsel gate | Statement names a person not listed in `people_involved` with explicit consent, except the named current spokesperson. | Remove the name or trigger counsel review. |
| Placeholder leak | Rubric / checks | Any publishable output contains `{name}`, `[DATE]`, `[Company]`, `<contact>`, or a similar placeholder. | Refuse the variant or ask for the missing fact. |
| Short-statement slop | Banned phrase list | The short statement contains any banned phrase. | Rewrite before returning. |
| Landmine still live | Sample 3 / landmine_newsjack | User says the offending post is still up. | Stop and tell the user to pull it before drafting. |
| No output contract | Output format | JSON or markdown rendering is missing. | Return both, unless the STOP block applies. |

## Score

Score each criterion 0-2.

- **0** - Broken or missing
- **1** - Present but weak, risky, vague, or incomplete
- **2** - Solid and usable

Total possible: 28 points.

| Points | Verdict | Meaning |
|---|---|---|
| 24-28 | **usable for review** | Tight enough to hand to counsel or the comms lead. Still not legal approval. |
| 18-23 | **revise before review** | Core structure works, but one or more risk surfaces are loose. |
| 10-17 | **reduce to short statement** | Too much exposure. Keep only the short statement and Q&A posture. |
| 0-9 | **do not draft** | Intake, counsel gate, or factual discipline failed. |

## Criteria

### 1. Intake completeness

Source: Inputs / structured prompt.

**Score 0:** Required fields are absent or vague enough that the agent has to infer facts.
**Score 1:** Required fields are present, but `known_facts`, `unknown_or_unverified`, or `actions_taken_so_far` are mixed with aspirations or disputed claims.
**Score 2:** Required fields are present, facts are separated from unknowns, and actions are concrete.

### 2. Legal-counsel gate

Source: Legal counsel gate / legal auto-fire keywords.

**Score 0:** A trigger is missed, softened, or treated as a normal drafting case.
**Score 1:** Gate fires, but the trigger is vague or next steps are generic.
**Score 2:** Gate fires or clears correctly, names the exact trigger and field, and gives the narrow next step.

### 3. Factual containment

Source: Drafting rules / anti-hallucination doctrine.

**Score 0:** Statements include invented facts, unnamed sources, inferred scope, or invented timelines.
**Score 1:** Mostly grounded, but one sentence overreaches or implies more certainty than the intake supports.
**Score 2:** Every factual claim maps cleanly to `known_facts`.

### 4. Unknown handling

Source: Drafting rules / Q&A scaffold.

**Score 0:** Unknowns are asserted, denied, or buried.
**Score 1:** Unknowns are acknowledged, but the language is vague or defensive.
**Score 2:** Unknowns are stated plainly and routed to Q&A with `decline-and-name-why` or `refer-to-counsel`.

### 5. Action and commitment discipline

Source: Drafting rules / Pushback pattern "we need to deny this" / "blame third party".

**Score 0:** Promises, deadlines, investigations, outside firms, regulator notices, refunds, or discipline are invented.
**Score 1:** Actions are real, but owner/window language is too loose or uses "soon."
**Score 2:** Only confirmed actions appear, and commitments preserve the user-provided owner or window.

### 6. Short statement fitness

Source: Short statement structure.

**Score 0:** More than 50 words, contains slop, or tries to litigate the incident.
**Score 1:** Short enough, but lacks either a specific confirmed fact or a specific confirmed action.
**Score 2:** 50 words or fewer, plain, defensible, and usable for inbound press.

### 7. Medium statement fitness

Source: Medium statement structure.

**Score 0:** Reads like a press release, values statement, apology essay, or legal memo.
**Score 1:** Structure is present, but audience priority or unknowns are mishandled.
**Score 2:** Around 120 words, audience-led, action-focused, and free of brand positioning.

### 8. Cautious-legal-pass quality

Source: Cautious-legal-pass rules.

**Score 0:** Merely duplicates the medium statement or adds meaningless hedges.
**Score 1:** Softens some assertions, but misses cause, remediation, third-party, or commitment language.
**Score 2:** Rewrites the medium statement with targeted qualifiers and lists every delta from the medium.

### 9. Q&A scaffold usefulness

Source: Q&A scaffold.

**Score 0:** Provides a generic FAQ or full answers that invent facts.
**Score 1:** Includes relevant questions, but categories or postures are thin.
**Score 2:** 10-20 realistic journalist questions, sorted by category, with posture, rationale, and a defensible holding line.

### 10. What-not-to-say specificity

Source: What-not-to-say list / banned phrase list.

**Score 0:** Lists generic advice or misses banned phrases in the draft.
**Score 1:** Catches obvious phrases, but reasons or rewrites are vague.
**Score 2:** Flags exact risky phrases, explains why each is risky, and gives a recoverable rewrite when one exists.

### 11. Decay discipline

Source: Decay.

**Score 0:** No `issued_at`, `valid_until`, or refresh trigger.
**Score 1:** Decay exists but uses the wrong window for urgency, data-security regulation, or landmine newsjacking.
**Score 2:** Correct window is applied and refresh triggers are concrete.

### 12. Landmine-newsjack handling

Source: Trigger / landmine recovery / Sample 3.

**Score 0:** Mentions product, campaign, donations, values, or a follow-up activation.
**Score 1:** Removes product copy but still lets the brand reframe the moment.
**Score 2:** Pull-post-first if live, short contrite language, no product or narrative recovery, 30-minute decay.

### 13. Voice and anti-slop

Source: Banned in all crisis output / anti-slop doctrine.

**Score 0:** Uses templated sympathy, corporate filler, passive distancing, or AI-signature structure.
**Score 1:** Mostly clean, with one or two phrases that need tightening.
**Score 2:** Narrow, active, direct, and free of banned phrases and performative positivity.

### 14. Output contract

Source: Outputs / Output format.

**Score 0:** Missing JSON, missing markdown, or returns prose around the artifact.
**Score 1:** Both formats exist but fields, order, or empty arrays are inconsistent.
**Score 2:** JSON and markdown match exactly, with word counts, deltas, tables, decay, and refusals.

## Banned phrase checks

Soft-fail and rewrite any statement containing these. Hard-fail if they appear in the short statement.

```text
out of an abundance of caution
isolated incident
our hearts go out
our thoughts and prayers
hearts go out
deeply committed
deeply troubled
deeply concerned
deeply saddened
we take [X] seriously
we are committed to transparency
we are committed to integrity
we are committed to excellence
we are committed to our customers
we are committed to our employees
swiftly
promptly
immediately without a timestamp
robust
comprehensive
industry-leading
best-in-class
world-class
regret any inconvenience
regret any confusion
regret any distress
unfortunate situation
regrettable circumstances
rogue employee
rogue actor
rogue agent
fully cooperating with authorities
external investigation
external review
no comment
this does not reflect our values
moving forward
going forward
It's not just [X], it's [Y]
```

## Legal auto-fire keyword checks

Substring-match case-insensitively across `incident_summary`, `known_facts`, `unknown_or_unverified`, and `regulatory_exposure`.

```text
death | fatal | killed | died | hospitaliz | serious injury | bodily harm
harassment | assault | abuse | discriminat
fraud | theft | embezzl | misappropriat
SEC | FDA | OSHA | FTC | DOJ | EPA | CFPB | EU Commission | regulator
GDPR | CCPA | HIPAA | DPA | data subject | PII | personally identifiable
CSAM | child | minor
ransomware | breach | exfiltrat | leaked
recall | hazard | defect
indict | subpoena | warrant | criminal
immigration | ICE | deport
weapons | defense | export control | export-control
class action | lawsuit | suit | litigation
```
