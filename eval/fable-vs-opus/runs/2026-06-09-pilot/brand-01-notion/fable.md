**Four of the five largest US banks are piloting Notion after it let them hold the encryption keys**

- **Story type:** customer-story
- **Why a journalist cares:** A banking-technology trade reporter covering what software the biggest US banks actually approve and deploy. The hook is power elite plus surprise: the largest, most security-conservative buyers in the country are piloting a tool with consumer-app DNA, and the unlock was key custody via AWS KMS. Not for consumer productivity press, app-review outlets, or general startup desks — and not for fintech reporters who cover banking products rather than bank IT.
- **Why now:** The CMEK launch is the stated reason the pilots became possible; the announcement is the peg.
- **Decay:** 24hr — standard company-launch news cycle; the bank-pilot hook fades fast once the launch day passes, unless a bank goes on record later.
- **What makes it distinct:** News-values lens, power elite — the banks are the protagonist, not the product. The other angles are about the architecture, the compliance category, and the geography of data.
- **Proof it needs:** at least one of the four banks willing to confirm the pilot, or a credible compliance/procurement source at one; clarity on what "pilot" means (seats, scope, timeline to production); permission to characterize the banks even if unnamed.
- **Facts it rests on:** four of the five largest US banks run pilots; enterprise customers can hold their own encryption keys via AWS KMS; the company says CMEK unblocks regulated buyers in finance.

**Notion ships offline editing, the feature cloud-native software spent a decade treating as optional**

- **Story type:** contrarian
- **Why a journalist cares:** A product or infrastructure reporter at a technically literate outlet who covers software architecture and sync engineering. The prevailing belief: modern collaboration SaaS assumes you are always online, and offline is a legacy concern. The evidence against it: a ~100M-user, cloud-native app investing in offline-first editing with sync on reconnect. Not for enterprise-security desks, banking trade press, or business reporters who don't cover how software is built.
- **Why now:** Offline editing ships now; the engineering-story window is tied to the launch but holds longer than the announcement itself.
- **Decay:** week — architecture and how-it-was-built angles survive the launch day if the engineering detail is real.
- **What makes it distinct:** Inversion lens — the product architecture is the protagonist. Every other kept angle is about compliance and buyers; this one would run even if the banks didn't exist.
- **Proof it needs:** an engineer or product lead who can explain conflict resolution on reconnect; honest scope of offline (full editing vs. limited surfaces, databases, multi-user merge); why the company prioritized this now.
- **Facts it rests on:** offline editing now works with no connection and syncs on reconnect; Notion has ~100M users.

**What it took for a 100M-user productivity app to clear a bank's security review**

- **Story type:** trend
- **Why a journalist cares:** An enterprise-security trade reporter covering cloud security, key management, and what regulated buyers now demand from SaaS vendors. The angle zooms up from the launch: CMEK and residency as the admission price for finance and healthcare, with Notion as the fresh, concrete data point. Not for consumer tech press, founder-profile writers, or general business desks — and not the same inbox as the bank-pilot story, which is about the buyer, not the requirements list.
- **Why now:** The launch is the peg, but the requirements-for-regulated-buyers framing supports a longer reported piece.
- **Decay:** week — trend and context pieces outlive the announcement.
- **What makes it distinct:** Ladder lens, zoomed up — the compliance bar for regulated industries is the story; Notion is the evidence. The bank angle keeps one protagonist at ground level; this one makes the category the protagonist.
- **Proof it needs:** specifics of what regulated buyers required before saying no (CMEK, residency, what else); ideally a second vendor or analyst data point so the trend isn't a one-company claim; a named regulated customer or prospect who can describe the prior blocker.
- **Facts it rests on:** enterprise customers can hold their own encryption keys via AWS KMS; data residency added for the EU and Australia; the company says this unblocks regulated buyers in finance and healthcare that previously could not adopt; ~100M users.

**Data residency comes to the docs app: Notion adds EU and Australia regions**

- **Story type:** defensive-comment
- **Why a journalist cares:** A data-protection or digital-sovereignty reporter at a policy-adjacent or regional enterprise outlet (EU tech-policy desks, Australian enterprise IT trade press) covering where companies are required to keep data. The hook is that residency demands have reached everyday productivity software, not just databases and clouds. Not for US banking press, product reviewers, or security-engineering desks — those are the other angles' inboxes.
- **Why now:** The residency options ship now. There is no live regulatory event in hand to peg this to — the launch itself is the only time hook, and the angle should say so honestly.
- **Decay:** week — regional residency news holds for regional trade press through the week; it strengthens sharply if a real sovereignty news event appears.
- **What makes it distinct:** Perspective lens, regulator/rules protagonist — the geography and the rules are the story. CMEK is about who holds the keys; this is about where the data lives, and it targets a different (and partly non-US) press corps.
- **Proof it needs:** which regulations or customer demands drove the EU and Australia choices specifically; whether residency covers all customer data or a subset; an EU or Australian customer or prospect who can speak to the requirement.
- **Facts it rests on:** data residency options added for the EU and Australia; the company says this unblocks regulated buyers that previously could not adopt.

**Refused angles**

- "Notion takes on Microsoft and Atlassian in the regulated enterprise" — `hallucinated_fact`. No competitor, displacement, or win-loss fact was supplied; a counterposition angle with no named incumbent evidence is invented conflict.
- "Notion hits ~100M users" — `no_journalist_shape`. A bare user number with no comparison, growth rate, or timeframe is a topic, not a story; it works as supporting context inside the other angles, not as a lead.
- "Hospitals can finally run Notion" — `hallucinated_fact`. Healthcare appears only inside the company's own claim about regulated buyers; there is no healthcare customer, pilot, or named system in the facts. The bank pilots are evidence; healthcare is aspiration.
- "Notion refuses the security upsell: CMEK at no extra charge" — `hallucinated_fact`. The fact says no pricing change *for existing Enterprise plans*; whether CMEK is included or an add-on for new buyers is unknown, and the angle collapses without that answer.
- "Notion revolutionizes enterprise collaboration with seamless offline and best-in-class security" — `slop`. Press-release language end to end; rewriting it produces one of the kept angles, so it dies as written.

**Uncomfortable questions**

- Will any of the four banks confirm their pilot, even unnamed-but-characterized? Without one source, the strongest angle here rests entirely on your own claim.
- "Pilot" is doing heavy lifting. How many seats, which teams, and is there a path to production deployment — or could all four quietly end?
- Is CMEK and residency included for *new* Enterprise customers at current pricing, or is "no pricing change for existing plans" hiding an add-on SKU?
- What does offline actually cover? If databases or shared editing don't work offline, the contrarian angle thins fast and a technical reporter will find that in the first demo.
- Is the ~100M figure registered users, monthly actives, or something else — and can you defend it on the record?

**Next step:** `journalist-fit-check` — four viable shapes exist across four different press corps; pick one angle and resolve real names against it before any drafting.
