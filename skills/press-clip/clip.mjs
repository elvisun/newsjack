#!/usr/bin/env node
// press-clip clipper — render a live article to a high-fidelity PDF press clip.
//
// Keeps the publication's real look (logo, fonts, photos, layout). Isolates the
// article STRUCTURALLY — keep the article and its ancestor chain, drop everything
// that is a sibling of that chain (header, nav, sidebar, footer, recirc) — so no
// site-specific class names are baked in. Per-publisher junk that lives INSIDE the
// article (ads, sponsored rails, newsletter boxes, comment embeds) is removed at
// runtime via --drop, which the caller supplies after inspecting the page.
// Optionally isolates one section (the client's) in a roundup, and highlights the
// client's mentions. Contains no per-site logic by design.
//
// Usage:
//   node clip.mjs --url <URL> --out <file.pdf> [options]
//
// Options:
//   --client "<Name>"      Highlight this brand/name wherever it appears.
//   --section "<Heading>"  Isolate the roundup section whose heading names this
//                          client: keep the lead + that section, drop the rest.
//                          Omit for single-subject articles (keep whole article).
//   --preview <file.png>   Also write a full-page screenshot to verify.
//   --chrome <path>        Browser executable (defaults to system Chrome/Chromium).
//   --keep "<sel,sel>"     Extra CSS selectors to force-keep (never remove).
//   --drop "<sel,sel>"     Extra CSS selectors to remove (site-specific junk).
//
// Requires a Chromium-based browser on the machine and the `playwright-core`
// npm package (npm i playwright-core). Edge works as the browser too.

// Resolve playwright-core from the user's working directory (where they ran
// `npm i playwright-core`) or the global npm root, not just next to this script.
import { createRequire } from 'module';
import { pathToFileURL } from 'url';
import { execSync } from 'child_process';

function loadChromium() {
  const bases = [process.cwd() + '/', process.env.PRESS_CLIP_MODULES ? process.env.PRESS_CLIP_MODULES + '/' : null].filter(Boolean);
  try { bases.push(execSync('npm root -g').toString().trim() + '/../'); } catch {}
  for (const base of bases) {
    try { return createRequire(pathToFileURL(base))('playwright-core').chromium; } catch {}
  }
  console.error('Could not find playwright-core. Run:  npm i playwright-core   (in this folder), then retry.');
  process.exit(1);
}
const chromium = loadChromium();

const args = {};
for (let i = 2; i < process.argv.length; i++) {
  const a = process.argv[i];
  if (a.startsWith('--')) args[a.slice(2)] = process.argv[++i];
}
const { url, out, client, section, preview, keep, drop } = args;
if (!url || !out) { console.error('need --url and --out'); process.exit(1); }

const CHROME = args.chrome
  || process.env.PRESS_CLIP_CHROME
  || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';

(async () => {
  const browser = await chromium.launch({ executablePath: CHROME, headless: true });
  const page = await browser.newPage({ viewport: { width: 1180, height: 1600 }, deviceScaleFactor: 2 });
  await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 60000 });
  await page.waitForTimeout(3500);
  // nudge lazy-loaded images into view, then return to top
  await page.evaluate(async () => {
    for (let y = 0; y < document.body.scrollHeight; y += 800) { window.scrollTo(0, y); await new Promise(r => setTimeout(r, 80)); }
    window.scrollTo(0, 0);
  });
  await page.waitForTimeout(1500);

  await page.evaluate(({ client, section, keep, drop }) => {
    // --- capture outlet identity BEFORE we remove anything (logo often lives in a header we strip) ---
    const meta = (sel, attr = 'content') => { const e = document.querySelector(sel); return e ? (e.getAttribute(attr) || '').trim() : ''; };
    const outlet = meta('meta[property="og:site_name"]')
      || document.title.replace(/.*[-|–—]\s*/, '').trim()
      || location.hostname.replace(/^www\./, '');
    let dateStr = '';
    const dateRaw = meta('meta[property="article:published_time"]') || meta('meta[itemprop="datePublished"]')
      || meta('time[datetime]', 'datetime');
    if (dateRaw) { const d = new Date(dateRaw); if (!isNaN(d)) dateStr = d.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' }); }
    // find the masthead logo (the key trust signal): prefer the homepage-linking logo near the top.
    // support inline <svg> wordmarks as well as <img> logos; exclude article thumbnails/icons.
    let logoSrc = '', logoSvg = '';
    const badImg = /sprite|emoji|avatar|gravatar|icon-|\/thumbs?\/|uploads\/sites/i;
    const originRe = new RegExp('^' + location.origin.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + '\\/?$');
    const brandCands = [
      ...[...document.querySelectorAll('a')].filter(a => {
        const href = a.getAttribute('href') || ''; const r = a.getBoundingClientRect();
        return (href === '/' || originRe.test(href)) && r.top < 300 && r.width > 60;
      }),
      ...document.querySelectorAll('[class*="site-logo" i], [class*="masthead" i], [class*="navbar-brand" i], [class*="logo" i]'),
    ];
    for (const c of brandCands) {
      const img = c.matches('img') ? c : c.querySelector('img');
      if (img) { const s = img.currentSrc || img.src || ''; if (s && !badImg.test(s)) { logoSrc = s; break; } }
      const svg = c.matches('svg') ? c : c.querySelector('svg');
      if (svg && svg.getBoundingClientRect().width >= 60) { logoSvg = svg.outerHTML; break; }
    }
    if (!logoSrc && !logoSvg) logoSrc = meta('meta[property="og:logo"]') || meta('link[rel*="icon"]', 'href');

    // The article body and the wrappers it sits inside must never be deleted, even if a
    // wrapper's class happens to contain a junk word (e.g. a "recirc" container around the article).
    // Pick by priority and by most text — not document order — so a page-level <main> wrapper or a
    // teaser-card <article> in the header strip can't win over the real story body.
    const articleRoot = (() => {
      const arts = [...document.querySelectorAll('article')];
      if (arts.length) return arts.sort((a, b) => b.innerText.length - a.innerText.length)[0];
      for (const s of ['[class*="article-body" i]', '[class*="article-content" i]', '[class*="post-content" i]', '[class*="entry-content" i]', 'main']) {
        const el = document.querySelector(s);
        if (el) return el;
      }
      return document.body;
    })();

    const keepSel = (keep || '').split(',').map(s => s.trim()).filter(Boolean);
    const protectedEls = new Set();
    keepSel.forEach(sel => document.querySelectorAll(sel).forEach(n => protectedEls.add(n)));
    const isProtected = n => { for (let p = n; p; p = p.parentElement) if (protectedEls.has(p)) return true; return false; };

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
    [...HARD_JUNK, ...(drop || '').split(',').map(s => s.trim()).filter(Boolean)].forEach(sel => {
      document.querySelectorAll(sel).forEach(n => { if (!n.contains(articleRoot) && !isProtected(n)) n.remove(); });
    });

    // widen the main column in case a removed sidebar left the article in a narrow grid track
    document.querySelectorAll('[class*="span8"], [class*="main-content"], [class*="content-area"]')
      .forEach(n => { n.style.width = '100%'; n.style.maxWidth = '100%'; n.style.flex = '0 0 100%'; });

    // --- isolate one roundup section, if asked ---
    if (section) {
      const content = articleRoot;
      const heads = [...content.querySelectorAll('h1,h2,h3,h4')];
      const norm = s => (s || '').trim().toLowerCase();
      const target = norm(section).slice(0, 8);
      const start = heads.find(h => norm(h.textContent).startsWith(target));
      if (start) {
        const level = start.tagName;                     // boundary = next heading at same level
        const kids = [...start.parentElement.children];
        const sIdx = kids.indexOf(start);
        let eIdx = kids.length;
        for (let i = sIdx + 1; i < kids.length; i++) { if (kids[i].tagName === level) { eIdx = i; break; } }
        // drop everything after our section
        for (let i = kids.length - 1; i >= eIdx; i--) kids[i].remove();
        // drop other sections before ours, but keep the article's lead (nodes before the first heading)
        const firstHeadIdx = kids.findIndex(n => /^H[1-4]$/.test(n.tagName));
        if (firstHeadIdx !== -1 && firstHeadIdx < sIdx) for (let i = sIdx - 1; i >= firstHeadIdx; i--) kids[i].remove();
      }
    }

    // --- highlight client mentions ---
    if (client) {
      const rx = new RegExp('(' + client.replace(/[.*+?^${}()|[\]\\]/g, '\\$&') + ')', 'g');
      const root = articleRoot;
      const walker = document.createTreeWalker(root, NodeFilter.SHOW_TEXT);
      const hits = [];
      while (walker.nextNode()) {
        const t = walker.currentNode;
        if (rx.test(t.nodeValue) && t.parentElement && !['SCRIPT', 'STYLE', 'MARK'].includes(t.parentElement.tagName)) hits.push(t);
      }
      hits.forEach(t => {
        const span = document.createElement('span');
        span.innerHTML = t.nodeValue.replace(rx, '<mark style="background:#fde68a;padding:0 .08em;border-radius:2px">$1</mark>');
        t.replaceWith(span);
      });
    }

    // --- clip header band: the outlet LOGO (the key trust signal) + outlet name, date, source link ---
    const esc = s => (s || '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/"/g, '&quot;');
    const sizer = document.createElement('style');
    sizer.textContent = '.pc-logo svg{height:40px !important;width:auto !important}';
    document.head && document.head.appendChild(sizer);
    // header logos are often white (for a dark masthead); recolor white fills so they show on our band
    const logoSvgDark = logoSvg
      .replace(/fill\s*=\s*"(#fff(fff)?|#ffffff|white)"/gi, 'fill="#111"')
      .replace(/fill\s*:\s*(#fff(fff)?|#ffffff|white)/gi, 'fill:#111');
    const logoImg = logoSvg
      ? '<span class="pc-logo" style="display:inline-block;height:40px;line-height:0;color:#111">' + logoSvgDark + '</span>'
      : logoSrc
      ? '<img src="' + esc(logoSrc) + '" alt="' + esc(outlet) + '" style="height:40px;max-width:260px;width:auto;object-fit:contain;display:block">'
      : '<span style="font:700 22px/1 Georgia,serif;color:#111">' + esc(outlet) + '</span>';
    const meta2 = [
      client && ('<b style="color:#0f172a">PRESS CLIP — ' + esc(client) + '</b>'),
      esc(outlet),
      dateStr && ('Published ' + esc(dateStr)),
      '<a href="' + esc(location.href) + '" style="color:#475569;text-decoration:none">' + esc(location.href) + '</a>',
    ].filter(Boolean).join(' &nbsp;·&nbsp; ');
    const band = document.createElement('div');
    band.style.cssText = 'background:#fff;border-bottom:2px solid #0f172a;padding:14px 18px 12px;margin:0 0 10px;display:flex;align-items:center;gap:16px';
    band.innerHTML =
      '<div style="flex:0 0 auto">' + logoImg + '</div>' +
      '<div style="font:500 11.5px/1.5 -apple-system,Segoe UI,Roboto,sans-serif;color:#475569;word-break:break-word;border-left:1px solid #cbd5e1;padding-left:16px">' + meta2 + '</div>';
    document.body.prepend(band);
  }, { client, section, keep, drop });

  await page.waitForTimeout(600);
  if (preview) await page.screenshot({ path: preview, fullPage: true });
  await page.pdf({ path: out, format: 'A4', printBackground: true, margin: { top: '10mm', bottom: '12mm', left: '8mm', right: '8mm' } });
  await browser.close();
  console.log('clip written:', out);
})().catch(e => { console.error('ERR', e.message); process.exit(1); });
