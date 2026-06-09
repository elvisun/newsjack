**Four of the five largest US banks are piloting Notion — after it agreed to hand over the keys**

- **Story type:** customer-story
- **Why a journalist cares:** A banking-technology trade reporter covering how the largest US banks vet and adopt SaaS tools. The hook is the buyer, not the feature list: the biggest regulated institutions in the country running pilots of a 100M-user collaboration app is a procurement story about what it took to get them in the door — customer-held keys via AWS KMS, offline mode, data residency. Not for consumer productivity reviewers, general startup press, or security-research desks looking for vulnerability stories.
- **Why now:** The CMEK and offline-mode release ships now, and the bank pilots are the proof it landed with the hardest buyers.
- **Decay:** 24hr — standard launch-cycle news; the pilots claim keeps it alive slightly longer, but the announcement is the peg.
- **What makes it distinct:** Perspective lens, customer protagonist, plus the power-elite news value. The banks are the story; the other kept angles make the rules, the architecture, and the pricing the story.
- **Proof it needs:** at least one bank willing to be named or referenced on background; substantiation for the "four of the five largest US banks" claim that survives a journalist's verification call; pilot scope (seats, departments, timeline) so "pilot" doesn't deflate on first question.
- **Facts it rests on:** four of the five largest US banks now run pilots; enterprise customers can hold their own encryption keys via AWS KMS; ~100M users; the company says this unblocks regulated buyers in finance and healthcare.

**Data residency comes for collaboration software: Notion adds EU and Australia options as sovereignty demands spread**

- **Story type:** trend
- **Why a journalist cares:** A data-policy or privacy-compliance reporter at a policy outlet or enterprise-IT trade covering data sovereignty, cross-border transfer rules, and key custody. Notion adding EU and Australia residency plus customer-managed keys is a concrete data point in the story that sovereignty requirements are now shaping product roadmaps at workspace-software vendors, not just cloud infrastructure providers. Not for banking trade press, product reviewers, or founder-profile writers.
- **Why now:** The residency and CMEK options ship now; the sovereignty context gives it a longer runway than the launch itself.
- **Decay:** week — a context piece pegged to the launch, not a same-day race.
- **What makes it distinct:** Perspective lens, regulator/rules protagonist — the compliance regime is the story, and Notion is the evidence. The bank angle is about one buyer cohort; this is about the rules forcing the whole category to change.
- **Proof it needs:** which specific regulatory or procurement requirements (EU and Australian) the residency options were built to satisfy; whether residency covers all customer data or a subset; someone at Notion who can speak to what regulated buyers actually demanded.
- **Facts it rests on:** data residency options added for the EU and Australia; enterprise customers can hold their own encryption keys via AWS KMS; the company says this unblocks regulated buyers in finance and healthcare.

**Notion shipping offline mode in 2026 is an admission the always-online SaaS bet had a ceiling**

- **Story type:** contrarian
- **Why a journalist cares:** An enterprise-software reporter at a tech business outlet or developer-leaning newsletter who covers product architecture decisions. The prevailing belief: modern cloud collaboration tools treat offline as a legacy concern not worth the engineering cost. A 100M-user, cloud-native app doing the work to make editing function with no connection — and citing regulated buyers as the reason — cuts against that. Not for policy reporters, banking trades, or general business desks.
- **Why now:** Offline editing ships now; the architecture argument is pegged to the release.
- **Decay:** week — the think-piece window around a launch, not breaking news.
- **What makes it distinct:** Inversion lens — the product architecture is the protagonist. The other angles take the announcement at face value; this one asks why a cloud-native company reversed a core assumption.
- **Proof it needs:** whether offline is full-fidelity (databases, conflict resolution on reconnect) or a cached subset — the angle dies if it's read-only; how long it took to build and why now; an engineering or product lead who can speak to the tradeoffs on record.
- **Facts it rests on:** offline editing now works with no connection and syncs on reconnect; ~100M users; the company says the release unblocks regulated buyers that previously could not adopt.

**Notion adds enterprise key custody without raising prices, cutting against the security-upsell norm**

- **Story type:** counterposition
- **Why a journalist cares:** A SaaS-business reporter covering enterprise software pricing and procurement — the beat that has chronicled vendors gating security features behind premium tiers (the "SSO tax" pattern). Notion shipping CMEK and residency to existing Enterprise plans with no pricing change is a pricing-strategy story with competitive teeth. Not for policy desks, consumer tech, or healthcare trades.
- **Why now:** The no-pricing-change decision ships with the release; the pricing-norms contrast keeps it warm past launch day.
- **Decay:** week — pricing-strategy angles run on the launch peg but don't expire same-day.
- **What makes it distinct:** Contrarian/counterposition lens with the business model as protagonist — none of the other angles touch pricing. It positions against an industry norm, not against the announcement itself.
- **Proof it needs:** confirmation that CMEK and residency are included in the existing Enterprise tier with no add-on fee or new SKU; user-supplied examples of competitors charging extra for comparable features — do not pitch the contrast without them; whether "no pricing change for existing plans" leaves room for higher pricing on new plans (a journalist will ask).
- **Facts it rests on:** no pricing change for existing Enterprise plans; enterprise customers can hold their own encryption keys via AWS KMS; data residency options added for the EU and Australia.

**Refused angles**

- "Notion revolutionizes enterprise collaboration with offline-first security" — `slop`.
- "Hospitals and health systems can finally adopt Notion" — `hallucinated_fact`. The facts contain a company claim about healthcare buyers and zero healthcare customers. The banks carry the finance claim; nothing carries the healthcare one.
- "Notion hits 100M users" — `no_why_now_but_required`. The ~100M figure is supplied as context, not as a new milestone, and a bare user count with no comparison is a number, not a story.
- "What Notion's enterprise push means for the future of work" — `no_journalist_shape`. No sub-beat would claim it.
- "Enterprise IT angle: what it takes for a consumer-grade tool to enter regulated industries" — `duplicate`. It collapses into the bank-pilot and data-sovereignty angles; the checklist framing adds no new protagonist.

**Uncomfortable questions**

- Can any of the four banks be named, or even referenced on background? "Four of the five largest US banks" is a verifiable claim — a banking reporter will try to confirm it, and if it wobbles, it damages every angle, not just one.
- Pilots are not deployments. What is the actual scope, and what is your answer when a journalist asks "how many seats?"
- Is offline editing full-fidelity — databases, merges, conflict resolution — or a cached subset? The contrarian angle and part of the regulated-buyer story rest on it being real.
- CMEK is via AWS KMS only. What do you say to a buyer (or reporter) who asks about other key-management systems, or about which data flows the customer key actually covers?
- The "unblocks regulated buyers" claim cites no certification or audit in the facts you gave me. Is there one? If not, the claim is your assertion, and journalists will treat it that way.

**Next step:** `journalist-fit-check` — four viable shapes exist across four different beats; pick the angle with the strongest proof in hand (the bank-pilot story if a bank can be referenced, the sovereignty trend if not) and resolve real names against it.
