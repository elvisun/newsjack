#!/usr/bin/env node
// press-clip clipper — render a live article to a high-fidelity PDF press clip.
//
// Keeps the publication's real look (logo, fonts, photos, layout). Isolates the
// article STRUCTURALLY — keep the article and its ancestor chain, drop everything
// that is a sibling of that chain (header, nav, sidebar, footer, recirc) — so no
// site-specific class names are baked in. Per-publisher junk that lives INSIDE the
// article (ads, sponsored rails, newsletter boxes, comment embeds) is removed at
// runtime via --drop, which the caller supplies after inspecting the page.
// Optionally isolates one section (the client's) in a roundup. Stamps the outlet's
// logo at the top as the trust signal. Contains no per-site logic by design.
//
// Usage:
//   node clip.mjs --url <URL> --out <file.pdf> [options]
//
// Options:
//   --section "<Heading>"  Isolate the roundup section whose heading names this
//                          client: keep the lead + that section, drop the rest.
//                          Omit for single-subject articles (keep whole article).
//   --preview <file.png>   Also write a full-page screenshot to verify. Pages taller
//                          than Chrome's capture limit are captured at 1x, or cut
//                          at 16,000 px with a warning; content is never repeated.
//   --chrome <path>        Browser executable (defaults to system Chrome/Chromium).
//   --keep "<sel,sel>"     Extra CSS selectors to force-keep (never remove).
//   --drop "<sel,sel>"     Extra CSS selectors to remove (site-specific junk).
//   --root "<selector>"    Force the article container instead of auto-detecting it.
//                          Escape hatch for templates the heuristic picks wrong —
//                          tailor at runtime here rather than forking the script.
//   --logo "<url>"         Outlet logo image URL to stamp at the top of the clip.
//                          Overrides auto-detection — pass it when the article page
//                          has no masthead logo (grab one from the outlet's home page).
//
// Animation assets (for coverage-reel and other motion uses):
//   --mention "<term>"     The client's name or an alias to find in the article.
//                          Repeat the flag or pass a comma list. Matching is
//                          case-sensitive and prefers the longest term. A single
//                          capitalised word that is also a common word ("Nothing")
//                          only counts when it is not the first word of a sentence.
//   --assets <dir>         Write animation-ready captures plus a clip.json sidecar
//                          (page metadata, logo, captures, ranked mentions with
//                          rectangles, warnings) into this folder.
//   --scale 2|3            Device scale for the asset captures (default 2). Zooming
//                          into a mention needs at least 2; coverage-reel uses 3.
//   --pick <n>             Capture the n-th ranked mention instead of the best one.
//
// Every clip carries the outlet logo: it is the key trust signal. When the article
// page has no masthead logo, the script automatically opens the home page to find
// one; --logo lets the caller supply it explicitly. The console line reports which
// source the logo came from so a reviewer can confirm a real logo (not a text fallback).
// The logo is stamped exactly as the site renders it: its own colours, on a plate of
// the header colour it sat on, with an animated logo frozen on its fullest frame.
//
// Generic capture safeguards (no site names): the article root is pinned BEFORE any
// scrolling, from the <h1> that matches the page's own headline metadata; same-site
// scripts can no longer fetch whole next articles as HTML or rewrite the URL; fixed
// and sticky overlays outside the article are hidden; lazy images are loaded eagerly
// and decoded; animations are frozen. A page that serves a bot challenge instead of
// the article stops the run with exit code 3: ask the user for their own PDF or
// screenshot instead. The script never tries to get around a block or a paywall.
//
// Requires a Chromium-based browser on the machine and the `playwright-core`
// npm package (npm i playwright-core). Edge works as the browser too.

// Resolve playwright-core from the user's working directory (where they ran
// `npm i playwright-core`) or the global npm root, not just next to this script.
import { createRequire } from 'module';
import { pathToFileURL, fileURLToPath } from 'url';
import { execSync } from 'child_process';
import { mkdirSync, unlinkSync, writeFileSync } from 'fs';
import { join, resolve } from 'path';

export const TOOL_VERSION = 'press-clip/clip.mjs 2';
export const SIDECAR_VERSION = 1;
// Chromium silently repeats a capture's content past 16,384 device px. Asset tiles stay
// far below that so a decoded tile is also kind to the video compositor's memory.
export const MAX_TILE_DEVICE_PX = 8000;
export const EXIT_BLOCKED = 3;

function loadChromium() {
  const bases = [process.cwd() + '/', process.env.PRESS_CLIP_MODULES ? process.env.PRESS_CLIP_MODULES + '/' : null].filter(Boolean);
  try { bases.push(execSync('npm root -g').toString().trim() + '/../'); } catch {}
  for (const base of bases) {
    try { return createRequire(pathToFileURL(base))('playwright-core').chromium; } catch {}
  }
  console.error('Could not find playwright-core. Run:  npm i playwright-core   (in this folder), then retry.');
  process.exit(1);
}

// Ad / recirculation / comment / video-widget networks. Blocking them at the request level stops
// the whole class of lazy-injected junk (display ads, sponsored rails, "around the web", comment
// embeds, autoplay video) from ever loading — generically, by domain, with no per-site selectors.
// The article's own text and images are first-party and load normally.
const BLOCK_HOSTS = /(doubleclick|googlesyndication|googletagservices|google-analytics|googletagmanager|adservice\.google|amazon-adsystem|adsystem|taboola|outbrain|zergnet|connatix|spot\.im|openweb|disqus|criteo|pubmatic|rubiconproject|adnxs|moatads|scorecardresearch|zemanta|sharethrough|teads|indexww|casalemedia|3lift|districtm|smartadserver|yieldmo|sailthru|piano\.io|permutive|chartbeat|parsely|nativo|bidswitch|adlightning|confiant|exco|primis|anyclip|jwpltx)\./i;

// ---------------------------------------------------------------------------
// Pure helpers (exported for tests; no browser needed)
// ---------------------------------------------------------------------------

export function parseArgs(argv) {
  const args = { mention: [] };
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (!a.startsWith('--')) continue;
    const key = a.slice(2);
    const value = argv[++i];
    if (key === 'mention') args.mention.push(value ?? '');
    else args[key] = value;
  }
  return args;
}

// Comma lists and repeated flags both work. Longest first, so a product name wins over
// the bare company name that it starts with.
export function parseMentionTerms(values) {
  const terms = [];
  for (const value of [values].flat()) {
    for (const part of String(value ?? '').split(',')) {
      const term = part.replace(/\s+/g, ' ').trim();
      if (term && !terms.includes(term)) terms.push(term);
    }
  }
  return terms.sort((a, b) => b.length - a.length || a.localeCompare(b));
}

// A single capitalised word ("Nothing", "Apple", "Notion") doubles as an ordinary word at the
// start of a sentence or in a title-case heading. Multi-word names, all-caps names and names
// with digits or inner capitals ("OnePlus", "4a") are distinctive on their own.
export function isAmbiguousTerm(term) {
  return /^\p{Lu}\p{Ll}+$/u.test(term);
}

// Score one located hit. Body sentences beat headlines, links, captions, recirculation
// modules and navigation. Returns { score, reasons, rejected }.
export function scoreMention(hit) {
  const reasons = [];
  let score = 0;
  const add = (points, why) => { score += points; reasons.push(`${points > 0 ? '+' : ''}${points} ${why}`); };
  if (hit.ambiguous && (hit.sentenceInitial || hit.titleCaseHeading)) {
    return { score: -100, reasons: ['rejected: ambiguous common word used at a sentence start or in a title-case heading'], rejected: true };
  }
  if (hit.inAside) add(-4, 'inside an aside, nav or footer');
  if (hit.linkDense) add(-5, 'link-heavy block (related stories or navigation)');
  else if (hit.inLink) add(hit.sameSiteLink ? -5 : -2, hit.sameSiteLink ? 'inside a link to another page on the site' : 'inside a link');
  if (hit.inCaption) add(-2, 'in a caption');
  if (hit.inHeadline) add(-3, 'in the headline (shown separately)');
  else if (hit.inHeading) add(-2, 'in a subheading');
  if (hit.block === 'P') add(3, 'body paragraph');
  else if (hit.block === 'LI' || hit.block === 'BLOCKQUOTE') add(1, 'list item or quote');
  if (hit.sectionScoped && !hit.inSection) add(-4, 'outside the requested section');
  if (hit.paragraphIndex >= 0 && hit.paragraphIndex < 5) add(1, 'in the first five paragraphs');
  else if (hit.paragraphIndex >= 15) add(-1, 'deep in the article');
  const len = (hit.sentence || '').length;
  if (len >= 60 && len <= 240) add(1, 'readable sentence length');
  if (len > 0 && len < 30) add(-2, 'very short sentence');
  if (/\s/.test(hit.term || '')) add(2, 'full product or company name');
  else if (hit.ambiguous) add(-1, 'bare name that is also a common word');
  if (hit.attributedQuote) add(1, 'attributed quote');
  return { score, reasons, rejected: false };
}

export function rankMentions(hits) {
  const ranked = hits
    .map((hit, order) => ({ ...hit, order, ...scoreMention(hit) }))
    .filter((hit) => !hit.rejected)
    .sort((a, b) => b.score - a.score || a.order - b.order);
  // one entry per sentence: a sentence that names the client twice is still one candidate
  const seen = new Set();
  return ranked.filter((hit) => {
    const key = `${hit.paragraphIndex}|${hit.sentence}`;
    if (seen.has(key)) return false;
    seen.add(key);
    return true;
  });
}

// Merge the rectangles of one text range into one bar per rendered line.
export function mergeLineRects(rects) {
  const lines = [];
  for (const r of [...rects].filter((q) => q.w > 1 && q.h > 1).sort((a, b) => a.y - b.y || a.x - b.x)) {
    const line = lines.find((l) => Math.min(l.y + l.h, r.y + r.h) - Math.max(l.y, r.y) > 0.5 * Math.min(l.h, r.h));
    if (line) {
      const x0 = Math.min(line.x, r.x), y0 = Math.min(line.y, r.y);
      const x1 = Math.max(line.x + line.w, r.x + r.w), y1 = Math.max(line.y + line.h, r.y + r.h);
      Object.assign(line, { x: x0, y: y0, w: x1 - x0, h: y1 - y0 });
    } else {
      lines.push({ ...r });
    }
  }
  return lines.sort((a, b) => a.y - b.y).map(roundRect);
}

export function roundRect(r) {
  const q = (n) => Math.round(n * 100) / 100;
  return { x: q(r.x), y: q(r.y), w: q(r.w), h: q(r.h) };
}

export function unionRect(rects) {
  if (!rects.length) return null;
  const x0 = Math.min(...rects.map((r) => r.x)), y0 = Math.min(...rects.map((r) => r.y));
  const x1 = Math.max(...rects.map((r) => r.x + r.w)), y1 = Math.max(...rects.map((r) => r.y + r.h));
  return roundRect({ x: x0, y: y0, w: x1 - x0, h: y1 - y0 });
}

// Split [y0, y1) into tiles no taller than MAX_TILE_DEVICE_PX device pixels.
export function planStrips(y0, y1, scale, maxDevicePx = MAX_TILE_DEVICE_PX) {
  const tileCss = Math.floor(maxDevicePx / scale);
  const tiles = [];
  for (let top = Math.floor(y0); top < y1; top += tileCss) {
    tiles.push({ y: top, h: Math.min(tileCss, Math.ceil(y1) - top) });
  }
  return tiles;
}

// Convert page-space CSS rects into an asset's pixel space; drop rects that miss the asset.
export function toAssetRects(rects, asset) {
  const out = [];
  for (const r of rects) {
    const x0 = Math.max(r.x, asset.rect.x), y0 = Math.max(r.y, asset.rect.y);
    const x1 = Math.min(r.x + r.w, asset.rect.x + asset.rect.w), y1 = Math.min(r.y + r.h, asset.rect.y + asset.rect.h);
    if (x1 <= x0 || y1 <= y0) continue;
    out.push(roundRect({ x: (x0 - asset.rect.x) * asset.scale, y: (y0 - asset.rect.y) * asset.scale, w: (x1 - x0) * asset.scale, h: (y1 - y0) * asset.scale }));
  }
  return out;
}

const CHALLENGE_TITLE = /just a moment|attention required|access denied|are you a robot|verify (?:that )?you are (?:a )?human|checking (?:your|the) browser|pardon our interruption|request unsuccessful|security check|bot (?:check|detection)|please enable (?:js|javascript) and cookies/i;

// Decide whether the page is a bot-protection interstitial rather than the article.
export function detectBlock({ status = 200, title = '', textLength = 0, hasH1 = true, challengeMarkup = false }) {
  const challengeTitle = CHALLENGE_TITLE.test(title);
  const badStatus = [401, 403, 429, 503].includes(status);
  if (challengeTitle && (badStatus || textLength < 2000 || !hasH1)) return { blocked: true, reason: `challenge page ("${title.trim().slice(0, 60)}")` };
  if (challengeMarkup && (badStatus || !hasH1)) return { blocked: true, reason: 'bot-challenge markup on the page' };
  if (badStatus && (textLength < 600 || !hasH1)) return { blocked: true, reason: `HTTP ${status} with no article content` };
  return { blocked: false, reason: '' };
}

// Full-page previews must never repeat content past the device-pixel capture limit.
export function pickPreviewMode(cssHeight, dpr, limit = 16000) {
  if (cssHeight * dpr <= limit) return { scale: 'device', clipHeight: null, warning: null };
  if (cssHeight <= limit) return { scale: 'css', clipHeight: null, warning: { code: 'preview_downscaled', detail: `Page is ${Math.round(cssHeight)} px tall; preview captured at 1x to stay under Chrome's capture limit.` } };
  return { scale: 'css', clipHeight: limit, warning: { code: 'preview_truncated', detail: `Page is ${Math.round(cssHeight)} px tall; preview stops at ${limit} px. The PDF has the full article.` } };
}

export function pngSize(buffer) {
  if (buffer.length < 24 || buffer.toString('latin1', 12, 16) !== 'IHDR') return null;
  return { width: buffer.readUInt32BE(16), height: buffer.readUInt32BE(20) };
}

// ---------------------------------------------------------------------------
// In-page steps (each runs inside the page via page.evaluate / addInitScript)
// ---------------------------------------------------------------------------

// Infinite-scroll templates rewrite the address bar to the next article as you scroll.
// Keep the path pinned so the page never re-renders itself as a different story.
function pinPathInit() {
  const pin = location.pathname;
  for (const k of ['pushState', 'replaceState']) {
    const orig = history[k].bind(history);
    history[k] = (state, title, url) => {
      try { if (url && new URL(url, location.href).pathname !== pin) return undefined; } catch {}
      return orig(state, title, url);
    };
  }
}

function pageProbe() {
  const markers = '#challenge-form, #challenge-running, #cf-challenge-running, .cf-browser-verification, [id^="cf-chl"], iframe[src*="challenges.cloudflare.com"], iframe[src*="captcha-delivery.com"], #px-captcha, iframe[src*="hcaptcha.com"], iframe[src*="recaptcha"]';
  return {
    title: document.title || '',
    textLength: (document.body?.innerText || '').trim().length,
    hasH1: !!document.querySelector('h1'),
    challengeMarkup: !!document.querySelector(markers),
  };
}

// Page metadata, read never inferred. Each field is { value, source } or { value: null, source: null }.
function readPageMeta() {
  const decoder = document.createElement('textarea');
  const clean = (s) => { decoder.innerHTML = s == null ? '' : String(s); return decoder.value.replace(/\s+/g, ' ').trim(); };
  const out = {};
  const put = (k, v, source) => { const value = clean(v); if (value && !out[k]) out[k] = { value, source }; };
  const types = new Set(['NewsArticle', 'Article', 'ReportageNewsArticle', 'AnalysisNewsArticle', 'OpinionNewsArticle', 'ReviewNewsArticle', 'BlogPosting', 'LiveBlogPosting', 'Review', 'TechArticle', 'BackgroundNewsArticle']);
  const nodes = [...document.querySelectorAll('script[type="application/ld+json"]')].flatMap((s) => {
    try { const j = JSON.parse(s.textContent); return [j].flat().flatMap((o) => (o && o['@graph'] ? o['@graph'] : [o])); } catch { return []; }
  }).filter((o) => o && typeof o === 'object');
  const art = nodes.find((o) => [o['@type']].flat().some((t) => types.has(t)));
  let paywalled = false;
  if (art) {
    const pub = art.publisher && typeof art.publisher === 'object'
      ? (art.publisher.name ? art.publisher : nodes.find((o) => o['@id'] && o['@id'] === art.publisher['@id']))
      : null;
    put('outlet', pub && pub.name, 'json-ld');
    put('headline', art.headline, 'json-ld');
    const authors = [art.author].flat().filter(Boolean)
      .map((a) => (typeof a === 'string' ? a : a.name))
      .map((name) => clean(name).replace(/\S+@\S+\.\S+/g, '').replace(/\s+/g, ' ').trim())
      .filter((name) => name && !/^https?:/i.test(name));
    if (authors.length) put('byline', [...new Set(authors)].join(', '), 'json-ld');
    put('published_at', art.datePublished, 'json-ld');
    paywalled = art.isAccessibleForFree === false || String(art.isAccessibleForFree).toLowerCase() === 'false';
  }
  const meta = (sel) => document.querySelector(sel)?.getAttribute('content');
  put('outlet', meta('meta[property="og:site_name"]'), 'og');
  put('headline', meta('meta[property="og:title"]'), 'og');
  put('published_at', meta('meta[property="article:published_time"]'), 'og');
  const ogAuthor = meta('meta[property="article:author"]');
  if (ogAuthor && !/^https?:/i.test(ogAuthor)) put('byline', ogAuthor, 'og');
  put('outlet', meta('meta[name="application-name"]'), 'meta');
  put('headline', meta('meta[name="parsely-title"]') || meta('meta[name="twitter:title"]'), 'meta');
  put('byline', meta('meta[name="author"]') || meta('meta[name="parsely-author"]'), 'meta');
  put('published_at', meta('meta[name="parsely-pub-date"]') || meta('meta[name="sailthru.date"]') || meta('meta[name="date"]') || meta('meta[itemprop="datePublished"]'), 'meta');
  const h1 = document.querySelector('[data-pc-root] h1') || document.querySelector('h1');
  const visibleHeadline = h1 ? clean(h1.textContent) : '';
  put('byline', document.querySelector('[rel="author"]')?.innerText, 'visible');
  put('published_at', document.querySelector('[data-pc-root] time[datetime], time[datetime]')?.getAttribute('datetime'), 'visible');
  const field = (k) => out[k] || { value: null, source: null };
  return {
    outlet: field('outlet'),
    headline: field('headline'),
    byline: field('byline'),
    published_at: field('published_at'),
    visible_headline: visibleHeadline || null,
    canonical_url: document.querySelector('link[rel="canonical"]')?.href || null,
    paywalled,
    headline_candidates: [out.headline?.value, meta('meta[property="og:title"]'), document.title].filter(Boolean),
  };
}

// Choose the article container STRUCTURALLY, never by first-match — first-match let a
// 115-char footer recirc card classed "post-content" beat the 2,257-char real body. Gather
// every plausible candidate, drop any that lives inside site chrome (header/nav/footer/aside
// is by definition not the story — pure structure, no class guessing), then pick the TIGHTEST
// container that still holds the headline and most of the page text. That beats both a small
// teaser (fails the text bar) and a loose <main> (drags nav/comments back in). All generic.
// It runs BEFORE any scrolling, seeded with the <h1> that matches the page's own headline
// metadata, so infinite-scroll templates cannot swap in the next article first. Anything
// article-shaped that a script inserts outside the pinned root afterwards is removed.
function pinRoot({ root, headlines }) {
  const norm = (s) => (s || '').replace(/\s+/g, ' ').trim().toLowerCase();
  const wanted = (headlines || []).map(norm).filter((h) => h.length >= 8);
  const h1s = [...document.querySelectorAll('h1')];
  const h1 = h1s.find((h) => wanted.includes(norm(h.textContent)))
    || h1s.find((h) => { const t = norm(h.textContent); return t.length >= 12 && wanted.some((w) => w.includes(t) || t.includes(w)); })
    || h1s.find((h) => h.getBoundingClientRect().height > 0)
    || h1s[0] || null;
  let picked = null;
  // explicit override wins — the --root escape hatch, same runtime-tailoring rule as --drop/--keep
  if (root) { try { picked = document.querySelector(root); } catch {} }
  if (!picked) {
    // include WordPress post wrappers ([id^=post-], class token "post") — a huge slice of the web
    // whose ideal root (the .post div holding headline + byline + body) none of the others match.
    const SEL = ['article', '[class*="article-body" i]', '[class*="article-content" i]',
      '[class*="post-content" i]', '[class*="entry-content" i]', '[id^="post-" i]', '[class~="post"]', 'main'];
    let cands = [];
    SEL.forEach((s) => { try { document.querySelectorAll(s).forEach((el) => cands.push(el)); } catch {} });
    cands = [...new Set(cands)].filter((el) => !el.closest('header, nav, footer, aside'));
    if (!cands.length) picked = document.body;
    else {
      const txt = (el) => (el.innerText || '').trim().length;
      const withHeadline = h1 ? cands.filter((el) => el.contains(h1)) : cands;
      const pool = withHeadline.length ? withHeadline : cands;
      const maxText = Math.max(...pool.map(txt));
      const solid = pool.filter((el) => txt(el) >= 0.6 * maxText);
      picked = solid.length ? solid.sort((a, b) => txt(a) - txt(b))[0]   // tightest that qualifies
        : pool.sort((a, b) => txt(b) - txt(a))[0];                       // fallback: most text
      // some templates put the headline in a hero block beside the <article>: climb to the
      // nearest ancestor that holds both, as long as that adds little beyond the hero
      if (h1 && !picked.contains(h1)) {
        let lca = picked.parentElement;
        while (lca && !lca.contains(h1)) lca = lca.parentElement;
        if (lca && lca !== document.body && lca !== document.documentElement && txt(lca) <= 1.6 * txt(picked)) picked = lca;
      }
    }
  }
  picked.setAttribute('data-pc-root', '');
  if (h1 && picked.contains(h1)) h1.setAttribute('data-pc-h1', '');
  new MutationObserver((records) => records.forEach((m) => m.addedNodes.forEach((n) => {
    if (n.nodeType !== 1 || picked.contains(n) || n.contains(picked) || n.closest?.('.pc-header')) return;
    if (n.matches?.('article') || n.querySelector?.('h1, article')) n.remove();
  }))).observe(document.body, { childList: true, subtree: true });
  return { h1Text: h1 ? norm(h1.textContent) : null, h1Count: h1s.length, tag: picked.tagName.toLowerCase() };
}

// Lazy images: beyond-viewport captures do not trigger lazy loaders, so make the article's
// images eager, promote data-* sources, nudge loaders by scrolling only within the pinned
// article (next-article loading is already blocked), then wait for decode.
async function loadArticleImages() {
  const root = document.querySelector('[data-pc-root]') || document.body;
  for (const img of root.querySelectorAll('img')) {
    img.loading = 'eager'; img.decoding = 'sync';
    for (const [from, to] of [['data-src', 'src'], ['data-srcset', 'srcset'], ['data-lazy-src', 'src'], ['data-original', 'src']]) {
      const v = img.getAttribute(from);
      if (v && !/^https?:/.test(img.getAttribute(to) || '')) img.setAttribute(to, v);
    }
  }
  for (const s of root.querySelectorAll('source[data-srcset]')) s.srcset = s.dataset.srcset;
  const bottom = root.getBoundingClientRect().bottom + scrollY;
  for (let y = 0; y < bottom; y += 700) { window.scrollTo(0, y); await new Promise((r) => setTimeout(r, 70)); }
  window.scrollTo(0, 0);
  await Promise.all([...root.querySelectorAll('img')].map((i) => (i.complete && i.naturalWidth)
    ? null
    : Promise.race([i.decode().catch(() => {}), new Promise((r) => setTimeout(r, 8000))])));
  await document.fonts.ready;
}

// Find the masthead logo (the key trust signal) on whatever page is loaded — the
// article page first, the home page as a fallback. Prefers the homepage-linking logo
// near the top; supports inline <svg> wordmarks as well as <img> logos; excludes
// article thumbnails/icons. Returns {logoSrc, logoSvg, plate, color}; never a favicon.
// The logo is returned exactly as rendered: computed fills are inlined into the SVG copy,
// sprite <use> references are resolved, and `plate` is the solid header colour it sat on
// (null when that is white), so the stamp can reproduce it without recolouring the mark.
function detectMastheadLogo() {
  let logoSrc = '', logoSvg = '', plate = null, color = '', animated = false;
  const badImg = /sprite|emoji|avatar|gravatar|icon-|\/thumbs?\/|uploads\/sites/i;
  const originRe = new RegExp('^' + location.origin.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + '\\/?$');
  // the masthead sits above the headline; banners above it can push it well below 300 px
  const h1 = document.querySelector('[data-pc-h1]') || document.querySelector('h1');
  const limit = Math.min(900, Math.max(300, h1 ? h1.getBoundingClientRect().top : 600));
  const markWidth = (a) => { const m = a.querySelector('svg, img'); return m ? m.getBoundingClientRect().width : 0; };
  const brandCands = [
    ...[...document.querySelectorAll('a')].filter((a) => {
      const href = a.getAttribute('href') || ''; const r = a.getBoundingClientRect();
      return (href === '/' || originRe.test(href)) && r.top < limit && r.width > 60 && r.left > -1 && r.right < innerWidth + 1;
    }).sort((a, b) => markWidth(b) - markWidth(a)),
    ...document.querySelectorAll('[class*="site-logo" i], [class*="masthead" i], [class*="navbar-brand" i], [class*="logo" i]'),
  ];
  const plateOf = (el) => {
    for (let p = el; p && p !== document.documentElement; p = p.parentElement) {
      const m = getComputedStyle(p).backgroundColor.match(/rgba?\(([^)]+)\)/);
      if (!m) continue;
      const [r, g, b, a = 1] = m[1].split(',').map((v) => parseFloat(v));
      if (a < 0.5) continue;
      const lum = (0.2126 * r + 0.7152 * g + 0.0722 * b) / 255;
      return lum > 0.94 ? null : `rgb(${Math.round(r)}, ${Math.round(g)}, ${Math.round(b)})`;
    }
    return null;
  };
  const SHAPES = 'path, rect, circle, ellipse, polygon, polyline, line, text, tspan, use, g';
  const inlineSvg = (svg) => {
    const copy = svg.cloneNode(true);
    const src = [svg, ...svg.querySelectorAll(SHAPES)], dst = [copy, ...copy.querySelectorAll(SHAPES)];
    src.forEach((el, i) => {
      const cs = getComputedStyle(el), d = dst[i];
      if (!d) return;
      for (const prop of ['fill', 'stroke', 'fill-opacity', 'stroke-width', 'color']) {
        const v = cs.getPropertyValue(prop);
        if (v) d.style.setProperty(prop, v);
      }
    });
    // inline the symbols that sprite <use> elements point at, so the copy survives isolation
    const defs = document.createElementNS('http://www.w3.org/2000/svg', 'defs');
    copy.querySelectorAll('use').forEach((u) => {
      const ref = (u.getAttribute('href') || u.getAttribute('xlink:href') || '').split('#')[1];
      const target = ref && document.getElementById(ref);
      if (target && !copy.querySelector('#' + CSS.escape(ref))) defs.appendChild(target.cloneNode(true));
    });
    if (defs.childNodes.length) copy.prepend(defs);
    const r = svg.getBoundingClientRect();
    if (!copy.getAttribute('viewBox') && r.width && r.height) copy.setAttribute('viewBox', `0 0 ${r.width} ${r.height}`);
    return copy.outerHTML;
  };
  for (const c of brandCands) {
    const img = c.matches('img') ? c : c.querySelector('img');
    if (img) { const s = img.currentSrc || img.src || ''; if (s && !badImg.test(s)) { logoSrc = s; plate = plateOf(img); break; } }
    const svg = c.matches('svg') ? c : c.querySelector('svg');
    if (svg && svg.getBoundingClientRect().width >= 60) {
      logoSvg = inlineSvg(svg);
      plate = plateOf(svg);
      color = getComputedStyle(svg).color;
      animated = svg.getAnimations({ subtree: true }).length > 0 || !!svg.querySelector('animate, animateTransform, animateMotion, set');
      break;
    }
  }
  return { logoSrc, logoSvg, plate, color, animated };
}

// Last-resort logo: og:logo or the site icon. Lower fidelity than a masthead, but
// still the outlet's own mark — better than a text wordmark. Returns an absolute URL.
function detectLogoFallback() {
  const meta = (sel, attr = 'content') => { const e = document.querySelector(sel); if (!e) return ''; const v = (e.getAttribute(attr) || '').trim(); return v ? new URL(v, location.href).href : ''; };
  return meta('meta[property="og:logo"]') || meta('link[rel*="icon"]', 'href');
}

function surgery({ section, keep, drop, logoSrc, logoSvg, plate, color, outletName }) {
  const articleRoot = document.querySelector('[data-pc-root]') || document.body;

  const keepSel = (keep || '').split(',').map((s) => s.trim()).filter(Boolean);
  const protectedEls = new Set();
  keepSel.forEach((sel) => document.querySelectorAll(sel).forEach((n) => protectedEls.add(n)));
  const isProtected = (n) => { for (let p = n; p; p = p.parentElement) if (protectedEls.has(p)) return true; return false; };

  // Isolate the story WITHOUT guessing class names: hide everything that is not on the path from
  // <body> down to the article. The site header, nav, sidebars, footer, and recirculation rails
  // are all siblings of the article's ancestor chain, so they fall away — while legitimate layout
  // wrappers around the body (even ones classed "...--has-sidebar") are on the path and survive.
  // This avoids the brittle "[class*=sidebar]" substring match that deletes real content on some
  // templates. (Broad class-keyword removal does not generalize; isolation by structure does.)
  for (let node = articleRoot; node && node.parentElement && node !== document.body; node = node.parentElement) {
    for (const sib of [...node.parentElement.children]) {
      if (sib === node || sib.contains(articleRoot) || isProtected(sib)) continue;
      if (sib.tagName === 'STYLE' || sib.tagName === 'SCRIPT' || sib.tagName === 'LINK') continue;
      sib.remove();
    }
  }

  // Inside the article, remove only high-confidence junk (never broad layout words): ad slots and
  // embeds, consent/overlay dialogs, in-article newsletter sign-ups and tables of contents.
  const HARD_JUNK = [
    'iframe', 'ins.adsbygoogle', '.adsbygoogle', 'amp-ad', '[id^="div-gpt"]', '[id*="google_ads"]',
    '[data-ad]', '[aria-label*="advertisement" i]', '[role="dialog"]', '[aria-modal="true"]',
    '[class*="newsletter" i]', '[class*="ez-toc" i]', '[class*="table-of-contents" i]',
  ];
  [...HARD_JUNK, ...(drop || '').split(',').map((s) => s.trim()).filter(Boolean)].forEach((sel) => {
    try { document.querySelectorAll(sel).forEach((n) => { if (!n.contains(articleRoot) && !isProtected(n)) n.remove(); }); } catch {}
  });

  // Fixed and sticky chrome outside the article (video players, "keep scrolling" bars, sticky
  // headers, consent overlays) would float over captures far down the page: hide it, unstick
  // the rest, and undo the scroll lock modals put on <html>/<body>. Pure structure, no names.
  const vw = innerWidth, vh = innerHeight;
  for (const el of document.querySelectorAll('body *')) {
    if (el.closest('.pc-header')) continue;
    const cs = getComputedStyle(el);
    if (cs.position === 'fixed') {
      if (el.contains(articleRoot)) el.style.setProperty('position', 'static', 'important');
      else if (!isProtected(el)) el.style.setProperty('display', 'none', 'important');
    } else if (cs.position === 'sticky') {
      el.style.setProperty('position', 'static', 'important');
    } else if (cs.position === 'absolute' && !articleRoot.contains(el) && !el.contains(articleRoot)) {
      const r = el.getBoundingClientRect(), z = parseInt(cs.zIndex, 10) || 0;
      if (z > 1000 && r.width * r.height > 0.3 * vw * vh) el.style.setProperty('display', 'none', 'important');
    }
  }
  // Release scroll locks (consent dialogs pin <body> with position:fixed; overflow:hidden, which
  // also cuts the PDF to one page). Inline !important beats any stylesheet rule, print included.
  for (const el of [document.documentElement, document.body]) {
    for (const [prop, value] of [['overflow', 'visible'], ['position', 'static'], ['height', 'auto'], ['max-height', 'none'], ['top', 'auto']]) el.style.setProperty(prop, value, 'important');
  }

  // widen the main column in case a removed sidebar left the article in a narrow grid track
  document.querySelectorAll('[class*="span8"], [class*="main-content"], [class*="content-area"]')
    .forEach((n) => { n.style.width = '100%'; n.style.maxWidth = '100%'; n.style.flex = '0 0 100%'; });

  // --- isolate one roundup section, if asked ---
  // Works in document order, so it also finds the section when headings and entries live in
  // different wrappers: keep the lead (everything before the first section heading), keep
  // the client's section up to the next heading of the same or a higher level, drop the rest.
  let sectionFound = null;
  if (section) {
    const content = articleRoot;
    const h1 = content.querySelector('[data-pc-h1]') || content.querySelector('h1');
    const heads = [...content.querySelectorAll('h1,h2,h3,h4')].filter((h) => h !== h1 && (!h1 || (h1.compareDocumentPosition(h) & Node.DOCUMENT_POSITION_FOLLOWING)));
    const norm = (s) => (s || '').trim().toLowerCase();
    const target = norm(section).slice(0, 8);
    const start = heads.find((h) => norm(h.textContent).startsWith(target));
    sectionFound = !!start;
    if (start) {
      start.setAttribute('data-pc-section', '');
      const rank = (h) => Number(h.tagName[1]);
      const after = (h) => start.compareDocumentPosition(h) & Node.DOCUMENT_POSITION_FOLLOWING;
      const boundaries = heads.filter((h) => rank(h) <= rank(start));
      const end = boundaries.find((h) => after(h) && !start.contains(h));
      if (end) { const r = document.createRange(); r.setStartBefore(end); r.setEnd(content, content.childNodes.length); r.deleteContents(); }
      const first = boundaries[0];
      if (first && first !== start) { const r = document.createRange(); r.setStartBefore(first); r.setEndBefore(start); r.deleteContents(); }
    }
  }

  // Sweep visual dead space left by request-level blocking: a blocked ad slot or emptied embed
  // (e.g. a removed tweet <blockquote>) becomes an empty box that still takes height. Remove
  // elements INSIDE the article that render with size but carry no text, no media, and no caption.
  // This is generic — it keys on "renders empty", not on any publisher's class names. A genuinely
  // failed first-party image still has its <img>/<figure> in the DOM, so it counts as media and is
  // kept; we also log everything stripped so a real photo can never vanish silently.
  const stripped = [];
  const MEDIA = 'img, picture, video, svg, iframe, embed, object, figcaption, [class*="caption" i]';
  [...articleRoot.querySelectorAll('div, aside, section, blockquote, ins, figure')].forEach((el) => {
    if (!el.isConnected || isProtected(el)) return;        // already gone with an ancestor, or kept
    const r = el.getBoundingClientRect();
    if (r.height < 8 || r.width < 8) return;               // not occupying visible space
    if ((el.innerText || '').trim().length || el.querySelector(MEDIA)) return;  // has real content
    const cls = (typeof el.className === 'string' && el.className.trim()) ? '.' + el.className.trim().split(/\s+/).join('.') : '';
    stripped.push(el.tagName.toLowerCase() + cls + ' [' + Math.round(r.width) + '×' + Math.round(r.height) + ']');
    el.remove();
  });

  // Sweep link farms left inside the article: "most viewed" rails, related-story lists, tag
  // clouds. Structure only: a block with several links whose text is almost all link text and
  // that holds no real prose. The widest such block is removed (and logged), never the headline.
  const linkLen = (el) => [...el.querySelectorAll('a')].reduce((n, a) => n + (a.innerText || '').trim().length, 0);
  const prose = (el) => [...el.querySelectorAll('p')].some((p) => ((p.innerText || '').trim().length - linkLen(p)) > 120);
  const isFarm = (el) => {
    if (!el.isConnected || isProtected(el) || el.querySelector('h1') || el.contains(articleRoot)) return false;
    const text = (el.innerText || '').trim().length;
    return text >= 60 && el.querySelectorAll('a').length >= 3 && linkLen(el) >= 0.8 * text && !prose(el);
  };
  [...articleRoot.querySelectorAll('div, aside, section, ul, ol, nav')].forEach((el) => {
    if (!isFarm(el) || (el.parentElement && el.parentElement !== articleRoot && isFarm(el.parentElement))) return;
    const r = el.getBoundingClientRect();
    const cls = (typeof el.className === 'string' && el.className.trim()) ? '.' + el.className.trim().split(/\s+/).slice(0, 2).join('.') : '';
    stripped.push('link list ' + el.tagName.toLowerCase() + cls + ' [' + Math.round(r.width) + '×' + Math.round(r.height) + ']');
    el.remove();
  });

  // --- clip header: the outlet LOGO only (the key trust signal), shown large above the article.
  // The mark is never recoloured: it keeps the colours it had on the site and, when the site
  // showed it on a coloured header, it sits on a plate of that same colour. ---
  const esc = (s) => (s || '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/"/g, '&quot;');
  const LOGO_H = 64;
  const sizer = document.createElement('style');
  sizer.textContent = '.pc-logo svg{height:' + LOGO_H + 'px !important;width:auto !important;display:block}';
  document.head && document.head.appendChild(sizer);
  const plateCss = plate ? 'background:' + plate + ';padding:14px 24px;' : '';
  const logoImg = logoSvg
    ? '<span class="pc-logo" style="display:inline-block;height:' + LOGO_H + 'px;line-height:0;color:' + (color || '#111') + '">' + logoSvg + '</span>'
    : logoSrc
    ? '<img class="pc-logo" src="' + esc(logoSrc) + '" alt="' + esc(outletName) + '" style="height:' + LOGO_H + 'px;max-width:420px;width:auto;object-fit:contain;display:block">'
    : '<span class="pc-logo pc-wordmark" style="font:700 30px/1 Georgia,serif;color:#111">' + esc(outletName) + '</span>';
  const header = document.createElement('div');
  header.className = 'pc-header';
  header.style.cssText = 'background:#fff;padding:4px 4px 16px;margin:0 0 14px;display:flex;justify-content:center;align-items:center';
  header.innerHTML = '<span class="pc-plate" style="display:inline-flex;align-items:center;line-height:0;' + plateCss + '">' + logoImg + '</span>';
  document.body.prepend(header);
  return { stripped, sectionFound };
}

// Centre the stamped logo over the article's own content column rather than the window, so
// it sits above the story once side rails are gone.
function alignStamp() {
  const root = document.querySelector('[data-pc-root]') || document.body;
  const header = document.querySelector('.pc-header'), plate = document.querySelector('.pc-plate');
  if (!header || !plate) return;
  let x0 = Infinity, x1 = -Infinity;
  for (const el of root.querySelectorAll('h1, h2, h3, p, figure, img, picture, video, blockquote')) {
    const r = el.getBoundingClientRect();
    if (r.width < 2 || r.height < 2 || r.left < -1 || r.right > innerWidth + 1) continue;   // skip off-canvas slides
    x0 = Math.min(x0, r.left); x1 = Math.max(x1, r.right);
  }
  if (!Number.isFinite(x0)) return;
  const w = plate.getBoundingClientRect().width;
  header.style.justifyContent = 'flex-start';
  header.style.paddingLeft = Math.max(4, Math.round((x0 + x1) / 2 - w / 2)) + 'px';
}

// Freeze the page: finish finite animations, pause infinite ones at their current frame
// (cancel() would drop them to a base style), pause SMIL clocks and videos. The stamped
// logo is left alone here; its frame is chosen separately.
function freezePage() {
  for (const a of document.getAnimations()) {
    const target = a.effect && a.effect.target;
    if (target && target.closest && target.closest('.pc-header')) continue;
    try { Number.isFinite(a.effect?.getComputedTiming().endTime) ? a.finish() : a.pause(); } catch {}
  }
  document.querySelectorAll('svg').forEach((s) => { if (!s.closest('.pc-header')) { try { s.pauseAnimations(); } catch {} } });
  document.querySelectorAll('video, audio').forEach((v) => { try { v.pause(); } catch {} });
}

// Copying an animated SVG restarts its animations, so a stamped copy is caught mid-draw.
// Seek every animation of the stamped logo to its first or last frame and hold it there.
function setLogoFrame(which) {
  const svg = document.querySelector('.pc-logo svg');
  if (!svg) return false;
  for (const a of svg.getAnimations({ subtree: true })) {
    try {
      a.pause();
      const end = a.effect.getComputedTiming().endTime;
      a.currentTime = which === 'start' ? 0 : (Number.isFinite(end) ? end : 0);
    } catch {}
  }
  try { svg.pauseAnimations(); svg.setCurrentTime(which === 'start' ? 0 : 1e6); } catch {}
  return true;
}

// Width of the non-transparent pixels of a PNG (measured in the page, no image library).
async function inkWidth(b64) {
  const img = new Image(); img.src = 'data:image/png;base64,' + b64; await img.decode();
  const c = new OffscreenCanvas(img.naturalWidth, img.naturalHeight); const g = c.getContext('2d');
  g.drawImage(img, 0, 0);
  const { data, width, height } = g.getImageData(0, 0, c.width, c.height);
  let x0 = width, x1 = -1;
  for (let y = 0; y < height; y += 2) for (let x = 0; x < width; x++) if (data[(y * width + x) * 4 + 3] > 24) { if (x < x0) x0 = x; if (x > x1) x1 = x; }
  return x1 < 0 ? 0 : x1 - x0 + 1;
}

// Locate every mention of the client's terms inside the pinned article. Walks text nodes,
// matches case-sensitively across inline elements (never across blocks), expands each hit to
// its sentence and key clause, and returns page-space rectangles plus the features used to
// score the hit. Ranges are kept on window so rectangles can be re-measured after capture.
function locateMentions({ terms, ambiguous }) {
  const root = document.querySelector('[data-pc-root]') || document.body;
  const BLOCK = 'p,li,blockquote,h1,h2,h3,h4,h5,h6,figcaption,td,th,dd,dt,pre,summary,caption';
  const SKIP = 'script,style,noscript,template,svg,[hidden],[aria-hidden="true"],.pc-header';
  const norm = (ch) => ch.replace(/[\u2018\u2019\u02bc\u2032]/g, "'").replace(/[\u00ad\u200b-\u200d\u2060\ufeff]/g, '').replace(/[\u00a0\u2007\u202f]/g, ' ').replace(/[\u2010\u2011]/g, '-');
  const blocks = []; let cur = null;
  const tw = document.createTreeWalker(root, NodeFilter.SHOW_TEXT, { acceptNode(n) {
    const p = n.parentElement;
    if (!p || p.closest(SKIP)) return NodeFilter.FILTER_REJECT;
    const cs = getComputedStyle(p);
    if (cs.display === 'none' || cs.visibility === 'hidden') return NodeFilter.FILTER_REJECT;
    return NodeFilter.FILTER_ACCEPT;
  } });
  for (let n; (n = tw.nextNode());) {
    const el = n.parentElement.closest(BLOCK) || n.parentElement;
    if (!cur || cur.el !== el) { cur = { el, text: '', map: [] }; blocks.push(cur); }
    const v = n.nodeValue;
    for (let i = 0; i < v.length; i++) for (const ch of norm(v[i])) { cur.text += ch; cur.map.push([n, i]); }
  }
  const esc = (s) => s.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
  const alts = terms.map((t) => t.split(/\s+/).map((w) => esc(norm(w))).join('\\s+'));
  const re = new RegExp('(?<![\\p{L}\\p{N}])(?:' + alts.join('|') + ")(?:'s)?(?![\\p{L}\\p{N}])", 'gu');
  const segmenter = new Intl.Segmenter(document.documentElement.lang || 'en', { granularity: 'sentence' });
  const ABBR = /(?:\b[A-Z]|\b(?:Mr|Mrs|Ms|Dr|Prof|St|vs|etc|Inc|Co|Corp|Ltd|No|Jan|Feb|Mar|Apr|Jun|Jul|Aug|Sep|Sept|Oct|Nov|Dec|U\.S|e\.g|i\.e))\.\s*$/;
  const paragraphs = [...root.querySelectorAll('p')];
  const sectionHead = root.querySelector('[data-pc-section]');
  const pageRects = (range) => [...range.getClientRects()].filter((q) => q.width > 1 && q.height > 1)
    .map((q) => ({ x: q.left + scrollX, y: q.top + scrollY, w: q.width, h: q.height }));
  const rangeOf = (blk, a, b) => {
    const s = blk.map[a], e = blk.map[b - 1];
    const r = document.createRange(); r.setStart(s[0], s[1]); r.setEnd(e[0], e[1] + 1); return r;
  };
  const trimSpan = (text, a, b) => { while (a < b && /\s/.test(text[a])) a++; while (b > a && /\s/.test(text[b - 1])) b--; return [a, b]; };
  const host = location.hostname.replace(/^www\./, '');
  const hits = [];
  window.__pcHits = [];
  for (const blk of blocks) {
    if (!root.contains(blk.el) && blk.el !== root) continue;
    // sentence spans, re-joining splits after abbreviations ("U.S.", "Inc.")
    const sents = [];
    for (const s of segmenter.segment(blk.text)) {
      const prev = sents[sents.length - 1];
      if (prev && ABBR.test(blk.text.slice(prev[0], prev[1]))) prev[1] = s.index + s.segment.length;
      else sents.push([s.index, s.index + s.segment.length]);
    }
    for (const m of blk.text.matchAll(re)) {
      const start = m.index, end = m.index + m[0].length;
      const [sa0, sb0] = sents.find(([a, b]) => start >= a && start < b) || [0, blk.text.length];
      const [sa, sb] = trimSpan(blk.text, sa0, sb0);
      const sentence = blk.text.slice(sa, sb).replace(/\s+/g, ' ').trim();
      // key clause: for a long sentence, the verbatim span around the term that reads on its own
      // (30-80 characters, the shorter the better), bounded by punctuation or before a joining
      // word. The reel highlights this span and its phone formats zoom to fit it, so a shorter
      // clause reads larger and long sentences never need long holds.
      let ca = sa, cb = sb;
      if (sb - sa > 60) {
        const cuts = [{ at: sa, hard: true }];
        let depth = 0, quoted = false;
        for (let i = sa; i < sb; i++) {
          const ch = blk.text[i];
          if (ch === '(' || ch === '[') depth++;
          else if ((ch === ')' || ch === ']') && depth) depth--;
          else if (ch === '"') quoted = !quoted;
          else if (ch === '\u201c') quoted = true;
          else if (ch === '\u201d') quoted = false;
          if (depth || quoted) continue;                // never cut inside brackets or quotes
          if (/[,;:\u2013\u2014]/.test(ch) && /\s/.test(blk.text[i + 1] || '')) cuts.push({ at: i + 1, hard: true });
          else if (/\s/.test(ch) && /^(?:and|but|or|so|yet|which|while|whereas|because|though|although|from|with|for|to|that|than|as|in|on|at|by|after|before|since|without)\s/i.test(blk.text.slice(i + 1, i + 12))) cuts.push({ at: i + 1, hard: false });
        }
        cuts.push({ at: sb, hard: true });
        const trim = (a, b) => {
          while (a < b && /[\s,;:]/.test(blk.text[a])) a++;
          for (let guard = 0; guard < 4; guard++) {
            while (b > a && /[,;:\s]$/.test(blk.text.slice(a, b))) b--;
            const tail = /\s(?:and|but|or|so|yet|which|while)$/i.exec(blk.text.slice(a, b));
            if (!tail || b - tail[0].length <= end) break;
            b -= tail[0].length;
          }
          return [a, b];
        };
        let best = null;
        for (const c0 of cuts) {
          if (c0.at > start) break;
          for (const c1 of cuts) {
            if (c1.at < end) continue;
            const [a, b] = trim(c0.at, c1.at);
            const len = b - a;
            if (len > 80 || b < end || a > start) continue;
            const score = (c0.hard ? 0 : -1) + (c1.hard ? 0 : -2) - (len < 30 ? 3 : 0) - Math.abs(len - 40) / 30;
            if (!best || score > best.score) best = { a, b, score };
          }
        }
        if (best) [ca, cb] = [best.a, best.b];
      }
      const clause = blk.text.slice(ca, cb).replace(/\s+/g, ' ').trim();
      const termRange = rangeOf(blk, start, end);
      const termRects = pageRects(termRange);
      if (!termRects.length) continue;                               // visually hidden
      const hostEl = blk.map[start][0].parentElement;
      const link = hostEl.closest('a[href]');
      let sameSiteLink = false;
      if (link) { try { const u = new URL(link.href, location.href); sameSiteLink = u.hostname.replace(/^www\./, '') === host && u.pathname !== location.pathname; } catch {} }
      const blockText = (blk.el.innerText || '').trim();
      const linkText = [...blk.el.querySelectorAll('a')].reduce((n, a) => n + (a.innerText || '').trim().length, 0);
      const before = blk.text.slice(sa, start);
      const words = blockText.split(/\s+/).filter((w) => /\p{L}/u.test(w));
      const heading = /^H[1-6]$/.test(blk.el.tagName);
      const br = blk.el.getBoundingClientRect();
      const termText = m[0].replace(/'s$/, '').replace(/\s+/g, ' ');
      const sentenceRange = rangeOf(blk, sa, sb), clauseRange = rangeOf(blk, ca, cb);
      window.__pcHits.push({ termRange, sentenceRange, clauseRange });
      hits.push({
        index: hits.length,
        term: terms.find((t) => t.replace(/\s+/g, ' ') === termText) || termText,
        text: m[0].replace(/\s+/g, ' '),
        sentence, clause,
        block: blk.el.tagName,
        ambiguous: ambiguous.includes(termText),
        sentenceInitial: !/[\p{L}\p{N}]/u.test(before),
        titleCaseHeading: heading && words.length > 1 && words.filter((w) => /^\p{Lu}/u.test(w)).length / words.length >= 0.6,
        inLink: !!link, sameSiteLink,
        linkDense: blockText.length > 0 && linkText / blockText.length > 0.6,
        inAside: !!hostEl.closest('aside, nav, footer, [role="complementary"], [role="navigation"]'),
        inCaption: !!hostEl.closest('figcaption, figure'),
        inHeadline: !!hostEl.closest('h1'),
        inHeading: heading,
        attributedQuote: /["\u201c\u201d]/.test(sentence) && /\b(said|says|according to|told)\b/i.test(sentence),
        paragraphIndex: paragraphs.indexOf(blk.el),
        sectionScoped: !!sectionHead,
        inSection: !!sectionHead && (sectionHead.contains(blk.el) || !!(sectionHead.compareDocumentPosition(blk.el) & Node.DOCUMENT_POSITION_FOLLOWING)),
        rects: {
          term: termRects,
          sentence: pageRects(sentenceRange),
          clause: pageRects(clauseRange),
          paragraph: { x: br.left + scrollX, y: br.top + scrollY, w: br.width, h: br.height },
        },
        context: (() => {
          const prev = blk.el.previousElementSibling, next = blk.el.nextElementSibling;
          const pr = prev ? prev.getBoundingClientRect() : br, nr = next ? next.getBoundingClientRect() : br;
          return { top: Math.min(pr.top, br.top) + scrollY, bottom: Math.max(nr.bottom, br.bottom) + scrollY };
        })(),
      });
    }
  }
  return hits;
}

// Layout facts measured as the last step before capture.
function measureLayout() {
  const pr = (el) => { if (!el) return null; const r = el.getBoundingClientRect(); return { x: r.left + scrollX, y: r.top + scrollY, w: r.width, h: r.height }; };
  const root = document.querySelector('[data-pc-root]') || document.body;
  const h1 = document.querySelector('[data-pc-h1]') || root.querySelector('h1') || document.querySelector('h1');
  const rr = pr(root), hr = pr(h1);
  let hero = null;
  if (hr) {
    for (const el of root.querySelectorAll('img, picture, video, figure')) {
      const r = pr(el);
      if (r.y >= hr.y - 40 && r.y < hr.y + hr.h + 1400 && r.w >= 0.4 * rr.w && r.h > 120) { hero = r; break; }
    }
  }
  // horizontal extent of the article's real content (the root is often a full-width grid
  // with empty gutters once side rails are gone)
  let cx0 = Infinity, cx1 = -Infinity;
  for (const el of root.querySelectorAll('h1, h2, h3, p, figure, img, picture, video, blockquote, ul, ol, table, time, address')) {
    const r = el.getBoundingClientRect();
    if (r.width < 2 || r.height < 2 || r.left < -1 || r.right > innerWidth + 1) continue;   // skip off-canvas slides
    cx0 = Math.min(cx0, r.left + scrollX); cx1 = Math.max(cx1, r.right + scrollX);
  }
  return {
    root: rr, h1: hr, hero,
    content: Number.isFinite(cx0) ? { x0: cx0, x1: cx1 } : null,
    header: pr(document.querySelector('.pc-header')),
    plate: pr(document.querySelector('.pc-plate')),
    logo: pr(document.querySelector('.pc-logo')),
    docHeight: Math.max(document.documentElement.scrollHeight, document.body.scrollHeight),
    h1Count: document.querySelectorAll('h1').length,
    path: location.pathname,
  };
}

// Wait until the article stops moving (late images, embeds, web fonts) before measuring.
async function waitForStableLayout() {
  const root = document.querySelector('[data-pc-root]') || document.body;
  const sample = () => { const r = root.getBoundingClientRect(); return `${Math.round(r.top + scrollY)}|${Math.round(r.height)}|${document.documentElement.scrollHeight}`; };
  let last = sample();
  for (let i = 0; i < 12; i++) {
    await new Promise((r) => setTimeout(r, 400));
    const now = sample();
    if (now === last) return true;
    last = now;
  }
  return false;
}

function remeasure(index) {
  const h = (window.__pcHits || [])[index];
  const root = document.querySelector('[data-pc-root]') || document.body;
  const r = root.getBoundingClientRect();
  const t = h ? [...h.termRange.getClientRects()][0] : null;
  return { root: { x: r.left + scrollX, y: r.top + scrollY, w: r.width, h: r.height }, term: t ? { x: t.left + scrollX, y: t.top + scrollY } : null };
}

// ---------------------------------------------------------------------------
// Main
// ---------------------------------------------------------------------------

async function main() {
  const args = parseArgs(process.argv.slice(2));
  const { url, out, section, preview, keep, drop, root, logo: logoArg } = args;
  if (!url || !out) { console.error('need --url and --out'); process.exit(1); }
  const terms = parseMentionTerms(args.mention);
  const assetsDir = args.assets ? resolve(args.assets) : '';
  const scale = args.scale === undefined ? 2 : Number(args.scale);
  if (![2, 3].includes(scale)) { console.error('--scale must be 2 or 3'); process.exit(1); }
  const pick = args.pick === undefined ? 1 : Number(args.pick);
  if (!Number.isInteger(pick) || pick < 1) { console.error('--pick must be a whole number from 1'); process.exit(1); }

  const CHROME = args.chrome
    || process.env.PRESS_CLIP_CHROME
    || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';
  const chromium = loadChromium();
  const warnings = [];
  const warn = (code, detail) => { warnings.push({ code, detail }); console.warn(`WARNING (${code}): ${detail}`); };

  const browser = await chromium.launch({ executablePath: CHROME, headless: true, args: ['--hide-scrollbars', '--force-color-profile=srgb'] });
  try {
  const contextOptions = { viewport: { width: 1180, height: 1600 }, deviceScaleFactor: assetsDir ? scale : 2, reducedMotion: 'reduce', serviceWorkers: 'block', colorScheme: 'light' };
  const context = await browser.newContext(contextOptions);
  await context.addInitScript(pinPathInit);
  const page = await context.newPage();

  // Request rules: ad/recirc networks never load; once the article has settled, same-site
  // scripts may not fetch whole HTML documents (that is how next-article loaders append
  // the following story). JSON/API calls still pass.
  const articleSite = (() => { try { return new URL(url).hostname.split('.').slice(-2).join('.'); } catch { return ''; } })();
  let settled = false, blockedFragments = 0;
  await page.route('**/*', async (route) => {
    const req = route.request();
    let host = '';
    try { host = new URL(req.url()).hostname; } catch { return route.continue(); }
    if (BLOCK_HOSTS.test(host)) return route.abort();
    if (settled && ['fetch', 'xhr'].includes(req.resourceType()) && host.split('.').slice(-2).join('.') === articleSite) {
      const resp = await route.fetch().catch(() => null);
      if (!resp) return route.abort();
      if (/text\/html/i.test(resp.headers()['content-type'] || '')) { blockedFragments++; return route.abort('blockedbyclient'); }
      return route.fulfill({ response: resp });
    }
    return route.continue();
  });
  const response = await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 60000 });
  await page.waitForTimeout(3500);
  settled = true;

  // A bot-protection interstitial is not the article. Stop honestly; never try to get around it.
  const probe = await page.evaluate(pageProbe);
  const block = detectBlock({ status: response ? response.status() : 0, ...probe });
  if (block.blocked) {
    console.error(`BLOCKED: ${url} served ${block.reason} to the automated browser instead of the article.`);
    console.error('Do not try to get around it. Ask the user for their own PDF or screenshot of the article and use that instead.');
    await browser.close();
    process.exit(EXIT_BLOCKED);
  }

  const meta = await page.evaluate(readPageMeta);
  const pinned = await page.evaluate(pinRoot, { root, headlines: meta.headline_candidates });
  await page.evaluate(loadArticleImages);
  await page.waitForTimeout(800);

  // --- resolve the outlet logo BEFORE surgery: every clip must carry one (it is the
  // key trust signal). Priority: explicit --logo > article masthead > HOME-PAGE masthead
  // > og:logo/favicon > text wordmark. The home-page step exists because some article
  // templates render no masthead logo at all; we navigate to the home page to grab one. ---
  let logo = { logoSrc: '', logoSvg: '', plate: null, color: '', animated: false }, logoFrom = '';
  if (logoArg) { logo.logoSrc = logoArg; logoFrom = 'explicit (--logo)'; }
  else { logo = await page.evaluate(detectMastheadLogo); if (logo.logoSrc || logo.logoSvg) logoFrom = 'article page'; }
  if (!logo.logoSrc && !logo.logoSvg) {
    try {
      const origin = new URL(url).origin;
      const home = await context.newPage();
      await home.route('**/*', (route) => {
        try { return BLOCK_HOSTS.test(new URL(route.request().url()).hostname) ? route.abort() : route.continue(); }
        catch { return route.continue(); }
      });
      await home.goto(origin, { waitUntil: 'domcontentloaded', timeout: 60000 });
      await home.waitForTimeout(2000);
      logo = await home.evaluate(detectMastheadLogo);
      if (logo.logoSrc || logo.logoSvg) logoFrom = 'home page';
      await home.close();
    } catch { /* home-page fetch is best-effort */ }
  }
  if (!logo.logoSrc && !logo.logoSvg) { logo.logoSrc = await page.evaluate(detectLogoFallback); if (logo.logoSrc) logoFrom = 'og:logo/favicon fallback'; }
  if (!logo.logoSrc && !logo.logoSvg) logoFrom = 'TEXT WORDMARK — no logo found';

  const outletName = meta.outlet.value || probe.title.replace(/.*[-|\u2013\u2014]\s*/, '').trim() || new URL(url).hostname.replace(/^www\./, '');
  const { stripped, sectionFound } = await page.evaluate(surgery, { section, keep, drop, outletName, ...logo });
  const pinnedHeadline = await page.evaluate(() => {
    const h = document.querySelector('[data-pc-h1]') || document.querySelector('[data-pc-root] h1');
    return h ? h.textContent.replace(/\s+/g, ' ').trim() : null;
  });
  if (pinnedHeadline) meta.visible_headline = pinnedHeadline;
  if (section && !sectionFound) warn('section_not_found', `No heading starting with "${section}" was found; the whole article was kept.`);

  // Force a white page background so Chromium doesn't paint the below-content area of the final
  // A4 page with a site's off-white/grey body color (a common trailing-band artifact). Injected
  // last so it wins the cascade even against the site's own !important html/body background.
  await page.addStyleTag({ content: 'html,body{background:#fff !important}@media print{html,body{background:#fff !important}}' });
  await page.evaluate(loadArticleImages);
  await page.evaluate(alignStamp);
  await page.evaluate(freezePage);

  // An animated logo has no single right frame. Try its first and last frames and keep the one
  // with more ink (the full wordmark rather than a collapsed monogram).
  let logoFrame = null;
  if (logo.logoSvg && await page.evaluate(() => !!document.querySelector('.pc-logo svg') && (document.querySelector('.pc-logo svg').getAnimations({ subtree: true }).length > 0 || !!document.querySelector('.pc-logo svg animate, .pc-logo svg animateTransform, .pc-logo svg set')))) {
    const ink = {};
    for (const which of ['start', 'end']) {
      await page.evaluate(setLogoFrame, which);
      const shot = await page.locator('.pc-logo').first().screenshot({ omitBackground: true, animations: 'allow', style: 'html,body,.pc-header,.pc-plate{background:transparent !important}' });
      ink[which] = await page.evaluate(inkWidth, shot.toString('base64'));
    }
    logoFrame = ink.start >= ink.end ? 'start' : 'end';
    await page.evaluate(setLogoFrame, logoFrame);
    console.log(`animated logo frozen on its ${logoFrame} frame (ink width start=${ink.start}px, end=${ink.end}px)`);
  }
  await page.waitForTimeout(600);

  if (pinned.h1Count > 1) warn('multiple_h1', `The page had ${pinned.h1Count} <h1> elements; the article was pinned to "${pinned.h1Text || 'the first one'}".`);
  if (blockedFragments) console.log(`blocked ${blockedFragments} next-article HTML fetch(es) after load`);
  if (meta.paywalled) warn('paywall_suspected', 'The page marks itself as not free to read (isAccessibleForFree: false). Capture only what is visible; never work around a paywall.');
  if (logoFrom.startsWith('TEXT')) warn('logo_text_fallback', 'No outlet logo was found on the article or home page; the clip shows a text wordmark. Re-run with --logo "<url>".');
  if (meta.visible_headline && meta.headline.source !== null && meta.headline.source !== 'visible') {
    const a = meta.visible_headline.toLowerCase(), b = (meta.headline.value || '').toLowerCase();
    if (a && b && a !== b && !a.includes(b) && !b.includes(a)) warn('headline_mismatch', `The page's metadata headline ("${meta.headline.value}") differs from the headline shown on the page ("${meta.visible_headline}"). The visible one is used.`);
  }
  // The headline shown in the capture is the one to display; metadata is the cross-check.
  if (meta.visible_headline) meta.headline = { value: meta.visible_headline, source: 'visible' };

  // ---- animation assets --------------------------------------------------------------
  let sidecar = null;
  if (assetsDir) {
    mkdirSync(assetsDir, { recursive: true });
    const ambiguous = terms.filter(isAmbiguousTerm);
    // measure LAST: wait for the layout to settle, locate mentions and lay out the captures
    // immediately before the shots, then re-measure. One automatic retry if anything moved.
    const captureOnce = async () => {
      await page.evaluate(waitForStableLayout);
      const hits = terms.length ? await page.evaluate(locateMentions, { terms, ambiguous }) : [];
      const ranked = rankMentions(hits);
      const chosen = ranked[pick - 1] || null;
      const layout = await page.evaluate(measureLayout);
      const plateRect = layout.plate || layout.logo;
      const content = layout.content || { x0: layout.root.x, x1: layout.root.x + layout.root.w };
      const colX0 = Math.floor(Math.min(content.x0 - 24, plateRect ? plateRect.x - 16 : Infinity));
      const colX1 = Math.ceil(Math.max(content.x1 + 24, plateRect ? plateRect.x + plateRect.w + 16 : 0));
      const column = { x: Math.max(0, colX0), w: Math.min(1180, colX1) - Math.max(0, colX0) };
      const pageTop = Math.floor(Math.min(layout.header ? layout.header.y : 0, layout.root.y));
      const rootBottom = layout.root.y + layout.root.h;
      const maxCss = Math.floor(MAX_TILE_DEVICE_PX / scale);
      let topBottom = layout.hero ? layout.hero.y + layout.hero.h + 32 : (layout.h1 ? layout.h1.y + layout.h1.h + 700 : pageTop + 1400);
      topBottom = Math.min(topBottom, rootBottom, pageTop + maxCss);
      const assets = [];
      const shoot = async (file, role, rect, extra = {}) => {
        const clip = { x: column.x, y: Math.max(0, Math.floor(rect.y)), width: column.w, height: Math.ceil(rect.h) };
        const buffer = await page.screenshot({ path: join(assetsDir, file), fullPage: true, clip, animations: 'allow', caret: 'hide' });
        const size = pngSize(buffer);
        const asset = { role, file, width: size.width, height: size.height, scale, rect: { x: clip.x, y: clip.y, w: clip.width, h: clip.height }, ...extra };
        assets.push(asset);
        return asset;
      };
      await shoot('top.png', 'top', { y: pageTop, h: topBottom - pageTop });
      const stripBottom = Math.min(rootBottom, Math.max(topBottom, chosen ? Math.max(chosen.rects.paragraph.y + chosen.rects.paragraph.h, chosen.context.bottom) + 900 : topBottom));
      const tiles = planStrips(pageTop, stripBottom, scale);
      for (let i = 0; i < tiles.length; i++) await shoot(`strip-${i + 1}.png`, 'strip', tiles[i], { index: i });
      if (chosen) {
        const y0 = Math.max(pageTop, Math.min(chosen.context.top, chosen.rects.paragraph.y - 40) - 24);
        const y1 = Math.min(rootBottom, Math.max(chosen.context.bottom, chosen.rects.paragraph.y + chosen.rects.paragraph.h + 40) + 24);
        await shoot('mention-1.png', 'mention', { y: y0, h: Math.min(y1 - y0, maxCss) }, { mention: chosen.index });
      }
      // the mark itself on a transparent background (the reel prefers the crop from top.png)
      let mastheadAsset = null;
      if (layout.logo && (logo.logoSrc || logo.logoSvg)) {
        const buffer = await page.locator('.pc-logo').first().screenshot({ path: join(assetsDir, 'masthead.png'), omitBackground: true, animations: 'allow', style: 'html,body,.pc-header,.pc-plate{background:transparent !important}' });
        const size = pngSize(buffer);
        mastheadAsset = { role: 'masthead', file: 'masthead.png', width: size.width, height: size.height, scale, rect: roundRect(layout.logo) };
        assets.push(mastheadAsset);
      }
      // the captures must match the measurements: re-measure and flag any shift
      const after = await page.evaluate(remeasure, chosen ? chosen.index : -1);
      const shifted = Math.abs(after.root.y - layout.root.y) > 1 || Math.abs(after.root.h - layout.root.h) > 1
        || (chosen && after.term && (Math.abs(after.term.y - chosen.rects.term[0].y) > 1 || Math.abs(after.term.x - chosen.rects.term[0].x) > 1));
      const shiftDetail = !shifted ? null : `The page layout moved while the captures were taken (article top ${layout.root.y}→${after.root.y}, height ${Math.round(layout.root.h)}→${Math.round(after.root.h)}${chosen && after.term ? `, mention ${Math.round(chosen.rects.term[0].x)},${Math.round(chosen.rects.term[0].y)}→${Math.round(after.term.x)},${Math.round(after.term.y)}` : ''}); re-run the clip before using the rectangles.`;
      return { ranked, chosen, layout, assets, mastheadAsset, plateRect, shiftDetail };
    };
    let capture = await captureOnce();
    if (capture.shiftDetail) {
      console.log('layout moved during capture; measuring and capturing again');
      for (const a of capture.assets) { try { unlinkSync(join(assetsDir, a.file)); } catch {} }
      capture = await captureOnce();
    }
    const { ranked, chosen, layout, assets, mastheadAsset, plateRect } = capture;
    if (capture.shiftDetail) warn('layout_shifted', capture.shiftDetail);
    if (terms.length && !ranked.length) warn('no_mention_found', `None of ${terms.map((t) => `"${t}"`).join(', ')} appears in the article body. Do not clip it for the client unless you can confirm the mention yourself.`);
    else if (terms.length && !chosen) warn('pick_out_of_range', `--pick ${pick} is beyond the ${ranked.length} ranked mention(s); no mention was captured.`);

    const pageAssets = assets.filter((a) => a.role !== 'masthead');
    const mentions = ranked.slice(0, 20).map((hit, i) => {
      const lines = { term: mergeLineRects(hit.rects.term), sentence: mergeLineRects(hit.rects.sentence), clause: mergeLineRects(hit.rects.clause) };
      return {
        rank: i + 1,
        chosen: hit === chosen,
        term: hit.term,
        text: hit.text,
        sentence: hit.sentence,
        clause: hit.clause,
        block: hit.block.toLowerCase(),
        in_link: hit.inLink,
        score: hit.score,
        reasons: hit.reasons,
        rects: { ...lines, paragraph: roundRect(hit.rects.paragraph) },
        asset_rects: pageAssets
          .map((a) => ({ file: a.file, term: toAssetRects(lines.term, a), sentence: toAssetRects(lines.sentence, a), clause: toAssetRects(lines.clause, a) }))
          .filter((r) => r.term.length || r.sentence.length || r.clause.length),
      };
    });
    sidecar = {
      version: SIDECAR_VERSION,
      tool: TOOL_VERSION,
      source_url: url,
      final_url: page.url(),
      canonical_url: meta.canonical_url,
      captured_at: new Date().toISOString(),
      viewport: contextOptions.viewport,
      scale,
      scope: section && sectionFound ? 'section' : 'whole',
      section: section && sectionFound ? section : null,
      mention_terms: terms,
      page: { outlet: meta.outlet, headline: meta.headline, byline: meta.byline, published_at: meta.published_at },
      logo: {
        resolved_from: logoFrom.startsWith('TEXT') ? 'text fallback' : logoFrom,
        kind: logo.logoSvg ? 'svg' : logo.logoSrc ? 'image' : 'text',
        frame: logoFrame,
        plate_color: logo.plate || null,
        file: mastheadAsset ? mastheadAsset.file : null,
        rect: layout.logo ? roundRect(layout.logo) : null,
        plate_rect: plateRect ? roundRect(plateRect) : null,
      },
      root: roundRect(layout.root),
      headline_rect: layout.h1 ? roundRect(layout.h1) : null,
      assets,
      mentions,
      warnings,
    };
  }

  // ---- preview and PDF ---------------------------------------------------------------
  if (preview) {
    const docHeight = await page.evaluate(() => Math.max(document.documentElement.scrollHeight, document.body.scrollHeight));
    const mode = pickPreviewMode(docHeight, contextOptions.deviceScaleFactor);
    if (mode.warning) warn(mode.warning.code, mode.warning.detail);
    await page.screenshot({ path: preview, fullPage: true, scale: mode.scale, animations: 'allow', caret: 'hide',
      ...(mode.clipHeight ? { clip: { x: 0, y: 0, width: contextOptions.viewport.width, height: mode.clipHeight } } : {}) });
  }
  await page.pdf({ path: out, format: 'A4', printBackground: true, margin: { top: '10mm', bottom: '12mm', left: '8mm', right: '8mm' } });
  if (sidecar) {
    sidecar.warnings = warnings;
    writeFileSync(join(assetsDir, 'clip.json'), JSON.stringify(sidecar, null, 2) + '\n');
  }
  console.log('clip written:', out, '| logo:', logoFrom);
  if (stripped && stripped.length) console.log('swept ' + stripped.length + ' empty placeholder(s):\n  - ' + stripped.join('\n  - '));
  if (sidecar) {
    const m = sidecar.mentions.find((x) => x.chosen);
    console.log(`assets written: ${assetsDir} (${sidecar.assets.length} files + clip.json)`);
    if (m) console.log(`mention (rank ${m.rank}, score ${m.score}): "${m.sentence}"`);
  }
  } finally {
    await browser.close();
  }
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  main().catch((e) => { console.error('ERR', e.message); process.exit(1); });
}
