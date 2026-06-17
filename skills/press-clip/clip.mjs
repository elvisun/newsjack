#!/usr/bin/env node
// press-clip clipper — render a live article to a high-fidelity PDF press clip.
//
// Keeps the publication's real look (logo, fonts, photos, layout) and removes
// the ad/chrome clutter. Optionally isolates one section (the client's) in a
// per-item roundup, and highlights the client's mentions.
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

// Generic clutter that is never part of the article. Safe across sites; a
// selector that matches nothing is simply ignored.
const JUNK = [
  'iframe', 'ins.adsbygoogle', '.adsbygoogle', 'amp-ad',
  '[id^="div-gpt"]', '[id*="google_ads"]', '[id*="-ad-"]', '[class^="ad-"]', '[class*=" ad-"]',
  '.td-a-rec', '[class*="td-a-rec"]', '.td-a-ad',                       // tagDiv (WordPress) ad recs
  '[class*="advert"]', '[data-ad]', '[aria-label*="advertisement" i]',
  '[id*="cookie" i]', '[class*="cookie" i]', '[class*="consent" i]', '[id*="consent" i]',
  '[class*="newsletter" i]', '[class*="subscribe" i]', '[class*="signup" i]',
  '[class*="related" i]', '[class*="more-from" i]', '[class*="recirc" i]',
  '[class*="share" i]', '[class*="social" i]', '[class*="sharing" i]',
  '#comments', '.comments-area', '[class*="comment" i]',
  // overlays — target real dialogs/interstitials, NOT click-to-zoom image links (e.g. a.td-modal-image)
  '[role="dialog"]', '[aria-modal="true"]', '[class*="lightbox" i]', '[class*="-popup" i]', '[class*="popup-" i]', '[class*="paywall" i]',
  // in-article table of contents — lists sections we may have removed, so drop it on a clip
  '[class*="ez-toc" i]', '[class*="table-of-contents" i]', 'nav[class*="toc" i]', '#toc',
];

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

  await page.evaluate(({ JUNK, client, section, keep, drop }) => {
    // --- capture outlet identity BEFORE we remove anything (logo often lives in a header we strip) ---
    const meta = (sel, attr = 'content') => { const e = document.querySelector(sel); return e ? (e.getAttribute(attr) || '').trim() : ''; };
    const outlet = meta('meta[property="og:site_name"]')
      || document.title.replace(/.*[-|–—]\s*/, '').trim()
      || location.hostname.replace(/^www\./, '');
    let dateStr = '';
    const dateRaw = meta('meta[property="article:published_time"]') || meta('meta[itemprop="datePublished"]')
      || meta('time[datetime]', 'datetime');
    if (dateRaw) { const d = new Date(dateRaw); if (!isNaN(d)) dateStr = d.toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' }); }
    // find the masthead logo: a real <img> logo near the top, else a wordmark/og:logo/icon
    let logoSrc = '';
    const cands = [...document.querySelectorAll(
      '[class*="site-logo" i] img, [class*="navbar-brand" i] img, [class*="logo" i] img, #logo img, header img, a[href="/"] img, a[href="' + location.origin + '/"] img'
    )];
    for (const img of cands) {
      const s = img.currentSrc || img.src || '';
      const r = img.getBoundingClientRect();
      if (s && !/sprite|emoji|avatar|gravatar|icon-/i.test(s) && r.top < 700) { logoSrc = s; break; }
    }
    if (!logoSrc) logoSrc = meta('meta[property="og:logo"]') || meta('link[rel*="icon"]', 'href');

    const keepSel = (keep || '').split(',').map(s => s.trim()).filter(Boolean);
    const protectedEls = new Set();
    keepSel.forEach(sel => document.querySelectorAll(sel).forEach(n => protectedEls.add(n)));
    const isProtected = n => { for (let p = n; p; p = p.parentElement) if (protectedEls.has(p)) return true; return false; };
    const kill = sel => document.querySelectorAll(sel).forEach(n => { if (!isProtected(n)) n.remove(); });

    JUNK.forEach(kill);
    (drop || '').split(',').map(s => s.trim()).filter(Boolean).forEach(kill);
    // common layout chrome: sidebar, footer, sticky bars, and the site's top nav menu
    // (safe — we already captured the logo above and re-inject it in the clip header band)
    kill('[class*="sidebar" i], aside, [class*="footer" i], footer, [class*="sticky" i], [class*="affix" i], nav, [role="navigation"], [class*="menu-wrap" i], [class*="header-menu" i]');

    // widen the main column once the sidebar is gone
    document.querySelectorAll('[class*="span8"], [class*="main-content"], [class*="content-area"]')
      .forEach(n => { n.style.width = '100%'; n.style.maxWidth = '100%'; n.style.flex = '0 0 100%'; });

    // --- isolate one roundup section, if asked ---
    if (section) {
      const content = document.querySelector('[class*="post-content" i], [class*="entry-content" i], article, main') || document.body;
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
      const root = document.querySelector('[class*="post-content" i], [class*="entry-content" i], article, main') || document.body;
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
    const logoImg = logoSrc
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
  }, { JUNK, client, section, keep, drop });

  await page.waitForTimeout(600);
  if (preview) await page.screenshot({ path: preview, fullPage: true });
  await page.pdf({ path: out, format: 'A4', printBackground: true, margin: { top: '10mm', bottom: '12mm', left: '8mm', right: '8mm' } });
  await browser.close();
  console.log('clip written:', out);
})().catch(e => { console.error('ERR', e.message); process.exit(1); });
