import { fileURLToPath } from "node:url";
import { resolve } from "node:path";

import {
  buildOpenAIPluginArchive,
  openAIArchiveName,
} from "./lib/openai-plugin.mjs";

export { buildOpenAIPluginArchive, openAIArchiveName };

function main() {
  const repoRoot = resolve(import.meta.dirname, "..");
  const releaseVersion = process.env.NEWSJACK_VERSION || process.env.GITHUB_REF_NAME;
  if (!releaseVersion) {
    throw new Error("NEWSJACK_VERSION or GITHUB_REF_NAME is required");
  }
  const outRoot = resolve(
    repoRoot,
    process.env.NEWSJACK_RELEASE_DIST || ".tmp/newsjack-release",
  );
  const result = buildOpenAIPluginArchive({ repoRoot, outRoot, releaseVersion });
  console.log(
    `Built ${result.archiveName} (${result.skillCount} skills, manifest version ${result.version}).`,
  );
}

if (process.argv[1] && fileURLToPath(import.meta.url) === resolve(process.argv[1])) {
  main();
}
