---
name: journalist-fit-check
description: "Gate a pitch against one journalist at a time. Returns fit, soft-fit, no-fit, or unknown using recent byline anchors, decay checks, anti-slop refusals, and specific edits."
when_to_use: "User asks whether a specific journalist is a fit for a pitch, wants a pre-send relevance check, tries to add one journalist to a media list, or asks 'should I pitch this person?'"
---

# Journalist Fit Check

This is the **Journalist Fit Check** skill inside newsjack.sh. It is the gatekeeper. It exists because too many pitches go to the wrong people: irrelevant stories, stale contacts, and mail-merge "personalization" sent to journalists who never asked for it.

It checks **one journalist and one pitch at a time**, and gives you one of four answers:

- **Fit** — pitch this person, the match is real.
- **Soft-fit** — close, but the pitch needs a few specific edits first.
- **No-fit** — wrong person, do not pitch.
- **Unknown** — not enough solid evidence to say yes or no.

Every Fit, Soft-fit, or No-fit answer must point to a real, dated, recent article (a "byline" — a piece with this journalist's name on it) published in the last 90 days. No real recent piece, no confident answer.

This skill is blunt on purpose. It will not soften a no or pad a maybe. A clear, specific answer beats a polite one.

<!-- TODO: Reference skills/ETHICS.md and skills/WHY-NOT-SPAM.md once those doctrine files exist in this tree. -->

## Boundaries

- Do not build media lists.
- Do not rank journalists against each other.
- Do not send anything.
- Do not keep or trust a contact database.
- Do not call something a fit based on an outlet's category, a database tag, a bio, or a vibe.
- Do not make up articles, dates, titles, links, outlets, or social posts.
- Do not say **Soft-fit** when the honest answer is **Unknown**.

If you ask about 20 or 50 journalists, this skill checks them one at a time. Stopping the mass-blast is someone else's job; this skill will never be the thing that greases a batch send.

## Required Inputs

Accept one journalist identifier:

- name + outlet
- profile URL
- recent byline URL
- a beat description (the topic area they cover) is useful context only; a beat on its own cannot identify a specific journalist and cannot produce an answer

Accept one pitch:

- full pitch text
- subject line if present
- body text as written

Accept context:

- `current_time_iso` (today's date and time) is required. Never guess what "now" is from memory.
- `client_or_subject` (who or what the pitch is about) is optional.
- `decay_stage` is optional; it carries over from breaking-news workflows to flag how fast a story is moving.

If today's date and time is missing, the answer is **Unknown** (reason: missing current time).

## Retrieval

Find the journalist's recent work using the best source available:

- `medialyst` (the built-in media database) when you are logged in
- `host-agent-search` (public web search) otherwise
- `cache` only when you explicitly hand over saved article evidence

Look in places that are likely to hold their current work:

- the outlet's author page
- Google News or similar web results
- the journalist's personal site
- their Substack or newsletter archive
- LinkedIn snippets
- their Twitter/X profile or specific posts, when those can be opened

The answer must say which source it used. If no source turns up a named article with a date and a link, the answer is **Unknown**.

## Step-By-Step Flow

### Step 1 - Identify the journalist

Confirm this is a real person you can actually check. If you cannot find an author page, a recent article, a profile, a newsletter, or any public trace of them, stop and return **Unknown** (reason: unresolved).

If all you were given is a beat (a topic area, not a person), return **Unknown** (reason: unresolved) and ask for a named journalist, an outlet, a profile link, or a link to one of their recent articles. If the goal is to find journalists in the first place, that's a job for the `media-list-manager` or `newsjack-detector` skills.

Do not guess from the name alone. A confident "yes" about the wrong person is worse than an honest "I'm not sure."

### Step 2 - Scan the pitch for slop tells

A "slop tell" is a sign the pitch is a template, a bot draft, or generic corporate filler. Before checking fit, scan the pitch against the banned patterns listed in the Rubric section below. In plain terms, watch for:

- Fill-in-the-blank placeholders left in the text, like `{Company Name}`, `[TOPIC]`, or `<<<merge_field>>>`.
- Empty PR buzzwords such as "world-class," "innovative," "best-in-class," "revolutionary," "we are committed to," "we are excited to announce," or "we are thrilled."
- Robotic sentence shapes like "It's not just X, it's Y" or "In today's fast-paced world."
- Hollow greetings like "Hope you're well" or "Hope this finds you well."
- Vague flattery about the journalist's "amazing work" that never names a specific article.

If any clear slop tell shows up, return **Unknown** (reason: slop tells in pitch), and point the user to the `meanest-editor` and `voice-extractor` skills to clean up the draft. A pitch that still reads like a template never gets a Fit.

### Step 3 - Find the anchor article

An "anchor" is one specific recent article by this journalist that proves the match is real. For **Fit** or **Soft-fit**, you must name at least one, with all of:

- the exact title, word for word
- a working link (article or a real social post)
- the real publication date
- published within the last 90 days
- one sentence explaining how it connects to the pitch

If your reasoning leans on phrases like "their recent work," "the outlet covers," "given their beat," or "broadly relevant," you do not have an anchor. Find a real article or return **Unknown**.

For Substack writers and independents, anchor to their current newsletter, personal site, or recent posts. Don't judge them by an old staff job when their newsletter is now the work that matters.

### Step 4 - Check how fresh their work is (decay)

"Decay" means how stale the journalist's most recent work is. Every answer reports this:

- More than 90 days since their last article: stop and return **Unknown** (reason: stale data).
- Between 61 and 90 days: you can still give an answer, but flag a freshness warning.
- 60 days or fewer: no warning needed.

For independents, a newsletter post or an openable thread counts as a recent piece. The 90-day cutoff still applies either way.

### Step 5 - Decide the verdict

Use this ladder. The confidence number (0 to 1) is how sure the match is.

| Verdict | Confidence | What it means |
|---------|------------|---------------|
| **Fit** | 0.80 or higher | The journalist covered this exact angle, company, person, format, or problem in the last 90 days, and the pitch already names that coverage or could be tweaked to it in a minute. Save 0.85+ for an exact-angle match within the last 30 days. |
| **Soft-fit** | 0.55 to 0.80 | A real but indirect connection. They cover the broader topic or a nearby angle, but the pitch needs 1-3 specific edits before it's a true fit. |
| **No-fit** | 0.30 to 0.55 | Their recent work has no believable connection. Wrong beat, wrong outlet, wrong format, or wrong angle. Do not suggest wording fixes. |
| **Unknown** | below 0.30, or a refusal | Can't identify the journalist, evidence is stale, no anchor article, weak search, missing date, or the pitch failed the slop scan. |

There is no road from "broadly on their beat" to **Fit**. Broad topic categories are exactly how mass-blasting happens.

### Step 6 - Write the answer

Keep the "why" to 2-3 sentences:

1. State the verdict.
2. Name the anchor article.
3. Name the gap (why it's not a fit) or the driver (why it is).

For **Soft-fit**, give 1-3 concrete edits. Each edit must name the exact paragraph, sentence, hook, or angle to change, tie it to a specific anchor article, and be doable in under five minutes.

For **No-fit**, offer no edits. The journalist is wrong, not the wording.

For **Fit**, any suggested edits are optional and should be minimal.

## Output Format

Write the answer as a short, readable note a person can act on — not a data dump. Keep it terse. Use this shape:

- **A bold verdict line.** One of: **Fit**, **Soft-fit**, **No-fit**, or **Unknown**. Include the confidence (0 to 1) in parentheses.
- **Why (2-3 sentences).** Name the verdict, the specific anchor article (with its title, date, and link), and the reason it fits or the gap that holds it back. No throat-clearing.
- **What to do next.**
  - For **Soft-fit**: 1-3 specific edits. For each, say what to cut, replace, or add, and which anchor article justifies it.
  - For **No-fit**: no edits — just say plainly that the journalist is wrong for this pitch.
  - For **Fit**: optional, minimal polish notes only.
  - For **Unknown**: a clear remediation telling the user exactly what to do next (supply the date, give a named journalist, clean up the draft, etc.).
- **Freshness note.** The date of the most recent verified article and how many days ago that was. If it's between 61 and 90 days old, flag the freshness warning here.
- **Where this came from.** Which source you used (web search, the media database, or supplied cache) and a short trail: what you checked and which links supplied the anchor articles.

When the answer is a refusal, the verdict is **Unknown** with confidence below 0.30, no anchor articles (unless you need to show a stale one), and a clear next step. The reasons a check can refuse are:

- **Missing current time** — today's date and time wasn't provided.
- **Stale data** — the most recent article is more than 90 days old.
- **Unresolved** — the journalist couldn't be identified.
- **Slop tells in pitch** — the pitch failed the anti-slop scan.
- **Uncertainty above threshold** — no solid anchor article could be found.

## Pushback Rules

- If you're asked to "just call it a fit," decline. There is no override.
- If you say you'll personalize the pitch later, the pitch gets judged as written now. "I'll fix it later" is how spam ends up in inboxes.
- If you give only a broad beat, the answer is **Unknown** (unresolved); ask for a named journalist, an outlet, a profile link, or a link to a recent article.
- If the strongest evidence is that the outlet (not this journalist) covers the topic, the answer is **Unknown**, not **Soft-fit**.
- If the last article is more than 90 days old, the answer is **Unknown** (stale) even if the old beat looked perfect.
- If a breaking-news freshness stage is in play, check whether this journalist actually covers that kind of fast-moving story, not just the calmer everyday beat.

The detailed scoring lives in the Rubric section below; worked examples live in the Examples section below.

## Rubric

This rubric turns the design into concrete checks. The hard gates below always win: if any gate trips, the verdict is **Unknown** and the answer explains why.

### Verdict Ladder

| Verdict | Confidence | Standard |
|---------|------------|----------|
| **Fit** | 0.80 or higher | Exact or near-exact angle match, anchored to recent work. Save 0.85+ for an exact-angle match within the last 30 days, where the pitch already names or cleanly leads to that article. |
| **Soft-fit** | 0.55 to 0.80 | Real but indirect overlap. The journalist covers the broader topic, but the pitch needs 1-3 concrete edits. |
| **No-fit** | 0.30 to 0.55 | The journalist is identified and their work is recent, but the pitch is outside their lane. |
| **Unknown** | below 0.30, or a refusal | Missing date, unidentifiable journalist, stale evidence, slop tells, no anchor article, or evidence you can't trust. |

Most honest calls land between 0.50 and 0.75. If everything keeps coming out at 0.85, the check is flattering the pitch.

### Hard Gates

A "gate" is a hard stop. If a gate trips, the verdict is **Unknown** no matter how good the rest looks.

**Gate 1 - Missing date.** Fails when today's date and time wasn't provided. Result: **Unknown** (missing current time).

**Gate 2 - Can't identify the journalist.** Fails when the journalist can't be tied to a real, current public identity: no author page, profile, recent article, newsletter, personal site, or openable social footprint. A beat description alone is not a person. Ask for a named journalist, an outlet, a profile link, or a link to a recent article. Result: **Unknown** (unresolved).

**Gate 3 - Slop tells in the pitch.** Fails when any clear slop tell appears in the pitch. Don't bless copy that still reads like a template, a bot draft, or corporate filler. Result: **Unknown** (slop tells in pitch).

**Gate 4 - No anchor article.** Fails when a **Fit** or **Soft-fit** can't point to a real, dated, linked article by that journalist. Result: **Unknown** (uncertainty above threshold).

**Gate 5 - Stale work.** Fails when the most recent verifiable article is more than 90 days old as of today. Result: **Unknown** (stale data).

**Gate 6 - Made-up or unchecked anchor.** Fails when an anchor's title, link, or date didn't actually come from the source you searched, or when an anchor's link doesn't appear in your "where this came from" trail. Result: drop that anchor. If nothing real is left, the verdict is **Unknown**.

### Scored Criteria

After the hard gates pass, score the ten things below from 0 to 2. Use the total to calibrate your confidence number — not to overrule your judgment. In general:

- **0** - Missing, false, stale, or generic.
- **1** - Present but weak, indirect, or thinly checked.
- **2** - Specific, recent, cited, and usable.

Top score is 20 points. Rough guide:

| Points | Default verdict range |
|--------|-----------------------|
| 17-20 | **Fit**, if the fit eligibility rules below also pass |
| 12-16 | **Soft-fit** |
| 7-11 | **No-fit**, or a low **Soft-fit**, depending on angle overlap |
| 0-6 | **Unknown**, unless a clean **No-fit** is better supported |

What each criterion measures, and what a 0, 1, or 2 looks like:

| # | Criterion | Score 0 | Score 1 | Score 2 |
|---|-----------|---------|---------|---------|
| 1 | Source trail | No source named, or the trail is vague. | Source named, but the trail doesn't show enough of what was checked. | A clear source (web search, media database, or supplied cache) plus a trail listing what was checked and which links became anchors. |
| 2 | Journalist identity and current role | Identity is ambiguous, misspelled, stale, or the outlet can't be confirmed. | Likely the right person, but their current outlet or role is thinly supported. | Identified to a current outlet, newsletter, profile, or author page. |
| 3 | Anchor article quality | No specific article, no link, no date, or just "recent work" reasoning. | A specific article exists, but one part is weak: uncertain date, indirect link, paraphrased title, or a thin relevance note. | Exact title, working link, a real date within 90 days, and a relevance note tied to the pitch. |
| 4 | Freshness (decay) | Most recent article is over 90 days old, or the freshness note is missing. | Article is 61-90 days old and the freshness warning is present. | Article is 60 days old or newer; the freshness note is complete. |
| 5 | Beat and angle overlap | Recent work contradicts the pitch. Wrong beat, wrong outlet format, wrong audience, or wrong story type. | Broad topic overlap only — they cover nearby issues but not this exact angle, format, person, or problem. | Direct overlap with the exact angle, named person, problem, format, or story type in the pitch. |
| 6 | Pitch-to-anchor connection | The pitch doesn't mention or plausibly connect to the anchor article. | The pitch can be edited into relevance with a small bridge. | The pitch already names the anchor or clearly frames itself around the same gap, question, or problem. |
| 7 | Format fit | The pitch asks for a format this journalist doesn't do (e.g., a product launch to an essayist, a vendor briefing to a columnist, an evergreen pitch to a breaking-news reporter, a listicle angle to an enterprise reporter). | The format could work with a reframe, but the current ask is mismatched. | The pitch format matches how the journalist currently works: reported story, analysis, newsletter item, interview, embargo, data scoop, event invite, or another format you observed. |
| 8 | Confidence calibration | Confidence is inflated, unsupported, or outside the verdict's range. | Confidence roughly matches the verdict but ignores evidence quality. | Confidence reflects recency, directness, number of anchors, source quality, and whether the pitch already connects to the article. |
| 9 | Suggested-edit quality | Suggestions are vague, generic, or amount to "do more research." | Suggestions name the angle but not the exact edit. | For **Soft-fit**, each edit names what to cut, replace, or add and ties it to a specific anchor article. For **No-fit**, there are no suggestions. |
| 10 | No-fit discipline | The answer softens a clear no-fit to avoid conflict. | The answer says no-fit but hedges with needless workarounds. | The answer plainly says the journalist is wrong for this pitch and doesn't relabel the miss as a wording problem. |

### Banned Patterns (Slop Tells)

These are the banned patterns the pitch scan in Step 2 checks for. Read the pitch case-insensitively. Any clear match means the pitch fails the slop scan and the verdict is **Unknown** (slop tells in pitch).

**Leftover placeholders** — fill-in-the-blank text that was never replaced:

- Curly-brace tokens like `{Company Name}` or `{First Name}`.
- All-caps bracket tokens like `[TOPIC]` or `[OUTLET]`.
- Angle-bracket merge fields like `<<<merge_field>>>`.

**Banned PR buzzwords and phrases:**

- "world-class," "innovative," "best-in-class," "revolutionary," "cutting-edge"
- "leading" when used to puff up a provider, platform, company, firm, or solution ("leading provider," "leading platform," etc.)
- "we are committed to"
- "we are excited to announce / share," "we are thrilled to announce / share"
- "game-changer," "game-changing," "unlock value" / "unlocks value" / "unlocking value," "synergy"

**Robotic sentence shapes:**

- "It's not just X, it's Y."
- A short clause followed by an em-dash and "and that's why."
- "In today's fast-paced world" or "In today's rapidly-evolving world."

**Hollow greetings:**

- An opener like "Hi [Name]," immediately followed by "Hope you're well" or "Hope this (email) finds you well."
- "I hope this email / message finds you well."

**Em-dash overuse is a warning, not an automatic fail:** if a short pitch (under ~1,500 characters) uses more than two em-dashes, flag it. But if heavy em-dash use shows up alongside any banned phrase above, the pitch fails the scan.

### Generic Reasoning to Reject

If your own "why" leans on any of these phrases, it usually means you never found a real anchor article:

- "their recent work"
- "the outlet (often) covers"
- "she / he / they often (or tend to, or frequently) cover…"
- "given their beat"
- "broadly relevant"
- "aligns with their interests"

If your reasoning uses phrasing like this and there's no valid anchor article behind it, downgrade a **Fit** or **Soft-fit** to **Unknown**.

### What a Fit Requires

A **Fit** verdict needs all of the following:

- At least one anchor article from the last 30 days.
- Direct topical relevance, not just broad-beat relevance.
- A pitch that already names the article or could be tweaked to it in a minute.
- The journalist's most recent article is 60 days old or newer.
- No slop tells.
- No reasoning that rests only on what the outlet (rather than this journalist) covers.

If any one of these fails, drop down to **Soft-fit** or **Unknown**.

### Handling a No-Fit

For **No-fit**, still cite a recent article — but use the relevance note to explain why that article contradicts the pitch. Offer no edits. A no-fit is a targeting problem, not a writing exercise.

## Examples

Each example shows the input it received, the answer it returned, and the action that follows. The point is not to flatter the pitch. The point is to stop bad targeting before it becomes email.

### Example 1: Clean Fit, Recent Anchor

#### Before

- **Journalist:** Maxwell Zeff, TechCrunch
- **Client/subject:** AgentEval, an open-source agent benchmarking harness
- **Today's date:** May 18, 2026
- **The pitch:** Subject line "Open-source eval framework for agentic AI — built after the Anthropic constitutional AI launch." It opens on Maxwell's Apr. 28 piece about developer adoption of Claude's agent SDK, then offers an open-source eval harness that benchmarks agent tool-use against his "real workflow" tests rather than synthetic ones. Repo opens Tuesday; first look plus two pilot customers (legal and healthcare) with workflow-time data; embargo through 9am ET Tuesday.

#### After

**Fit** (confidence 0.86)

**Why:** Maxwell Zeff's Apr. 28 TechCrunch piece argued that agent tool-use benchmarks fail against real workflows; your pitch ships the exact thing he flagged as missing. The embargo, the repo opening, and the pilot customers give him a usable scoop.

**Anchor article:** "Developers say Claude's agent SDK is the easiest to deploy, hardest to evaluate" — TechCrunch, Apr. 28, 2026 — https://techcrunch.com/2026/04/28/claude-agent-sdk-eval-gap/. It names workflow-time benchmarks as the missing piece, which is exactly what the pitch launches.

**What to do next:** Nothing major — the target is right. Just normal copy polish.

**Freshness:** Most recent article May 16, 2026 (2 days ago). No freshness concern.

**Where this came from:** Web search. Checked the TechCrunch author page and recent results; the anchor link above supplied the match. Other recent bylines found May 3, May 11, and May 16.

#### Action

Send the pitch after normal copy polish. The target is right.

### Example 2: No-Fit, Do Not Rescue It

#### Before

- **Journalist:** Emily Chang (Bloomberg author page: https://www.bloomberg.com/authors/AS0E8x9aH-0w/emily-chang)
- **Today's date:** May 18, 2026
- **The pitch:** Subject line "New B2B accounting SaaS for freelancers." Two sentences: the team just launched an accounting platform for freelancers and would love coverage in Bloomberg.

#### After

**No-fit** (confidence 0.41)

**Why:** Emily Chang's recent Bloomberg work is executive interviews, venture capital, and AI company strategy — not B2B SaaS launches or freelancer tooling. A freelancer-accounting product is outside her lane, and copy edits won't change that.

**Anchor article:** "Inside Anduril's pitch to the Pentagon" — Bloomberg, May 11, 2026 — https://www.bloomberg.com/news/articles/2026/05/11/anduril-pitch. It shows the company-and-executive-strategy level she covers, which contradicts the freelancer-SaaS ask.

**What to do next:** Nothing to fix — she's the wrong target. No edits, because the journalist is wrong, not the wording.

**Freshness:** Most recent article May 15, 2026 (3 days ago). No freshness concern.

**Where this came from:** Web search. Reviewed the Bloomberg author page and her last five visible articles; none touched SMB accounting, freelancer tooling, fintech for individuals, or product reviews.

#### Action

Drop the contact. Do not rewrite this for her.
