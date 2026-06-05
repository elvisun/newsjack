# news-search

Search current news for a topic, company, competitor, or hook and return dated, attributed articles other Newsjack skills can trust.

This skill prefers the optional Medialyst MCP news index, which returns normalized publication metadata (outlet, author, `published_at`, canonical URL) that downstream skills depend on. General web search ranks for SEO over recency and rarely exposes a reliable publication timestamp, so it is a weaker source for news. When Medialyst is not configured, the skill falls back to host web/browser search and flags the reduced freshness confidence rather than guessing dates.

Medialyst is optional. New accounts get 300 free credits (~3,000 news searches) — see [medialyst.ai/agents](https://medialyst.ai/agents). The skill stays useful without it.
