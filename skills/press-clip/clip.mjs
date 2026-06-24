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
//   --preview <file.png>   Also write a full-page screenshot to verify.
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
// Every clip carries the outlet logo: it is the key trust signal. When the article
// page has no masthead logo, the script automatically opens the home page to find
// one; --logo lets the caller supply it explicitly. The console line reports which
// source the logo came from so a reviewer can confirm a real logo (not a text fallback).
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
const { url, out, section, preview, keep, drop, root, logo: logoArg } = args;
if (!url || !out) { console.error('need --url and --out'); process.exit(1); }

const CHROME = args.chrome
  || process.env.PRESS_CLIP_CHROME
  || '/Applications/Google Chrome.app/Contents/MacOS/Google Chrome';

// Ad / recirculation / comment / video-widget networks. Blocking them at the request level stops
// the whole class of lazy-injected junk (display ads, sponsored rails, "around the web", comment
// embeds, autoplay video) from ever loading — generically, by domain, with no per-site selectors.
// The article's own text and images are first-party and load normally.
const BLOCK_HOSTS = /(doubleclick|googlesyndication|googletagservices|google-analytics|googletagmanager|adservice\.google|amazon-adsystem|adsystem|taboola|outbrain|zergnet|connatix|spot\.im|openweb|disqus|criteo|pubmatic|rubiconproject|adnxs|moatads|scorecardresearch|zemanta|sharethrough|teads|indexww|casalemedia|3lift|districtm|smartadserver|yieldmo|sailthru|piano\.io|permutive|chartbeat|parsely|nativo|bidswitch|adlightning|confiant)\./i;

// Find the masthead logo (the key trust signal) on whatever page is loaded — the
// article page first, the home page as a fallback. Prefers the homepage-linking logo
// near the top; supports inline <svg> wordmarks as well as <img> logos; excludes
// article thumbnails/icons. Returns {logoSrc, logoSvg}; never a favicon (see fallback).
function detectMastheadLogo(pg) {
  return pg.evaluate(() => {
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
    return { logoSrc, logoSvg };
  });
}

// Last-resort logo: og:logo or the site icon. Lower fidelity than a masthead, but
// still the outlet's own mark — better than a text wordmark. Returns an absolute URL.
function detectLogoFallback(pg) {
  return pg.evaluate(() => {
    const meta = (sel, attr = 'content') => { const e = document.querySelector(sel); if (!e) return ''; const v = (e.getAttribute(attr) || '').trim(); return v ? new URL(v, location.href).href : ''; };
    return meta('meta[property="og:logo"]') || meta('link[rel*="icon"]', 'href');
  });
}

(async () => {
  const browser = await chromium.launch({ executablePath: CHROME, headless: true });
  try {
  const page = await browser.newPage({ viewport: { width: 1180, height: 1600 }, deviceScaleFactor: 2 });
  await page.route('**/*', route => {
    try { return BLOCK_HOSTS.test(new URL(route.request().url()).hostname) ? route.abort() : route.continue(); }
    catch { return route.continue(); }
  });
  await page.goto(url, { waitUntil: 'domcontentloaded', timeout: 60000 });
  await page.waitForTimeout(3500);
  // nudge lazy-loaded images into view, then return to top
  await page.evaluate(async () => {
    for (let y = 0; y < document.body.scrollHeight; y += 800) { window.scrollTo(0, y); await new Promise(r => setTimeout(r, 80)); }
    window.scrollTo(0, 0);
  });
  await page.waitForTimeout(1500);

  // --- resolve the outlet logo BEFORE surgery: every clip must carry one (it is the
  // key trust signal). Priority: explicit --logo > article masthead > HOME-PAGE masthead
  // > og:logo/favicon > text wordmark. The home-page step exists because some article
  // templates render no masthead logo at all; we navigate to the home page to grab one. ---
  let logoSrc = '', logoSvg = '', logoFrom = '';
  if (logoArg) { logoSrc = logoArg; logoFrom = 'explicit (--logo)'; }
  else { ({ logoSrc, logoSvg } = await detectMastheadLogo(page)); if (logoSrc || logoSvg) logoFrom = 'article page'; }
  if (!logoSrc && !logoSvg) {
    try {
      const origin = new URL(url).origin;
      const home = await browser.newPage({ viewport: { width: 1180, height: 1600 }, deviceScaleFactor: 2 });
      await home.route('**/*', route => {
        try { return BLOCK_HOSTS.test(new URL(route.request().url()).hostname) ? route.abort() : route.continue(); }
        catch { return route.continue(); }
      });
      await home.goto(origin, { waitUntil: 'domcontentloaded', timeout: 60000 });
      await home.waitForTimeout(2000);
      ({ logoSrc, logoSvg } = await detectMastheadLogo(home));
      if (logoSrc || logoSvg) logoFrom = 'home page';
      await home.close();
    } catch { /* home-page fetch is best-effort */ }
  }
  if (!logoSrc && !logoSvg) { logoSrc = await detectLogoFallback(page); if (logoSrc) logoFrom = 'og:logo/favicon fallback'; }
  if (!logoSrc && !logoSvg) logoFrom = 'TEXT WORDMARK — no logo found';

  const stripped = await page.evaluate(({ section, keep, drop, root, logoSrc, logoSvg }) => {
    // --- capture the outlet name BEFORE we remove anything (used only as logo alt / text fallback) ---
    const meta = (sel, attr = 'content') => { const e = document.querySelector(sel); return e ? (e.getAttribute(attr) || '').trim() : ''; };
    const outlet = meta('meta[property="og:site_name"]')
      || document.title.replace(/.*[-|–—]\s*/, '').trim()
      || location.hostname.replace(/^www\./, '');
    // The outlet logo (logoSrc / logoSvg, the key trust signal) was resolved before this
    // surgery ran — from the article masthead, the home page, or an explicit --logo — so the
    // header below always has a real logo to stamp, even on templates that render none.

    // Choose the article container STRUCTURALLY, never by first-match — first-match let a
    // 115-char footer recirc card classed "post-content" beat the 2,257-char real body. Gather
    // every plausible candidate, drop any that lives inside site chrome (header/nav/footer/aside
    // is by definition not the story — pure structure, no class guessing), then pick the TIGHTEST
    // container that still holds the headline and most of the page text. That beats both a small
    // teaser (fails the text bar) and a loose <main> (drags nav/comments back in). All generic.
    const articleRoot = (() => {
      // explicit override wins — the --root escape hatch, same runtime-tailoring rule as --drop/--keep
      if (root) { const el = document.querySelector(root); if (el) return el; }
      // include WordPress post wrappers ([id^=post-], class token "post") — a huge slice of the web
      // whose ideal root (the .post div holding headline + byline + body) none of the others match.
      const SEL = ['article', '[class*="article-body" i]', '[class*="article-content" i]',
        '[class*="post-content" i]', '[class*="entry-content" i]', '[id^="post-" i]', '[class~="post"]', 'main'];
      let cands = [];
      SEL.forEach(s => { try { document.querySelectorAll(s).forEach(el => cands.push(el)); } catch {} });
      cands = [...new Set(cands)].filter(el => !el.closest('header, nav, footer, aside'));
      if (!cands.length) return document.body;
      const txt = el => (el.innerText || '').trim().length;
      const maxText = Math.max(...cands.map(txt));
      const h1 = document.querySelector('h1');
      const solid = cands.filter(el => txt(el) >= 0.6 * maxText && (!h1 || el.contains(h1)));
      if (solid.length) return solid.sort((a, b) => txt(a) - txt(b))[0];   // tightest that qualifies
      return cands.sort((a, b) => txt(b) - txt(a))[0];                     // fallback: most text
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

    // Sweep visual dead space left by request-level blocking: a blocked ad slot or emptied embed
    // (e.g. a removed tweet <blockquote>) becomes an empty box that still takes height. Remove
    // elements INSIDE the article that render with size but carry no text, no media, and no caption.
    // This is generic — it keys on "renders empty", not on any publisher's class names. A genuinely
    // failed first-party image still has its <img>/<figure> in the DOM, so it counts as media and is
    // kept; we also log everything stripped so a real photo can never vanish silently.
    const stripped = [];
    const MEDIA = 'img, picture, video, svg, iframe, embed, object, figcaption, [class*="caption" i]';
    [...articleRoot.querySelectorAll('div, aside, section, blockquote, ins, figure')].forEach(el => {
      if (!el.isConnected || isProtected(el)) return;        // already gone with an ancestor, or kept
      const r = el.getBoundingClientRect();
      if (r.height < 8 || r.width < 8) return;               // not occupying visible space
      if ((el.innerText || '').trim().length || el.querySelector(MEDIA)) return;  // has real content
      const cls = (typeof el.className === 'string' && el.className.trim()) ? '.' + el.className.trim().split(/\s+/).join('.') : '';
      stripped.push(el.tagName.toLowerCase() + cls + ' [' + Math.round(r.width) + '×' + Math.round(r.height) + ']');
      el.remove();
    });

    // --- clip header: the outlet LOGO only (the key trust signal), shown large above the article ---
    const esc = s => (s || '').replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/"/g, '&quot;');
    const LOGO_H = 64;
    const sizer = document.createElement('style');
    sizer.textContent = '.pc-logo svg{height:' + LOGO_H + 'px !important;width:auto !important}';
    document.head && document.head.appendChild(sizer);
    // header logos are often white (for a dark masthead); recolor white fills so they show on the clip
    const logoSvgDark = logoSvg
      .replace(/fill\s*=\s*"(#fff(fff)?|#ffffff|white)"/gi, 'fill="#111"')
      .replace(/fill\s*:\s*(#fff(fff)?|#ffffff|white)/gi, 'fill:#111');
    const logoImg = logoSvg
      ? '<span class="pc-logo" style="display:inline-block;height:' + LOGO_H + 'px;line-height:0;color:#111">' + logoSvgDark + '</span>'
      : logoSrc
      ? '<img src="' + esc(logoSrc) + '" alt="' + esc(outlet) + '" style="height:' + LOGO_H + 'px;max-width:420px;width:auto;object-fit:contain;display:block">'
      : '<span style="font:700 30px/1 Georgia,serif;color:#111">' + esc(outlet) + '</span>';
    const header = document.createElement('div');
    header.style.cssText = 'background:#fff;padding:4px 4px 16px;margin:0 0 14px;display:flex;justify-content:center;align-items:center';
    header.innerHTML = logoImg;
    document.body.prepend(header);
    return stripped;
  }, { section, keep, drop, root, logoSrc, logoSvg });

  // Force a white page background so Chromium doesn't paint the below-content area of the final
  // A4 page with a site's off-white/grey body color (a common trailing-band artifact). Injected
  // last so it wins the cascade even against the site's own !important html/body background.
  await page.addStyleTag({ content: 'html,body{background:#fff !important}@media print{html,body{background:#fff !important}}' });

  await page.waitForTimeout(600);
  if (preview) await page.screenshot({ path: preview, fullPage: true });
  await page.pdf({ path: out, format: 'A4', printBackground: true, margin: { top: '10mm', bottom: '12mm', left: '8mm', right: '8mm' } });
  console.log('clip written:', out, '| logo:', logoFrom);
  if (stripped && stripped.length) console.log('swept ' + stripped.length + ' empty placeholder(s):\n  - ' + stripped.join('\n  - '));
  if (logoFrom.startsWith('TEXT')) console.warn('WARNING: no outlet logo found on the article OR home page — clip fell back to a text wordmark. Find a logo and re-run with --logo "<url>".');
  } finally {
    await browser.close();
  }
})().catch(e => { console.error('ERR', e.message); process.exit(1); });
