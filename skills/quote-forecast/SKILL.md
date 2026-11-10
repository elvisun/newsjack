---
name: quote-forecast
description: "Forecast how an upcoming scheduled news moment (a keynote, earnings call, product launch, ruling, or conference) could play out, and pre-draft a spokesperson quote for each likely scenario so the client can clear them before the day. Reads what the crowd expects on social media, verifies the facts any quote would rest on, then writes a scenario brief with estimated odds, angles, quotes, risks, and a clearance timeline. Never sends anything."
when_to_use: "User says \"quote forecast,\" \"what will happen at [event],\" \"pre-draft quotes for [event],\" or \"get us ready for [keynote/earnings/launch]\"; or a dated moment from pr-calendar is a week or more out and the client wants quotes cleared ahead of it. Use reactive-comment or newsjack-detector instead when the news has already broken."
metadata:
  category: Act
---

# Quote Forecast

Plan-ahead newsjacking. The event is on the calendar and hasn't happened yet. The client wants a cleared quote ready for whichever way it goes, so on the day you swap in the specifics and send within the hour.

The output is a scenario brief for the PR team and the client. It never goes to a journalist as written.

## Inputs
- **The client.** You need named spokespeople, how each one talks, what they can speak to, and what they never comment on. Stop and ask if you don't have this. A quote needs a named voice with rules. If the user has writing samples but no voice rules, run `voice-extractor` first.
- **The event or events.** Name and date. One or two events per brief. If the user has no event in mind, `pr-calendar` finds upcoming moments the client has standing to own.
- **Anything already approved.** Existing quotes, messaging docs, approved stats, and which partner or customer names the client allows in public.

## Steps

1. **Client ground rules.** Pull the spokespeople and their voice rules, the topics each can speak to, the companies they can and can't name, and standing bans (for example, no comment on lawsuits). These outrank anything the crowd says.

2. **What the crowd expects.** Search social media (X works best for tech and business events) with 6 to 10 queries, covering the event name and hashtag, the event plus the client's topic words, rumored product names, and pricing or benchmark chatter. Sort by engagement. Note the dominant narrative, the rumors and who started them, and the one or two posts that name a gap nobody is answering. That gap is often the client's opening. Use `news-search` for the press coverage of the same rumors.

3. **Verify the facts a quote rests on.** Run `fact-check` on every price, policy, launch date, or number a quote would state. Check each against a primary source, such as the company's newsroom, documentation, or pricing page. Rumors stay out of quotes until a primary source confirms them. List what you couldn't verify.

4. **Scenarios.** Write three to five per event. Each one gets:
   - a name
   - estimated odds as a percentage, labeled as an estimate
   - what happens
   - an angle, ideally a fresh phrase the client can own
   - the one-sentence line the quotes serve
   - one or two quotes
   - which reporters or outlets to pitch
   - what would kill it

   Always include a hold scenario where the client stays quiet, and a wildcard (a competitor counter-programs, a delay, a name collision).

5. **Quotes.** Write each quote in the spokesperson's voice. Reuse approved lines where they fit and mark them as approved. Every number traces to a verified source or an approved client fact. No quote names a company the client hasn't cleared, comments on a lawsuit, or knocks another company's product.

6. **Risks.** List the client's known sensitivities, any customer or partner relationships with companies at the event, unverified numbers, and topics to stay away from.

7. **Clearance timeline.** Make a dated table with internal review, the send to the client, the client's approval deadline, the event day and how often you'll check the news, and the follow-on moment that gives a second shot.

## Deliverable
One brief per event, saved where the team keeps client work. Use tables for odds and timing, one heading per scenario, and each quote as its own paragraph labeled with the speaker. Note which new quotes should join the client's approved quote bank once cleared.

## Rules
- Nothing goes to the client, a journalist, or a public channel without a person on the team reviewing it.
- Odds are estimates. Label them as such.
- The client's own rules beat the crowd on any conflict.
- On the event day, swap in the real specifics and run `fact-check` again before anything leaves.
- Before pitching any reporter from the brief, run `journalist-fit-check`. Follow `skills/ETHICS.md` and `skills/WHY-NOT-SPAM.md`.
