import assert from "node:assert/strict";
import { mkdtempSync, readFileSync, writeFileSync } from "node:fs";
import { tmpdir } from "node:os";
import { dirname, join, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "node:test";

import {
  InputError,
  chooseReelRows,
  deriveMetrics,
  parseCsv,
  renderArtifacts,
  validateBrand,
  validateClips,
  validateCoverage,
} from "../../skills/coverage-reel/render.mjs";

const fixtureRoot = dirname(fileURLToPath(import.meta.url));

function fixtureInputs() {
  const coverageRows = validateCoverage(parseCsv(readFileSync(join(fixtureRoot, "coverage.csv"), "utf8")));
  const brand = validateBrand(JSON.parse(readFileSync(join(fixtureRoot, "brand.json"), "utf8")), fixtureRoot);
  const clips = { raw: JSON.parse(readFileSync(join(fixtureRoot, "clips.json"), "utf8")), baseDir: fixtureRoot };
  return { coverageRows, brand, clips };
}

test("CSV parsing preserves quoted commas and newlines", () => {
  const rows = parseCsv('source_url,headline,note\r\nhttps://example.test/a,"A, B","line one\nline two"\r\n');
  assert.deepEqual(rows, [{ source_url: "https://example.test/a", headline: "A, B", note: "line one\nline two" }]);
});

test("coverage validation rejects unsafe URLs and unsourced reach", () => {
  assert.throws(
    () => validateCoverage([{ source_url: "javascript:alert(1)" }]),
    (error) => error instanceof InputError && /http or https/.test(error.message),
  );
  assert.throws(
    () => validateCoverage([{ source_url: "https://example.test/a", reach: "1200" }]),
    /reach requires reach_source_url/,
  );
  assert.throws(
    () => validateCoverage([{ source_url: "https://example.test/a", outlet_rank: "1" }]),
    /unsupported column\(s\): outlet_rank/,
  );
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
  assert.equal(metrics.sentiment.counts.positive, 1);

  const missing = deriveMetrics(validateCoverage([{ source_url: "https://example.test/c" }]));
  assert.equal(missing.outlets.status, "unavailable");
  assert.equal(missing.date_range.status, "unavailable");
  assert.equal(missing.reach.status, "unavailable");
  assert.equal(missing.sentiment.status, "unavailable");
});

test("clip manifest enforces the press-clip review gate", () => {
  const { coverageRows, clips } = fixtureInputs();
  const unsafe = structuredClone(clips.raw);
  unsafe.clips[0].review.logo_verified = false;
  assert.throws(() => validateClips(unsafe, coverageRows, clips.baseDir), /logo_verified must be true/);

  const noPreview = structuredClone(clips.raw);
  noPreview.clips[0].preview_path = "";
  assert.throws(() => validateClips(noPreview, coverageRows, clips.baseDir), /preview_path is required/);

  const noPdf = structuredClone(clips.raw);
  noPdf.clips[0].pdf_path = "";
  assert.throws(() => validateClips(noPdf, coverageRows, clips.baseDir), /pdf_path is required/);

  const missing = structuredClone(clips.raw);
  missing.clips.pop();
  assert.throws(() => validateClips(missing, coverageRows, clips.baseDir), /missing 1 CSV source/);
});

test("renderer escapes supplied text, copies only local assets, and is deterministic", () => {
  const inputs = fixtureInputs();
  inputs.coverageRows[0] = {
    ...inputs.coverageRows[0],
    headline: '<script>alert("x")</script>',
    note: '<img src=x onerror="alert(1)">',
  };
  const first = mkdtempSync(join(tmpdir(), "coverage-reel-a-"));
  const second = mkdtempSync(join(tmpdir(), "coverage-reel-b-"));
  const a = renderArtifacts({ ...inputs, outDir: first });
  const b = renderArtifacts({ ...inputs, outDir: second });

  assert.equal(a.dashboard, b.dashboard);
  assert.deepEqual(a.manifest, b.manifest);
  assert.doesNotMatch(a.dashboard, /<script>alert/);
  assert.match(a.dashboard, /&lt;script&gt;alert\(&quot;x&quot;\)&lt;\/script&gt;/);
  assert.match(a.dashboard, /default-src 'none'/);
  assert.match(a.dashboard, /assets\/clip-01-[a-f0-9]{12}\.svg/);
  assert.equal(readFileSync(join(first, "manifest.json"), "utf8"), readFileSync(join(second, "manifest.json"), "utf8"));
});

test("renderer leaves missing facts unavailable and writes source attribution", () => {
  const work = mkdtempSync(join(tmpdir(), "coverage-reel-missing-"));
  const csv = validateCoverage([{ source_url: "https://example.test/missing" }]);
  const preview = resolve(fixtureRoot, "assets/northstar-daily.svg");
  const pdf = resolve(fixtureRoot, "assets/synthetic-clip.pdf");
  const clips = {
    baseDir: fixtureRoot,
    raw: {
      version: 1,
      clips: [{
        source_url: "https://example.test/missing",
        preview_path: preview,
        pdf_path: pdf,
        review: {
          verdict: "clean",
          logo_verified: true,
          pdf_reviewed: true,
          scope: "whole",
          reviewed_at: "2026-10-10T14:30:00Z",
        },
      }],
    },
  };
  const brand = validateBrand({ name: "Example" });
  const result = renderArtifacts({ coverageRows: csv, brand, clips, outDir: work });
  assert.match(result.dashboard, /Headline unavailable/);
  assert.match(result.dashboard, /Outlet unavailable/);
  assert.match(result.dashboard, /reach: unavailable/);
  assert.match(readFileSync(join(work, "sources.txt"), "utf8"), /https:\/\/example\.test\/missing/);
});

test("reel selection never silently chooses a subset", () => {
  const rows = Array.from({ length: 7 }, (_, index) => ({ position: index + 1, include_in_reel: null }));
  assert.throws(() => chooseReelRows(rows, 6), /mark at most 6 rows/);
  const selected = rows.map((row, index) => ({ ...row, include_in_reel: index < 2 }));
  assert.deepEqual(chooseReelRows(selected, 6).rows.map((row) => row.position), [1, 2]);
});

test("brand validation rejects remote logo paths and unsafe colors", () => {
  assert.throws(() => validateBrand({ name: "Example", logo_path: "https://example.test/logo.svg" }), /local path/);
  assert.throws(() => validateBrand({ name: "Example", accent_color: "red; background:url(x)" }), /hex color/);
});

test("brand validation rejects active SVG assets", () => {
  const work = mkdtempSync(join(tmpdir(), "coverage-reel-svg-"));
  const unsafeSvg = join(work, "unsafe.svg");
  writeFileSync(unsafeSvg, '<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>');
  assert.throws(() => validateBrand({ name: "Example", logo_path: unsafeSvg }), /active or remote SVG content/);
});
