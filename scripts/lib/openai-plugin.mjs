import { execFileSync } from "node:child_process";
import {
  cpSync,
  lstatSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
  readdirSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { basename, isAbsolute, join, relative, resolve, sep } from "node:path";
import { isDeepStrictEqual } from "node:util";

export const PLUGIN_SCHEMA_URL =
  "https://agent-plugins.org/schemas/1.0.0/plugin.schema.json";
export const MCP_SCHEMA_URL =
  "https://agent-plugins.org/schemas/1.0.0/mcp.schema.json";
export const MAX_ARCHIVE_BYTES = 100 * 1024 * 1024;
export const MAX_ARCHIVE_ENTRIES = 5_000;
export const MAX_MEMBER_BYTES = 100 * 1024 * 1024;
export const MAX_UNCOMPRESSED_BYTES = 512 * 1024 * 1024;
export const MAX_IMAGE_BYTES = 5 * 1024 * 1024;

const RELEASE_VERSION_PATTERN =
  /^v?(\d+\.\d+\.\d+)(?:-([0-9A-Za-z][0-9A-Za-z.-]*))?(?:\+([0-9A-Za-z][0-9A-Za-z.-]*))?$/;
const SEMVER_PATTERN =
  /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$/;
const PLUGIN_NAME_PATTERN = /^[A-Za-z0-9][A-Za-z0-9_-]*$/;
const ALLOWED_CATEGORIES = new Set([
  "Productivity",
  "Creativity",
  "Developer Tools",
  "Business & Operations",
  "Data & Analytics",
  "Communication",
  "Education & Research",
  "Security",
  "Finance",
  "Healthcare",
  "Travel",
  "Entertainment",
  "Other",
]);
const FORBIDDEN_PATH_PARTS = new Set([
  ".git",
  ".github",
  ".DS_Store",
  "Thumbs.db",
  "node_modules",
  "eval",
  "fixtures",
  "runs",
  "credentials",
  "secrets",
]);

function invariant(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

function readJSON(path, label) {
  try {
    return JSON.parse(readFileSync(path, "utf8"));
  } catch (error) {
    throw new Error(`${label} is not valid JSON: ${error.message}`);
  }
}

function clone(value) {
  return JSON.parse(JSON.stringify(value));
}

function parseFrontmatterField(body, field, label) {
  invariant(body.startsWith("---\n"), `${label} must start with YAML frontmatter`);
  const end = body.indexOf("\n---", 4);
  invariant(end !== -1, `${label} has unclosed YAML frontmatter`);
  const frontmatter = body.slice(4, end);
  const match = frontmatter.match(new RegExp(`^${field}:\\s*(.+)$`, "m"));
  invariant(match, `${label} has no ${field} frontmatter field`);
  const raw = match[1].trim();
  if (raw.startsWith('"')) {
    try {
      return JSON.parse(raw);
    } catch (error) {
      throw new Error(`${label} has invalid quoted ${field}: ${error.message}`);
    }
  }
  return raw;
}

function walkRegularFiles(root, label, prefix = "") {
  const files = [];
  for (const entry of readdirSync(root, { withFileTypes: true }).sort((a, b) =>
    a.name.localeCompare(b.name),
  )) {
    const relativePath = prefix ? join(prefix, entry.name) : entry.name;
    const path = join(root, entry.name);
    const stat = lstatSync(path);
    invariant(!stat.isSymbolicLink(), `${label} contains a symlink: ${relativePath}`);
    if (stat.isDirectory()) {
      files.push(...walkRegularFiles(path, label, relativePath));
      continue;
    }
    invariant(stat.isFile(), `${label} contains a non-regular file: ${relativePath}`);
    assertSafeSourcePath(relativePath, label);
    files.push(relativePath);
  }
  return files;
}

function assertSafeSourcePath(path, label) {
  const normalized = path.split(sep).join("/");
  const parts = normalized.split("/");
  for (const part of parts) {
    invariant(!FORBIDDEN_PATH_PARTS.has(part), `${label} contains forbidden path ${path}`);
    invariant(!part.startsWith(".env"), `${label} contains forbidden path ${path}`);
    invariant(!/\.(?:key|pem|p12|pfx|token|secret|credential)$/i.test(part), `${label} contains secret-like path ${path}`);
  }
}

function assertInside(root, path, label) {
  const rel = relative(root, path);
  invariant(
    rel === "" || (!rel.startsWith(`..${sep}`) && rel !== ".." && !isAbsolute(rel)),
    `${label} resolves outside ${root}`,
  );
}

export function normalizeReleaseVersion(value) {
  invariant(typeof value === "string" && value.length > 0, "release version is required");
  const match = value.match(RELEASE_VERSION_PATTERN);
  invariant(match, `release version ${JSON.stringify(value)} is not a supported v* semver`);
  return {
    releaseVersion: value,
    coreVersion: match[1],
  };
}

export function openAIArchiveName(releaseVersion) {
  normalizeReleaseVersion(releaseVersion);
  return `newsjack-openai-plugin-${releaseVersion}.zip`;
}

function conciseDescription(description) {
  let concise = description
    .replace(/^\d+\s+open-source\s+/i, "")
    .replace(/\s+for\s+.+$/s, "")
    .trim();
  if (!concise || concise.length > 30) {
    concise = description.slice(0, 30).replace(/\s+\S*$/, "").trim();
  }
  invariant(concise.length > 0 && concise.length <= 30, "could not derive a valid short description");
  return concise;
}

function normalizeCategory(category) {
  if (typeof category !== "string" || category.trim() === "") {
    return "Other";
  }
  const normalized = category
    .trim()
    .split(/\s+/)
    .map((word) => `${word[0].toUpperCase()}${word.slice(1).toLowerCase()}`)
    .join(" ");
  return ALLOWED_CATEGORIES.has(normalized) ? normalized : "Other";
}

function repositoryDerivedURL(repository, suffix) {
  if (typeof repository !== "string" || !repository.startsWith("https://")) {
    return undefined;
  }
  return `${repository.replace(/\.git$/, "").replace(/\/$/, "")}${suffix}`;
}

export function generateOpenAIManifest({
  claudeManifest,
  marketplaceEntry,
  releaseVersion,
}) {
  const { coreVersion } = normalizeReleaseVersion(releaseVersion);
  invariant(claudeManifest && typeof claudeManifest === "object", "Claude plugin manifest is required");
  invariant(marketplaceEntry && typeof marketplaceEntry === "object", "marketplace entry is required");
  invariant(claudeManifest.name === marketplaceEntry.name, "plugin and marketplace names differ");
  invariant(claudeManifest.version === marketplaceEntry.version, "plugin and marketplace versions differ");
  invariant(typeof claudeManifest.description === "string" && claudeManifest.description.trim(), "plugin description is required");
  invariant(typeof claudeManifest.author?.name === "string" && claudeManifest.author.name.trim(), "plugin author name is required");

  const displayName = claudeManifest.displayName || marketplaceEntry.displayName || claudeManifest.name;
  const repository = claudeManifest.repository || marketplaceEntry.repository;
  const interfaceMetadata = {
    displayName,
    shortDescription: conciseDescription(claudeManifest.description),
    longDescription: claudeManifest.description,
    developerName: claudeManifest.author.name,
    category: normalizeCategory(marketplaceEntry.category),
    websiteURL: claudeManifest.homepage,
    supportURL: repositoryDerivedURL(repository, "/issues"),
    privacyPolicyURL: claudeManifest.privacyPolicyUrl,
    termsOfServiceURL: repositoryDerivedURL(repository, "/blob/main/LICENSE"),
    composerIcon: "./assets/icon.svg",
    logo: "./assets/icon.svg",
  };
  for (const [key, value] of Object.entries(interfaceMetadata)) {
    if (value === undefined) {
      delete interfaceMetadata[key];
    }
  }

  const manifest = {
    $schema: PLUGIN_SCHEMA_URL,
    name: claudeManifest.name,
    version: coreVersion,
    description: claudeManifest.description,
    author: clone(claudeManifest.author),
    homepage: claudeManifest.homepage,
    repository,
    license: claudeManifest.license,
    keywords: clone(claudeManifest.keywords || marketplaceEntry.keywords || []),
    extensions: {
      "com.openai": {
        interface: interfaceMetadata,
      },
    },
  };
  for (const [key, value] of Object.entries(manifest)) {
    if (value === undefined) {
      delete manifest[key];
    }
  }
  return manifest;
}

export function generatePortableMCP(sourceMCP) {
  invariant(sourceMCP && typeof sourceMCP === "object", "source MCP configuration is required");
  invariant(sourceMCP.mcpServers && typeof sourceMCP.mcpServers === "object", "source MCP configuration has no mcpServers object");
  const mcpServers = {};
  for (const [name, sourceServer] of Object.entries(sourceMCP.mcpServers)) {
    invariant(name.trim(), "MCP server names must be non-empty");
    invariant(sourceServer && typeof sourceServer === "object", `MCP server ${name} must be an object`);
    const server = clone(sourceServer);
    delete server.headersHelper;
    if (server.type === "http") {
      server.type = "streamable-http";
    }
    invariant(
      ["streamable-http", "sse", "stdio"].includes(server.type),
      `MCP server ${name} has unsupported portable transport ${JSON.stringify(server.type)}`,
    );
    if (server.type !== "stdio") {
      invariant(typeof server.url === "string" && server.url.startsWith("https://"), `MCP server ${name} must use an HTTPS URL`);
    }
    mcpServers[name] = server;
  }
  invariant(Object.keys(mcpServers).length > 0, "portable MCP configuration must contain a server");
  return {
    $schema: MCP_SCHEMA_URL,
    mcpServers,
  };
}

export function readOpenAISource(repoRoot) {
  const resolvedRepoRoot = resolve(repoRoot);
  const marketplacePath = join(resolvedRepoRoot, ".claude-plugin", "marketplace.json");
  const marketplace = readJSON(marketplacePath, ".claude-plugin/marketplace.json");
  const marketplaceEntry = (marketplace.plugins || []).find((plugin) => plugin?.name === "newsjack");
  invariant(marketplaceEntry, 'marketplace has no plugin named "newsjack"');
  invariant(typeof marketplaceEntry.source === "string" && marketplaceEntry.source.startsWith("./"), "newsjack marketplace source must be repository-relative");

  const pluginRoot = resolve(resolvedRepoRoot, marketplaceEntry.source);
  assertInside(resolvedRepoRoot, pluginRoot, "newsjack marketplace source");
  invariant(lstatSync(pluginRoot).isDirectory(), "newsjack marketplace source is not a directory");
  const claudeManifest = readJSON(
    join(pluginRoot, ".claude-plugin", "plugin.json"),
    "Newsjack Claude plugin manifest",
  );
  invariant(claudeManifest.name === marketplaceEntry.name, "plugin and marketplace names differ");
  invariant(claudeManifest.version === marketplaceEntry.version, "plugin and marketplace versions differ");

  const skillsRoot = join(pluginRoot, "skills");
  invariant(lstatSync(skillsRoot).isDirectory(), "Newsjack plugin has no skills directory");
  const skillFiles = walkRegularFiles(skillsRoot, "Newsjack plugin skills").filter(
    (path) => path !== ".gitkeep",
  );
  const skillNames = readdirSync(skillsRoot, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .map((entry) => {
      invariant(!entry.name.startsWith("."), `skill directory names must not begin with a dot: ${entry.name}`);
      const skillPath = join(skillsRoot, entry.name, "SKILL.md");
      invariant(statSync(skillPath).isFile(), `${entry.name} has no SKILL.md`);
      const body = readFileSync(skillPath, "utf8");
      const name = parseFrontmatterField(body, "name", `${entry.name}/SKILL.md`);
      const description = parseFrontmatterField(body, "description", `${entry.name}/SKILL.md`);
      invariant(name === entry.name, `${entry.name}/SKILL.md declares name ${JSON.stringify(name)}`);
      invariant(`${claudeManifest.name}:${name}`.length <= 64, `combined plugin and skill name is too long: ${name}`);
      invariant(description.length > 0 && description.length <= 1_024, `${entry.name}/SKILL.md has an invalid description`);
      invariant(body.slice(body.indexOf("\n---", 4) + 4).trim(), `${entry.name}/SKILL.md has no instruction body`);
      return name;
    })
    .sort();
  invariant(skillNames.length > 0, "Newsjack plugin has no skills");

  const sourceMCP = readJSON(join(pluginRoot, ".mcp.json"), "Newsjack plugin MCP configuration");
  const readmePath = join(pluginRoot, "README.md");
  const iconPath = join(pluginRoot, "icon.svg");
  invariant(statSync(readmePath).isFile(), "Newsjack plugin README is missing");
  invariant(statSync(iconPath).isFile(), "Newsjack plugin icon is missing");

  return {
    repoRoot: resolvedRepoRoot,
    pluginRoot,
    claudeManifest,
    marketplaceEntry,
    skillsRoot,
    skillFiles,
    skillNames,
    sourceMCP,
    readmePath,
    iconPath,
  };
}

export function buildOpenAIPluginArchive({ repoRoot, outRoot, releaseVersion }) {
  const source = readOpenAISource(repoRoot);
  const resolvedOutRoot = resolve(outRoot);
  mkdirSync(resolvedOutRoot, { recursive: true });
  const archiveName = openAIArchiveName(releaseVersion);
  const archivePath = join(resolvedOutRoot, archiveName);
  const workRoot = mkdtempSync(join(tmpdir(), "newsjack-openai-plugin-"));
  const packageRoot = join(workRoot, "newsjack");

  try {
    mkdirSync(join(packageRoot, "assets"), { recursive: true });
    cpSync(source.skillsRoot, join(packageRoot, "skills"), { recursive: true });
    rmSync(join(packageRoot, "skills", ".gitkeep"), { force: true });
    cpSync(source.readmePath, join(packageRoot, "README.md"));
    cpSync(source.iconPath, join(packageRoot, "assets", "icon.svg"));
    writeFileSync(
      join(packageRoot, "plugin.json"),
      `${JSON.stringify(generateOpenAIManifest({
        claudeManifest: source.claudeManifest,
        marketplaceEntry: source.marketplaceEntry,
        releaseVersion,
      }), null, 2)}\n`,
    );
    writeFileSync(
      join(packageRoot, "mcp.json"),
      `${JSON.stringify(generatePortableMCP(source.sourceMCP), null, 2)}\n`,
    );

    rmSync(archivePath, { force: true });
    execFileSync("zip", ["-q", "-X", "-r", archivePath, "newsjack"], {
      cwd: workRoot,
      stdio: "inherit",
    });
  } finally {
    rmSync(workRoot, { recursive: true, force: true });
  }

  return {
    archiveName,
    archivePath,
    version: normalizeReleaseVersion(releaseVersion).coreVersion,
    skillCount: source.skillNames.length,
    size: statSync(archivePath).size,
  };
}

function runUnzip(args, options = {}) {
  try {
    return execFileSync("unzip", args, {
      encoding: options.encoding,
      maxBuffer: 16 * 1024 * 1024,
      stdio: options.encoding ? ["ignore", "pipe", "pipe"] : ["ignore", "ignore", "pipe"],
    });
  } catch (error) {
    const detail = error.stderr?.toString().trim();
    throw new Error(`invalid or unreadable ZIP archive${detail ? `: ${detail}` : ""}`);
  }
}

function validateArchivePaths(entries) {
  invariant(entries.length > 0, "archive is empty");
  invariant(entries.length <= MAX_ARCHIVE_ENTRIES, `archive has more than ${MAX_ARCHIVE_ENTRIES} entries`);
  const exact = new Set();
  const normalized = new Set();
  for (const rawPath of entries) {
    invariant(rawPath.trim() === rawPath, `archive path has outer whitespace: ${JSON.stringify(rawPath)}`);
    invariant(!rawPath.includes("\\"), `archive path uses a backslash: ${rawPath}`);
    invariant(!rawPath.startsWith("/"), `archive path is absolute: ${rawPath}`);
    invariant(!/[\u0000-\u001f\u007f]/.test(rawPath), `archive path has a control character: ${JSON.stringify(rawPath)}`);
    const path = rawPath.endsWith("/") ? rawPath.slice(0, -1) : rawPath;
    const parts = path.split("/");
    invariant(parts.every((part) => part.length > 0), `archive path has an empty segment: ${rawPath}`);
    invariant(!parts.includes(".."), `archive path has a parent segment: ${rawPath}`);
    invariant(parts.length <= 20, `archive path is deeper than 20 segments: ${rawPath}`);
    invariant(rawPath.length <= 1_024, `archive path is too long: ${rawPath}`);
    invariant(!exact.has(rawPath), `archive path is duplicated: ${rawPath}`);
    exact.add(rawPath);
    const normalizedPath = rawPath.normalize("NFC").toLowerCase();
    invariant(!normalized.has(normalizedPath), `archive paths collide after normalization: ${rawPath}`);
    normalized.add(normalizedPath);
    for (const part of parts) {
      invariant(!FORBIDDEN_PATH_PARTS.has(part), `archive contains forbidden path: ${rawPath}`);
      invariant(!part.startsWith(".env"), `archive contains forbidden path: ${rawPath}`);
      invariant(!/\.(?:key|pem|p12|pfx|token|secret|credential)$/i.test(part), `archive contains secret-like path: ${rawPath}`);
    }
    invariant(parts[0] === "newsjack", `archive must contain one top-level newsjack directory: ${rawPath}`);
  }
}

function validateManifestShape(manifest) {
  invariant(manifest?.$schema === PLUGIN_SCHEMA_URL, `plugin.json must declare ${PLUGIN_SCHEMA_URL}`);
  invariant(typeof manifest.name === "string" && PLUGIN_NAME_PATTERN.test(manifest.name) && manifest.name.length <= 64, "plugin.json has an invalid name");
  invariant(typeof manifest.version === "string" && SEMVER_PATTERN.test(manifest.version) && manifest.version.length <= 64, "plugin.json has an invalid version");
  invariant(typeof manifest.description === "string" && manifest.description.length > 0 && manifest.description.length <= 1_024, "plugin.json has an invalid description");
  invariant(typeof manifest.author?.name === "string" && manifest.author.name.length > 0 && manifest.author.name.length <= 120, "plugin.json has an invalid author.name");
  const interfaceMetadata = manifest.extensions?.["com.openai"]?.interface;
  invariant(interfaceMetadata && typeof interfaceMetadata === "object", "plugin.json has no extensions.com.openai.interface object");
  invariant(typeof interfaceMetadata.displayName === "string" && interfaceMetadata.displayName.length > 0 && interfaceMetadata.displayName.length <= 30, "OpenAI displayName must be 1-30 characters");
  invariant(typeof interfaceMetadata.shortDescription === "string" && !interfaceMetadata.shortDescription.includes("\n") && interfaceMetadata.shortDescription.length > 0 && interfaceMetadata.shortDescription.length <= 30, "OpenAI shortDescription must be one line and 1-30 characters");
  invariant(typeof interfaceMetadata.longDescription === "string" && interfaceMetadata.longDescription.length > 0 && interfaceMetadata.longDescription.length <= 4_000, "OpenAI longDescription must be 1-4000 characters");
  invariant(typeof interfaceMetadata.developerName === "string" && interfaceMetadata.developerName.length > 0 && interfaceMetadata.developerName.length <= 80, "OpenAI developerName must be 1-80 characters");
  invariant(ALLOWED_CATEGORIES.has(interfaceMetadata.category), "OpenAI category is not supported");
  for (const field of ["websiteURL", "supportURL", "privacyPolicyURL", "termsOfServiceURL"]) {
    invariant(typeof interfaceMetadata[field] === "string" && interfaceMetadata[field].startsWith("https://") && interfaceMetadata[field].length <= 1_024, `OpenAI ${field} must be an HTTPS URL of at most 1024 characters`);
  }
  invariant(interfaceMetadata.composerIcon === "./assets/icon.svg", "OpenAI composerIcon must point to ./assets/icon.svg");
  invariant(interfaceMetadata.logo === "./assets/icon.svg", "OpenAI logo must point to ./assets/icon.svg");
}

function validateMCPShape(mcp) {
  invariant(mcp?.$schema === MCP_SCHEMA_URL, `mcp.json must declare ${MCP_SCHEMA_URL}`);
  invariant(mcp.mcpServers && typeof mcp.mcpServers === "object", "mcp.json has no mcpServers object");
  invariant(Object.keys(mcp.mcpServers).length > 0, "mcp.json has no MCP servers");
  for (const [name, server] of Object.entries(mcp.mcpServers)) {
    invariant(name.trim(), "mcp.json has an empty MCP server name");
    invariant(server && typeof server === "object", `MCP server ${name} is not an object`);
    invariant(!Object.hasOwn(server, "headersHelper"), `MCP server ${name} contains headersHelper`);
    invariant(["streamable-http", "sse", "stdio"].includes(server.type), `MCP server ${name} has an invalid transport`);
    if (server.type !== "stdio") {
      invariant(typeof server.url === "string" && server.url.startsWith("https://"), `MCP server ${name} must use HTTPS`);
    }
  }
}

function validateSVG(path) {
  const size = statSync(path).size;
  invariant(size <= MAX_IMAGE_BYTES, "plugin icon exceeds 5 MiB");
  const body = readFileSync(path, "utf8");
  invariant(/^\s*<svg\b/i.test(body), "plugin icon is not an SVG document");
  const viewBox = body.match(/\bviewBox=["']([^"']+)["']/i)?.[1];
  const width = body.match(/\bwidth=["']([0-9.]+)["']/i)?.[1];
  const height = body.match(/\bheight=["']([0-9.]+)["']/i)?.[1];
  let dimensions;
  if (viewBox) {
    const values = viewBox.trim().split(/[\s,]+/).map(Number);
    invariant(values.length === 4 && values.every(Number.isFinite), "plugin SVG has an invalid viewBox");
    dimensions = values.slice(2);
  } else {
    dimensions = [Number(width), Number(height)];
  }
  invariant(dimensions.every((value) => Number.isFinite(value) && value >= 48), "plugin SVG must be at least 48x48");
  invariant(dimensions[0] === dimensions[1], "plugin SVG must be square");
}

function comparePackagedFiles(packageRoot, source) {
  const actualFiles = walkRegularFiles(packageRoot, "OpenAI plugin archive").map((path) =>
    path.split(sep).join("/"),
  );
  const expectedFiles = [
    "README.md",
    "assets/icon.svg",
    "mcp.json",
    "plugin.json",
    ...source.skillFiles.map((path) => `skills/${path.split(sep).join("/")}`),
  ].sort();
  invariant(
    isDeepStrictEqual(actualFiles.sort(), expectedFiles),
    `archive file tree differs from the generated package\nexpected: ${expectedFiles.join(", ")}\nactual: ${actualFiles.sort().join(", ")}`,
  );

  for (const skillFile of source.skillFiles) {
    const sourceBody = readFileSync(join(source.skillsRoot, skillFile));
    const packagedBody = readFileSync(join(packageRoot, "skills", skillFile));
    invariant(sourceBody.equals(packagedBody), `packaged skill file differs from source: ${skillFile}`);
  }
  invariant(
    readFileSync(source.readmePath).equals(readFileSync(join(packageRoot, "README.md"))),
    "packaged README differs from source",
  );
  invariant(
    readFileSync(source.iconPath).equals(readFileSync(join(packageRoot, "assets", "icon.svg"))),
    "packaged icon differs from source",
  );
}

export function validateOpenAIPluginArchive({
  archivePath,
  repoRoot,
  expectedReleaseVersion,
}) {
  const resolvedArchivePath = resolve(archivePath);
  invariant(statSync(resolvedArchivePath).isFile(), `archive is not a regular file: ${archivePath}`);
  const archiveSize = statSync(resolvedArchivePath).size;
  invariant(archiveSize > 0, "archive is empty");
  invariant(archiveSize <= MAX_ARCHIVE_BYTES, "compressed archive exceeds 100 MB");
  if (expectedReleaseVersion) {
    invariant(
      basename(resolvedArchivePath) === openAIArchiveName(expectedReleaseVersion),
      `archive filename must be ${openAIArchiveName(expectedReleaseVersion)}`,
    );
  }

  const listing = runUnzip(["-Z1", resolvedArchivePath], { encoding: "utf8" });
  const entries = listing.split(/\r?\n/).filter((entry) => entry.length > 0);
  validateArchivePaths(entries);
  runUnzip(["-tqq", resolvedArchivePath]);

  const extractRoot = mkdtempSync(join(tmpdir(), "newsjack-openai-validate-"));
  try {
    runUnzip(["-qq", resolvedArchivePath, "-d", extractRoot]);
    const packageRoot = join(extractRoot, "newsjack");
    invariant(lstatSync(packageRoot).isDirectory(), "archive has no top-level newsjack directory");

    let uncompressedBytes = 0;
    for (const relativePath of walkRegularFiles(packageRoot, "OpenAI plugin archive")) {
      const size = statSync(join(packageRoot, relativePath)).size;
      invariant(size <= MAX_MEMBER_BYTES, `archive member exceeds 100 MiB: ${relativePath}`);
      uncompressedBytes += size;
    }
    invariant(uncompressedBytes <= MAX_UNCOMPRESSED_BYTES, "extracted archive exceeds 512 MiB");

    const source = readOpenAISource(repoRoot);
    comparePackagedFiles(packageRoot, source);
    const manifest = readJSON(join(packageRoot, "plugin.json"), "OpenAI plugin.json");
    validateManifestShape(manifest);
    const releaseVersion = expectedReleaseVersion || manifest.version;
    const expectedManifest = generateOpenAIManifest({
      claudeManifest: source.claudeManifest,
      marketplaceEntry: source.marketplaceEntry,
      releaseVersion,
    });
    invariant(isDeepStrictEqual(manifest, expectedManifest), "plugin.json was not generated from the Newsjack plugin metadata");

    const mcp = readJSON(join(packageRoot, "mcp.json"), "OpenAI mcp.json");
    validateMCPShape(mcp);
    invariant(isDeepStrictEqual(mcp, generatePortableMCP(source.sourceMCP)), "mcp.json was not generated from the Newsjack plugin MCP configuration");
    validateSVG(join(packageRoot, "assets", "icon.svg"));

    return {
      archivePath: resolvedArchivePath,
      archiveSize,
      uncompressedBytes,
      entryCount: entries.length,
      skillCount: source.skillNames.length,
      version: manifest.version,
    };
  } finally {
    rmSync(extractRoot, { recursive: true, force: true });
  }
}
