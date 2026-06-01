import { createHash } from "node:crypto";
import {
  cpSync,
  mkdirSync,
  readdirSync,
  readFileSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { basename, join, relative, resolve } from "node:path";
import { execFileSync } from "node:child_process";

const repoRoot = resolve(import.meta.dirname, "..");
const cliRoot = join(repoRoot, "apps/cli");
const skillsRoot = join(repoRoot, "skills");
const outRoot = resolve(
  repoRoot,
  process.env.NEWSJACK_RELEASE_DIST || ".tmp/newsjack-release",
);
const goVersion = process.env.NEWSJACK_GO_VERSION || "1.26.3";

const targets = [
  ["darwin", "amd64"],
  ["darwin", "arm64"],
  ["linux", "amd64"],
  ["linux", "arm64"],
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
  console.log(`Installing Go ${goVersion} for Newsjack release build`);
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

function copyIfExists(name, destRoot) {
  const source = join(repoRoot, name);
  try {
    statSync(source);
  } catch {
    return;
  }
  cpSync(source, join(destRoot, basename(name)), { recursive: true });
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

function buildSkillsManifest({ version, commit, builtAt }) {
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
    commit,
    built_at: builtAt,
    skills: skillNames,
    files,
  };
}

const goBinary = findGo();
const commit = process.env.GITHUB_SHA || output("git", ["rev-parse", "HEAD"]);
const tagName = process.env.GITHUB_REF_NAME || commandOutput("git", ["describe", "--tags", "--exact-match"]);
const version = process.env.NEWSJACK_VERSION || tagName || `v0.1.0-dev+${commit.slice(0, 12)}`;
const builtAt = new Date().toISOString();
const skillsManifest = buildSkillsManifest({ version, commit, builtAt });

function packageTarget({ os, arch }) {
  const workRoot = join(
    repoRoot,
    ".tmp",
    "newsjack-release-work",
    `newsjack_${os}_${arch}`,
  );
  const payloadRoot = join(workRoot, "payload");
  const binRoot = join(payloadRoot, "bin");
  const binary = join(binRoot, "newsjack");
  const artifact = `newsjack_${os}_${arch}.tar.gz`;
  const artifactPath = join(outRoot, artifact);

  rmSync(workRoot, { recursive: true, force: true });
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
      env: { CGO_ENABLED: "0", GOOS: os, GOARCH: arch },
    },
  );

  cpSync(skillsRoot, join(payloadRoot, "skills"), { recursive: true });
  copyIfExists("README.md", payloadRoot);
  copyIfExists("LICENSE", payloadRoot);
  copyIfExists(".mcp.json", payloadRoot);
  writeFileSync(join(payloadRoot, ".newsjack-prebuilt"), "1\n");
  writeFileSync(join(payloadRoot, "VERSION"), `${version}\n`);
  writeFileSync(join(payloadRoot, "COMMIT"), `${commit}\n`);
  writeFileSync(join(payloadRoot, "skills-manifest.json"), `${JSON.stringify(skillsManifest, null, 2)}\n`);
  writeFileSync(
    join(payloadRoot, "manifest.json"),
    `${JSON.stringify({ version, commit, os, arch, built_at: builtAt, artifact }, null, 2)}\n`,
  );

  run("tar", ["--no-xattrs", "-czf", artifactPath, "-C", payloadRoot, "."], {
    env: { COPYFILE_DISABLE: "1" },
  });
  const sha256 = sha256File(artifactPath);
  const size = statSync(artifactPath).size;
  rmSync(workRoot, { recursive: true, force: true });

  return { name: artifact, os, arch, sha256, size };
}

rmSync(outRoot, { recursive: true, force: true });
mkdirSync(outRoot, { recursive: true });

const artifacts = targets.map(([os, arch]) => packageTarget({ os, arch }));
const manifest = {
  version,
  commit,
  channel: "stable",
  built_at: builtAt,
  artifacts,
};
const checksums = artifacts
  .map((artifact) => `${artifact.sha256}  ${artifact.name}`)
  .join("\n");

writeFileSync(join(outRoot, "manifest.json"), `${JSON.stringify(manifest, null, 2)}\n`);
writeFileSync(join(outRoot, "skills-manifest.json"), `${JSON.stringify(skillsManifest, null, 2)}\n`);
writeFileSync(join(outRoot, "checksums.txt"), `${checksums}\n`);
cpSync(join(repoRoot, "install.sh"), join(outRoot, "install.sh"));

console.log(`Built Newsjack ${version} release assets in ${relative(repoRoot, outRoot)}.`);
