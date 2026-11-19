import { fileURLToPath } from "node:url";
import { resolve } from "node:path";

import {
  buildOpenAIPluginArchives,
  openAIArchiveName,
  writeStableChatGPTSkillsArchive,
} from "./lib/openai-plugin.mjs";

export {
  buildOpenAIPluginArchives,
  openAIArchiveName,
  writeStableChatGPTSkillsArchive,
};

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
  const results = buildOpenAIPluginArchives({ repoRoot, outRoot, releaseVersion });
  const stableSkills = writeStableChatGPTSkillsArchive({
    archives: results,
    outRoot,
  });
  for (const result of results) {
    console.log(
      `Built ${result.archiveName} (${result.skillCount} skills, manifest version ${result.version}).`,
    );
  }
  console.log(
    `Built ${stableSkills.archiveName} as a byte-identical alias of ${stableSkills.sourceArchiveName}.`,
  );
}

if (process.argv[1] && fileURLToPath(import.meta.url) === resolve(process.argv[1])) {
  main();
}
