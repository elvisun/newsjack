import { cpSync, mkdirSync } from "node:fs";
import { join, resolve } from "node:path";

const siteRoot = resolve(import.meta.dirname, "..");
const repoRoot = resolve(siteRoot, "../..");
const publicRoot = join(siteRoot, "public");

mkdirSync(publicRoot, { recursive: true });
cpSync(join(repoRoot, "install.sh"), join(publicRoot, "install.sh"));

console.log("Bundled Newsjack installer for newsjack.sh.");
