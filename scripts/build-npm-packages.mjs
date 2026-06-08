import { createHash } from "node:crypto";
import {
  chmodSync,
  cpSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { join, relative, resolve } from "node:path";
import { execFileSync } from "node:child_process";

const repoRoot = resolve(import.meta.dirname, "..");
const cliRoot = join(repoRoot, "apps/cli");
const skillsRoot = join(repoRoot, "skills");
const outRoot = resolve(
  repoRoot,
  process.env.NEWSJACK_NPM_DIST || ".tmp/newsjack-npm",
);
const goVersion = process.env.NEWSJACK_GO_VERSION || "1.26.3";

const targets = [
  {
    packageName: "newsjack-linux-arm64",
    packageDir: join(outRoot, "newsjack-linux-arm64"),
    nodeOS: "linux",
    nodeArch: "arm64",
    goOS: "linux",
    goArch: "arm64",
  },
  {
    packageName: "newsjack-linux-x64",
    packageDir: join(outRoot, "newsjack-linux-x64"),
    nodeOS: "linux",
    nodeArch: "x64",
    goOS: "linux",
    goArch: "amd64",
  },
  {
    packageName: "newsjack-darwin-arm64",
    packageDir: join(outRoot, "newsjack-darwin-arm64"),
    nodeOS: "darwin",
    nodeArch: "arm64",
    goOS: "darwin",
    goArch: "arm64",
  },
  {
    packageName: "newsjack-darwin-x64",
    packageDir: join(outRoot, "newsjack-darwin-x64"),
    nodeOS: "darwin",
    nodeArch: "x64",
    goOS: "darwin",
    goArch: "amd64",
  },
];

function run(command, args, options = {}) {
  execFileSync(command, args, {
    cwd: options.cwd ?? repoRoot,
    env: { ...process.env, ...options.env },
    stdio: "inherit",
  });
}

function output(command, args, options = {}) {
  return execFileSync(command, args, {
    cwd: options.cwd ?? repoRoot,
    env: { ...process.env, ...options.env },
    encoding: "utf8",
    stdio: ["ignore", "pipe", "ignore"],
  }).trim();
}

function commandOutput(command, args, options = {}) {
  try {
    return output(command, args, options);
  } catch {
    return "";
  }
}

function platformName() {
  switch (process.platform) {
    case "darwin":
      return "darwin";
    case "linux":
      return "linux";
    default:
      throw new Error(`unsupported build platform: ${process.platform}`);
  }
}

function archName() {
  switch (process.arch) {
    case "x64":
      return "amd64";
    case "arm64":
      return "arm64";
    default:
      throw new Error(`unsupported build architecture: ${process.arch}`);
  }
}

function findGo() {
  if (process.env.NEWSJACK_GO) {
    return process.env.NEWSJACK_GO;
  }
  const existing = commandOutput("sh", ["-c", "command -v go"]);
  if (existing) {
    return existing;
  }

  const os = platformName();
  const arch = archName();
  const goRoot = join(repoRoot, ".tmp", `go${goVersion}`);
  const goBin = join(goRoot, "go", "bin", "go");
  const archive = join(repoRoot, ".tmp", `go${goVersion}.${os}-${arch}.tar.gz`);
  const url = `https://go.dev/dl/go${goVersion}.${os}-${arch}.tar.gz`;

  mkdirSync(join(repoRoot, ".tmp"), { recursive: true });
  rmSync(goRoot, { recursive: true, force: true });
  console.log(`Installing Go ${goVersion} for Newsjack npm build`);
  run("curl", ["-fsSL", url, "-o", archive]);
  mkdirSync(goRoot, { recursive: true });
  run("tar", ["-xzf", archive, "-C", goRoot]);
  return goBin;
}

function sha256File(path) {
  const hash = createHash("sha256");
  hash.update(readFileSync(path));
  return hash.digest("hex");
}

function walkFiles(root, prefix = "") {
  const out = [];
  for (const entry of readdirSync(root, { withFileTypes: true })) {
    const rel = join(prefix, entry.name);
    const path = join(root, entry.name);
    if (entry.isDirectory()) {
      out.push(...walkFiles(path, rel));
    } else if (entry.isFile()) {
      out.push(rel);
    }
  }
  return out.sort();
}

function copyIfExists(name, destRoot) {
  const source = join(repoRoot, name);
  try {
    statSync(source);
  } catch {
    return;
  }
  cpSync(source, join(destRoot, name), { recursive: true });
}

function buildSkillsManifest({ version, npmVersion, commit, builtAt }) {
  const files = walkFiles(skillsRoot).map((relPath) => {
    const path = join(skillsRoot, relPath);
    return {
      path: `skills/${relPath}`,
      sha256: sha256File(path),
      size: statSync(path).size,
    };
  });
  const skillNames = readdirSync(skillsRoot, { withFileTypes: true })
    .filter((entry) => entry.isDirectory())
    .filter((entry) => {
      try {
        return statSync(join(skillsRoot, entry.name, "SKILL.md")).isFile();
      } catch {
        return false;
      }
    })
    .map((entry) => entry.name)
    .sort();

  return {
    version,
    npm_version: npmVersion,
    commit,
    built_at: builtAt,
    distribution: "npm",
    skills: skillNames,
    files,
  };
}

function npmVersionFrom(version) {
  const npmVersion = version.replace(/^v/, "");
  if (!/^\d+\.\d+\.\d+(?:-[0-9A-Za-z][0-9A-Za-z.-]*)?(?:\+[0-9A-Za-z][0-9A-Za-z.-]*)?$/.test(npmVersion)) {
    throw new Error(`NEWSJACK_VERSION must be a publishable semver version, got ${version}`);
  }
  return npmVersion;
}

function writeJSON(path, value) {
  writeFileSync(path, `${JSON.stringify(value, null, 2)}\n`);
}

function packageMetadata({ name, version, description, extra = {} }) {
  return {
    name,
    version,
    description,
    license: "MIT",
    homepage: "https://newsjack.sh",
    repository: {
      type: "git",
      url: "git+https://github.com/elvisun/newsjack.git",
    },
    bugs: {
      url: "https://github.com/elvisun/newsjack/issues",
    },
    ...extra,
  };
}

function wrapperScript() {
  const packageMap = Object.fromEntries(
    targets.map((target) => [`${target.nodeOS}-${target.nodeArch}`, target.packageName]),
  );
  return `#!/usr/bin/env node
const { spawnSync } = require("node:child_process");
const fs = require("node:fs");
const path = require("node:path");

const packages = ${JSON.stringify(packageMap, null, 2)};
const key = process.platform + "-" + process.arch;
const packageName = packages[key];

if (!packageName) {
  console.error("newsjack: unsupported platform " + key);
  process.exit(1);
}

function findPackageRoot(name) {
  const candidates = [
    path.resolve(__dirname, "..", "..", name),
  ];

  const argvPath = process.argv[1] ? path.resolve(process.argv[1]) : "";
  if (argvPath.includes(path.sep + ".bin" + path.sep)) {
    candidates.push(path.resolve(path.dirname(argvPath), "..", name));
  }

  for (const candidate of candidates) {
    if (fs.existsSync(path.join(candidate, "package.json"))) {
      return candidate;
    }
  }

  try {
    return path.dirname(require.resolve(name + "/package.json"));
  } catch {
    return "";
  }
}

const packageRoot = findPackageRoot(packageName);
if (!packageRoot) {
  console.error("newsjack: missing platform package " + packageName);
  console.error("Reinstall with optional dependencies enabled: npm i -g newsjack");
  console.error("Or run one-shot with: npx -y newsjack@latest <command>");
  process.exit(1);
}

const binary = path.join(packageRoot, "bin", "newsjack");
const mainRoot = path.resolve(__dirname, "..");
const env = {
  ...process.env,
  NEWSJACK_DISTRIBUTION: process.env.NEWSJACK_DISTRIBUTION || "npm",
  NEWSJACK_ROOT: process.env.NEWSJACK_ROOT || mainRoot,
  NEWSJACK_NPM_PACKAGE: process.env.NEWSJACK_NPM_PACKAGE || "newsjack",
};

const result = spawnSync(binary, process.argv.slice(2), {
  stdio: "inherit",
  env,
});

if (result.error) {
  console.error("newsjack: could not run " + binary + ": " + result.error.message);
  process.exit(1);
}

if (result.signal) {
  process.kill(process.pid, result.signal);
}

process.exit(result.status ?? 1);
`;
}

const goBinary = findGo();
const commit = process.env.GITHUB_SHA || output("git", ["rev-parse", "HEAD"]);
const tagName = process.env.GITHUB_REF_NAME || commandOutput("git", ["describe", "--tags", "--exact-match"]);
const version = process.env.NEWSJACK_VERSION || tagName || `v0.1.0-dev+${commit.slice(0, 12)}`;
const npmVersion = npmVersionFrom(version);
const builtAt = new Date().toISOString();
const skillsManifest = buildSkillsManifest({ version, npmVersion, commit, builtAt });
const mainPackageRoot = join(outRoot, "newsjack");

rmSync(outRoot, { recursive: true, force: true });
mkdirSync(outRoot, { recursive: true });

for (const target of targets) {
  const binRoot = join(target.packageDir, "bin");
  const binary = join(binRoot, "newsjack");
  mkdirSync(binRoot, { recursive: true });
  run(
    goBinary,
    [
      "build",
      "-trimpath",
      "-buildvcs=false",
      "-ldflags",
      `-s -w -X main.version=${version}`,
      "-o",
      binary,
      "./cmd/newsjack",
    ],
    {
      cwd: cliRoot,
      env: { CGO_ENABLED: "0", GOOS: target.goOS, GOARCH: target.goArch },
    },
  );
  chmodSync(binary, 0o755);
  writeJSON(
    join(target.packageDir, "package.json"),
    packageMetadata({
      name: target.packageName,
      version: npmVersion,
      description: `Newsjack CLI binary for ${target.nodeOS}/${target.nodeArch}`,
      extra: {
        os: [target.nodeOS],
        cpu: [target.nodeArch],
        files: ["bin/newsjack"],
        publishConfig: { access: "public" },
      },
    }),
  );
  copyIfExists("README.md", target.packageDir);
  copyIfExists("LICENSE", target.packageDir);
}

mkdirSync(join(mainPackageRoot, "bin"), { recursive: true });
cpSync(skillsRoot, join(mainPackageRoot, "skills"), { recursive: true });
copyIfExists("README.md", mainPackageRoot);
copyIfExists("LICENSE", mainPackageRoot);
copyIfExists(".mcp.json", mainPackageRoot);
writeFileSync(join(mainPackageRoot, "VERSION"), `${version}\n`);
writeFileSync(join(mainPackageRoot, "COMMIT"), `${commit}\n`);
writeFileSync(join(mainPackageRoot, ".newsjack-npm"), "1\n");
writeJSON(join(mainPackageRoot, "skills-manifest.json"), skillsManifest);
writeFileSync(join(mainPackageRoot, "bin", "newsjack"), wrapperScript());
chmodSync(join(mainPackageRoot, "bin", "newsjack"), 0o755);
writeJSON(
  join(mainPackageRoot, "package.json"),
  packageMetadata({
    name: "newsjack",
    version: npmVersion,
    description: "Open-source agent skills and CLI for PR operators.",
    extra: {
      bin: {
        newsjack: "bin/newsjack",
      },
      engines: {
        node: ">=18",
      },
      files: [
        "bin/newsjack",
        "skills",
        ".mcp.json",
        ".newsjack-npm",
        "VERSION",
        "COMMIT",
        "skills-manifest.json",
        "README.md",
        "LICENSE",
      ],
      optionalDependencies: Object.fromEntries(
        targets.map((target) => [target.packageName, npmVersion]),
      ),
    },
  }),
);

console.log(`Built Newsjack ${version} npm packages in ${relative(repoRoot, outRoot)}.`);
