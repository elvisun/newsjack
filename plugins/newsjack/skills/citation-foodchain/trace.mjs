#!/usr/bin/env node
/**
 * trace.mjs — deterministic evidence extraction for the citation-foodchain skill.
 *
 * Zero dependencies: Node >= 18 standard library only. An optional `--browser`
 * mode drives the user's own Chrome/Edge through `playwright-core` (the same
 * optional package press-clip uses) for pages that only render with JavaScript.
 *
 * This script never decides anything. It pulls the machine-readable facts an
 * agent needs to judge a provenance hop honestly: where a page says it got a
 * claim from, which links sit next to the matching sentence, when the page was
 * published, whether it is syndicated or sponsored, what Wikipedia cites for a
 * sentence, and when the Wayback Machine first saw a URL. The skill's SKILL.md
 * owns every judgment call (claim support, hop direction, root type).
 *
 * Subcommands:
 *   share  <share-url> [--browser] [--html saved.html]   ChatGPT / Perplexity share link → answer + cited URLs
 *   fetch  <url> [--claim "<text>"]... [--browser] [--brief] [--no-wayback] [--out file.json]
 *   unwrap <url> [--follow]            resolve redirect/tracking wrappers to the real URL
 *   wayback <url>                      earliest Wayback Machine capture (CDX API)
 *   reuse  <a> <b>                     verbatim overlap between two pages (URLs or fetch JSON files)
 *   batch  <urls.txt> [--out dir]      fetch many URLs, one JSON per URL
 *   demo-fixture <dir> [--serve]       write + serve the offline synthetic chains (for tests)
 *
 * Every subcommand prints one JSON document to stdout (or --out).
 */

import { createRequire } from 'module';
import { pathToFileURL, fileURLToPath } from 'url';
import { readFileSync, writeFileSync, mkdirSync, existsSync } from 'fs';
import { join, resolve, dirname } from 'path';
import { createServer } from 'http';

const VERSION = '0.2.0';
const UA = process.env.CITATION_FOODCHAIN_UA
  || 'Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0 Safari/537.36';
const FETCH_TIMEOUT_MS = Number(process.env.CITATION_FOODCHAIN_TIMEOUT_MS || 25000);
const MAX_BYTES = 4 * 1024 * 1024;

// ---------------------------------------------------------------------------
// CLI plumbing
// ---------------------------------------------------------------------------

export function parseArgs(argv) {
  const out = { _: [], claim: [] };
  for (let i = 0; i < argv.length; i++) {
    const a = argv[i];
    if (a === '--claim') { out.claim.push(argv[++i]); continue; }
    if (a === '--out') { out.out = argv[++i]; continue; }
    if (a === '--chrome') { out.chrome = argv[++i]; continue; }
    if (a === '--html') { out.html = argv[++i]; continue; }
    if (a === '--max-paragraphs') { out.maxParagraphs = Number(argv[++i]); continue; }
    if (a === '--port') { out.port = Number(argv[++i]); continue; }
    if (a === '--browser') { out.browser = true; continue; }
    if (a === '--brief') { out.brief = true; continue; }
    if (a === '--full') { out.full = true; continue; } // accepted for compatibility; full is the default
    if (a === '--no-wayback') { out.noWayback = true; continue; }
    if (a.startsWith('--')) { out[a.slice(2)] = true; continue; }
    out._.push(a);
  }
  return out;
}

function emit(obj, outPath) {
  const text = JSON.stringify(obj, null, 2);
  if (outPath) {
    mkdirSync(dirname(resolve(outPath)), { recursive: true });
    writeFileSync(outPath, text + '\n');
    process.stderr.write(`wrote ${outPath}\n`);
  } else {
    process.stdout.write(text + '\n');
  }
}

const sleep = (ms) => new Promise((r) => setTimeout(r, ms));

// ---------------------------------------------------------------------------
// Fetching
// ---------------------------------------------------------------------------

async function fetchWithTimeout(url, opts = {}) {
  const ctrl = new AbortController();
  const t = setTimeout(() => ctrl.abort(), opts.timeout || FETCH_TIMEOUT_MS);
  try {
    return await fetch(url, {
      redirect: opts.redirect || 'follow',
      signal: ctrl.signal,
      headers: {
        'user-agent': UA,
        accept: 'text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8',
        'accept-language': 'en-US,en;q=0.9',
        ...(opts.headers || {}),
      },
    });
  } finally {
    clearTimeout(t);
  }
}

async function readBodyBytes(res) {
  const reader = res.body?.getReader?.();
  if (!reader) return Buffer.from(await res.arrayBuffer());
  const chunks = [];
  let total = 0;
  while (true) {
    const { done, value } = await reader.read();
    if (done) break;
    chunks.push(value);
    total += value.length;
    if (total > MAX_BYTES) { try { await reader.cancel(); } catch {} break; }
  }
  return Buffer.concat(chunks);
}

function detectCharset(contentType, bytes) {
  const fromHeader = (contentType.match(/charset=\s*"?([\w-]+)/i) || [])[1];
  if (fromHeader) return fromHeader;
  const head = bytes.subarray(0, 4096).toString('latin1');
  const fromMeta = (head.match(/<meta[^>]+charset=["']?\s*([\w-]+)/i) || [])[1];
  return fromMeta || 'utf-8';
}

function decodeBytes(bytes, charset) {
  try { return new TextDecoder(charset.toLowerCase()).decode(bytes); } catch { return bytes.toString('utf8'); }
}

async function fetchHtml(url) {
  // Rate limits (Wayback especially) and transient 5xx get two retries with backoff.
  let res;
  for (let attempt = 0; ; attempt++) {
    res = await fetchWithTimeout(url);
    if ((res.status === 429 || res.status === 503 || res.status === 502) && attempt < 2) { await sleep(2000 * (attempt + 1)); continue; }
    break;
  }
  const contentType = res.headers.get('content-type') || '';
  let html = '';
  let truncated = false;
  if (/html|xml|text\/plain|json/i.test(contentType) || !contentType) {
    const bytes = await readBodyBytes(res);
    truncated = bytes.length >= MAX_BYTES;
    html = decodeBytes(bytes, detectCharset(contentType, bytes));
  }
  return { status: res.status, finalUrl: res.url || url, contentType, html, truncated };
}

function loadChromium(base) {
  const candidates = [base, process.cwd(), dirname(fileURLToPath(import.meta.url))];
  for (const b of candidates) {
    if (!b) continue;
    try { return createRequire(pathToFileURL(join(b, 'x.js')))('playwright-core').chromium; } catch {}
  }
  throw new Error('playwright-core not found. Run `npm i playwright-core` in this folder (only needed for --browser).');
}

function defaultChrome() {
  if (process.env.PRESS_CLIP_CHROME) return process.env.PRESS_CLIP_CHROME;
  const mac = [
    '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome',
    '/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge',
  ];
  const linux = ['/usr/bin/google-chrome', '/usr/bin/chromium', '/usr/bin/chromium-browser', '/usr/bin/microsoft-edge'];
  const win = [
    'C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe',
    'C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe',
    'C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe',
  ];
  for (const p of [...mac, ...linux, ...win]) if (existsSync(p)) return p;
  return mac[0];
}

async function fetchHtmlBrowser(url, chromePath) {
  const chromium = loadChromium();
  const browser = await chromium.launch({ executablePath: chromePath || defaultChrome(), headless: true });
  try {
    const page = await browser.newPage({ userAgent: UA });
    await page.route('**/*', (route) => {
      const type = route.request().resourceType();
      if (['image', 'media', 'font', 'stylesheet'].includes(type)) return route.abort();
      return route.continue();
    });
    const res = await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 60000 });
    await page.waitForTimeout(1500);
    const html = await page.content();
    return { status: res?.status() || 0, finalUrl: page.url(), contentType: 'text/html', html, truncated: false, viaBrowser: true };
  } finally {
    await browser.close();
  }
}

// ---------------------------------------------------------------------------
// HTML → structure (a tolerant, dependency-free parser; good enough for articles)
// ---------------------------------------------------------------------------

const BLOCK_TAGS = new Set(['p', 'div', 'section', 'article', 'main', 'li', 'h1', 'h2', 'h3', 'h4', 'h5', 'h6', 'blockquote', 'figcaption', 'td', 'th', 'tr', 'dd', 'dt', 'pre', 'br', 'hr', 'header', 'footer', 'aside', 'nav', 'ul', 'ol', 'table', 'body', 'html', 'form', 'figure', 'summary', 'details']);
const DROP_TAGS = new Set(['script', 'style', 'noscript', 'svg', 'template', 'iframe', 'canvas', 'video', 'audio', 'select', 'textarea']);
const VOID_TAGS = new Set(['br', 'hr', 'img', 'meta', 'link', 'input', 'source', 'track', 'wbr', 'embed', 'param', 'area', 'col', 'base']);
const CHROME_TAGS = new Set(['nav', 'header', 'footer', 'aside']);

function decodeEntities(s) {
  return s
    .replace(/&nbsp;/g, ' ')
    .replace(/&amp;/g, '&')
    .replace(/&lt;/g, '<')
    .replace(/&gt;/g, '>')
    .replace(/&quot;/g, '"')
    .replace(/&#39;|&apos;/g, "'")
    .replace(/&ldquo;|&rdquo;/g, '"')
    .replace(/&lsquo;|&rsquo;/g, "'")
    .replace(/&mdash;/g, '—')
    .replace(/&ndash;/g, '–')
    .replace(/&hellip;/g, '…')
    .replace(/&#x([0-9a-f]+);/gi, (_, h) => { try { return String.fromCodePoint(parseInt(h, 16)); } catch { return ''; } })
    .replace(/&#(\d+);/g, (_, d) => { try { return String.fromCodePoint(Number(d)); } catch { return ''; } });
}

function attrs(tagBody) {
  const out = {};
  const re = /([a-zA-Z_:][-a-zA-Z0-9_:.]*)\s*(?:=\s*("([^"]*)"|'([^']*)'|([^\s"'>]+)))?/g;
  let m;
  while ((m = re.exec(tagBody))) {
    const name = m[1].toLowerCase();
    out[name] = decodeEntities(m[3] ?? m[4] ?? m[5] ?? '');
  }
  return out;
}

/**
 * Walks the HTML once and produces: paragraphs (with their links, Wikipedia
 * ref markers, and chrome flags), all links, meta tags, JSON-LD blocks,
 * <time> elements, title/h1, and MediaWiki-style reference lists.
 *
 * Dropped elements (script/style/svg/template…) are skipped by jumping to
 * their closing tag, so an unbalanced tag cannot swallow the rest of the page.
 * Chrome flags (nav/header/footer/aside ancestry) are recorded, never used to
 * delete text: real pages leave those elements unclosed often enough that
 * deleting would lose the article.
 */
function parseHtml(html, baseUrl) {
  const metas = [];
  const jsonld = [];
  const times = [];
  const links = [];
  const paragraphs = [];
  const wikiRefs = {};
  let title = '';
  let h1 = '';
  let canonical = null;
  let inTitle = false, inH1 = false;
  const ancestry = [];
  let cur = { text: '', links: [], refs: [], flags: new Set() };
  let curRef = null;
  let curRefText = '';
  let openA = null;
  let refDepth = 0;

  const base = (() => { try { return new URL(baseUrl); } catch { return null; } })();
  const abs = (href) => { try { return new URL(href, base || undefined).href; } catch { return null; } };

  const flushParagraph = () => {
    const text = cur.text.replace(/\s+/g, ' ').trim();
    if (text) paragraphs.push({ i: paragraphs.length, text, links: cur.links, refs: [...new Set(cur.refs)], flags: [...cur.flags] });
    cur = { text: '', links: [], refs: [], flags: new Set() };
  };
  const chromeFlags = () => ancestry.filter((t) => CHROME_TAGS.has(t));

  const re = /<!--[\s\S]*?-->|<!\[CDATA\[[\s\S]*?\]\]>|<![^>]*>|<\?[^>]*>|<\/?([a-zA-Z][a-zA-Z0-9:-]*)\b([^>]*)>|([^<]+)/g;
  let m;
  while ((m = re.exec(html))) {
    if (m[0].startsWith('<!') || m[0].startsWith('<?')) continue;
    const tag = m[1] ? m[1].toLowerCase() : null;
    const body = m[2] || '';
    const text = m[3];

    if (text !== undefined) {
      const t = decodeEntities(text);
      if (inTitle) { title += t; continue; }
      if (inH1) h1 += t;
      cur.text += t;
      if (openA) openA.text += t;
      if (curRef) curRefText += t;
      continue;
    }

    const closing = m[0].startsWith('</');
    const selfClosing = /\/\s*>$/.test(m[0]) || VOID_TAGS.has(tag);

    if (!closing && DROP_TAGS.has(tag) && !selfClosing) {
      // Jump straight to the matching close tag (no nesting for these elements in practice).
      // An unclosed <template>/<iframe>/… must not swallow the page: only script/style may run to the end.
      const closeRe = new RegExp(`</${tag}\\s*>`, 'ig');
      closeRe.lastIndex = re.lastIndex;
      const cm = closeRe.exec(html);
      if (!cm && tag !== 'script' && tag !== 'style') continue;
      const inner = html.slice(re.lastIndex, cm ? cm.index : html.length);
      if (tag === 'script' && /ld\+json/i.test(attrs(body).type || '')) jsonld.push(inner);
      re.lastIndex = cm ? cm.index + cm[0].length : html.length;
      continue;
    }
    if (closing && DROP_TAGS.has(tag)) continue;

    if (!closing) {
      const a = attrs(body);
      if (tag === 'meta') { metas.push(a); continue; }
      if (tag === 'link') {
        if ((a.rel || '').toLowerCase().split(/\s+/).includes('canonical') && a.href) canonical = abs(a.href);
        continue;
      }
      if (tag === 'title') inTitle = true;
      if (tag === 'h1') inH1 = true;
      if (tag === 'time' && a.datetime) times.push(a.datetime);
      if (tag === 'a' && a.href) {
        const href = abs(a.href);
        openA = { href, raw: a.href, text: '', rel: a.rel || '' };
        if (href && /^#cite_note-/.test(a.href)) cur.refs.push(a.href.slice(1));
      }
      if (tag === 'li' && /^cite_note-/.test(a.id || '')) { curRef = a.id; curRefText = ''; wikiRefs[a.id] = { urls: [], text: '' }; refDepth = ancestry.length; }
      if (tag === 'ol' && /\breferences\b/.test(a.class || '')) cur.flags.add('references');
      if (BLOCK_TAGS.has(tag)) { flushParagraph(); for (const f of chromeFlags()) cur.flags.add(f); }
      if (!selfClosing) ancestry.push(tag);
      if (BLOCK_TAGS.has(tag)) for (const f of chromeFlags()) cur.flags.add(f);
    } else {
      if (tag === 'title') inTitle = false;
      if (tag === 'h1') inH1 = false;
      if (tag === 'a' && openA) {
        const link = { href: openA.href, text: openA.text.replace(/\s+/g, ' ').trim(), rel: openA.rel };
        if (link.href && !/^#/.test(openA.raw) && !/^(javascript|mailto|tel):/i.test(openA.raw)) {
          links.push({ ...link, paragraph: paragraphs.length });
          cur.links.push(link);
          if (curRef) wikiRefs[curRef].urls.push(link.href);
        }
        openA = null;
      }
      if (tag === 'li' && curRef && ancestry.length - 1 <= refDepth) {
        wikiRefs[curRef].text = curRefText.replace(/\s+/g, ' ').trim().slice(0, 400);
        curRef = null;
      }
      const idx = ancestry.lastIndexOf(tag);
      if (idx >= 0) ancestry.splice(idx);
      if (BLOCK_TAGS.has(tag)) flushParagraph();
    }
  }
  flushParagraph();

  const ld = [];
  for (const raw of jsonld) {
    try { ld.push(JSON.parse(raw.trim())); } catch { /* ignore malformed */ }
  }
  return { title: title.replace(/\s+/g, ' ').trim(), h1: h1.replace(/\s+/g, ' ').trim(), canonical, metas, jsonld: ld, times, links, paragraphs, wikiRefs };
}

function metaGet(metas, ...names) {
  for (const n of names) {
    const hit = metas.find((m) => (m.property || m.name || m.itemprop || '').toLowerCase() === n.toLowerCase());
    if (hit && hit.content) return hit.content;
  }
  return null;
}

function ldWalk(node, fn, depth = 0) {
  if (!node || depth > 6) return;
  if (Array.isArray(node)) return node.forEach((n) => ldWalk(n, fn, depth + 1));
  if (typeof node === 'object') {
    fn(node);
    for (const k of Object.keys(node)) if (typeof node[k] === 'object') ldWalk(node[k], fn, depth + 1);
  }
}

function pickLd(jsonld) {
  const out = { types: [], datePublished: null, dateModified: null, author: null, publisher: null, isBasedOn: [], citation: [], isAccessibleForFree: null, mainEntityOfPage: null };
  const strOrName = (v) => (typeof v === 'string' ? v : v && typeof v === 'object' ? v.name || v['@id'] || v.url || null : null);
  for (const doc of jsonld) {
    ldWalk(doc, (n) => {
      const t = n['@type'];
      if (t) out.types.push(...(Array.isArray(t) ? t : [t]));
      if (n.datePublished && !out.datePublished) out.datePublished = n.datePublished;
      if (n.dateModified && !out.dateModified) out.dateModified = n.dateModified;
      if (n.author && !out.author) out.author = Array.isArray(n.author) ? n.author.map(strOrName).filter(Boolean).join(', ') : strOrName(n.author);
      if (n.publisher && !out.publisher) out.publisher = strOrName(n.publisher);
      if (n.isBasedOn) out.isBasedOn.push(...[].concat(n.isBasedOn).map(strOrName).filter(Boolean));
      if (n.citation) out.citation.push(...[].concat(n.citation).map(strOrName).filter(Boolean));
      if ((typeof n.isAccessibleForFree === 'boolean' || typeof n.isAccessibleForFree === 'string') && out.isAccessibleForFree === null) out.isAccessibleForFree = String(n.isAccessibleForFree) === 'true';
      if (n.mainEntityOfPage && !out.mainEntityOfPage) out.mainEntityOfPage = strOrName(n.mainEntityOfPage);
    });
  }
  out.types = [...new Set(out.types)];
  return out;
}

// ---------------------------------------------------------------------------
// Signals
// ---------------------------------------------------------------------------

const ATTRIBUTION_PATTERNS = [
  ['according to', /\baccording to\b/i],
  ['citing', /\bcit(?:ing|ed)\b/i],
  ['reported', /\b(?:first\s+)?reported(?:\s+by|\s+that|\s+on)?\b/i],
  ['said / told', /\b(?:said|says|told|tells)\b/i],
  ['study / survey / report', /\b(?:a|the|new|recent|an?|its|their)\s+(?:\w+\s+){0,3}?(?:study|survey|report|analysis|paper|poll|index|benchmark|investigation|white\s*paper)\b/i],
  ['data from', /\b(?:data|figures|numbers|statistics|stats|estimates?)\s+(?:from|by|compiled by|published by)\b/i],
  ['published by', /\bpublished\s+(?:by|in)\b/i],
  ['found that', /\b(?:found|finds|estimates?|estimated|shows|showed|revealed|reveals|suggests?)\s+that\b/i],
  ['per / via', /\b(?:per|via)\s+[A-Z]/],
  ['press release', /\b(?:press\s+release|announced|announcement|statement)\b/i],
  ['wikipedia', /\bwikipedia\b/i],
  ['originally published', /\b(?:originally\s+(?:appeared|published|ran)|republished|syndicated|reprinted)\b/i],
  ['first reported', /\b(?:first|originally|initially)\s+reported\b/i],
  ['traced to', /\b(?:traced|tracked|linked)\s+(?:back\s+)?to\b/i],
  ['sources line', /^\s*(?:sources?|references?|credit)\s*:/i],
  ['parenthetical credit', /\((?:[Ss]ource|[Vv]ia|[Pp]er|[Cc]redit)?\s*:?\s*[A-Z][^(),\d]{2,60}\)\s*\.?\s*$/],
];

const WIRE_PATTERNS = [
  ['reuters', /\(Reuters\)|\bReuters\b\s*[—–-]/],
  ['associated press', /\(AP\)|\bAssociated Press\b|\bBy the Associated Press\b/i],
  ['bloomberg', /\(Bloomberg\)|\bBloomberg\b\s*[—–-]/],
  ['afp', /\(AFP\)|\bAgence France-Presse\b/],
  ['pr newswire', /\/PRNewswire\/|\bPR\s*Newswire\b|prnewswire\.com/i],
  ['business wire', /\(BUSINESS WIRE\)|\bBusiness\s*Wire\b|businesswire\.com/i],
  ['globenewswire', /\bGlobe\s*Newswire\b|globenewswire\.com/i],
  ['accesswire', /\bAccess\s*wire\b|accesswire\.com/i],
  ['ein presswire', /\bEIN\s*Presswire\b|einpresswire\.com/i],
];

const SYNDICATION_TEXT = [
  ['originally published', /this (?:article|story|post|piece|column|report) (?:was )?(?:originally|first) (?:appeared|published|ran)/i],
  ['republished', /\b(?:republished|reprinted|syndicated)\s+(?:from|with permission|under|courtesy)/i],
  ['read the original', /\b(?:read|view|see) the (?:original|full) (?:article|story|report)\b/i],
  ['courtesy of', /\b(?:courtesy|via) (?:of )?[A-Z][\w.&' ]{2,40}\b/],
  ['partner content', /\b(?:partner content|this content was (?:produced|created) by|in partnership with)\b/i],
];

const SPONSORED_TEXT = [
  ['sponsored', /\bsponsored\b/i],
  ['paid post', /\bpaid (?:post|content|partnership|placement)\b/i],
  ['advertorial', /\badvertorial\b/i],
  ['brand studio', /\b(?:brand ?studio|branded content|promoted content)\b/i],
  ['affiliate', /\b(?:affiliate (?:links?|commission)|we may earn a commission)\b/i],
];

const SATIRE_TEXT = [
  ['april fools', /\bapril fool/i],
  ['satire', /\b(?:satir(?:e|ical)|parody|spoof|this (?:is|was) a joke|humou?r piece|not a real)\b/i],
];

const PAYWALL_TEXT = /\b(?:subscribe to (?:read|continue)|subscribers? only|this article is for subscribers|this post is for (?:paid|paying) (?:members|subscribers)|sign in to (?:read|continue)|already a subscriber|unlock this article|continue reading with|become a member to read|paywall|members[- ]only)\b/i;
const BOT_WALL_TEXT = /\b(?:prove your humanity|just a moment|verify (?:that )?you are (?:a )?human|access denied|attention required|enable javascript and cookies|are you a robot|checking your browser|cf-browser-verification|captcha)\b/i;

const NOT_A_SOURCE = /^(?:Monday|Tuesday|Wednesday|Thursday|Friday|Saturday|Sunday|January|February|March|April|May|June|July|August|September|October|November|December|The|A|An|In|On|At|It|He|She|They|We|I|This|That|These|Those|Its|His|Her|Their|Our|Mr|Ms|Mrs|Dr|One|Two|Three|Four|Five|Six|Seven|Eight|Nine|Ten|While|Although|But|And|However|Engineers|Officials|Researchers|Experts|Analysts|Critics|Sources)$/;
const NAME = "\\p{Lu}[\\p{L}\\p{N}.&'’-]*(?:\\s+(?:of|for|and|&|de|the|du|der|von|van)?\\s*\\p{Lu}[\\p{L}\\p{N}.&'’-]*){0,6}";
const SOURCE_NOUNS = 'study|survey|report|analysis|paper|poll|index|benchmark|estimate|forecast|data|release|announcement|statement|briefing|filing|post|column|article|interview|review|white\\s?paper|presentation|memo|letter|email|investigation|probe|piece|story|exposé|documentary|podcast|episode';
const ci = (phrase) => phrase.split('|').map((p) => `(?:${p[0].toUpperCase()}|${p[0]})${p.slice(1)}`).join('|');
const NAMED_SOURCE_PATTERNS = [
  // "according to the Wall Street Journal", "citing Reuters", "first reported by The Verge"
  new RegExp(`\\b(?:${ci('according to|citing|cited by|reported by|published by|data from|figures from|statistics from|estimates from|per|via|traced back to|traced to')}|[Rr]esearch (?:by|from))\\s+(?:the\\s+)?(${NAME})`, 'u'),
  // "a June 6 ProPublica investigation", "the 2025 IBM report"
  new RegExp(`\\b(?:[Aa]n?|[Tt]he|[Ii]ts|[Tt]heir)\\s+(?:(?:\\d{4}|\\w+\\s+\\d{1,2}|new|recent|latest|\\w+)\\s+){0,2}?(${NAME})\\s+(?:${SOURCE_NOUNS})\\b`, 'u'),
  // "a release from Keystone Analytics", "a 1993 column in PC Professional", "a study by Pew"
  new RegExp(`\\b(?:[Aa]n?|[Tt]he|[Nn]ew|[Rr]ecent|[Ll]atest|[Ii]ts|[Tt]heir)\\s+(?:\\w+\\s+){0,4}?(?:${SOURCE_NOUNS})\\s+(?:(?:published|released|issued|conducted|commissioned|sent|written|put out|carried out)\\s+(?:\\w+\\s+)?)?(?:by|from|in|at|for)\\s+(?:the\\s+)?(${NAME})`, 'u'),
  // "Rosa Ilić told reporters", "Microsoft found" — take the LAST such match in the sentence
  new RegExp(`(${NAME})\\s+(?:said|says|told|tells|reported|reports|found|finds|estimates|estimated|announced|wrote|writes|confirmed|added)\\b`, 'gu'),
];

/** Best-effort name of the attributed source in a sentence, or null. Judgment stays with the agent. */
function namedSourceIn(sentence) {
  for (const re of NAMED_SOURCE_PATTERNS) {
    let name = null;
    if (re.global) {
      for (const m of sentence.matchAll(re)) name = m[1];
    } else {
      const m = sentence.match(re);
      name = m ? m[1] : null;
    }
    if (!name) continue;
    name = name.replace(/[.,;:]+$/, '').trim();
    const first = name.split(/\s+/)[0];
    if (!name || (NOT_A_SOURCE.test(first) && name.split(/\s+/).length === 1)) continue;
    return name;
  }
  return null;
}

const STOPWORDS = new Set('a an the and or but of to in on at for by with from as is are was were be been being that this these those it its into than then there their they them he she his her we our you your i me my not no yes do does did have has had will would can could should may might also about over under more most less least very just only such which who whom whose what when where why how if so because while during before after between among per via each every any all some both either neither percent said says'.split(' '));

const WORD_NUMBERS = { zero: 0, one: 1, two: 2, three: 3, four: 4, five: 5, six: 6, seven: 7, eight: 8, nine: 9, ten: 10, eleven: 11, twelve: 12, thirteen: 13, fourteen: 14, fifteen: 15, sixteen: 16, seventeen: 17, eighteen: 18, nineteen: 19, twenty: 20, thirty: 30, forty: 40, fifty: 50, sixty: 60, seventy: 70, eighty: 80, ninety: 90, hundred: 100, thousand: 1000 };
const FRACTION_PHRASES = [
  [/\b(?:nearly |almost |about |roughly |around |over |more than )?(?:a |one[- ])third\b|\bone in three\b|\b1 in 3\b/i, '33 %'],
  [/\b(?:a |one[- ])quarter\b|\bone in four\b|\b1 in 4\b/i, '25 %'],
  [/\bhalf\b|\bone in two\b|\b1 in 2\b/i, '50 %'],
  [/\btwo[- ]thirds\b|\b2 in 3\b/i, '67 %'],
  [/\bthree[- ]quarters\b|\b3 in 4\b/i, '75 %'],
  [/\bone in five\b|\b1 in 5\b|\ba fifth\b/i, '20 %'],
  [/\bone in six\b|\b1 in 6\b/i, '17 %'],
  [/\bone in seven\b|\b1 in 7\b/i, '14 %'],
  [/\bone in eight\b|\b1 in 8\b/i, '13 %'],
  [/\bone in ten\b|\b1 in 10\b|\ba tenth\b/i, '10 %'],
  [/\bnine in ten\b|\b9 in 10\b/i, '90 %'],
];

function wordsToDigits(s) {
  // "sixty-five percent" → "65 percent"; "eight spiders" → "8 spiders"; "twelve seconds" → "12 seconds"
  return s.replace(/\b([a-z]+)(?:[- ]([a-z]+))?\b/gi, (m, a, b) => {
    const A = WORD_NUMBERS[a.toLowerCase()], B = b ? WORD_NUMBERS[b.toLowerCase()] : undefined;
    if (A === undefined) return m;
    if (b && B !== undefined) {
      if (A >= 20 && B < 10) return String(A + B);
      if (B >= 100) return String(A * B);
      return `${A} ${B}`;
    }
    if (b) return `${A} ${b}`;
    return String(A);
  });
}

function tokens(s) {
  return (wordsToDigits(s).toLowerCase().normalize('NFKD').replace(/[\u0300-\u036f]/g, '').replace(/[-–—_/]/g, ' ').match(/[a-z0-9][a-z0-9'%.$]*/g) || [])
    .map((t) => t.replace(/^[.']+|[.']+$/g, ''))
    .map((t) => (t.length > 5 ? t.replace(/(?:ings?|ed|es|s)$/, (m) => (t.length - m.length >= 4 ? '' : m)) : t))
    .filter((t) => t && !STOPWORDS.has(t));
}

function numbersIn(s) {
  const out = new Set();
  let text = wordsToDigits(s);
  for (const [re, norm] of FRACTION_PHRASES) if (re.test(text)) out.add(norm);
  const re = /(?:\$|€|£)?\d[\d,]*(?:\.\d+)?\s*(?:%|per\s?cent|percent|pct|million|billion|trillion|mn|m|bn|b|k|x)?(?![\w])/gi;
  let m;
  while ((m = re.exec(text))) {
    const raw = m[0].trim();
    const num = raw.replace(/[^\d.]/g, '');
    if (!num) continue;
    const unit = (raw.match(/(%|per\s?cent|percent|pct|million|billion|trillion|mn|m|bn|b|k|x)$/i) || [''])[0].toLowerCase().replace(/\s+/g, '');
    const normUnit = unit === 'percent' || unit === 'pct' ? '%' : unit === 'm' || unit === 'mn' ? 'million' : unit === 'bn' || unit === 'b' ? 'billion' : unit;
    out.add(num.replace(/\.0+$/, '') + (normUnit ? ' ' + normUnit : ''));
  }
  return [...out];
}

function numberNear(a, b) {
  // same unit and within ±1 absolute or ±5% relative → near match (drift candidates like 32% vs 33%)
  const [na, ua = ''] = a.split(' '), [nb, ub = ''] = b.split(' ');
  if (ua !== ub) return false;
  const x = Number(na), y = Number(nb);
  if (!Number.isFinite(x) || !Number.isFinite(y)) return false;
  return Math.abs(x - y) <= 1 || Math.abs(x - y) / Math.max(Math.abs(x), Math.abs(y), 1) <= 0.05;
}

const ABBREV = /\b(Jan|Feb|Mar|Apr|Jun|Jul|Aug|Sep|Sept|Oct|Nov|Dec|Mr|Mrs|Ms|Dr|Prof|Sr|Jr|St|No|Nos|vs|Inc|Ltd|Co|Corp|Gov|Sen|Rep|Gen|Col|Lt|Capt|Mt|Ft|U\.S|U\.K|E\.U|e\.g|i\.e|etc|approx|est|dept|Univ|Fig|al)\./g;
function splitSentences(text) {
  const protectedText = text.replace(ABBREV, (m) => m.replace('.', '\u2024'));
  return protectedText
    .split(/(?<=[.!?]["”’')\]]*(?:\[\d+\])*)\s+(?=[A-Z0-9“"(\[])/)
    .map((s) => s.replace(/\u2024/g, '.').trim())
    .filter((s) => s.length > 0);
}

function ngrams(toks, n) {
  const out = [];
  for (let i = 0; i + n <= toks.length; i++) out.push(toks.slice(i, i + n).join(' '));
  return out;
}

function longestCommonRun(aToks, bToks) {
  const bIndex = new Map();
  bToks.forEach((t, j) => { if (!bIndex.has(t)) bIndex.set(t, []); bIndex.get(t).push(j); });
  let best = 0, bestI = -1;
  for (let i = 0; i < aToks.length; i++) {
    for (const j of bIndex.get(aToks[i]) || []) {
      let k = 0;
      while (i + k < aToks.length && j + k < bToks.length && aToks[i + k] === bToks[j + k]) k++;
      if (k > best) { best = k; bestI = i; }
    }
  }
  return { length: best, span: best ? aToks.slice(bestI, bestI + best).join(' ') : '' };
}

function scoreSentence(claim, sentence) {
  const ct = tokens(claim), st = tokens(sentence);
  if (!ct.length || !st.length) return { score: 0 };
  const sset = new Set(st);
  const overlap = ct.filter((t) => sset.has(t)).length / ct.length;
  const cn = numbersIn(claim), sn = numbersIn(sentence);
  let numMatch = null, numNear = 0;
  if (cn.length) {
    const exact = cn.filter((n) => sn.includes(n)).length;
    numNear = cn.filter((n) => !sn.includes(n) && sn.some((m) => numberNear(n, m))).length;
    numMatch = exact / cn.length;
  }
  const lcr = longestCommonRun(ct, st);
  const quote = Math.min(1, lcr.length / 6);
  const numPart = numMatch === null ? 0.2 * overlap : 0.2 * (numMatch + 0.5 * (numNear / cn.length));
  const score = 0.55 * overlap + 0.25 * quote + numPart;
  return { score: Math.round(score * 1000) / 1000, token_overlap: Math.round(overlap * 100) / 100, number_match: numMatch, number_near_match: cn.length ? numNear / cn.length : null, longest_common_run: lcr.length, common_span: lcr.span };
}

function hostOf(u) { try { return new URL(u).hostname.replace(/^www\./, ''); } catch { return null; } }
function registrable(host) {
  if (!host) return null;
  if (/^\d{1,3}(\.\d{1,3}){3}$/.test(host) || host.includes(':') || host === 'localhost') return host;
  const parts = host.split('.');
  if (parts.length <= 2) return host;
  const twoLevel = /^(co|com|org|net|gov|ac|edu)\.[a-z]{2}$/.test(parts.slice(-2).join('.'));
  return parts.slice(twoLevel ? -3 : -2).join('.');
}

const TRACKING_PARAMS = /^(utm_\w+|fbclid|gclid|dclid|msclkid|mc_cid|mc_eid|ref|ref_src|source|srsltid|_hsenc|_hsmi|igshid|cmpid|ocid|ncid|sref|partner|via|pplx_\w+)$/i;
function stripTracking(u) {
  try {
    const url = new URL(u);
    for (const k of [...url.searchParams.keys()]) if (TRACKING_PARAMS.test(k)) url.searchParams.delete(k);
    url.hash = '';
    return url.href.replace(/\?$/, '');
  } catch { return u; }
}

// ---------------------------------------------------------------------------
// fetch subcommand
// ---------------------------------------------------------------------------

export async function analyzePage(html, { requestedUrl, finalUrl, status, contentType, claims = [], brief = false, maxParagraphs = 400, viaBrowser = false, truncated = false }) {
  const p = parseHtml(html, finalUrl || requestedUrl);
  const host = hostOf(finalUrl || requestedUrl);
  const canonicalHost = p.canonical ? hostOf(p.canonical) : null;
  const ld = pickLd(p.jsonld);

  // Dates
  const dateSources = [];
  const push = (source, value) => { if (value) dateSources.push({ source, value }); };
  push('meta:article:published_time', metaGet(p.metas, 'article:published_time', 'og:article:published_time'));
  push('jsonld:datePublished', ld.datePublished);
  push('meta:date', metaGet(p.metas, 'date', 'pubdate', 'publishdate', 'publish-date', 'dc.date', 'dc.date.issued', 'dcterms.created', 'parsely-pub-date', 'sailthru.date', 'article.published', 'citation_publication_date', 'citation_date', 'datePublished', 'publication_date'));
  if (p.times[0]) push('time:datetime', p.times[0]);
  const modSources = [];
  const pushMod = (source, value) => { if (value) modSources.push({ source, value }); };
  pushMod('meta:article:modified_time', metaGet(p.metas, 'article:modified_time', 'og:updated_time'));
  pushMod('jsonld:dateModified', ld.dateModified);
  pushMod('meta:last-modified', metaGet(p.metas, 'last-modified', 'dc.date.modified', 'dcterms.modified'));

  // Body: every paragraph is kept; chrome flags are informational. Pure-chrome paragraphs are
  // down-weighted in matching and excluded from signal scanning when article paragraphs exist.
  const all = p.paragraphs;
  const isChrome = (q) => q.flags.some((f) => CHROME_TAGS.has(f));
  const article = all.filter((q) => !isChrome(q));
  const scanSet = article.length >= 3 ? article : all;
  const bodyText = scanSet.map((q) => q.text).join('\n');
  const wordCount = all.map((q) => q.text).join(' ').split(/\s+/).filter(Boolean).length;
  const articleWordCount = article.map((q) => q.text).join(' ').split(/\s+/).filter(Boolean).length;

  // Outbound links. On a Wayback capture every link is rewritten to web.archive.org/web/<ts>/<url>;
  // unwrap those so the original hosts are visible (and mark them archived).
  const isArchive = /(^|\.)web\.archive\.org$/.test(host || '');
  const unArchive = (u) => { const m = u.match(/^https?:\/\/web\.archive\.org\/web\/\d{4,17}(?:[a-z_]{2,3})?\/(https?:\/\/.+)$/i); return m ? m[1] : null; };
  const pageHost = isArchive ? hostOf(unArchive(finalUrl || requestedUrl) || '') || host : host;
  const seen = new Set();
  const outbound = [];
  for (const l of p.links) {
    if (!l.href) continue;
    const inner = unArchive(l.href);
    const href = inner || l.href;
    if (seen.has(href)) continue;
    seen.add(href);
    const lh = hostOf(href);
    if (!lh || (isArchive && !inner && /web\.archive\.org$/.test(lh))) continue;
    const external = registrable(lh) !== registrable(pageHost);
    outbound.push({ href, host: lh, external, text: l.text.slice(0, 160), paragraph: l.paragraph, rel: l.rel || undefined, archived: inner ? true : undefined });
  }
  const externalLinks = outbound.filter((l) => l.external);

  // Attribution sentences (article paragraphs; chrome paragraphs only if nothing else)
  const attribution = [];
  for (const q of scanSet) {
    for (const s of splitSentences(q.text)) {
      const hits = ATTRIBUTION_PATTERNS.filter(([, re]) => re.test(s)).map(([k]) => k);
      if (!hits.length) continue;
      const next = all.find((r) => r.i === q.i + 1);
      attribution.push({
        paragraph: q.i, sentence: s.slice(0, 600), phrases: hits, named_source: namedSourceIn(s),
        links: q.links.filter((l) => l.href && l.text.length > 1 && s.includes(l.text)).map((l) => ({ href: l.href, text: l.text })),
        paragraph_links: q.links.map((l) => ({ href: l.href, text: l.text.slice(0, 120) })),
        next_paragraph_links: next ? next.links.map((l) => ({ href: l.href, text: l.text.slice(0, 120) })) : [],
        refs: q.refs,
      });
    }
  }

  // Wire / syndication / sponsored / satire / paywall / bot-wall signals
  const head = bodyText.slice(0, 1500), tail = bodyText.slice(-1200);
  const signals = [];
  for (const [k, re] of WIRE_PATTERNS) { const m = head.match(re) || tail.match(re); if (m) signals.push({ kind: 'wire', label: k, evidence: m[0].slice(0, 80) }); }
  for (const [k, re] of SYNDICATION_TEXT) { const m = head.match(re) || tail.match(re); if (m) signals.push({ kind: 'syndication', label: k, evidence: m[0].slice(0, 120) }); }
  for (const q of scanSet) {
    const short = q.text.split(/\s+/).length <= 12;
    const nearTop = q.i <= 6;
    for (const [k, re] of SPONSORED_TEXT) { const m = q.text.match(re); if (m && (short || nearTop) && !signals.some((s) => s.kind === 'sponsored' && s.label === k)) signals.push({ kind: 'sponsored', label: k, evidence: q.text.slice(0, 120) }); }
  }
  // Satire markers count only where a self-label would sit: the title, the opening, or a closing note.
  for (const [k, re] of SATIRE_TEXT) {
    const zone = `${p.title}\n${head.slice(0, 800)}\n${tail.slice(-600)}`;
    const m = zone.match(re);
    if (m) { const idx = zone.search(re); signals.push({ kind: 'satire', label: k, evidence: zone.slice(Math.max(0, idx - 60), idx + 80).replace(/\s+/g, ' '), note: 'a self-label on this page, or a mention of satire near its edges — read it' }); }
  }
  if (p.canonical && canonicalHost && registrable(canonicalHost) !== registrable(host)) signals.push({ kind: 'syndication', label: 'canonical points to another domain', evidence: p.canonical });
  else if (p.canonical && stripTracking(p.canonical).replace(/\/$/, '') !== stripTracking(finalUrl || requestedUrl).replace(/\/$/, '')) signals.push({ kind: 'syndication', label: 'canonical points to a different page', evidence: p.canonical });
  const siteName = metaGet(p.metas, 'og:site_name');
  if (ld.isBasedOn.length) signals.push({ kind: 'provenance', label: 'jsonld:isBasedOn', evidence: ld.isBasedOn.slice(0, 5).join(' | ') });
  if (ld.citation.length) signals.push({ kind: 'provenance', label: 'jsonld:citation', evidence: ld.citation.slice(0, 5).join(' | ') });
  const botWall = BOT_WALL_TEXT.test(p.title) || (wordCount < 200 && BOT_WALL_TEXT.test(bodyText));
  // Paywall only when the text is actually short: a readable article with a "subscribe" footer is not paywalled.
  const paywall = ld.isAccessibleForFree === false || (status === 401 || status === 402) || (PAYWALL_TEXT.test(bodyText) && articleWordCount < 350) || (status === 200 && articleWordCount < 120 && /subscri|sign in|log in|member/i.test(bodyText) && !botWall);
  const redirected = (finalUrl && requestedUrl && stripTracking(finalUrl) !== stripTracking(requestedUrl)) || false;

  // MediaWiki-style reference lists
  const isWikipedia = /(^|\.)wikipedia\.org$/.test(host || '');
  const wikipediaReferences = Object.entries(p.wikiRefs)
    .map(([id, r]) => ({ id, urls: [...new Set(r.urls.filter((u) => !/wikipedia\.org\/wiki\//.test(u)))], text: r.text }))
    .filter((r) => r.urls.length || r.text);

  // Claim matching: single sentences and two-sentence windows; headline-like and chrome paragraphs down-weighted
  const claimMatches = claims.map((claim) => {
    const cands = [];
    for (const q of all) {
      const sents = splitSentences(q.text);
      const headlineLike = sents.length === 1 && !/[.!?]["”’']?$/.test(q.text) && q.text.split(/\s+/).length < 16;
      const weight = (isChrome(q) ? 0.8 : 1) * (headlineLike ? 0.75 : 1);
      const scores = sents.map((s) => scoreSentence(claim, s));
      for (let k = 0; k < sents.length; k++) {
        const s1 = sents[k];
        const sc = scores[k];
        if (sc.score > 0.1) cands.push({ ...sc, score: Math.round(sc.score * weight * 1000) / 1000, paragraph: q.i, sentence: s1.slice(0, 700), window: 1 });
        if (k + 1 < sents.length) {
          // A two-sentence window only counts when it beats both of its sentences on their own
          // (a claim whose halves sit in adjacent sentences), never as a tie-breaker.
          const s2 = `${s1} ${sents[k + 1]}`;
          const sc2 = scoreSentence(claim, s2);
          if (sc2.score > Math.max(sc.score, scores[k + 1].score) + 0.08) cands.push({ ...sc2, score: Math.round(sc2.score * weight * 1000) / 1000, paragraph: q.i, sentence: s2.slice(0, 900), window: 2 });
        }
      }
    }
    cands.sort((a, b) => b.score - a.score || a.window - b.window);
    const top = cands.slice(0, 4).map((c) => {
      const para = all.find((q) => q.i === c.paragraph) || { links: [], refs: [], text: '' };
      const next = all.find((q) => q.i === c.paragraph + 1);
      // Reference markers inside the matched sentence ("…Leeds.[1]") pick that sentence's refs; otherwise the paragraph's.
      const markers = (c.sentence.match(/\[(\d+)\]/g) || []).map((m) => m.slice(1, -1));
      const refIds = (para.refs || []).filter((id) => !markers.length || markers.some((n) => id === `cite_note-${n}` || id.endsWith(`-${n}`)));
      const refs = (refIds.length ? refIds : (para.refs || [])).map((id) => ({ id, urls: (p.wikiRefs[id]?.urls || []).filter((u) => !/wikipedia\.org\/wiki\//.test(u)), text: p.wikiRefs[id]?.text || '', sentence_level: markers.length > 0 && refIds.length > 0 }));
      const att = attribution.filter((a) => a.paragraph === c.paragraph && c.sentence.includes(a.sentence.slice(0, 80)));
      return {
        ...c,
        paragraph_text: para.text.slice(0, 1200),
        paragraph_links: para.links.map((l) => ({ href: l.href, text: l.text.slice(0, 120) })),
        next_paragraph_links: next ? next.links.map((l) => ({ href: l.href, text: l.text.slice(0, 120) })) : [],
        wikipedia_refs: refs,
        attribution: att.map((a) => ({ phrases: a.phrases, named_source: a.named_source })),
      };
    });
    const best = top[0];
    const hint = !best ? 'no_match' : best.score >= 0.6 ? 'strong_match' : best.score >= 0.35 ? 'partial_match' : 'weak_match';
    return { claim, numbers_in_claim: numbersIn(claim), best: top, verdict_hint: hint, note: 'verdict_hint locates the sentence; whether the page supports the claim is the agent\'s reading.' };
  });

  let hint = null;
  if (/pdf/i.test(contentType)) hint = 'This is a PDF, which the script does not parse. Download it and read it with your PDF tool (or `pdftotext`); quote from that text and note `quote_via: page`.';
  else if (botWall) hint = viaBrowser
    ? 'Bot challenge page (Cloudflare or similar) even in browser mode. Try the Wayback snapshot (`wayback <url>`), a reader/API route for this site, or read it in your own browser; record the page as unreachable if none works.'
    : 'Bot challenge page. Retry with --browser; if that also fails, try the Wayback snapshot or read it in your own browser.';
  else if (status >= 400) hint = `HTTP ${status}. ${viaBrowser ? 'Already in browser mode — ' : 'Try --browser (uses your Chrome); otherwise '}try the Wayback snapshot (\`wayback <url>\`). Do not treat the page as absent.`;
  else if (status === 200 && articleWordCount < 25 && wordCount < 80) hint = viaBrowser
    ? 'Very little text even in browser mode: the page may be blocked, empty, or a consent wall. Try the Wayback snapshot or read it in your own browser.'
    : 'Very little text extracted. The page may render with JavaScript — retry with --browser — or be a paywall/consent wall.';

  const out = {
    tool: 'citation-foodchain/trace.mjs', version: VERSION, fetched_at: new Date().toISOString(),
    requested_url: requestedUrl, final_url: finalUrl, redirected, status, content_type: contentType, via_browser: viaBrowser || undefined, truncated: truncated || undefined,
    host, registrable_domain: registrable(host), canonical_url: p.canonical, canonical_host: canonicalHost, canonical_host_mismatch: !!(canonicalHost && registrable(canonicalHost) !== registrable(host)),
    title: p.title || null, h1: p.h1 || null, site_name: siteName,
    byline: [ld.author, metaGet(p.metas, 'author', 'parsely-author', 'sailthru.author', 'dc.creator', 'citation_author', 'byl'), metaGet(p.metas, 'article:author')].find((v) => v && !/^https?:\/\//i.test(v)) || null,
    publisher: ld.publisher || siteName || null,
    published_at: dateSources[0]?.value || null, published_at_sources: dateSources,
    modified_at: modSources[0]?.value || null, modified_at_sources: modSources,
    jsonld_types: ld.types, jsonld_is_based_on: ld.isBasedOn, jsonld_citation: ld.citation,
    word_count: wordCount, article_word_count: articleWordCount, paragraph_count: all.length,
    outbound_link_count: outbound.length, external_link_count: externalLinks.length,
    external_domains: [...new Set(externalLinks.map((l) => registrable(l.host)))].slice(0, 80),
    has_byline: !!(metaGet(p.metas, 'author', 'article:author') || ld.author), has_date: !!dateSources.length,
    paywall_suspected: !!paywall, bot_wall: !!botWall, is_wikipedia: isWikipedia,
    signals, attribution_sentences: attribution.slice(0, 80),
    wikipedia_references: wikipediaReferences.slice(0, 400),
    claim_matches: claimMatches,
    external_links: externalLinks.slice(0, 300),
    hint,
  };
  if (!brief) { out.paragraphs = all.slice(0, maxParagraphs).map((q) => ({ i: q.i, text: q.text, links: q.links.map((l) => l.href), refs: q.refs, flags: q.flags.length ? q.flags : undefined })); out.all_links = outbound; }
  else out.body_excerpt = scanSet.slice(0, 12).map((q) => q.text.slice(0, 500));
  return out;
}

async function fetchAny(url, args) {
  let got;
  let error = null;
  try {
    got = args.browser ? await fetchHtmlBrowser(url, args.chrome) : await fetchHtml(url);
  } catch (e) {
    error = String(e?.message || e);
    if (e?.cause?.code) error += ` (${e.cause.code})`;
    got = { status: 0, finalUrl: url, contentType: '', html: '' };
  }
  return { got, error };
}

async function cmdFetch(args) {
  const url = args._[1];
  if (!url) throw new Error('usage: fetch <url> [--claim "text"]... [--browser] [--brief] [--no-wayback] [--out file]');
  const { got, error } = await fetchAny(url, args);
  const res = await analyzePage(got.html, { requestedUrl: url, finalUrl: got.finalUrl, status: got.status, contentType: got.contentType, claims: args.claim, brief: args.brief, maxParagraphs: args.maxParagraphs || 400, viaBrowser: got.viaBrowser, truncated: got.truncated });
  if (error) { res.fetch_error = error; res.hint = `Fetch failed: ${error}. ${args.browser ? 'The browser could not load the page; ' : 'Retry with --browser; '}then try the Wayback snapshot (\`wayback <url>\`). Record the page as unreachable if nothing works.`; }
  if (!args.noWayback && /^https?:/.test(url)) {
    res.wayback = await waybackEarliest(url);
  }
  emit(res, args.out);
}

// ---------------------------------------------------------------------------
// unwrap
// ---------------------------------------------------------------------------

export function unwrapStatic(u) {
  try {
    const url = new URL(u);
    const h = url.hostname;
    if (/^(www\.)?google\.[a-z.]+$/.test(h) && url.pathname === '/url') {
      const q = url.searchParams.get('q') || url.searchParams.get('url');
      if (q) return { url: stripTracking(q), method: 'google-url-param' };
    }
    if (/duckduckgo\.com$/.test(h) && url.pathname.startsWith('/l/')) {
      const q = url.searchParams.get('uddg');
      if (q) return { url: stripTracking(q), method: 'ddg-uddg' };
    }
    if (/bing\.com$/.test(h) && url.pathname.startsWith('/ck/')) {
      const b = url.searchParams.get('u');
      if (b && b.startsWith('a1')) {
        try {
          const dec = Buffer.from(b.slice(2).replace(/-/g, '+').replace(/_/g, '/'), 'base64').toString('utf8');
          if (/^https?:/.test(dec)) return { url: stripTracking(dec), method: 'bing-ck-base64' };
        } catch {}
      }
    }
    if (/vertexaisearch\.cloud\.google\.com$/.test(h) || /grounding-api-redirect/.test(url.pathname)) return { url: null, method: 'needs-request', note: 'Google grounding redirect — follow it with one request' };
    const stripped = stripTracking(u);
    return { url: stripped, method: stripped === u ? 'none' : 'strip-tracking-params' };
  } catch { return { url: u, method: 'invalid-url' }; }
}

async function cmdUnwrap(args) {
  const u = args._[1];
  if (!u) throw new Error('usage: unwrap <url> [--follow]');
  const s = unwrapStatic(u);
  const hops = [];
  let current = s.url || u;
  if (s.method === 'needs-request' || args.follow) {
    current = u;
    for (let i = 0; i < 6; i++) {
      let res;
      try { res = await fetchWithTimeout(current, { redirect: 'manual', timeout: 10000 }); } catch (e) { hops.push({ url: current, error: String(e?.message || e) }); break; }
      hops.push({ url: current, status: res.status });
      const loc = res.headers.get('location');
      if (res.status >= 300 && res.status < 400 && loc) { current = new URL(loc, current).href; continue; }
      break;
    }
  }
  emit({ input: u, resolved_url: stripTracking(current), static: s, hops: hops.length ? hops : undefined }, args.out);
}

// ---------------------------------------------------------------------------
// wayback
// ---------------------------------------------------------------------------

export async function waybackEarliest(url) {
  const api = `https://web.archive.org/cdx/search/cdx?url=${encodeURIComponent(url)}&output=json&limit=1&fl=timestamp,original,statuscode&filter=statuscode:200`;
  let lastErr = null;
  for (let attempt = 0; attempt < 3; attempt++) {
    try {
      const res = await fetchWithTimeout(api, { timeout: 20000, headers: { accept: 'application/json' } });
      if (res.status === 429 || res.status >= 500) { lastErr = `HTTP ${res.status}`; await sleep(1500 * (attempt + 1)); continue; }
      if (!res.ok) return { available: false, error: `HTTP ${res.status}`, earliest_capture: null };
      const rows = await res.json();
      if (!Array.isArray(rows) || rows.length < 2) return { available: true, earliest_capture: null, note: 'no capture found (not proof the page is new)' };
      const [ts, original] = rows[1];
      const iso = `${ts.slice(0, 4)}-${ts.slice(4, 6)}-${ts.slice(6, 8)}T${ts.slice(8, 10)}:${ts.slice(10, 12)}:${ts.slice(12, 14)}Z`;
      return { available: true, earliest_capture: iso, snapshot_url: `https://web.archive.org/web/${ts}/${original}`, raw_snapshot_url: `https://web.archive.org/web/${ts}id_/${original}`, note: 'lower bound on page age only; absence of an earlier capture proves nothing' };
    } catch (e) {
      lastErr = String(e?.message || e) + (e?.cause?.code ? ` (${e.cause.code})` : '');
      await sleep(1500 * (attempt + 1));
    }
  }
  return { available: false, error: `Wayback unreachable after 3 attempts: ${lastErr}`, earliest_capture: null, note: 'network failure, NOT "no capture" — retry later or open web.archive.org in your browser' };
}

async function cmdWayback(args) {
  const u = args._[1];
  if (!u) throw new Error('usage: wayback <url>');
  emit({ url: u, ...(await waybackEarliest(u)) }, args.out);
}

// ---------------------------------------------------------------------------
// reuse (text reuse between two pages)
// ---------------------------------------------------------------------------

function pageText(doc) {
  if (doc.paragraphs) return doc.paragraphs.filter((q) => !(q.flags || []).some((f) => CHROME_TAGS.has(f))).map((q) => q.text).join('\n') || doc.paragraphs.map((q) => q.text).join('\n');
  if (doc.body_excerpt) return doc.body_excerpt.join('\n');
  return '';
}

/**
 * Text reuse between two pages. Parameters follow the published news-reuse
 * literature: 7-word shingles for precision, 5-word for recall (Nicholls 2019,
 * IJoC), and the churnalism.com flag at >20% overlap (Media Standards Trust 2011).
 */
export function textReuse(aText, bText, n = 7) {
  const at = tokens(aText), bt = tokens(bText);
  const shingles = (k) => {
    const ag = new Set(ngrams(at, k)), bg = new Set(ngrams(bt, k));
    let shared = 0;
    for (const g of ag) if (bg.has(g)) shared++;
    return { a: ag.size, b: bg.size, shared, a_in_b: ag.size ? Math.round((shared / ag.size) * 1000) / 1000 : 0, b_in_a: bg.size ? Math.round((shared / bg.size) * 1000) / 1000 : 0 };
  };
  const p = shingles(n), r = shingles(5);
  const lcr = longestCommonRun(at, bt);
  const an = new Set(numbersIn(aText)), bn = new Set(numbersIn(bText));
  const sharedNumbers = [...an].filter((x) => bn.has(x));
  const maxOverlap = Math.max(p.a_in_b, p.b_in_a);
  const tooSmall = Math.min(at.length, bt.length) < 12;
  return {
    shingle_size: n,
    a_tokens: at.length, b_tokens: bt.length,
    a_shingles: p.a, b_shingles: p.b, shared_shingles: p.shared,
    a_in_b: p.a_in_b, b_in_a: p.b_in_a,
    recall_5gram: { shared: r.shared, a_in_b: r.a_in_b, b_in_a: r.b_in_a },
    longest_common_run_tokens: lcr.length, longest_common_span: lcr.span.slice(0, 400),
    shared_numbers: sharedNumbers,
    churn_flag: !tooSmall && maxOverlap > 0.2,
    reading: tooSmall ? 'one page has too little text to compare (fetch it fully or with --browser first)' : p.shared === 0 ? 'no verbatim reuse' : (lcr.length >= 25 || maxOverlap > 0.2) ? 'substantial verbatim reuse (wire copy, syndication, or rewrite of one by the other)' : 'some shared phrasing (quotes or boilerplate)',
  };
}

async function loadDocForReuse(ref, args) {
  if (/^https?:\/\//.test(ref)) {
    const { got, error } = await fetchAny(ref, args);
    const doc = await analyzePage(got.html, { requestedUrl: ref, finalUrl: got.finalUrl, status: got.status, contentType: got.contentType, viaBrowser: got.viaBrowser });
    if (error) doc.fetch_error = error;
    return doc;
  }
  return JSON.parse(readFileSync(ref, 'utf8'));
}

async function cmdReuse(args) {
  const [, a, b] = args._;
  if (!a || !b) throw new Error('usage: reuse <a.json|url> <b.json|url> [--browser]');
  const A = await loadDocForReuse(a, args), B = await loadDocForReuse(b, args);
  const r = textReuse(pageText(A), pageText(B));
  const dates = { a_published_at: A.published_at || null, b_published_at: B.published_at || null, a_wayback: A.wayback?.earliest_capture || null, b_wayback: B.wayback?.earliest_capture || null };
  let direction = 'unknown';
  const da = Date.parse(A.published_at || A.wayback?.earliest_capture || ''), db = Date.parse(B.published_at || B.wayback?.earliest_capture || '');
  if (!Number.isNaN(da) && !Number.isNaN(db) && r.shared_shingles > 0) {
    const gapH = Math.abs(da - db) / 36e5;
    direction = gapH < 24 ? `too_close_to_order (${gapH.toFixed(1)}h apart; the skill does not order pages inside a day)` : da < db ? 'a_before_b' : 'b_before_a';
  }
  emit({ a: A.final_url, b: B.final_url, ...r, dates, earlier_by_timestamp: direction, caution: 'Timestamps order the pages; they do not prove who copied whom. A wire or press release can predate both.' }, args.out);
}

// ---------------------------------------------------------------------------
// share (capture answer + first-degree citations from a public share link)
// ---------------------------------------------------------------------------

function balancedJson(text, start) {
  const open = text[start];
  if (open !== '{' && open !== '[') return null;
  const close = open === '{' ? '}' : ']';
  let depth = 0, i = start, inStr = false;
  while (i < text.length) {
    const c = text[i];
    if (inStr) {
      if (c === '\\') { i += 2; continue; }
      if (c === '"') inStr = false;
    } else if (c === '"') inStr = true;
    else if (c === open) depth++;
    else if (c === close) { depth--; if (depth === 0) break; }
    i++;
  }
  try { return [JSON.parse(text.slice(start, i + 1)), i + 1]; } catch { return null; }
}

function cleanCitationUrls(urls) {
  const out = [];
  const seen = new Set();
  for (const u of urls || []) {
    if (typeof u !== 'string' || !/^https?:/.test(u)) continue;
    if (/images\.openai\.com|oaiusercontent\.com/.test(u)) continue;
    const s = stripTracking(u);
    if (seen.has(s)) continue;
    seen.add(s);
    out.push(s);
  }
  return out;
}

function lastSentenceBefore(text, idx, max = 400) {
  const before = text.slice(Math.max(0, idx - max), idx);
  const parts = before.split(/(?<=[.!?]["”’']?)\s+/);
  return (parts[parts.length - 1] || before).trim();
}

/**
 * ChatGPT share page. The SSR HTML carries one "frame" entry per message:
 *   {"key":"<message id>","origin":"server","frame":{"id","role",...},
 *    "source":{"messageId","markdown","references":[{messageId,startIndex,endIndex,approvedUrls,data:{...}}]}}
 * `data.type` is "grouped_webpages" (a per-span citation: safe_urls, items with title/url/snippet)
 * or "sources_footnote" (the Sources panel). Index fields point into `markdown`.
 * A smaller "serialized-conversation-state" blob (recent messages only) is the fallback.
 */
export function parseChatGptShare(html) {
  const messages = [];
  const citations = [];
  const footnotes = [];
  const seenMsg = new Set();

  const refToCitation = (ref, messageId, markdown) => {
    const data = ref.data || ref;
    const startIndex = Number.isFinite(ref.startIndex) ? ref.startIndex : (Number.isFinite(data.start_idx) ? data.start_idx : null);
    const endIndex = Number.isFinite(ref.endIndex) ? ref.endIndex : (Number.isFinite(data.end_idx) ? data.end_idx : null);
    if (data.type === 'sources_footnote') {
      footnotes.push({ message_id: messageId, sources: (data.sources || []).map((s) => ({ title: s.title, url: stripTracking(s.url), attribution: s.attribution })) });
      return;
    }
    if (data.type !== 'grouped_webpages') return;
    citations.push({
      message_id: messageId,
      start_index: startIndex, end_index: endIndex,
      preceding_sentence: markdown && Number.isFinite(startIndex) ? lastSentenceBefore(markdown, startIndex).replace(/\s*\(\[[^\]]*\]\([^)]*\)\)/g, '').trim() || null : null,
      urls: cleanCitationUrls(data.safe_urls),
      alt_markdown: data.alt || null,
      items: (data.items || []).map((it) => ({ title: it.title, url: stripTracking(it.url), attribution: it.attribution, pub_date: it.pub_date || null, snippet: it.snippet || null, supporting: (it.supporting_websites || []).map((s) => ({ title: s.title, url: stripTracking(s.url), snippet: s.snippet || null })) })),
    });
  };

  let pos = 0;
  while (true) {
    const f = html.indexOf('"origin":"server","frame":{"id":"', pos);
    if (f < 0) break;
    pos = f + 10;
    const start = html.lastIndexOf('{"key":"', f);
    if (start < 0 || f - start > 200) continue;
    const parsed = balancedJson(html, start);
    if (!parsed) continue;
    const [entry] = parsed;
    const frame = entry.frame || {};
    const source = entry.source || {};
    const id = source.messageId || frame.id || entry.key;
    const role = frame.role || (frame.attribution && /you said/i.test(frame.attribution) ? 'user' : 'assistant');
    const markdown = typeof source.markdown === 'string' ? source.markdown : null;
    if (id && !seenMsg.has(id) && markdown !== null) { seenMsg.add(id); messages.push({ id, role, content: markdown }); }
    const refs = source.references || entry.references || frame.references || [];
    for (const ref of refs) refToCitation(ref, id, markdown);
  }

  if (!messages.length) {
    const stateKey = html.indexOf('"key":"serialized-conversation-state"');
    if (stateKey >= 0) {
      const v = html.indexOf('["string","', stateKey);
      if (v >= 0) {
        const strStart = v + '["string",'.length;
        let i = strStart + 1;
        while (i < html.length) { if (html[i] === '\\') { i += 2; continue; } if (html[i] === '"') break; i++; }
        try {
          const state = JSON.parse(JSON.parse(html.slice(strStart, i + 1)));
          for (const m of state.messages || []) messages.push({ id: m.id, role: m.role, content: typeof m.content === 'string' ? m.content : JSON.stringify(m.content) });
        } catch { /* ignore */ }
      }
    }
  }

  if (!citations.length && !footnotes.length) {
    let p = 0;
    while (true) {
      const d = html.indexOf('"data":{"matched_text"', p);
      if (d < 0) break;
      p = d + 10;
      const parsed = balancedJson(html, d + 7);
      if (parsed) refToCitation({ data: parsed[0] }, null, null);
    }
  }

  const seen = new Set();
  const uniq = citations.filter((c) => { const k = `${c.message_id}|${c.start_index}|${c.urls.join(',')}`; if (seen.has(k)) return false; seen.add(k); return true; });
  const fseen = new Set();
  const funiq = footnotes.filter((f) => { const k = `${f.message_id}|${f.sources.map((s) => s.url).join(',')}`; if (fseen.has(k)) return false; fseen.add(k); return true; });
  const allUrls = cleanCitationUrls([...uniq.flatMap((c) => c.urls), ...funiq.flatMap((f) => f.sources.map((s) => s.url))]);
  return { engine: 'chatgpt', messages, citations: uniq, sources_footnotes: funiq, cited_urls: allUrls, per_claim_mapping: uniq.some((c) => c.preceding_sentence) };
}

/** Perplexity thread JSON (`/rest/thread/<slug>?with_parent_info=true&source=web`). */
export function parsePerplexityThread(json) {
  const out = { engine: 'perplexity', turns: [], cited_urls: [], per_claim_mapping: true };
  for (const e of json.entries || []) {
    let steps = e.text;
    if (typeof steps === 'string') { try { steps = JSON.parse(steps); } catch { steps = []; } }
    const fin = (Array.isArray(steps) ? steps : []).find((s) => s.step_type === 'FINAL');
    let ans = fin?.content?.answer;
    if (typeof ans === 'string') { try { ans = JSON.parse(ans); } catch { ans = { answer: ans }; } }
    const web = [...(ans?.web_results || []), ...(ans?.extra_web_results || [])].map((w, i) => ({ n: i + 1, title: w.name, url: stripTracking(w.url), snippet: w.snippet || null }));
    const text = ans?.answer || '';
    const sentences = splitSentences(text).map((s) => ({ sentence: s, refs: [...new Set((s.match(/\[(\d+)\]/g) || []).map((m) => Number(m.slice(1, -1))))].map((n) => web[n - 1]?.url).filter(Boolean) })).filter((s) => s.refs.length);
    out.turns.push({ query: e.query_str, answer: text, web_results: web, sentence_citations: sentences, created: e.updated_datetime || null });
    out.cited_urls.push(...web.map((w) => w.url));
  }
  out.cited_urls = cleanCitationUrls(out.cited_urls);
  return out;
}

async function cmdShare(args) {
  const u = args._[1];
  if (!u) throw new Error('usage: share <share-url> [--html saved.html] [--browser] [--out file]');
  let host;
  try { host = new URL(u).hostname; } catch { throw new Error('not a URL'); }
  let result;
  if (/(^|\.)chatgpt\.com$/.test(host) || /chat\.openai\.com$/.test(host)) {
    const html = args.html ? readFileSync(args.html, 'utf8') : (args.browser ? await fetchHtmlBrowser(u, args.chrome) : await fetchHtml(u)).html;
    result = parseChatGptShare(html);
    if (!result.messages.length && !result.citations.length) result.hint = 'No conversation found in the page. The link may be private/expired, or the page shape changed — retry with --browser and read the sources panel from the rendered page.';
  } else if (/perplexity\.ai$/.test(host)) {
    const slug = new URL(u).pathname.split('/').filter(Boolean).pop();
    const api = `https://www.perplexity.ai/rest/thread/${slug}?with_parent_info=true&source=web`;
    const res = await fetchWithTimeout(api, { headers: { accept: 'application/json' } });
    if (!res.ok) throw new Error(`perplexity thread endpoint returned ${res.status}; the thread may be private or deleted — fall back to --browser on the share page`);
    result = parsePerplexityThread(await res.json());
    result.thread_api = api;
  } else {
    result = { engine: 'other', hint: `No plain-fetch parser for ${host}. Claude, Gemini, Google AI Mode, Grok and Copilot share pages render client-side: open the link with --browser (or in your own browser) and read the answer and its sources panel from the rendered page.` };
    if (args.browser) {
      const got = await fetchHtmlBrowser(u, args.chrome);
      const page = await analyzePage(got.html, { requestedUrl: u, finalUrl: got.finalUrl, status: got.status, contentType: got.contentType, viaBrowser: true });
      result.rendered = { title: page.title, external_links: page.external_links, paragraphs: page.paragraphs };
    }
  }
  result.share_url = u;
  result.captured_at = new Date().toISOString();
  result.note = 'first-degree citations = the URLs the engine itself attached to this answer, tracking params stripped. One capture is one sample; re-runs may cite different pages.';
  emit(result, args.out);
}

// ---------------------------------------------------------------------------
// batch
// ---------------------------------------------------------------------------

async function cmdBatch(args) {
  const list = args._[1];
  if (!list) throw new Error('usage: batch <urls.txt> [--out dir] [--claim ...]');
  const urls = readFileSync(list, 'utf8').split(/\r?\n/).map((s) => s.trim()).filter((s) => s && !s.startsWith('#'));
  const outDir = args.out || 'citation-foodchain-out';
  mkdirSync(outDir, { recursive: true });
  const index = [];
  for (const u of urls) {
    const slug = u.replace(/^https?:\/\//, '').replace(/[^a-z0-9]+/gi, '-').replace(/^-|-$/g, '').slice(0, 90).toLowerCase();
    const file = join(outDir, `${slug}.json`);
    try {
      await cmdFetch({ ...args, _: ['fetch', u], out: file });
      index.push({ url: u, file, ok: true });
    } catch (e) {
      index.push({ url: u, file, ok: false, error: String(e?.message || e) });
    }
  }
  emit({ out_dir: outDir, fetched: index }, undefined);
}

// ---------------------------------------------------------------------------
// demo fixture (offline synthetic chains for tests)
// ---------------------------------------------------------------------------

export function fixturePages() {
  const pr = `<!doctype html><html><head><title>Northline Logistics Reports 14% Failed-First-Attempt Rate Across 12 Canadian Cities</title>
<meta property="og:site_name" content="PR Newswire"><meta property="article:published_time" content="2026-02-03T13:00:00Z">
<link rel="canonical" href="http://127.0.0.1:PORT/prnewswire/northline-release.html">
<script type="application/ld+json">{"@context":"https://schema.org","@type":"NewsArticle","datePublished":"2026-02-03T13:00:00Z","publisher":{"@type":"Organization","name":"PR Newswire"},"author":{"@type":"Organization","name":"Northline Logistics"}}</script></head>
<body><header><nav><a href="/">Home</a> <a href="/news">News</a></nav></header>
<main><article><h1>Northline Logistics Reports 14% Failed-First-Attempt Rate Across 12 Canadian Cities</h1>
<p>TORONTO, Feb. 3, 2026 /PRNewswire/ -- Northline Logistics today released its Q1 Last-Mile Benchmark, which measured 18,240 deliveries across 12 Canadian cities between January and March 2026 and found a 14% failed-first-attempt rate.</p>
<p>"Failed first attempts cost Canadian retailers an estimated $212 million a year," said Priya Raman, Chief Operating Officer of Northline Logistics. The full methodology is available at <a href="http://127.0.0.1:PORT/northline/benchmark.html">northline.example/benchmark</a>.</p>
<p>About Northline Logistics: Northline is a last-mile delivery company serving retailers across Canada.</p></article></main>
<footer><p>SOURCE Northline Logistics</p></footer></body></html>`;

  const study = `<!doctype html><html><head><title>Q1 2026 Last-Mile Benchmark — Northline Logistics</title>
<meta name="author" content="Northline Logistics Research"><meta property="article:published_time" content="2026-02-03T12:00:00Z">
<script type="application/ld+json">{"@context":"https://schema.org","@type":"Report","datePublished":"2026-02-03","publisher":{"@type":"Organization","name":"Northline Logistics"}}</script></head>
<body><main><h1>Q1 2026 Last-Mile Benchmark</h1>
<p>Method: we logged every delivery attempt handled by Northline drivers in 12 Canadian cities from January 5 to January 27, 2026 — 18,240 deliveries in total. A failed first attempt is any delivery not completed on the first visit.</p>
<p>Result: 2,554 of 18,240 deliveries (14%) failed on the first attempt. The rate ranged from 9% in Halifax to 19% in Vancouver.</p>
<p>The $212 million annual cost figure is an extrapolation using Statistics Canada's 2025 retail e-commerce volume and an assumed $6.40 re-delivery cost.</p></main></body></html>`;

  const outlet = `<!doctype html><html><head><title>One in seven Canadian parcels fails on the first delivery attempt, study finds</title>
<meta property="og:site_name" content="The Maple Ledger"><meta name="author" content="Dana Whitfield"><meta property="article:published_time" content="2026-02-04T15:20:00Z">
<link rel="canonical" href="http://127.0.0.1:PORT/mapleledger/parcels.html">
<script type="application/ld+json">{"@context":"https://schema.org","@type":"NewsArticle","datePublished":"2026-02-04T15:20:00Z","author":{"@type":"Person","name":"Dana Whitfield"},"publisher":{"@type":"Organization","name":"The Maple Ledger"}}</script></head>
<body><header><nav><a href="/">Home</a></nav></header><main><article>
<h1>One in seven Canadian parcels fails on the first delivery attempt, study finds</h1>
<p>About 14 per cent of parcel deliveries in Canada's largest cities fail on the first attempt, according to a benchmark released Tuesday by Northline Logistics, a Toronto last-mile carrier.</p>
<p>The study tracked 18,240 deliveries across 12 cities in the first quarter of 2026. Northline estimates failed first attempts cost retailers $212 million a year, a figure the company says is extrapolated from Statistics Canada e-commerce data.</p>
<p>Independent analysts cautioned that the sample covers only Northline's own network. "It's one carrier's numbers," said Marc Leblanc of the Retail Council of Canada.</p>
<p>Northline published its <a href="http://127.0.0.1:PORT/prnewswire/northline-release.html">announcement</a> on Tuesday morning.</p>
</article></main><aside><p>Related: <a href="/other">Other story</a></p></aside></body></html>`;

  const aggregator = `<!doctype html><html><head><title>Study: 14% of Canadian Deliveries Fail First Attempt</title>
<meta property="og:site_name" content="Logistics Daily Brief"><meta property="article:published_time" content="2026-02-05T09:00:00Z">
<link rel="canonical" href="http://127.0.0.1:PORT/mapleledger/parcels.html"></head>
<body><main><article><h1>Study: 14% of Canadian Deliveries Fail First Attempt</h1>
<p>About 14 per cent of parcel deliveries in Canada's largest cities fail on the first attempt, according to a benchmark released Tuesday by Northline Logistics, a Toronto last-mile carrier.</p>
<p>The study tracked 18,240 deliveries across 12 cities in the first quarter of 2026. Northline estimates failed first attempts cost retailers $212 million a year.</p>
<p><em>This article originally appeared in <a href="http://127.0.0.1:PORT/mapleledger/parcels.html">The Maple Ledger</a>.</em></p>
</article></main></body></html>`;

  const listicle = `<!doctype html><html><head><title>27 Last-Mile Delivery Statistics You Need to Know in 2026</title>
<meta property="og:site_name" content="StatsVault"></head>
<body><main><article><h1>27 Last-Mile Delivery Statistics You Need to Know in 2026</h1>
<p>Editor's top 3 last-mile delivery statistics:</p>
<ul><li>14% of deliveries in Canada fail on the first attempt.</li><li>Failed deliveries cost retailers $212 million per year.</li><li>Last-mile delivery accounts for 53% of total shipping costs.</li></ul>
<p>Sources: Logistics Daily Brief, industry reports.</p>
<p>Sponsored: Try ParcelPilot free for 30 days.</p>
</article></main></body></html>`;

  const wiki = `<!doctype html><html><head><title>Last-mile delivery - Wikipedia</title></head>
<body><div id="content"><h1>Last-mile delivery</h1>
<p>In Canada, a 2026 industry benchmark found that about 14% of urban parcel deliveries fail on the first attempt.<sup class="reference" id="cite_ref-1"><a href="#cite_note-1">[1]</a></sup> Failed deliveries are a major cost driver for retailers.<sup class="reference" id="cite_ref-2"><a href="#cite_note-2">[2]</a></sup></p>
<h2>References</h2>
<ol class="references">
<li id="cite_note-1"><span class="reference-text">Whitfield, Dana (4 February 2026). <a rel="nofollow" class="external text" href="http://127.0.0.1:PORT/mapleledger/parcels.html">"One in seven Canadian parcels fails on the first delivery attempt, study finds"</a>. The Maple Ledger.</span></li>
<li id="cite_note-2"><span class="reference-text"><a rel="nofollow" class="external text" href="http://127.0.0.1:PORT/statsvault/last-mile-stats.html">"27 Last-Mile Delivery Statistics"</a>. StatsVault.</span></li>
</ol></div></body></html>`;

  const original = `<!doctype html><html><head><title>Harbour City to close Elm Street bridge for 14 months</title>
<meta property="og:site_name" content="Harbour City Gazette"><meta name="author" content="Lena Okafor"><meta property="article:published_time" content="2026-03-02T08:10:00Z">
<link rel="canonical" href="http://127.0.0.1:PORT/original/story.html"></head>
<body><main><article><h1>Harbour City to close Elm Street bridge for 14 months</h1>
<p>The Elm Street bridge will close for 14 months starting June 1, the city's transport department said in a briefing attended by the Gazette on Monday.</p>
<p>Engineers found corrosion in four of the six support beams during a February inspection, department head Rosa Ilić told reporters.</p></article></main></body></html>`;
  const mirror = `<!doctype html><html><head><title>Harbour City to close Elm Street bridge for 14 months</title>
<meta property="og:site_name" content="Regional News Wire"><meta property="article:published_time" content="2026-03-02T10:40:00Z">
<link rel="canonical" href="http://127.0.0.1:PORT/original/story.html"></head>
<body><main><article><h1>Harbour City to close Elm Street bridge for 14 months</h1>
<p>The Elm Street bridge will close for 14 months starting June 1, the city's transport department said in a briefing attended by the Gazette on Monday.</p>
<p>Engineers found corrosion in four of the six support beams during a February inspection, department head Rosa Ilić told reporters.</p></article></main></body></html>`;

  const driftStudy = `<!doctype html><html><head><title>Employee Retention Survey 2026 — Keystone Analytics</title>
<meta name="author" content="Keystone Analytics Research"><meta property="article:published_time" content="2026-01-20T09:00:00Z">
<script type="application/ld+json">{"@context":"https://schema.org","@type":"Report","datePublished":"2026-01-20","publisher":{"@type":"Organization","name":"Keystone Analytics"}}</script></head>
<body><main><h1>Employee Retention Survey 2026</h1>
<p>Method: online survey of 412 HR managers at companies with 50 to 500 employees in the United Kingdom, fielded 6 to 17 January 2026.</p>
<p>Result: 32% of respondents (132 of 412) said at least one employee had resigned in the past year because flexible-work arrangements were withdrawn.</p></main></body></html>`;
  const driftPr = `<!doctype html><html><head><title>Nearly a Third of UK Employers Lost Staff Over Flexible-Work Rollbacks, Keystone Analytics Finds</title>
<meta property="og:site_name" content="PR Newswire"><meta property="article:published_time" content="2026-01-21T07:00:00Z"></head>
<body><main><article><h1>Nearly a Third of UK Employers Lost Staff Over Flexible-Work Rollbacks, Keystone Analytics Finds</h1>
<p>LONDON, Jan. 21, 2026 /PRNewswire/ -- Nearly a third of UK employers have lost staff after withdrawing flexible-work arrangements, according to the <a href="http://127.0.0.1:PORT/study/retention-report.html">Employee Retention Survey 2026</a> published today by Keystone Analytics.</p></article></main></body></html>`;
  const driftOutlet = `<!doctype html><html><head><title>One in three UK firms lost staff over flexible-work rollbacks</title>
<meta property="og:site_name" content="The Workplace Post"><meta name="author" content="Tom Adeyemi"><meta property="article:published_time" content="2026-01-22T11:30:00Z"></head>
<body><main><article><h1>One in three UK firms lost staff over flexible-work rollbacks</h1>
<p>One in three UK employers has lost staff after rolling back flexible working, according to a release from Keystone Analytics, an HR analytics firm.</p>
<p>The company said the finding came from a survey of HR managers. Its <a href="http://127.0.0.1:PORT/pr/retention-release.html">announcement</a> did not give the sample size.</p></article></main></body></html>`;
  const driftListicle = `<!doctype html><html><head><title>41 Employee Retention Statistics for 2026</title><meta property="og:site_name" content="StatsVault"></head>
<body><main><article><h1>41 Employee Retention Statistics for 2026</h1>
<ul><li>33% of UK employers lost staff because they rolled back flexible work (<a href="http://127.0.0.1:PORT/outlet/one-in-three.html">The Workplace Post</a>).</li><li>Replacing an employee costs 6 to 9 months of salary.</li></ul></article></main></body></html>`;

  const deadEnd = `<!doctype html><html><head><title>19 Sleep Facts That Will Keep You Up at Night</title><meta property="og:site_name" content="StatsVault"></head>
<body><main><article><h1>19 Sleep Facts That Will Keep You Up at Night</h1>
<p>The average person swallows eight spiders a year in their sleep, according to <a href="http://127.0.0.1:PORT/archive/pc-professional-1993.html">a 1993 column in PC Professional</a>.</p>
<p>Adults need seven to nine hours of sleep per night, according to the <a href="http://127.0.0.1:PORT/study/sleep-guidelines.html">National Sleep Foundation guidelines</a>.</p></article></main></body></html>`;

  const companyAbout = `<!doctype html><html><head><title>About Acme Robotics</title><meta property="og:site_name" content="Acme Robotics"></head>
<body><main><h1>About Acme Robotics</h1><p>Acme Robotics was founded in Leeds in 2011 by Priya Nair and Jonas Weber. The company builds warehouse picking robots and employs 240 people.</p>
<p>Press: <a href="http://127.0.0.1:PORT/company/press/2021-tenth-anniversary.html">Acme marks its tenth anniversary (2021)</a></p></main></body></html>`;
  const companyPress = `<!doctype html><html><head><title>Acme Robotics marks ten years</title><meta property="og:site_name" content="Acme Robotics"><meta property="article:published_time" content="2021-09-14T09:00:00Z"></head>
<body><main><article><h1>Acme Robotics marks ten years</h1><p>LEEDS, 14 September 2021 — Acme Robotics today marked ten years since its founding in September 2011.</p></article></main></body></html>`;
  const wikiAcme = `<!doctype html><html><head><title>Acme Robotics - Wikipedia</title></head>
<body><div id="content"><h1>Acme Robotics</h1>
<p>Acme Robotics is a British robotics company founded in 2012 in Leeds.<sup class="reference" id="cite_ref-1"><a href="#cite_note-1">[1]</a></sup> It builds warehouse picking robots.<sup class="reference" id="cite_ref-2"><a href="#cite_note-2">[2]</a></sup></p>
<h2>References</h2><ol class="references">
<li id="cite_note-1"><span class="reference-text"><a rel="nofollow" class="external text" href="http://127.0.0.1:PORT/listicle/top-startups.html">"Top 50 UK Robotics Startups"</a>. StatsVault.</span></li>
<li id="cite_note-2"><span class="reference-text"><a rel="nofollow" class="external text" href="http://127.0.0.1:PORT/company/about.html">"About Acme Robotics"</a>. Acme Robotics.</span></li>
</ol></div><div id="footer"><p>This page was last edited on 3 March 2024.</p></div></body></html>`;
  const listicleStartups = `<!doctype html><html><head><title>Top 50 UK Robotics Startups</title><meta property="og:site_name" content="StatsVault"></head>
<body><main><article><h1>Top 50 UK Robotics Startups</h1><ol><li>Acme Robotics (Leeds, founded 2012) — warehouse picking robots.</li><li>Borealis Drones (Aberdeen, founded 2015).</li></ol></article></main></body></html>`;

  const wikiFestival = `<!doctype html><html><head><title>Bramble Festival - Wikipedia</title></head>
<body><div id="content"><h1>Bramble Festival</h1>
<p>The Bramble Festival is an annual music festival in Dorset, England. The 2018 edition was attended by 40,000 people.<sup class="reference" id="cite_ref-1"><a href="#cite_note-1">[1]</a></sup></p>
<h2>References</h2><ol class="references">
<li id="cite_note-1"><span class="reference-text">Hale, Ruth (10 June 2021). <a rel="nofollow" class="external text" href="http://127.0.0.1:PORT/gazette/bramble-festival.html">"Bramble Festival returns after two-year gap"</a>. Dorset Gazette.</span></li>
</ol></div><div id="footer"><p>This page was last edited on 14 February 2019.</p></div></body></html>`;
  const gazetteFestival = `<!doctype html><html><head><title>Bramble Festival returns after two-year gap</title><meta property="og:site_name" content="Dorset Gazette"><meta name="author" content="Ruth Hale"><meta property="article:published_time" content="2021-06-10T07:45:00Z"></head>
<body><main><article><h1>Bramble Festival returns after two-year gap</h1>
<p>The festival, which according to Wikipedia drew 40,000 people in 2018, returns to Combe Farm this August after two cancelled years.</p>
<p>Organiser Sam Petrie said ticket sales were "well ahead" of 2019.</p></article></main></body></html>`;

  return {
    'prnewswire/northline-release.html': pr,
    'northline/benchmark.html': study,
    'mapleledger/parcels.html': outlet,
    'logisticsdaily/study-14-percent.html': aggregator,
    'statsvault/last-mile-stats.html': listicle,
    'wiki/Last-mile_delivery': wiki,
    'original/story.html': original,
    'mirror/story.html': mirror,
    'study/retention-report.html': driftStudy,
    'pr/retention-release.html': driftPr,
    'outlet/one-in-three.html': driftOutlet,
    'listicle/retention-stats.html': driftListicle,
    'listicle/sleep-stats.html': deadEnd,
    'company/about.html': companyAbout,
    'company/press/2021-tenth-anniversary.html': companyPress,
    'wiki/Acme_Robotics': wikiAcme,
    'listicle/top-startups.html': listicleStartups,
    'wiki/Bramble_Festival': wikiFestival,
    'gazette/bramble-festival.html': gazetteFestival,
  };
}

async function cmdDemoFixture(args) {
  const dir = args._[1] || 'citation-foodchain-fixture';
  const port = args.port || 0;
  const pages = fixturePages();
  const server = createServer((req, res) => {
    const key = decodeURIComponent((req.url || '/').replace(/^\//, '').split('?')[0]);
    if (key === '' || key === 'index.html') {
      res.writeHead(200, { 'content-type': 'text/html; charset=utf-8' });
      return res.end(`<html><body><h1>citation-foodchain fixture</h1><ul>${Object.keys(pages).map((k) => `<li><a href="/${k}">${k}</a></li>`).join('')}</ul></body></html>`);
    }
    const page = pages[key];
    if (!page) { res.writeHead(404, { 'content-type': 'text/plain' }); return res.end('not found'); }
    res.writeHead(200, { 'content-type': 'text/html; charset=utf-8' });
    res.end(page.replace(/PORT/g, String(server.address().port)));
  });
  await new Promise((r) => server.listen(port, '127.0.0.1', r));
  const actual = server.address().port;
  mkdirSync(dir, { recursive: true });
  for (const [k, v] of Object.entries(pages)) {
    const f = join(dir, k);
    mkdirSync(dirname(f), { recursive: true });
    writeFileSync(f, v.replace(/PORT/g, String(actual)));
  }
  const base = `http://127.0.0.1:${actual}/`;
  emit({ base_url: base, index: base, pages: Object.keys(pages).map((k) => base + k), note: args.serve ? 'serving until killed; GET / lists every page' : 'server closes after emit unless --serve' }, undefined);
  if (!args.serve) server.close();
  else await new Promise(() => {});
}

// ---------------------------------------------------------------------------
// main
// ---------------------------------------------------------------------------

async function main() {
  const args = parseArgs(process.argv.slice(2));
  const cmd = args._[0];
  try {
    if (cmd === 'fetch') await cmdFetch(args);
    else if (cmd === 'unwrap') await cmdUnwrap(args);
    else if (cmd === 'wayback') await cmdWayback(args);
    else if (cmd === 'reuse') await cmdReuse(args);
    else if (cmd === 'share') await cmdShare(args);
    else if (cmd === 'batch') await cmdBatch(args);
    else if (cmd === 'demo-fixture') await cmdDemoFixture(args);
    else {
      process.stderr.write(`citation-foodchain trace.mjs v${VERSION}\n\nusage:\n  node trace.mjs share <share-url> [--browser] [--html saved.html] [--out file.json]   # ChatGPT / Perplexity share links → answer + cited URLs\n  node trace.mjs fetch <url> [--claim "text"]... [--browser] [--brief] [--no-wayback] [--out file.json]\n  node trace.mjs unwrap <url> [--follow]\n  node trace.mjs wayback <url>\n  node trace.mjs reuse <a.json|url> <b.json|url> [--browser]\n  node trace.mjs batch <urls.txt> [--out dir] [--claim ...]\n  node trace.mjs demo-fixture <dir> [--serve] [--port N]\n`);
      process.exit(cmd ? 2 : 0);
    }
  } catch (e) {
    process.stderr.write(`error: ${e?.message || e}\n`);
    process.exit(1);
  }
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) main();
