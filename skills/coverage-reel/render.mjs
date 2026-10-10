#!/usr/bin/env node
// coverage-reel renderer: validated coverage + reviewed press-clip captures -> dashboard,
// manifest, sources list, an offline reel.html player and (optionally) highlight.mp4.
//
// Node standard library only for the dashboard and reel.html. Previews and video use the
// user's Chrome or Edge through playwright-core (the same setup as press-clip); the MP4 is
// encoded inside Chrome with WebCodecs and written by the skill-local mp4mux.mjs. No ffmpeg.

import { createHash } from "node:crypto";
import { execFileSync } from "node:child_process";
import {
  copyFileSync,
  existsSync,
  mkdirSync,
  readFileSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { createRequire } from "node:module";
import {
  basename,
  dirname,
  extname,
  isAbsolute,
  join,
  resolve,
} from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

import { muxMp4, readMp4Summary } from "./mp4mux.mjs";
import {
  FORMAT_NAMES,
  buildTimeline,
  dateLabel,
  dateSpanLabel,
  lintTimeline,
  shortUrl,
} from "./timeline.mjs";

const MODULE_PATH = fileURLToPath(import.meta.url);
const SKILL_DIR = dirname(MODULE_PATH);
const ALLOWED_SENTIMENTS = new Set(["positive", "neutral", "negative", "mixed"]);
const ALLOWED_IMAGE_EXTENSIONS = new Set([".png", ".jpg", ".jpeg", ".webp", ".svg"]);
const MAX_ASSET_DEVICE_PX = 16000;
const PAGE_SOURCES = new Set(["json-ld", "og", "meta", "visible"]);
const OMIT_REASONS = new Set(["blocked", "paywall", "capture_failed", "no_mention", "user_excluded"]);
// Sidecar warnings that make a capture unusable for the reel until the clip is re-run.
const BLOCKING_WARNINGS = new Set(["blocked", "no_mention_found", "layout_shifted", "pick_out_of_range"]);
const DISCLAIMER = "Coverage belongs to its publishers. Mastheads are trademarks, shown as credit, not endorsement.";
const COVERAGE_KEYS = new Set([
  "source_url",
  "outlet",
  "headline",
  "published_at",
  "byline",
  "note",
  "coverage_type",
  "sentiment",
  "reach",
  "reach_source_url",
  "clip_section",
  "include_in_reel",
]);
const BRAND_KEYS = new Set([
  "name",
  "mention_terms",
  "report_title",
  "subtitle",
  "period_label",
  "website",
  "logo_path",
  "primary_color",
  "accent_color",
  "background_color",
  "closing_note",
  "reel_max_items",
]);

export class InputError extends Error {}

function fail(message) {
  throw new InputError(message);
}

export function parseCsv(input) {
  const text = input.replace(/^\uFEFF/, "");
  const records = [];
  let record = [];
  let field = "";
  let quoted = false;

  for (let i = 0; i < text.length; i += 1) {
    const char = text[i];
    if (quoted) {
      if (char === '"') {
        if (text[i + 1] === '"') {
          field += '"';
          i += 1;
        } else {
          quoted = false;
        }
      } else {
        field += char;
      }
      continue;
    }

    if (char === '"' && field === "") {
      quoted = true;
    } else if (char === ",") {
      record.push(field);
      field = "";
    } else if (char === "\n" || char === "\r") {
      if (char === "\r" && text[i + 1] === "\n") i += 1;
      record.push(field);
      if (record.some((value) => value !== "")) records.push(record);
      record = [];
      field = "";
    } else {
      field += char;
    }
  }

  if (quoted) fail("CSV has an unterminated quoted field");
  record.push(field);
  if (record.some((value) => value !== "")) records.push(record);
  if (records.length === 0) fail("CSV is empty");

  const headers = records[0].map((value) => value.trim());
  if (headers.some((value) => value === "")) fail("CSV contains an empty header");
  if (new Set(headers).size !== headers.length) fail("CSV contains duplicate headers");

  return records.slice(1).map((values, index) => {
    if (values.length > headers.length) {
      fail(`CSV row ${index + 2} has more fields than the header`);
    }
    return Object.fromEntries(headers.map((header, fieldIndex) => [header, values[fieldIndex] ?? ""]));
  });
}

function text(value, label, maxLength = 500) {
  if (value === undefined || value === null || value === "") return "";
  if (typeof value !== "string") fail(`${label} must be text`);
  const normalized = value.trim();
  if (normalized.includes("\0")) fail(`${label} contains a null byte`);
  if (normalized.length > maxLength) fail(`${label} exceeds ${maxLength} characters`);
  return normalized;
}

function httpUrl(value, label, { required = false } = {}) {
  const normalized = text(value, label, 4096);
  if (!normalized) {
    if (required) fail(`${label} is required`);
    return "";
  }
  let parsed;
  try {
    parsed = new URL(normalized);
  } catch {
    fail(`${label} must be a valid URL`);
  }
  if (parsed.protocol !== "http:" && parsed.protocol !== "https:") {
    fail(`${label} must use http or https`);
  }
  parsed.hash = "";
  return parsed.href;
}

function isoDate(value, label) {
  const normalized = text(value, label, 10);
  if (!normalized) return "";
  if (!/^\d{4}-\d{2}-\d{2}$/.test(normalized)) fail(`${label} must use YYYY-MM-DD`);
  const parsed = new Date(`${normalized}T00:00:00Z`);
  if (Number.isNaN(parsed.valueOf()) || parsed.toISOString().slice(0, 10) !== normalized) {
    fail(`${label} is not a real calendar date`);
  }
  return normalized;
}

function isoTimestamp(value, label) {
  const normalized = text(value, label, 64);
  if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:\d{2})$/.test(normalized)) {
    fail(`${label} must be an ISO 8601 timestamp with a timezone`);
  }
  if (Number.isNaN(Date.parse(normalized))) fail(`${label} is not a valid timestamp`);
  return normalized;
}

function booleanCell(value, label) {
  const normalized = text(value, label, 5).toLowerCase();
  if (!normalized) return null;
  if (["true", "yes", "1"].includes(normalized)) return true;
  if (["false", "no", "0"].includes(normalized)) return false;
  fail(`${label} must be true or false`);
}

function nonNegativeInteger(value, label) {
  const normalized = text(value, label, 32);
  if (!normalized) return null;
  if (!/^\d+$/.test(normalized)) fail(`${label} must be a non-negative integer without separators`);
  const parsed = Number(normalized);
  if (!Number.isSafeInteger(parsed)) fail(`${label} is larger than JavaScript's safe integer range`);
  return parsed;
}

export function validateCoverage(rows) {
  if (!Array.isArray(rows) || rows.length === 0) fail("CSV must contain at least one coverage row");
  const unknown = [...new Set(rows.flatMap((row) => Object.keys(row).filter((key) => !COVERAGE_KEYS.has(key))))];
  if (unknown.length) fail(`CSV has unsupported column(s): ${unknown.join(", ")}`);
  const seen = new Set();
  return rows.map((row, index) => {
    const line = index + 2;
    const sourceUrl = httpUrl(row.source_url, `CSV row ${line} source_url`, { required: true });
    if (seen.has(sourceUrl)) fail(`CSV row ${line} duplicates source_url ${sourceUrl}`);
    seen.add(sourceUrl);

    const reach = nonNegativeInteger(row.reach, `CSV row ${line} reach`);
    const reachSourceUrl = httpUrl(row.reach_source_url, `CSV row ${line} reach_source_url`);
    if (reach !== null && !reachSourceUrl) fail(`CSV row ${line} reach requires reach_source_url`);
    if (reach === null && reachSourceUrl) fail(`CSV row ${line} reach_source_url requires reach`);

    const sentiment = text(row.sentiment, `CSV row ${line} sentiment`, 16).toLowerCase();
    if (sentiment && !ALLOWED_SENTIMENTS.has(sentiment)) {
      fail(`CSV row ${line} sentiment must be positive, neutral, negative, or mixed`);
    }

    return {
      position: index + 1,
      source_url: sourceUrl,
      source_host: new URL(sourceUrl).hostname.replace(/^www\./, ""),
      outlet: text(row.outlet, `CSV row ${line} outlet`, 120),
      headline: text(row.headline, `CSV row ${line} headline`, 240),
      published_at: isoDate(row.published_at, `CSV row ${line} published_at`),
      byline: text(row.byline, `CSV row ${line} byline`, 160),
      note: text(row.note, `CSV row ${line} note`, 280),
      coverage_type: text(row.coverage_type, `CSV row ${line} coverage_type`, 80),
      sentiment,
      reach,
      reach_source_url: reachSourceUrl,
      clip_section: text(row.clip_section, `CSV row ${line} clip_section`, 160),
      include_in_reel: booleanCell(row.include_in_reel, `CSV row ${line} include_in_reel`),
    };
  });
}

function hexColor(value, label, fallback) {
  const normalized = text(value, label, 64) || fallback;
  if (!/^#[0-9a-fA-F]{6}$/.test(normalized)) fail(`${label} must be a six-digit hex color`);
  return normalized.toLowerCase();
}

function localPath(value, label, baseDir, extensions) {
  const normalized = text(value, label, 4096);
  if (!normalized) return "";
  if (/^[a-z][a-z0-9+.-]*:/i.test(normalized)) fail(`${label} must be a local path, not a URL`);
  const absolute = isAbsolute(normalized) ? normalized : resolve(baseDir, normalized);
  if (!existsSync(absolute) || !statSync(absolute).isFile()) fail(`${label} does not exist: ${absolute}`);
  if (extensions && !extensions.has(extname(absolute).toLowerCase())) {
    fail(`${label} has an unsupported file type: ${extname(absolute) || "none"}`);
  }
  if (extname(absolute).toLowerCase() === ".svg") {
    const svg = readFileSync(absolute, "utf8");
    if (/<script\b|<foreignObject\b|\bon[a-z]+\s*=|(?:href|xlink:href)\s*=\s*["']\s*(?:https?:|\/\/)|url\(\s*["']?\s*(?:https?:|\/\/)/i.test(svg)) {
      fail(`${label} contains active or remote SVG content; use a self-contained SVG or raster image`);
    }
  }
  return absolute;
}

export function validateBrand(raw, baseDir = process.cwd()) {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) fail("brand JSON must be an object");
  const unknown = Object.keys(raw).filter((key) => !BRAND_KEYS.has(key));
  if (unknown.length) fail(`brand JSON has unknown field(s): ${unknown.join(", ")}`);

  const name = text(raw.name, "brand.name", 120);
  if (!name) fail("brand.name is required");
  const reelMaxItems = raw.reel_max_items ?? 6;
  if (!Number.isInteger(reelMaxItems) || reelMaxItems < 1 || reelMaxItems > 8) {
    fail("brand.reel_max_items must be an integer from 1 to 8");
  }
  const terms = raw.mention_terms ?? [];
  if (!Array.isArray(terms) || terms.length > 20) fail("brand.mention_terms must be a list of up to 20 names");
  const mentionTerms = [...new Set(terms.map((term, i) => text(term, `brand.mention_terms[${i}]`, 120)).filter(Boolean))];

  return {
    name,
    mention_terms: mentionTerms.length ? mentionTerms : [name],
    report_title: text(raw.report_title, "brand.report_title", 160) || `${name} earned coverage`,
    subtitle: text(raw.subtitle, "brand.subtitle", 240),
    period_label: text(raw.period_label, "brand.period_label", 80),
    website: httpUrl(raw.website, "brand.website"),
    logo_path: localPath(raw.logo_path, "brand.logo_path", baseDir, ALLOWED_IMAGE_EXTENSIONS),
    primary_color: hexColor(raw.primary_color, "brand.primary_color", "#13243a"),
    accent_color: hexColor(raw.accent_color, "brand.accent_color", "#f05d3d"),
    background_color: hexColor(raw.background_color, "brand.background_color", "#f5f1e8"),
    closing_note: text(raw.closing_note, "brand.closing_note", 240),
    reel_max_items: reelMaxItems,
  };
}

// ---------------------------------------------------------------------------
// press-clip sidecars (clip.json) and the review gate (clips.json)
// ---------------------------------------------------------------------------

export function imageSize(path) {
  const bytes = readFileSync(path);
  if (bytes.length >= 24 && bytes.toString("latin1", 1, 4) === "PNG" && bytes.toString("latin1", 12, 16) === "IHDR") {
    return { width: bytes.readUInt32BE(16), height: bytes.readUInt32BE(20) };
  }
  if (extname(path).toLowerCase() === ".svg") {
    const head = bytes.toString("utf8", 0, 2000);
    const w = /<svg[^>]*\swidth="(\d+(?:\.\d+)?)"/.exec(head), h = /<svg[^>]*\sheight="(\d+(?:\.\d+)?)"/.exec(head);
    if (w && h) return { width: Number(w[1]), height: Number(h[1]) };
  }
  if (bytes.length > 4 && bytes[0] === 0xff && bytes[1] === 0xd8) {
    for (let at = 2; at + 9 < bytes.length;) {
      if (bytes[at] !== 0xff) break;
      const marker = bytes[at + 1], length = bytes.readUInt16BE(at + 2);
      if (marker >= 0xc0 && marker <= 0xc3) return { width: bytes.readUInt16BE(at + 7), height: bytes.readUInt16BE(at + 5) };
      at += 2 + length;
    }
  }
  return null;
}

function rect(value, label) {
  if (!value || typeof value !== "object") fail(`${label} must be a rectangle`);
  for (const key of ["x", "y", "w", "h"]) {
    if (typeof value[key] !== "number" || !Number.isFinite(value[key])) fail(`${label}.${key} must be a number`);
  }
  if (value.w < 0 || value.h < 0) fail(`${label} has a negative size`);
  return { x: value.x, y: value.y, w: value.w, h: value.h };
}

function rects(list, label) {
  if (!Array.isArray(list)) fail(`${label} must be a list of rectangles`);
  return list.map((r, i) => rect(r, `${label}[${i}]`));
}

function pageField(value, label) {
  if (!value || typeof value !== "object") fail(`${label} must be { value, source }`);
  if (value.value === null || value.value === undefined) return { value: null, source: null };
  if (!PAGE_SOURCES.has(value.source)) fail(`${label}.source must be json-ld, og, meta or visible`);
  return { value: text(value.value, `${label}.value`, 400), source: value.source };
}

// Validate one press-clip sidecar and resolve its files. Rejects anything that would make
// the reel lie: a different URL, missing or mis-sized captures, captures past Chrome's limit,
// rectangles outside the capture, unresolved capture warnings, or no chosen mention.
export function validateSidecar(raw, sidecarDir, sourceUrl, label = "clip.json") {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) fail(`${label} must be an object`);
  if (raw.version !== 1) fail(`${label} version must be 1 (from press-clip --assets)`);
  const url = httpUrl(raw.source_url, `${label}.source_url`, { required: true });
  if (url !== sourceUrl) fail(`${label}.source_url (${url}) does not match the CSV row (${sourceUrl})`);
  if (raw.scale !== 2 && raw.scale !== 3) fail(`${label}.scale must be 2 or 3`);
  const warnings = Array.isArray(raw.warnings) ? raw.warnings : [];
  for (const w of warnings) {
    if (BLOCKING_WARNINGS.has(w?.code)) fail(`${label} has the capture warning "${w.code}": ${text(w.detail, "warning", 600) || "re-run press-clip"}`);
  }
  const page = raw.page || {};
  const meta = {
    outlet: pageField(page.outlet ?? { value: null }, `${label}.page.outlet`),
    headline: pageField(page.headline ?? { value: null }, `${label}.page.headline`),
    byline: pageField(page.byline ?? { value: null }, `${label}.page.byline`),
    published_at: pageField(page.published_at ?? { value: null }, `${label}.page.published_at`),
  };
  if (meta.byline.value) meta.byline.value = meta.byline.value.replace(/\S+@\S+\.\S+/g, "").replace(/\s+/g, " ").trim() || null;

  if (!Array.isArray(raw.assets) || !raw.assets.length) fail(`${label}.assets is empty`);
  const assets = raw.assets.map((asset, i) => {
    const at = `${label}.assets[${i}]`;
    if (!["top", "strip", "mention", "masthead"].includes(asset?.role)) fail(`${at}.role is not top, strip, mention or masthead`);
    const file = text(asset.file, `${at}.file`, 200);
    if (!file || basename(file) !== file || file.startsWith(".") || extname(file).toLowerCase() !== ".png") fail(`${at}.file must be a PNG file name inside the clip folder`);
    const path = join(sidecarDir, file);
    if (!existsSync(path)) fail(`${at}.file does not exist: ${path}`);
    const size = imageSize(path);
    if (!size || size.width !== asset.width || size.height !== asset.height) fail(`${at} says ${asset.width}x${asset.height} but ${file} is ${size ? `${size.width}x${size.height}` : "not a PNG"}`);
    if (size.height > MAX_ASSET_DEVICE_PX || size.width > MAX_ASSET_DEVICE_PX) fail(`${at} is ${size.height} px tall; captures must stay under ${MAX_ASSET_DEVICE_PX} px (Chrome repeats content past its limit)`);
    if (asset.scale !== raw.scale) fail(`${at}.scale must match the sidecar scale`);
    return { role: asset.role, file, path, width: size.width, height: size.height, scale: asset.scale, rect: rect(asset.rect, `${at}.rect`), index: asset.index ?? null };
  });
  const top = assets.find((a) => a.role === "top");
  if (!top) fail(`${label} has no top capture`);
  const strips = assets.filter((a) => a.role === "strip").sort((a, b) => a.rect.y - b.rect.y);
  if (!strips.length) fail(`${label} has no strip captures`);
  strips.forEach((s, i) => {
    if (Math.abs(s.rect.x - top.rect.x) > 0.5 || Math.abs(s.rect.w - top.rect.w) > 0.5) fail(`${label} strip ${s.file} does not share the top capture's column`);
    if (i && Math.abs(strips[i - 1].rect.y + strips[i - 1].rect.h - s.rect.y) > 0.5) fail(`${label} strips leave a gap before ${s.file}`);
    if (Math.abs(s.width / s.rect.w - raw.scale) > 0.02) fail(`${label} strip ${s.file} is not at ${raw.scale}x`);
  });
  const covered = { y0: strips[0].rect.y, y1: strips[strips.length - 1].rect.y + strips[strips.length - 1].rect.h };

  const mentions = Array.isArray(raw.mentions) ? raw.mentions : [];
  const chosenList = mentions.filter((m) => m && m.chosen === true);
  if (chosenList.length !== 1) fail(`${label} must have exactly one chosen mention (found ${chosenList.length}); re-run press-clip with --mention`);
  const chosen = chosenList[0];
  const mention = {
    term: text(chosen.term, `${label} mention term`, 200),
    sentence: text(chosen.sentence, `${label} mention sentence`, 2000),
    clause: text(chosen.clause, `${label} mention clause`, 2000),
    rank: chosen.rank,
    sentence_bars: rects(chosen.rects?.sentence, `${label} mention sentence rects`),
    clause_bars: rects(chosen.rects?.clause ?? [], `${label} mention clause rects`),
    term_bars: rects(chosen.rects?.term ?? [], `${label} mention term rects`),
    paragraph: rect(chosen.rects?.paragraph, `${label} mention paragraph`),
  };
  if (!mention.sentence || !mention.sentence_bars.length) fail(`${label} chosen mention has no sentence`);
  if (mention.term && !mention.sentence.includes(mention.term.split(" ")[0])) fail(`${label} chosen sentence does not contain its matched term`);
  for (const bar of [...mention.sentence_bars, ...mention.clause_bars]) {
    if (bar.y < covered.y0 - 1 || bar.y + bar.h > covered.y1 + 1 || bar.x < top.rect.x - 1 || bar.x + bar.w > top.rect.x + top.rect.w + 1) {
      fail(`${label} mention rectangle lies outside the captured strips`);
    }
  }
  const logo = raw.logo || {};
  let plate = null;
  if (logo.plate_rect) {
    const pr = rect(logo.plate_rect, `${label}.logo.plate_rect`);
    if (pr.x >= top.rect.x - 0.5 && pr.y >= top.rect.y - 0.5 && pr.x + pr.w <= top.rect.x + top.rect.w + 0.5 && pr.y + pr.h <= top.rect.y + top.rect.h + 0.5 && pr.w > 4 && pr.h > 4) {
      plate = { rect: pr, color: typeof logo.plate_color === "string" && /^rgb\(\d{1,3}, \d{1,3}, \d{1,3}\)$/.test(logo.plate_color) ? logo.plate_color : null };
    }
  }
  return {
    source_url: url,
    captured_at: isoTimestamp(raw.captured_at, `${label}.captured_at`),
    scale: raw.scale,
    scope: raw.scope === "section" ? "section" : "whole",
    page: meta,
    logo: { resolved_from: text(logo.resolved_from, `${label}.logo.resolved_from`, 120) || null, kind: logo.kind || null, plate },
    top,
    strips,
    mention_asset: assets.find((a) => a.role === "mention") || null,
    mention,
    warnings: warnings.map((w) => ({ code: text(w.code, "warning code", 60), detail: text(w.detail, "warning detail", 600) })),
  };
}

// clips.json (version 2): one entry per CSV row. A row is either a reviewed press-clip
// capture, the user's own screenshot, or omitted with a reason (for example a page that
// blocked automated capture). Only rows that passed the separate reviewer appear.
export function validateClips(raw, coverageRows, baseDir = process.cwd()) {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) fail("clips JSON must be an object");
  if (raw.version !== 2) fail("clips JSON version must be 2 (each entry points at a press-clip clip.json sidecar)");
  if (!Array.isArray(raw.clips)) fail("clips JSON must contain a clips array");

  const coverageUrls = new Set(coverageRows.map((row) => row.source_url));
  const byUrl = new Map();
  for (let index = 0; index < raw.clips.length; index += 1) {
    const label = `clips[${index}]`;
    const entry = raw.clips[index] || {};
    const sourceUrl = httpUrl(entry.source_url, `${label}.source_url`, { required: true });
    if (!coverageUrls.has(sourceUrl)) fail(`${label}.source_url is not present in the coverage CSV`);
    if (byUrl.has(sourceUrl)) fail(`clips JSON duplicates source_url ${sourceUrl}`);
    const kinds = ["sidecar_path", "user_capture", "omitted"].filter((k) => entry[k]);
    if (kinds.length !== 1) fail(`${label} needs exactly one of sidecar_path, user_capture or omitted`);

    if (entry.omitted) {
      const reason = text(entry.omitted.reason, `${label}.omitted.reason`, 40);
      if (!OMIT_REASONS.has(reason)) fail(`${label}.omitted.reason must be one of ${[...OMIT_REASONS].join(", ")}`);
      byUrl.set(sourceUrl, { kind: "omitted", source_url: sourceUrl, reason, detail: text(entry.omitted.detail, `${label}.omitted.detail`, 280) });
      continue;
    }

    const review = entry.review;
    if (!review || typeof review !== "object" || Array.isArray(review)) fail(`${label}.review is required`);
    if (review.verdict !== "clean") fail(`${label}.review.verdict must be clean`);
    const reviewedAt = isoTimestamp(review.reviewed_at, `${label}.review.reviewed_at`);
    const notes = text(review.notes, `${label}.review.notes`, 280);

    if (entry.user_capture) {
      const imagePath = localPath(entry.user_capture.image_path, `${label}.user_capture.image_path`, baseDir, new Set([".png", ".jpg", ".jpeg"]));
      if (!imagePath) fail(`${label}.user_capture.image_path is required`);
      const size = imageSize(imagePath);
      if (!size) fail(`${label}.user_capture.image_path is not a readable PNG or JPEG`);
      const pdfPath = localPath(entry.user_capture.pdf_path, `${label}.user_capture.pdf_path`, baseDir, new Set([".pdf"]));
      byUrl.set(sourceUrl, { kind: "user", source_url: sourceUrl, image_path: imagePath, image: size, pdf_path: pdfPath, review: { verdict: "clean", supplied_by_user: true, reviewed_at: reviewedAt, notes } });
      continue;
    }

    for (const flag of ["logo_verified", "pdf_reviewed", "mention_verified", "captures_clean"]) {
      if (review[flag] !== true) fail(`${label}.review.${flag} must be true`);
    }
    if (review.scope !== "whole" && review.scope !== "section") fail(`${label}.review.scope must be whole or section`);
    const sidecarPath = localPath(entry.sidecar_path, `${label}.sidecar_path`, baseDir, new Set([".json"]));
    const pdfPath = localPath(entry.pdf_path, `${label}.pdf_path`, baseDir, new Set([".pdf"]));
    if (!pdfPath) fail(`${label}.pdf_path is required`);
    let sidecarRaw;
    try {
      sidecarRaw = JSON.parse(readFileSync(sidecarPath, "utf8"));
    } catch (error) {
      fail(`${label}.sidecar_path is not valid JSON: ${error.message}`);
    }
    const sidecar = validateSidecar(sidecarRaw, dirname(sidecarPath), sourceUrl, `${label} sidecar`);
    if (sidecar.scope !== review.scope) fail(`${label}.review.scope is ${review.scope} but the capture is ${sidecar.scope}`);
    byUrl.set(sourceUrl, {
      kind: "clip", source_url: sourceUrl, sidecar_path: sidecarPath, pdf_path: pdfPath, sidecar,
      review: { verdict: "clean", logo_verified: true, pdf_reviewed: true, mention_verified: true, captures_clean: true, scope: review.scope, reviewed_at: reviewedAt, notes },
    });
  }

  const missing = coverageRows.filter((row) => !byUrl.has(row.source_url));
  if (missing.length) fail(`clips JSON is missing ${missing.length} CSV source(s): ${missing.map((row) => row.source_url).join(", ")}`);
  return coverageRows.map((row) => ({ ...row, clip: byUrl.get(row.source_url) }));
}

// CSV values win. A blank outlet, headline or byline may be filled from the captured page and
// is labelled as read from the page. A blank publish date is never filled: the page's date is
// only reported back so the user can confirm it and add it to the CSV.
export function applyPageMetadata(row) {
  const out = { ...row, outlet_source: row.outlet ? "csv" : null, headline_source: row.headline ? "csv" : null, byline_source: row.byline ? "csv" : null, published_at_source: row.published_at ? "csv" : null, page_published_at: null };
  const page = row.clip && row.clip.kind === "clip" ? row.clip.sidecar.page : null;
  if (!page) return out;
  for (const key of ["outlet", "headline", "byline"]) {
    if (!out[key] && page[key].value) {
      out[key] = page[key].value.slice(0, key === "headline" ? 240 : 160);
      out[`${key}_source`] = `page:${page[key].source}`;
    }
  }
  if (!row.published_at && page.published_at.value) out.page_published_at = { value: page.published_at.value, source: page.published_at.source };
  return out;
}

function formatNumber(value) {
  return new Intl.NumberFormat("en-US", { maximumFractionDigits: 0 }).format(value);
}

export function deriveMetrics(rows) {
  const outlets = [...new Set(rows.map((row) => row.outlet).filter(Boolean))];
  const dates = rows.map((row) => row.published_at).filter(Boolean).sort();
  const reachRows = rows.filter((row) => row.reach !== null);
  const sentimentRows = rows.filter((row) => row.sentiment);
  const sentiments = Object.fromEntries([...ALLOWED_SENTIMENTS].map((name) => [name, 0]));
  sentimentRows.forEach((row) => { sentiments[row.sentiment] += 1; });

  return {
    coverage: {
      label: "Coverage",
      value: formatNumber(rows.length),
      count: rows.length,
      status: "calculated",
      detail: "Count of validated CSV rows",
    },
    outlets: outlets.length
      ? {
          label: "Named outlets",
          value: formatNumber(outlets.length),
          count: outlets.length,
          status: "calculated",
          detail: `Distinct outlet names across ${outlets.length} outlet${outlets.length === 1 ? "" : "s"}`,
        }
      : { label: "Named outlets", value: "Unavailable", status: "unavailable", detail: "No outlet names were supplied" },
    date_range: dates.length
      ? {
          label: "Supplied date range",
          value: dates[0] === dates.at(-1) ? dates[0] : `${dates[0]} — ${dates.at(-1)}`,
          first: dates[0],
          last: dates.at(-1),
          supplied: dates.length,
          status: "calculated",
          detail: `From ${dates.length}/${rows.length} supplied publish date${dates.length === 1 ? "" : "s"}${dates.length < rows.length ? "; partial" : ""}`,
        }
      : { label: "Supplied date range", value: "Unavailable", status: "unavailable", detail: "No publish dates were supplied" },
    reach: reachRows.length
      ? {
          label: reachRows.length === rows.length ? "Reported reach" : "Partial reported reach",
          value: formatNumber(reachRows.reduce((sum, row) => sum + row.reach, 0)),
          count: reachRows.reduce((sum, row) => sum + row.reach, 0),
          rows: reachRows.length,
          status: "sourced",
          detail: `Sum calculated from ${reachRows.length}/${rows.length} source-linked value${reachRows.length === 1 ? "" : "s"}${reachRows.length < rows.length ? "; not total reach" : ""}`,
        }
      : { label: "Reported reach", value: "Unavailable", status: "unavailable", detail: "No source-linked reach values were supplied" },
    sentiment: {
      label: "Supplied sentiment",
      status: sentimentRows.length ? "calculated" : "unavailable",
      detail: sentimentRows.length
        ? `Shares calculated from ${sentimentRows.length}/${rows.length} user-supplied label${sentimentRows.length === 1 ? "" : "s"}; no automated analysis`
        : "No sentiment labels were supplied; no automated analysis was run",
      labeled_count: sentimentRows.length,
      counts: sentiments,
    },
  };
}

// Rows with a reviewed capture (or the user's own screenshot) can be featured. Omitted rows
// are left out and reported, never faked. The CSV decides order and, via include_in_reel,
// which rows to feature; nothing is silently ranked as "top" coverage.
export function chooseReelRows(rows, limit) {
  const hasFlags = rows.some((row) => row.include_in_reel !== null);
  const wanted = hasFlags ? rows.filter((row) => row.include_in_reel === true) : rows;
  if (hasFlags && wanted.length === 0) fail("include_in_reel is present but no row is marked true");
  const left_out = wanted.filter((row) => !row.clip || row.clip.kind === "omitted");
  const selected = wanted.filter((row) => row.clip && row.clip.kind !== "omitted");
  if (selected.length > limit) {
    fail(`${selected.length} rows are selected for the reel, above reel_max_items ${limit}; mark at most ${limit} rows include_in_reel=true`);
  }
  if (!selected.length) fail("no selected row has a reviewed capture; the reel needs at least one");
  return {
    rows: selected,
    left_out,
    rule: hasFlags ? "include_in_reel=true in CSV, preserved in CSV order" : "all CSV rows with a reviewed capture, preserved in CSV order",
  };
}

export function escapeHtml(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

// JSON that is safe inside <script type="application/json">.
export function scriptJson(value) {
  return JSON.stringify(value)
    .replaceAll("<", "\\u003c")
    .replaceAll(">", "\\u003e")
    .replaceAll("&", "\\u0026")
    .replaceAll("\u2028", "\\u2028")
    .replaceAll("\u2029", "\\u2029");
}

function assetName(path, prefix) {
  const bytes = readFileSync(path);
  const digest = createHash("sha256").update(bytes).digest("hex").slice(0, 12);
  return `${prefix}-${digest}${extname(path).toLowerCase()}`;
}

function copyAsset(path, assetsDir, prefix) {
  if (!path) return "";
  const name = assetName(path, prefix);
  mkdirSync(assetsDir, { recursive: true });
  copyFileSync(path, join(assetsDir, name));
  return `assets/${name}`;
}

function articleName(row) {
  return row.outlet || row.source_host;
}

function articleTitle(row) {
  return row.headline || "Headline unavailable";
}

function truncate(value, max) {
  const chars = [...String(value || "")];
  return chars.length <= max ? chars.join("") : `${chars.slice(0, max - 1).join("").trimEnd()}\u2026`;
}

// ---------------------------------------------------------------------------
// Colour
// ---------------------------------------------------------------------------

function hexToRgb(hex) {
  const n = parseInt(hex.slice(1), 16);
  return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
}

function rgbToHex([r, g, b]) {
  return `#${[r, g, b].map((v) => Math.round(Math.max(0, Math.min(255, v))).toString(16).padStart(2, "0")).join("")}`;
}

function mix(a, b, amount) {
  const x = hexToRgb(a), y = hexToRgb(b);
  return rgbToHex(x.map((v, i) => v + (y[i] - v) * amount));
}

function luminance(hex) {
  const [r, g, b] = hexToRgb(hex).map((v) => { const c = v / 255; return c <= 0.03928 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4; });
  return 0.2126 * r + 0.7152 * g + 0.0722 * b;
}

export function contrast(a, b) {
  const [x, y] = [luminance(a), luminance(b)].sort((p, q) => q - p);
  return (x + 0.05) / (y + 0.05);
}

// The reel's colours from the brand file: a deep stage from the primary colour, the accent
// lifted until it reads on that stage, and ink that reads on the paper end card.
export function reelPalette(brand) {
  let stage = brand.primary_color;
  for (let i = 0; luminance(stage) > 0.03 && i < 20; i++) stage = mix(stage, "#08090b", 0.25);
  let accentOnDark = brand.accent_color;
  for (let i = 0; contrast(accentOnDark, stage) < 4.5 && i < 20; i++) accentOnDark = mix(accentOnDark, "#ffffff", 0.15);
  const paper = brand.background_color;
  let ink = brand.primary_color;
  for (let i = 0; contrast(ink, paper) < 7 && i < 20; i++) ink = mix(ink, "#000000", 0.3);
  let accentOnPaper = brand.accent_color;
  for (let i = 0; contrast(accentOnPaper, paper) < 4.5 && i < 20; i++) accentOnPaper = mix(accentOnPaper, "#000000", 0.2);
  return {
    stage,
    stageLift: mix(stage, "#ffffff", 0.07),
    text: "#ffffff",
    muted: "rgba(255, 255, 255, 0.7)",
    accent: brand.accent_color,
    accentOnDark,
    paper,
    ink,
    inkMuted: `rgba(${hexToRgb(ink).join(", ")}, 0.66)`,
    accentOnPaper,
    highlight: "rgba(255, 225, 77, 0.62)",
    highlightSolid: "#ffe14d",
    calculated: "#ffd582",
    sourced: "#9eeacb",
  };
}

// ---------------------------------------------------------------------------
// Dashboard
// ---------------------------------------------------------------------------

function metricCard(metric) {
  return `<article class="metric"><span class="metric-label">${escapeHtml(metric.label)}</span><strong>${escapeHtml(metric.value)}</strong><span class="status status-${escapeHtml(metric.status)}">${escapeHtml(metric.status)}</span><p>${escapeHtml(metric.detail)}</p></article>`;
}

function sourceLabel(source) {
  if (!source || source === "csv") return "";
  return ` <span class="from-page" title="Read from the article page (${escapeHtml(source.replace("page:", ""))}); not in the CSV">read from page</span>`;
}

function dashboardHtml({ brand, rows, metrics, brandLogoAsset }) {
  const logo = brandLogoAsset
    ? `<img class="brand-logo" src="${escapeHtml(brandLogoAsset)}" alt="${escapeHtml(brand.name)} logo">`
    : `<span class="brand-wordmark">${escapeHtml(brand.name)}</span>`;
  const articleCards = rows.map((row, index) => {
    const omitted = row.clip.kind === "omitted";
    const visual = omitted
      ? `<div class="clip-missing"><strong>No capture</strong><span>${escapeHtml(omittedReason(row.clip))}</span></div>`
      : `<img src="${escapeHtml(row.thumb_asset)}" alt="${row.clip.kind === "user" ? "Screenshot supplied by the user" : "Reviewed press-clip capture"} of ${escapeHtml(articleName(row))}">`;
    const mention = row.clip.kind === "clip" ? row.clip.sidecar.mention.sentence : "";
    return `
    <article class="coverage-card${omitted ? " is-omitted" : ""}">
      <a class="clip-link" href="${escapeHtml(row.source_url)}" target="_blank" rel="noopener noreferrer" aria-label="Open source ${index + 1}">${visual}</a>
      <div class="coverage-copy">
        <div class="coverage-kicker"><span>${String(index + 1).padStart(2, "0")}</span><span>${escapeHtml(row.coverage_type || "earned coverage")}</span></div>
        <h3>${escapeHtml(articleTitle(row))}${sourceLabel(row.headline_source)}</h3>
        <p class="outlet">${escapeHtml(row.outlet || "Outlet unavailable")}${sourceLabel(row.outlet_source)} <span>·</span> ${escapeHtml(row.published_at || "Date not supplied")}</p>
        ${row.byline ? `<p class="byline">By ${escapeHtml(row.byline)}${sourceLabel(row.byline_source)}</p>` : ""}
        ${mention ? `<blockquote class="mention"><span>Mention</span>${escapeHtml(truncate(mention, 220))}</blockquote>` : ""}
        ${row.note ? `<p class="note">${escapeHtml(row.note)}</p>` : ""}
        <a class="source-url" href="${escapeHtml(row.source_url)}" target="_blank" rel="noopener noreferrer">${escapeHtml(row.source_url)}</a>
        <div class="facts">
          ${omitted ? `<span>not in the reel: ${escapeHtml(row.clip.reason.replaceAll("_", " "))}</span>` : row.clip.kind === "user" ? "<span>user-supplied screenshot</span>" : "<span>clip review: clean</span><span>logo: verified</span><span>mention: verified</span>"}
          ${row.sentiment ? `<span>sentiment: supplied ${escapeHtml(row.sentiment)}</span>` : `<span>sentiment: unavailable</span>`}
          ${row.reach !== null ? `<a href="${escapeHtml(row.reach_source_url)}" target="_blank" rel="noopener noreferrer">reach: ${escapeHtml(formatNumber(row.reach))} · source</a>` : `<span>reach: unavailable</span>`}
        </div>
      </div>
    </article>`;
  }).join("");

  const sentimentTotal = metrics.sentiment.labeled_count;
  const sentimentRows = [...ALLOWED_SENTIMENTS].map((name) => {
    const count = metrics.sentiment.counts[name];
    const percentage = sentimentTotal ? Math.round((count / sentimentTotal) * 100) : 0;
    return `<div class="sentiment-row"><span>${escapeHtml(name)}</span><div class="bar"><i style="width:${percentage}%"></i></div><strong>${count}${sentimentTotal ? ` · ${percentage}%` : ""}</strong></div>`;
  }).join("");

  return `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src 'self' data:; style-src 'unsafe-inline'; base-uri 'none'; form-action 'none'; frame-ancestors 'none'">
  <title>${escapeHtml(brand.report_title)}</title>
  <style>
    :root{--ink:${brand.primary_color};--accent:${brand.accent_color};--paper:${brand.background_color};--white:#fff;--muted:#607083;--line:color-mix(in srgb,var(--ink) 15%,transparent)}
    *{box-sizing:border-box}html{background:var(--paper);color:var(--ink);font-family:Arial,Helvetica,sans-serif}body{margin:0}a{color:inherit}.shell{max-width:1440px;margin:0 auto;padding:48px}
    .masthead{display:grid;grid-template-columns:1.4fr .6fr;gap:24px;align-items:end;border-bottom:2px solid var(--ink);padding-bottom:34px}.eyebrow{text-transform:uppercase;letter-spacing:.18em;font-size:12px;font-weight:800;color:var(--accent)}h1{font-size:clamp(42px,6vw,84px);line-height:.94;letter-spacing:-.055em;margin:16px 0 18px;max-width:950px}.subtitle{font-size:19px;line-height:1.5;max-width:760px;color:var(--muted);margin:0}.brand-block{text-align:right}.brand-logo{display:block;max-width:280px;max-height:92px;margin-left:auto}.brand-wordmark{font-weight:900;font-size:28px;letter-spacing:-.03em}.period{display:block;margin-top:18px;font-size:13px;text-transform:uppercase;letter-spacing:.14em;color:var(--muted)}
    .metrics{display:grid;grid-template-columns:repeat(4,1fr);gap:14px;margin:22px 0 50px}.metric{min-height:190px;background:var(--ink);color:white;padding:24px;border-radius:18px;display:flex;flex-direction:column}.metric-label{text-transform:uppercase;letter-spacing:.12em;font-size:11px;opacity:.72}.metric strong{font-size:35px;line-height:1.05;letter-spacing:-.04em;margin:18px 0 10px}.metric p{font-size:12px;line-height:1.4;opacity:.68;margin:auto 0 0}.status{align-self:flex-start;border:1px solid currentColor;border-radius:999px;padding:4px 8px;text-transform:uppercase;font-size:9px;letter-spacing:.1em}.status-sourced{color:#9eeacb}.status-unavailable{color:#d7dbe0}.status-calculated{color:#ffd582}
    .section-head{display:flex;justify-content:space-between;align-items:end;margin:0 0 22px}.section-head h2{font-size:32px;letter-spacing:-.035em;margin:0}.section-head p{margin:0;color:var(--muted);font-size:13px}.coverage-list{display:grid;gap:20px}.coverage-card{display:grid;grid-template-columns:minmax(280px,38%) 1fr;gap:28px;background:white;border:1px solid var(--line);border-radius:22px;padding:18px;box-shadow:0 14px 45px rgba(19,36,58,.08)}.coverage-card.is-omitted{opacity:.82}.clip-link{display:block;background:#eef1f4;border-radius:13px;overflow:hidden;min-height:300px;text-decoration:none}.clip-link img{width:100%;height:100%;max-height:440px;object-fit:cover;object-position:top;display:block}.clip-missing{display:flex;flex-direction:column;gap:8px;justify-content:center;height:100%;min-height:300px;padding:28px;color:var(--muted);font-size:14px;line-height:1.45}.clip-missing strong{color:var(--ink);font-size:18px}.coverage-copy{padding:8px 10px 8px 0}.coverage-kicker{display:flex;justify-content:space-between;text-transform:uppercase;letter-spacing:.12em;font-weight:800;font-size:10px;color:var(--accent)}.coverage-copy h3{font-size:29px;line-height:1.05;letter-spacing:-.035em;margin:26px 0 14px}.outlet,.byline{font-size:13px;color:var(--muted)}.from-page{display:inline-block;vertical-align:middle;border:1px dashed currentColor;border-radius:999px;padding:2px 7px;font-size:9px;letter-spacing:.08em;text-transform:uppercase;font-weight:700;color:var(--muted)}.mention{margin:20px 0;padding:14px 16px;background:color-mix(in srgb,#ffe14d 28%,white);border-radius:10px;font-size:15px;line-height:1.5}.mention span{display:block;text-transform:uppercase;letter-spacing:.12em;font-size:9px;font-weight:800;color:var(--muted);margin-bottom:6px}.note{font-size:16px;line-height:1.55;border-left:3px solid var(--accent);padding-left:14px;margin:24px 0}.source-url{display:block;overflow-wrap:anywhere;font:12px/1.4 ui-monospace,SFMono-Regular,Menlo,monospace;color:var(--muted);margin:22px 0}.facts{display:flex;gap:7px;flex-wrap:wrap}.facts span,.facts a{border:1px solid var(--line);border-radius:999px;padding:7px 9px;font-size:10px;text-decoration:none;background:var(--paper)}
    .sentiment{margin:50px 0 20px;background:white;border-radius:22px;padding:28px;border:1px solid var(--line)}.sentiment h2{font-size:26px;margin:0 0 8px}.sentiment>p{color:var(--muted);font-size:12px;margin:0 0 24px}.sentiment-row{display:grid;grid-template-columns:90px 1fr 90px;gap:14px;align-items:center;margin:12px 0;text-transform:capitalize;font-size:13px}.bar{height:10px;background:var(--paper);border-radius:999px;overflow:hidden}.bar i{display:block;height:100%;background:var(--accent);border-radius:inherit}.sentiment-row strong{text-align:right;font-size:12px}.footer{display:grid;grid-template-columns:1fr auto;gap:30px;border-top:2px solid var(--ink);margin-top:50px;padding:26px 0;color:var(--muted);font-size:12px;line-height:1.5}.footer strong{color:var(--ink)}
    @media(max-width:900px){.shell{padding:24px}.masthead,.coverage-card{grid-template-columns:1fr}.brand-block{text-align:left}.brand-logo{margin:0}.metrics{grid-template-columns:repeat(2,1fr)}.coverage-copy{padding:8px}.footer{grid-template-columns:1fr}}@media(max-width:560px){.metrics{grid-template-columns:1fr}.shell{padding:16px}h1{font-size:46px}.coverage-copy h3{font-size:24px}}
    @media print{.shell{max-width:none;padding:18mm}.coverage-card,.metric,.sentiment{break-inside:avoid}.coverage-card{box-shadow:none}a{text-decoration:none}}
  </style>
</head>
<body>
  <main class="shell">
    <header class="masthead"><div><div class="eyebrow">Earned media · source-linked</div><h1>${escapeHtml(brand.report_title)}</h1>${brand.subtitle ? `<p class="subtitle">${escapeHtml(brand.subtitle)}</p>` : ""}</div><div class="brand-block">${logo}<span class="period">${escapeHtml(brand.period_label || "Period not supplied")}</span></div></header>
    <section class="metrics">${metricCard(metrics.coverage)}${metricCard(metrics.outlets)}${metricCard(metrics.date_range)}${metricCard(metrics.reach)}</section>
    <div class="section-head"><h2>The coverage</h2><p>Reviewed press-clip captures · open each live source · "read from page" marks details taken from the article, not the CSV</p></div>
    <section class="coverage-list">${articleCards}</section>
    <section class="sentiment"><h2>Supplied sentiment labels</h2><p><span class="status status-${escapeHtml(metrics.sentiment.status)}">${escapeHtml(metrics.sentiment.status)}</span> ${escapeHtml(metrics.sentiment.detail)}</p>${sentimentRows}</section>
    <footer class="footer"><div><strong>Evidence, not reconstruction.</strong><br>Article visuals are reviewed press-clip captures. Reporting and outlet branding remain the publishers' property; highlights in the reel are added emphasis over unaltered text.</div><div>${brand.website ? `<a href="${escapeHtml(brand.website)}" target="_blank" rel="noopener noreferrer">${escapeHtml(brand.website)}</a>` : escapeHtml(brand.name)}</div></footer>
  </main>
</body>
</html>`;
}

function omittedReason(clip) {
  const base = {
    blocked: "The page blocked automated capture. Ask the user for their own screenshot to include it.",
    paywall: "The article is behind a paywall. Only the user's own subscriber screenshot can be used.",
    capture_failed: "The capture failed. Re-run press-clip or use the user's own screenshot.",
    no_mention: "The client is not named in the article body.",
    user_excluded: "Left out at the user's request.",
  }[clip.reason];
  return clip.detail ? `${base} ${clip.detail}` : base;
}

// ---------------------------------------------------------------------------
// Reel data
// ---------------------------------------------------------------------------

function reelFacts(rows, reelRows, metrics) {
  const parts = [`${metrics.coverage.count} piece${metrics.coverage.count === 1 ? "" : "s"} of coverage`];
  if (metrics.outlets.status !== "unavailable") parts.push(`${metrics.outlets.count} outlet${metrics.outlets.count === 1 ? "" : "s"}`);
  const span = metrics.date_range.status !== "unavailable" ? dateSpanLabel(metrics.date_range.first, metrics.date_range.last) : null;
  if (span) parts.push(span);
  const counters = [{ kind: "count", value: metrics.coverage.count, display: String(metrics.coverage.count), label: "pieces of coverage", status: "calculated" }];
  if (metrics.outlets.status !== "unavailable") counters.push({ kind: "count", value: metrics.outlets.count, display: String(metrics.outlets.count), label: "named outlets", status: "calculated" });
  if (metrics.reach.status === "sourced") {
    counters.push({ kind: "count", value: metrics.reach.count, display: formatNumber(metrics.reach.count), label: metrics.reach.rows < rows.length ? `reported reach, ${metrics.reach.rows} of ${rows.length} items (not total reach)` : "reported reach", status: "sourced" });
  }
  const more = rows.length - reelRows.length;
  return {
    line: parts.join(" · "),
    counters,
    caption: more > 0 ? `${reelRows.length} shown in this reel · ${rows.length} in the coverage list` : "",
    more: more > 0 ? `+${more} more in the full coverage list` : "",
  };
}

function reelItem(row, n, assets) {
  const item = {
    n,
    outlet: articleName(row),
    headline: row.headline || "",
    byline: row.byline || "",
    date_label: dateLabel(row.published_at),
    url: row.source_url,
    short_url: shortUrl(row.source_url),
    kicker: row.coverage_type || "",
    thumb: assets.thumb,
    masthead: null,
  };
  const clip = row.clip;
  if (clip.kind === "clip") {
    const s = clip.sidecar;
    item.capture = {
      kind: "clip",
      scale: s.scale,
      column: { x: s.top.rect.x, w: s.top.rect.w },
      top_y: s.strips[0].rect.y,
      bottom_y: s.strips[s.strips.length - 1].rect.y + s.strips[s.strips.length - 1].rect.h,
      strips: s.strips.map((st, i) => ({ src: assets.strips[i], y: st.rect.y, h: st.rect.h, width: st.width, height: st.height })),
      mention: {
        term: s.mention.term,
        sentence: s.mention.sentence,
        clause: s.mention.clause || s.mention.sentence,
        sentence_bars: s.mention.sentence_bars,
        clause_bars: s.mention.clause_bars.length ? s.mention.clause_bars : s.mention.sentence_bars,
        paragraph: s.mention.paragraph,
      },
    };
    if (s.logo.plate) {
      const p = s.logo.plate.rect, k = s.scale;
      item.masthead = {
        src: assets.thumb.src,
        crop: { x: (p.x - s.top.rect.x) * k, y: (p.y - s.top.rect.y) * k, w: p.w * k, h: p.h * k },
        plate: s.logo.plate.color,
      };
    }
  } else {
    item.capture = { kind: "image", src: assets.thumb.src, width: clip.image.width, height: clip.image.height };
  }
  return item;
}

// Inline timeline.mjs into reel.html: drop the `export` keywords so the module body runs as a
// plain script (browsers block ES module imports from file:// pages).
export function inlineModule(source) {
  const stripped = source.replace(/^export (?=(?:async\s+)?function\b|const\b|let\b|class\b)/gm, "");
  if (/^export\b/m.test(stripped)) throw new Error("timeline.mjs may only use `export function` / `export const` declarations");
  return stripped;
}

export function reelHtml({ brand, timeline, palette }) {
  const bundle = `(() => {\n"use strict";\n${inlineModule(readFileSync(join(SKILL_DIR, "timeline.mjs"), "utf8"))}\n${readFileSync(join(SKILL_DIR, "reel-player.js"), "utf8")}\n})();`;
  const hash = createHash("sha256").update(bundle).digest("base64");
  const data = scriptJson({ timeline, palette });
  return `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src 'self' data:; style-src 'unsafe-inline'; script-src 'sha256-${hash}'; base-uri 'none'; form-action 'none'">
<title>${escapeHtml(brand.report_title)} · reel</title>
<style>
html,body{margin:0;height:100%;background:#0b0b0d;color:#e6e7ea;font:14px/1.4 system-ui,-apple-system,"Segoe UI",sans-serif}
.wrap{box-sizing:border-box;height:100%;display:flex;flex-direction:column;align-items:center;justify-content:center;gap:14px;padding:18px}
canvas{display:block;max-width:100%;max-height:calc(100vh - 90px);width:auto;height:auto;box-shadow:0 24px 70px rgba(0,0,0,.55);border-radius:6px}
.controls{display:flex;gap:12px;align-items:center;width:min(100%,960px)}
button{background:#f2f2f0;color:#111;border:0;border-radius:999px;padding:8px 18px;font-weight:600;cursor:pointer}
input[type=range]{flex:1;accent-color:${palette.accent}}
#clock{font-variant-numeric:tabular-nums;min-width:110px;text-align:right;opacity:.75}
</style>
</head>
<body>
<div class="wrap">
<canvas id="reel" width="${timeline.width}" height="${timeline.height}" role="img" aria-label="${escapeHtml(`${brand.report_title}: coverage reel, ${timeline.items.length} articles. Sources are listed at the end and in sources.txt.`)}"></canvas>
<div class="controls"><button id="play" type="button">Play</button><input id="scrub" type="range" min="0" max="0" value="0" aria-label="Scrub through the reel"><span id="clock"></span></div>
</div>
<script type="application/json" id="reel-data">${data}</script>
<script>${bundle}</script>
</body>
</html>`;
}

function manifestPayload({ brand, rows, metrics, reel, timeline, format, lints }) {
  return {
    version: 2,
    brand: {
      name: brand.name,
      mention_terms: brand.mention_terms,
      report_title: brand.report_title,
      subtitle: brand.subtitle || null,
      period_label: brand.period_label || null,
      website: brand.website || null,
      logo_asset: brand.logo_asset || null,
      primary_color: brand.primary_color,
      accent_color: brand.accent_color,
      background_color: brand.background_color,
      closing_note: brand.closing_note || null,
      reel_max_items: brand.reel_max_items,
    },
    metrics,
    reel: {
      format,
      fps: timeline.fps,
      duration_seconds: timeline.duration,
      frames: timeline.frames,
      selected_source_urls: reel.rows.map((row) => row.source_url),
      left_out: reel.left_out.map((row) => ({ source_url: row.source_url, reason: row.clip ? row.clip.reason : "no capture" })),
      selection_rule: reel.rule,
      scenes: timeline.scenes.map((s) => ({ id: s.id, kind: s.kind, start: s.start, end: s.end })),
      motion_lints: lints.length ? lints : "all passed",
    },
    coverage: rows.map((row) => ({
      position: row.position,
      source_url: row.source_url,
      source_host: row.source_host,
      outlet: row.outlet || null,
      outlet_source: row.outlet_source,
      headline: row.headline || null,
      headline_source: row.headline_source,
      published_at: row.published_at || null,
      published_at_source: row.published_at_source,
      page_published_at_unconfirmed: row.page_published_at,
      byline: row.byline || null,
      byline_source: row.byline_source,
      note: row.note || null,
      coverage_type: row.coverage_type || null,
      sentiment: row.sentiment || null,
      reach: row.reach,
      reach_source_url: row.reach_source_url || null,
      clip_section: row.clip_section || null,
      include_in_reel: row.include_in_reel,
      capture: row.clip.kind === "omitted"
        ? { kind: "omitted", reason: row.clip.reason, detail: row.clip.detail || null }
        : row.clip.kind === "user"
        ? { kind: "user_screenshot", image_asset: row.thumb_asset, review: row.clip.review }
        : {
            kind: "press_clip",
            captured_at: row.clip.sidecar.captured_at,
            scope: row.clip.sidecar.scope,
            logo_resolved_from: row.clip.sidecar.logo.resolved_from,
            top_asset: row.thumb_asset,
            mention: { term: row.clip.sidecar.mention.term, sentence: row.clip.sidecar.mention.sentence, highlighted: row.clip.sidecar.mention.clause || row.clip.sidecar.mention.sentence },
            capture_warnings: row.clip.sidecar.warnings,
            review: row.clip.review,
          },
    })),
  };
}

export function renderArtifacts({ coverageRows, brand, clips, outDir, format = "landscape", fps = 30 }) {
  if (!FORMAT_NAMES.includes(format)) fail(`--format must be one of ${FORMAT_NAMES.join(", ")}`);
  const destination = resolve(outDir);
  const combined = validateClips(clips.raw, coverageRows, clips.baseDir).map(applyPageMetadata);
  const assetsDir = join(destination, "assets");
  mkdirSync(assetsDir, { recursive: true });

  const rows = combined.map((row) => {
    const prefix = `clip-${String(row.position).padStart(2, "0")}`;
    if (row.clip.kind === "clip") {
      const s = row.clip.sidecar;
      const thumb = copyAsset(s.top.path, assetsDir, `${prefix}-top`);
      return {
        ...row,
        thumb_asset: thumb,
        reel_assets: {
          thumb: { src: thumb, width: s.top.width, height: s.top.height },
          strips: s.strips.map((st, i) => copyAsset(st.path, assetsDir, `${prefix}-strip-${i + 1}`)),
        },
      };
    }
    if (row.clip.kind === "user") {
      const thumb = copyAsset(row.clip.image_path, assetsDir, `${prefix}-user`);
      return { ...row, thumb_asset: thumb, reel_assets: { thumb: { src: thumb, width: row.clip.image.width, height: row.clip.image.height } } };
    }
    return { ...row, thumb_asset: "" };
  });
  const brandLogoAsset = copyAsset(brand.logo_path, assetsDir, "brand-logo");
  const safeBrand = { ...brand, logo_asset: brandLogoAsset };
  const metrics = deriveMetrics(rows);
  const reel = chooseReelRows(rows, brand.reel_max_items);
  const palette = reelPalette(brand);
  const logoSize = brandLogoAsset ? imageSize(brand.logo_path) : null;
  const timeline = buildTimeline({
    brand: {
      name: brand.name,
      report_title: brand.report_title,
      overline: brand.period_label || "Coverage recap",
      logo: brandLogoAsset && logoSize ? { src: brandLogoAsset, width: logoSize.width, height: logoSize.height } : null,
    },
    items: reel.rows.map((row, i) => reelItem(row, i + 1, row.reel_assets)),
    facts: reelFacts(rows, reel.rows, metrics),
    format,
    fps,
    disclaimer: DISCLAIMER,
  });
  const lints = lintTimeline(timeline);
  const dashboard = dashboardHtml({ brand: safeBrand, rows, metrics, brandLogoAsset });
  const reelDocument = reelHtml({ brand: safeBrand, timeline, palette });
  const manifest = manifestPayload({ brand: safeBrand, rows, metrics, reel, timeline, format, lints });
  const sources = rows.map((row, index) => `${index + 1}. ${articleName(row)}${row.published_at ? ` (${row.published_at})` : ""}${row.clip.kind === "omitted" ? " [not in the reel]" : ""}\n${row.source_url}`).join("\n\n");

  writeFileSync(join(destination, "index.html"), dashboard);
  writeFileSync(join(destination, "manifest.json"), `${JSON.stringify(manifest, null, 2)}\n`);
  writeFileSync(join(destination, "sources.txt"), `${sources}\n`);
  writeFileSync(join(destination, "reel.html"), reelDocument);
  return { destination, dashboard, manifest, reelDocument, rows, reelRows: reel.rows, timeline, lints, palette };
}

// ---------------------------------------------------------------------------
// Browser steps: dashboard preview, storyboard, frames, video
// ---------------------------------------------------------------------------

function loadChromium() {
  const bases = [process.cwd(), SKILL_DIR, join(SKILL_DIR, "..", "..")];
  if (process.env.PRESS_CLIP_MODULES) bases.push(process.env.PRESS_CLIP_MODULES);
  try {
    bases.push(execFileSync("npm", ["root", "-g"], { encoding: "utf8" }).trim());
  } catch {}
  for (const base of bases) {
    try {
      return createRequire(join(resolve(base), "package.json"))("playwright-core").chromium;
    } catch {}
  }
  fail("Previews and video need playwright-core. Run: npm i playwright-core");
}

export function findBrowser(chromeArg = "") {
  const explicit = chromeArg || process.env.COVERAGE_REEL_CHROME || process.env.PRESS_CLIP_CHROME;
  if (explicit) {
    if (!existsSync(explicit)) fail(`Chrome/Edge executable does not exist: ${explicit}`);
    return explicit;
  }
  const candidates = [
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
    "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
    "/usr/bin/google-chrome",
    "/usr/bin/google-chrome-stable",
    "/usr/bin/microsoft-edge",
    "/usr/bin/chromium",
    "/usr/bin/chromium-browser",
    "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
    "C:\\Program Files (x86)\\Google\\Chrome\\Application\\chrome.exe",
    "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe",
  ];
  return candidates.find((candidate) => existsSync(candidate)) || "";
}

async function withBrowser(chromeArg, callback) {
  const executablePath = findBrowser(chromeArg);
  if (!executablePath) fail("Could not find Chrome or Edge. Install Google Chrome, set COVERAGE_REEL_CHROME, or pass --chrome <path>");
  // file:// captures must not taint the canvas, or Chrome refuses to turn frames into video
  const browser = await loadChromium().launch({ executablePath, headless: true, args: ["--allow-file-access-from-files", "--force-color-profile=srgb", "--hide-scrollbars"] });
  try {
    return await callback(browser);
  } finally {
    await browser.close();
  }
}

async function openReel(browser, reelPath, width, height) {
  const page = await browser.newPage({ viewport: { width: Math.min(width, 1280), height: Math.min(height, 1000) }, deviceScaleFactor: 1 });
  const errors = [];
  page.on("pageerror", (error) => errors.push(error.message));
  await page.goto(`${pathToFileURL(reelPath).href}?export=1`, { waitUntil: "load" });
  await page.evaluate(() => window.reel.ready).catch((error) => { throw new Error(`reel.html failed to load: ${errors.join("; ") || error.message}`); });
  if (errors.length) throw new Error(`reel.html raised an error: ${errors.join("; ")}`);
  // frame 0 is the thumbnail: the whole report title has to fit on it
  const layout = await page.evaluate(() => window.reel.layoutCheck());
  if (layout.title.truncated) fail(`The report title does not fit the ${width}x${height} opening frame ("${layout.title.lines.join(" ")}"). Shorten report_title in brand.json.`);
  return page;
}

export async function renderDashboardPreview(indexPath, outputPath, chromeArg = "") {
  await withBrowser(chromeArg, async (browser) => {
    const page = await browser.newPage({ viewport: { width: 1440, height: 1000 }, deviceScaleFactor: 1 });
    await page.goto(pathToFileURL(indexPath).href, { waitUntil: "load" });
    const height = await page.evaluate(() => document.documentElement.scrollHeight);
    await page.screenshot({ path: outputPath, fullPage: true, animations: "disabled", ...(height > 16000 ? { clip: { x: 0, y: 0, width: 1440, height: 16000 } } : {}) });
  });
}

// Key moments of every scene, for a contact sheet a reviewer can scan at a glance.
export function storyboardTimes(timeline) {
  const times = [0, Math.min(1.2, timeline.duration)];
  for (const s of timeline.scenes) {
    if (s.kind === "opening") times.push(s.chipReveal.start + 0.9);
    if (s.kind === "article") {
      times.push(s.start + 0.9);
      if (s.camera && s.camera.travel) times.push(s.camera.travel.start + s.camera.travel.dur * 0.5);
      if (s.mark) times.push(s.mark.end + 0.3);
      // the end of the hold, at full drift: where the highlight sits closest to the card edge
      if (s.mark) times.push(s.exitAt - 0.05);
    }
    if (s.kind === "wall") times.push(s.counterReveal.start + 1.6);
    if (s.kind === "end") times.push(Math.min(s.end - 0.05, s.start + 2.4));
  }
  return [...new Set(times.map((t) => Math.round(Math.min(timeline.duration - 1 / timeline.fps, t) * 100) / 100))].sort((a, b) => a - b);
}

export async function renderStoryboard({ reelPath, outputPath, timeline, chromeArg = "", framesAt = [], framesDir = "" }) {
  return withBrowser(chromeArg, async (browser) => {
    const page = await openReel(browser, reelPath, timeline.width, timeline.height);
    const columns = timeline.width > timeline.height ? 4 : 5;
    const tileWidth = timeline.width > timeline.height ? 480 : 300;
    const dataUrl = await page.evaluate(([t, c, w]) => window.reel.storyboard(t, c, w), [storyboardTimes(timeline), columns, tileWidth]);
    writeFileSync(outputPath, Buffer.from(dataUrl.split(",")[1], "base64"));
    const written = [];
    if (framesAt.length) {
      mkdirSync(framesDir, { recursive: true });
      for (const t of framesAt) {
        const url = await page.evaluate((x) => window.reel.frameAt(x), t);
        const file = join(framesDir, `frame-${t.toFixed(2).padStart(6, "0")}s.png`);
        writeFileSync(file, Buffer.from(url.split(",")[1], "base64"));
        written.push(file);
      }
    }
    return written;
  });
}

export async function renderVideo({ reelPath, outputPath, timeline, chromeArg = "", maxFrames = Infinity }) {
  return withBrowser(chromeArg, async (browser) => {
    const page = await openReel(browser, reelPath, timeline.width, timeline.height);
    const fps = timeline.fps;
    const codec = fps > 30 ? "avc1.64002A" : "avc1.640028";      // H.264 High, level 4.2 / 4.0
    const bitrate = fps > 30 ? 12_000_000 : 8_000_000;
    const acceleration = await page.evaluate((cfg) => window.reel.startEncoder(cfg), { codec, bitrate });
    if (!acceleration) {
      fail("This browser cannot encode H.264 video (WebCodecs). Open-source Chromium builds usually leave H.264 out. Install Google Chrome and point the renderer at it with --chrome \"/path/to/Google Chrome\" or COVERAGE_REEL_CHROME.");
    }
    const frames = Math.min(timeline.frames, maxFrames);
    const samples = [];
    let avcC = null;
    const collect = (batch) => {
      const data = Buffer.from(batch.data, "base64");
      let at = 0;
      batch.sizes.forEach((size, i) => { samples.push({ data: data.subarray(at, at + size), key: batch.keys[i] }); at += size; });
      if (batch.meta && !avcC) avcC = Buffer.from(batch.meta.avcC, "base64");
    };
    const started = Date.now();
    const batchSize = fps * 3;
    for (let from = 0; from < frames; from += batchSize) {
      await page.evaluate(([a, b, k]) => window.reel.encodeRange(a, b, k), [from, Math.min(frames, from + batchSize), fps]);
      collect(await page.evaluate(() => window.reel.drain(false)));
    }
    collect(await page.evaluate(() => window.reel.drain(true)));
    if (!avcC) fail("The encoder returned no H.264 configuration; try Google Chrome with --chrome");
    if (samples.length !== frames) fail(`The encoder returned ${samples.length} frames for ${frames} drawn`);
    const mp4 = muxMp4({ width: timeline.width, height: timeline.height, fps, avcC, samples });
    writeFileSync(outputPath, mp4);
    return { acceleration, seconds: (Date.now() - started) / 1000, summary: readMp4Summary(readFileSync(outputPath)) };
  });
}

// ---------------------------------------------------------------------------
// CLI
// ---------------------------------------------------------------------------

function usage() {
  return `Usage:
  node render.mjs --csv coverage.csv --brand brand.json --clips clips.json --out output [options]

Options:
  --format <name>      landscape (1920x1080, default), portrait (1080x1350, 4:5),
                       vertical (1080x1920, 9:16) or square (1080x1080)
  --fps 30|60          Frame rate (default 30)
  --preview            Write dashboard.png and storyboard.png (needs Chrome or Edge)
  --video              Also write highlight.mp4, encoded inside Chrome (no ffmpeg)
  --frames-at <list>   Also write full-size frames at these times, e.g. "3.2,8.5"
  --chrome <path>      Chrome or Edge executable (or set COVERAGE_REEL_CHROME)
  --help               Show this help`;
}

function parseArgs(argv) {
  const args = {};
  const booleans = new Set(["preview", "video", "help"]);
  const values = new Set(["csv", "brand", "clips", "out", "chrome", "format", "fps", "frames-at"]);
  for (let index = 0; index < argv.length; index += 1) {
    const token = argv[index];
    if (!token.startsWith("--")) fail(`Unexpected argument: ${token}`);
    const key = token.slice(2);
    if (booleans.has(key)) {
      args[key] = true;
    } else {
      if (!values.has(key)) fail(`Unknown option: --${key}`);
      const value = argv[index + 1];
      if (!value || value.startsWith("--")) fail(`Missing value for --${key}`);
      args[key] = value;
      index += 1;
    }
  }
  return args;
}

async function main() {
  const args = parseArgs(process.argv.slice(2));
  if (args.help) {
    console.log(usage());
    return;
  }
  for (const required of ["csv", "brand", "clips", "out"]) {
    if (!args[required]) fail(`--${required} is required\n\n${usage()}`);
  }
  const fps = args.fps === undefined ? 30 : Number(args.fps);
  if (fps !== 30 && fps !== 60) fail("--fps must be 30 or 60");
  const format = args.format || "landscape";
  const framesAt = args["frames-at"] ? args["frames-at"].split(",").map((v) => Number(v.trim())) : [];
  if (framesAt.some((t) => !Number.isFinite(t) || t < 0)) fail("--frames-at must be a comma list of seconds");

  const csvPath = resolve(args.csv);
  const brandPath = resolve(args.brand);
  const clipsPath = resolve(args.clips);
  const coverageRows = validateCoverage(parseCsv(readFileSync(csvPath, "utf8")));
  const brand = validateBrand(JSON.parse(readFileSync(brandPath, "utf8")), dirname(brandPath));
  const clips = { raw: JSON.parse(readFileSync(clipsPath, "utf8")), baseDir: dirname(clipsPath) };
  const result = renderArtifacts({ coverageRows, brand, clips, outDir: args.out, format, fps });
  const out = result.destination;

  console.log(`Coverage dashboard written to ${join(out, "index.html")}`);
  console.log(`Reel player written to ${join(out, "reel.html")} (${format}, ${result.timeline.duration.toFixed(1)} s, ${result.reelRows.length} article${result.reelRows.length === 1 ? "" : "s"}; open it in a browser to watch and scrub)`);
  console.log(`Metrics: coverage=${result.rows.length}; reach=${result.manifest.metrics.reach.status}; sentiment=${result.manifest.metrics.sentiment.status}`);
  for (const row of result.rows) {
    if (row.clip.kind === "omitted") console.log(`Left out of the reel: ${row.source_url} (${row.clip.reason}). ${omittedReason(row.clip)}`);
    for (const key of ["outlet", "headline", "byline"]) if (row[`${key}_source`]?.startsWith("page:")) console.log(`Row ${row.position} ${key} read from the page (${row[`${key}_source`].slice(5)}): "${row[key]}"`);
    if (row.page_published_at) console.log(`Row ${row.position} has no publish date. The page says ${row.page_published_at.value} (${row.page_published_at.source}); confirm it with the user and add it to the CSV. It was not filled in.`);
  }
  if (result.lints.length) {
    for (const lint of result.lints) console.error(`Motion check failed (${lint.rule}): ${lint.detail}`);
    fail("the reel timeline failed its motion checks; nothing was exported");
  }
  console.log("Motion checks: all passed");

  if (args.preview || args.video || framesAt.length) {
    await renderDashboardPreview(join(out, "index.html"), join(out, "dashboard.png"), args.chrome);
    const frames = await renderStoryboard({ reelPath: join(out, "reel.html"), outputPath: join(out, "storyboard.png"), timeline: result.timeline, chromeArg: args.chrome, framesAt, framesDir: join(out, "frames") });
    console.log(`Dashboard preview written to ${join(out, "dashboard.png")}`);
    console.log(`Storyboard written to ${join(out, "storyboard.png")}`);
    for (const file of frames) console.log(`Frame written to ${file}`);
  }
  if (args.video) {
    const video = await renderVideo({ reelPath: join(out, "reel.html"), outputPath: join(out, "highlight.mp4"), timeline: result.timeline, chromeArg: args.chrome });
    const s = video.summary;
    console.log(`Highlight reel written to ${join(out, "highlight.mp4")} in ${video.seconds.toFixed(1)} s (${video.acceleration === "prefer-software" ? "software encoder, repeatable output" : "hardware encoder"})`);
    console.log(`MP4 check: ${s.codec} ${s.width}x${s.height}, ${s.fps} fps, ${s.frames} frames, ${s.duration_seconds} s, ${s.keyframes} keyframes, ${s.faststart ? "moov before mdat" : "moov AFTER mdat"}`);
    if (s.width !== result.timeline.width || s.height !== result.timeline.height || s.frames !== result.timeline.frames) fail("the MP4 does not match the timeline (size or frame count)");
  }
}

if (process.argv[1] && resolve(process.argv[1]) === resolve(MODULE_PATH)) {
  main().catch((error) => {
    const prefix = error instanceof InputError ? "Input error" : "Error";
    console.error(`${prefix}: ${error.message}`);
    process.exitCode = 1;
  });
}
