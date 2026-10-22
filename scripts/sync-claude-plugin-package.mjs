import {
  copyFileSync,
  cpSync,
  lstatSync,
  readFileSync,
  readdirSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { join, resolve } from "node:path";

const repoRoot = resolve(process.argv[2] || resolve(import.meta.dirname, ".."));
const sourceSkillsRoot = join(repoRoot, "skills");
const pluginRoot = join(repoRoot, "plugins", "newsjack");
const packagedSkillsRoot = join(pluginRoot, "skills");

function assertRegularTree(root, label) {
  for (const entry of readdirSync(root, { withFileTypes: true })) {
    const path = join(root, entry.name);
    if (entry.isSymbolicLink()) {
      throw new Error(`${label} contains a symlink: ${path}`);
    }
    if (entry.isDirectory()) {
      assertRegularTree(path, label);
    }
  }
}

function readFrontmatterField(body, field, label) {
  const match = body.match(new RegExp(`^${field}:\\s*(.+)$`, "m"));
  if (!match) {
    throw new Error(`${label} has no ${field} frontmatter field`);
  }
  const raw = match[1].trim();
  return raw.startsWith('"') ? JSON.parse(raw) : raw;
}

assertRegularTree(sourceSkillsRoot, "skills source");

rmSync(packagedSkillsRoot, { force: true, recursive: true });
cpSync(sourceSkillsRoot, packagedSkillsRoot, { recursive: true });

const packagedMcpPath = join(pluginRoot, ".mcp.json");
rmSync(packagedMcpPath, { force: true });
copyFileSync(join(repoRoot, ".mcp.json"), packagedMcpPath);

const skillCatalog = readdirSync(sourceSkillsRoot, { withFileTypes: true })
  .filter((entry) => entry.isDirectory())
  .map((entry) => {
    const body = readFileSync(join(sourceSkillsRoot, entry.name, "SKILL.md"), "utf8");
    return {
      name: readFrontmatterField(body, "name", entry.name),
      description: readFrontmatterField(body, "description", entry.name),
    };
  })
  .sort((a, b) => a.name.localeCompare(b.name));

const catalogLines = skillCatalog
  .map(({ name, description }) => `- **\`${name}\`** — ${description}`)
  .join("\n");
const readmePath = join(pluginRoot, "README.md");
const readme = readFileSync(readmePath, "utf8")
  .replace(
    /<!-- BEGIN SKILL CATALOG -->\n[\s\S]*?\n<!-- END SKILL CATALOG -->/,
    `<!-- BEGIN SKILL CATALOG -->\n${catalogLines}\n<!-- END SKILL CATALOG -->`,
  )
  .replace(
    /Newsjack is an open-source plugin that gives an agent \d+ focused/,
    `Newsjack is an open-source plugin that gives an agent ${skillCatalog.length} focused`,
  );
writeFileSync(readmePath, readme);

if (!lstatSync(packagedMcpPath).isFile() || !lstatSync(packagedSkillsRoot).isDirectory()) {
  throw new Error("plugin package sync did not create regular MCP and skills paths");
}

console.log(
  `Synced .mcp.json and ${skillCatalog.length} skills into plugins/newsjack`,
);
