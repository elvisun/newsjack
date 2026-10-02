import { readdirSync, readFileSync, writeFileSync } from "node:fs";
import { join, resolve } from "node:path";

// Everything the site shows about a skill comes from the skill itself:
// frontmatter for name, description, when_to_use, and metadata.category,
// and the SKILL.md H1 for the display title. No site-only copy.

const siteRoot = resolve(import.meta.dirname, "..");
const repoRoot = resolve(siteRoot, "../..");
const skillsDir = join(repoRoot, "skills");
const outFile = join(siteRoot, "lib", "skills-data.json");

const CATEGORY_ORDER = ["Detect", "Act", "Strategize"];

function unquote(value) {
  return value.trim().replace(/^"(.*)"$/, "$1");
}

// Top-level `key: value` pairs plus one level of nesting (`metadata:` maps).
function parseFrontmatter(content) {
  const match = content.match(/^---\n([\s\S]*?)\n---/);
  if (!match) return null;
  const fields = {};
  let parent = null;
  for (const line of match[1].split("\n")) {
    const nested = line.match(/^\s+(\w[\w_-]*):\s*(.*)$/);
    if (nested && parent) {
      fields[parent][nested[1]] = unquote(nested[2]);
      continue;
    }
    const top = line.match(/^(\w[\w_-]*):\s*(.*)$/);
    if (!top) continue;
    if (top[2] === "") {
      parent = top[1];
      fields[parent] = {};
    } else {
      parent = null;
      fields[top[1]] = unquote(top[2]);
    }
  }
  return fields;
}

function headingTitle(content, name) {
  const heading = content.match(/^# (.+)$/m)?.[1]?.trim();
  if (heading && heading !== name) return heading;
  return name.replace(/-/g, " ").replace(/\b\w/g, (c) => c.toUpperCase());
}

const entries = readdirSync(skillsDir, { withFileTypes: true })
  .filter((d) => d.isDirectory())
  .map((d) => {
    const skillMd = join(skillsDir, d.name, "SKILL.md");
    try {
      const raw = readFileSync(skillMd, "utf-8");
      const fm = parseFrontmatter(raw);
      if (!fm?.name) return null;
      const category = fm.metadata?.category;
      if (!category) {
        console.warn(`prebuild-skills: ${fm.name} has no metadata.category; listing it under "More"`);
      }
      return {
        name: fm.name,
        title: headingTitle(raw, fm.name),
        description: fm.description || "",
        whenToUse: fm.when_to_use || "",
        category: category || "More",
      };
    } catch {
      return null;
    }
  })
  .filter(Boolean);

const rank = (category) => {
  const i = CATEGORY_ORDER.indexOf(category);
  return i === -1 ? CATEGORY_ORDER.length : i;
};
entries.sort(
  (a, b) => rank(a.category) - rank(b.category) || a.name.localeCompare(b.name),
);

writeFileSync(outFile, JSON.stringify(entries, null, 2));
console.log(`Wrote ${entries.length} skills to lib/skills-data.json`);
