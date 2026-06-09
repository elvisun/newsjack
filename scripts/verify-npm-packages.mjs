import { execFileSync } from "node:child_process";
import {
  accessSync,
  constants,
  readFileSync,
  statSync,
} from "node:fs";
import { join, resolve } from "node:path";

const outRoot = resolve(process.argv[2] || ".tmp/newsjack-npm");
const releaseVersion = process.argv[3] || process.env.NEWSJACK_VERSION || "";

const targets = [
  {
    packageName: "newsjack-linux-arm64",
    nodeOS: "linux",
    nodeArch: "arm64",
  },
  {
    packageName: "newsjack-linux-x64",
    nodeOS: "linux",
    nodeArch: "x64",
  },
  {
    packageName: "newsjack-darwin-arm64",
    nodeOS: "darwin",
    nodeArch: "arm64",
  },
  {
    packageName: "newsjack-darwin-x64",
    nodeOS: "darwin",
    nodeArch: "x64",
  },
];

function npmVersionFrom(version) {
  const npmVersion = version.replace(/^v/, "");
  if (!/^\d+\.\d+\.\d+(?:-[0-9A-Za-z][0-9A-Za-z.-]*)?(?:\+[0-9A-Za-z][0-9A-Za-z.-]*)?$/.test(npmVersion)) {
    throw new Error(`release version must be publishable semver, got ${version}`);
  }
  return npmVersion;
}

function readJSON(path) {
  return JSON.parse(readFileSync(path, "utf8"));
}

function assert(condition, message) {
  if (!condition) {
    throw new Error(message);
  }
}

function assertExecutable(path) {
  const stats = statSync(path);
  assert(stats.isFile(), `${path} must be a file`);
  accessSync(path, constants.X_OK);
}

function packDryRun(packageDir) {
  const raw = execFileSync("npm", ["pack", "--dry-run", "--json", packageDir], {
    encoding: "utf8",
    stdio: ["ignore", "pipe", "inherit"],
  });
  const result = JSON.parse(raw);
  assert(Array.isArray(result) && result.length === 1, `unexpected npm pack output for ${packageDir}`);
  return new Set(result[0].files.map((file) => file.path));
}

function assertCommonMetadata(pkg, packageName, npmVersion) {
  assert(pkg.name === packageName, `${packageName}: package name mismatch`);
  assert(pkg.version === npmVersion, `${packageName}: package version mismatch`);
  assert(pkg.license === "MIT", `${packageName}: license mismatch`);
  assert(pkg.repository?.url === "git+https://github.com/elvisun/newsjack.git", `${packageName}: repository.url mismatch`);
  assert(pkg.bugs?.url === "https://github.com/elvisun/newsjack/issues", `${packageName}: bugs.url mismatch`);
}

const npmVersion = npmVersionFrom(releaseVersion);

for (const target of targets) {
  const packageDir = join(outRoot, target.packageName);
  const pkg = readJSON(join(packageDir, "package.json"));
  assertCommonMetadata(pkg, target.packageName, npmVersion);
  assert(pkg.publishConfig?.access === "public", `${target.packageName}: publishConfig.access must be public`);
  assert(pkg.os?.length === 1 && pkg.os[0] === target.nodeOS, `${target.packageName}: os mismatch`);
  assert(pkg.cpu?.length === 1 && pkg.cpu[0] === target.nodeArch, `${target.packageName}: cpu mismatch`);
  assertExecutable(join(packageDir, "bin", "newsjack"));

  const files = packDryRun(packageDir);
  assert(files.has("bin/newsjack"), `${target.packageName}: npm pack is missing bin/newsjack`);
  assert(files.has("package.json"), `${target.packageName}: npm pack is missing package.json`);
}

const mainPackageDir = join(outRoot, "newsjack");
const mainPkg = readJSON(join(mainPackageDir, "package.json"));
assertCommonMetadata(mainPkg, "newsjack", npmVersion);
assert(mainPkg.bin?.newsjack === "bin/newsjack", "newsjack: bin.newsjack mismatch");
assert(mainPkg.publishConfig?.access === "public", "newsjack: publishConfig.access must be public");
assert(mainPkg.engines?.node === ">=18", "newsjack: engines.node mismatch");

for (const target of targets) {
  assert(
    mainPkg.optionalDependencies?.[target.packageName] === npmVersion,
    `newsjack: optional dependency ${target.packageName} must equal ${npmVersion}`,
  );
}

const manifest = readJSON(join(mainPackageDir, "skills-manifest.json"));
assert(manifest.version === releaseVersion, "newsjack: skills manifest release version mismatch");
assert(manifest.npm_version === npmVersion, "newsjack: skills manifest npm version mismatch");
assert(manifest.distribution === "npm", "newsjack: skills manifest distribution mismatch");
assertExecutable(join(mainPackageDir, "bin", "newsjack"));

const mainFiles = packDryRun(mainPackageDir);
for (const expectedFile of [
  "bin/newsjack",
  "skills-manifest.json",
  "VERSION",
  "COMMIT",
  ".newsjack-npm",
  "package.json",
]) {
  assert(mainFiles.has(expectedFile), `newsjack: npm pack is missing ${expectedFile}`);
}

execFileSync(join(mainPackageDir, "bin", "newsjack"), ["version"], {
  env: {
    ...process.env,
    NEWSJACK_NO_AUTO_UPDATE: "1",
  },
  stdio: "inherit",
});

console.log(`Verified Newsjack npm packages for ${releaseVersion} in ${outRoot}.`);
