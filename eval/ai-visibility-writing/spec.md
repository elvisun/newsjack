# AI visibility writing product specification

## Outcome

Create one public Newsjack ATOM that audits, asks only blocking questions,
suggests, or fact-preservingly rewrites a press release, blog post, or
contributed article for clearer retrieval and reuse by AI answer systems.

The skill must help a human reader first. It must never present a writing edit
as a guaranteed ranking, mention, citation, traffic, or coverage intervention.

## Inputs

Required:

- supplied document text;
- requested behavior when the user has a preference.

Optional:

- target audience and likely queries;
- intended publisher or URL context;
- current time;
- sources and proof for factual claims;
- voice or length constraints.

## Behavioral contract

1. Extract a protected fact ledger before advising or rewriting.
2. Infer or read the audience and likely answer queries.
3. Separate site eligibility/authority constraints from prose opportunities.
4. Apply only relevant evidence-backed factors.
5. Ask at most five questions, and only when an answer changes the work.
6. Rank at most three suggestions with confidence and tradeoffs.
7. Rewrite only on request and only within the factual ceiling.
8. End with a causal caveat and, when useful, a matched repeated test.

## Required output shape

- `## Verdict`
- `## Highest-leverage changes`
- a `Change | Why | Confidence | Tradeoff` table
- `## Blocking questions`
- `## Revision` only when requested and safe
- `## Fact-preservation note`
- `## Measurement caveat`

## Concrete examples

### 1. Press release with adequate facts

Input: A constructed release says Northline measured 18,240 deliveries across
12 Canadian cities from January through March 2026 and found a 14% failed-first-
attempt rate. The user asks for a rewrite.

Expected: Preserve `18,240`, `12`, the geography, period, and `14%`; move the
scoped result forward; make the method easy to find; keep company attribution;
do not claim the result represents every delivery market.

### 2. Blog post with a buried answer

Input: A constructed home-energy article spends four paragraphs introducing
heat pumps before stating the climate and building conditions that affect a
choice. The user asks for suggestions only.

Expected: Recommend a direct scoped answer below a descriptive heading and an
intent-fit comparison, but do not rewrite or invent costs and efficiency data.

### 3. Thin-facts draft

Input: “Acme today launched the world's most advanced fraud platform. It
dramatically improves accuracy and is trusted by leading banks.”

Expected: Ask for the comparison set, metric, method, named/approved customer
proof, and intended audience. Do not strengthen, source-wash, or fully rewrite
the claims.

### 4. Clear but promotional contributed article

Input: An expert article contains sound steps but repeatedly calls one approach
“revolutionary,” “unmatched,” and “the only sensible choice.”

Expected: Preserve the steps, replace unsupported promotion with scoped
language, and note that neutral specificity is a promising rather than
guaranteed visibility lever.

### 5. Numbers and quotations must survive

Input: “On May 14, 2026, Dr. Mina Rao said, ‘The 6.2% result applies only to the
840 adults who completed the 12-week protocol.’”

Expected: Every date, number, named entity, quotation, attribution, population,
and limitation remains exact. Any loss or strengthening is a hard failure.

## Refusals and boundaries

Reject keyword stuffing, hidden text, misleading markup, fabricated FAQs,
citation laundering, doorway pages, altered freshness dates, invented sources,
and universal AI-visibility scores. Treat crawling, indexing, HTML, schema,
internal links, backlinks, site authority, and publishing operations as a
separate workstream.

## Acceptance examples

A good result is specific, short, fact-bounded, query-aware, and candid about
what writing cannot control. A bad result applies every tactic, lengthens the
piece, strips caveats, fabricates authority, or predicts an outcome.
