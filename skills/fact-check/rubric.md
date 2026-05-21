# Fact Check Rubric

Use this rubric to evaluate a `fact-check` output before it leaves the agent.
Every criterion is scored 0-2.

- **0** - Missing, unsafe, or likely to hallucinate.
- **1** - Present but incomplete, vague, or too trusting.
- **2** - Specific, cited, and faithful to the skill.

Total possible: 20 points.

| Points | Verdict |
|--------|---------|
| 18-20 | **ship** |
| 14-17 | **revise** |
| 8-13 | **regenerate** |
| 0-7 | **refuse / ask for better input** |

---

## 1. Claim Extraction Completeness

The output must extract every material factual claim, not just the obvious
statistics.

**Score 0:** Misses named people, title claims, bylines, dates, quoted speech,
or statistics that affect whether the draft is safe.

**Score 1:** Captures major claims but misses secondary claims or merges
separate assertions into one line.

**Score 2:** Lists each material claim separately, in input order, with clear
claim text.

Red flags:

- A sentence contains two facts but only one claim is listed.
- A name is checked but the role attached to that name is not checked.
- "Recent", "first", "largest", or "last week" is ignored.

---

## 2. Independent Verification

Each claim must be verified independently.

**Score 0:** Uses one source to bless a whole paragraph, relies on model
memory, or treats the user's confidence as evidence.

**Score 1:** Searches most claims but lets some adjacent claims inherit
support.

**Score 2:** Each claim has its own evidence decision and notes explain when a
source supports only part of the claim.

Red flags:

- "The company page checks out, so the claims are verified."
- A source proves a company exists but not the stated title, date, or statistic.
- The output verifies a quote from a page that only paraphrases the topic.

---

## 3. Citation Quality

Verified and disputed claims need concrete URLs and credible sources.

**Score 0:** Uses vague citations such as "news reports", "LinkedIn", or
"company website" with no URL.

**Score 1:** URLs are present but weak, secondary, stale, or not clearly tied
to the claim.

**Score 2:** Citations are concrete, source quality is appropriate to the
claim, and source limitations are named.

Red flags:

- Citation is a search result page rather than an underlying source.
- A statistic cites an article that mentions the number but not the original
report.
- A disputed status has no URL showing the contradiction.

---

## 4. Status Discipline

Statuses must use the skill's four-label ladder: `Verified`, `Disputed`,
`Unverifiable`, `Missing source`.

**Score 0:** Invents softer labels, uses "likely true", or marks unsupported
claims as verified.

**Score 1:** Uses the labels but applies them inconsistently.

**Score 2:** Applies the labels conservatively and explains ambiguity in notes.

Red flags:

- "Partially verified" appears as a status instead of a note.
- A no-result search becomes `Verified`.
- A claim without a source becomes `Verified` because it seems plausible.

---

## 5. Missing-Source Handling

Missing sources are first-class failures.

**Score 0:** Ignores source gaps or silently supplies a citation without
warning that the draft lacked one.

**Score 1:** Flags missing citations but buries them in notes.

**Score 2:** Uses `Missing source` when needed and calls source gaps out in
the verdict and warning.

Red flags:

- "Research shows" has no named report and no failure status.
- A statistic is verified by a secondary mention but the original report is
missing and not noted.
- The draft's own missing citation would still leave the reader unable to
audit the claim.

---

## 6. Dispute And Contradiction Handling

Contradictions must be prominent and cannot be softened.

**Score 0:** Contradicted claims are marked "needs review", "unclear", or
otherwise underplayed.

**Score 1:** Contradictions are labeled but the verdict summary does not make
the risk clear.

**Score 2:** `Disputed` claims cite the contradictory source and the verdict
summary says the draft is not safe to send as written.

Red flags:

- Official source contradicts the draft, but a weaker source is used to keep it
  verified.
- The note says "source says X" but the status stays `Verified`.
- The warning block omits a disputed claim.

---

## 7. Recency And Staleness

Recency-sensitive claims must be anchored to `current_time`.

**Score 0:** Treats old evidence as current, or infers the current date from
memory.

**Score 1:** Mentions recency but does not adjust statuses when evidence is
too stale.

**Score 2:** Applies the staleness thresholds and marks title, role, and
"recent" claims conservatively.

Red flags:

- A title claim is verified from a two-year-old bio.
- "Last week" is not checked against the current date.
- "Latest report" is accepted without publication date.

---

## 8. Warning Block

The final section must be `## Warning` and must summarize residual risk.

**Score 0:** No warning block, or the warning is generic.

**Score 1:** Warning exists but misses unresolved claims or stale-source risk.

**Score 2:** Warning names every material unresolved, disputed, missing-source,
or stale risk and says what needs human review.

Red flags:

- Final section has another title.
- "No issues" appears despite unverifiable claims.
- Human review items are vague.

---

## 9. Multi-Agent Readiness

The output should be usable by separate extraction, verification, and
adjudication agents.

**Score 0:** The output is a prose blob with no stable claim structure.

**Score 1:** The numbered list exists but lacks enough detail for another
agent to audit decisions.

**Score 2:** Claim order, status, citations, and notes are structured enough
for a later adjudication pass.

Red flags:

- Evidence is mixed into the verdict summary only.
- Claims are grouped by topic instead of numbered.
- Notes do not distinguish extraction ambiguity from source ambiguity.

---

## 10. Output Contract

The final output must match the required Markdown sections.

**Score 0:** Missing one of the required sections, adds a rewrite, or returns
JSON when Markdown was requested by the skill.

**Score 1:** Has the sections but field names are inconsistent or citations
are hard to audit.

**Score 2:** Uses exactly `## Fact-check verdict`, `## Facts & Citations`, and
`## Warning`, with numbered facts and the required fields.

Red flags:

- The warning is not last.
- Field labels are renamed in a way another agent cannot parse.
- The answer includes unsupported reassurance outside the warning.
