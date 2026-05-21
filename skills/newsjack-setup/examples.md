# Newsjack Setup Examples

## Chatbase-Style Profile

```json
{
  "profile": {
    "company": "Chatbase",
    "website": "https://chatbase.com",
    "description": "AI customer service platform for building and deploying AI support agents.",
    "topics": [
      "AI customer support",
      "AI support agents",
      "customer service automation",
      "AI chatbots for business"
    ],
    "competitors": [
      "Intercom",
      "Zendesk AI",
      "Ada",
      "Gorgias"
    ],
    "feed_urls": [
      "https://www.techmeme.com/feed.xml",
      "https://news.google.com/rss/headlines/section/topic/TECHNOLOGY?hl=en-US&gl=US&ceid=US:en",
      "https://news.google.com/rss/headlines/section/topic/BUSINESS?hl=en-US&gl=US&ceid=US:en"
    ],
    "spokespeople": [
      "Founder or CEO with AI customer support expertise",
      "Product lead for AI support agents"
    ],
    "proof_assets": [
      "Product pages",
      "customer support automation examples",
      "customer case studies",
      "AI agent analytics and escalation workflows"
    ],
    "standing": [
      "AI customer support",
      "AI agent deployment",
      "support automation",
      "customer experience workflows"
    ],
    "exclusions": []
  },
  "feed_rationale": [
    {
      "feed": "https://www.techmeme.com/feed.xml",
      "why": "High-signal tech and AI business stories where major platform/customer-support moves appear early."
    },
    {
      "feed": "https://news.google.com/rss/headlines/section/topic/TECHNOLOGY?hl=en-US&gl=US&ceid=US:en",
      "why": "Broader technology backstop for AI product launches and platform changes not surfaced by Techmeme."
    },
    {
      "feed": "https://news.google.com/rss/headlines/section/topic/BUSINESS?hl=en-US&gl=US&ceid=US:en",
      "why": "Catches enterprise software, M&A, funding, layoffs, and customer-experience business stories."
    }
  ],
  "run_commands": {
    "hourly_major_news": "python3 skills/newsjack-detector/scripts/newsjack_detector.py run --profile profile.json --feed-only --save --emit json",
    "profile_relevance": "python3 skills/newsjack-detector/scripts/newsjack_detector.py run \"AI customer support\" --profile profile.json --save --emit json"
  },
  "missing_inputs": [
    "Which customer proof or metrics can be used publicly?"
  ]
}
```
