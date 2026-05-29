# Reverse Eval Targets: 2026-05-28 US Google News

Seed surface: Google News Business and Technology, viewed through US VPN on 2026-05-28.

These are target stories for reverse-engineered company-profile evals. Canonical URLs should be captured during eval execution from primary/source pages or the Google News cluster.

## First Batch

| ID | Batch | Originating Story | Reverse Profile To Test | Expected Story Entities | Bias Tag |
|---|---|---|---|---|---|
| rev-20260528-001 | CPG / marketplace compliance | European Commission fines Temu $232M for illegal or unsafe product risks | Product safety compliance SaaS for marketplaces and importers | Temu, European Commission, illegal products, unsafe toys/electronics | same_source_assisted if Google News Business US is enabled |
| rev-20260528-002 | Healthtech / benefits | CVS restores coverage for Lilly's Zepbound / Foundayo | Employer benefits or PBM analytics company focused on GLP-1 access | CVS, Eli Lilly, Zepbound, Foundayo, obesity drugs | same_source_assisted if Google News Business US is enabled |
| rev-20260528-003 | SaaS / AI workforce | Wix lays off about 20% of staff, citing AI and exchange-rate pressure | Workforce planning SaaS or AI adoption consultancy for SMB software companies | Wix, layoffs, AI capabilities, exchange rates | same_source_assisted if Google News Business US or Techmeme is enabled |
| rev-20260528-004 | Fintech / compliance | Google engineer Michele Spagnuolo charged with using confidential data for Polymarket bets | Internal-data security or prediction-market compliance startup | Google, Michele Spagnuolo, Polymarket, insider trading, AlphaRaccoon | same_source_assisted if Google News Business US is enabled |
| rev-20260528-005 | SaaS / adtech | OpenAI prepares ChatGPT ads around conversational intent | AI search or conversational ad measurement platform | OpenAI, ChatGPT ads, conversational intent, smaller advertisers | same_source_assisted if Google News Business US is enabled |
| rev-20260528-006 | Cloud / data infra | Snowflake jumps after expanded $6B AWS partnership tied to enterprise AI demand | Cloud cost optimization or data-stack observability startup | Snowflake, AWS, Amazon, $6B partnership, enterprise AI | same_source_assisted if Google News Business US is enabled |
| rev-20260528-007 | Cybersecurity / open source | IBM and Red Hat commit $5B to open-source security / AI initiative | Open-source security posture management company | IBM, Red Hat, $5B, open source, AI, security | same_source_assisted if Google News Business US or Techmeme is enabled |
| rev-20260528-008 | Wearables / healthtech | Oura Ring 5 adds smaller form factor plus blood-pressure and sleep-disturbance signals | Remote patient monitoring or wearable health analytics company | Oura, Ring 5, blood pressure, sleep disturbance, wearable | same_source_assisted if Google News Technology US is enabled |
| rev-20260528-009 | Consumer hardware / supply chain | Steam Deck prices jump because of rising component costs | Hardware supply-chain analytics company or gaming-handheld accessory company | Valve, Steam Deck, component costs, memory, price increase | same_source_assisted if Google News Technology US or Techmeme is enabled |
| rev-20260528-010 | Privacy / cybersecurity | Websites can infer visitor behavior by analyzing SSD activity | Browser privacy or endpoint security company | SSD activity, browser API, side-channel attack, web tracking | same_source_assisted if Google News Technology US is enabled |
| rev-20260528-011 | AI governance / HR tech | Study finds racial disparities in AI hiring algorithms | AI audit or HR compliance platform | AI hiring tools, racial disparities, job applicants, bias | same_source_assisted if Google News Business US is enabled |
| rev-20260528-012 | Agentic fintech | Robinhood enables AI agents to trade and make purchases | Agent authorization or financial automation risk platform | Robinhood, AI agents, trading, credit card purchases | same_source_assisted if Google News Business US is enabled |
| rev-20260528-013 | Retail AI | AWS helps retailers build AI-powered shopping assistants | E-commerce personalization or retail AI tooling company | AWS, Amazon, AI shopping assistants, retailers | same_source_assisted if Google News Business US is enabled |
| rev-20260528-014 | Media / AI copyright | CNN sues Perplexity over alleged AI copyright infringement | Publisher licensing or AI content provenance company | CNN, Perplexity, copyright infringement, AI search | same_source_assisted if Google News Business US or Techmeme is enabled |
| rev-20260528-015 | EV / consumer auto | Rivian starts R2 SUV deliveries and order invites on June 9 | EV charging, fleet readiness, or auto retail analytics company | Rivian, R2, deliveries, order invites, June 9 | same_source_assisted if Google News Business US is enabled |

## Lower-Priority Or Excluded Seeds

| Story | Reason |
|---|---|
| Iran war, oil, Strait of Hormuz updates | Brand-safety risk and too broad for positive newsjack recall testing. Use only for rejection/brand-safety evals. |
| Inflation/PCE and live market updates | Broad macro story with noisy profile standing and weak single-origin identity. |
| National Hamburger Day deals | Low seriousness; useful only for local retail/CPG edge-case testing. |
| Mark Zuckerberg yacht coverage | Tabloid-shaped and weak client standing. |
| SpaceX/Tesla merger chatter | Noisy unless the reverse profile is narrowly about space/EV capital markets. |

## Suggested First Run

Start with:

1. Temu product safety.
2. CVS / Zepbound coverage.
3. Wix layoffs and AI.
4. Google / Polymarket insider trading.
5. OpenAI / ChatGPT ads.
6. Snowflake / AWS partnership.
7. Oura Ring 5.
8. CNN / Perplexity copyright lawsuit.

This covers CPG, healthtech, SaaS, fintech, adtech, cloud infra, wearables, and media/AI.
