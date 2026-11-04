# Evidence signals for a provenance hop

What `trace.mjs` extracts, what each signal is worth, and the published work the thresholds come from. The agent makes every judgment; this note explains which facts to weigh.

## Why first-degree citations are not enough

- Muck Rack's October 2026 *What Is AI Reading?* run (40M responses, 207M citations): earned and third-party pages make up more than 84% of citations; journalism 22–27%; press releases well under 2%. Its summit demo traced one Claude citation three links deep: Wikipedia → trade article → the brand's own press release, which was never cited directly. https://muckrack.com/blog/what-is-ai-reading-october-2026
- Trakkr Research #013, *Zombie stats* (28 Oct 2026): 1,285 AI-cited statistics hand-followed up to five hops. 40% reached a study that said the same thing; 8% reached a study whose number or context had changed; 40% dead-ended on a locked, gone, or silent page; 11% had no study anywhere. The cited page contained the figure 82% of the time but was original research only 21% of the time; company blogs were 54% of first citations. Of the statistics that did reach a source, 17% had changed meaning on the way (year, population, what was measured). https://trakkr.ai/trakkr-research/zombie-stats
- Liu, Zhang & Liang (2023): across four generative search engines, only 51.5% of generated sentences were fully supported by their citations, and 74.5% of citations supported the sentence they sat beside. https://arxiv.org/abs/2304.09848
- Tow Center / CJR (March 2025): eight AI search tools were wrong about the source of a quoted passage more than 60% of the time; several cited syndicated copies (Yahoo, AOL) rather than the original; Grok 3 returned error pages for 154 of 200 citations. https://www.cjr.org/tow_center/we-compared-eight-ai-search-engines-theyre-all-bad-at-citing-news.php

Takeaways encoded in the skill: label support before hopping; a syndicated copy is never a root; expect dead ends and say so; report drift separately from falsity.

## Run-to-run variance (why one capture is one sample)

- Muck Rack (Oct 2026): the same prompt run twice on the same day repeated its cited domains 41% of the time on ChatGPT, 45% on Gemini, 94% on Perplexity.
- EWR Digital (Jan–Jun 2026, 12,778 answers): median domain-set Jaccard between repeats — Claude 0.42, Perplexity 0.33, Gemini 0.23, ChatGPT 0.13; an identical source set recurred in ≤2.5% of repeats on any engine. Error bars: 1 run ±33 pp, 3 runs ±18, 5 ±13, 7–8 ±10.
- Ahrefs (2025): consecutive AI Overviews share about 54% of their URLs.

Encoded: ask for three runs; quote rates only from seven or more; label a single capture as one observation.

## Signals on a page, in the order the skill uses them

| Signal | Where `trace.mjs` surfaces it | Worth | Source of the rule |
|---|---|---|---|
| Link inside the matching sentence or paragraph | `claim_matches[].best[].paragraph_links`, `attribution_sentences[].links` | Strongest: the author pointed at it | Fact-checkers' "trace to the original" step (SIFT; First Draft provenance pillar) |
| Attribution phrase with a named source | `attribution_sentences[].named_source` ("according to", "a study by", "first reported by", "told") | Strong when the named source can be found carrying the claim | Quotebank-style attribution patterns (Vaucher et al. 2021) |
| Canonical URL on another domain; "originally appeared"; "republished"; wire boilerplate `(Reuters)`, `/PRNewswire/`, `BUSINESS WIRE` | `canonical_host_mismatch`, `signals[]` of kind `syndication` / `wire` | The page is a copy; go to the original. Google stopped recommending canonical for syndication in 2023, so many copies are indexed and get cited instead of the original | Nicholls 2019 treats explicit wire credit as the marker of a wire article; Tow 2025 on cited copies |
| Wikipedia sentence reference | `wikipedia_references[]`, `claim_matches[].best[].wikipedia_refs` | The reference is the next hop, not a root. Wikipedia carries a list of 50+ documented "citogenesis" incidents where a wiki edit became the source of the press that later cited it | https://en.wikipedia.org/wiki/Wikipedia:List_of_citogenesis_incidents |
| Verbatim overlap with another page | `reuse` subcommand: 7-word shingles (precision) and 5-word (recall), longest common run, shared numbers, `churn_flag` at >20% overlap | Copy without credit; order the two pages by date to decide direction | Nicholls 2019 (5/7-gram thresholds); churnalism.com (2011) flagged "churn" above 20% overlap; Sunlight's SuperFastMatch scored overlap × density |
| Numeric match | `claim_matches[].best[].number_match`, `numbers_in_claim` | Cheapest hop signal: same number, unit, and entity. "14%" ≈ "14 per cent" ≈ "one in seven" needs the agent's reading | Trakkr found the figure on the cited page 82% of the time |
| Sponsored / partner / affiliate markers | `signals[]` of kind `sponsored` | Lowers the page's standing as a root; a sponsored page is brand-owned content wearing an outlet's masthead | — |
| No byline, no date, no outbound links, templated title ("27 X statistics for 2026") | `has_byline`, `has_date`, `external_link_count`, title | Content-farm fingerprint; not proof. Trakkr: company blogs were 54% of first citations and only 37% of those cited a source | NewsGuard tracks 3,000+ AI-generated "news" sites (2026) |
| Paywall / block / 404 | `paywall_suspected`, `status`, `hint` | Unreachable, not absent. Try `--browser`, then the Wayback snapshot; otherwise say so | Pew (2024): 38% of 2013 web pages were gone by 2023; 54% of Wikipedia articles carry at least one dead reference |

## Dates

On-page timestamps are unreliable for ordering: sites restamp for freshness, publish in the wrong time zone, or show the syndication date (Nicholls 2019 used crawler fetch time instead). Order of trust the skill uses: a primary document's own date → Wayback first capture (a lower bound: the page existed by then; absence of a capture proves nothing) → `article:published_time` / `datePublished` → visible dates. Do not claim an order inside a one-day gap. A "source" that is newer than the page citing it is a citogenesis flag.

## Root types, and the trap each one carries

- **original_reporting** — the reporter's own reporting. Trap: a headline that omits the company name makes it hard to find by search; "exclusively learned" and "obtained documents" are the tells.
- **primary_source** — release, filing, official statistics, court record, official statement. Trap: the wire mirror of a release is a copy of the primary, not the primary; the brand's newsroom or the regulator's site is the root.
- **study_or_paper** — research with a stated method. Trap: the number is real but the scope was lost on the way (Mehrabian's 7-38-55 applied to all communication; a vendor "survey" of its own customers presented as the market).
- **brand_own_site** — blog, marketing page, docs. Trap: a brand page that cites "a study" with no link is a dead end wearing a logo.
- **wikipedia** — a wiki sentence that predates everything it cites. Trap: the newspaper "source" copied the wiki.
- **forum_or_social** — Reddit, X, Quora, YouTube, LinkedIn, fan wikis. Trap: a fandom wiki is not Wikipedia.
- **content_farm_or_aggregator** — templated statistics sites, AI-generated farms, unsourced rewrites, mirror networks. Trap: an 80-item "sources" block at the bottom of the page that does not connect to the number.
- **listicle_or_review_site** — "best X" and comparison pages, review marketplaces. Trap: affiliate ranking presented as a finding.
- **unknown** — the trail ends. Not a failure of the trace; the finding is that the claim rests on nothing reachable.

Satire (The Onion, April Fools posts) is a flag on whichever type the page is, and a complete trace: the AI took a joke literally; nothing upstream needs finding.

## Further reading

- Lewis, Williams & Franklin (Cardiff, 2008): 60% of UK press articles were wholly or mainly wire or PR copy. https://orca.cardiff.ac.uk/id/eprint/18439/
- Sumner et al., BMJ 2014: when a university press release exaggerated, 58–86% of the news stories that followed did too; when it did not, 10–18%. Exaggeration enters at the release. https://doi.org/10.1136/bmj.g7015
- Leskovec, Backstrom & Kleinberg (2009), MemeTracker: quote variants cluster and news leads blogs by about 2.5 hours. https://snap.stanford.edu/memetracker/
- Wayback CDX API for first-capture dates. https://github.com/internetarchive/wayback/tree/master/wayback-cdx-server
- Ray, "The AI Slop Loop" (2026): a fabricated claim planted in AI-written SEO blogs reached Perplexity, ChatGPT and AI Overviews within a day. https://lilyraynyc.substack.com/p/the-ai-slop-loop
