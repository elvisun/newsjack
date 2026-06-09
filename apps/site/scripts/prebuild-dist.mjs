import { cpSync, mkdirSync, rmSync } from "node:fs";
import { join, resolve } from "node:path";

const siteRoot = resolve(import.meta.dirname, "..");
const repoRoot = resolve(siteRoot, "../..");
const publicRoot = join(siteRoot, "public");
const generatedRoot = join(publicRoot, "_generated");

mkdirSync(generatedRoot, { recursive: true });
rmSync(join(publicRoot, "install.sh"), { force: true });
cpSync(join(repoRoot, "install.sh"), join(generatedRoot, "install.sh"));

console.log("Bundled Newsjack installer for newsjack.sh.");
