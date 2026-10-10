#!/usr/bin/env node
// Writes the synthetic press-clip captures and clip.json sidecars for the coverage-reel
// fixture, so the whole reel renders offline. No publisher material: every outlet, headline,
// byline and sentence below is made up. PNGs are encoded here with node:zlib and drawn with a
// tiny 5x7 pixel font, the same shapes press-clip's --assets mode writes.
//
//   node fixtures/coverage-reel/make-fixture.mjs [output-dir]   (default: this folder's clips/)

import { mkdirSync, writeFileSync } from "node:fs";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { deflateSync } from "node:zlib";

import { mergeLineRects, toAssetRects, unionRect } from "../../skills/press-clip/clip.mjs";

const HERE = dirname(fileURLToPath(import.meta.url));
const SCALE = 2;
const COLUMN = { x: 0, w: 560 };

// 5x7 glyphs, one row per number, high bit on the left.
const GLYPHS = {
  A: [14, 17, 17, 31, 17, 17, 17], B: [30, 17, 17, 30, 17, 17, 30], C: [14, 17, 16, 16, 16, 17, 14], D: [30, 17, 17, 17, 17, 17, 30],
  E: [31, 16, 16, 30, 16, 16, 31], F: [31, 16, 16, 30, 16, 16, 16], G: [14, 17, 16, 23, 17, 17, 15], H: [17, 17, 17, 31, 17, 17, 17],
  I: [14, 4, 4, 4, 4, 4, 14], J: [7, 2, 2, 2, 2, 18, 12], K: [17, 18, 20, 24, 20, 18, 17], L: [16, 16, 16, 16, 16, 16, 31],
  M: [17, 27, 21, 21, 17, 17, 17], N: [17, 17, 25, 21, 19, 17, 17], O: [14, 17, 17, 17, 17, 17, 14], P: [30, 17, 17, 30, 16, 16, 16],
  Q: [14, 17, 17, 17, 21, 18, 13], R: [30, 17, 17, 30, 20, 18, 17], S: [15, 16, 16, 14, 1, 1, 30], T: [31, 4, 4, 4, 4, 4, 4],
  U: [17, 17, 17, 17, 17, 17, 14], V: [17, 17, 17, 17, 17, 10, 4], W: [17, 17, 17, 21, 21, 21, 10], X: [17, 17, 10, 4, 10, 17, 17],
  Y: [17, 17, 10, 4, 4, 4, 4], Z: [31, 1, 2, 4, 8, 16, 31], 0: [14, 17, 19, 21, 25, 17, 14], 1: [4, 12, 4, 4, 4, 4, 14],
  2: [14, 17, 1, 2, 4, 8, 31], 3: [31, 2, 4, 2, 1, 17, 14], 4: [2, 6, 10, 18, 31, 2, 2], 5: [31, 16, 30, 1, 1, 17, 14],
  6: [6, 8, 16, 30, 17, 17, 14], 7: [31, 1, 2, 4, 8, 8, 8], 8: [14, 17, 17, 14, 17, 17, 14], 9: [14, 17, 17, 15, 1, 2, 12],
  " ": [0, 0, 0, 0, 0, 0, 0], ".": [0, 0, 0, 0, 0, 12, 12], ",": [0, 0, 0, 0, 12, 4, 8], "'": [4, 4, 8, 0, 0, 0, 0],
  "-": [0, 0, 0, 31, 0, 0, 0], ":": [0, 12, 12, 0, 12, 12, 0], "/": [1, 1, 2, 4, 8, 16, 16], "&": [12, 18, 20, 8, 21, 18, 13],
};

const ARTICLES = [
  {
    slug: "northstar-daily",
    url: "https://example.test/northstar-launch",
    outlet: "NORTHSTAR DAILY", plate: [16, 42, 67], ink: [16, 42, 67], accent: [239, 106, 71],
    kicker: "INNOVATION", headline: "ASTER LABS OPENS A SAFER ROBOTICS WORKSPACE", byline: "BY JAMIE RIVERA - SEPT 14, 2026",
    meta: { outlet: "Northstar Daily", headline: "Aster Labs opens a safer robotics workspace", byline: "Jamie Rivera", published_at: "2026-09-14T09:00:00Z" },
    paragraphs: [
      "THE LAUNCH CAME AFTER TWO YEARS OF QUIET TESTING IN A SMALL WAREHOUSE NEAR THE HARBOUR.",
      "TEAMS THERE TRACKED EVERY NEAR MISS AND PUBLISHED THE NUMBERS EACH MONTH.",
      "ASTER LABS SAYS ITS NEW WORKSPACE KEEPS PEOPLE AND ROBOTS IN SEPARATE LANES BY DEFAULT.",
      "OTHER FIRMS ARE WATCHING THE PILOT CLOSELY BEFORE THEY COMMIT.",
      "A SECOND SITE IS PLANNED FOR NEXT YEAR IF THE FIRST ONE HOLDS UP.",
    ],
    mention: { paragraph: 2, sentence: "ASTER LABS SAYS ITS NEW WORKSPACE KEEPS PEOPLE AND ROBOTS IN SEPARATE LANES BY DEFAULT.", term: "ASTER LABS" },
  },
  {
    slug: "field-notes",
    url: "https://example.test/field-notes",
    outlet: "FIELD / NOTES", plate: null, ink: [31, 61, 47], accent: [208, 168, 92],
    kicker: "THE WORK ISSUE", headline: "INSIDE THE SMALL TEAMS TESTING WAREHOUSE ROBOTS", byline: "BY ROBIN CHEN - SEPT 16, 2026",
    meta: { outlet: "Field Notes", headline: "Inside the small teams testing warehouse robots", byline: "Robin Chen", published_at: "2026-09-16T08:00:00Z" },
    paragraphs: [
      "FOUR TEAMS LET US WATCH THEIR ROBOT TRIALS OVER A LONG SUMMER.",
      "EACH ONE TOOK A DIFFERENT VIEW OF HOW MUCH TO AUTOMATE.",
      "AT ASTER LABS, THE FOCUS IS ON A CLEAR SAFETY LOG THAT ANY WORKER CAN READ.",
      "THE NEXT ROUND OF TRIALS STARTS IN THE SPRING.",
    ],
    mention: { paragraph: 2, sentence: "AT ASTER LABS, THE FOCUS IS ON A CLEAR SAFETY LOG THAT ANY WORKER CAN READ.", term: "ASTER LABS" },
    scope: "section",
  },
  {
    slug: "morning-signal",
    url: "https://example.test/morning-signal",
    outlet: "MORNING SIGNAL", plate: [23, 37, 62], ink: [23, 37, 62], accent: [242, 184, 75],
    kicker: "LOCAL BUSINESS", headline: "LOCAL ROBOTICS TEAMS MOVE FROM PROTOTYPES TO PILOTS", byline: "BY MORGAN LEE",
    meta: { outlet: "Morning Signal", headline: "Local robotics teams move from prototypes to pilots", byline: "Morgan Lee", published_at: "2026-09-18T07:30:00Z" },
    paragraphs: [
      "THREE LOCAL COMPANIES SHOWED THEIR MACHINES ON THE MORNING SHOW.",
      "THE HOSTS ASKED WHAT CHANGES FOR PEOPLE ON THE FLOOR.",
      "ASTER LABS ANSWERED WITH A LIVE DEMO OF ITS LANE SYSTEM.",
      "VIEWERS SENT IN MORE QUESTIONS THAN THE SEGMENT COULD TAKE.",
      "THE STATION PLANS A FOLLOW-UP IN THE NEW YEAR.",
    ],
    mention: { paragraph: 2, sentence: "ASTER LABS ANSWERED WITH A LIVE DEMO OF ITS LANE SYSTEM.", term: "ASTER LABS" },
  },
];

// ---- raster + PNG ---------------------------------------------------------------------

function canvas(width, height) {
  const data = Buffer.alloc(width * height * 4, 255);
  return { width, height, data };
}

function fill(img, x, y, w, h, [r, g, b], a = 255) {
  const x0 = Math.max(0, Math.round(x)), y0 = Math.max(0, Math.round(y));
  const x1 = Math.min(img.width, Math.round(x + w)), y1 = Math.min(img.height, Math.round(y + h));
  for (let yy = y0; yy < y1; yy++) {
    for (let xx = x0; xx < x1; xx++) {
      const i = (yy * img.width + xx) * 4;
      img.data[i] = r; img.data[i + 1] = g; img.data[i + 2] = b; img.data[i + 3] = a;
    }
  }
}

// Draw text in CSS px at font pixel size `px`; returns per-character boxes (CSS px).
function drawText(img, str, x, y, px, color, advance = 6) {
  const boxes = [];
  [...str].forEach((ch, i) => {
    const glyph = GLYPHS[ch] || GLYPHS[" "];
    const gx = x + i * advance * px;
    glyph.forEach((row, ry) => {
      for (let rx = 0; rx < 5; rx++) if (row & (1 << (4 - rx))) fill(img, (gx + rx * px) * SCALE, (y + ry * px) * SCALE, px * SCALE, px * SCALE, color);
    });
    boxes.push({ x: gx, y: y - px, w: advance * px, h: 9 * px });
  });
  return boxes;
}

function wrap(text, maxChars) {
  const lines = [];
  let line = "";
  for (const word of text.split(" ")) {
    if (line && (line + " " + word).length > maxChars) { lines.push(line); line = word; } else line = line ? `${line} ${word}` : word;
  }
  if (line) lines.push(line);
  return lines;
}

const CRC = new Uint32Array(256).map((_, n) => { let c = n; for (let k = 0; k < 8; k++) c = c & 1 ? 0xedb88320 ^ (c >>> 1) : c >>> 1; return c >>> 0; });
function crc32(buf) {
  let c = 0xffffffff;
  for (const byte of buf) c = CRC[(c ^ byte) & 255] ^ (c >>> 8);
  return (c ^ 0xffffffff) >>> 0;
}

function chunk(type, body) {
  const len = Buffer.alloc(4); len.writeUInt32BE(body.length);
  const crc = Buffer.alloc(4); crc.writeUInt32BE(crc32(Buffer.concat([Buffer.from(type), body])));
  return Buffer.concat([len, Buffer.from(type), body, crc]);
}

export function encodePng(img, crop = { x: 0, y: 0, w: img.width, h: img.height }) {
  const rows = [];
  for (let y = crop.y; y < crop.y + crop.h; y++) {
    rows.push(Buffer.from([0]), img.data.subarray((y * img.width + crop.x) * 4, (y * img.width + crop.x + crop.w) * 4));
  }
  const ihdr = Buffer.alloc(13);
  ihdr.writeUInt32BE(crop.w, 0); ihdr.writeUInt32BE(crop.h, 4);
  ihdr[8] = 8; ihdr[9] = 6; ihdr[10] = 0; ihdr[11] = 0; ihdr[12] = 0;
  return Buffer.concat([
    Buffer.from([137, 80, 78, 71, 13, 10, 26, 10]),
    chunk("IHDR", ihdr),
    chunk("IDAT", deflateSync(Buffer.concat(rows), { level: 9 })),
    chunk("IEND", Buffer.alloc(0)),
  ]);
}

// ---- one synthetic article ---------------------------------------------------------------

function buildArticle(a) {
  const pageH = 1900;
  const img = canvas(COLUMN.w * SCALE, pageH * SCALE);
  const fg = a.ink, body = [40, 44, 52];
  // stamped masthead: on a plate of the header colour when the "site" had one
  const nameW = a.outlet.length * 6 * 5;
  const plate = { x: (COLUMN.w - nameW) / 2 - 22, y: 6, w: nameW + 44, h: 64 };
  if (a.plate) fill(img, plate.x * SCALE, plate.y * SCALE, plate.w * SCALE, plate.h * SCALE, a.plate);
  const logo = { x: plate.x + 22, y: 20, w: nameW, h: 35 };
  drawText(img, a.outlet, logo.x, logo.y, 5, a.plate ? [255, 255, 255] : fg);
  drawText(img, a.kicker, 24, 106, 2, a.accent);
  let y = 132;
  const headlineLines = wrap(a.headline, 22);
  headlineLines.forEach((line) => { drawText(img, line, 24, y, 4, fg); y += 38; });
  const h1 = { x: 24, y: 132, w: 512, h: y - 132 };
  drawText(img, a.byline, 24, y + 6, 2, [102, 120, 138]);
  y += 40;
  const hero = { x: 24, y, w: 512, h: 300 };
  fill(img, hero.x * SCALE, hero.y * SCALE, hero.w * SCALE, hero.h * SCALE, [217, 229, 234]);
  fill(img, (hero.x + 176) * SCALE, (hero.y + 70) * SCALE, 160 * SCALE, 160 * SCALE, a.plate || a.accent);
  fill(img, (hero.x + 226) * SCALE, (hero.y + 120) * SCALE, 60 * SCALE, 60 * SCALE, a.accent);
  y += 300 + 40;
  const paragraphs = [];
  let mention = null;
  a.paragraphs.forEach((text, index) => {
    const top = y;
    const lines = wrap(text, 42);
    const charBoxes = [];
    lines.forEach((line) => {
      charBoxes.push(...drawText(img, line, 24, y, 2, body).map((b) => ({ ...b })), null);
      y += 26;
    });
    const para = { x: 24, y: top - 2, w: 512, h: y - top };
    paragraphs.push(para);
    if (index === a.mention.paragraph) {
      // map sentence characters (one box per char, line breaks as null) back to rects
      const flat = lines.join(" ");
      const boxes = [];
      let k = 0;
      lines.forEach((line, li) => {
        for (let c = 0; c < line.length; c++) boxes.push(charBoxes[k++]);
        k++;                                    // the null after each line
        if (li < lines.length - 1) boxes.push(null);  // the joining space
      });
      const rectsFor = (from, to) => mergeLineRects(boxes.slice(from, to).filter(Boolean));
      const s0 = flat.indexOf(a.mention.sentence), t0 = flat.indexOf(a.mention.term);
      mention = {
        term: rectsFor(t0, t0 + a.mention.term.length),
        sentence: rectsFor(s0, s0 + a.mention.sentence.length),
        paragraph: para,
      };
    }
    y += 22;
  });
  return { img, pageH, plate, logo, h1, hero, paragraphs, mention };
}

function write(dir, file, png) {
  writeFileSync(join(dir, file), png);
}

function makeArticle(a, outDir) {
  const dir = join(outDir, a.slug);
  mkdirSync(dir, { recursive: true });
  const art = buildArticle(a);
  const k = SCALE;
  const asset = (role, file, rect, extra = {}) => ({ role, file, width: rect.w * k, height: rect.h * k, scale: k, rect: { x: COLUMN.x + rect.x, y: rect.y, w: rect.w, h: rect.h }, ...extra });
  const topRect = { x: 0, y: 0, w: COLUMN.w, h: art.hero.y + art.hero.h + 32 };
  const para = art.mention.paragraph;
  const stripRect = { x: 0, y: 0, w: COLUMN.w, h: Math.min(art.pageH, Math.ceil(para.y + para.h + 900)) };
  const mentionRect = { x: 0, y: Math.floor(para.y - 60), w: COLUMN.w, h: Math.ceil(para.h + 120) };
  const crop = (r) => ({ x: r.x * k, y: r.y * k, w: r.w * k, h: r.h * k });
  write(dir, "top.png", encodePng(art.img, crop(topRect)));
  write(dir, "strip-1.png", encodePng(art.img, crop(stripRect)));
  write(dir, "mention-1.png", encodePng(art.img, crop(mentionRect)));
  // the mark alone, on a transparent background
  const mark = canvas(art.logo.w * k, art.logo.h * k);
  mark.data.fill(0);
  drawTextOn(mark, a.outlet, a.plate ? [255, 255, 255] : a.ink);
  write(dir, "masthead.png", encodePng(mark));
  const assets = [
    asset("top", "top.png", topRect),
    asset("strip", "strip-1.png", stripRect, { index: 0 }),
    asset("mention", "mention-1.png", mentionRect, { mention: 0 }),
    { role: "masthead", file: "masthead.png", width: art.logo.w * k, height: art.logo.h * k, scale: k, rect: art.logo },
  ];
  const lines = { term: art.mention.term, sentence: art.mention.sentence, clause: art.mention.sentence };
  const pageAssets = assets.filter((x) => x.role !== "masthead");
  const sidecar = {
    version: 1,
    tool: "fixtures/coverage-reel/make-fixture.mjs (synthetic)",
    source_url: a.url,
    final_url: a.url,
    canonical_url: a.url,
    captured_at: "2026-09-30T12:00:00.000Z",
    viewport: { width: 1180, height: 1600 },
    scale: k,
    scope: a.scope || "whole",
    section: a.scope === "section" ? "Aster Labs" : null,
    mention_terms: ["ASTER LABS"],
    page: {
      outlet: { value: a.meta.outlet, source: "json-ld" },
      headline: { value: a.meta.headline, source: "visible" },
      byline: { value: a.meta.byline, source: "json-ld" },
      published_at: { value: a.meta.published_at, source: "json-ld" },
    },
    logo: { resolved_from: "article page", kind: "svg", frame: null, plate_color: a.plate ? `rgb(${a.plate.join(", ")})` : null, file: "masthead.png", rect: art.logo, plate_rect: a.plate ? art.plate : { x: art.logo.x - 16, y: art.logo.y - 12, w: art.logo.w + 32, h: art.logo.h + 24 } },
    root: { x: 0, y: 0, w: COLUMN.w, h: art.pageH },
    headline_rect: art.h1,
    assets,
    mentions: [{
      rank: 1,
      chosen: true,
      term: a.mention.term,
      text: a.mention.term,
      sentence: a.mention.sentence,
      clause: a.mention.sentence,
      block: "p",
      in_link: false,
      score: 7,
      reasons: ["+3 body paragraph", "+2 full product or company name", "+1 in the first five paragraphs", "+1 readable sentence length"],
      rects: { ...lines, paragraph: art.mention.paragraph },
      asset_rects: pageAssets
        .map((x) => ({ file: x.file, term: toAssetRects(lines.term, x), sentence: toAssetRects(lines.sentence, x), clause: toAssetRects(lines.clause, x) }))
        .filter((x) => x.term.length || x.sentence.length),
    }],
    warnings: [],
  };
  writeFileSync(join(dir, "clip.json"), `${JSON.stringify(sidecar, null, 2)}\n`);
  return { dir, sidecar, union: unionRect(art.mention.sentence) };
}

function drawTextOn(img, str, color) {
  // the masthead PNG holds only the wordmark: draw at the origin of its own small canvas
  [...str].forEach((ch, i) => {
    const glyph = GLYPHS[ch] || GLYPHS[" "];
    glyph.forEach((row, ry) => {
      for (let rx = 0; rx < 5; rx++) if (row & (1 << (4 - rx))) fill(img, (i * 30 + rx * 5) * SCALE, ry * 5 * SCALE, 5 * SCALE, 5 * SCALE, color);
    });
  });
}

export function makeFixture(outDir = join(HERE, "clips")) {
  mkdirSync(outDir, { recursive: true });
  return ARTICLES.map((a) => makeArticle(a, resolve(outDir)));
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const made = makeFixture(process.argv[2] ? resolve(process.argv[2]) : undefined);
  for (const m of made) console.log(`wrote ${m.dir}`);
}
