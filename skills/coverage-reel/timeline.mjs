// coverage-reel timeline: the motion system as pure, deterministic code.
//
// buildTimeline() turns reviewed coverage (rows + press-clip sidecars + metrics) into a
// list of scenes with exact start/end times, camera keyframes computed from the mention
// rectangles, reading-speed holds, and layout boxes for one output format. The browser
// compositor (reel-player.js) draws a frame as a pure function of time from this data;
// lintTimeline() checks the motion against the anti-pattern list before anything renders.
//
// This file runs in Node (tests, render.mjs) and in the browser: the renderer inlines it
// into reel.html. Keep it free of imports and side effects, and export only with
// `export function` / `export const` declarations.

export const TIMELINE_VERSION = 1;

// ---------------------------------------------------------------------------
// Easing palette. Every animated value names one of these; nothing else is allowed.
// ---------------------------------------------------------------------------

export const EASINGS = {
  ARRIVE: [0.05, 0.7, 0.1, 1],   // things landing: camera on the mention, cards, chips
  MOVE: [0.2, 0, 0, 1],          // on-screen moves: word reveals, push transitions, scrolls
  LEAVE: [0.3, 0, 0.8, 0.15],    // exits, at about 70% of the paired entry duration
  DRIFT: [0.37, 0, 0.63, 1],     // slow symmetric moves under holds
  MARK: [0, 0, 0.3, 1],          // highlighter and underline draws
};
export const EASE_NAMES = ["ARRIVE", "MOVE", "LEAVE", "DRIFT", "MARK", "SPRING", "LINEAR"];
// constant speed reads as mechanical except for drifts, marquees and the progress bar
export const LINEAR_ROLES = ["drift", "marquee", "progress"];

function bezier(p1x, p1y, p2x, p2y, x) {
  const cx = 3 * p1x, bx = 3 * (p2x - p1x) - cx, ax = 1 - cx - bx;
  const cy = 3 * p1y, by = 3 * (p2y - p1y) - cy, ay = 1 - cy - by;
  const sampleX = (t) => ((ax * t + bx) * t + cx) * t;
  const slopeX = (t) => (3 * ax * t + 2 * bx) * t + cx;
  let t = x;
  for (let i = 0; i < 8; i++) {
    const err = sampleX(t) - x;
    if (Math.abs(err) < 1e-7) break;
    const d = slopeX(t);
    if (Math.abs(d) < 1e-6) break;
    t -= err / d;
  }
  if (t < 0 || t > 1 || Math.abs(sampleX(t) - x) > 1e-5) {
    let lo = 0, hi = 1;
    t = x;
    for (let i = 0; i < 40; i++) {
      const v = sampleX(t);
      if (Math.abs(v - x) < 1e-7) break;
      if (v < x) lo = t; else hi = t;
      t = (lo + hi) / 2;
    }
  }
  return ((ay * t + by) * t + cy) * t;
}

// Critically damped spring (damping = 2 * sqrt(stiffness)): no oscillation, no overshoot,
// normalised so it lands exactly on 1 at the end of its duration.
const SPRING_W = 9.23;
const springRaw = (x) => 1 - (1 + SPRING_W * x) * Math.exp(-SPRING_W * x);
const SPRING_END = springRaw(1);

export function ease(name, x) {
  const p = x <= 0 ? 0 : x >= 1 ? 1 : x;
  if (name === "LINEAR") return p;
  if (name === "SPRING") return springRaw(p) / SPRING_END;
  const c = EASINGS[name];
  if (!c) throw new Error(`unknown easing ${name}`);
  if (p === 0 || p === 1) return p;
  return bezier(c[0], c[1], c[2], c[3], p);
}

// Progress of a timed move: { start, dur, ease } -> 0..1
export function progress(move, t) {
  if (!move) return 1;
  if (t <= move.start) return 0;
  if (t >= move.start + move.dur) return 1;
  return ease(move.ease, (t - move.start) / move.dur);
}

const lerp = (a, b, p) => a + (b - a) * p;

// Staggered reveal of n parts: { start, stagger, dur, ease } -> progress of part i
export function staggerProgress(spec, i, t) {
  return progress({ start: spec.start + i * (spec.stagger || 0), dur: spec.dur, ease: spec.ease }, t);
}

// ---------------------------------------------------------------------------
// Formats: semantic slots per output size. Motion is computed relative to these slots,
// so the same timeline recomposes for every format.
// ---------------------------------------------------------------------------

export const FORMATS = {
  landscape: {
    width: 1920, height: 1080,
    safe: { x: 96, y: 54, w: 1728, h: 972 },
    article: {
      card: { x: 104, y: 72, w: 920, h: 936, r: 18 },
      focus: { x: 150, y: 300, w: 828, h: 420 },
      masthead: { x: 1108, y: 128, w: 700, h: 64 },
      kicker: { x: 1108, y: 246, w: 700, h: 30, size: 22 },
      headline: { x: 1108, y: 290, w: 700, h: 380, size: 60, min: 38, lines: 5, serif: true },
      byline: { x: 1108, y: 690, w: 700, h: 40, size: 27 },
      lower: { x: 1108, y: 812, w: 700, h: 128, size: 31, small: 25 },
      micro: { x: 1108, y: 958, w: 700, h: 30, size: 20 },
    },
    opening: {
      brand: { x: 112, y: 92, w: 600, h: 60 },
      overline: { x: 112, y: 300, w: 1400, h: 34, size: 26 },
      title: { x: 112, y: 350, w: 1560, h: 260, size: 112, min: 64, lines: 2, serif: true },
      facts: { x: 112, y: 650, w: 1560, h: 46, size: 36 },
      strip: { x: 112, y: 790, w: 1696, h: 168, chip: 72, gap: 22 },
    },
    wall: { metrics: { x: 112, y: 300, w: 1696, h: 480 }, value: 150, label: 30, caption: { x: 112, y: 200, w: 1696, h: 40, size: 28 } },
    end: {
      overline: { x: 112, y: 104, w: 1696, h: 34, size: 28 },
      list: { x: 112, y: 196, w: 1696, h: 600, row: 150, size: 36, small: 26, columns: 2 },
      disclaimer: { x: 112, y: 900, w: 1250, h: 80, size: 23 },
      brand: { x: 1508, y: 896, w: 300, h: 64 },
    },
    progress: { y: 1072, h: 8 },
  },
  portrait: {
    width: 1080, height: 1350,
    safe: { x: 54, y: 54, w: 972, h: 1242 },
    article: {
      masthead: { x: 60, y: 62, w: 960, h: 50 },
      kicker: { x: 60, y: 128, w: 960, h: 26, size: 20 },
      headline: { x: 60, y: 160, w: 960, h: 150, size: 46, min: 32, lines: 3, serif: true },
      byline: { x: 60, y: 318, w: 960, h: 30, size: 22 },
      card: { x: 60, y: 366, w: 960, h: 740, r: 16 },
      focus: { x: 92, y: 520, w: 896, h: 400 },
      lower: { x: 60, y: 1124, w: 960, h: 100, size: 30, small: 24 },
      micro: { x: 60, y: 1240, w: 960, h: 30, size: 19 },
    },
    opening: {
      brand: { x: 64, y: 72, w: 500, h: 52 },
      overline: { x: 64, y: 300, w: 950, h: 30, size: 24 },
      title: { x: 64, y: 346, w: 950, h: 300, size: 90, min: 56, lines: 3, serif: true },
      facts: { x: 64, y: 680, w: 950, h: 80, size: 32 },
      strip: { x: 64, y: 820, w: 952, h: 360, chip: 64, gap: 20 },
    },
    wall: { metrics: { x: 64, y: 330, w: 952, h: 640 }, value: 132, label: 28, caption: { x: 64, y: 230, w: 952, h: 40, size: 26 } },
    end: {
      overline: { x: 64, y: 72, w: 952, h: 30, size: 24 },
      list: { x: 64, y: 130, w: 952, h: 900, row: 118, size: 30, small: 23, columns: 1 },
      disclaimer: { x: 64, y: 1100, w: 952, h: 80, size: 21 },
      brand: { x: 64, y: 1210, w: 400, h: 56 },
    },
    progress: { y: 1342, h: 8 },
  },
  vertical: {
    width: 1080, height: 1920,
    // keep key text out of the top 14%, the bottom 35% and 6% at the sides (platform UI)
    safe: { x: 65, y: 269, w: 950, h: 979 },
    article: {
      masthead: { x: 65, y: 284, w: 950, h: 54 },
      kicker: { x: 65, y: 356, w: 950, h: 26, size: 21 },
      headline: { x: 65, y: 388, w: 950, h: 170, size: 50, min: 34, lines: 3, serif: true },
      byline: { x: 65, y: 566, w: 950, h: 30, size: 23 },
      lower: { x: 65, y: 612, w: 700, h: 92, size: 29, small: 23 },
      micro: { x: 765, y: 616, w: 250, h: 30, size: 19, align: "right" },
      card: { x: 36, y: 720, w: 1008, h: 1060, r: 18 },
      focus: { x: 80, y: 790, w: 920, h: 400 },
    },
    opening: {
      brand: { x: 65, y: 284, w: 500, h: 54 },
      overline: { x: 65, y: 470, w: 950, h: 30, size: 24 },
      title: { x: 65, y: 514, w: 950, h: 330, size: 96, min: 56, lines: 3, serif: true },
      facts: { x: 65, y: 870, w: 950, h: 80, size: 32 },
      strip: { x: 65, y: 990, w: 950, h: 250, chip: 60, gap: 18 },
    },
    wall: { metrics: { x: 65, y: 420, w: 950, h: 760 }, value: 136, label: 28, caption: { x: 65, y: 300, w: 950, h: 40, size: 26 } },
    end: {
      overline: { x: 65, y: 290, w: 950, h: 30, size: 24 },
      list: { x: 65, y: 340, w: 950, h: 720, row: 100, size: 28, small: 22, columns: 1 },
      disclaimer: { x: 65, y: 1080, w: 950, h: 90, size: 21 },
      brand: { x: 65, y: 1184, w: 400, h: 50 },
    },
    progress: { y: 1912, h: 8 },
  },
  square: {
    width: 1080, height: 1080,
    safe: { x: 54, y: 40, w: 972, h: 1000 },
    article: {
      masthead: { x: 60, y: 44, w: 960, h: 44 },
      kicker: { x: 60, y: 102, w: 960, h: 24, size: 19 },
      headline: { x: 60, y: 130, w: 960, h: 100, size: 40, min: 28, lines: 2, serif: true },
      byline: { x: 60, y: 236, w: 960, h: 28, size: 20 },
      card: { x: 60, y: 278, w: 960, h: 620, r: 16 },
      focus: { x: 92, y: 400, w: 896, h: 360 },
      lower: { x: 60, y: 912, w: 960, h: 92, size: 28, small: 22 },
      micro: { x: 60, y: 1008, w: 960, h: 28, size: 18 },
    },
    opening: {
      brand: { x: 64, y: 64, w: 500, h: 50 },
      overline: { x: 64, y: 250, w: 950, h: 30, size: 23 },
      title: { x: 64, y: 292, w: 950, h: 240, size: 84, min: 52, lines: 3, serif: true },
      facts: { x: 64, y: 560, w: 950, h: 76, size: 30 },
      strip: { x: 64, y: 690, w: 952, h: 300, chip: 56, gap: 18 },
    },
    wall: { metrics: { x: 64, y: 250, w: 952, h: 600 }, value: 120, label: 26, caption: { x: 64, y: 160, w: 952, h: 36, size: 24 } },
    end: {
      overline: { x: 64, y: 60, w: 952, h: 30, size: 23 },
      list: { x: 64, y: 112, w: 952, h: 740, row: 104, size: 27, small: 21, columns: 1 },
      disclaimer: { x: 64, y: 870, w: 952, h: 80, size: 20 },
      brand: { x: 64, y: 966, w: 400, h: 50 },
    },
    progress: { y: 1072, h: 8 },
  },
};

export const FORMAT_NAMES = Object.keys(FORMATS);

// ---------------------------------------------------------------------------
// Timing constants (seconds), from the motion research. Holds come from reading speed.
// ---------------------------------------------------------------------------

export const TIMING = {
  push: 0.45,             // push transition between scenes (400-550 ms)
  dissolve: 0.8,          // the one cross-dissolve, into the end card
  chipIn: 0.5,            // masthead chips and lower-thirds (400-600 ms)
  wordReveal: 0.5,        // per word (450-600 ms)
  wordStagger: 0.05,      // between words (40-70 ms)
  lineStagger: 0.07,
  zoom: 1.0,              // zoom to mention (0.9-1.3 s)
  zoomSettle: 0.35,       // share of the ARRIVE zoom after which the sentence reads as landed (~85% of the move)
  dim: 0.65,              // dim the surround (500-700 ms)
  markDelay: 0.2,         // land, then mark (150-250 ms)
  markGap: 0.06,
  driftRate: 0.011,       // page drift, scale per second (1-1.5%/s)
  readCps: 20,            // reading speed for holds: the 20 cps ceiling (Netflix adult subtitles)
  maxCps: 20,             // lint ceiling
  minEndCard: 5.5,
  // block budget (research): ~0.5 push + ~0.8 drift + ~1.2 travel + ~0.5 mark + reading hold
  lead: [0.85, 0.5, 0.6], // drift on the top before travelling: first block, then alternating
  maxBlock: 6.6,          // article blocks run 5-6 s; never longer than this
  minZoom: 1.3,           // the zoom to the mention always reads as a move
};

// Square-root distance timing: two viewports take ~1.4x as long as one, capped.
export function scrollDuration(distance, viewport) {
  if (distance <= 1) return 0;
  return Math.min(1.3, Math.max(0.9, 0.9 * Math.sqrt(distance / Math.max(1, viewport))));
}

// Total length budget: 30-40 s for 4-6 articles, 5 s more per article beyond six. Longer reels
// lose viewers before the end card.
export function lengthBudget(articles) {
  return 40 + 5 * Math.max(0, articles - 6);
}

export function readSeconds(text, cps = TIMING.readCps) {
  return [...String(text || "")].length / cps;
}

// ---------------------------------------------------------------------------
// Camera math
// ---------------------------------------------------------------------------

// The camera maps a page point P (CSS px) to the screen: S = A + s * (P - F).
export function cameraAt(cam, t) {
  const p = progress(cam.travel, t);
  const q = progress(cam.zoom, t);
  const drift = 1 + cam.driftRate * Math.max(0, t - cam.driftStart);
  return {
    s: lerp(cam.s0, cam.s1, q) * drift,
    fx: lerp(cam.f0[0], cam.f1[0], p),
    fy: lerp(cam.f0[1], cam.f1[1], p),
    ax: lerp(cam.a0[0], cam.a1[0], p),
    ay: lerp(cam.a0[1], cam.a1[1], p),
  };
}

export function toScreen(c, rect) {
  return { x: c.ax + c.s * (rect.x - c.fx), y: c.ay + c.s * (rect.y - c.fy), w: rect.w * c.s, h: rect.h * c.s };
}

export function unionRect(rects) {
  if (!rects.length) return null;
  const x0 = Math.min(...rects.map((r) => r.x)), y0 = Math.min(...rects.map((r) => r.y));
  const x1 = Math.max(...rects.map((r) => r.x + r.w)), y1 = Math.max(...rects.map((r) => r.y + r.h));
  return { x: x0, y: y0, w: x1 - x0, h: y1 - y0 };
}

// Plan the article camera from the capture geometry and the mention rectangles.
// Returns page-fit and sentence-fit scales, focus points and anchors. The zoom never
// asks for more screen pixels per CSS pixel than the capture holds, drift included.
export function planCamera({ capture, card, focus, sceneSeconds, driftRate = TIMING.driftRate }) {
  const native = capture.scale;
  const headroom = 1 + driftRate * sceneSeconds;
  let s0 = Math.min(card.w / capture.column.w, native / headroom);
  const f0 = [capture.column.x + capture.column.w / 2, capture.top_y];
  const a0 = [card.x + card.w / 2, card.y];
  const m = capture.mention;
  if (!m || !m.sentence_bars.length) {
    return { s0, s1: s0, f0, f1: f0, a0, a1: a0, distance: 0, zoomFactor: 1 };
  }
  const sentence = unionRect(m.sentence_bars);
  const para = m.paragraph || sentence;
  const frameW = Math.max(sentence.w, Math.min(para.w, capture.column.w)) + 24;
  const frameH = sentence.h + 40;
  const s1 = Math.min(focus.w / frameW, focus.h / frameH, native / headroom);
  // a narrow column already fills the card: start it smaller so the zoom still lands
  s0 = Math.min(s0, s1 / TIMING.minZoom);
  const a1 = [focus.x + focus.w / 2, focus.y + focus.h / 2];
  let fx = Math.max(para.x, Math.min(sentence.x, para.x)) + frameW / 2 - 12;
  fx = Math.min(Math.max(fx, sentence.x + sentence.w / 2 - (focus.w / s1) / 2 + 12), sentence.x + sentence.w / 2 + (focus.w / s1) / 2 - 12);
  let fy = sentence.y + sentence.h / 2;
  // keep the zoomed-in window inside the captured strips (no blank page above or below)
  const above = (a1[1] - card.y) / s1, below = (card.y + card.h - a1[1]) / s1;
  fy = Math.min(Math.max(fy, capture.top_y + above), capture.bottom_y - below);
  const distance = Math.max(0, fy - (capture.top_y + (card.h / 2) / s0));
  return { s0, s1, f0, f1: [fx, fy], a0, a1, distance, zoomFactor: s1 / s0 };
}

// ---------------------------------------------------------------------------
// Text helpers shared with the compositor
// ---------------------------------------------------------------------------

export function shortUrl(url, max = 40) {
  let u;
  try { u = new URL(url); } catch { return String(url).slice(0, max); }
  const text = (u.hostname.replace(/^www\./, "") + u.pathname).replace(/\/$/, "");
  return text.length <= max ? text : text.slice(0, max - 1) + "\u2026";
}

const MONTHS = ["Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"];
export function dateLabel(iso) {
  const m = /^(\d{4})-(\d{2})-(\d{2})$/.exec(iso || "");
  if (!m) return null;
  return `${Number(m[3])} ${MONTHS[Number(m[2]) - 1]} ${m[1]}`;
}

export function dateSpanLabel(first, last) {
  if (!first) return null;
  if (!last || first === last) return dateLabel(first);
  const [y0, m0] = first.split("-"), [y1, m1] = last.split("-");
  const a = dateLabel(first), b = dateLabel(last);
  if (y0 === y1 && m0 === m1) return `${Number(first.slice(8))}\u2013${b}`;
  if (y0 === y1) return `${a.replace(` ${y0}`, "")} \u2013 ${b}`;
  return `${a} \u2013 ${b}`;
}

// ---------------------------------------------------------------------------
// Timeline builder
// ---------------------------------------------------------------------------

function round(n) {
  return Math.round(n * 1000) / 1000;
}

// items: reel rows prepared by render.mjs (display facts + capture geometry, see render.mjs)
// facts: calculated/sourced metric values for the hook line and the wall
export function buildTimeline({ brand, items, facts, format = "landscape", fps = 30, disclaimer }) {
  const F = FORMATS[format];
  if (!F) throw new Error(`unknown format ${format}`);
  if (!items.length) throw new Error("the reel needs at least one article with a capture");
  const scenes = [];
  const transitions = [];
  const texts = [];
  const read = (scene, role, text, from, until, exempt = false) => texts.push({ scene, role, text, from: round(from), until: round(until), exempt });

  // ---- opening: hook + masthead strip ------------------------------------------------
  // LinkedIn and X use frame 0 as the thumbnail, so the brand, period and the whole report
  // title are on screen from frame 0. The backdrop, facts line and mastheads animate in.
  const backdropIn = { start: 0, dur: 1.4, ease: "MOVE" };
  const factsIn = { start: 0.12, stagger: TIMING.lineStagger, dur: TIMING.wordReveal, ease: "MOVE" };
  const factsReadable = factsIn.start + factsIn.dur * 0.5;   // MOVE has covered ~90% by then
  const chips = items.filter((it) => it.masthead || it.outlet);
  const chipReveal = { start: 0.8, stagger: 0.05, dur: TIMING.chipIn, ease: "ARRIVE" };
  const chipsShown = chipReveal.start + Math.max(0, chips.length - 1) * chipReveal.stagger + chipReveal.dur;
  const openingExit = round(Math.max(2.8, chipsShown + 0.9, readSeconds(brand.report_title), factsReadable + readSeconds(facts.line)));
  const opening = {
    id: "opening", kind: "opening", start: 0, end: round(openingExit + TIMING.push),
    camera: { scale: { start: 0, dur: openingExit + TIMING.push, from: 1, to: 1.03, ease: "LINEAR", role: "drift" } },
    backdrop: { drift: 26, dim: 0.8, from: 0.94 },
    backdropIn, factsIn, chipReveal, chips: chips.map((it) => it.n),
  };
  scenes.push(opening);
  read("opening", "title", brand.report_title, 0, openingExit);
  read("opening", "facts", facts.line, factsReadable, openingExit);

  // ---- article blocks ------------------------------------------------------------------
  // Pass 1 finds each block's minimum length (reading speed decides it); the rhythm rules
  // then set the final lengths; pass 2 lays the blocks out on the clock.
  const A = F.article;
  const required = items.map((it, index) => articleBlock(it, index, 0, 0, A).duration);
  const lengths = rhythm(required);
  let cursor = openingExit;   // when the next push starts
  let prevId = "opening";
  items.forEach((it, index) => {
    const block = articleBlock(it, index, cursor, lengths[index], A);
    scenes.push(block.scene);
    transitions.push({ kind: "push", from: prevId, to: block.scene.id, start: round(cursor), dur: TIMING.push, ease: "MOVE" });
    block.reads.forEach((r) => read(block.scene.id, r.role, r.text, r.from, r.until));
    prevId = block.scene.id;
    cursor = block.scene.exitAt;
  });
  const lastArticle = scenes[scenes.length - 1];

  // ---- wall + metrics --------------------------------------------------------------------
  const wallStart = lastArticle.exitAt;
  const counters = facts.counters;
  const counterReveal = { start: round(wallStart + 0.4), stagger: 0.14, dur: 0.8, ease: "ARRIVE" };
  const countersShown = counterReveal.start + Math.max(0, counters.length - 1) * counterReveal.stagger + counterReveal.dur;
  // the counters are read one after another as they land
  const wallText = counters.map((c) => `${c.display} ${c.label}`).join(" ");
  const firstShown = counterReveal.start + counterReveal.dur;
  const wallExit = round(Math.max(wallStart + 3.0, countersShown + 1.2, firstShown + readSeconds(wallText), wallStart + 0.6 + readSeconds(facts.caption || "")));
  scenes.push({
    id: "wall", kind: "wall", start: round(wallStart), end: round(wallExit + TIMING.dissolve), exitAt: wallExit,
    camera: { scale: { start: wallStart, dur: wallExit + TIMING.dissolve - wallStart, from: 1.02, to: 1.06, ease: "LINEAR", role: "drift" } },
    marquee: 28, scrim: 0.78, counterReveal, counters, caption: facts.caption || "",
  });
  transitions.push({ kind: "push", from: lastArticle.id, to: "wall", start: round(wallStart), dur: TIMING.push, ease: "MOVE" });
  read("wall", "metrics", wallText, firstShown, wallExit);
  if (facts.caption) read("wall", "caption", facts.caption, wallStart + 0.6, wallExit);

  // ---- end card ---------------------------------------------------------------------------
  const endStart = wallExit;
  const endShown = endStart + TIMING.dissolve;
  // the end card is legible once the dissolve is half done
  const endExit = round(Math.max(endStart + TIMING.minEndCard, endStart + TIMING.dissolve / 2 + readSeconds(disclaimer) + 0.2));
  const listReveal = { start: round(endStart + 0.35), stagger: 0.07, dur: TIMING.chipIn, ease: "ARRIVE" };
  scenes.push({
    id: "end", kind: "end", start: round(endStart), end: endExit, exitAt: endExit,
    camera: { scale: { start: endStart, dur: endExit - endStart, from: 1, to: 1.018, ease: "LINEAR", role: "drift" } },
    listReveal, disclaimer,
  });
  transitions.push({ kind: "dissolve", from: "wall", to: "end", start: round(endStart), dur: TIMING.dissolve, ease: "DRIFT" });
  read("end", "sources", items.map((it) => `${it.outlet} ${it.date_label || ""} ${it.short_url}`).join(" "), endShown, endExit, true);
  read("end", "disclaimer", disclaimer, endStart + TIMING.dissolve / 2, endExit);

  const duration = endExit;
  return {
    version: TIMELINE_VERSION,
    format, width: F.width, height: F.height, fps,
    duration: round(duration),
    frames: Math.round(duration * fps),
    safe: F.safe,
    progress: { ...F.progress, ease: "LINEAR", role: "progress" },
    brand,
    facts,
    items,
    scenes,
    transitions,
    texts,
  };
}

// One article block starting at `start` (when its push-in begins). `length` is the time
// from push-in to the next block's push-in; 0 means "as short as reading allows".
function articleBlock(it, index, start, length, A) {
  const arrived = start + TIMING.push;
  const lead = index === 0 ? TIMING.lead[0] : TIMING.lead[1 + (index % 2)];   // first block teaches the pattern
  const headlineLines = Math.min(A.headline.lines, Math.max(1, Math.ceil([...(it.headline || "")].length / 24)));
  const headlineReveal = { start: round(start + 0.25), stagger: TIMING.lineStagger, dur: 0.55, ease: "MOVE" };
  const headlineShown = headlineReveal.start + (headlineLines - 1) * headlineReveal.stagger + headlineReveal.dur;
  const lowerIn = { start: round(start + 0.42), dur: TIMING.chipIn, ease: "ARRIVE" };
  const lowerText = [it.outlet, it.date_label, it.short_url].filter(Boolean).join(" ");
  const capture = it.capture;
  let camera = null, mark = null, dim = null, still = null, settle = arrived + lead, keyText = "";
  if (capture.kind === "clip") {
    // drift headroom sized for a long block so the zoom never upsamples, whatever the length
    const cam = planCamera({ capture, card: A.card, focus: A.focus, sceneSeconds: Math.max(12, length + TIMING.push) });
    camera = { ...cam, driftStart: round(start), driftRate: TIMING.driftRate, travel: null, zoom: null };
    if (capture.mention && capture.mention.sentence_bars.length) {
      const viewport = A.card.h / cam.s0;
      const scrollDur = scrollDuration(cam.distance, viewport) || 0.8;
      camera.travel = { start: round(arrived + lead), dur: round(scrollDur), ease: "MOVE" };
      const zoomStart = camera.travel.start + Math.max(0.35, scrollDur - 0.6);
      camera.zoom = { start: round(zoomStart), dur: TIMING.zoom, ease: "ARRIVE" };
      settle = zoomStart + TIMING.zoom * TIMING.zoomSettle;
      dim = { start: round(zoomStart + 0.15), dur: TIMING.dim, ease: "MOVE", to: 0.62 };
      const lines = capture.mention.clause_bars.length ? capture.mention.clause_bars : capture.mention.sentence_bars;
      const widest = Math.max(...lines.map((l) => l.w));
      let t = settle + TIMING.markDelay;
      const moves = lines.map((l) => {
        const dur = round(0.35 + 0.25 * (l.w / widest));
        const move = { start: round(t), dur, ease: "MARK" };
        t += dur + TIMING.markGap;
        return move;
      });
      mark = { lines: moves, end: round(t - TIMING.markGap) };
      keyText = capture.mention.clause || capture.mention.sentence;
    }
  } else {
    still = { driftStart: round(start), driftRate: TIMING.driftRate };
  }
  // the exit waits until every text block has been readable at the target speed
  const exitAt = Math.max(
    mark ? Math.max(mark.end + 1.0, settle + readSeconds(keyText)) : settle + 2.2,
    headlineShown + readSeconds(it.headline),
    lowerIn.start + lowerIn.dur + readSeconds(lowerText),
    start + length,
  );
  const scene = {
    id: `article-${it.n}`, kind: "article", item: it.n,
    start: round(start), end: round(exitAt + TIMING.push), exitAt: round(exitAt),
    headlineReveal, lowerIn,
    mastheadIn: { start: round(start + 0.2), dur: TIMING.chipIn, ease: "ARRIVE" },
    dim, mark, keyText, settle: round(settle), camera, still,
  };
  const reads = [
    { role: "headline", text: it.headline, from: headlineShown, until: exitAt },
    { role: "lower-third", text: lowerText, from: lowerIn.start + lowerIn.dur, until: exitAt },
  ];
  if (keyText) reads.push({ role: "mention", text: keyText, from: settle, until: exitAt });
  return { scene, reads, duration: exitAt - start };
}

// Final block lengths from the reading minimums: the first block is the longest (it teaches
// the viewer the pattern), neighbours never match, and the set varies by at least 10%.
export function rhythm(required) {
  const out = required.map((d) => d);
  if (out.length < 2) return out.map(round);
  // neighbours never match
  for (let i = 1; i < out.length; i++) {
    const near = Math.abs(out[i] - out[i - 1]) / Math.max(out[i], out[i - 1]) < 0.02;
    if (near) out[i] = Math.max(out[i], out[i - 1]) * 1.03;
  }
  // the first block teaches the pattern: later blocks run at least 10% shorter
  out[0] = Math.max(out[0], Math.max(...out.slice(1)) / 0.9);
  return out.map((d) => round(d + 1e-4));
}

// ---------------------------------------------------------------------------
// Frame-level state shared by the compositor and the lints
// ---------------------------------------------------------------------------

// How far each opening layer has animated in at t (0..1). `title` is 1 from frame 0 unless
// a scene asks for a title entrance, which the thumbnail lint rejects.
export function openingState(scene, t) {
  return {
    title: scene.titleIn ? progress(scene.titleIn, t) : 1,
    backdrop: progress(scene.backdropIn, t),
    facts: (line) => staggerProgress(scene.factsIn, line, t),
  };
}

// Which scenes are on screen at t, with their transition offsets.
export function scenesAt(timeline, t) {
  const out = [];
  for (const scene of timeline.scenes) {
    if (t < scene.start || t >= scene.end) continue;
    const incoming = timeline.transitions.find((tr) => tr.to === scene.id);
    const outgoing = timeline.transitions.find((tr) => tr.from === scene.id);
    const state = { scene, x: 0, scale: 1, alpha: 1, dim: 0 };
    if (incoming && t < incoming.start + incoming.dur) {
      const p = progress(incoming, t);
      if (incoming.kind === "push") state.x = (1 - p) * timeline.width;
      else state.alpha = p;
    }
    if (outgoing && t >= outgoing.start) {
      const p = progress(outgoing, t);
      if (outgoing.kind === "push") { state.x = -p * 0.3 * timeline.width; state.scale = 1 - 0.04 * p; state.dim = 0.45 * p; }
    }
    out.push(state);
  }
  return out;
}

// Scene-level camera: a number signature used to prove the frame never freezes.
export function cameraSignature(timeline, t) {
  const parts = [];
  for (const st of scenesAt(timeline, t)) {
    const s = st.scene;
    parts.push(st.x, st.scale, st.alpha);
    if (s.kind === "article" && s.camera) {
      const c = cameraAt(s.camera, t);
      parts.push(c.s, c.fx, c.fy, c.ax, c.ay);
    } else if (s.kind === "article" && s.still) {
      parts.push(1 + s.still.driftRate * (t - s.still.driftStart));
    } else if (s.camera && s.camera.scale) {
      const m = s.camera.scale;
      parts.push(lerp(m.from, m.to, progress(m, t)));
    }
  }
  return parts;
}

// ---------------------------------------------------------------------------
// Lints: the anti-pattern checklist as assertions
// ---------------------------------------------------------------------------

function inside(inner, outer, slack = 0.5) {
  return inner.x >= outer.x - slack && inner.y >= outer.y - slack
    && inner.x + inner.w <= outer.x + outer.w + slack && inner.y + inner.h <= outer.y + outer.h + slack;
}

function* movesOf(timeline) {
  for (const s of timeline.scenes) {
    const tag = (name) => `${s.id}.${name}`;
    if (s.titleIn) yield [tag("title"), s.titleIn, "reveal"];
    if (s.backdropIn) yield [tag("backdrop"), s.backdropIn, "reveal"];
    if (s.factsIn) yield [tag("facts"), s.factsIn, "reveal"];
    if (s.chipReveal) yield [tag("chips"), s.chipReveal, "reveal"];
    if (s.headlineReveal) yield [tag("headline"), s.headlineReveal, "reveal"];
    if (s.lowerIn) yield [tag("lower-third"), s.lowerIn, "reveal"];
    if (s.mastheadIn) yield [tag("masthead"), s.mastheadIn, "reveal"];
    if (s.dim) yield [tag("dim"), s.dim, "reveal"];
    if (s.mark) for (const [i, m] of s.mark.lines.entries()) yield [tag(`mark${i}`), m, "mark"];
    if (s.camera && s.camera.travel) yield [tag("travel"), s.camera.travel, "camera"];
    if (s.camera && s.camera.zoom) yield [tag("zoom"), s.camera.zoom, "camera"];
    if (s.camera && s.camera.scale) yield [tag("drift"), s.camera.scale, s.camera.scale.role];
    if (s.counterReveal) yield [tag("counters"), s.counterReveal, "reveal"];
    if (s.listReveal) yield [tag("sources"), s.listReveal, "reveal"];
  }
  for (const tr of timeline.transitions) yield [`transition ${tr.from}->${tr.to}`, tr, "transition"];
  yield ["progress", timeline.progress, timeline.progress.role];
}

export function lintTimeline(timeline) {
  const problems = [];
  const fail = (rule, detail) => problems.push({ rule, detail });
  const dt = 1 / timeline.fps;

  // 1. only named easings; linear only for drifts, marquees and the progress bar
  for (const [name, move, role] of movesOf(timeline)) {
    if (!EASE_NAMES.includes(move.ease)) fail("named-easing", `${name} uses "${move.ease}"`);
    else if (move.ease === "LINEAR" && !LINEAR_ROLES.includes(role)) fail("linear-only-for-drift", `${name} is linear`);
  }

  // 2. no overshoot beyond 2% on any move
  for (const name of EASE_NAMES) {
    let lo = 0, hi = 1;
    for (let i = 0; i <= 200; i++) { const v = ease(name, i / 200); lo = Math.min(lo, v); hi = Math.max(hi, v); }
    if (lo < -0.02 || hi > 1.02) fail("no-overshoot", `${name} reaches ${lo.toFixed(3)}..${hi.toFixed(3)}`);
  }

  // 3. at most two transition types
  const kinds = [...new Set(timeline.transitions.map((t) => t.kind))];
  if (kinds.length > 2) fail("max-two-transition-types", kinds.join(", "));
  const dissolves = timeline.transitions.filter((t) => t.kind === "dissolve");
  if (dissolves.some((t) => t.to !== "end")) fail("dissolve-only-into-end-card", dissolves.map((t) => t.to).join(", "));

  // 4. article blocks are never uniform: spread of at least 10%
  const blocks = timeline.scenes.filter((s) => s.kind === "article").map((s) => s.exitAt - s.start);
  if (blocks.length > 1) {
    const spread = (Math.max(...blocks) - Math.min(...blocks)) / Math.max(...blocks);
    if (spread < 0.1 - 1e-9) fail("varied-block-lengths", `block lengths ${blocks.map((b) => b.toFixed(2)).join(", ")} vary by ${(spread * 100).toFixed(1)}%`);
  }

  // 5. the camera moves on every frame (no slideshow freezes)
  let prev = null, frozen = 0;
  for (let i = 0; i < timeline.frames; i++) {
    const sig = cameraSignature(timeline, i * dt);
    if (prev && sig.length === prev.length && sig.every((v, k) => Math.abs(v - prev[k]) < 1e-7)) {
      frozen += 1;
      if (frozen > 6) { fail("camera-never-static", `frame ${i} (${(i * dt).toFixed(2)} s) repeats the previous camera`); break; }
    } else frozen = 0;
    prev = sig;
  }

  // 6. never upsample a capture past its native pixels
  for (const s of timeline.scenes) {
    if (s.kind !== "article" || !s.camera) continue;
    const it = timeline.items.find((x) => x.n === s.item);
    for (let t = s.start; t < s.end; t += dt) {
      const c = cameraAt(s.camera, t);
      if (c.s > it.capture.scale + 1e-6) { fail("no-upsampling", `${s.id} draws ${c.s.toFixed(3)} px per CSS px from a ${it.capture.scale}x capture at ${t.toFixed(2)} s`); break; }
    }
  }

  // 7. reading speed: every text block stays up long enough
  for (const r of timeline.texts) {
    const seconds = r.until - r.from;
    if (r.exempt) {
      if (seconds < TIMING.minEndCard - 1 - 1e-6) fail("reference-hold", `${r.scene} ${r.role} is up ${seconds.toFixed(2)} s`);
      continue;
    }
    const cps = [...String(r.text)].length / Math.max(0.001, seconds);
    if (cps > TIMING.maxCps + 1e-6) fail("reading-speed", `${r.scene} ${r.role} needs ${cps.toFixed(1)} characters per second ("${String(r.text).slice(0, 40)}")`);
  }

  // 8. key text inside the format's safe area, and the mention lands in its focus box
  const F = FORMATS[timeline.format];
  const textSlots = [
    ...["masthead", "kicker", "headline", "byline", "lower", "micro"].map((k) => [`article.${k}`, F.article[k]]),
    ...["overline", "title", "facts", "strip"].map((k) => [`opening.${k}`, F.opening[k]]),
    ["wall.metrics", F.wall.metrics], ["wall.caption", F.wall.caption],
    ...["overline", "list", "disclaimer"].map((k) => [`end.${k}`, F.end[k]]),
    ["article.focus", F.article.focus],
  ];
  for (const [name, slot] of textSlots) if (!inside(slot, F.safe)) fail("safe-area", `${name} leaves the ${timeline.format} safe area`);
  if (!inside(F.article.focus, F.article.card)) fail("safe-area", "article.focus leaves the card");
  const overlaps = (a, b) => a.x < b.x + b.w && b.x < a.x + a.w && a.y < b.y + b.h && b.y < a.y + a.h;
  for (const k of ["masthead", "kicker", "headline", "byline", "lower", "micro"]) {
    if (overlaps(F.article[k], F.article.card)) fail("text-off-card", `article.${k} overlaps the page card in ${timeline.format}`);
  }
  for (const s of timeline.scenes) {
    if (s.kind !== "article" || !s.camera || !s.keyText) continue;
    const it = timeline.items.find((x) => x.n === s.item);
    const c = cameraAt(s.camera, Math.min(s.exitAt, s.settle + 0.4));
    const box = toScreen(c, unionRect(it.capture.mention.sentence_bars));
    if (!inside(box, { x: F.article.card.x, y: F.article.card.y, w: F.article.card.w, h: F.article.card.h }, 1)
      || !inside(box, F.safe, 1)) fail("mention-in-safe-area", `${s.id} mention lands outside the card or safe area`);
  }

  // 9. total length stays in budget (30-40 s for 4-6 articles) and no article block drags
  const articleCount = timeline.scenes.filter((s) => s.kind === "article").length;
  if (timeline.duration > lengthBudget(articleCount) + 1e-6) fail("reel-length", `${timeline.duration.toFixed(1)} s for ${articleCount} article${articleCount === 1 ? "" : "s"}; the budget is ${lengthBudget(articleCount)} s (feature fewer articles with include_in_reel, or re-capture with --pick for a shorter mention)`);
  for (const s of timeline.scenes) {
    if (s.kind === "article" && s.exitAt - s.start > TIMING.maxBlock + 1e-6) fail("block-length", `${s.id} runs ${(s.exitAt - s.start).toFixed(2)} s; article blocks stay under ${TIMING.maxBlock} s (shorten the highlighted clause)`);
  }

  // 10. frame 0 is a usable thumbnail: the opening is on screen and fully opaque, and the
  // whole report title is already up (LinkedIn and X use frame 0 as the thumbnail)
  const first = scenesAt(timeline, 0);
  if (!first.length || first[0].scene.kind !== "opening" || first[0].alpha < 1) fail("no-fade-from-black", "frame 0 is not the opening at full opacity");
  const opening = timeline.scenes.find((s) => s.kind === "opening");
  const titleRead = timeline.texts.find((r) => r.scene === "opening" && r.role === "title");
  if (!opening || openingState(opening, 0).title < 1 || !titleRead || titleRead.from > 0) fail("thumbnail-title", "the report title is not fully visible at frame 0");

  return problems;
}
