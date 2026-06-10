// Playwright validation + screenshotting for Newsjack eval figures.
//
// Renders an eval-graphics HTML file in headless Chromium, asserts the design
// system rendered correctly (no JS errors, fonts applied, accent present,
// data-driven figures populated, every figure has real geometry), and writes
// PNG screenshots — full page plus one per figure block — for embedding in a
// published study.
//
// Usage:
//   node validate.mjs [path/to/figure.html] [--out DIR] [--width N]
// Defaults to ../chart-room.html, screenshots to ../screenshots/, width 1280.
//
// Exit code 0 = all checks passed; 1 = at least one check failed.

import { chromium } from 'playwright';
import { fileURLToPath, pathToFileURL } from 'node:url';
import { dirname, resolve, basename, join } from 'node:path';
import { mkdirSync } from 'node:fs';

const __dirname = dirname(fileURLToPath(import.meta.url));
const ACCENT = '#e05a47';
const ACCENT_RGB = 'rgb(224, 90, 71)';

function parseArgs(argv) {
  const a = { file: null, out: null, width: 1280 };
  for (let i = 0; i < argv.length; i++) {
    const v = argv[i];
    if (v === '--out') a.out = argv[++i];
    else if (v === '--width') a.width = parseInt(argv[++i], 10);
    else if (!v.startsWith('--')) a.file = v;
  }
  return a;
}

const args = parseArgs(process.argv.slice(2));
const file = resolve(args.file || join(__dirname, '..', 'chart-room.html'));
const outDir = resolve(args.out || join(__dirname, '..', 'screenshots'));
mkdirSync(outDir, { recursive: true });

const checks = [];
const ok = (name, cond, detail = '') => checks.push({ name, pass: !!cond, detail });

const browser = await chromium.launch();
const page = await browser.newPage({
  viewport: { width: args.width, height: 900 },
  deviceScaleFactor: 2,
});

const jsErrors = [];
page.on('pageerror', (e) => jsErrors.push(String(e)));
page.on('console', (m) => { if (m.type() === 'error') jsErrors.push(m.text()); });

await page.goto(pathToFileURL(file).href, { waitUntil: 'networkidle' });
// fonts + data-driven heatmap script have run by networkidle, but be explicit:
await page.evaluate(() => document.fonts && document.fonts.ready);
await page.waitForTimeout(250);

// ---- assertions -------------------------------------------------------------

ok('no JS/console errors', jsErrors.length === 0, jsErrors.join(' | '));

// Heatmap is data-driven — only assert it populated if the figure has one.
const hasHeatmap = await page.locator('#heatmap').count();
if (hasHeatmap) {
  const heatCells = await page.locator('#heatmap rect.hm-cell').count();
  ok('heatmap cells rendered', heatCells > 0, `${heatCells} cells`);
}

// Accent token resolved and actually painted on a primary bar.
const barFill = await page
  .locator('.bar-primary')
  .first()
  .evaluate((el) => getComputedStyle(el).fill)
  .catch(() => '');
ok('accent painted on primary series', barFill.replace(/\s/g, '') === ACCENT_RGB.replace(/\s/g, ''),
   `bar-primary fill = ${barFill}`);

// Editorial serif applied to figure titles (font loaded, not a fallback only).
const h3Font = await page
  .locator('.fig h3, h3.nj')
  .first()
  .evaluate((el) => getComputedStyle(el).fontFamily)
  .catch(() => '');
ok('serif display font applied to titles', /newsreader/i.test(h3Font), h3Font);

// Every figure block has real geometry (nothing collapsed / empty).
const figBoxes = await page.locator('.fig').evaluateAll((els) =>
  els.map((el) => { const r = el.getBoundingClientRect(); return { w: r.width, h: r.height }; })
);
ok('figure blocks present', figBoxes.length >= 1, `${figBoxes.length} .fig blocks`);
ok('all figures have geometry', figBoxes.every((b) => b.w > 50 && b.h > 50),
   JSON.stringify(figBoxes.filter((b) => !(b.w > 50 && b.h > 50))));

// Headline stat callouts present.
const stats = await page.locator('.stat').count();
ok('headline stat callouts present', stats >= 1, `${stats} stats`);

// Every chart SVG has drawable content.
const emptySvgs = await page.locator('svg.chart').evaluateAll((els) =>
  els.filter((el) => el.querySelectorAll('rect,circle,line,polyline,polygon,path,text').length === 0).length
);
ok('no empty chart SVGs', emptySvgs === 0, `${emptySvgs} empty`);

// ---- screenshots ------------------------------------------------------------

const base = basename(file).replace(/\.html?$/i, '');
await page.screenshot({ path: join(outDir, `${base}--full.png`), fullPage: true });

// per-figure crops (great for dropping a single chart into a post)
const figs = page.locator('.fig');
const n = await figs.count();
for (let i = 0; i < n; i++) {
  const tag = await figs.nth(i).locator('.fig-tag').textContent().catch(() => null);
  const label = (tag || `fig-${i + 1}`).toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-|-$/g, '');
  await figs.nth(i).screenshot({ path: join(outDir, `${base}--${label}.png`) }).catch(() => {});
}
// stat grid as one strip
if (stats > 0) {
  await page.locator('.stat-grid').first()
    .screenshot({ path: join(outDir, `${base}--headline-stats.png`) }).catch(() => {});
}

await browser.close();

// ---- report -----------------------------------------------------------------

const failed = checks.filter((c) => !c.pass);
console.log(`\nValidating ${file}`);
for (const c of checks) {
  console.log(`  ${c.pass ? 'PASS' : 'FAIL'}  ${c.name}${c.detail ? `  — ${c.detail}` : ''}`);
}
console.log(`\nScreenshots → ${outDir}`);
console.log(`${checks.length - failed.length}/${checks.length} checks passed.`);
if (failed.length) {
  console.error(`\n${failed.length} check(s) FAILED.`);
  process.exit(1);
}
console.log('All checks passed.');
