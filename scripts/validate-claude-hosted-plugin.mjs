import {
  lstatSync,
  readFileSync,
  readdirSync,
  statSync,
} from "node:fs";
import { isAbsolute, join, relative, resolve } from "node:path";
import { spawnSync } from "node:child_process";

const repoRoot = resolve(process.argv[2] || resolve(import.meta.dirname, ".."));
const marketplacePath = join(repoRoot, ".claude-plugin", "marketplace.json");

function fail(message) {
  console.error(`Claude hosted plugin validation failed: ${message}`);
  process.exit(1);
}

function readJSON(path, label) {
  try {
    return JSON.parse(readFileSync(path, "utf8"));
  } catch (error) {
    fail(`${label} is not valid JSON: ${error.message}`);
  }
}

function assertInsideRepo(path, label) {
  const rel = relative(repoRoot, path);
  if (rel === "" || (!rel.startsWith("..") && !isAbsolute(rel))) {
    return;
  }
  fail(`${label} resolves outside the marketplace repository`);
}

function regularTreeEntries(root, label, prefix = "") {
  const entries = [];
  for (const entry of readdirSync(root, { withFileTypes: true }).sort((a, b) =>
    a.name.localeCompare(b.name),
  )) {
    const relativePath = prefix ? join(prefix, entry.name) : entry.name;
    const path = join(root, entry.name);
    const stat = lstatSync(path);
    if (stat.isSymbolicLink()) {
      fail(`${label} contains a symlink: ${relativePath}`);
    }
    if (stat.isDirectory()) {
      entries.push({ path: relativePath, type: "directory" });
      entries.push(...regularTreeEntries(path, label, relativePath));
      continue;
    }
    if (!stat.isFile()) {
      fail(`${label} contains a non-regular file: ${relativePath}`);
    }
    entries.push({ path: relativePath, type: "file" });
  }
  return entries;
}

const marketplace = readJSON(marketplacePath, ".claude-plugin/marketplace.json");
const plugins = Array.isArray(marketplace.plugins) ? marketplace.plugins : [];
const newsjack = plugins.find((plugin) => plugin?.name === "newsjack");

if (!newsjack) {
  fail('marketplace has no plugin named "newsjack"');
}
if (typeof newsjack.source !== "string" || !newsjack.source.startsWith("./")) {
  fail('newsjack must use a repository-relative source such as "./plugins/newsjack"');
}
if (newsjack.source === "./") {
  fail("newsjack cannot package the marketplace repository root");
}

const pluginRoot = resolve(repoRoot, newsjack.source);
assertInsideRepo(pluginRoot, "newsjack source");

try {
  if (!lstatSync(pluginRoot).isDirectory()) {
    fail(`newsjack source is not a directory: ${newsjack.source}`);
  }
} catch (error) {
  fail(`cannot read newsjack source ${newsjack.source}: ${error.message}`);
}

const manifestPath = join(pluginRoot, ".claude-plugin", "plugin.json");
const manifest = readJSON(manifestPath, "newsjack plugin manifest");
if (Object.hasOwn(manifest, "$schema")) {
  fail(
    'manifest contains "$schema"; claude.ai strips this unrecognized top-level key',
  );
}
if (manifest.name !== newsjack.name) {
  fail(
    `manifest name ${JSON.stringify(manifest.name)} does not match marketplace name ${JSON.stringify(newsjack.name)}`,
  );
}

const readFrontmatterField = (body, field, label) => {
  const match = body.match(new RegExp(`^${field}:\\s*(.+)$`, "m"));
  if (!match) {
    fail(`${label} has no ${field} frontmatter field`);
  }
  const raw = match[1].trim();
  if (!raw.startsWith('"')) {
    return raw;
  }
  try {
    return JSON.parse(raw);
  } catch (error) {
    fail(`${label} has invalid quoted ${field}: ${error.message}`);
  }
};

const sourceMcpPath = join(repoRoot, ".mcp.json");
const packagedMcpPath = join(pluginRoot, ".mcp.json");
if (!lstatSync(packagedMcpPath).isFile()) {
  fail("newsjack .mcp.json must be a regular file, not a symlink");
}
if (!readFileSync(sourceMcpPath).equals(readFileSync(packagedMcpPath))) {
  fail("newsjack .mcp.json differs from the repository .mcp.json; run the package sync script");
}

const sourceSkillsRoot = join(repoRoot, "skills");
const skillsRoot = join(pluginRoot, "skills");
if (!lstatSync(skillsRoot).isDirectory()) {
  fail("newsjack skills must be a regular directory, not a symlink");
}
const sourceSkillEntries = regularTreeEntries(sourceSkillsRoot, "skills source");
const packagedSkillEntries = regularTreeEntries(skillsRoot, "packaged skills");
if (JSON.stringify(sourceSkillEntries) !== JSON.stringify(packagedSkillEntries)) {
  fail("packaged skills file tree differs from skills/; run the package sync script");
}
for (const entry of sourceSkillEntries) {
  if (entry.type !== "file") {
    continue;
  }
  const source = readFileSync(join(sourceSkillsRoot, entry.path));
  const packaged = readFileSync(join(skillsRoot, entry.path));
  if (!source.equals(packaged)) {
    fail(`packaged skill file differs from skills/: ${entry.path}`);
  }
}

const skillCatalog = readdirSync(skillsRoot, { withFileTypes: true })
  .filter((entry) => entry.isDirectory())
  .map((entry) => {
    const skillPath = join(skillsRoot, entry.name, "SKILL.md");
    try {
      if (!statSync(skillPath).isFile()) {
        fail(`${entry.name} has no SKILL.md`);
      }
    } catch (error) {
      fail(`${entry.name} has no readable SKILL.md: ${error.message}`);
    }
    const body = readFileSync(skillPath, "utf8");
    const name = readFrontmatterField(body, "name", entry.name);
    if (name !== entry.name) {
      fail(`${entry.name}/SKILL.md declares name ${JSON.stringify(name)}`);
    }
    return {
      name,
      description: readFrontmatterField(body, "description", entry.name),
    };
  })
  .sort((a, b) => a.name.localeCompare(b.name));

const pluginReadme = readFileSync(join(pluginRoot, "README.md"), "utf8");
const catalogMatch = pluginReadme.match(
  /<!-- BEGIN SKILL CATALOG -->\n([\s\S]*?)\n<!-- END SKILL CATALOG -->/,
);
if (!catalogMatch) {
  fail("README.md is missing the checked skill catalog markers");
}
const documentedSkills = [...catalogMatch[1].matchAll(/^- \*\*`([^`]+)`\*\* — (.+)$/gm)]
  .map((match) => ({ name: match[1], description: match[2] }))
  .sort((a, b) => a.name.localeCompare(b.name));
if (JSON.stringify(documentedSkills) !== JSON.stringify(skillCatalog)) {
  fail(
    `README.md skill catalog does not match ${skillCatalog.length} skill directories and their frontmatter`,
  );
}
const expectedDescriptionPrefix = `${skillCatalog.length} open-source`;
if (!marketplace.description?.includes(expectedDescriptionPrefix)) {
  fail(`marketplace summary must include ${JSON.stringify(expectedDescriptionPrefix)}`);
}
if (!manifest.description?.startsWith(expectedDescriptionPrefix)) {
  fail(`plugin description must start with ${JSON.stringify(expectedDescriptionPrefix)}`);
}
if (!newsjack.description?.startsWith(expectedDescriptionPrefix)) {
  fail(`marketplace description must start with ${JSON.stringify(expectedDescriptionPrefix)}`);
}
if (manifest.icon !== "./icon.svg") {
  fail('plugin manifest icon must be "./icon.svg"');
}
if (!statSync(join(pluginRoot, "icon.svg")).isFile()) {
  fail("plugin manifest icon does not resolve to a file inside the plugin");
}
if (manifest.privacyPolicyUrl !== "https://newsjack.sh/privacy") {
  fail("plugin manifest privacyPolicyUrl must point to the Newsjack privacy notice");
}
for (const manifestOnlyField of ["icon", "privacyPolicyUrl"]) {
  if (Object.hasOwn(newsjack, manifestOnlyField)) {
    fail(`${manifestOnlyField} belongs in plugin.json, not the marketplace entry`);
  }
}

try {
  lstatSync(join(pluginRoot, "bin"));
  fail(
    "newsjack contains a top-level bin path; claude.ai-hosted plugins may not ship PATH executables",
  );
} catch (error) {
  if (error.code !== "ENOENT") {
    throw error;
  }
}

for (const entry of readdirSync(pluginRoot, { withFileTypes: true })) {
  if (entry.isSymbolicLink()) {
    fail(`newsjack package contains a top-level symlink: ${entry.name}`);
  }
}

console.log(
  `Claude hosted plugin package is valid: ${relative(repoRoot, pluginRoot)} (${skillCatalog.length} documented skills, no top-level bin)`,
);

const claudeVersion = spawnSync("claude", ["--version"], { encoding: "utf8" });
if (claudeVersion.error?.code === "ENOENT") {
  console.log("Claude CLI not found; skipped claude plugin validate");
} else if (claudeVersion.error || claudeVersion.status !== 0) {
  fail(
    `Claude CLI is present but unavailable: ${claudeVersion.error?.message || claudeVersion.stderr.trim()}`,
  );
} else {
  const claudeValidation = spawnSync(
    "claude",
    ["plugin", "validate", pluginRoot, "--profile", "directory"],
    { stdio: "inherit" },
  );
  if (claudeValidation.error || claudeValidation.status !== 0) {
    fail(
      `claude plugin validate failed${claudeValidation.error ? `: ${claudeValidation.error.message}` : ""}`,
    );
  }
}
