# Angle Generator - Worked Examples

These examples show the pattern: realistic input, strict output, no padded angles. The examples are abbreviated only where repetition would teach nothing; keep the JSON shape intact when using the skill.

---

## Example 1: Series A Funding With Real Substance

### Before

```json
{
  "update": {
    "headline": "ForgeLedger raises $12M Series A",
    "facts": [
      "Round led by Foundry Climate; Sequoia participated",
      "Company sells carbon-accounting software to mid-market manufacturers",
      "127 paying customers, $2.4M ARR",
      "Founders are ex-Stripe (CEO) and ex-Watershed (CTO)",
      "Plans to hire 20 engineers in New York and Berlin"
    ],
    "links": [],
    "embargo": "none"
  },
  "company": {
    "name": "ForgeLedger",
    "one_liner": "Carbon accounting for mid-market manufacturers",
    "category": "climate-tech",
    "stage": "series-a",
    "geo": "USA + Germany",
    "prior_coverage": []
  },
  "profile": {
    "target_beats": ["climate-tech", "manufacturing", "venture capital"],
    "ban_list_outlets": [],
    "voice_notes": null
  },
  "context": {
    "current_time": "2026-05-18T10:00:00Z",
    "signal_from_newsjack_detector": null,
    "moments_from_story_calendar": null
  },
  "constraints": {
    "min_angles": 3,
    "max_angles": 7,
    "require_data_angle": false,
    "require_contrarian": false
  }
}
```

### After

```json
{
  "angles": [
    {
      "id": "a1-midmarket-manufacturing-gap",
      "headline_frame": "ForgeLedger raises $12M for carbon accounting in mid-market manufacturing",
      "story_type": "category-creation",
      "journalist_shape": {
        "beat_description": "Climate-tech reporter at a B2B trade or newsletter outlet covering industrial decarbonization for mid-market manufacturers.",
        "outlet_archetype": "Climate trade, manufacturing trade, or technical climate newsletter.",
        "evidence_they_care": "The angle is not the round amount; it is the buyer segment. A reporter on industrial decarbonization can test whether mid-market manufacturers have different reporting pain than enterprise ESG teams.",
        "do_not_target": "Consumer tech press, general startup blogs, broad enterprise ESG desks."
      },
      "why_now": "The Series A is live now; the broader mid-market compliance story has a longer window if ForgeLedger can prove customer demand.",
      "decay": {
        "stage": "week",
        "rationale": "Funding news is a 24hr event, but the segment thesis can support a week-long trend angle."
      },
      "distinctness_check": {
        "compared_to_other_angles_in_this_set": "This angle makes the mid-market manufacturing segment the story. The other kept angles focus on founder lineage and the investor thesis.",
        "compared_to_prior_coverage": null
      },
      "required_proof": [
        "One named mid-market manufacturer willing to describe the compliance pain",
        "Evidence that existing ESG tools overserve enterprise buyers",
        "Source for any regulatory-deadline claim before using it in a pitch"
      ],
      "anti_slop_pass": true,
      "facts_used": [
        "Round led by Foundry Climate; Sequoia participated",
        "Company sells carbon-accounting software to mid-market manufacturers",
        "127 paying customers, $2.4M ARR"
      ]
    },
    {
      "id": "a2-founder-lineage",
      "headline_frame": "Stripe and Watershed alumni are building climate software for factory floors",
      "story_type": "founder-profile",
      "journalist_shape": {
        "beat_description": "Founder-focused reporter at a VC newsletter or tech business outlet tracking operator-to-founder pipelines.",
        "outlet_archetype": "VC-adjacent newsletter or founder-profile desk.",
        "evidence_they_care": "Founder lineage is the entry point: ex-Stripe commercial discipline plus ex-Watershed climate domain credibility applied to manufacturers.",
        "do_not_target": "Manufacturing trade press that does not cover founder backstories."
      },
      "why_now": "The funding round gives the founder-profile angle a reason to run now.",
      "decay": {
        "stage": "24hr",
        "rationale": "Founder-lineage angles decay with the funding announcement unless tied to a broader reported trend."
      },
      "distinctness_check": {
        "compared_to_other_angles_in_this_set": "This angle makes the founders the protagonist, not the customer segment or the investor.",
        "compared_to_prior_coverage": null
      },
      "required_proof": [
        "Confirmed roles and dates at Stripe and Watershed",
        "Founder quote on why mid-market manufacturers were chosen",
        "At least one detail showing what the founders learned at prior companies"
      ],
      "anti_slop_pass": true,
      "facts_used": [
        "Founders are ex-Stripe (CEO) and ex-Watershed (CTO)",
        "Round led by Foundry Climate; Sequoia participated"
      ]
    },
    {
      "id": "a3-foundry-thesis",
      "headline_frame": "Foundry Climate's Series A bet says mid-market carbon accounting is not a back-office chore",
      "story_type": "funding-mechanics",
      "journalist_shape": {
        "beat_description": "VC reporter covering climate-tech fund theses, especially checks that go against the default enterprise-software narrative.",
        "outlet_archetype": "Venture trade, paid tech newsletter, or climate finance desk.",
        "evidence_they_care": "The investor is the news hook. A climate-finance reporter can use the round to examine whether investors see mid-market manufacturing as a distinct software market.",
        "do_not_target": "Local hiring reporters, product-review outlets, broad consumer business desks."
      },
      "why_now": "The round gives Foundry's thesis a timely peg, but the thesis must be stated on record.",
      "decay": {
        "stage": "week",
        "rationale": "Investor-thesis stories can run after the funding day if the partner supplies a real argument."
      },
      "distinctness_check": {
        "compared_to_other_angles_in_this_set": "This angle makes the investor's market bet the story. It is not another version of the company milestone.",
        "compared_to_prior_coverage": null
      },
      "required_proof": [
        "On-record Foundry partner quote explaining the mid-market manufacturing thesis",
        "Comparable recent climate-software rounds or market data supplied by the user",
        "Clarify whether Sequoia participated with capital, board role, or both"
      ],
      "anti_slop_pass": true,
      "facts_used": [
        "Round led by Foundry Climate; Sequoia participated",
        "Company sells carbon-accounting software to mid-market manufacturers",
        "127 paying customers, $2.4M ARR"
      ]
    }
  ],
  "refused_angles": [
    {
      "would_have_been": "ForgeLedger revolutionizes carbon accounting for manufacturers",
      "refusal_reason": "slop"
    },
    {
      "would_have_been": "ForgeLedger proves climate-tech funding is back",
      "refusal_reason": "hallucinated_fact"
    }
  ],
  "uncomfortable_questions": [
    "Are any of the 127 customers willing to be named? Without one, the strongest manufacturing angle becomes much thinner.",
    "Has Foundry actually said the mid-market thesis on record, or are you inferring it from the investment?",
    "How many of the 20 engineering hires are in Berlin versus New York? The regional hiring angle needs a real split."
  ],
  "follow_up_suggestions": {
    "next_skill": "journalist-fit-check",
    "rationale": "The output has three viable journalist shapes. Resolve names only after choosing the shape."
  }
}
```

### Why This Works

The output finds three different protagonists: customer segment, founders, and investor thesis. It refuses the slop and the unsupported market-wide claim instead of dressing them up.

---

## Example 2: Weak Product Update That Should Mostly Refuse

### Before

```json
{
  "update": {
    "headline": "Northstar Workflow launches v2",
    "facts": ["Faster than v1", "New UI"],
    "links": [],
    "embargo": "none"
  },
  "company": {
    "name": "Northstar Workflow",
    "one_liner": "Workflow software for B2B operations teams",
    "category": "B2B SaaS",
    "stage": "seed",
    "geo": "USA",
    "prior_coverage": []
  },
  "profile": {
    "target_beats": ["SaaS", "future of work"],
    "ban_list_outlets": [],
    "voice_notes": null
  },
  "context": {
    "current_time": "2026-05-18T10:00:00Z",
    "signal_from_newsjack_detector": null,
    "moments_from_story_calendar": null
  },
  "constraints": {
    "min_angles": 3,
    "max_angles": 7,
    "require_data_angle": false,
    "require_contrarian": false
  }
}
```

### After

```json
{
  "angles": [],
  "refused_angles": [
    {
      "would_have_been": "Northstar Workflow launches v2 with faster performance and a new UI",
      "refusal_reason": "no_journalist_shape"
    },
    {
      "would_have_been": "Northstar Workflow reshapes how operations teams manage work",
      "refusal_reason": "slop"
    },
    {
      "would_have_been": "Why Northstar Workflow v2 matters for the future of work",
      "refusal_reason": "hallucinated_fact"
    }
  ],
  "uncomfortable_questions": [
    "What specifically is faster: page load, task completion, report generation, or something else?",
    "By how much is it faster, measured over what baseline, and in whose environment?",
    "Why did the UI need replacing? Customer complaints, a new buyer, a new workflow, or a pivot?",
    "Who is already using v2, and can one customer describe what changed in their work?",
    "What can v2 do that v1 could not do? A new UI is a refresh. A new capability might be news."
  ],
  "follow_up_suggestions": {
    "next_skill": null,
    "rationale": "There is no pitchable angle yet. Get measurable proof, a customer, or a real capability change, then rerun angle-generator."
  }
}
```

### Why This Works

The skill does not pad to three angles. It tells the user exactly what facts would turn a product refresh into a story.

---

## Example 3: Newsjack-Detector Handoff With A Live Signal

### Before

```json
{
  "update": {
    "headline": "ClearGov AI helps public companies prepare AI-risk disclosures",
    "facts": [
      "ClearGov AI offers AI-governance software for public companies",
      "Tracks model usage, audit logs, and disclosure workflows",
      "47 public-company customers",
      "CEO can take calls today from 15:00-18:00 UTC"
    ],
    "links": [],
    "embargo": "none"
  },
  "company": {
    "name": "ClearGov AI",
    "one_liner": "AI governance for public companies",
    "category": "regtech",
    "stage": "series-b",
    "geo": "USA",
    "prior_coverage": []
  },
  "profile": {
    "target_beats": ["SEC", "AI governance", "public-company compliance"],
    "ban_list_outlets": [],
    "voice_notes": null
  },
  "context": {
    "current_time": "2026-05-18T14:00:00Z",
    "signal_from_newsjack_detector": {
      "signal_id": "sec-ai-disclosure-final-rule-2026-05-18",
      "headline": "SEC finalizes AI disclosure rule for public companies, effective 90 days from publication",
      "fetched_at": "2026-05-18T13:42:00Z",
      "published_at": "2026-05-18T13:00:00Z",
      "decay_stage": "4hr"
    },
    "moments_from_story_calendar": null
  },
  "constraints": {
    "min_angles": 3,
    "max_angles": 5,
    "require_data_angle": false,
    "require_contrarian": false
  }
}
```

### After

```json
{
  "angles": [
    {
      "id": "a1-sec-rule-reaction",
      "headline_frame": "SEC just finalized AI disclosure rules. ClearGov has 47 public-company customers in the affected market.",
      "story_type": "defensive-comment",
      "journalist_shape": {
        "beat_description": "Reporter writing same-day SEC rule reaction for a securities-law trade, governance newsletter, or business regulatory desk.",
        "outlet_archetype": "Regulatory trade publication or business press compliance desk.",
        "evidence_they_care": "The reporter needs credible company reaction within hours, and ClearGov has public-company customer exposure plus CEO availability today.",
        "do_not_target": "General AI product press; they will not care inside the 4hr regulatory-reaction window."
      },
      "why_now": "The SEC signal was published at 13:00 UTC and fetched at 13:42 UTC; the reaction window is live.",
      "decay": {
        "stage": "4hr",
        "rationale": "Direct handoff from newsjack-detector with a 4hr decay stage."
      },
      "distinctness_check": {
        "compared_to_other_angles_in_this_set": "This is the same-day reaction angle. The second angle is a slower implementation story.",
        "compared_to_prior_coverage": null
      },
      "required_proof": [
        "Link to final SEC rule text",
        "CEO available during the stated window",
        "One example of a disclosure workflow ClearGov tracks, without revealing confidential customer data"
      ],
      "anti_slop_pass": true,
      "facts_used": [
        "ClearGov AI offers AI-governance software for public companies",
        "47 public-company customers",
        "CEO can take calls today from 15:00-18:00 UTC",
        "SEC finalizes AI disclosure rule for public companies, effective 90 days from publication"
      ]
    },
    {
      "id": "a2-90-day-implementation",
      "headline_frame": "The 90-day AI disclosure clock starts now for public companies",
      "story_type": "trend",
      "journalist_shape": {
        "beat_description": "Enterprise-risk or governance reporter writing implementation pieces after the first wave of rule-reaction stories.",
        "outlet_archetype": "CFO, legal, governance, or risk-management trade outlet.",
        "evidence_they_care": "The story moves from what the rule says to what companies must inventory, log, and disclose before the deadline.",
        "do_not_target": "Breaking-news desks that only want immediate comment."
      },
      "why_now": "The rule creates a 90-day implementation clock; this angle has a slower but still timely window.",
      "decay": {
        "stage": "week",
        "rationale": "Implementation stories can run after the initial same-day reaction cycle."
      },
      "distinctness_check": {
        "compared_to_other_angles_in_this_set": "This angle is operational and deadline-driven, not a same-day reaction quote.",
        "compared_to_prior_coverage": null
      },
      "required_proof": [
        "Checklist of model-usage, audit-log, and disclosure-workflow steps",
        "One anonymized customer pattern if named customer approval is unavailable",
        "Legal review that the 90-day deadline is stated correctly"
      ],
      "anti_slop_pass": true,
      "facts_used": [
        "Tracks model usage, audit logs, and disclosure workflows",
        "47 public-company customers",
        "SEC finalizes AI disclosure rule for public companies, effective 90 days from publication"
      ]
    }
  ],
  "refused_angles": [
    {
      "would_have_been": "ClearGov AI is the future of public-company AI governance",
      "refusal_reason": "slop"
    }
  ],
  "uncomfortable_questions": [
    "Can the CEO actually take a call in the next two hours? If not, downgrade a1 from 4hr reaction to a week-long implementation angle.",
    "Can you share a customer example without violating confidentiality? Without it, reporters may treat this as vendor commentary only."
  ],
  "follow_up_suggestions": {
    "next_skill": "journalist-fit-check",
    "rationale": "The 4hr angle needs named reporters now, not tomorrow."
  }
}
```

### Why This Works

The live signal controls the first angle's decay. The skill keeps one fast reaction angle and one slower implementation angle instead of pretending every angle is breaking news.

---

## Example 4: Spray-And-Pray Partnership Request

### Before

```json
{
  "update": {
    "headline": "RelayDesk announces Zendesk partnership",
    "facts": [
      "RelayDesk integrates with Zendesk",
      "Both companies issued a joint press release"
    ],
    "links": [],
    "embargo": "none"
  },
  "company": {
    "name": "RelayDesk",
    "one_liner": "Workflow automation for support teams",
    "category": "B2B SaaS",
    "stage": "growth",
    "geo": "USA",
    "prior_coverage": []
  },
  "profile": {
    "target_beats": ["tech", "SaaS", "customer support", "AI", "future of work", "remote work", "startups"],
    "ban_list_outlets": [],
    "voice_notes": null
  },
  "context": {
    "current_time": "2026-05-18T10:00:00Z",
    "signal_from_newsjack_detector": null,
    "moments_from_story_calendar": null
  },
  "constraints": {
    "min_angles": 7,
    "max_angles": 7,
    "require_data_angle": false,
    "require_contrarian": false
  }
}
```

### After

```json
{
  "angles": [
    {
      "id": "a1-support-integration-mechanics",
      "headline_frame": "RelayDesk's Zendesk integration automates support triage without a developer",
      "story_type": "trend",
      "journalist_shape": {
        "beat_description": "Customer-experience trade reporter covering support operations and Zendesk's integration ecosystem.",
        "outlet_archetype": "CX trade publication or support-ops newsletter.",
        "evidence_they_care": "The only honest beat is support operations. The integration mechanics may matter if they change how teams triage tickets.",
        "do_not_target": "AI press, remote-work press, general startup press, future-of-work columnists."
      },
      "why_now": "The joint announcement is live today.",
      "decay": {
        "stage": "24hr",
        "rationale": "Standard partnership-news cycle."
      },
      "distinctness_check": {
        "compared_to_other_angles_in_this_set": "Only one angle survived. The other requested beats would be rephrasings or unsupported claims.",
        "compared_to_prior_coverage": null
      },
      "required_proof": [
        "A named customer using the integration",
        "What the integration automates that Zendesk alone does not",
        "Quote from a Zendesk product lead or marketplace owner"
      ],
      "anti_slop_pass": true,
      "facts_used": [
        "RelayDesk integrates with Zendesk",
        "Both companies issued a joint press release"
      ]
    }
  ],
  "refused_angles": [
    {
      "would_have_been": "RelayDesk and Zendesk shape the future of support",
      "refusal_reason": "slop"
    },
    {
      "would_have_been": "RelayDesk announces an innovative Zendesk integration",
      "refusal_reason": "slop"
    },
    {
      "would_have_been": "Why AI reporters should care about RelayDesk's Zendesk partnership",
      "refusal_reason": "hallucinated_fact"
    },
    {
      "would_have_been": "RelayDesk brings remote-work teams a better support workflow",
      "refusal_reason": "hallucinated_fact"
    },
    {
      "would_have_been": "What this partnership means for startups",
      "refusal_reason": "no_journalist_shape"
    },
    {
      "would_have_been": "The future-of-work angle on RelayDesk and Zendesk",
      "refusal_reason": "duplicate"
    }
  ],
  "uncomfortable_questions": [
    "You asked for seven angles. From these facts, one honest angle exists. The rest are the spray-and-pray pattern in different clothes.",
    "What does the integration do that Zendesk's own automation cannot do?",
    "Do you have usage data: tickets triaged, time saved, deflection rate, or implementation time?",
    "Can a customer or Zendesk product lead speak on record?"
  ],
  "follow_up_suggestions": {
    "next_skill": "meanest-editor",
    "rationale": "Draft one tight pitch around the support-ops angle. Do not pad the beat list."
  }
}
```

### Why This Works

The output is quietly hostile to the volume request. It gives the one defensible angle and shows exactly why the other six would be inbox spam.
