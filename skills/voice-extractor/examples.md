# Voice Extractor - Worked Examples

Real-format examples showing how `voice-extractor` behaves in extract, check, and enforce modes.

---

## Example 1: Founder First-Time Voice Init

### Before

> "I'm doing my first newsjack voice init. Here are 8 samples in `~/samples/`:
> 3 tweets, 2 Slack messages to my cofounder, 2 old emails to journalists from
> 2024, and 1 LinkedIn post. Audience is mostly tech journalists. I write
> pitches and a bit of social."

Sample inventory:

| ID | Source | Audience | Date | Words |
|----|--------|----------|------|------:|
| s_001 | tweet | public | 2026-05-04 | 28 |
| s_002 | tweet | public | 2026-05-08 | 41 |
| s_003 | tweet | public | 2026-05-11 | 36 |
| s_004 | slack | internal | 2026-04-30 | 96 |
| s_005 | slack | internal | 2026-05-02 | 122 |
| s_006 | email | journalist | 2024-11-18 | 310 |
| s_007 | email | journalist | 2024-12-02 | 275 |
| s_008 | linkedin | public | 2026-03-19 | 332 |

### Voice Extractor Output

```yaml
profile_id: jane-doe-personal
sample_count: 8
sample_word_count: 1240
sample_age_p50_days: 27
register: casual-professional
intent: [pitches, social]
extraction:
  confidence: medium
  warnings:
    - "1240 words is usable but light; add 8-10 more native samples for high confidence."
cadence:
  sentence_length: { mean: 11.2, median: 9, p10: 3, p90: 24, stdev: 6.8 }
  paragraph_length: { mean_sentences: 1.7, one_sentence_paragraph_frequency: 0.55 }
  rhythm_signature: short-burst
mechanics:
  contractions: yes
  contraction_rate: 0.81
  em_dash_usage: never
  em_dash_per_1k_words: 0.0
  oxford_comma: no
  exclamation_rate_per_1k_words: 0.6
sentence_initial:
  conjunction_starts_allowed: true
  conjunction_start_rate: 0.18
  uses_however_furthermore_moreover: false
idioms:
  signature_phrases: ["the shape of this is", "two things at once", "fwiw"]
  signature_words: [actually, basically, fwiw, ship]
openers:
  observed: ["Quick one:", "Saw this:", "Heads up:"]
  banned_from_use: ["I hope this email finds you well", "I wanted to reach out"]
```

### After

```text
Voice fingerprint: jane-doe-personal
Saved: ~/.newsjack/voice/jane-doe-personal.yaml
Active profile: yes
Samples: 8 (1240 words)
Register: casual-professional
Confidence: medium

What I captured:
- Cadence: short-burst, mean 11.2 words/sentence, 55% one-sentence paragraphs
- Mechanics: contractions yes, em-dashes never, Oxford comma no
- Signature phrases: the shape of this is; two things at once; fwiw
- Banned for this profile: em-dashes, however/furthermore/moreover, stock pitch openers, global anti-slop list

Warnings:
- Usable but light. Add 8-10 more native samples when you have them.

Refresh after: 2026-08-16
```

Why this works: the skill accepts the 8-sample set, stamps medium confidence, stores a local fingerprint, and makes the em-dash rule explicit before other skills draft as Jane.

---

## Example 2: Enforce Mode Catches A Bot Pitch

### Before

Draft from another newsjack skill:

> Hi Sarah — Hope this finds you well. We're excited to announce that Acme has
> unveiled its revolutionary new platform, which leverages cutting-edge AI to
> deliver world-class results for enterprise customers. In today's
> ever-evolving landscape, it's not just a product, it's a paradigm shift.
> Looking forward to hearing from you. Best, Jane

Active fingerprint: `jane-doe-personal@2026-05-18`, confidence `medium`, em-dash usage `never`.

### Voice Check Output

```json
{
  "verdict": "fail",
  "pass_rate": 0.11,
  "fingerprint_used": "jane-doe-personal@2026-05-18",
  "violations": [
    {"rule": "em_dash_against_fingerprint", "match": "—", "span": [9, 10], "severity": "block", "fix_hint": "fingerprint says em_dash_usage=never; use a comma, period, or colon"},
    {"rule": "banned-opener", "match": "Hope this finds you well", "span": [11, 35], "severity": "block", "fix_hint": "open with the news"},
    {"rule": "banned-word-global", "match": "revolutionary", "span": [76, 89], "severity": "block", "fix_hint": "make a specific claim instead"},
    {"rule": "banned-word-global", "match": "leverages", "span": [125, 134], "severity": "block", "fix_hint": "use 'uses' or rewrite"},
    {"rule": "banned-word-global", "match": "cutting-edge", "span": [138, 150], "severity": "block", "fix_hint": "name the actual method or omit"},
    {"rule": "banned-word-global", "match": "world-class", "span": [171, 182], "severity": "block", "fix_hint": "replace self-awarded praise with evidence"},
    {"rule": "in-todays-world", "match": "In today's ever-evolving landscape", "span": [220, 254], "severity": "block", "fix_hint": "delete the stock setup"},
    {"rule": "not-just-x-its-y", "match": "it's not just a product, it's a paradigm shift", "span": [256, 302], "severity": "block", "fix_hint": "rewrite as a single direct claim"},
    {"rule": "banned-closer", "match": "Looking forward to hearing from you", "span": [304, 339], "severity": "block", "fix_hint": "close with a concrete ask"}
  ],
  "stats": {
    "sentence_length_mean": 24.8,
    "fingerprint_sentence_length_mean": 11.2,
    "drift_score": 0.74
  },
  "regenerate": true
}
```

### After

The drafting skill retries with the fingerprint loaded:

> Quick one: Acme shipped a search tool today that finds duplicate vendor
> contracts before finance approves a renewal.
>
> 14 companies used it in beta. The cleanest result: one customer found
> $1.8M in duplicate renewals in two weeks.
>
> CEO Maya Chen can talk Thursday or Friday. Worth a look?
>
> Jane

Why this works: the retry removes block violations, shortens cadence, uses a documented opener shape, keeps contractions, and closes with a concrete ask.

---

## Example 3: Standalone Check On A Launch Tweet

### Before

Command:

```bash
newsjack voice check ~/Documents/launch-tweet.txt
```

`launch-tweet.txt`:

> Excited to unveil our cutting-edge platform — it's not just a tool, it's a
> paradigm shift for the modern PR practitioner.

### Voice Check Output

```text
voice-check: launch-tweet.txt
fingerprint: jane-doe-personal@2026-05-18 (confidence: medium)
verdict: FAIL (4 blocks, 1 warn)

  x banned-word-global             cutting-edge          [22-34]
  x em_dash_against_fingerprint    —                     [44-45]
  x not-just-x-its-y               "it's not just a tool, it's a paradigm shift"  [46-90]
  x banned-word-global             paradigm              [76-84]
  ! cadence_drift                  sentence_length_mean = 22 (fingerprint: 11)

suggested rewrite:
  "Shipping our new platform today. Not a paradigm shift. Just the
  thing we wish existed."

(rewrite is a hint, not authoritative; run through meanest-editor for the real edit)
```

### After

> Shipping our new platform today. Not a paradigm shift. Just the thing we wish
> existed.

Why this works: check mode flags the mechanical tells and offers a hint without pretending to be the editorial rewrite layer.

---

## Example 4: Mixed-Register Agency Dump

### Before

> "I run PR for 5 clients. Build one voice profile from these 15 samples:
> 5 casual Slack messages I sent my team, 5 formal client approval emails,
> and 5 polished blog posts from Acme's CEO."

### Voice Extractor Response

```text
These are clearly different voices: your internal-casual voice, your formal
client-approval voice, and Acme's CEO voice. If I average them, I get a
fingerprint that's none of them.

Default path:
1. Build your personal agency-sender profile from your own Slack/messages/emails.
2. Build a separate Acme profile only with Acme-approved samples and consent.
3. Switch active profiles with `newsjack voice use <name>`.

Which one should I extract first?
```

### After

Profiles created separately:

```text
~/.newsjack/voice/sarah-agency-personal.yaml
~/.newsjack/voice/sarah-client-approval.yaml
~/.newsjack/voice/acme-ceo-approved.yaml
```

Why this works: the skill refuses to blend incompatible registers and preserves the source-doc boundary between sender voice and brand/client voice.
