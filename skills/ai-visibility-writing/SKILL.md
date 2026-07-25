---
name: ai-visibility-writing
description: Audit, question, suggest, or fact-preservingly revise a press release, blog post, contributed article, or expert explainer so AI answer systems can more easily retrieve, understand, quote, and cite its useful information. Use when someone asks for AI visibility, AI search, answer-engine, AEO, GEO, AI Overview, or ChatGPT citation optimization of supplied writing; when they want an evidence-aware pre-publication audit; or when another Newsjack workflow needs a prose-level AI-discoverability pass. Do not use as a technical SEO audit, rank tracker, publishing system, or promise of rankings, mentions, citations, traffic, or coverage.
---

# AI Visibility Writing

Make useful information easier to find and reuse without making the writing
worse for people. Treat AI visibility as a probabilistic outcome with strong
retrieval, authority, query, and platform confounders—not as a property a
rewrite can guarantee.

This skill inherits the ethical floor from `skills/ETHICS.md`. If local
instructions conflict with that doctrine, `skills/ETHICS.md` wins. Follow
`skills/WHY-NOT-SPAM.md` when the document will support media outreach.

## Choose the requested behavior

- **Audit:** diagnose the supplied draft and stop before rewriting.
- **Question:** return only the few answers needed before sound advice is
  possible.
- **Suggest:** rank concrete edits without silently applying them.
- **Rewrite:** apply safe, grounded edits and report what changed.

If the request is ambiguous, audit and suggest. Do not rewrite by default.

## 1. Build the fact ledger first

Before judging style, record the material the output must preserve:

- every number, date, named entity, title, quotation, attribution, comparison,
  qualification, and source relationship;
- the document type, intended publisher, audience, and stated purpose;
- supplied links or evidence, including which claim each source supports;
- claims that are promotional, unverifiable from the supplied material, or
  likely to require `fact-check` before publication.

Treat the ledger as a ceiling. Never add a statistic, testimonial, citation,
link, customer, credential, superlative, causal claim, or other fact merely to
make a passage look authoritative. Preserve uncertainty words such as “may,”
“estimated,” “in this sample,” and “as of.” If a proposed improvement needs new
proof, ask for it instead of writing around the gap.

## 2. Read the audience and likely queries

State the primary human audience and two to four plausible information needs
the piece can honestly answer. Prefer the user's target queries when supplied.
Otherwise infer cautiously from the document and label the inference.

Distinguish:

- the reader's question;
- the answer this document can support;
- the answer the organization wishes it could support but cannot yet prove.

Ask a blocking question only when its answer would change the target query,
the factual ceiling, the recommended intervention, or whether a rewrite is
safe. Group questions by priority and ask no more than five at once. Do not ask
for optional analytics, personas, or keywords merely to appear thorough.

## 3. Separate eligibility from writing

Give two clearly separated diagnoses:

- **Retrieval and authority limits:** indexing, crawl access, snippet
  eligibility, page HTML, internal links, canonicalization, structured data,
  publisher reputation, backlinks, and off-site mentions. These can dominate
  visibility but are outside a prose-only edit.
- **Writing-level opportunities:** relevance, extractable answers, evidence,
  attribution, entity clarity, structure, specificity, and human readability
  in the supplied text.

Do not imply that prose can repair an unindexed page or weak publisher
authority. Google says its ordinary Search eligibility and people-first
practices remain foundational for AI features and that no special AI markup is
required. Treat platform behavior as changeable and cross-engine findings as
non-universal.

## 4. Select only applicable levers

Use the smallest useful set. Label each recommendation `supported`,
`promising`, or `speculative`; the label describes the evidence for the
recommendation, not a prediction for this page.

1. **Add unique, verifiable information — promising.** Prefer firsthand data,
   a defined method, an expert observation, or a real example over a commodity
   summary. Ask for proof when it is missing.
2. **Put a scoped answer near its descriptive heading — promising.** Make the
   first useful sentence answer the likely question directly, then add nuance.
   Do not manufacture certainty or flatten a complex answer.
3. **Tie claims to evidence and attribution — promising.** Name who found what,
   when, in which population or context, and from which supplied source. More
   citations are not automatically better.
4. **Clarify entities and relationships — promising.** On first reference,
   disambiguate organizations, products, people, places, and acronyms when a
   reader could reasonably confuse them.
5. **Use descriptive sections and coherent chunks — promising.** Give each
   section one job. Use headings that name the subject and consequence; avoid
   vague labels such as “Overview” when a precise label is available.
6. **Match format to intent — promising and conditional.** Use steps for a real
   procedure and tables for genuine comparisons. Do not bolt FAQs, tables, or
   question headings onto announcements that do not need them.
7. **Expose legitimate freshness — promising and conditional.** State a real
   publication, update, measurement, or effective date for time-sensitive
   information. Never add, hide, or alter a date to simulate freshness.
8. **Replace promotion with specific, qualified prose — promising.** Trade
   empty superiority claims for the exact supported outcome, scope, and
   limitation. Keep necessary brand voice.
9. **Preserve useful precision — supported human-quality guard.** Simplify
   syntax when it helps, but retain technical terms, caveats, and register the
   audience needs. A lower reading grade is not a universal win.
10. **Remove repetition and checklist padding — supported safety guard.** Use a
    natural term when it is the right term; do not repeat keywords, create
    doorway copy, or expand the document so every lever appears.

For a press release, the strongest intervention may be better proof or a
separate owned analysis page with method, data, and limitations. Say so when
the announcement cannot credibly answer the target query. Do not disguise a
thin announcement as an authoritative explainer with cosmetic headings.

## 5. Audit, ask, suggest, or rewrite

### Audit

Identify the extractable claims, evidence gaps, ambiguous entities, buried
answers, mismatched structure, promotional passages, and applicable levers.
Name the dominant non-writing limitation before recommending prose changes.

### Question

Ask only blockers. Explain in a phrase what each answer would unlock. If a
question only changes polish, continue without it and state the assumption.

### Suggest

Rank no more than three changes by expected usefulness to the intended reader
and likely query. For each, identify the exact passage, edit, evidence label,
reason, and tradeoff. Avoid a universal checklist or fake score.

### Rewrite

Rewrite only when requested and the fact ledger is sufficient. Keep the
document's meaning, voice, document type, and evidentiary ceiling. Do not:

- invent facts or strengthen an unsupported claim;
- convert correlation into causation;
- change who said or found something;
- detach a number from its denominator, period, population, or source;
- remove caveats to create a cleaner answer;
- recommend hidden text, misleading schema, citation laundering, or keyword
  stuffing.

After rewriting, compare the revision against the ledger. If a protected item
changed or disappeared, restore it or stop and disclose the blocker.

## Output

Return readable Markdown in this order:

## Verdict

Two sentences: what the document can credibly become and the dominant limit.

## Highest-leverage changes

List up to three changes. Then show:

| Change | Why | Confidence | Tradeoff |
|---|---|---|---|

Use only `supported`, `promising`, or `speculative` in the Confidence column.

## Blocking questions

Include only when needed. Say `None` when the supplied facts are sufficient.

## Revision

Include only when requested and safe. If blocked, say what proof or decision is
needed; do not emit a knowingly misleading partial rewrite.

## Fact-preservation note

Name protected details retained, any claim intentionally softened, and any
material passage left unchanged because evidence was missing.

## Measurement caveat

State in one sentence that the edits may improve clarity or extractability but
do not guarantee retrieval, mention, citation, ranking, traffic, or coverage.
When useful, propose a before/after test using the same query set, platform,
location, time window, and repeated runs; measure mentions, citations, and
accurate answer use separately.

## Short examples

**Thin release:** If “Acme launches the leading fraud platform” has no metric,
method, customer evidence, or comparison set, ask for proof and do not upgrade
the claim. Recommend an evidence page if the release itself cannot carry it.

**Technical blog:** If a security article defines a result precisely, preserve
its caveats and terminology. Improve the heading and opening answer without
rewriting it to a universal low reading grade.

**Freshness-sensitive article:** If the draft supplies an effective date and
an official source, keep both beside the affected claim. Do not change an old
date or add “updated” without a real update.

## Final quality gate

Before returning, confirm that:

- every suggestion is applicable to this document and target query;
- every stronger claim is supported by supplied or verified evidence;
- every protected ledger item survives the rewrite with its relationships;
- out-of-scope technical work is labeled rather than smuggled into prose advice;
- the result remains useful and natural for the intended human reader;
- no sentence promises an AI visibility outcome.
