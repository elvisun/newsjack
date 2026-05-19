# Reactive Comment - Worked Examples

Four realistic examples showing the expected shape: input query and profile
before, YAML verdict after. These are not templates to blast. They show
what the gate should accept, kill, or pause.

---

## Example 1: Clean Fit, Draft

### Before

```yaml
profile:
  name: "Jane Doe"
  title: "Co-founder & CTO"
  company: "Acme Security"
  expertise_areas:
    - "application security for SaaS startups"
    - "secrets management in CI/CD"
  do_not_comment_on:
    - "cryptocurrency"
    - "criminal cases"
    - "competitor product comparisons"
  proof_points:
    - claim: "BlackHat 2024 talk on CI/CD secret leakage"
      verifiable_at: "https://blackhat.com/us-24/briefings/schedule/#leak"
    - claim: "10 years building auth systems at Stripe and Okta"
      verifiable_at: "https://linkedin.com/in/janedoe"
  contact_block: |
    Jane Doe, Co-founder & CTO, Acme Security
    jane@acme.security
    linkedin.com/in/janedoe
  response_cap_per_week: 5
query:
  source: "SourceOfSources"
  journalist_name: "Alex Rivera"
  journalist_outlet: "TechCrunch"
  query_text: "Looking for CTOs at Series A-B SaaS startups: has Copilot/Cursor changed how secrets leak into your repos? Real anecdotes welcome. No vendor pitches. Deadline EOD Friday."
  deadline_iso: "2026-05-22T23:59:00-07:00"
  received_at_iso: "2026-05-20T08:14:00-07:00"
recent_context:
  journalist_bylines:
    - title: "GitHub Copilot's hidden cost: a year of leaked tokens"
      url: "https://techcrunch.com/2026/03/14/copilot-leak-year"
      published_iso: "2026-03-14"
internal_state:
  responses_to_this_source_this_week: 2
```

### After

```yaml
verdict: draft
fit_score: 84
fit_reasoning: |
  Direct fit on CI/CD secrets and SaaS application security. The query asks
  for CTOs discussing Copilot/Cursor secret leakage, and the profile has a
  BlackHat proof point on CI/CD secret leakage plus auth tenure at Stripe
  and Okta. The journalist's recent byline is on leaked tokens. "No vendor
  pitches" is compatible because the response offers an operator anecdote,
  not a product pitch.
decay_flags:
  hours_until_deadline: 56
  is_fresh: true
  warning: null
draft_response:
  subject: "Re: AI-code secret leakage, CTO at Acme Security"
  body: |
    Hi Alex,

    Your March piece on Copilot's leaked-token year matches what we saw from the founder side at Acme. After we instrumented Cursor and Copilot against our repos earlier this year, AI-suggested code raised our secret-leak rate roughly 3x before we put server-side scanning in front of every push. The pattern worth flagging is autocompletion of `.env.example` files with real values, not malice, just muscle memory. Happy to go on record: 10 years building auth at Stripe and Okta before Acme, plus the CI/CD secrets talk at BlackHat 2024.

    Jane Doe, Co-founder & CTO, Acme Security
    jane@acme.security
    linkedin.com/in/janedoe
provenance:
  - claim: "BlackHat 2024 CI/CD secret leakage talk"
    sourced_from: "profile.proof_points[0]"
  - claim: "10 years building auth at Stripe and Okta"
    sourced_from: "profile.proof_points[1]"
  - claim: "TechCrunch byline on Copilot leaked tokens"
    sourced_from: "recent_context.journalist_bylines[0]"
  - claim: "Acme instrumented Cursor and Copilot, roughly 3x secret-leak rate"
    sourced_from: "USER MUST CONFIRM"
slop_check:
  banned_words_found: []
  emdash_count: 0
  placeholders_found: []
  ai_tells_found: []
  passed: true
anti_spray:
  responses_to_this_source_this_week: 2
  cap: 5
  passed: true
next_action: |
  Confirm the Cursor/Copilot instrumentation and 3x figure before sending.
  If the number is directional, change it to "noticeably raised." Send
  manually from your mail client. Do not auto-send.
```

---

## Example 2: Adjacent Topic, Hard Kill

### Before

```yaml
profile:
  name: "Jane Doe"
  title: "Co-founder & CTO"
  company: "Acme Security"
  expertise_areas:
    - "application security for SaaS startups"
    - "SOC 2 Type 2 readiness for sub-100-headcount teams"
  do_not_comment_on:
    - "cryptocurrency"
    - "competitor product comparisons"
  proof_points:
    - claim: "Led SOC 2 Type 2 readiness at Acme Security"
      verifiable_at: "https://acme.security/security"
  contact_block: |
    Jane Doe, Co-founder & CTO, Acme Security
    jane@acme.security
query:
  source: "Featured"
  journalist_name: "Sam Lin"
  journalist_outlet: "CoinDesk"
  query_text: "How are SaaS CTOs thinking about on-chain authentication for B2B SSO? Looking for comments on wallet-based identity and enterprise readiness."
  deadline_iso: "2026-05-23T17:00:00-07:00"
internal_state:
  responses_to_this_source_this_week: 1
```

### After

```yaml
verdict: kill
fit_score: 0
kill_reason: |
  The query is about on-chain authentication and wallet-based identity.
  Your profile's `do_not_comment_on` includes cryptocurrency. The SaaS SSO
  angle is adjacent, but the requested comment would land inside crypto
  coverage at CoinDesk. Hard veto.
why_not_your_fight: |
  You will be tempted because "SSO" is close to your work. The actual piece
  is about wallet-based identity. Anything useful you say will be excerpted
  into a crypto-auth story, and then you will spend time clarifying a
  position your profile already told us not to take. Skip it.
suggested_alternative: |
  If this topic matters to you, publish your own post on why wallet-based
  SSO is not ready for mainstream SaaS. That gives future reporters a clean
  reason to come to you without violating the profile.
decay_flags:
  hours_until_deadline: 72
  is_fresh: true
  warning: null
anti_spray:
  responses_to_this_source_this_week: 1
  cap: 5
  passed: true
```

---

## Example 3: Missing Proof, Ask

### Before

```yaml
profile:
  name: "Maya Patel"
  title: "VP Engineering"
  company: "Northstar Data"
  expertise_areas:
    - "data platform reliability"
    - "incident response for enterprise analytics teams"
  do_not_comment_on:
    - "layoffs"
    - "criminal cases"
  proof_points:
    - claim: "Built Northstar's incident review process after a 2024 warehouse outage"
      verifiable_at: "https://northstar.example.com/engineering/outage-review"
  contact_block: |
    Maya Patel, VP Engineering, Northstar Data
    maya@northstar.example.com
query:
  source: "Qwoted"
  journalist_name: "Priya Mehta"
  journalist_outlet: "InformationWeek"
  query_text: "Looking for CISOs at companies with 500-5,000 employees to discuss how AI governance policies are being structured. Need examples of internal policy language and ownership."
  deadline_iso: "2026-05-21T18:00:00-07:00"
recent_context: null
internal_state:
  responses_to_this_source_this_week: 0
```

### After

```yaml
verdict: ask
missing_info:
  - "The query asks for CISOs. Profile says VP Engineering. Are you the accountable security-policy owner, or is this outside your title scope?"
  - "The query requires companies with 500-5,000 employees. What is Northstar's current headcount?"
  - "The query asks for internal AI governance policy language. Add a proof point or paste approved policy language that can be quoted."
  - "Recent context is missing. Paste 1-2 recent Priya Mehta bylines or rerun with browse enabled."
why_we_paused: |
  Drafting now would guess title scope, company size, policy ownership, and
  journalist context. Any one of those could make this a clean kill.
decay_flags:
  hours_until_deadline: 30
  is_fresh: true
  warning: null
anti_spray:
  responses_to_this_source_this_week: 0
  cap: 5
  passed: true
```

---

## Example 4: Strong Fit, Cap Refusal

### Before

```yaml
profile:
  name: "Owen Chen"
  title: "Founder"
  company: "Ledgerfield"
  expertise_areas:
    - "restaurant payroll compliance"
    - "tip-credit law for multi-state hospitality groups"
  do_not_comment_on:
    - "immigration enforcement"
  proof_points:
    - claim: "Testified to the California Senate Labor Committee on tip-credit compliance in 2025"
      verifiable_at: "https://senate.ca.gov/archive/tip-credit-hearing-2025"
  contact_block: |
    Owen Chen, Founder, Ledgerfield
    owen@ledgerfield.example.com
  response_cap_per_week: 5
query:
  source: "HelpAB2BWriter"
  journalist_name: "Dana Ortiz"
  journalist_outlet: "Restaurant Business"
  query_text: "Need payroll operators who can explain the practical impact of the new state tip-credit rules for restaurant groups operating in 3+ states. Deadline tomorrow noon ET."
  deadline_iso: "2026-05-20T12:00:00-04:00"
recent_context:
  journalist_bylines:
    - title: "Restaurant chains brace for state wage-law divergence"
      url: "https://restaurantbusiness.example.com/labor/state-wage-law-divergence"
      published_iso: "2026-05-01"
internal_state:
  responses_to_this_source_this_week: 5
```

### After

```yaml
verdict: kill
fit_score: 78
kill_reason: |
  Topic fit is real: restaurant payroll compliance, multi-state tip-credit
  rules, and a recent journalist byline on state wage-law divergence. Cap
  is the blocker. You have already drafted 5 Help A B2B Writer responses
  this rolling week.
why_not_your_fight: |
  This is exactly why the cap exists. A good-fit sixth response still makes
  the account look like it is spraying. If this is better than one of the
  five already drafted, retire one of those first. Otherwise hold it.
suggested_alternative: |
  Turn the point into a short owned post on how multi-state restaurant
  groups should audit tip-credit exposure. It can support the next query
  without spending another source-platform slot.
decay_flags:
  hours_until_deadline: 20
  is_fresh: true
  warning: "Less than 24 hours until deadline."
anti_spray:
  responses_to_this_source_this_week: 5
  cap: 5
  passed: false
```
