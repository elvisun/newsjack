import {
  lstatSync,
  readFileSync,
  readdirSync,
  realpathSync,
} from "node:fs";
import { isAbsolute, join, relative, resolve } from "node:path";

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
if (manifest.name !== newsjack.name) {
  fail(
    `manifest name ${JSON.stringify(manifest.name)} does not match marketplace name ${JSON.stringify(newsjack.name)}`,
  );
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
  if (!entry.isSymbolicLink()) {
    continue;
  }
  const linkPath = join(pluginRoot, entry.name);
  let target;
  try {
    target = realpathSync(linkPath);
  } catch (error) {
    fail(`top-level symlink ${entry.name} cannot be resolved: ${error.message}`);
  }
  assertInsideRepo(target, `top-level symlink ${entry.name}`);
}

console.log(
  `Claude hosted plugin package is valid: ${relative(repoRoot, pluginRoot)} (no top-level bin)`,
);
