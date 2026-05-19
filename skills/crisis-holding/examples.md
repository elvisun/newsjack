# crisis-holding - Worked Examples

Realistic before/after examples for the four crisis patterns the skill must handle: counsel gate, counsel-engaged drafting, landmine-newsjack recovery, and intake refusal.

---

## Example 1: Product Safety, Counsel Not Engaged

### Before

The user wants a fast publishable statement.

```yaml
incident_summary: "We have reports from three customers that our smart lock model SL-200 unlocked unexpectedly in the last 48 hours. We've confirmed two of the three. We don't know the root cause yet."
incident_type: product_safety
incident_first_known_at: 2026-05-18T17:00:00Z
org_name: Northgate Security
org_role_of_user: head of comms
audience: [press, customers]
known_facts:
  - two of three reports independently verified
  - both verified incidents involved model SL-200 firmware v3.2
  - no injuries reported in the verified cases
unknown_or_unverified:
  - third report
  - root cause
  - whether other firmware versions are affected
  - total number of SL-200 units in the field
actions_taken_so_far:
  - support tickets opened with the two verified customers
  - engineering investigation started this morning
actions_committed_to: []
people_involved: []
legal_status: no_counsel_yet
regulatory_exposure: "CPSC may be relevant under Consumer Product Safety Act if root cause is a defect"
media_inquiry_timing: inbound_within_4h
```

Unsafe draft the user wanted:

> Northgate Security has identified an isolated firmware issue affecting a small number of SL-200 locks. Out of an abundance of caution, we are launching a comprehensive investigation and will issue a fix within 24 hours.

### After

The skill refuses to draft publishable copy.

````markdown
## STOP - Legal counsel required before any external statement

Trigger: `incident_type == product_safety` and `legal_status == no_counsel_yet`. Additional trigger: `regulatory_exposure` mentions CPSC.

Why this gate exists: A holding statement on a product-safety incident can become an admission of defect, scope, or notice. Making a public statement before knowing reporting obligations adds procedural exposure to the underlying incident.

Next steps:
1. Page general counsel or product-liability outside counsel now.
2. Inbound press gets exactly: "We are aware of the reports and are investigating. We'll have more to share once we've completed our initial review."
3. Re-run this skill with `legal_status: counsel_engaged_reviewing` to get drafts.

If you need draft language for counsel to review, re-invoke with `--counsel-review-mode`.
````

What the skill caught:

- "isolated" is unverifiable.
- "small number" is unsupported because total scope is unknown.
- "out of an abundance of caution" is banned hedge.
- "comprehensive investigation" is invented.
- "fix within 24 hours" is invented.

---

## Example 2: Data Security, Counsel Engaged

### Before

The user has a structured intake and counsel is already involved.

```yaml
incident_summary: "We detected unauthorized access to a customer database table containing email addresses and hashed passwords on the morning of May 17. We rotated keys and forced password resets for affected accounts. We don't know if data was exfiltrated."
incident_type: data_security
incident_first_known_at: 2026-05-17T08:30:00Z
org_name: Loomwork
org_role_of_user: VP comms
audience: [press, customers, regulators]
known_facts:
  - unauthorized access detected at 08:30 UTC on May 17
  - affected table contained email addresses and bcrypt-hashed passwords
  - affected table did not contain payment information, message content, or document content
  - 47,200 accounts were in the affected table
  - access vector was a compromised internal API key
unknown_or_unverified:
  - whether data was exfiltrated
  - full root cause
actions_taken_so_far:
  - rotated all internal API keys, completed May 17 by 11:00 UTC
  - forced password reset for 47,200 affected accounts, in progress and about 80 percent complete
  - engaged Mandiant for forensic review
  - notified our DPO and Irish Data Protection Commission
actions_committed_to:
  - publish a post-incident write-up within 14 days
  - notify any user whose data is confirmed exfiltrated within 72 hours of confirmation
people_involved: []
legal_status: counsel_engaged_reviewing
regulatory_exposure: "GDPR; Irish DPC notified under Article 33"
media_inquiry_timing: inbound_within_24h
```

Unsafe draft the user considered:

> Loomwork takes customer security seriously. Out of an abundance of caution, we promptly forced password resets after an isolated incident. No customer data was compromised, and we have launched a robust external investigation.

### After

````markdown
# Holding draft - Loomwork - 2026-05-18T17:14:00Z - valid until 2026-05-18T19:14:00Z

## Short (39 words)

```text
Loomwork detected unauthorized access to a customer database table on May 17 affecting 47,200 accounts. We rotated internal API keys, forced password resets for affected accounts, engaged Mandiant, and notified the Irish Data Protection Commission.
```

## Medium (121 words)

```text
On May 17 at 08:30 UTC, Loomwork detected unauthorized access to a customer database table containing email addresses and bcrypt-hashed passwords for 47,200 accounts. The affected table did not contain payment information, message content, or document content. We do not yet know whether data was exfiltrated.

We rotated all internal API keys by 11:00 UTC that day, forced password resets for affected accounts, engaged Mandiant for forensic review, and notified our DPO and the Irish Data Protection Commission. Password resets are in progress and about 80 percent complete.

We will publish a post-incident write-up within 14 days and notify any user whose data is confirmed exfiltrated within 72 hours of confirmation.
```

## Cautious legal pass (116 words)

```text
On May 17 at 08:30 UTC, Loomwork detected what appears to be unauthorized access to a customer database table containing email addresses and bcrypt-hashed passwords for 47,200 accounts. Based on what we currently know, the affected table did not contain payment information, message content, or document content. We do not yet know whether data was exfiltrated.

We have rotated internal API keys, have begun forcing password resets for affected accounts, and engaged Mandiant for forensic review. We have notified our DPO and the Irish Data Protection Commission.

We will update this statement as our understanding develops.
```

Deltas from medium:
- "detected unauthorized access" became "detected what appears to be unauthorized access."
- "The affected table did not contain" became "Based on what we currently know, the affected table did not contain."
- Removed exact completion time for key rotation.
- "forced password resets" became "have begun forcing password resets."
- Removed the 14-day post-incident write-up and 72-hour notification commitments for counsel to decide.

## Q&A scaffold

| Category | Question | Posture | Rationale | Draft response or holding line |
|---|---|---|---|---|
| facts | When did you detect the access? | answer | Timestamp is confirmed. | We detected it at 08:30 UTC on May 17. |
| scope | How many accounts were affected? | answer | Account count is confirmed. | 47,200 accounts were in the affected table. |
| scope | What data was in the table? | answer | Data categories are confirmed. | Email addresses and bcrypt-hashed passwords. The table did not contain payment information, message content, or document content. |
| responsibility | Was this an attack or a misconfiguration? | decline-and-name-why | Root cause is not confirmed. | Mandiant's forensic review is underway. We'll share findings when we can confirm them. |
| remediation | Have all passwords been reset? | answer | Status is confirmed but incomplete. | Password resets are in progress and about 80 percent complete. |
| legal | Have you notified regulators? | answer | Irish DPC notice is confirmed. | We notified our DPO and the Irish Data Protection Commission. |
| business | Is this material to the business? | decline-and-name-why | The intake does not include materiality facts. | We are not making forward-looking statements at this point. |

## What not to say

| Phrase | Reason | Suggested rewrite |
|---|---|---|
| "takes customer security seriously" | Parodied crisis boilerplate. Demonstrate seriousness with actions. | Name the key rotation, password resets, Mandiant review, and DPC notice. |
| "out of an abundance of caution" | Banned hedge. | State the action and why it was taken. |
| "promptly" | Vague timing. | Use 11:00 UTC if counsel clears it. |
| "isolated incident" | Scope is not fully known. | Omit. |
| "No customer data was compromised" | Exfiltration is unknown. | "We do not yet know whether data was exfiltrated." |
| "robust external investigation" | "Robust" is filler; the firm matters. | "Mandiant forensic review." |
````

Why this works: the draft says less than the unsafe version, but every sentence is defensible from the intake.

---

## Example 3: Landmine Newsjack

### Before

The company's social account used an active armed conflict as a hook for a product promotion.

```yaml
incident_summary: "Our social team posted a product promo at 9:15 ET this morning that referenced an active armed conflict and is getting backlash. The post is still up. We don't know who approved it."
incident_type: landmine_newsjack
incident_first_known_at: 2026-05-18T13:15:00Z
org_name: Aster & Finch
org_role_of_user: head of comms
audience: [press, public_social]
known_facts:
  - post went live at 13:15 UTC
  - post is still live as of 13:42 UTC
  - the social team has autonomous posting authority for standard product content
unknown_or_unverified:
  - who approved the post
  - whether anyone reviewed the timing
  - whether internal social-content guidelines were followed
actions_taken_so_far: []
actions_committed_to: []
people_involved: []
legal_status: no_counsel_yet
regulatory_exposure: none
media_inquiry_timing: inbound_within_1h
```

Unsafe draft the user wanted:

> Aster & Finch has always stood with affected communities. We regret any offense caused by today's post and remain deeply committed to using our platform for good. Moving forward, we will donate a portion of proceeds from the featured product to relief efforts.

### After

First response:

```markdown
## STOP - One action before drafting

The post is still up. Pull it. Then return to this skill.

Every minute the post stays live is another screenshot added to the search results. This is not legal counsel. This is the unranked first step. Pull the post, then draft.
```

After the user confirms the post is down:

````markdown
# Holding draft - Aster & Finch - 2026-05-18T13:49:00Z - valid until 2026-05-18T14:19:00Z

## Short (26 words)

```text
Earlier today, we posted content that referenced an active armed conflict in a way that was wrong. We took the post down and are reviewing how it went live.
```

## Medium (82 words)

```text
Earlier today, at 13:15 UTC, Aster & Finch posted promotional content that referenced an active armed conflict. The post was inappropriate. We took it down.

We are reviewing how this content was approved and posted. The people most affected by the conflict deserve better than a brand using their situation as a product hook. We're sorry.

We are not making further statements about the conflict or our internal process until the review is complete.
```

## Cautious legal pass (73 words)

```text
Earlier today, at 13:15 UTC, Aster & Finch posted promotional content that referenced an active armed conflict. We understand the post caused offense. We have taken the post down.

We have begun an internal review of how this content was approved and posted. We will share findings when the review is complete. We will update this statement as our understanding develops.
```

Deltas from medium:
- "was inappropriate" became "caused offense."
- "took it down" became "have taken the post down."
- Removed the audience-facing apology sentence for counsel to decide.

## Q&A scaffold

| Category | Question | Posture | Rationale | Draft response or holding line |
|---|---|---|---|---|
| facts | When did the post go live? | answer | Timestamp is confirmed. | It went live at 13:15 UTC. |
| facts | When did it come down? | answer | Answer only after the user confirms removal time. | We took it down at the confirmed removal time. |
| responsibility | Who approved the post? | decline-and-name-why | Approval path is unknown. | We are reviewing how it was approved and posted. We will not name people before that review is complete. |
| people | Has anyone been disciplined? | refer-to-counsel | Personnel actions are not for a first statement. | We are not discussing personnel matters. |
| legal | Are you donating to the cause? | decline-and-name-why | Donation talk extends the landmine. | We are not announcing anything related to the conflict today. |

## What not to say

| Phrase | Reason | Suggested rewrite |
|---|---|---|
| "has always stood with affected communities" | Defensive values preamble. | Omit. |
| "regret any offense caused" | Non-apology that centers reaction, not action. | "The post was inappropriate." |
| "deeply committed" | Banned crisis filler. | Omit. |
| "using our platform for good" | Brand mission copy in a crisis. | Omit. |
| product donation language | Repeats the original error by tying the crisis to the product. | Omit. |
````

Why this works: the statement removes the product, does not reframe the conflict around the brand, and keeps the decay window to 30 minutes.

---

## Example 4: User Skips Intake on a Minor-Harm Allegation

### Before

User:

> We have a problem. A customer is going to the press tomorrow morning saying our product hurt their kid. Write me a holding statement now.

### After

```markdown
Stop. Two things before drafting.

1. The phrase "hurt their kid" puts this in the legal-counsel-required category before we discuss anything else. If you do not have counsel engaged on this, that is the first call. Not after a draft. Before.

2. I cannot responsibly draft from a one-line summary on an incident involving a minor and alleged product harm. Walk me through the intake. Start with: your org, your role, the specific allegation as stated by the customer, what you can confirm, and what you cannot.

If counsel is engaged and you want language for counsel to review, not publish, re-invoke with `--counsel-review-mode` and the full intake.
```

Why this works: the skill refuses the wrong task but keeps the user moving. It names the trigger, asks for the next facts, and preserves the counsel-review path.
