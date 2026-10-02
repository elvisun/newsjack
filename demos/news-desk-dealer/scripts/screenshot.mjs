// Screenshots every headline's page through ScrapingBee and writes small JPEG
// thumbnails to public/shots/<id>.jpg plus a manifest at data/shots.json.
// Usage: SCRAPINGBEE_API_KEY=... node scripts/screenshot.mjs [--limit N] [--concurrency 4]
// Google News links are redirect pages; ScrapingBee renders JS so the redirect
// is followed before the shot is taken. Each shot costs ScrapingBee credits.

import fs from "node:fs/promises";
import path from "node:path";
import sharp from "sharp";

const key = process.env.SCRAPINGBEE_API_KEY;
if (!key) { console.error("SCRAPINGBEE_API_KEY is not set"); process.exit(1); }
const arg = (n, d) => { const i = process.argv.indexOf(n); return i > -1 ? process.argv[i + 1] : d; };
const limit = +arg("--limit", Infinity);
const concurrency = +arg("--concurrency", 4);

const doc = JSON.parse(await fs.readFile("data/headlines.json", "utf8"));
const outDir = "public/shots";
await fs.mkdir(outDir, { recursive: true });
const manifestPath = "data/shots.json";
const manifest = JSON.parse(await fs.readFile(manifestPath, "utf8").catch(() => '{"ids":[]}'));
const done = new Set(manifest.ids);

const todo = doc.headlines.filter((h) => !done.has(h.id)).slice(0, limit);
console.error(`${todo.length} to shoot, ${done.size} already done`);

async function shoot(h) {
  const params = new URLSearchParams({
    api_key: key, url: h.url, screenshot: "true", render_js: "true",
    window_width: "1024", window_height: "768", block_ads: "true", wait: "1500", timeout: "20000",
  });
  const res = await fetch(`https://app.scrapingbee.com/api/v1/?${params}`);
  if (!res.ok) throw new Error(`${res.status} ${(await res.text()).slice(0, 120)}`);
  const png = Buffer.from(await res.arrayBuffer());
  await sharp(png).resize({ width: 360, height: 270, fit: "cover", position: "top" }).jpeg({ quality: 72 }).toFile(path.join(outDir, `${h.id}.jpg`));
}

let i = 0, ok = 0, fail = 0;
await Promise.all(Array.from({ length: concurrency }, async () => {
  while (i < todo.length) {
    const h = todo[i++];
    try { await shoot(h); done.add(h.id); ok++; console.error(`ok   ${h.id} ${h.source}`); }
    catch (e) { fail++; console.error(`fail ${h.id} ${h.source}: ${e.message}`); }
    await fs.writeFile(manifestPath, JSON.stringify({ ids: [...done].sort() }, null, 2));
  }
}));
console.error(`done: ${ok} ok, ${fail} failed, ${done.size} total`);
