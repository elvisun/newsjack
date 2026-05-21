# Fact Check - Worked Examples

These examples show the expected shape. Reserved `example.com` URLs are used
only for fictional companies and publications; live runs must use real,
reachable URLs.

---

## Example 1: Mixed Result With Verified, Missing Source, And Disputed Claims

### Input

```text
current_time: 2026-05-18T14:00:00Z

Hi Rowan,

Saw your April 30 piece in Growth Ledger on CFOs cutting SaaS seats.
AcmeOps CEO Maya Patel says the company reduced finance-team software spend
by 41% in Q1. Rivet Research's latest finance software survey shows 68% of CFOs are
consolidating tools this year, so this is becoming a board-level issue.

Maya was PioneerPay's first CFO before founding AcmeOps.
```

### Output

```md
## Fact-check verdict
Risky as written. The byline reference verifies, but the statistic is missing
an auditable source and the PioneerPay title claim is disputed by the source found.

## Facts & Citations
1. **Claim:** "your April 30 piece in Growth Ledger on CFOs cutting SaaS seats"
   - **Status:** Verified
   - **Citation(s):** Growth Ledger author page and article archive: https://growthledger.example/rowan/cfo-saas-seats-april-30
   - **Notes:** The page lists Rowan as author and publication date as 2026-04-30. Fictional example URL; live runs must cite the real publication URL.

2. **Claim:** "AcmeOps CEO Maya Patel"
   - **Status:** Verified
   - **Citation(s):** AcmeOps leadership page: https://acmeops.example/company/leadership
   - **Notes:** The page lists Maya Patel as CEO and was updated 2026-05-10, inside the 30-day freshness window.

3. **Claim:** "the company reduced finance-team software spend by 41% in Q1"
   - **Status:** Missing source
   - **Citation(s):** None found
   - **Notes:** This is an internal performance claim. It may be true, but the draft gives no public source, customer proof, or document a journalist can audit.

4. **Claim:** "Rivet Research's latest finance software survey shows 68% of CFOs are consolidating tools this year"
   - **Status:** Missing source
   - **Citation(s):** None found
   - **Notes:** The claim names a report and exact statistic but supplies no URL, report title, or publication date. Search did not identify the exact 68% figure with confidence.

5. **Claim:** "Maya was PioneerPay's first CFO before founding AcmeOps"
   - **Status:** Disputed
   - **Citation(s):** PioneerPay archived leadership history: https://pioneerpay.example/history/leadership; AcmeOps founder bio: https://acmeops.example/company/maya-patel
   - **Notes:** The sources support that Maya Patel worked in finance leadership at PioneerPay, but not that she was PioneerPay's first CFO. The official history lists another person in that role first.

## Warning
Do not send this draft as written. Add auditable citations for the 41% internal
spend claim and the Rivet Research statistic, or remove them. The "first CFO" line is
contradicted by the cited leadership history and should be corrected before
any journalist sees it. All fictional example URLs here are format examples,
not live evidence.
```

---

## Example 2: Clean Pass With Supplied Source URLs

### Input

```text
current_time: 2026-05-18T09:00:00Z

Source links:
- https://northstar.example/news/series-a-2026
- https://northstar.example/team

Northstar Workflow raised a $12M Series A led by Vale Ventures on May 14,
2026. CEO Lina Torres founded the company after leading operations at a
regional logistics provider.
```

### Output

```md
## Fact-check verdict
Safe on the checked facts. Both material claims are supported by supplied
company sources, though an independent investor source would strengthen the
funding claim.

## Facts & Citations
1. **Claim:** "Northstar Workflow raised a $12M Series A led by Vale Ventures on May 14, 2026"
   - **Status:** Verified
   - **Citation(s):** Northstar press release: https://northstar.example/news/series-a-2026
   - **Notes:** The supplied source states the amount, lead investor, and date. Because this is a fictional example URL, a live run should prefer a real press release plus investor confirmation.

2. **Claim:** "CEO Lina Torres founded the company after leading operations at a regional logistics provider"
   - **Status:** Verified
   - **Citation(s):** Northstar team page: https://northstar.example/team
   - **Notes:** The source supports current CEO status and founder background. If the pitch names the prior employer, verify that employer separately.

## Warning
Residual risk is low for this example, but the funding claim relies on a
company-owned source. For a real pitch, add the investor announcement, filing,
or another independent source if available.
```

---

## Example 3: No Live Retrieval Available

### Input

```text
current_time: 2026-05-18T09:00:00Z

I saw your piece last week about AI agents replacing SDR teams. Our CTO
David Kim says 73% of enterprise sales teams will consolidate outbound tools
this year.
```

### Output

```md
## Fact-check verdict
Not safe to rely on yet. The claim ledger is extracted, but without retrieval
or supplied URLs the factual claims cannot be verified.

## Facts & Citations
1. **Claim:** "your piece last week about AI agents replacing SDR teams"
   - **Status:** Unverifiable
   - **Citation(s):** None found
   - **Notes:** This needs the article URL, publication name, author, and date. "Last week" depends on the supplied current time and cannot be verified without retrieval.

2. **Claim:** "Our CTO David Kim"
   - **Status:** Unverifiable
   - **Citation(s):** None found
   - **Notes:** Current title claims require recent public evidence or an authoritative supplied source.

3. **Claim:** "73% of enterprise sales teams will consolidate outbound tools this year"
   - **Status:** Missing source
   - **Citation(s):** None found
   - **Notes:** Exact statistic with no report name, URL, methodology, or date. Do not leave this in a journalist-facing draft without a citation.

## Warning
Source lookup was unavailable and no URLs were supplied. Treat this as a
failed fact-check, not a pass. Add the byline URL, a current team or company
source for David Kim's title, and the report behind the 73% statistic before
sending.
```

---

## Example 4: Placeholder And Quote Risk

### Input

```text
current_time: 2026-05-18T09:00:00Z

According to recent industry analysis, "AI will replace half of comms work by
2027," said Dr. Evelyn Shaw at the Global Comms Forum. Our customer results
show a XX% reduction in response time.
```

### Output

```md
## Fact-check verdict
Blocked by source and placeholder risk. The quote has no verifiable source in
the draft, and `XX%` is an unfilled factual placeholder.

## Facts & Citations
1. **Claim:** "According to recent industry analysis"
   - **Status:** Missing source
   - **Citation(s):** None found
   - **Notes:** The phrase names no report, publisher, date, or URL. "Recent" also requires a current-time anchor and publication date.

2. **Claim:** "\"AI will replace half of comms work by 2027,\" said Dr. Evelyn Shaw at the Global Comms Forum"
   - **Status:** Unverifiable
   - **Citation(s):** None found
   - **Notes:** The quote requires a transcript, event page, recording, or published article tying the exact words to the named speaker and venue.

3. **Claim:** "Our customer results show a XX% reduction in response time"
   - **Status:** Missing source
   - **Citation(s):** None found
   - **Notes:** `XX%` is an unfilled placeholder. The metric also needs a source, measurement period, and baseline.

## Warning
Do not send this. Replace or remove the placeholder, cite the industry
analysis, and provide a source for the quote. If the quote was private or
paraphrased, do not present it as a public quotation.
```
