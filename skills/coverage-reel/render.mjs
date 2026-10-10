#!/usr/bin/env node

import { createHash } from "node:crypto";
import { execFileSync, spawnSync } from "node:child_process";
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
  dirname,
  extname,
  isAbsolute,
  join,
  resolve,
} from "node:path";
import { fileURLToPath, pathToFileURL } from "node:url";

const MODULE_PATH = fileURLToPath(import.meta.url);
const ALLOWED_SENTIMENTS = new Set(["positive", "neutral", "negative", "mixed"]);
const ALLOWED_IMAGE_EXTENSIONS = new Set([".png", ".jpg", ".jpeg", ".webp", ".svg"]);
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

  return {
    name,
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

export function validateClips(raw, coverageRows, baseDir = process.cwd()) {
  if (!raw || typeof raw !== "object" || Array.isArray(raw)) fail("clips JSON must be an object");
  if (raw.version !== 1) fail("clips JSON version must be 1");
  if (!Array.isArray(raw.clips)) fail("clips JSON must contain a clips array");

  const coverageUrls = new Set(coverageRows.map((row) => row.source_url));
  const byUrl = new Map();
  for (let index = 0; index < raw.clips.length; index += 1) {
    const label = `clips[${index}]`;
    const sourceUrl = httpUrl(raw.clips[index]?.source_url, `${label}.source_url`, { required: true });
    if (!coverageUrls.has(sourceUrl)) fail(`${label}.source_url is not present in the coverage CSV`);
    if (byUrl.has(sourceUrl)) fail(`clips JSON duplicates source_url ${sourceUrl}`);

    const review = raw.clips[index]?.review;
    if (!review || typeof review !== "object" || Array.isArray(review)) fail(`${label}.review is required`);
    if (review.verdict !== "clean") fail(`${label}.review.verdict must be clean`);
    if (review.logo_verified !== true) fail(`${label}.review.logo_verified must be true`);
    if (review.pdf_reviewed !== true) fail(`${label}.review.pdf_reviewed must be true`);
    if (review.scope !== "whole" && review.scope !== "section") {
      fail(`${label}.review.scope must be whole or section`);
    }

    const previewPath = localPath(
      raw.clips[index]?.preview_path,
      `${label}.preview_path`,
      baseDir,
      ALLOWED_IMAGE_EXTENSIONS,
    );
    if (!previewPath) fail(`${label}.preview_path is required`);
    const pdfPath = localPath(
      raw.clips[index]?.pdf_path,
      `${label}.pdf_path`,
      baseDir,
      new Set([".pdf"]),
    );
    if (!pdfPath) fail(`${label}.pdf_path is required`);

    byUrl.set(sourceUrl, {
      source_url: sourceUrl,
      preview_path: previewPath,
      pdf_path: pdfPath,
      review: {
        verdict: "clean",
        logo_verified: true,
        pdf_reviewed: true,
        scope: review.scope,
        reviewed_at: isoTimestamp(review.reviewed_at, `${label}.review.reviewed_at`),
        notes: text(review.notes, `${label}.review.notes`, 280),
      },
    });
  }

  const missing = coverageRows.filter((row) => !byUrl.has(row.source_url));
  if (missing.length) fail(`clips JSON is missing ${missing.length} CSV source(s): ${missing.map((row) => row.source_url).join(", ")}`);
  return coverageRows.map((row) => ({ ...row, clip: byUrl.get(row.source_url) }));
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
      status: "calculated",
      detail: "Count of validated CSV rows",
    },
    outlets: outlets.length
      ? {
          label: "Named outlets",
          value: formatNumber(outlets.length),
          status: "calculated",
          detail: `Distinct supplied outlet names across ${outlets.length} outlet${outlets.length === 1 ? "" : "s"}`,
        }
      : { label: "Named outlets", value: "Unavailable", status: "unavailable", detail: "No outlet names were supplied" },
    date_range: dates.length
      ? {
          label: "Supplied date range",
          value: dates[0] === dates.at(-1) ? dates[0] : `${dates[0]} — ${dates.at(-1)}`,
          status: "calculated",
          detail: `From ${dates.length}/${rows.length} supplied publish date${dates.length === 1 ? "" : "s"}${dates.length < rows.length ? "; partial" : ""}`,
        }
      : { label: "Supplied date range", value: "Unavailable", status: "unavailable", detail: "No publish dates were supplied" },
    reach: reachRows.length
      ? {
          label: reachRows.length === rows.length ? "Reported reach" : "Partial reported reach",
          value: formatNumber(reachRows.reduce((sum, row) => sum + row.reach, 0)),
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

export function chooseReelRows(rows, limit) {
  const hasFlags = rows.some((row) => row.include_in_reel !== null);
  const selected = hasFlags ? rows.filter((row) => row.include_in_reel === true) : rows;
  if (hasFlags && selected.length === 0) fail("include_in_reel is present but no row is marked true");
  if (selected.length > limit) {
    fail(`${selected.length} rows are selected for the reel, above reel_max_items ${limit}; mark at most ${limit} rows include_in_reel=true`);
  }
  return {
    rows: selected,
    rule: hasFlags ? "include_in_reel=true in CSV, preserved in CSV order" : "all CSV rows, preserved in CSV order",
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

function lengthClass(value, longAt, veryLongAt) {
  if (value.length > veryLongAt) return "very-long";
  if (value.length > longAt) return "long";
  return "";
}

function metricCard(metric) {
  return `<article class="metric"><span class="metric-label">${escapeHtml(metric.label)}</span><strong>${escapeHtml(metric.value)}</strong><span class="status status-${escapeHtml(metric.status)}">${escapeHtml(metric.status)}</span><p>${escapeHtml(metric.detail)}</p></article>`;
}

function dashboardHtml({ brand, rows, metrics, brandLogoAsset }) {
  const logo = brandLogoAsset
    ? `<img class="brand-logo" src="${escapeHtml(brandLogoAsset)}" alt="${escapeHtml(brand.name)} logo">`
    : `<span class="brand-wordmark">${escapeHtml(brand.name)}</span>`;
  const articleCards = rows.map((row, index) => `
    <article class="coverage-card">
      <a class="clip-link" href="${escapeHtml(row.source_url)}" target="_blank" rel="noopener noreferrer" aria-label="Open source ${index + 1}">
        <img src="${escapeHtml(row.preview_asset)}" alt="Reviewed press-clip preview for ${escapeHtml(articleName(row))}">
      </a>
      <div class="coverage-copy">
        <div class="coverage-kicker"><span>${String(index + 1).padStart(2, "0")}</span><span>${escapeHtml(row.coverage_type || "earned coverage")}</span></div>
        <h3>${escapeHtml(articleTitle(row))}</h3>
        <p class="outlet">${escapeHtml(row.outlet || "Outlet unavailable")} <span>·</span> ${escapeHtml(row.published_at || "Date unavailable")}</p>
        ${row.byline ? `<p class="byline">By ${escapeHtml(row.byline)}</p>` : ""}
        ${row.note ? `<p class="note">${escapeHtml(row.note)}</p>` : ""}
        <a class="source-url" href="${escapeHtml(row.source_url)}" target="_blank" rel="noopener noreferrer">${escapeHtml(row.source_url)}</a>
        <div class="facts">
          <span>clip review: clean</span>
          <span>logo: verified</span>
          ${row.sentiment ? `<span>sentiment: supplied ${escapeHtml(row.sentiment)}</span>` : `<span>sentiment: unavailable</span>`}
          ${row.reach !== null ? `<a href="${escapeHtml(row.reach_source_url)}" target="_blank" rel="noopener noreferrer">reach: ${escapeHtml(formatNumber(row.reach))} · source</a>` : `<span>reach: unavailable</span>`}
        </div>
      </div>
    </article>`).join("");

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
    .section-head{display:flex;justify-content:space-between;align-items:end;margin:0 0 22px}.section-head h2{font-size:32px;letter-spacing:-.035em;margin:0}.section-head p{margin:0;color:var(--muted);font-size:13px}.coverage-list{display:grid;gap:20px}.coverage-card{display:grid;grid-template-columns:minmax(280px,38%) 1fr;gap:28px;background:white;border:1px solid var(--line);border-radius:22px;padding:18px;box-shadow:0 14px 45px rgba(19,36,58,.08)}.clip-link{display:block;background:#eef1f4;border-radius:13px;overflow:hidden;min-height:300px}.clip-link img{width:100%;height:100%;max-height:440px;object-fit:cover;object-position:top;display:block}.coverage-copy{padding:8px 10px 8px 0}.coverage-kicker{display:flex;justify-content:space-between;text-transform:uppercase;letter-spacing:.12em;font-weight:800;font-size:10px;color:var(--accent)}.coverage-copy h3{font-size:29px;line-height:1.05;letter-spacing:-.035em;margin:26px 0 14px}.outlet,.byline{font-size:13px;color:var(--muted)}.note{font-size:16px;line-height:1.55;border-left:3px solid var(--accent);padding-left:14px;margin:24px 0}.source-url{display:block;overflow-wrap:anywhere;font:12px/1.4 ui-monospace,SFMono-Regular,Menlo,monospace;color:var(--muted);margin:22px 0}.facts{display:flex;gap:7px;flex-wrap:wrap}.facts span,.facts a{border:1px solid var(--line);border-radius:999px;padding:7px 9px;font-size:10px;text-decoration:none;background:var(--paper)}
    .sentiment{margin:50px 0 20px;background:white;border-radius:22px;padding:28px;border:1px solid var(--line)}.sentiment h2{font-size:26px;margin:0 0 8px}.sentiment>p{color:var(--muted);font-size:12px;margin:0 0 24px}.sentiment-row{display:grid;grid-template-columns:90px 1fr 90px;gap:14px;align-items:center;margin:12px 0;text-transform:capitalize;font-size:13px}.bar{height:10px;background:var(--paper);border-radius:999px;overflow:hidden}.bar i{display:block;height:100%;background:var(--accent);border-radius:inherit}.sentiment-row strong{text-align:right;font-size:12px}.footer{display:grid;grid-template-columns:1fr auto;gap:30px;border-top:2px solid var(--ink);margin-top:50px;padding:26px 0;color:var(--muted);font-size:12px;line-height:1.5}.footer strong{color:var(--ink)}
    @media(max-width:900px){.shell{padding:24px}.masthead,.coverage-card{grid-template-columns:1fr}.brand-block{text-align:left}.brand-logo{margin:0}.metrics{grid-template-columns:repeat(2,1fr)}.coverage-copy{padding:8px}.footer{grid-template-columns:1fr}}@media(max-width:560px){.metrics{grid-template-columns:1fr}.shell{padding:16px}h1{font-size:46px}.coverage-copy h3{font-size:24px}}
    @media print{.shell{max-width:none;padding:18mm}.coverage-card,.metric,.sentiment{break-inside:avoid}.coverage-card{box-shadow:none}a{text-decoration:none}}
  </style>
</head>
<body>
  <main class="shell">
    <header class="masthead"><div><div class="eyebrow">Earned media · source-linked</div><h1>${escapeHtml(brand.report_title)}</h1>${brand.subtitle ? `<p class="subtitle">${escapeHtml(brand.subtitle)}</p>` : ""}</div><div class="brand-block">${logo}<span class="period">${escapeHtml(brand.period_label || "Period not supplied")}</span></div></header>
    <section class="metrics">${metricCard(metrics.coverage)}${metricCard(metrics.outlets)}${metricCard(metrics.date_range)}${metricCard(metrics.reach)}</section>
    <div class="section-head"><h2>The coverage</h2><p>Reviewed press-clip previews · open each live source</p></div>
    <section class="coverage-list">${articleCards}</section>
    <section class="sentiment"><h2>Supplied sentiment labels</h2><p><span class="status status-${escapeHtml(metrics.sentiment.status)}">${escapeHtml(metrics.sentiment.status)}</span> ${escapeHtml(metrics.sentiment.detail)}</p>${sentimentRows}</section>
    <footer class="footer"><div><strong>Evidence, not reconstruction.</strong><br>Article visuals are reviewed press-clip previews. Reporting and outlet branding remain the publishers' property.</div><div>${brand.website ? `<a href="${escapeHtml(brand.website)}" target="_blank" rel="noopener noreferrer">${escapeHtml(brand.website)}</a>` : escapeHtml(brand.name)}</div></footer>
  </main>
</body>
</html>`;
}

function reelHtml({ brand, rows, metrics, brandLogoAsset }) {
  const logo = brandLogoAsset
    ? `<img class="reel-logo" src="${escapeHtml(brandLogoAsset)}" alt="${escapeHtml(brand.name)} logo">`
    : `<span class="reel-wordmark">${escapeHtml(brand.name)}</span>`;
  const slides = [];
  slides.push(`<section class="slide title-slide"><div class="accent-rule"></div><div class="title-copy"><span class="overline">Earned media highlights</span><h1 class="${lengthClass(brand.report_title, 56, 95)}">${escapeHtml(brand.report_title)}</h1><p>${escapeHtml(brand.subtitle || brand.period_label || "Source-linked coverage recap")}</p></div><div class="title-brand">${logo}</div></section>`);

  slides.push(`<section class="slide mosaic-slide"><div class="slide-heading"><span>Reviewed coverage</span><strong>${rows.length} source${rows.length === 1 ? "" : "s"}</strong></div><div class="mosaic">${rows.map((row, index) => `<figure style="--i:${index}"><img src="${escapeHtml(row.preview_asset)}" alt=""><figcaption>${escapeHtml(articleName(row))}</figcaption></figure>`).join("")}</div><div class="slide-footer">Real press-clip previews · outlet branding preserved</div></section>`);

  rows.forEach((row, index) => {
    const title = articleTitle(row);
    slides.push(`<section class="slide article-slide"><div class="article-visual"><div class="stack-card one"></div><div class="stack-card two"></div><img src="${escapeHtml(row.preview_asset)}" alt="Reviewed press-clip preview"></div><div class="article-copy"><span class="overline">Coverage ${String(index + 1).padStart(2, "0")} · ${escapeHtml(row.coverage_type || "earned media")}</span><h2 class="${lengthClass(title, 75, 120)}">${escapeHtml(title)}</h2><p class="article-meta">${escapeHtml(row.outlet || "Outlet unavailable")} · ${escapeHtml(row.published_at || "Date unavailable")}</p>${row.note ? `<p class="article-note ${lengthClass(row.note, 150, 230)}">${escapeHtml(row.note)}</p>` : ""}<p class="article-source">${escapeHtml(row.source_url)}</p></div><div class="slide-footer">${escapeHtml(brand.name)} · source linked</div></section>`);
  });

  slides.push(`<section class="slide metrics-slide"><div><span class="overline">Coverage at a glance</span><h2>What the supplied evidence supports</h2></div><div class="metric-grid">${[metrics.coverage, metrics.outlets, metrics.date_range, metrics.reach].map((metric) => `<article><span>${escapeHtml(metric.label)}</span><strong>${escapeHtml(metric.value)}</strong><i>${escapeHtml(metric.status)}</i><p>${escapeHtml(metric.detail)}</p></article>`).join("")}</div><div class="slide-footer">No inferred reach, rank, sentiment, quotes, dates, or bylines</div></section>`);

  for (let offset = 0; offset < rows.length; offset += 3) {
    const group = rows.slice(offset, offset + 3);
    slides.push(`<section class="slide sources-slide"><div><span class="overline">Source links</span><h2>Read the original coverage</h2></div><ol start="${offset + 1}">${group.map((row) => `<li><strong>${escapeHtml(articleName(row))}</strong><span>${escapeHtml(row.source_url)}</span></li>`).join("")}</ol><div class="source-close">${escapeHtml(brand.closing_note || "Thank you to the journalists and outlets behind the reporting.")}</div><div class="slide-footer">${escapeHtml(brand.name)} · ${escapeHtml(brand.period_label || "earned coverage")}</div></section>`);
  }

  return `<!doctype html>
<html lang="en"><head><meta charset="utf-8"><meta http-equiv="Content-Security-Policy" content="default-src 'none'; img-src 'self' data:; style-src 'unsafe-inline'; base-uri 'none'; frame-ancestors 'none'"><title>${escapeHtml(brand.report_title)} reel</title><style>
  :root{--ink:${brand.primary_color};--accent:${brand.accent_color};--paper:${brand.background_color}}*{box-sizing:border-box}html,body{margin:0;background:#1a1a1a;font-family:Arial,Helvetica,sans-serif;color:var(--ink)}.slide{width:1920px;height:1080px;position:relative;overflow:hidden;background:var(--paper);padding:90px 110px;display:flex}.overline{text-transform:uppercase;letter-spacing:.2em;font-size:19px;font-weight:800;color:var(--accent)}.slide-footer{position:absolute;bottom:44px;left:110px;right:110px;padding-top:16px;border-top:2px solid color-mix(in srgb,var(--ink) 22%,transparent);font-size:17px;letter-spacing:.04em}.title-slide{background:var(--ink);color:white;align-items:center}.accent-rule{position:absolute;top:0;bottom:0;left:0;width:25px;background:var(--accent)}.title-copy{max-width:1260px}.title-copy h1{font-size:116px;line-height:.9;letter-spacing:-.065em;margin:34px 0}.title-copy h1.long{font-size:92px}.title-copy h1.very-long{font-size:72px}.title-copy p{font-size:32px;line-height:1.35;color:#cbd4dc;max-width:1050px}.title-brand{position:absolute;right:110px;bottom:72px;background:white;padding:14px 20px;border-radius:16px;line-height:0}.reel-logo{max-width:320px;max-height:100px}.reel-wordmark{font-size:36px;line-height:1;font-weight:900;color:var(--ink)}.mosaic-slide{background:var(--ink);color:white}.slide-heading{position:absolute;top:72px;left:110px;right:110px;display:flex;justify-content:space-between;align-items:center}.slide-heading span{text-transform:uppercase;letter-spacing:.2em;color:var(--accent);font-weight:800}.slide-heading strong{font-size:30px}.mosaic{position:absolute;inset:175px 100px 115px;display:flex;align-items:center;justify-content:center}.mosaic figure{position:absolute;width:440px;height:620px;margin:0;background:white;border:12px solid white;border-radius:18px;overflow:hidden;box-shadow:0 28px 80px rgba(0,0,0,.45);transform:translateX(calc((var(--i) - 2.5)*245px)) rotate(calc((var(--i) - 2.5)*2deg));z-index:calc(10 - var(--i))}.mosaic figure:nth-child(n+7){display:none}.mosaic img{width:100%;height:545px;object-fit:cover;object-position:top}.mosaic figcaption{height:52px;color:var(--ink);font-weight:900;font-size:20px;padding:15px 16px;white-space:nowrap;overflow:hidden;text-overflow:ellipsis}.mosaic-slide .slide-footer{border-color:#4a5b70;color:#b9c4cf}.article-slide{align-items:center;gap:110px;background:linear-gradient(130deg,var(--paper) 0 68%,color-mix(in srgb,var(--accent) 16%,var(--paper)) 68%)}.article-visual{width:690px;height:760px;position:relative;flex:0 0 auto}.article-visual>img{position:absolute;inset:20px 40px 20px 20px;width:630px;height:720px;object-fit:cover;object-position:top;border:14px solid white;border-radius:20px;box-shadow:0 30px 80px rgba(19,36,58,.28);transform:rotate(-2deg)}.stack-card{position:absolute;width:620px;height:700px;border-radius:20px;background:var(--accent);right:0;top:46px;transform:rotate(4deg)}.stack-card.two{background:var(--ink);right:18px;top:25px;transform:rotate(1deg)}.article-copy{max-width:850px}.article-copy h2{font-size:70px;line-height:.98;letter-spacing:-.05em;margin:32px 0 28px}.article-copy h2.long{font-size:56px}.article-copy h2.very-long{font-size:46px}.article-meta{font-size:25px;color:#526378}.article-note{font-size:28px;line-height:1.45;border-left:7px solid var(--accent);padding-left:24px;margin:40px 0}.article-note.long{font-size:24px}.article-note.very-long{font-size:21px}.article-source{font:18px/1.4 ui-monospace,SFMono-Regular,Menlo,monospace;overflow-wrap:anywhere;color:#526378}.metrics-slide{display:block;padding-top:80px}.metrics-slide h2,.sources-slide h2{font-size:64px;line-height:1;margin:20px 0 46px;letter-spacing:-.045em}.metric-grid{display:grid;grid-template-columns:1fr 1fr;gap:18px}.metric-grid article{background:var(--ink);color:white;border-radius:22px;padding:28px 34px;min-height:285px;position:relative}.metric-grid span{text-transform:uppercase;letter-spacing:.14em;font-size:15px;color:#aebbc8}.metric-grid strong{display:block;font-size:58px;line-height:1;margin:25px 0 14px;letter-spacing:-.04em}.metric-grid i{position:absolute;right:28px;top:28px;border:1px solid var(--accent);border-radius:999px;padding:7px 11px;color:var(--accent);font-style:normal;text-transform:uppercase;font-size:12px}.metric-grid p{font-size:17px;line-height:1.4;color:#bdc7d2;max-width:660px}.sources-slide{display:block;background:var(--ink);color:white}.sources-slide ol{margin:0;padding:0;list-style:none;display:grid;gap:18px;counter-reset:item calc(var(--start,1) - 1)}.sources-slide li{background:#203754;border-left:8px solid var(--accent);padding:23px 28px;border-radius:0 14px 14px 0}.sources-slide li strong{display:block;font-size:25px;margin-bottom:10px}.sources-slide li span{display:block;font:17px/1.4 ui-monospace,SFMono-Regular,Menlo,monospace;color:#c8d2dc;overflow-wrap:anywhere}.source-close{margin-top:36px;color:#c8d2dc;font-size:24px}.sources-slide .slide-footer{border-color:#4a5b70;color:#b9c4cf}
  </style></head><body>${slides.map((slide, index) => slide.replace('class="slide ', `data-slide-index="${index}" class="slide `)).join("\n")}</body></html>`;
}

function manifestPayload({ brand, rows, metrics, reel }) {
  return {
    version: 1,
    brand: {
      name: brand.name,
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
      selected_source_urls: reel.rows.map((row) => row.source_url),
      selection_rule: reel.rule,
    },
    coverage: rows.map((row) => ({
      position: row.position,
      source_url: row.source_url,
      source_host: row.source_host,
      outlet: row.outlet || null,
      headline: row.headline || null,
      published_at: row.published_at || null,
      byline: row.byline || null,
      note: row.note || null,
      coverage_type: row.coverage_type || null,
      sentiment: row.sentiment || null,
      reach: row.reach,
      reach_source_url: row.reach_source_url || null,
      clip_section: row.clip_section || null,
      include_in_reel: row.include_in_reel,
      preview_asset: row.preview_asset,
      clip_review: row.clip.review,
    })),
  };
}

export function renderArtifacts({ coverageRows, brand, clips, outDir }) {
  const destination = resolve(outDir);
  const combined = validateClips(clips.raw, coverageRows, clips.baseDir);
  const assetsDir = join(destination, "assets");
  mkdirSync(assetsDir, { recursive: true });

  const rows = combined.map((row) => ({
    ...row,
    preview_asset: copyAsset(row.clip.preview_path, assetsDir, `clip-${String(row.position).padStart(2, "0")}`),
  }));
  const brandLogoAsset = copyAsset(brand.logo_path, assetsDir, "brand-logo");
  const safeBrand = { ...brand, logo_asset: brandLogoAsset };
  const metrics = deriveMetrics(rows);
  const reel = chooseReelRows(rows, brand.reel_max_items);
  const dashboard = dashboardHtml({ brand: safeBrand, rows, metrics, brandLogoAsset });
  const reelDocument = reelHtml({ brand: safeBrand, rows: reel.rows, metrics, brandLogoAsset });
  const manifest = manifestPayload({ brand: safeBrand, rows, metrics, reel });
  const sources = rows.map((row, index) => `${index + 1}. ${articleName(row)}\n${row.source_url}`).join("\n\n");

  writeFileSync(join(destination, "index.html"), dashboard);
  writeFileSync(join(destination, "manifest.json"), `${JSON.stringify(manifest, null, 2)}\n`);
  writeFileSync(join(destination, "sources.txt"), `${sources}\n`);
  writeFileSync(join(destination, "reel.html"), reelDocument);
  return { destination, dashboard, manifest, reelDocument, rows, reelRows: reel.rows };
}

function loadChromium() {
  const bases = [process.cwd(), dirname(MODULE_PATH)];
  if (process.env.PRESS_CLIP_MODULES) bases.push(process.env.PRESS_CLIP_MODULES);
  try {
    bases.push(execFileSync("npm", ["root", "-g"], { encoding: "utf8" }).trim());
  } catch {}
  for (const base of bases) {
    try {
      return createRequire(join(resolve(base), "package.json"))("playwright-core").chromium;
    } catch {}
  }
  fail("Preview/video rendering needs playwright-core. Run: npm i playwright-core");
}

function browserLaunchOptions(chromeArg) {
  const explicit = chromeArg || process.env.COVERAGE_REEL_CHROME || process.env.PRESS_CLIP_CHROME;
  if (explicit) {
    if (!existsSync(explicit)) fail(`Chrome/Edge executable does not exist: ${explicit}`);
    return { executablePath: explicit, headless: true };
  }
  const candidates = [
    "/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
    "/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
    "/usr/bin/google-chrome",
    "/usr/bin/chromium",
    "/usr/bin/chromium-browser",
    "C:\\Program Files\\Google\\Chrome\\Application\\chrome.exe",
    "C:\\Program Files (x86)\\Microsoft\\Edge\\Application\\msedge.exe",
  ];
  const found = candidates.find((candidate) => existsSync(candidate));
  if (!found) fail("Could not find Chrome or Edge. Set COVERAGE_REEL_CHROME or pass --chrome <path>");
  return { executablePath: found, headless: true };
}

async function withBrowser(chromeArg, callback) {
  const browser = await loadChromium().launch(browserLaunchOptions(chromeArg));
  try {
    return await callback(browser);
  } finally {
    await browser.close();
  }
}

export async function renderDashboardPreview(indexPath, outputPath, chromeArg = "") {
  await withBrowser(chromeArg, async (browser) => {
    const page = await browser.newPage({ viewport: { width: 1440, height: 1000 }, deviceScaleFactor: 1 });
    await page.goto(pathToFileURL(indexPath).href, { waitUntil: "load" });
    await page.screenshot({ path: outputPath, fullPage: true, animations: "disabled" });
  });
}

function requireFfmpeg(command) {
  const probe = spawnSync(command, ["-version"], { encoding: "utf8" });
  if (probe.error || probe.status !== 0) {
    fail(`Video rendering needs ffmpeg. Install ffmpeg or pass --ffmpeg <path> (tried: ${command})`);
  }
}

export async function renderVideo({ reelPath, outputPath, framesDir, slideCount, chromeArg = "", ffmpeg = "ffmpeg", secondsPerSlide = 3.5 }) {
  requireFfmpeg(ffmpeg);
  mkdirSync(framesDir, { recursive: true });
  await withBrowser(chromeArg, async (browser) => {
    const page = await browser.newPage({ viewport: { width: 1920, height: 1080 }, deviceScaleFactor: 1 });
    await page.goto(pathToFileURL(reelPath).href, { waitUntil: "load" });
    for (let index = 0; index < slideCount; index += 1) {
      const slide = page.locator(`[data-slide-index="${index}"]`);
      await slide.screenshot({ path: join(framesDir, `slide-${String(index).padStart(3, "0")}.png`), animations: "disabled" });
    }
  });

  const frameRate = `1/${secondsPerSlide}`;
  const result = spawnSync(
    ffmpeg,
    [
      "-hide_banner",
      "-loglevel", "error",
      "-y",
      "-framerate", frameRate,
      "-start_number", "0",
      "-i", join(framesDir, "slide-%03d.png"),
      "-vf", "fps=30,format=yuv420p",
      "-c:v", "libx264",
      "-preset", "medium",
      "-crf", "20",
      "-an",
      "-map_metadata", "-1",
      "-movflags", "+faststart",
      outputPath,
    ],
    { encoding: "utf8" },
  );
  if (result.error || result.status !== 0) {
    fail(`ffmpeg could not render the MP4: ${(result.stderr || result.error?.message || "unknown error").trim()}`);
  }
}

function usage() {
  return `Usage:
  node render.mjs --csv coverage.csv --brand brand.json --clips clips.json --out output [--preview] [--video]

Options:
  --preview                 Write dashboard.png using Chrome/Edge and playwright-core
  --video                   Write a silent 1920x1080 highlight.mp4 and reel frames
  --chrome <path>           Chrome or Edge executable
  --ffmpeg <path>           ffmpeg executable (default: ffmpeg)
  --slide-seconds <number>  Seconds per reel slide (default: 3.5)
  --help                    Show this help`;
}

function parseArgs(argv) {
  const args = {};
  const booleans = new Set(["preview", "video", "help"]);
  const values = new Set(["csv", "brand", "clips", "out", "chrome", "ffmpeg", "slide-seconds"]);
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
  const slideSeconds = args["slide-seconds"] === undefined ? 3.5 : Number(args["slide-seconds"]);
  if (!Number.isFinite(slideSeconds) || slideSeconds < 1 || slideSeconds > 10) {
    fail("--slide-seconds must be a number from 1 to 10");
  }

  const csvPath = resolve(args.csv);
  const brandPath = resolve(args.brand);
  const clipsPath = resolve(args.clips);
  const coverageRows = validateCoverage(parseCsv(readFileSync(csvPath, "utf8")));
  const brand = validateBrand(JSON.parse(readFileSync(brandPath, "utf8")), dirname(brandPath));
  const clips = { raw: JSON.parse(readFileSync(clipsPath, "utf8")), baseDir: dirname(clipsPath) };
  const result = renderArtifacts({ coverageRows, brand, clips, outDir: args.out });

  if (args.preview || args.video) {
    await renderDashboardPreview(join(result.destination, "index.html"), join(result.destination, "dashboard.png"), args.chrome);
  }
  if (args.video) {
    await renderVideo({
      reelPath: join(result.destination, "reel.html"),
      outputPath: join(result.destination, "highlight.mp4"),
      framesDir: join(result.destination, "reel-frames"),
      slideCount: result.reelRows.length + Math.ceil(result.reelRows.length / 3) + 3,
      chromeArg: args.chrome,
      ffmpeg: args.ffmpeg || "ffmpeg",
      secondsPerSlide: slideSeconds,
    });
  }

  console.log(`Coverage dashboard written to ${join(result.destination, "index.html")}`);
  console.log(`Metrics: coverage=${result.rows.length}; reach=${result.manifest.metrics.reach.status}; sentiment=${result.manifest.metrics.sentiment.status}`);
  if (args.preview || args.video) console.log(`Dashboard preview written to ${join(result.destination, "dashboard.png")}`);
  if (args.video) console.log(`Highlight reel written to ${join(result.destination, "highlight.mp4")} (${result.reelRows.length} article${result.reelRows.length === 1 ? "" : "s"})`);
}

if (process.argv[1] && resolve(process.argv[1]) === resolve(MODULE_PATH)) {
  main().catch((error) => {
    const prefix = error instanceof InputError ? "Input error" : "Error";
    console.error(`${prefix}: ${error.message}`);
    process.exitCode = 1;
  });
}
