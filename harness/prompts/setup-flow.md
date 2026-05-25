You are testing the installed Newsjack skills.

Create a Newsjack monitor profile for Fixture Coffee, a specialty coffee company.
Save it exactly at ARTIFACT_DIR/profile.json.

Use this JSON shape:

{
  "company": "Fixture Coffee",
  "website": "https://example.com",
  "description": "...",
  "topics": ["specialty coffee", "coffee roasting", "coffee shops"],
  "competitors": ["Blue Bottle Coffee", "Stumptown Coffee Roasters", "Intelligentsia Coffee"],
  "search_terms": ["specialty coffee", "coffee roasting", "coffee shops"],
  "feed_urls": ["https://news.google.com/rss/headlines/section/topic/BUSINESS?hl=en-US&gl=US&ceid=US:en"],
  "x_news": {"enabled": true},
  "x_trends": {"mode": "none", "woeids": [], "locations": []},
  "spokespeople": ["Founder", "Coffee sourcing lead"],
  "proof_assets": ["company website", "product pages"],
  "standing": ["specialty coffee", "coffee roasting"]
}

Then run the local newsjack detector in mock mode with that profile.
Save a short Markdown result at ARTIFACT_DIR/result.md.

Do not ask follow-up questions. Make reasonable assumptions.
