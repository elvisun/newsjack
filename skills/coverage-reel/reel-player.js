// coverage-reel compositor and player (browser only).
//
// render.mjs inlines timeline.mjs (exports stripped) above this file inside reel.html, so
// FORMATS, cameraAt, scenesAt, progress, staggerProgress, toScreen and friends are in scope.
// drawFrame(ctx, t) is a pure function of time: the same t always draws the same pixels.
// Captured pages are drawn as they are; highlights are multiply-blended over the original
// pixels, never retyped. Mastheads are cropped from the captures and only moved or faded.

const DATA = JSON.parse(document.getElementById("reel-data").textContent);
const TL = DATA.timeline;
const FMT = FORMATS[TL.format];
const W = TL.width, H = TL.height;
const PAL = DATA.palette;
const SANS = 'ui-sans-serif, system-ui, -apple-system, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif';
const SERIF = 'ui-serif, "New York", Georgia, "Times New Roman", serif';
const ITEMS = new Map(TL.items.map((it) => [it.n, it]));
const IMAGES = new Map();

// ---------------------------------------------------------------------------
// Assets
// ---------------------------------------------------------------------------

function imageSources() {
  const srcs = new Set();
  for (const it of TL.items) {
    if (it.thumb) srcs.add(it.thumb.src);
    if (it.masthead) srcs.add(it.masthead.src);
    if (it.capture.kind === "clip") it.capture.strips.forEach((s) => srcs.add(s.src));
    else srcs.add(it.capture.src);
  }
  if (TL.brand.logo) srcs.add(TL.brand.logo.src);
  return [...srcs];
}

async function loadImages() {
  await Promise.all(imageSources().map(async (src) => {
    const img = new Image();
    img.decoding = "sync";
    img.src = src;
    await img.decode();
    IMAGES.set(src, img);
  }));
  await document.fonts.ready;
  // warm the system fonts so the first frame never draws with a fallback face
  const probe = document.createElement("canvas").getContext("2d");
  for (const f of [`700 40px ${SERIF}`, `600 40px ${SANS}`, `400 40px ${SANS}`]) { probe.font = f; probe.measureText("Warm up 0123"); }
}

// ---------------------------------------------------------------------------
// Drawing helpers
// ---------------------------------------------------------------------------

function roundRectPath(ctx, x, y, w, h, r) {
  const rr = Math.max(0, Math.min(r, w / 2, h / 2));
  ctx.beginPath();
  ctx.moveTo(x + rr, y);
  ctx.arcTo(x + w, y, x + w, y + h, rr);
  ctx.arcTo(x + w, y + h, x, y + h, rr);
  ctx.arcTo(x, y + h, x, y, rr);
  ctx.arcTo(x, y, x + w, y, rr);
  ctx.closePath();
}

function rgba(hex, alpha) {
  const n = parseInt(hex.slice(1), 16);
  return `rgba(${(n >> 16) & 255}, ${(n >> 8) & 255}, ${n & 255}, ${alpha})`;
}

const LAYOUT_CACHE = new Map();

// Wrap text into lines that fit `maxWidth` with the current ctx.font.
function wrap(ctx, text, maxWidth) {
  const words = String(text).split(/\s+/).filter(Boolean);
  const lines = [];
  let line = "";
  for (const word of words) {
    const next = line ? `${line} ${word}` : word;
    if (ctx.measureText(next).width <= maxWidth || !line) {
      if (!line && ctx.measureText(word).width > maxWidth) {
        // a single word wider than the box: break it by characters
        let part = "";
        for (const ch of word) {
          if (ctx.measureText(part + ch).width > maxWidth && part) { lines.push(part); part = ""; }
          part += ch;
        }
        line = part;
      } else line = next;
    } else {
      lines.push(line);
      line = word;
    }
  }
  if (line) lines.push(line);
  return lines;
}

// Largest font size (stepping down to `min`) at which the text fits the box. Past `min`,
// the last line is cut with an ellipsis, so text can never spill out of its slot.
function fitText(ctx, text, box, { size, min = size, lines: maxLines = 1, weight = 400, family = SANS, lineHeight = 1.15 }) {
  const key = [text, box.w, box.h, size, min, maxLines, weight, family, lineHeight].join("|");
  if (LAYOUT_CACHE.has(key)) return LAYOUT_CACHE.get(key);
  let result = null;
  for (let s = size; s >= min; s -= 2) {
    ctx.font = `${weight} ${s}px ${family}`;
    const lines = wrap(ctx, text, box.w);
    if (lines.length <= maxLines && lines.length * s * lineHeight <= box.h + 1) { result = { size: s, lines, lh: s * lineHeight, font: ctx.font }; break; }
  }
  if (!result) {
    ctx.font = `${weight} ${min}px ${family}`;
    const lines = wrap(ctx, text, box.w);
    const keep = Math.max(1, Math.min(maxLines, Math.floor((box.h + 1) / (min * lineHeight))));
    const cut = lines.slice(0, keep);
    if (lines.length > keep) {
      let last = cut[keep - 1];
      while (last && ctx.measureText(`${last}\u2026`).width > box.w) last = last.replace(/\s*\S+$/, "");
      cut[keep - 1] = `${last}\u2026`;
    }
    result = { size: min, lines: cut, lh: min * lineHeight, font: ctx.font };
  }
  LAYOUT_CACHE.set(key, result);
  return result;
}

function drawLines(ctx, layout, x, y, color, revealSpec, t, align = "left", width = 0) {
  ctx.font = layout.font;
  ctx.fillStyle = color;
  ctx.textBaseline = "alphabetic";
  layout.lines.forEach((line, i) => {
    const p = revealSpec ? staggerProgress(revealSpec, i, t) : 1;
    if (p <= 0) return;
    const top = y + i * layout.lh;
    const lw = ctx.measureText(line).width;
    const lx = align === "right" ? x + width - lw : align === "center" ? x + (width - lw) / 2 : x;
    ctx.save();
    ctx.beginPath();
    ctx.rect(x - 4, top - layout.lh * 0.05, Math.max(width, 4000) + 8, layout.lh * 1.12);
    ctx.clip();
    ctx.fillText(line, lx, top + layout.size * 0.92 + (1 - p) * layout.lh);
    ctx.restore();
  });
}

function text(ctx, str, x, y, { size, weight = 400, family = SANS, color, align = "left", spacing = 0, baseline = "alphabetic" }) {
  ctx.font = `${weight} ${size}px ${family}`;
  ctx.fillStyle = color;
  ctx.textAlign = align;
  ctx.textBaseline = baseline;
  ctx.letterSpacing = `${spacing}px`;
  ctx.fillText(str, x, y);
  ctx.letterSpacing = "0px";
  ctx.textAlign = "left";
}

// Draw `src` sub-rect of an image into a destination rect, trimmed to `clip` (screen space).
function drawImageRect(ctx, img, sx, sy, sw, sh, dx, dy, dw, dh) {
  if (sw <= 0 || sh <= 0 || dw <= 0 || dh <= 0) return;
  ctx.drawImage(img, sx, sy, sw, sh, dx, dy, dw, dh);
}

// A masthead exactly as captured: the crop of the stamped logo (on its own plate) from the
// article's top capture, scaled to a common height. Only the whole crop moves or fades.
function mastheadChip(ctx, it, x, y, h, alpha, align = "left", maxW = Infinity) {
  if (alpha <= 0) return 0;
  const m = it.masthead;
  ctx.save();
  ctx.globalAlpha *= alpha;
  if (m) {
    const img = IMAGES.get(m.src);
    let w = (m.crop.w / m.crop.h) * h;
    let hh = h;
    if (w > maxW) { hh = h * (maxW / w); w = maxW; }
    const dx = align === "right" ? x - w : x;
    const dy = y + (h - hh) / 2;
    roundRectPath(ctx, dx, dy, w, hh, Math.min(8, hh * 0.12));
    ctx.fillStyle = m.plate || "#ffffff";
    ctx.fill();
    ctx.clip();
    drawImageRect(ctx, img, m.crop.x, m.crop.y, m.crop.w, m.crop.h, dx, dy, w, hh);
    ctx.restore();
    return w;
  }
  // no captured masthead (a user-supplied screenshot): the outlet's name as plain text
  ctx.font = `700 ${Math.round(h * 0.5)}px ${SANS}`;
  const w = Math.min(maxW, ctx.measureText(it.outlet).width + h * 0.7);
  const dx = align === "right" ? x - w : x;
  roundRectPath(ctx, dx, y, w, h, Math.min(8, h * 0.12));
  ctx.fillStyle = "#ffffff";
  ctx.fill();
  text(ctx, it.outlet, dx + w / 2, y + h * 0.66, { size: Math.round(h * 0.5), weight: 700, color: "#111111", align: "center" });
  ctx.restore();
  return w;
}

function stageBackground(ctx) {
  const g = ctx.createLinearGradient(0, 0, W * 0.4, H);
  g.addColorStop(0, PAL.stageLift);
  g.addColorStop(1, PAL.stage);
  ctx.fillStyle = g;
  ctx.fillRect(0, 0, W, H);
}

// A top capture as a mosaic tile (the article's real top: masthead, headline, hero).
function drawThumb(ctx, it, x, y, w, h) {
  const th = it.thumb;
  if (!th) return;
  const img = IMAGES.get(th.src);
  const scale = w / th.width;
  const sh = Math.min(th.height, h / scale);
  ctx.save();
  roundRectPath(ctx, x, y, w, h, 10);
  ctx.fillStyle = "#ffffff";
  ctx.fill();
  ctx.clip();
  drawImageRect(ctx, img, 0, 0, th.width, sh, x, y, w, sh * scale);
  ctx.restore();
}

// ---------------------------------------------------------------------------
// Scenes
// ---------------------------------------------------------------------------

function drawOpening(ctx, scene, t) {
  const O = FMT.opening;
  const st = openingState(scene, t);
  stageBackground(ctx);
  const zoom = progress(scene.camera.scale, t);
  // the backdrop settles in from slightly closer and darker while it starts to drift
  const s = (scene.camera.scale.from + (scene.camera.scale.to - scene.camera.scale.from) * zoom) * (1 + 0.04 * (1 - st.backdrop));
  // backdrop: the real article tops, drifting, under a deep scrim
  ctx.save();
  ctx.translate(W / 2, H / 2); ctx.scale(s, s); ctx.translate(-W / 2, -H / 2);
  const tiles = TL.items.filter((it) => it.thumb);
  if (tiles.length) {
    const cols = W > H ? 5 : 3;
    const tw = (W * 1.15) / cols, th = tw * 1.25, gap = 26;
    const rows = Math.ceil(H / (th + gap)) + 1;
    for (let r = 0; r < rows; r++) {
      const dir = r % 2 ? -1 : 1;
      const off = (dir * scene.backdrop.drift * t) % (tw + gap) - (tw + gap);
      for (let c = 0; c < cols + 2; c++) {
        const it = tiles[(r * 3 + c) % tiles.length];
        drawThumb(ctx, it, off + c * (tw + gap) - W * 0.06, r * (th + gap) - th * 0.35, tw, th);
      }
    }
  }
  ctx.restore();
  const dim = scene.backdrop.from + (scene.backdrop.dim - scene.backdrop.from) * st.backdrop;
  const scrim = ctx.createLinearGradient(0, 0, W, H * 0.3);
  scrim.addColorStop(0, rgba(PAL.stage, Math.max(0.95, dim)));
  scrim.addColorStop(1, rgba(PAL.stage, dim));
  ctx.fillStyle = scrim;
  ctx.fillRect(0, 0, W, H);

  // brand mark, overline and the whole verbatim title are on screen from frame 0: it is
  // the thumbnail on LinkedIn and X
  drawBrand(ctx, O.brand, PAL.text, 1);
  text(ctx, (TL.brand.overline || "").toUpperCase(), O.overline.x, O.overline.y + O.overline.size, { size: O.overline.size, weight: 700, color: PAL.accentOnDark, spacing: 3 });
  const T = O.title;
  const layout = titleLayout(ctx);
  drawLines(ctx, layout, T.x, T.y, PAL.text, st.title < 1 ? scene.titleIn : null, t, "left", T.w);
  // the calculated facts line rises in under it
  const facts = fitText(ctx, TL.facts.line, O.facts, { size: O.facts.size, min: 22, lines: 2, weight: 500 });
  drawLines(ctx, facts, O.facts.x, O.facts.y, PAL.muted, scene.factsIn, t, "left", O.facts.w);

  // masthead strip: each captured masthead, same height, staggered in
  const S = O.strip;
  let x = S.x, y = S.y, h = S.chip;
  const chips = scene.chips.map((n) => ITEMS.get(n));
  const widthOf = (it, hh) => (it.masthead ? (it.masthead.crop.w / it.masthead.crop.h) * hh : hh * 4);
  const total = chips.reduce((sum, it) => sum + widthOf(it, h) + S.gap, -S.gap);
  if (total > S.w * 2) h = Math.max(36, h * (S.w * 2) / total);
  chips.forEach((it, i) => {
    const w = Math.min(widthOf(it, h), S.w);
    if (x + w > S.x + S.w + 0.5) { x = S.x; y += h + S.gap; }
    const p = staggerProgress(scene.chipReveal, i, t);
    if (p > 0) mastheadChip(ctx, it, x, y + (1 - p) * 18, h, p, "left", S.w);
    x += w + S.gap;
  });
}

function titleLayout(ctx) {
  const T = FMT.opening.title;
  return fitText(ctx, TL.brand.report_title, T, { size: T.size, min: T.min, lines: T.lines, weight: 700, family: SERIF, lineHeight: 1.04 });
}

function drawBrand(ctx, slot, color, alpha) {
  ctx.save();
  ctx.globalAlpha *= alpha;
  const logo = TL.brand.logo;
  if (logo) {
    const img = IMAGES.get(logo.src);
    const w = Math.min(slot.w, (logo.width / logo.height) * slot.h);
    const h = w * (logo.height / logo.width);
    ctx.drawImage(img, slot.x, slot.y + (slot.h - h) / 2, w, h);
  } else {
    text(ctx, TL.brand.name, slot.x, slot.y + slot.h * 0.78, { size: Math.round(slot.h * 0.72), weight: 700, color });
  }
  ctx.restore();
}

function drawArticle(ctx, scene, t) {
  const it = ITEMS.get(scene.item);
  const A = FMT.article;
  const card = A.card;
  stageBackground(ctx);

  // the page card
  ctx.save();
  ctx.shadowColor = "rgba(0, 0, 0, 0.2)";
  ctx.shadowBlur = 60;
  ctx.shadowOffsetY = 20;
  roundRectPath(ctx, card.x, card.y, card.w, card.h, card.r);
  ctx.fillStyle = "#ffffff";
  ctx.fill();
  ctx.restore();
  ctx.save();
  roundRectPath(ctx, card.x, card.y, card.w, card.h, card.r);
  ctx.clip();
  if (it.capture.kind === "clip") drawPage(ctx, scene, it, t);
  else drawStill(ctx, scene, it, t);
  ctx.restore();

  // the text layer travels a little faster than the card during the push in (parallax depth)
  const pushIn = TL.transitions.find((tr) => tr.to === scene.id);
  const lag = pushIn ? (1 - progress(pushIn, t)) * W * 0.16 : 0;
  ctx.save();
  ctx.translate(lag, 0);
  // outlet, kicker, verbatim headline, byline
  const mp = progress(scene.mastheadIn, t);
  mastheadChip(ctx, it, A.masthead.x, A.masthead.y + (1 - mp) * 16, A.masthead.h, mp, "left", A.masthead.w);
  if (it.kicker) {
    const kp = progress(scene.mastheadIn, t - 0.08);
    ctx.save(); ctx.globalAlpha *= kp;
    text(ctx, it.kicker.toUpperCase(), A.kicker.x, A.kicker.y + A.kicker.size, { size: A.kicker.size, weight: 700, color: PAL.accentOnDark, spacing: 2.5 });
    ctx.restore();
  }
  const HL = A.headline;
  const headline = fitText(ctx, it.headline, HL, { size: HL.size, min: HL.min, lines: HL.lines, weight: 700, family: SERIF, lineHeight: 1.08 });
  drawLines(ctx, headline, HL.x, HL.y, PAL.text, scene.headlineReveal, t, "left", HL.w);
  if (it.byline) {
    const bp = staggerProgress(scene.headlineReveal, headline.lines.length, t);
    ctx.save(); ctx.globalAlpha *= bp;
    const by = Math.min(A.byline.y, HL.y + headline.lines.length * headline.lh + 18);
    const bl = fitText(ctx, `By ${it.byline}`, A.byline, { size: A.byline.size, min: 16, lines: 1, weight: 500 });
    drawLines(ctx, bl, A.byline.x, by, PAL.muted, null, t, "left", A.byline.w);
    ctx.restore();
  }

  // lower-third: outlet, publish date, short URL (persistent through the block)
  const lp = progress(scene.lowerIn, t);
  if (lp > 0) {
    const L = A.lower;
    ctx.save();
    ctx.globalAlpha *= lp;
    ctx.translate(0, (1 - lp) * 24);
    ctx.fillStyle = PAL.accent;
    ctx.fillRect(L.x, L.y, 6, L.h);
    const line1 = [it.outlet, it.date_label].filter(Boolean).join("  \u00b7  ");
    const l1 = fitText(ctx, line1, { w: L.w - 28, h: L.size * 1.3 }, { size: L.size, min: 18, lines: 1, weight: 700 });
    drawLines(ctx, l1, L.x + 26, L.y + L.h * 0.08, PAL.text, null, t, "left", L.w - 28);
    const l2 = fitText(ctx, it.short_url, { w: L.w - 28, h: L.small * 1.3 }, { size: L.small, min: 16, lines: 1, weight: 400 });
    drawLines(ctx, l2, L.x + 26, L.y + L.h * 0.08 + L.size * 1.45, PAL.muted, null, t, "left", L.w - 28);
    ctx.restore();
  }
  // "Highlights added" whenever a highlight is on screen
  if (scene.mark) {
    const hp = progress({ start: scene.mark.lines[0].start, dur: 0.4, ease: "MOVE" }, t);
    if (hp > 0) {
      const M = A.micro;
      ctx.save(); ctx.globalAlpha *= hp;
      ctx.font = `500 ${M.size}px ${SANS}`;
      const label = "Highlights added";
      const width = M.size * 1.35 + ctx.measureText(label).width;
      const x0 = M.align === "right" ? M.x + M.w - width : M.x;
      ctx.fillStyle = PAL.highlightSolid;
      roundRectPath(ctx, x0, M.y + M.h * 0.18, M.size * 0.9, M.size * 0.9, 3);
      ctx.fill();
      text(ctx, label, x0 + M.size * 1.35, M.y + M.size, { size: M.size, weight: 500, color: PAL.muted });
      ctx.restore();
    }
  }
  ctx.restore();
}

// The captured page inside the card: stacked strips under the planned camera.
function drawPage(ctx, scene, it, t) {
  const A = FMT.article, card = A.card, cap = it.capture;
  const cam = cameraAt(scene.camera, t);
  const vis = { x: cam.fx + (card.x - cam.ax) / cam.s, y: cam.fy + (card.y - cam.ay) / cam.s, w: card.w / cam.s, h: card.h / cam.s };
  const strips = () => {
    for (const st of cap.strips) {
      const y0 = Math.max(st.y, vis.y), y1 = Math.min(st.y + st.h, vis.y + vis.h);
      if (y1 <= y0) continue;
      const k = st.width / cap.column.w;   // image px per CSS px
      const img = IMAGES.get(st.src);
      const sx = Math.max(0, (vis.x - cap.column.x) * k), ex = Math.min(st.width, (vis.x + vis.w - cap.column.x) * k);
      const sy = (y0 - st.y) * k, ey = (y1 - st.y) * k;
      const d = toScreen(cam, { x: cap.column.x + sx / k, y: y0, w: (ex - sx) / k, h: y1 - y0 });
      drawImageRect(ctx, img, sx, sy, ex - sx, ey - sy, d.x, d.y, d.w, d.h);
    }
  };
  const m = cap.mention;
  const dimP = scene.dim ? progress(scene.dim, t) : 0;
  const holes = m ? m.sentence_bars.map((b) => toScreen(cam, b)) : [];
  const holePath = () => {
    ctx.beginPath();
    for (const h of holes) {
      ctx.roundRect(h.x - 4, h.y - 4, h.w + 8, h.h + 8, Math.min(6, h.h * 0.25));
    }
  };
  if (dimP > 0.002) {
    ctx.save();
    ctx.filter = `blur(${(3 * dimP).toFixed(2)}px)`;
    strips();
    ctx.restore();
    ctx.save();
    holePath();
    ctx.clip();
    ctx.fillStyle = "#ffffff";
    ctx.fillRect(card.x, card.y, card.w, card.h);
    strips();
    ctx.restore();
  } else {
    strips();
  }
  // highlighter: multiply-blended over the original pixels, line by line
  if (scene.mark) {
    const bars = m.clause_bars.length ? m.clause_bars : m.sentence_bars;
    ctx.save();
    ctx.globalCompositeOperation = "multiply";
    ctx.fillStyle = PAL.highlight;
    bars.forEach((b, i) => {
      const p = progress(scene.mark.lines[i], t);
      if (p <= 0) return;
      const r = toScreen(cam, b);
      const padX = 3, top = r.y - r.h * 0.04, h = r.h * 1.08;
      const w = (r.w + padX * 2) * p;
      ctx.beginPath();
      ctx.moveTo(r.x - padX + 1.5, top);
      ctx.lineTo(r.x - padX + w + 1.5, top);
      ctx.quadraticCurveTo(r.x - padX + w + 4, top + h / 2, r.x - padX + w - 1.5, top + h);
      ctx.lineTo(r.x - padX - 1.5, top + h);
      ctx.closePath();
      ctx.fill();
    });
    ctx.restore();
  }
  // wash the surround so the sentence reads first
  if (dimP > 0.002) {
    ctx.save();
    ctx.beginPath();
    ctx.rect(card.x, card.y, card.w, card.h);
    for (const h of holes) ctx.roundRect(h.x - 4, h.y - 4, h.w + 8, h.h + 8, Math.min(6, h.h * 0.25));
    ctx.fillStyle = `rgba(255, 255, 255, ${(scene.dim.to * dimP).toFixed(4)})`;
    ctx.fill("evenodd");
    ctx.restore();
  }
}

// A user-supplied screenshot: fit to the card width with a slow drift, no highlight.
function drawStill(ctx, scene, it, t) {
  const card = FMT.article.card, cap = it.capture;
  const img = IMAGES.get(cap.src);
  const s = Math.min(card.w / cap.width, 1) * (1 + scene.still.driftRate * (t - scene.still.driftStart));
  ctx.drawImage(img, card.x + card.w / 2 - (cap.width * s) / 2, card.y, cap.width * s, cap.height * s);
}

function drawWall(ctx, scene, t) {
  const Wl = FMT.wall;
  stageBackground(ctx);
  const zoom = progress(scene.camera.scale, t);
  const s = scene.camera.scale.from + (scene.camera.scale.to - scene.camera.scale.from) * zoom;
  ctx.save();
  ctx.translate(W / 2, H / 2); ctx.scale(s, s); ctx.translate(-W / 2, -H / 2);
  ctx.filter = "blur(2px)";   // a texture of the real coverage; the numbers carry the scene
  const tiles = TL.items.filter((it) => it.thumb);
  const rows = W > H ? 3 : 4;
  const th = H / (rows - 0.4), tw = th * 0.82, gap = 22;
  for (let r = 0; r < rows; r++) {
    const dir = r % 2 ? 1 : -1;
    const off = ((dir * scene.marquee * (t - scene.start)) % (tw + gap)) - (tw + gap) * 1.5;
    const count = Math.ceil(W / (tw + gap)) + 3;
    for (let c = 0; c < count; c++) {
      const it = tiles[(c + r * 2) % tiles.length];
      drawThumb(ctx, it, off + c * (tw + gap), r * (th + gap) - th * 0.3, tw, th);
    }
  }
  ctx.restore();
  ctx.fillStyle = rgba(PAL.stage, scene.scrim);
  ctx.fillRect(0, 0, W, H);
  const vg = ctx.createRadialGradient(W / 2, H / 2, Math.min(W, H) * 0.2, W / 2, H / 2, Math.max(W, H) * 0.75);
  vg.addColorStop(0, rgba(PAL.stage, 0));
  vg.addColorStop(1, rgba(PAL.stage, 0.7));
  ctx.fillStyle = vg;
  ctx.fillRect(0, 0, W, H);

  // counters: calculated or sourced metrics only, each with its status, centred as a group
  const M = Wl.metrics;
  const n = scene.counters.length;
  const cols = W > H ? Math.min(4, n) : Math.min(2, n);
  const rowsN = Math.ceil(n / cols);
  const cw = Math.min(M.w / cols, W > H ? 520 : 470);
  const ch = Math.min(M.h / rowsN, Wl.value * 2.6);
  const gx = W / 2 - (cols * cw) / 2, gy = M.y + (M.h - rowsN * ch) / 2;
  const glow = ctx.createRadialGradient(W / 2, gy + (rowsN * ch) / 2, 10, W / 2, gy + (rowsN * ch) / 2, Math.max(cols * cw, rowsN * ch) * 0.75);
  glow.addColorStop(0, rgba(PAL.stage, 0.72));
  glow.addColorStop(1, rgba(PAL.stage, 0));
  ctx.fillStyle = glow;
  ctx.fillRect(0, 0, W, H);
  if (scene.caption) {
    const C = Wl.caption;
    const cp = progress({ start: scene.start + 0.5, dur: 0.5, ease: "ARRIVE" }, t);
    ctx.save(); ctx.globalAlpha *= cp;
    const cl = fitText(ctx, scene.caption, { w: W - 2 * C.x, h: C.h }, { size: C.size, min: 18, lines: 1, weight: 600 });
    drawLines(ctx, cl, C.x, gy - C.size * 2.4, PAL.muted, null, t, "center", W - 2 * C.x);
    ctx.restore();
  }
  scene.counters.forEach((c, i) => {
    const p = staggerProgress(scene.counterReveal, i, t);
    if (p <= 0) return;
    const x = gx + (i % cols) * cw, y = gy + Math.floor(i / cols) * ch;
    ctx.save();
    ctx.globalAlpha *= Math.min(1, p * 1.6);
    ctx.translate(0, (1 - p) * 26);
    const valueText = c.kind === "count" ? formatCount(Math.round(c.value * p)) : c.display;
    const vl = fitText(ctx, c.display, { w: cw - 40, h: Wl.value * 1.2 }, { size: Wl.value, min: 28, lines: 1, weight: 700 });
    ctx.font = `700 ${vl.size}px ${SANS}`;
    drawDigits(ctx, valueText, x + cw / 2 - digitsWidth(ctx, c.display) / 2, y + vl.size * 0.95, PAL.text);
    const ly = y + vl.size * 1.18;
    const lab = fitText(ctx, c.label, { w: cw - 40, h: Wl.label * 2.6 }, { size: Wl.label, min: 18, lines: 2, weight: 500 });
    drawLines(ctx, lab, x + 20, ly, PAL.text, null, t, "center", cw - 40);
    const chipSize = Math.round(Wl.label * 0.62);
    statusChip(ctx, c.status, x + cw / 2, ly + lab.lines.length * lab.lh + 14, chipSize, true);
    ctx.restore();
  });
}

function digitsWidth(ctx, str) {
  const cell = ctx.measureText("0").width;
  let w = 0;
  for (const ch of str) w += /\d/.test(ch) ? cell : ctx.measureText(ch).width;
  return w;
}

function formatCount(n) {
  return n.toLocaleString("en-US");
}

// Tabular digits: every digit sits in a cell as wide as "0", so counters never jitter.
function drawDigits(ctx, str, x, baseline, color) {
  ctx.fillStyle = color;
  ctx.textBaseline = "alphabetic";
  const cell = ctx.measureText("0").width;
  let cx = x;
  for (const ch of str) {
    if (/\d/.test(ch)) {
      const w = ctx.measureText(ch).width;
      ctx.fillText(ch, cx + (cell - w) / 2, baseline);
      cx += cell;
    } else {
      ctx.fillText(ch, cx, baseline);
      cx += ctx.measureText(ch).width;
    }
  }
}

function statusChip(ctx, status, x, y, size, centered = false) {
  ctx.font = `700 ${size}px ${SANS}`;
  ctx.letterSpacing = "1.5px";
  const label = status.toUpperCase();
  const w = ctx.measureText(label).width + size * 1.4;
  if (centered) x -= w / 2;
  const h = size * 1.9;
  roundRectPath(ctx, x, y, w, h, h / 2);
  ctx.strokeStyle = status === "sourced" ? PAL.sourced : PAL.calculated;
  ctx.lineWidth = 1.5;
  ctx.stroke();
  ctx.fillStyle = status === "sourced" ? PAL.sourced : PAL.calculated;
  ctx.textBaseline = "middle";
  ctx.fillText(label, x + size * 0.7, y + h / 2 + 1);
  ctx.letterSpacing = "0px";
  ctx.textBaseline = "alphabetic";
}

function drawEnd(ctx, scene, t) {
  const E = FMT.end;
  ctx.fillStyle = PAL.paper;
  ctx.fillRect(0, 0, W, H);
  const zoom = progress(scene.camera.scale, t);
  const s = scene.camera.scale.from + (scene.camera.scale.to - scene.camera.scale.from) * zoom;
  ctx.save();
  ctx.translate(W / 2, H / 2); ctx.scale(s, s); ctx.translate(-W / 2, -H / 2);
  text(ctx, "COVERAGE SOURCES", E.overline.x, E.overline.y + E.overline.size, { size: E.overline.size, weight: 700, color: PAL.accentOnPaper, spacing: 3 });
  const L = E.list;
  const n = TL.items.length;
  const cols = n > 3 ? L.columns : 1;
  const perCol = Math.ceil(n / cols);
  const colW = (L.w - (cols - 1) * 56) / cols;
  const row = Math.min(L.row, L.h / Math.max(1, perCol + (TL.facts.more ? 0.5 : 0)));
  TL.items.forEach((it, i) => {
    const p = staggerProgress(scene.listReveal, i, t);
    if (p <= 0) return;
    const x = L.x + Math.floor(i / perCol) * (colW + 56), y = L.y + (i % perCol) * row;
    ctx.save();
    ctx.globalAlpha *= p;
    ctx.translate(0, (1 - p) * 14);
    const num = String(i + 1).padStart(2, "0");
    text(ctx, num, x, y + L.size * 1.0, { size: L.size, weight: 700, color: PAL.accentOnPaper });
    const indent = L.size * 2.1;
    const head = [it.outlet, it.date_label || "date not supplied"].join("  \u00b7  ");
    const hl = fitText(ctx, head, { w: colW - indent, h: L.size * 1.25 }, { size: L.size, min: 18, lines: 1, weight: 700 });
    drawLines(ctx, hl, x + indent, y, PAL.ink, null, t, "left", colW - indent);
    const ul = fitText(ctx, it.short_url, { w: colW - indent, h: L.small * 1.25 }, { size: L.small, min: 16, lines: 1, weight: 400 });
    drawLines(ctx, ul, x + indent, y + L.size * 1.34, PAL.inkMuted, null, t, "left", colW - indent);
    ctx.restore();
  });
  if (TL.facts.more) {
    const p = staggerProgress(scene.listReveal, n, t);
    ctx.save(); ctx.globalAlpha *= p;
    const ml = fitText(ctx, TL.facts.more, { w: L.w, h: L.small * 1.3 }, { size: L.small, min: 16, lines: 1, weight: 500 });
    drawLines(ctx, ml, L.x, L.y + perCol * row + 4, PAL.inkMuted, null, t, "left", L.w);
    ctx.restore();
  }
  const D = E.disclaimer;
  const dl = fitText(ctx, scene.disclaimer, D, { size: D.size, min: 15, lines: 3, weight: 400, lineHeight: 1.35 });
  drawLines(ctx, dl, D.x, D.y, PAL.inkMuted, null, t, "left", D.w);
  drawBrand(ctx, E.brand, PAL.ink, 1);
  ctx.restore();
}

// ---------------------------------------------------------------------------
// Frame
// ---------------------------------------------------------------------------

function drawFrame(ctx, t) {
  ctx.setTransform(1, 0, 0, 1, 0, 0);
  ctx.globalAlpha = 1;
  ctx.globalCompositeOperation = "source-over";
  ctx.filter = "none";
  ctx.fillStyle = PAL.stage;
  ctx.fillRect(0, 0, W, H);
  for (const st of scenesAt(TL, t)) {
    ctx.save();
    ctx.globalAlpha = st.alpha;
    ctx.translate(W / 2 + st.x, H / 2);
    ctx.scale(st.scale, st.scale);
    ctx.translate(-W / 2, -H / 2);
    // a scene never paints outside its own frame (the wall's marquee would spill over the
    // outgoing shot during the push)
    ctx.beginPath();
    ctx.rect(0, 0, W, H);
    ctx.clip();
    const s = st.scene;
    if (s.kind === "opening") drawOpening(ctx, s, t);
    else if (s.kind === "article") drawArticle(ctx, s, t);
    else if (s.kind === "wall") drawWall(ctx, s, t);
    else if (s.kind === "end") drawEnd(ctx, s, t);
    if (st.dim > 0) { ctx.fillStyle = rgba(PAL.stage, st.dim); ctx.fillRect(-W, -H, W * 3, H * 3); }
    ctx.restore();
  }
  // a thin progress bar runs the whole reel
  const P = TL.progress;
  const onPaper = scenesAt(TL, t).some((st) => st.scene.kind === "end" && st.alpha > 0.5);
  ctx.fillStyle = onPaper ? "rgba(0, 0, 0, 0.08)" : "rgba(255, 255, 255, 0.12)";
  ctx.fillRect(0, P.y, W, P.h);
  ctx.fillStyle = PAL.accent;
  ctx.fillRect(0, P.y, W * Math.min(1, t / TL.duration), P.h);
}

// ---------------------------------------------------------------------------
// Player (real-time preview with a scrub bar) and the export API used by render.mjs
// ---------------------------------------------------------------------------

const canvas = document.getElementById("reel");
const ctx = canvas.getContext("2d", { alpha: false });
const playButton = document.getElementById("play");
const scrub = document.getElementById("scrub");
const clock = document.getElementById("clock");
const exporting = new URLSearchParams(location.search).has("export");
let playing = false, origin = 0, current = 0;

function show(t) {
  current = Math.max(0, Math.min(TL.duration - 1 / TL.fps, t));
  drawFrame(ctx, current);
  scrub.value = String(Math.round(current * TL.fps));
  clock.textContent = `${current.toFixed(1)} / ${TL.duration.toFixed(1)} s`;
}

function tick(now) {
  if (!playing) return;
  let t = (now - origin) / 1000;
  if (t >= TL.duration) { origin = now; t = 0; }
  show(t);
  requestAnimationFrame(tick);
}

function play() { playing = true; origin = performance.now() - current * 1000; playButton.textContent = "Pause"; requestAnimationFrame(tick); }
function pause() { playing = false; playButton.textContent = "Play"; }

const ready = loadImages().then(() => {
  scrub.max = String(TL.frames - 1);
  show(0);
  playButton.addEventListener("click", () => (playing ? pause() : play()));
  scrub.addEventListener("input", () => { pause(); show(Number(scrub.value) / TL.fps); });
  window.addEventListener("keydown", (e) => { if (e.code === "Space") { e.preventDefault(); playing ? pause() : play(); } });
  if (!exporting) play();
  return true;
});

let encoder = null, chunks = [], decoderMeta = null, encodeError = null;

function toBase64(bytes) {
  let s = "";
  for (let i = 0; i < bytes.length; i += 0x8000) s += String.fromCharCode.apply(null, bytes.subarray(i, i + 0x8000));
  return btoa(s);
}

window.reel = {
  ready,
  info: { duration: TL.duration, frames: TL.frames, fps: TL.fps, width: W, height: H },
  pause,
  drawAt(t) { pause(); show(t); return true; },
  // Configure H.264 encoding. Prefers the software encoder (byte-identical output); returns
  // which acceleration was used, or null when this browser cannot encode H.264 at all.
  async startEncoder({ codec, bitrate }) {
    const base = { codec, width: W, height: H, framerate: TL.fps, bitrate, bitrateMode: "variable", latencyMode: "quality", avc: { format: "avc" } };
    if (typeof VideoEncoder === "undefined") return null;
    for (const hardwareAcceleration of ["prefer-software", "no-preference"]) {
      const config = { ...base, hardwareAcceleration };
      let ok = false;
      try { ok = (await VideoEncoder.isConfigSupported(config)).supported; } catch { ok = false; }
      if (!ok) continue;
      chunks = []; decoderMeta = null; encodeError = null;
      encoder = new VideoEncoder({
        output: (chunk, meta) => {
          const bytes = new Uint8Array(chunk.byteLength);
          chunk.copyTo(bytes);
          chunks.push({ bytes, key: chunk.type === "key", ts: chunk.timestamp });
          if (meta && meta.decoderConfig && meta.decoderConfig.description && !decoderMeta) {
            const d = meta.decoderConfig.description;
            decoderMeta = { avcC: toBase64(new Uint8Array(d.buffer ? d.buffer.slice(d.byteOffset, d.byteOffset + d.byteLength) : d)) };
          }
        },
        error: (e) => { encodeError = String(e); },
      });
      encoder.configure(config);
      return hardwareAcceleration;
    }
    return null;
  },
  async encodeRange(from, to, keyEvery) {
    for (let i = from; i < to; i++) {
      drawFrame(ctx, i / TL.fps);
      const frame = new VideoFrame(canvas, { timestamp: Math.round((i * 1e6) / TL.fps), duration: Math.round(1e6 / TL.fps) });
      while (encoder.encodeQueueSize > 8) await new Promise((r) => setTimeout(r, 0));
      encoder.encode(frame, { keyFrame: i % keyEvery === 0 });
      frame.close();
      if (encodeError) throw new Error(encodeError);
    }
    return true;
  },
  // Hand finished chunks to Node in batches, keeping page memory flat.
  async drain(final) {
    if (final) { await encoder.flush(); encoder.close(); }
    if (encodeError) throw new Error(encodeError);
    const out = chunks;
    chunks = [];
    const total = out.reduce((n, c) => n + c.bytes.length, 0);
    const joined = new Uint8Array(total);
    let at = 0;
    for (const c of out) { joined.set(c.bytes, at); at += c.bytes.length; }
    return { data: toBase64(joined), sizes: out.map((c) => c.bytes.length), keys: out.map((c) => c.key), ts: out.map((c) => c.ts), meta: decoderMeta };
  },
  // Contact sheet of key frames, drawn in the page (no image tools).
  storyboard(times, columns, tileWidth) {
    const tileH = Math.round(tileWidth * (H / W));
    const label = 30;
    const rows = Math.ceil(times.length / columns);
    const sheet = document.createElement("canvas");
    sheet.width = columns * tileWidth + (columns + 1) * 12;
    sheet.height = rows * (tileH + label) + (rows + 1) * 12;
    const g = sheet.getContext("2d");
    g.fillStyle = "#0d0d0f";
    g.fillRect(0, 0, sheet.width, sheet.height);
    times.forEach((t, i) => {
      drawFrame(ctx, t);
      const x = 12 + (i % columns) * (tileWidth + 12), y = 12 + Math.floor(i / columns) * (tileH + label + 12);
      g.drawImage(canvas, x, y, tileWidth, tileH);
      g.fillStyle = "#cfd3da";
      g.font = `500 18px ${SANS}`;
      g.fillText(`${t.toFixed(2)} s`, x, y + tileH + 22);
    });
    show(current);
    return sheet.toDataURL("image/png");
  },
  // Layout facts the renderer checks before exporting: the thumbnail title must fit whole.
  layoutCheck() {
    const layout = titleLayout(ctx);
    return { title: { lines: layout.lines, size: layout.size, truncated: layout.lines.join(" ") !== TL.brand.report_title.split(/\s+/).filter(Boolean).join(" ") } };
  },
  frameAt(t) {
    drawFrame(ctx, t);
    const url = canvas.toDataURL("image/png");
    show(current);
    return url;
  },
};
