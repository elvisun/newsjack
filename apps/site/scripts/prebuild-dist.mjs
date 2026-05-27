import { createHash } from "node:crypto";
import {
  cpSync,
  mkdirSync,
  readFileSync,
  rmSync,
  statSync,
  writeFileSync,
} from "node:fs";
import { basename, join, resolve } from "node:path";
import { execFileSync } from "node:child_process";

const siteRoot = resolve(import.meta.dirname, "..");
const repoRoot = resolve(siteRoot, "../..");
const publicRoot = join(siteRoot, "public");
const distRoot = join(publicRoot, "dist");
const cliRoot = join(repoRoot, "apps/cli");
const skillsRoot = join(repoRoot, "skills");
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
  console.log(`Installing Go ${goVersion} for Newsjack dist build`);
  run("curl", ["-fsSL", url, "-o", archive]);
  mkdirSync(goRoot, { recursive: true });
  run("tar", ["-xzf", archive, "-C", goRoot]);
  return goBin;
}

function fileSha256(path) {
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

const goBinary = findGo();

function packageTarget({ os, arch, commit, version, commitDir }) {
  const workRoot = join(
    repoRoot,
    ".tmp",
    "newsjack-dist",
    `newsjack_${os}_${arch}`,
  );
  const payloadRoot = join(workRoot, "payload");
  const binRoot = join(payloadRoot, "bin");
  const binary = join(binRoot, os === "windows" ? "newsjack.exe" : "newsjack");
  const artifact = `newsjack_${os}_${arch}.tar.gz`;
  const artifactPath = join(commitDir, artifact);

  rmSync(workRoot, { recursive: true, force: true });
  mkdirSync(binRoot, { recursive: true });

  run(goBinary, ["build", "-trimpath", "-buildvcs=false", "-ldflags", `-s -w -X main.version=${version}`, "-o", binary, "./cmd/newsjack"], {
    cwd: cliRoot,
    env: { CGO_ENABLED: "0", GOOS: os, GOARCH: arch },
  });

  cpSync(skillsRoot, join(payloadRoot, "skills"), { recursive: true });
  copyIfExists("README.md", payloadRoot);
  copyIfExists("LICENSE", payloadRoot);
  copyIfExists(".mcp.json", payloadRoot);
  writeFileSync(join(payloadRoot, ".newsjack-prebuilt"), "1\n");
  writeFileSync(join(payloadRoot, "VERSION"), `${commit}\n`);
  writeFileSync(
    join(payloadRoot, "manifest.json"),
    `${JSON.stringify({ commit, version, os, arch }, null, 2)}\n`,
  );

  run("tar", ["-czf", artifactPath, "-C", payloadRoot, "."]);
  const sha256 = fileSha256(artifactPath);
  const size = statSync(artifactPath).size;
  writeFileSync(join(commitDir, `${artifact}.sha256`), `${sha256}  ${artifact}\n`);
  rmSync(workRoot, { recursive: true, force: true });

  return { name: artifact, os, arch, path: `commits/${commit}/${artifact}`, sha256, size };
}

const commit =
  process.env.VERCEL_GIT_COMMIT_SHA ||
  process.env.GITHUB_SHA ||
  output("git", ["rev-parse", "HEAD"]);
const shortCommit = commit.slice(0, 12);
const version = process.env.NEWSJACK_VERSION || `0.2.0-go+${shortCommit}`;
const builtAt = new Date().toISOString();
const commitDir = join(distRoot, "commits", commit);
const channelDir = join(distRoot, "channels");

rmSync(distRoot, { recursive: true, force: true });
mkdirSync(commitDir, { recursive: true });
mkdirSync(channelDir, { recursive: true });
cpSync(join(repoRoot, "install.sh"), join(publicRoot, "install.sh"));

const artifacts = targets.map(([os, arch]) =>
  packageTarget({ os, arch, commit, version, commitDir }),
);

const manifest = {
  channel: "main",
  commit,
  version,
  built_at: builtAt,
  artifacts,
};

writeFileSync(join(commitDir, "manifest.json"), `${JSON.stringify(manifest, null, 2)}\n`);
writeFileSync(join(channelDir, "main.txt"), `${commit}\n`);
writeFileSync(join(channelDir, "main.json"), `${JSON.stringify(manifest, null, 2)}\n`);
writeFileSync(join(distRoot, "manifest.json"), `${JSON.stringify(manifest, null, 2)}\n`);

console.log(`Bundled Newsjack dist ${commit} for ${artifacts.length} targets.`);
