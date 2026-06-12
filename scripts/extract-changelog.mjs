// Extract the CHANGELOG.md section for a release tag, for use as the
// GitHub Release body. The release workflow fails if no usable section
// exists, which is what enforces "no release without release notes".
//
// Rules:
// - An exact `## <tag>` section always wins.
// - Prerelease tags (containing "-") may fall back to a non-empty
//   `## Unreleased` section.
// - Stable tags require their own section: rename Unreleased before tagging.
//
// Usage: node scripts/extract-changelog.mjs v0.1.11 [path/to/CHANGELOG.md]

import { readFileSync } from "node:fs";

const tag = (process.argv[2] || "").trim();
const changelogPath = process.argv[3] || "CHANGELOG.md";

if (!tag) {
  console.error("usage: node scripts/extract-changelog.mjs <tag> [changelog]");
  process.exit(2);
}

let body;
try {
  body = readFileSync(changelogPath, "utf8");
} catch {
  console.error(`${changelogPath} not found; every release needs a changelog entry.`);
  process.exit(1);
}

function extractSection(name) {
  const lines = body.split("\n");
  const matchesHeading = (line) => {
    if (!line.startsWith("## ")) {
      return false;
    }
    const heading = line.slice(3).trim();
    if (heading.toLowerCase() === name.toLowerCase()) {
      return true;
    }
    // Allow a date or annotation suffix: "## v0.1.10 — 2026-06-12".
    // The boundary check keeps "## v0.1.10" from matching "v0.1.10-rc.1".
    if (!heading.toLowerCase().startsWith(name.toLowerCase())) {
      return false;
    }
    const next = heading[name.length];
    return next === " " || next === "\t";
  };

  const start = lines.findIndex(matchesHeading);
  if (start === -1) {
    return null;
  }
  let end = lines.length;
  for (let i = start + 1; i < lines.length; i++) {
    if (lines[i].startsWith("## ")) {
      end = i;
      break;
    }
  }
  const content = lines.slice(start + 1, end).join("\n").trim();
  return content === "" ? null : content;
}

let notes = extractSection(tag);
let source = `the \`## ${tag}\` section`;

if (!notes && tag.includes("-")) {
  notes = extractSection("Unreleased");
  source = "the `## Unreleased` section (prerelease fallback)";
}

if (!notes) {
  console.error(
    [
      `No release notes for ${tag} in ${changelogPath}.`,
      "",
      tag.includes("-")
        ? `Add bullets under \`## Unreleased\` (prereleases use it automatically), or add a \`## ${tag}\` section.`
        : `Stable releases need their own section: rename \`## Unreleased\` to \`## ${tag} — YYYY-MM-DD\` and start a fresh Unreleased.`,
      "",
      "Then push the tag again.",
    ].join("\n"),
  );
  process.exit(1);
}

console.error(`release notes: using ${source}`);
process.stdout.write(notes + "\n");
