import assert from "node:assert/strict";
import { existsSync, mkdtempSync, readFileSync, writeFileSync, cpSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";
import { spawnSync } from "node:child_process";
import { createHash } from "node:crypto";
import { test } from "node:test";

import {
  InputError,
  applyPageMetadata,
  chooseReelRows,
  deriveMetrics,
  findBrowser,
  inlineModule,
  parseCsv,
  reelPalette,
  contrast,
  renderArtifacts,
  renderVideo,
  scriptJson,
  validateBrand,
  validateClips,
  validateCoverage,
  validateSidecar,
} from "../../skills/coverage-reel/render.mjs";
import { muxMp4, readMp4Summary } from "../../skills/coverage-reel/mp4mux.mjs";
import {
  EASE_NAMES,
  EASINGS,
  FORMATS,
  FORMAT_NAMES,
  TIMING,
  buildTimeline,
  cameraAt,
  ease,
  lintTimeline,
  openingState,
  staggerProgress,
  planCamera,
  lengthBudget,
  rhythm,
  scrollDuration,
  shortUrl,
  toScreen,
  unionRect,
} from "../../skills/coverage-reel/timeline.mjs";
import {
  detectBlock,
  isAmbiguousTerm,
  mergeLineRects,
  parseMentionTerms,
  pickPreviewMode,
  planStrips,
  rankMentions,
  toAssetRects,
} from "../../skills/press-clip/clip.mjs";
import { makeFixture } from "./make-fixture.mjs";

const fixtureRoot = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(fixtureRoot, "../..");

function fixtureInputs() {
  const coverageRows = validateCoverage(parseCsv(readFileSync(join(fixtureRoot, "coverage.csv"), "utf8")));
  const brand = validateBrand(JSON.parse(readFileSync(join(fixtureRoot, "brand.json"), "utf8")), fixtureRoot);
  const clips = { raw: JSON.parse(readFileSync(join(fixtureRoot, "clips.json"), "utf8")), baseDir: fixtureRoot };
  return { coverageRows, brand, clips };
}

function fixtureRender(options = {}) {
  const inputs = fixtureInputs();
  return renderArtifacts({ ...inputs, outDir: mkdtempSync(join(tmpdir(), "coverage-reel-")), ...options });
}

function sidecar(slug) {
  return JSON.parse(readFileSync(join(fixtureRoot, "clips", slug, "clip.json"), "utf8"));
}

// --- inputs -----------------------------------------------------------------------------

test("CSV parsing preserves quoted commas and newlines", () => {
  const rows = parseCsv('source_url,headline,note\r\nhttps://example.test/a,"A, B","line one\nline two"\r\n');
  assert.deepEqual(rows, [{ source_url: "https://example.test/a", headline: "A, B", note: "line one\nline two" }]);
});

test("coverage validation rejects unsafe URLs, unsourced reach and unknown columns", () => {
  assert.throws(() => validateCoverage([{ source_url: "javascript:alert(1)" }]), (error) => error instanceof InputError && /http or https/.test(error.message));
  assert.throws(() => validateCoverage([{ source_url: "https://example.test/a", reach: "1200" }]), /reach requires reach_source_url/);
  assert.throws(() => validateCoverage([{ source_url: "https://example.test/a", outlet_rank: "1" }]), /unsupported column\(s\): outlet_rank/);
  assert.throws(() => validateCoverage([{ source_url: "https://example.test/a", published_at: "2026-02-30" }]), /real calendar date/);
});

test("brand validation keeps mention terms and rejects remote logos, unsafe colours and active SVG", () => {
  assert.deepEqual(validateBrand({ name: "Example", mention_terms: ["Example Co", "Example Co"] }).mention_terms, ["Example Co"]);
  assert.deepEqual(validateBrand({ name: "Example" }).mention_terms, ["Example"]);
  assert.throws(() => validateBrand({ name: "Example", logo_path: "https://example.test/logo.svg" }), /local path/);
  assert.throws(() => validateBrand({ name: "Example", accent_color: "red; background:url(x)" }), /hex color/);
  const work = mkdtempSync(join(tmpdir(), "coverage-reel-svg-"));
  writeFileSync(join(work, "unsafe.svg"), '<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>');
  assert.throws(() => validateBrand({ name: "Example", logo_path: join(work, "unsafe.svg") }), /active or remote SVG content/);
});

test("metrics separate calculations, sourced values, partials, and unavailable data", () => {
  const rows = validateCoverage(parseCsv([
    "source_url,outlet,published_at,sentiment,reach,reach_source_url",
    "https://example.test/a,Daily Example,2026-09-01,positive,1200,https://metrics.example.test/a",
    "https://example.test/b,Daily Example,,, ,",
  ].join("\n")));
  const metrics = deriveMetrics(rows);
  assert.equal(metrics.coverage.status, "calculated");
  assert.equal(metrics.outlets.value, "1");
  assert.equal(metrics.date_range.detail, "From 1/2 supplied publish date; partial");
  assert.equal(metrics.reach.status, "sourced");
  assert.equal(metrics.reach.label, "Partial reported reach");
  assert.match(metrics.reach.detail, /not total reach/);
  const missing = deriveMetrics(validateCoverage([{ source_url: "https://example.test/c" }]));
  for (const key of ["outlets", "date_range", "reach", "sentiment"]) assert.equal(missing[key].status, "unavailable");
});

// --- review gate and sidecars ---------------------------------------------------------------

test("clips.json enforces the press-clip review gate", () => {
  const { coverageRows, clips } = fixtureInputs();
  const mutate = (fn) => { const raw = structuredClone(clips.raw); fn(raw); return raw; };
  assert.throws(() => validateClips(mutate((r) => { r.version = 1; }), coverageRows, clips.baseDir), /version must be 2/);
  for (const flag of ["logo_verified", "pdf_reviewed", "mention_verified", "captures_clean"]) {
    assert.throws(() => validateClips(mutate((r) => { r.clips[0].review[flag] = false; }), coverageRows, clips.baseDir), new RegExp(`${flag} must be true`));
  }
  assert.throws(() => validateClips(mutate((r) => { r.clips[0].review.verdict = "has_junk"; }), coverageRows, clips.baseDir), /verdict must be clean/);
  assert.throws(() => validateClips(mutate((r) => { r.clips[1].review.scope = "whole"; }), coverageRows, clips.baseDir), /capture is section/);
  assert.throws(() => validateClips(mutate((r) => { r.clips.pop(); }), coverageRows, clips.baseDir), /missing 1 CSV source/);
  assert.throws(() => validateClips(mutate((r) => { r.clips[3].omitted.reason = "too_hard"; }), coverageRows, clips.baseDir), /omitted.reason/);
  assert.throws(() => validateClips(mutate((r) => { r.clips[0].omitted = { reason: "blocked" }; }), coverageRows, clips.baseDir), /exactly one of/);
});

test("sidecar validation rejects captures that would make the reel lie", () => {
  const dir = join(fixtureRoot, "clips", "northstar-daily");
  const url = "https://example.test/northstar-launch";
  const good = sidecar("northstar-daily");
  assert.equal(validateSidecar(good, dir, url).mention.sentence_bars.length, 3);
  const bad = (fn, pattern) => { const raw = structuredClone(good); fn(raw); assert.throws(() => validateSidecar(raw, dir, url), pattern); };
  bad((r) => { r.source_url = "https://example.test/other"; }, /does not match the CSV row/);
  bad((r) => { r.assets[0].width += 2; }, /says .* but top.png is/);
  bad((r) => { r.assets[0].file = "../top.png"; }, /PNG file name inside the clip folder/);
  bad((r) => { r.warnings = [{ code: "layout_shifted", detail: "moved" }]; }, /layout_shifted/);
  bad((r) => { r.warnings = [{ code: "blocked", detail: "challenge" }]; }, /blocked/);
  bad((r) => { r.mentions[0].chosen = false; }, /exactly one chosen mention/);
  bad((r) => { r.mentions[0].rects.sentence[0].y = 99999; }, /outside the captured strips/);
  bad((r) => { r.version = 2; }, /version must be 1/);

  // a capture taller than Chrome's limit is refused even when its header is consistent
  const work = mkdtempSync(join(tmpdir(), "coverage-reel-tall-"));
  cpSync(dir, work, { recursive: true });
  const header = Buffer.from(readFileSync(join(work, "strip-1.png")));
  header.writeUInt32BE(17000, 20);
  writeFileSync(join(work, "strip-1.png"), header);
  const tall = structuredClone(good);
  tall.assets.find((a) => a.file === "strip-1.png").height = 17000;
  assert.throws(() => validateSidecar(tall, work, url), /must stay under 16000 px/);
});

test("metadata precedence: CSV wins, page fills are labelled, publish dates are never filled", () => {
  const { coverageRows, clips } = fixtureInputs();
  const rows = validateClips(clips.raw, coverageRows, clips.baseDir).map(applyPageMetadata);
  assert.equal(rows[0].headline, "Aster Labs opens a safer robotics workspace");
  assert.equal(rows[0].headline_source, "csv");
  assert.equal(rows[1].headline, "Inside the small teams testing warehouse robots");
  assert.equal(rows[1].headline_source, "page:visible");
  assert.equal(rows[2].published_at, "");
  assert.deepEqual(rows[2].page_published_at, { value: "2026-09-18T07:30:00Z", source: "json-ld" });
  const result = fixtureRender();
  const coverage = result.manifest.coverage;
  assert.equal(coverage[1].headline_source, "page:visible");
  assert.equal(coverage[2].published_at, null);
  assert.equal(coverage[2].page_published_at_unconfirmed.value, "2026-09-18T07:30:00Z");
  assert.match(result.dashboard, /read from page/);
  assert.match(result.dashboard, /Date not supplied/);
});

test("blocked rows are left out of the reel and reported, never faked", () => {
  const result = fixtureRender();
  assert.deepEqual(result.reelRows.map((r) => r.source_url), [
    "https://example.test/northstar-launch", "https://example.test/field-notes", "https://example.test/morning-signal",
  ]);
  assert.equal(result.manifest.coverage[3].capture.kind, "omitted");
  assert.match(result.dashboard, /blocked automated capture/);
  assert.match(readFileSync(join(result.destination, "sources.txt"), "utf8"), /ledger-weekly\n?/);
  assert.equal(result.timeline.facts.more, "+1 more in the full coverage list");
  const rows = Array.from({ length: 7 }, (_, i) => ({ position: i + 1, include_in_reel: null, clip: { kind: "clip" } }));
  assert.throws(() => chooseReelRows(rows, 6), /mark at most 6 rows/);
});

// --- timeline, camera and motion lints ------------------------------------------------------

test("the fixture timeline is deterministic in every format and passes every motion lint", () => {
  for (const format of FORMAT_NAMES) {
    const a = fixtureRender({ format });
    const b = fixtureRender({ format });
    assert.deepEqual(a.timeline, b.timeline);
    assert.equal(a.reelDocument, b.reelDocument);
    assert.equal(JSON.stringify(a.manifest), JSON.stringify(b.manifest));
    assert.deepEqual(lintTimeline(a.timeline), [], `${format} lints`);
    assert.equal(a.timeline.width, FORMATS[format].width);
    assert.equal(a.timeline.frames, Math.round(a.timeline.duration * 30));
  }
});

function lintRules(timeline) {
  return lintTimeline(timeline).map((p) => p.rule);
}

test("motion lints catch each anti-pattern", () => {
  const base = fixtureRender().timeline;
  const clone = () => structuredClone(base);

  let t = clone(); t.scenes[1].mastheadIn.ease = "ease-in-out";
  assert.ok(lintRules(t).includes("named-easing"));
  t = clone(); t.scenes[1].headlineReveal.ease = "LINEAR";
  assert.ok(lintRules(t).includes("linear-only-for-drift"));
  t = clone(); t.transitions[1].kind = "wipe";
  assert.ok(lintRules(t).includes("max-two-transition-types"));
  t = clone(); t.transitions[0].kind = "dissolve";
  assert.ok(lintRules(t).includes("dissolve-only-into-end-card"));
  t = clone(); t.scenes.filter((s) => s.kind === "article").forEach((s) => { s.exitAt = s.start + 6; });
  assert.ok(lintRules(t).includes("varied-block-lengths"));
  t = clone();
  for (const s of t.scenes) {
    if (s.camera && s.camera.scale) s.camera.scale.to = s.camera.scale.from;
    if (s.kind === "article") { s.camera.driftRate = 0; s.camera.travel = null; s.camera.zoom = null; }
  }
  t.transitions = t.transitions.map((tr) => ({ ...tr, dur: 0.001 }));
  assert.ok(lintRules(t).includes("camera-never-static"));
  t = clone(); t.scenes[1].camera.s1 = 3;
  assert.ok(lintRules(t).includes("no-upsampling"));
  t = clone(); const mention = t.texts.find((r) => r.role === "mention"); mention.until = mention.from + 1;
  assert.ok(lintRules(t).includes("reading-speed"));
  t = clone(); t.scenes[1].camera.a1 = [t.width + 400, t.height + 400];
  assert.ok(lintRules(t).includes("mention-in-safe-area"));
  t = clone(); t.scenes[0].start = 0.5;
  assert.ok(lintRules(t).includes("no-fade-from-black"));

  const saved = EASINGS.MOVE;
  try {
    EASINGS.MOVE = [0.3, -0.6, 0.3, 1.8];
    assert.ok(lintRules(clone()).includes("no-overshoot"));
  } finally {
    EASINGS.MOVE = saved;
  }
  const slot = FORMATS.landscape.article.headline;
  const savedX = slot.x;
  try {
    slot.x = 1900;
    assert.ok(lintRules(clone()).includes("safe-area"));
  } finally {
    slot.x = savedX;
  }
  const micro = FORMATS.landscape.article.micro;
  const savedY = micro.y, savedMx = micro.x;
  try {
    Object.assign(micro, { x: 200, y: 500 });   // a label drawn over the page card
    assert.ok(lintRules(clone()).includes("text-off-card"));
  } finally {
    Object.assign(micro, { x: savedMx, y: savedY });
  }
});

test("frame 0 is the thumbnail: the whole report title is up, the rest animates in", () => {
  for (const format of FORMAT_NAMES) {
    const tl = fixtureRender({ format }).timeline;
    const opening = tl.scenes.find((s) => s.kind === "opening");
    const at0 = openingState(opening, 0);
    assert.equal(at0.title, 1, `${format}: title fully visible at t=0`);
    assert.equal(tl.texts.find((r) => r.role === "title").from, 0);
    assert.ok(at0.backdrop < 1 && at0.facts(0) < 1, "backdrop and facts animate in");
    assert.equal(staggerProgress(opening.chipReveal, 0, 0), 0, "mastheads arrive after frame 0");
    assert.equal(openingState(opening, opening.end - 0.5).backdrop, 1);
  }
  const base = fixtureRender().timeline;
  let t = structuredClone(base);
  t.scenes[0].titleIn = { start: 0.04, dur: 0.5, ease: "MOVE" };   // a title entrance hides it on frame 0
  assert.ok(lintRules(t).includes("thumbnail-title"));
  t = structuredClone(base);
  t.texts.find((r) => r.role === "title").from = 0.9;
  assert.ok(lintRules(t).includes("thumbnail-title"));
});

test("length lints: 30-40 s for 4-6 articles, and no article block drags", () => {
  assert.equal(lengthBudget(4), 40);
  assert.equal(lengthBudget(6), 40);
  assert.equal(lengthBudget(8), 50);
  const base = fixtureRender().timeline;
  assert.ok(base.duration <= lengthBudget(3));
  // stretch the end card past the budget
  let t = structuredClone(base);
  t.duration = lengthBudget(3) + 5;
  t.frames = Math.round(t.duration * t.fps);
  t.scenes.at(-1).end = t.duration;
  assert.ok(lintRules(t).includes("reel-length"));
  // an article block that holds too long
  t = structuredClone(base);
  t.scenes[1].exitAt = t.scenes[1].start + TIMING.maxBlock + 1;
  assert.ok(lintRules(t).includes("block-length"));
  // every block in the fixture stays inside the 5-6 s target band's ceiling
  for (const s of base.scenes.filter((x) => x.kind === "article")) assert.ok(s.exitAt - s.start <= TIMING.maxBlock);
});

test("easing palette is monotonic with no overshoot, and pins its ends", () => {
  for (const name of EASE_NAMES) {
    assert.equal(ease(name, 0), 0);
    assert.equal(ease(name, 1), 1);
    let prev = 0;
    for (let i = 1; i <= 100; i++) {
      const v = ease(name, i / 100);
      assert.ok(v >= prev - 1e-9 && v <= 1 + 1e-9, `${name} at ${i}`);
      prev = v;
    }
  }
  assert.throws(() => ease("BOUNCE", 0.5), /unknown easing/);
});

test("camera math frames the mention from its rectangles without upsampling", () => {
  const card = { x: 100, y: 80, w: 900, h: 900 };
  const focus = { x: 140, y: 300, w: 820, h: 420 };
  const capture = {
    scale: 2, column: { x: 0, w: 800 }, top_y: 0, bottom_y: 4000,
    mention: { sentence_bars: [{ x: 40, y: 2000, w: 600, h: 24 }, { x: 40, y: 2028, w: 300, h: 24 }], paragraph: { x: 40, y: 1990, w: 640, h: 90 } },
  };
  const cam = planCamera({ capture, card, focus, sceneSeconds: 12 });
  assert.ok(cam.s1 / cam.s0 >= TIMING.minZoom - 1e-9, "the zoom always reads as a move");
  assert.ok(cam.s1 * (1 + TIMING.driftRate * 12) <= 2 + 1e-9, "never past the capture's pixels");
  // at the end of the move the sentence sits at the focus anchor
  const c = cameraAt({ ...cam, driftStart: 0, driftRate: 0, travel: { start: 0, dur: 1, ease: "MOVE" }, zoom: { start: 0, dur: 1, ease: "ARRIVE" } }, 5);
  const box = toScreen(c, unionRect(capture.mention.sentence_bars));
  assert.ok(box.x >= focus.x - 1 && box.x + box.w <= focus.x + focus.w + 1);
  assert.ok(Math.abs(box.y + box.h / 2 - (focus.y + focus.h / 2)) < 1);
  assert.ok(Math.abs(scrollDuration(2 * 900, 900) / scrollDuration(900, 900) - Math.SQRT2) < 0.01, "square-root distance timing");
  assert.equal(scrollDuration(50 * 900, 900), 1.3, "long scrolls are capped");
});

test("block rhythm varies and the first block teaches the pattern", () => {
  const out = rhythm([6, 6, 6, 6]);
  assert.ok((Math.max(...out) - Math.min(...out)) / Math.max(...out) >= 0.1);
  assert.equal(Math.max(...out), out[0]);
  for (let i = 1; i < out.length; i++) assert.ok(Math.abs(out[i] - out[i - 1]) > 0.01);
});

// --- MP4 muxer --------------------------------------------------------------------------------

function parseBoxes(buf, start = 0, end = buf.length) {
  const boxes = [];
  for (let at = start; at < end;) {
    const size = buf.readUInt32BE(at);
    boxes.push({ type: buf.toString("latin1", at + 4, at + 8), start: at, body: at + 8, end: at + size });
    at += size;
  }
  return boxes;
}

function child(buf, box, type, skip = 0) {
  return parseBoxes(buf, box.body + skip, box.end).find((b) => b.type === type);
}

test("the MP4 muxer writes a fast-start file with the right tables", () => {
  const avcC = Buffer.from([1, 0x64, 0, 0x28, 0xff, 0xe1, 0, 4, 0x67, 0x64, 0, 0x28, 1, 0, 4, 0x68, 0xee, 0x3c, 0x80]);
  const samples = [
    { data: Buffer.from([0, 0, 0, 2, 0x65, 1]), key: true },
    { data: Buffer.from([0, 0, 0, 3, 0x41, 2, 3]), key: false },
    { data: Buffer.from([0, 0, 0, 1, 0x41]), key: false },
    { data: Buffer.from([0, 0, 0, 2, 0x65, 9]), key: true },
  ];
  const mp4 = muxMp4({ width: 1080, height: 1920, fps: 30, avcC, samples });
  const top = parseBoxes(mp4);
  assert.deepEqual(top.map((b) => b.type), ["ftyp", "moov", "mdat"]);
  const moov = top[1], mdat = top[2];
  const trak = child(mp4, moov, "trak");
  const tkhd = child(mp4, trak, "tkhd");
  assert.equal(mp4.readUInt32BE(tkhd.end - 8) >>> 16, 1080);
  assert.equal(mp4.readUInt32BE(tkhd.end - 4) >>> 16, 1920);
  const mdia = child(mp4, trak, "mdia");
  const mdhd = child(mp4, mdia, "mdhd");
  assert.equal(mp4.readUInt32BE(mdhd.body + 12), 90000);
  assert.equal(mp4.readUInt32BE(mdhd.body + 16), 4 * 3000);
  const stbl = child(mp4, child(mp4, mdia, "minf"), "stbl");
  const stsd = child(mp4, stbl, "stsd");
  const avc1 = child(mp4, stsd, "avc1", 8);
  assert.equal(mp4.readUInt16BE(avc1.body + 24), 1080);
  const avcBox = child(mp4, avc1, "avcC", 78);
  assert.deepEqual(mp4.subarray(avcBox.body, avcBox.end), avcC);
  assert.ok(child(mp4, avc1, "colr", 78), "colour box present");
  const stts = child(mp4, stbl, "stts");
  assert.deepEqual([mp4.readUInt32BE(stts.body + 4), mp4.readUInt32BE(stts.body + 8), mp4.readUInt32BE(stts.body + 12)], [1, 4, 3000]);
  const stss = child(mp4, stbl, "stss");
  assert.deepEqual([mp4.readUInt32BE(stss.body + 4), mp4.readUInt32BE(stss.body + 8), mp4.readUInt32BE(stss.body + 12)], [2, 1, 4]);
  const stsz = child(mp4, stbl, "stsz");
  assert.deepEqual([0, 1, 2, 3].map((i) => mp4.readUInt32BE(stsz.body + 12 + 4 * i)), [6, 7, 5, 6]);
  const stco = child(mp4, stbl, "stco");
  assert.equal(mp4.readUInt32BE(stco.body + 8), mdat.body, "chunk offset points at the mdat payload");
  assert.deepEqual(mp4.subarray(mdat.body, mdat.end), Buffer.concat(samples.map((s) => s.data)));
  assert.ok(!parseBoxes(mp4, moov.body, moov.end).some((b) => b.type === "edts"), "no edit lists");
  assert.deepEqual(
    (({ width, height, fps, frames, keyframes, faststart }) => ({ width, height, fps, frames, keyframes, faststart }))(readMp4Summary(mp4)),
    { width: 1080, height: 1920, fps: 30, frames: 4, keyframes: 2, faststart: true },
  );
  assert.throws(() => muxMp4({ width: 1081, height: 1920, fps: 30, avcC, samples }), /even/);
  assert.throws(() => muxMp4({ width: 1080, height: 1920, fps: 30, avcC, samples: samples.slice(1) }), /keyframe/);
});

// --- escaping and safety ------------------------------------------------------------------------

test("supplied text is escaped in the dashboard and inert in reel.html", () => {
  const inputs = fixtureInputs();
  inputs.coverageRows[0] = { ...inputs.coverageRows[0], headline: '<script>alert("x")</script>', note: '<img src=x onerror="alert(1)">', outlet: "</script><b>Outlet" };
  const result = renderArtifacts({ ...inputs, outDir: mkdtempSync(join(tmpdir(), "coverage-reel-esc-")) });
  assert.doesNotMatch(result.dashboard, /<script>alert/);
  assert.match(result.dashboard, /&lt;script&gt;alert\(&quot;x&quot;\)&lt;\/script&gt;/);
  assert.match(result.dashboard, /default-src 'none'/);
  const data = result.reelDocument.match(/<script type="application\/json" id="reel-data">([\s\S]*?)<\/script>/)[1];
  assert.doesNotMatch(data, /<\/script/i);
  assert.equal(JSON.parse(data).timeline.items[0].headline, '<script>alert("x")</script>');
  assert.equal(scriptJson({ a: "</script>\u2028" }), '{"a":"\\u003c/script\\u003e\\u2028"}');
  // the inline player runs only because its hash is in the CSP
  const bundle = result.reelDocument.match(/<script>([\s\S]*)<\/script>\n<\/body>/)[1];
  const hash = createHash("sha256").update(bundle).digest("base64");
  assert.match(result.reelDocument, new RegExp(`script-src 'sha256-${hash.replace(/[+/]/g, "\\$&")}'`));
  assert.doesNotMatch(result.reelDocument, /https?:\/\/(?!example\.test|aster\.example\.test|metrics\.example\.test|www\.w3\.org)/, "no remote resources");
});

test("the shared timeline module inlines into the page as a plain script", () => {
  const source = readFileSync(join(repoRoot, "skills/coverage-reel/timeline.mjs"), "utf8");
  const inlined = inlineModule(source);
  assert.doesNotMatch(inlined, /^export\b/m);
  assert.doesNotThrow(() => new Function(inlined));
  assert.throws(() => inlineModule("export default 1;\n"), /export function/);
});

test("palette keeps text readable and short URLs stay short", () => {
  const p = reelPalette(validateBrand({ name: "X", primary_color: "#d0d0d0", accent_color: "#ffee00", background_color: "#ffffff" }));
  assert.ok(contrast(p.accentOnDark, p.stage) >= 4.5);
  assert.ok(contrast(p.ink, p.paper) >= 7);
  assert.ok(contrast("#ffffff", p.stage) >= 7);
  assert.equal(shortUrl("https://www.example.test/a/very/long/path/that/keeps/going/on-and-on/forever"), "example.test/a/very/long/path/that/keep\u2026");
});

// --- press-clip helpers (pure) -------------------------------------------------------------------

test("press-clip mention terms, ranking and the ambiguous-word rule", () => {
  assert.deepEqual(parseMentionTerms(["Nothing, Nothing Phone (4a) Pro", "Phone (4a) Pro"]), ["Nothing Phone (4a) Pro", "Phone (4a) Pro", "Nothing"]);
  assert.ok(isAmbiguousTerm("Nothing"));
  assert.ok(!isAmbiguousTerm("OnePlus") && !isAmbiguousTerm("Nothing Phone") && !isAmbiguousTerm("NASA"));
  const body = { block: "P", paragraphIndex: 2, sentence: "x".repeat(90), term: "Acme Robotics" };
  const ranked = rankMentions([
    { ...body, sentence: "Nothing is certain.", term: "Nothing", ambiguous: true, sentenceInitial: true },
    { ...body, inLink: true, sameSiteLink: true, term: "Acme" },
    { ...body, linkDense: true },
    { ...body, inAside: true },
    { ...body, inHeadline: true, block: "H1" },
    body,
    { ...body },
  ]);
  assert.equal(ranked[0].order, 5, "a body sentence with the full name wins");
  assert.ok(!ranked.some((h) => h.sentence === "Nothing is certain."), "sentence-initial common word rejected");
  assert.equal(ranked.filter((h) => h.sentence === body.sentence && h.paragraphIndex === 2).length, 1, "one entry per sentence");
  const scoped = rankMentions([{ ...body, sectionScoped: true, inSection: false, paragraphIndex: 1 }, { ...body, sentence: "y".repeat(90), sectionScoped: true, inSection: true, paragraphIndex: 9 }]);
  assert.equal(scoped[0].sentence, "y".repeat(90), "inside the requested section wins");
});

test("press-clip geometry: line merging, tiles under the capture limit, asset rects", () => {
  assert.deepEqual(mergeLineRects([{ x: 10, y: 100, w: 50, h: 20 }, { x: 60, y: 101, w: 40, h: 19 }, { x: 10, y: 124, w: 80, h: 20 }]), [{ x: 10, y: 100, w: 90, h: 20 }, { x: 10, y: 124, w: 80, h: 20 }]);
  for (const scale of [2, 3]) {
    const tiles = planStrips(0, 20000, scale);
    assert.ok(tiles.every((t) => t.h * scale <= 8000));
    assert.equal(tiles.reduce((n, t) => n + t.h, 0), 20000);
  }
  const asset = { rect: { x: 20, y: 1000, w: 500, h: 400 }, scale: 2 };
  assert.deepEqual(toAssetRects([{ x: 30, y: 1100, w: 100, h: 20 }, { x: 30, y: 5000, w: 10, h: 10 }], asset), [{ x: 20, y: 200, w: 200, h: 40 }]);
});

test("press-clip detects bot walls and never repeats tall previews", () => {
  assert.equal(detectBlock({ status: 403, title: "Just a moment...", textLength: 120, hasH1: false }).blocked, true);
  assert.equal(detectBlock({ status: 200, title: "Attention Required! | Cloudflare", textLength: 300, hasH1: false }).blocked, true);
  assert.equal(detectBlock({ status: 503, title: "", textLength: 40, hasH1: false }).blocked, true);
  assert.equal(detectBlock({ status: 200, title: "Nothing Phone review", textLength: 9000, hasH1: true }).blocked, false);
  assert.equal(pickPreviewMode(5000, 2).scale, "device");
  assert.equal(pickPreviewMode(12000, 2).scale, "css");
  assert.equal(pickPreviewMode(20000, 2).clipHeight, 16000);
});

test("the committed fixture captures match the generator byte for byte", () => {
  const out = mkdtempSync(join(tmpdir(), "coverage-reel-fixture-"));
  makeFixture(out);
  for (const slug of ["northstar-daily", "field-notes", "morning-signal"]) {
    for (const file of ["clip.json", "top.png", "strip-1.png", "mention-1.png", "masthead.png"]) {
      assert.ok(readFileSync(join(out, slug, file)).equals(readFileSync(join(fixtureRoot, "clips", slug, file))), `${slug}/${file}`);
    }
  }
});

// --- browser integration (skips without Chrome or Edge) -------------------------------------------

const browserPath = (() => { try { return findBrowser(); } catch { return ""; } })();
const hasPlaywright = existsSync(join(repoRoot, "node_modules", "playwright-core"));

test("Chrome renders a short reel to a valid MP4 with no ffmpeg", { skip: !browserPath || !hasPlaywright ? "no Chrome/Edge or playwright-core" : false, timeout: 120000 }, async () => {
  const result = fixtureRender({ format: "vertical" });
  const out = join(result.destination, "highlight.mp4");
  const video = await renderVideo({ reelPath: join(result.destination, "reel.html"), outputPath: out, timeline: result.timeline, maxFrames: 45 });
  assert.deepEqual(
    (({ codec, width, height, fps, frames, faststart }) => ({ codec, width, height, fps, frames, faststart }))(video.summary),
    { codec: "avc1", width: 1080, height: 1920, fps: 30, frames: 45, faststart: true },
  );
  assert.ok(video.summary.keyframes >= 2);
});

test("press-clip --assets captures a local synthetic article", { skip: !browserPath || !hasPlaywright ? "no Chrome/Edge or playwright-core" : false, timeout: 120000 }, () => {
  const work = mkdtempSync(join(tmpdir(), "press-clip-assets-"));
  const url = pathToFileURL(join(fixtureRoot, "synthetic-article.html")).href;
  const run = spawnSync(process.execPath, [join(repoRoot, "skills/press-clip/clip.mjs"), "--url", url, "--out", join(work, "clip.pdf"), "--assets", join(work, "assets"), "--mention", "Aster Labs,Aster", "--chrome", browserPath], { cwd: repoRoot, encoding: "utf8" });
  assert.equal(run.status, 0, run.stderr || run.stdout);
  const raw = JSON.parse(readFileSync(join(work, "assets", "clip.json"), "utf8"));
  assert.equal(raw.source_url, url);
  // the renderer only accepts http(s) sources; check the rest of the sidecar under a stand-in URL
  const checked = validateSidecar({ ...raw, source_url: "https://example.test/synthetic-article" }, join(work, "assets"), "https://example.test/synthetic-article");
  assert.match(checked.mention.sentence, /Aster Labs/);
  assert.doesNotMatch(checked.mention.sentence, /^Aster is/, "the common-word use at a sentence start is not the client");
  assert.equal(checked.page.headline.value, "Aster Labs opens a safer robotics workspace");
  assert.equal(checked.page.byline.value, "Jamie Rivera");
  assert.ok(!JSON.stringify(raw).includes("@example.test"), "author emails are dropped");
  assert.ok(raw.mentions.every((m) => !/Related story/.test(m.sentence)), "link lists are not mentions");
  assert.ok(raw.assets.every((a) => a.height <= 16000));
  // the animated wordmark is frozen on its fullest frame and keeps its own colours on its header plate
  assert.equal(raw.logo.kind, "svg");
  assert.equal(raw.logo.frame, "start");
  assert.equal(raw.logo.plate_color, "rgb(16, 42, 67)");
});
