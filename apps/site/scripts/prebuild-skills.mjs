import { readdirSync, readFileSync, writeFileSync } from "node:fs";
import { join, resolve } from "node:path";

const siteRoot = resolve(import.meta.dirname, "..");
const repoRoot = resolve(siteRoot, "../..");
const skillsDir = join(repoRoot, "skills");
const outFile = join(siteRoot, "lib", "skills-data.json");

const CATEGORIES = {
  "pr-strategist": "Strategize",
  "pr-calendar": "Strategize",
  "newsworthiness-check": "Strategize",
  "angle-generator": "Act",
  "headline-generator": "Act",
  "meanest-editor": "Act",
  "crisis-holding": "Act",
  "reactive-comment": "Act",
  "fact-check": "Act",
  "journalist-fit-check": "Act",
  "same-outlet-ranker": "Act",
  "voice-extractor": "Act",
  "find-journalists": "Act",
  "press-clip": "Act",
  "news-search": "Detect",
  "story-origin-check": "Detect",
  "relevance-coarse-filter": "Detect",
  "newsjack-triage": "Detect",
  "newsjack-detector": "Detect",
  "newsjack-monitor-setup": "Detect",
  "coverage-tracker": "Detect",
  "coverage-tracker-setup": "Detect",
};

function parseFrontmatter(content) {
  const match = content.match(/^---\n([\s\S]*?)\n---/);
  if (!match) return null;
  const fields = {};
  for (const line of match[1].split("\n")) {
    const m = line.match(/^(\w[\w_-]*):\s*"?(.*?)"?\s*$/);
    if (m) fields[m[1]] = m[2];
  }
  return fields;
}

const entries = readdirSync(skillsDir, { withFileTypes: true })
  .filter((d) => d.isDirectory())
  .map((d) => {
    const skillMd = join(skillsDir, d.name, "SKILL.md");
    try {
      const raw = readFileSync(skillMd, "utf-8");
      const fm = parseFrontmatter(raw);
      if (!fm?.name) return null;
      return {
        name: fm.name,
        description: fm.description || "",
        whenToUse: fm.when_to_use || "",
        category: CATEGORIES[fm.name] || "Other",
      };
    } catch {
      return null;
    }
  })
  .filter(Boolean);

const order = ["Strategize", "Act", "Detect"];
entries.sort(
  (a, b) => order.indexOf(a.category) - order.indexOf(b.category) || a.name.localeCompare(b.name),
);

writeFileSync(outFile, JSON.stringify(entries, null, 2));
console.log(`Wrote ${entries.length} skills to lib/skills-data.json`);
