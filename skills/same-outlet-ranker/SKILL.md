---
name: same-outlet-ranker
description: "Decide which single journalist at one publication to approach first with a pitch. Fit-checks each candidate through journalist-fit-check, then ranks the survivors on the comparative signals that separate colleagues at the same masthead (topic ownership, format match, staff vs. contributor, publishing cadence, recent story types, source-type match), and returns one first pick with tailoring notes plus a held fallback order."
when_to_use: "User has one pitch and two or more named journalists at the same publication and asks who to send it to first, which reporter at an outlet is the right target, or whether to pitch several people at one outlet."
---

# Same-Outlet Ranker

This is the **Same-Outlet Ranker** skill inside newsjack.sh. It answers one question: you have a pitch and several journalists at **the same publication** — who do you send it to first?

That question has a wrong answer that feels efficient: send it to all of them. Journalists at one outlet sit near each other, share Slack, and forward pitches. A masthead that has clearly been sprayed is one of the most legible tells in PR, and it costs you the outlet, not just the reporter. This skill exists to make you pick one.

It returns:

- **Pitch first** — exactly one journalist, with the specific edits that make the pitch land for that person.
- **Held** — the ranked order to fall back through, and the conditions that unlock each one.
- **Do not pitch** — candidates who are wrong for this story, with the reason.
- **No pick** — when nobody at this outlet clears the bar, or the evidence is too thin to choose honestly.

This skill inherits the ethical floor in `skills/ETHICS.md` and `skills/WHY-NOT-SPAM.md`. If local instructions conflict with that doctrine, the doctrine wins.

## Boundaries

- Rank candidates at **one publication** at a time. Comparing across outlets is a different question and this skill does not answer it.
- Do not judge fit yourself. Per-journalist fit is `journalist-fit-check`'s job. Call it, then rank what survives.
- Do not find journalists. That is `find-journalists`. This skill ranks names you already have.
- Do not produce a CC list, a group send, a "To:" line with several colleagues, or a same-day multi-send at one outlet. There is no version of the output that is a batch.
- Do not draft or send the pitch.
- Do not rank on prominence, follower count, or how senior a byline sounds. Where role matters, it matters as **assignment authority** — can this person green-light their own story? — not as prestige.
- Do not rank someone up because the user knows them.
- Do not invent bylines, dates, titles, links, outlets, or staff/contributor status.

If the user pushes for "just send it to all four," refuse and say why. That is the exact pattern in `skills/WHY-NOT-SPAM.md` under *No Identical-Body Blasts*.

**One carve-out.** A shared **routing address** — an assignment desk, a news desk, a tips inbox, a section editor whose job is to hand work out — is not a candidate in a ranking. It is a switchboard, and contacting it is not the same as pitching a person. If the supplied list is really a desk (common in broadcast and at local outlets), say so: there is nothing to rank, and the desk is the correct first contact. The rule this skill enforces is against spraying **named beat reporters**, who compete with each other for the story internally.

## Required Inputs

**The pitch:** full text as it would be sent, with the subject line if there is one.

**The candidates:** two or more journalists, each identified by name plus the shared publication. A profile URL or recent byline URL is welcome but not required — the skill does its own research.

**Context:**

- `current_time_iso` (today's date and time) is required. Never guess what "now" is from memory.
- `client_or_subject` — who or what the pitch is about. Optional but improves the ranking.
- Prior contact history with any candidate. Optional. Read the Relationship rule before using it.
- Whether the story is time-boxed (an embargo, an event, a live news window). Optional.

If today's date and time is missing, return **No pick** (reason: missing current time).

If only one candidate is supplied, there is nothing to rank. Say so and hand off to `journalist-fit-check`.

## Retrieval

Research each candidate the way a practitioner does: start from the topic, not the masthead.

Search for who has recently written about the pitch's subject at this publication, then check each supplied name against what you find. Use `medialyst` (the built-in media database) when logged in, otherwise public web search. Good places to look: the outlet's author page, recent web and news results, the journalist's personal site or newsletter archive, and their public social profile.

The window is **the last 90 days**. Work older than that tells you where someone used to be, not where they are.

Say which source you used. If a candidate produces no dated, linked, named article inside the window, that candidate is **Unverified** and cannot be ranked first — see Hard Gates.

## Step One — Fit-check every candidate

Run each candidate through `journalist-fit-check` against this pitch, one at a time. Do not reimplement its checks and do not shortcut them because you are comparing several people.

You now have a verdict per candidate: **Fit**, **Soft-fit**, **No-fit**, or **Unknown**.

That verdict is a floor, and it is not negotiable by ranking:

- **No-fit** candidates go straight to *Do not pitch*. They are never ranked, never held as a fallback, and no amount of relative strength promotes them. A No-fit is not "fourth best" — they are wrong for this story.
- **Unknown** candidates go to *Unverified*. They can be mentioned, but they cannot be the first pick.
- Only **Fit** and **Soft-fit** candidates enter the ranking.

If the fit-check flags slop tells in the pitch, stop the whole run. The pitch is not ready for anyone at this outlet. Hand off to `meanest-editor` and `voice-extractor`.

## Step Two — Rank the survivors

Fit-check tells you whether each person is plausible. It deliberately refuses to say who is *best* — that is this skill's job.

Apply these signals in order. Earlier signals outrank later ones. A signal only breaks the tie left by the one above it.

### 1. Verdict tier

A **Fit** beats a **Soft-fit**. Always. A Soft-fit only outranks a Fit if the Fit trips a hard gate below.

### 2. Topic ownership

Between two people on the same beat, prefer the one who owns the **specific angle**, not the broad category. Someone who has written three pieces on warehouse automation owns that topic; a general logistics reporter who mentioned it once does not.

Ask: whose recent work would this pitch be a *continuation* of, rather than a departure from? That person is closer.

The strongest possible version of this signal is prior coverage of the client, a named competitor, or the exact problem the pitch describes. That is the one thing that most reliably separates two otherwise-equal colleagues.

### 3. Format match

The pitch has a shape. So does each candidate's output. A launch pitched to someone who only writes first-person columns is a mismatch even when the topic is perfect; a data report belongs with the person who writes data-driven news, not the profile writer.

Use the form table in `journalist-fit-check` for identification. Here it is a *comparative* question: which of these colleagues writes the form this pitch actually fits?

### 4. Recent story types

Read the last 90 days of each candidate's work and ask what kind of story they keep choosing. Not the topic — the *type*. Someone filing scoops and news breaks wants different material than someone filing weekly analysis.

Prefer the candidate whose recent story types the pitch can actually feed.

### 5. Staff writer vs. contributor

Establish which each candidate is, and say so in the output. It changes both the odds and the value:

- A **staff writer** or full-time reporter is the outlet's real byline. Their coverage carries the outlet's weight, but they are pitched more heavily and often need an editor's buy-in.
- A **contributor**, columnist, or freelancer is usually more reachable and more actively looking for material, but the piece may carry less institutional weight and may sit in a section readers and aggregators treat differently.

Neither wins automatically. Decide on what the pitch is:

- **Hard news** — a launch, a raise, a data drop, anything time-pegged. Rank the contributor **below** an equally-fitting staff writer. A contributor usually cannot green-light a news story on their own, and the piece carries less institutional weight.
- **Expertise, commentary, or evergreen** — rank the contributor **at or above** staff when their column is the story's natural home.

Never treat "contributor" as a soft target for a pitch that would not survive a staff writer's scrutiny. That is how outlets end up with a contributor section nobody trusts.

**Freelancers are a separate track, not a rung on this ladder.** A freelancer is not employed by the outlet: they can take the idea elsewhere, and they still have to win a commission before anything runs. So pitching a freelancer does not consume your approach to that masthead, and it does not burn the staff reporters — but the reverse trap is real. You cannot promise a freelancer an exclusive at a named publication, because they cannot deliver that publication. Treat "I'll pitch it to them" as a lead, never as placed coverage.

### 6. Publishing cadence — tiebreaker only

How often is each candidate publishing right now?

Treat this as a tiebreaker and nothing more. There is no good evidence that high- or low-cadence reporters are likelier to respond, so cadence must never overturn a topic, format, or story-type read above it.

It is also two-sided, and the output must say which side it is reading. Someone publishing several times a week is demonstrably active and has slots to fill — but is also buried and working fast. Someone publishing every other week has more room to consider a pitch, and more time to do something with it, but fewer openings. Read it against the pitch: a live news window favors the reporter who can move today; a story that needs reporting time favors the one with room.

A candidate whose cadence has visibly **dropped to nothing** in the last 90 days may be on leave, between jobs, or moving off the beat. Treat that as a caution, not a ranking.

### 7. Source-type match

Check who each candidate actually quotes. Practitioners or academics? Named executives or independent analysts? If the pitch offers a founder and this candidate only quotes customers and researchers, a colleague who does quote founders is the better first send.

### 8. Relationship — tiebreaker only

A prior relationship separates **two candidates who are already genuine fits**. It does nothing else.

It cannot promote a Soft-fit over a Fit. It cannot rescue a No-fit. It is not a reason to send someone a story outside what they cover.

Knowing a journalist is a reason they might *read* your email. It is not a reason the story is right for them, and spending a real relationship on a story that does not fit is how the relationship stops working. If the person you know is not the best fit here, the output says so plainly and puts them in *Held*.

## Step Three — Write the tailoring notes

The first pick gets specific edits. This is the part that turns a ranking into a sent email.

Give 1–4 edits, each naming what to cut, replace, or add, tied to a specific piece of that journalist's recent work, and doable in a few minutes. If the first pick came out of the fit-check as a **Soft-fit**, its required edits are not optional — carry them through and mark them as blocking.

One asymmetry worth knowing: prior coverage of a **competitor** is among the strongest reasons to *choose* a journalist, and one of the worst things to *put in the pitch*. Rival coverage inside the email argues that someone else already told this story. Use it to rank; never cite it as a selling point.

Do not write tailoring notes for held candidates. Those notes belong to whoever you actually pitch, and writing four versions in advance is the batch this skill exists to prevent.

## Step Four — Set the fallback order and its conditions

Held candidates are ranked, but ranking is not permission. Each one needs a condition that unlocks them.

Default sequencing, unless the user has their own rule:

- Give the first pick a real chance. One follow-up, three to seven days after the original — never same-day. Silence inside that window is not a pass.
- An explicit **pass** unlocks the next candidate immediately. "Not for me" is information; no reply is not. On silence, move only after the single follow-up has also gone unanswered.
- Moving to a colleague works best when something has changed — a new development, a sharper angle, a different form. Re-sending the same pitch down the masthead is the spray this skill is built to stop, just slowed down.
- **The second email must name the first journalist.** This is a rule, not a courtesy. Say plainly that you approached the colleague and did not hear back. Newsrooms discover this anyway, and the discovery is what kills stories that were otherwise alive — a reporter who finds out secondhand can drop a piece they had already agreed to.
- Disclose the fact without framing the person as a downgrade. "I reached out to [name] and didn't hear back, so I wanted to bring it to you" does the job. Never tell someone they were the backup.
- **Stop at two people per publication for one pitch.** If both have passed or gone quiet, the outlet is not interested right now. A third approach converts a targeting problem into a reputation problem. Take it elsewhere or change the story.

**Exclusives.** An exclusive is one story, one outlet, one reporter. Never offer one without an expiry — an open-ended exclusive leaves you no clean way to move without appearing to renege. State the deadline in the offer. Never have two live exclusive offers at one publication at the same time; that is not sequencing, it is a lie. If the first pick passes or the deadline lapses, a colleague may be approached — with the same disclosure as above.

**When the news is time-boxed,** the waiting periods compress to hours. The disclosure requirement does not compress.

## Hard Gates

A gate is a hard stop. If one trips, the outcome is what the gate says, no matter how good the rest looks.

- **Missing date.** Today's date and time was not supplied. → **No pick** (missing current time).
- **Slop tells in the pitch.** The fit-check flagged the draft. → Stop the run. No ranking. → `meanest-editor`.
- **Unverified first pick.** A candidate with no real, dated, linked article inside 90 days can never be ranked first, however promising they look. If nobody is verifiable, → **No pick** (uncertainty above threshold).
- **All No-fit.** Every candidate is wrong for this story. → **No pick**. Do not promote the least-wrong one. Say the outlet is not the target and hand back to `find-journalists`.
- **One candidate only.** Nothing to rank. → hand off to `journalist-fit-check`.
- **More than one first pick.** Never. If two candidates are genuinely inseparable, pick one on the earliest signal that differs, and say the call was close. A tie is not a licence to pitch both.
- **Batch request.** Any request to CC, group-send, or same-day multi-send named beat reporters at this outlet. → Refuse, and offer the ranked single-send instead.
- **Non-editorial candidate.** The person works in ad sales, events, production, or the business side. → Remove them. Pitching across the editorial firewall does not work and marks the sender.
- **Wrong side of the opinion line.** A news pitch aimed at someone who writes only opinion or editorial-page work, or the reverse. → Remove them; this is a role mismatch no edit fixes.
- **Live exclusive elsewhere at this outlet.** An exclusive offer is already outstanding to anyone at this publication. → Stop. Resolve that offer before ranking anyone.
- **Repeat wrong targeting.** The user has already sent this journalist a pitch outside their beat, or has already followed up more than once on a pitch that did not fit. → Do not rank that person first. Say plainly that the account is overdrawn there.
- **A colleague already pitched this outlet.** Someone on the user's own team approached this publication with this story recently. → Stop and resolve internally. Two people from one company pitching one masthead is the same spray with extra steps.

## Pushback Rules

- "Just send it to all of them." — No. Name the cost: colleagues at one outlet see each other's inboxes, and a sprayed masthead is remembered as a sender problem, not a story problem.
- "I know this one, put them first." — Only if they are already a genuine fit. Otherwise they stay in *Held*, and say why.
- "Rank them anyway" when nothing is verifiable — No. An invented ranking is worse than no ranking, because it will be acted on.
- "Can I pitch the second person now, in parallel?" — That is the batch with extra steps. Give the first pick their window.
- "The senior one will just forward it to the right reporter." — Maybe, and maybe they kill it. Being forwarded is not a strategy; pick the person who would write it.
- If the user wants the ranking to justify a decision they have already made, give the honest order anyway.

## Output Format

Write it as a short, readable note. Terse. Use this shape:

- **Pitch first: [Name]** — one line on why they beat the others, naming the signal that decided it, plus their staff/contributor status.
- **The anchor** — one specific recent piece by that person: exact title, real date, working link, and one sentence tying it to the pitch.
- **Tailor it** — 1–4 concrete edits. Mark any as blocking if the first pick was a Soft-fit.
- **Held** — the remaining fits in order, one line each: their strongest signal, and the condition that would unlock them.
- **Do not pitch** — any No-fit candidates, one line each on what is wrong. No edits; the person is wrong, not the wording.
- **Unverified** — any candidate whose recent work could not be confirmed, and what would resolve it.
- **If this goes nowhere** — the sequencing rule in one or two lines, including when to stop pitching this outlet.
- **Where this came from** — which source you used and a short trail: what you checked, which links produced the anchors, and how staff/contributor status was established.

When the outcome is **No pick**, give the reason, the remediation, and no ranking. The reasons are: *missing current time*, *uncertainty above threshold*, *no fit at this outlet*, and *pitch not ready*.

## Quality Bar

Before the answer leaves the agent, it must clear all of these:

- **Single-send** — exactly one first pick. No CC, no batch, no parallel track.
- **Delegated** — every candidate went through `journalist-fit-check`; no fit judgment was invented here.
- **Anchored** — the first pick cites a real article with exact title, working link, and a real date inside 90 days.
- **Comparative** — the reasoning says why this person beat *these specific colleagues*, not why they are good in the abstract. If the explanation would read the same with the other candidates deleted, it is not a ranking.
- **Signal-named** — the output names which signal decided the call, and admits when it was close.
- **Status-checked** — staff vs. contributor is established from evidence and stated, not guessed from the byline.
- **Cadence-read** — where cadence influenced the order, the output says which way it was read and why.
- **Relationship-honest** — a known contact who is not the best fit is placed on fit, and the demotion is stated.
- **Actionable** — the tailoring notes name what to cut, replace, or add, and tie to a real piece.
- **Bounded** — the fallback order carries conditions, not just positions, and says when to stop.

## Hand-Offs

- **Check one journalist properly:** `journalist-fit-check`. This skill calls it; use it directly when you only have one name.
- **Find candidates in the first place:** `find-journalists`.
- **The pitch failed the slop scan:** `meanest-editor`, then `voice-extractor`.
- **No fit at this outlet:** back to `find-journalists` for a different publication.
- **The story itself is thin:** `newsworthiness-check` before spending anyone's attention.
- **Claims need verifying before send:** `fact-check`.

## Examples

Outlets and journalists below are fictional, so the examples illustrate the reasoning without putting invented articles under a real person's name.

### Example 1: The obvious pick loses

#### Before

- **Publication:** The Grocery Ledger (trade)
- **Candidates:** Dana Ruiz, Peter Amoah, Lila Fenn
- **Today's date:** July 24, 2026
- **Client/subject:** ColdHop, a refrigerated last-mile delivery startup
- **The pitch:** ColdHop's cold-chain failure data across 40,000 deliveries, showing spoilage concentrated in the final 90 minutes. Offers the dataset and the COO.
- **User note:** "I've met Dana twice at conferences, start with her?"

#### After

**Pitch first: Peter Amoah** (staff reporter). He beat Dana on topic ownership: three pieces on cold-chain loss in the last 90 days, including one that explicitly asked for spoilage data nobody would share. Dana is a real fit but covers retail media; this is a continuation of Peter's reporting and a departure from hers.

**The anchor:** "Nobody wants to talk about what spoils in the last mile" — The Grocery Ledger, June 30, 2026 — https://grocery-ledger.example/last-mile-spoilage. He named the missing dataset; this pitch is that dataset.

**Tailor it:**
1. Open on his June 30 piece by name and lead with the 90-minute finding — he already framed the question, so answer it in sentence one.
2. Cut the company boilerplate paragraph. He quotes operators and data, not positioning.
3. Offer the raw dataset, not a summary. His last three pieces all built their own charts.

**Held:**
- **Dana Ruiz** (staff, retail media) — genuine fit on the retail-operations edge, and you have a relationship. Unlocked if Peter passes and you reframe toward what spoilage costs retailers. Not first: relationship is a tiebreaker, and topic ownership is not close here.
- **Lila Fenn** (contributor, weekly column) — writes first-person trade commentary. Unlocked only if this becomes an opinion piece from your COO rather than a data story.

**Do not pitch:** none — all three are plausible, they are just not equally right.

**If this goes nowhere:** Give Peter five to seven business days. An explicit pass moves you to Dana sooner; silence does not. If two of the three go quiet, stop and take the data elsewhere.

**Where this came from:** Web search. Checked the Grocery Ledger author pages for all three and their last 90 days of bylines. Staff status confirmed from the outlet's staff page; Lila Fenn's contributor status from her byline note.

#### Action

Pitch Peter. Do not copy Dana, even though you know her.

### Example 2: No pick

#### Before

- **Publication:** Northbeam Quarterly
- **Candidates:** four names from the masthead
- **Today's date:** July 24, 2026
- **The pitch:** a Series A announcement for a developer-tools startup.

#### After

**No pick** — no fit at this outlet.

All four candidates came back No-fit. Northbeam Quarterly's recent output is long-form manufacturing policy analysis; none of the four has written a funding story, a company announcement, or anything on developer tooling in the last 90 days. The closest is a March piece on industrial software procurement, which is outside the window and outside the angle.

**Do not pitch:** all four. The problem is the outlet, not the wording, and no edits change that.

**What to do next:** Take this to `find-journalists` for publications that cover developer-tools funding. If Northbeam matters to you strategically, the way in is a manufacturing-sector customer story, not the raise.

**Where this came from:** Web search. Reviewed the outlet's author pages and the last 90 days of bylines for all four names.

#### Action

Do not pitch this outlet. Rebuild the list.
