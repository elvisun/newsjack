// Pulls today's headlines from Google News topic feeds (+ Techmeme) and writes
// data/headlines.json. Dependency-free: fetch + a small RSS item parser.
// Usage: node scripts/ingest.mjs [--out data/headlines.json]

const FEEDS = [
  ["technology", "https://news.google.com/rss/headlines/section/topic/TECHNOLOGY?hl=en-US&gl=US&ceid=US:en"],
  ["business", "https://news.google.com/rss/headlines/section/topic/BUSINESS?hl=en-US&gl=US&ceid=US:en"],
  ["science", "https://news.google.com/rss/headlines/section/topic/SCIENCE?hl=en-US&gl=US&ceid=US:en"],
  ["health", "https://news.google.com/rss/headlines/section/topic/HEALTH?hl=en-US&gl=US&ceid=US:en"],
  ["us", "https://news.google.com/rss/headlines/section/topic/NATION?hl=en-US&gl=US&ceid=US:en"],
  ["world", "https://news.google.com/rss/headlines/section/topic/WORLD?hl=en-US&gl=US&ceid=US:en"],
  ["techmeme", "https://www.techmeme.com/feed.xml"],
];

const out = process.argv.includes("--out")
  ? process.argv[process.argv.indexOf("--out") + 1]
  : "data/headlines.json";

const decode = (s = "") =>
  s
    .replace(/<!\[CDATA\[([\s\S]*?)\]\]>/g, "$1")
    .replace(/&lt;/g, "<").replace(/&gt;/g, ">").replace(/&quot;/g, '"')
    .replace(/&#39;|&apos;/g, "'").replace(/&amp;/g, "&").replace(/&nbsp;/g, " ")
    .replace(/<[^>]+>/g, "").replace(/\s+/g, " ").trim();

const tag = (xml, name) => {
  const m = xml.match(new RegExp(`<${name}(?:\\s[^>]*)?>([\\s\\S]*?)</${name}>`));
  return m ? decode(m[1]) : "";
};

function parseItems(xml, feed) {
  const items = [];
  for (const m of xml.matchAll(/<item>([\s\S]*?)<\/item>/g)) {
    const x = m[1];
    let title = tag(x, "title");
    let source = tag(x, "source");
    const sourceUrl = x.match(/<source[^>]*url="([^"]+)"/)?.[1] ?? "";
    // Google News titles end with " - Source"; strip it when we know the source.
    if (source && title.endsWith(` - ${source}`)) title = title.slice(0, -(source.length + 3));
    if (!source) {
      const dash = title.lastIndexOf(" - ");
      if (dash > 20) { source = title.slice(dash + 3); title = title.slice(0, dash); }
    }
    items.push({
      title,
      url: tag(x, "link") || (x.match(/<link>([\s\S]*?)<\/link>/)?.[1] ?? "").trim(),
      source: source || (feed === "techmeme" ? "Techmeme" : ""),
      domain: (() => { try { return new URL(sourceUrl || (feed === "techmeme" ? "https://www.techmeme.com" : "")).hostname.replace(/^www\./, ""); } catch { return ""; } })(),
      published_at: new Date(tag(x, "pubDate") || Date.now()).toISOString(),
      excerpt: tag(x, "description").slice(0, 280),
      feed,
    });
  }
  return items;
}

const norm = (t) => t.toLowerCase().replace(/[^a-z0-9 ]/g, "").split(" ").filter((w) => w.length > 3).slice(0, 8).join(" ");

const all = [];
for (const [feed, url] of FEEDS) {
  try {
    const res = await fetch(url, { headers: { "user-agent": "news-desk-dealer/0.1" } });
    const xml = await res.text();
    const items = parseItems(xml, feed);
    console.error(`${feed}: ${items.length}`);
    all.push(...items);
  } catch (e) {
    console.error(`${feed}: failed (${e.message})`);
  }
}

const seen = new Set();
const headlines = [];
for (const h of all) {
  if (!h.title) continue;
  const key = norm(h.title);
  if (seen.has(key)) continue;
  seen.add(key);
  headlines.push({ id: `h${String(headlines.length + 1).padStart(3, "0")}`, ...h });
}
headlines.sort((a, b) => (a.published_at < b.published_at ? 1 : -1));

const doc = { pulled_at: new Date().toISOString(), feeds: FEEDS.map(([f]) => f), count: headlines.length, headlines };
await import("node:fs/promises").then((fs) => fs.writeFile(out, JSON.stringify(doc, null, 2)));
console.error(`wrote ${headlines.length} headlines to ${out}`);
