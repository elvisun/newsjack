# Newsjack Detector Rubric

Use this rubric after the engine returns queued evidence. The engine exposes mechanical scores and `routing.queue_priority`; neither is a PR judgment.

The engine has two discovery lanes:

- `profile_relevance` - profile/topic/competitor queries. These catch highly relevant but sometimes minor stories.
- `major_news` - curated RSS/Atom feed items. These catch broader major news first, then require a stricter client-relevance judgment.

Do not treat a `major_news` item as pitchable because it is big. The client still needs standing, proof, and a journalist shape.

## Story Size

Use `story_size` to calibrate effort, not to approve a pitch. It is a deterministic media-attention proxy based on news-search publication metadata:

- log-scaled estimated monthly traffic
- domain authority
- coverage spread across independently surfaced domains

`major` or `high` story size means the opportunity may justify faster review and sharper proof asks. It does not compensate for stale timing, weak standing, missing proof, or a bad journalist shape.

## Freshness Gate

For recurring or beta cron output, the LLM `story-origin-check` recovers the first-public timestamp and canonical coverage, then the Go CLI `origin-apply` computes the freshness gate. News-search `published_at` values are reliable evidence for article timestamps and should be used to find candidate originals, but they are not alone a same-story or first-publication judgment.

Before assigning `pitch_now`, `develop_angle`, or `monitor`, inspect `freshness_gate.computed_status`:

- `fresh` - eligible for normal judgment.
- `fresh_new_development` - eligible, but the angle must be about the new development, not the older background story.
- `stale` - reject as stale.
- `freshness_unverified` or missing - reject as `freshness_unverified` for cron/beta output.

Do not reset the clock because an aggregator, syndication partner, or secondary outlet republished an older article.

When citing the story, prefer `story_origin.canonical_coverage_url` when present. It should be the major or most authoritative same-story coverage, such as a primary source, wire, major publisher, or recognized trade, instead of the small pickup that triggered retrieval.

## Verdict Ladder

### pitch_now

Use only when all are true:

- Evidence is fresh: usually `30min`, `4hr`, or `24hr`.
- The first public story clock is verified as inside the last 24 hours, or the new development is inside the last 24 hours.
- At least one credible news source exists, preferably `news_search`.
- The client has direct standing to comment.
- The client can supply proof or a real spokesperson now.
- A specific reporter shape is obvious.
- No hard brand-safety block applies.

### develop_angle

Use when the signal is real but needs framing:

- Fresh or still within the week.
- Client standing is plausible but not yet sharp.
- Proof exists or can be requested.
- A journalist shape exists, but the angle needs work.
- Major-news lane items often belong here when they are important but the client angle is indirect.

Handoff: `angle-generator`.

### monitor

Use when the signal is interesting but not pitch-ready:

- Single-source or weak cross-source confirmation.
- Early chatter without enough news confirmation.
- The client might have standing, but proof is missing.
- The signal may matter if it gains traction.

### reject

Use when any core gate fails:

- stale
- freshness unverified in cron/beta output
- no client standing
- missing proof that cannot be supplied quickly
- no plausible journalist shape
- off-beat
- already seen with no new development
- weak source quality

## Decay

Decay uses the verified first-public timestamp from `story-origin-check`. Engine `features.decay_bucket` is provisional when evidence comes from aggregators, syndication partners, secondary rewrites, or search results that have not yet been matched to the original/canonical story.

- `30min` - live/breaking. Only use for immediate comment if the client can respond now.
- `4hr` - same-cycle. Good for reactive comment.
- `24hr` - still fresh. Good for angle generation or same-day response.
- `week` - trend/context only. Do not call it breaking.
- `month` - usually not a newsjack unless paired with a new data point or fresh hook.
- `unknown` - do not pitch as timely without independent timestamp verification.

## Standing

Strong standing:

- The client operates directly in the affected market.
- The client has first-party data, customer evidence, technical expertise, or a named executive who can speak concretely.
- The signal names the client's category, customers, regulators, technology, or competitors.

Partial standing:

- The client has adjacent expertise but needs a narrower angle.
- The client can explain impact but not the core event.

Weak standing:

- The client merely sells into the broad category.
- The client wants to comment because the topic is popular.
- The proof is generic thought leadership.

For `major_news` lane signals, standing must explain the bridge from the public story to the client:

- same buyer being affected
- same regulator or policy surface
- named competitor or platform move
- client has first-party data about the consequence
- client can explain a non-obvious operational effect

If the bridge is "this is about AI and the client uses AI," reject or monitor.

## Proof

Required proof must be specific:

- first-party data
- named customer or user example
- real executive quote
- technical artifact
- regulatory or market analysis the client can defend
- recent product or customer evidence

Do not accept generic claims like "we help companies with AI" as proof.

## Journalist Shape

A useful journalist shape names:

- exact beat
- outlet archetype
- why the beat cares now
- who should not receive it

Bad shapes:

- "business reporter"
- "AI journalist"
- "tech media"
- "industry press"

Good shapes:

- "enterprise AI reporter covering vendor compliance claims after regulator action"
- "cybersecurity trade reporter covering identity-risk fallout from new enforcement"
- "retail operations reporter covering labor-cost impact of a same-day policy change"

## Hard Blocks

Block signals built on:

- death
- violence
- disaster
- war
- abuse
- sexual violence
- missing people
- humanitarian crisis
- hate crime
- terror
- suicide

The only acceptable work around these topics is restrained expert commentary with direct public-interest standing. Promotional hooks are refused.
