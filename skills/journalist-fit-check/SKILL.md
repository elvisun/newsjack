---
name: journalist-fit-check
description: "Gate a pitch against one journalist at a time. Runs the pair through proven media-relations checks (last-10-bylines audit, 90-day topic sweep, beat-vs-angle-vs-one-off, the 5 forms of journalism, source-mirror, stated-preferences, database triangulation) and returns fit, soft-fit, no-fit, or unknown with a real recent anchor and specific edits."
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

Look in places that are likely to hold their current work: the outlet's author page, Google News or similar web results, the journalist's personal site, their Substack or newsletter archive, LinkedIn snippets, and their Twitter/X profile or specific posts when those can be opened.

The answer must say which source it used. If no source turns up a named article with a date and a link, the answer is **Unknown**.

## The Checks — how to judge fit

These are the actual methods media-relations practitioners use to test one journalist against one pitch. They replace "do your research" with specific moves. Run the pitch through them, gather the evidence, and let it drive the verdict. A real **Fit** needs all three of the triangulation axes (relevance, timing, format) to hold; two of three is a **Soft-fit**; a missing relevance axis is a **No-fit**.

A single journalist will light up several checks at once. That convergence is the verdict — not any one check alone.

### 1. The Last-10-Bylines Audit — beat verification

**Mechanic:** Don't trust the outlet section or the database beat label. Pull this journalist's last ~10 bylines and ask what percentage actually covers your topic. Below roughly 80% topic-hit across the recent window means the *targeting* is wrong, not the pitch wording.

**Worked example:** Pitch = a new enterprise password manager. Journalist is tagged "Technology" at a national outlet, but their last 10 bylines are 7 on AI regulation, 2 on chip supply chains, 1 on a Big Tech earnings call. Cybersecurity-product = 0/10 in the recent window. → **No-fit** — right vertical label, wrong actual coverage.

### 2. The 90-Day Topic Sweep — recency / decay gate

**Mechanic:** Start from the byline, not the outlet. Pull the last 90 days of *this journalist's* coverage on your specific topic. Coverage older than ~3 months is a red flag they've moved on; reporters change beats every 12–18 months. One old hit does not establish a current beat.

**Worked example:** Pitch = a fintech fraud-detection launch. The journalist wrote a strong fraud feature — but 5 months ago, and nothing on fraud or fintech since (recent work is all crypto regulation). The fraud piece is now outside the freshness window and the beat has drifted. → **Soft-fit at best** — pitch only with a tight, current peg, or find a fresher fraud byline.

### 3. Beat vs. Angle Preference vs. One-Off — the three-way read

**Mechanic:** Read across 6–12 months of bylines and separate three things. **Standing beat** = the topic recurs repeatedly → safe to pitch. **Angle preference** = within that beat, the slice they keep choosing (covers "AI" but always the labor angle, never infrastructure). **One-off** = the topic appears once with no follow-up → coincidence, not a beat. Repetition is a beat; a consistent lens is an angle preference; a single appearance is a one-off.

**Worked example:** A year of AI coverage is consistently the worker-displacement/labor angle (10+ pieces), with exactly one piece on AI chip exports. Your pitch is an AI data-center cooling product. Beat (AI) matches, but their angle preference is labor, and the chip piece is a one-off — not a hardware beat. → **No-fit on the product angle; Soft-fit only if you reframe to "what AI infrastructure means for jobs."**

### 4. Format Match — the 5 forms of journalism

**Mechanic:** Match your pitch *shape* to the form the writer actually produces. Identify the form from their bylines, then check whether the pitch supplies what that form needs:

| Form | Identify it by | What the pitch must supply |
|---|---|---|
| **Investigative** | Data-heavy, multi-source, evidence trails | A data set, expert source, factual narrative — never a product promo |
| **News reporting** | Straight, timely, who/what/when | A genuine time peg plus facts, fast |
| **Feature** | Narrative, human-centric, longer | A person/story (Who/What/When/How), not a spec sheet |
| **Opinion** | First-person ("I believe"), narrow theme | A first-person POV angle — never a product |
| **Column** | Commentary + reporting, recurring themes | A perspective hook tied to current events — not a launch |
| **Review/roundup** | "Top 20…" comparisons or single-item tests | Honest product access, comparison framing, audience value |

**Worked example:** Pitch = "Our new project-management app launched today," sent to an opinion columnist who writes first-person takes on workplace culture. Product launches are exactly what their non-promotional mandate rejects. → **No-fit** (wrong form). The same app pitched to a review/roundup writer who does "best productivity tools" comparisons = **Fit**.

### 5. Source-Mirror Check — who they trust

**Mechanic:** Read 2–3 recent pieces and catalog who they quote and what sources they lean on — academics vs. practitioners, named executives vs. independent analysts, primary data/interviews vs. vendor reports. Then ask whether the source you're offering matches the type they actually use. Tech reporters especially favor credible/independent voices over marketing ones.

**Worked example:** Pitch offers your VP of Marketing as the expert; the journalist's last three pieces quote only independent academics and named end-users, never vendor marketers. The source type mismatches. → **Soft-fit** — swap the offered source to a data scientist or a customer and it becomes a **Fit**.

### 6. Stated-Preferences Read — the journalist's own words

**Mechanic:** Before judging anything inferred from bylines, check the journalist's *declared* rules: Muck Rack "pitching preferences," bio, X/social bio, a "how to pitch me" page. A stated "I don't cover X" overrides any byline inference. If they say no attachments, exclusives-only, or no funding rounds, those are hard constraints.

**Worked example:** Beat and recency look good, but the Muck Rack profile says "exclusives only, no embargoes, and I don't cover funding rounds." Your pitch is a Series B announcement going to five outlets. Two stated rules broken. → **No-fit by self-declaration**, regardless of beat match.

### 7. Database Triangulation — three axes before a yes

**Mechanic:** Use database signals (Muck Rack and similar) as *evidence inputs*, not the verdict. They surface recent articles, self-declared preferences, geographic focus, and 12-month topic coverage. Triangulate three axes before any **Fit**:

1. **Relevance** — beat + angle match, from bylines.
2. **Timing** — within the 90-day freshness window plus a live peg.
3. **Format** — the form they write matches your pitch shape.

All three = **Fit**. Two of three = **Soft-fit**. Missing relevance = **No-fit**.

**Worked example:** Pitch = a SaaS security data report. Relevance ✓ (5 cybersecurity pieces in 90 days), format ✓ (writes data-driven news), timing ⚠️ (no current peg; last breach cycle was weeks ago). → **Soft-fit** — hold until a news peg, or lead with the data's own novelty as the peg.

### Anti-patterns these checks catch

Right outlet, wrong section (news pitch to the opinion desk). A stale or one-off byline treated as a current beat. A "today only" peg sent to a feature writer who works on weeks-long lead times. A commercial promo sent to a non-promotional columnist. A marketer offered to a reporter who only quotes independent experts. Irrelevance is the single biggest reason journalists block a contact — these checks exist to catch it before the send.

## Step-By-Step Flow

### Step 1 — Identify the journalist

Confirm this is a real person you can actually check (author page, recent article, profile, newsletter, or any public trace). If you can't, stop and return **Unknown** (reason: unresolved). If all you were given is a beat (a topic, not a person), return **Unknown** (reason: unresolved) and ask for a named journalist, an outlet, a profile link, or a recent-article link. Finding journalists in the first place is a job for `find-journalists` or `newsjack-detector`. Do not guess from the name alone — a confident "yes" about the wrong person is worse than an honest "I'm not sure."

### Step 2 — Scan the pitch for slop tells

Run the pitch through the Anti-Slop principle below. If a clear slop tell shows up, return **Unknown** (reason: slop tells in pitch) and point the user to `meanest-editor` and `voice-extractor`. A pitch that still reads like a template never gets a Fit.

### Step 3 — Find the anchor article

An "anchor" is one specific recent article by this journalist that proves the match is real. For **Fit** or **Soft-fit** you must name at least one, with all of: the exact title word for word, a working link (article or real social post), the real publication date, published within the last 90 days, and one sentence connecting it to the pitch.

If your reasoning leans on "their recent work," "the outlet covers," "given their beat," or "broadly relevant," you do not have an anchor — find a real article or return **Unknown**. For Substack writers and independents, anchor to their current newsletter, personal site, or recent posts; don't judge them by an old staff job their newsletter has replaced.

### Step 4 — Check how fresh their work is (decay)

"Decay" is how stale their most recent work is. Every answer reports it:

- More than 90 days since their last article: stop and return **Unknown** (reason: stale data).
- 61–90 days: you can answer, but flag a freshness warning.
- 60 days or fewer: no warning needed.

For independents, a newsletter post or an openable thread counts as a recent piece. The 90-day cutoff still applies either way.

### Step 5 — Decide the verdict

Run the checks above and place the result on this ladder. The confidence number (0 to 1) is how sure the match is. Most honest calls land between 0.50 and 0.75; if everything keeps coming out at 0.85, the check is flattering the pitch.

| Verdict | Confidence | What it means |
|---------|------------|---------------|
| **Fit** | 0.80 or higher | All three triangulation axes hold: the journalist covered this exact angle, company, person, format, or problem in the last 90 days, and the pitch already names that coverage or could be tweaked to it in a minute. Save 0.85+ for an exact-angle match within the last 30 days. |
| **Soft-fit** | 0.55 to 0.80 | A real but indirect connection — two of three axes hold. They cover the broader topic or a nearby angle, but the pitch needs 1–3 specific edits first. |
| **No-fit** | 0.30 to 0.55 | Relevance is missing. Wrong beat, wrong outlet, wrong format, or wrong angle. Do not suggest wording fixes. |
| **Unknown** | below 0.30, or a refusal | Can't identify the journalist, stale evidence, no anchor article, weak search, missing date, or the pitch failed the slop scan. |

There is no road from "broadly on their beat" to **Fit**. Broad topic categories are exactly how mass-blasting happens.

### Step 6 — Write the answer

Keep the "why" to 2–3 sentences: state the verdict, name the anchor article, then name the gap (why it's not a fit) or the driver (why it is). For **Soft-fit**, give 1–3 concrete edits — each naming the exact paragraph, sentence, hook, or angle to change, tied to a specific anchor article, and doable in under five minutes. For **No-fit**, offer no edits: the journalist is wrong, not the wording. For **Fit**, any edits are optional and minimal.

## Hard Gates

A gate is a hard stop. If any gate trips, the verdict is **Unknown** no matter how good the rest looks. These are the safety floor of the skill — they are what stops a confident wrong answer from greasing a bad send.

- **Missing date.** Today's date and time wasn't provided. → **Unknown** (missing current time).
- **Can't identify the journalist.** No author page, profile, recent article, newsletter, personal site, or openable social footprint ties them to a real current identity. A beat description is not a person. → **Unknown** (unresolved).
- **Slop tells in the pitch.** Any clear slop tell appears. Don't bless copy that still reads like a template or bot draft. → **Unknown** (slop tells in pitch).
- **No anchor article.** A **Fit** or **Soft-fit** can't point to a real, dated, linked article by that journalist. → **Unknown** (uncertainty above threshold).
- **Stale work.** The most recent verifiable article is more than 90 days old as of today. → **Unknown** (stale data).
- **Made-up or unchecked anchor.** An anchor's title, link, or date didn't actually come from the source you searched, or its link isn't in your "where this came from" trail. Drop that anchor; if nothing real is left, → **Unknown**.

## Pushback Rules

- If you're asked to "just call it a fit," decline. There is no override.
- If the user says they'll personalize later, the pitch is judged as written now. "I'll fix it later" is how spam reaches inboxes.
- If the strongest evidence is that the *outlet* (not this journalist) covers the topic, the answer is **Unknown**, not **Soft-fit**.
- If the last article is more than 90 days old, the answer is **Unknown** (stale) even if the old beat looked perfect.
- If a breaking-news freshness stage is in play, check whether this journalist actually covers fast-moving stories, not just their calmer everyday beat.

## Anti-Slop

The principle: a pitch must read like something a person wrote for *this* journalist, not a template a marketing team approved. Reject leftover merge fields, hollow buzzwords, robotic sentence shapes, and generic flattery that never names a specific article. Read the pitch case-insensitively; one clear match fails the slop scan and the verdict is **Unknown** (slop tells in pitch).

Representative offenders to catch (not exhaustive — judge by the principle): leftover placeholders like `{Company Name}` / `[TOPIC]` / `<<<merge_field>>>`; buzzwords like "world-class," "best-in-class," "revolutionary," "leading provider," "game-changing," "we are thrilled to announce"; robotic shapes like "It's not just X, it's Y" or "In today's fast-paced world"; and hollow greetings like "Hope this email finds you well." Heavy em-dash use in a short pitch is a warning on its own, and a fail when it rides alongside any of the above.

## Output Format

Write the answer as a short, readable note a person can act on — not a data dump. Keep it terse. Use this shape:

- **A bold verdict line.** One of: **Fit**, **Soft-fit**, **No-fit**, or **Unknown**. Include the confidence (0 to 1) in parentheses.
- **Why (2–3 sentences).** Name the verdict, the specific anchor article (with its title, date, and link), and the reason it fits or the gap that holds it back. No throat-clearing.
- **What to do next.**
  - **Soft-fit:** 1–3 specific edits. For each, say what to cut, replace, or add, and which anchor article justifies it.
  - **No-fit:** no edits — just say plainly that the journalist is wrong for this pitch.
  - **Fit:** optional, minimal polish notes only.
  - **Unknown:** a clear remediation telling the user exactly what to do next (supply the date, give a named journalist, clean up the draft, etc.).
- **Freshness note.** The date of the most recent verified article and how many days ago that was. If it's 61–90 days old, flag the freshness warning here.
- **Where this came from.** Which source you used (web search, the media database, or supplied cache) and a short trail: what you checked and which links supplied the anchor articles.

When the answer is a refusal, the verdict is **Unknown** with confidence below 0.30, no anchor articles (unless you need to show a stale one), and a clear next step. The refusal reasons are:

- **Missing current time** — today's date and time wasn't provided.
- **Stale data** — the most recent article is more than 90 days old.
- **Unresolved** — the journalist couldn't be identified.
- **Slop tells in pitch** — the pitch failed the anti-slop scan.
- **Uncertainty above threshold** — no solid anchor article could be found.

## Quality Bar

Before the answer leaves the agent, it must clear all of these. Any miss means revise or regenerate:

- **Sourced** — names the source used (web search, media database, or cache) and a short trail of what was checked and which links became anchors.
- **Anchored** — every Fit/Soft-fit cites at least one real article with exact title, working link, and a real date inside 90 days, plus a one-line relevance note. No "their recent work" hand-waving standing in for a real piece.
- **Fresh** — reports decay; flags the warning at 61–90 days; refuses as stale past 90.
- **On-beat and on-format** — relevance rests on this journalist's *recent bylines* (not the outlet), distinguishes standing beat from angle preference from one-off, and matches the form they actually write.
- **Calibrated** — the confidence number reflects recency, directness, source quality, and whether the pitch already connects to the anchor; it sits inside the verdict's range.
- **Actionable** — Soft-fit edits each name what to cut, replace, or add and tie to a specific anchor; No-fit offers none.
- **Disciplined** — a clear no-fit is stated plainly, not softened into a wording problem or relabeled as Soft-fit.

## Hand-Offs

- **Clean up a slop-flagged draft:** `meanest-editor` and `voice-extractor`.
- **Find journalists in the first place / build a list:** `find-journalists` or `newsjack-detector`. This skill checks one named person; it does not discover or rank them.

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
