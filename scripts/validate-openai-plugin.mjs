import { resolve } from "node:path";

import {
  openAIArchiveName,
  validateOpenAIPluginArchive,
} from "./lib/openai-plugin.mjs";

const repoRoot = resolve(import.meta.dirname, "..");
const expectedReleaseVersion = process.argv[3] || process.env.NEWSJACK_VERSION;
const defaultArchive = expectedReleaseVersion
  ? resolve(
      repoRoot,
      process.env.NEWSJACK_RELEASE_DIST || ".tmp/newsjack-release",
      openAIArchiveName(expectedReleaseVersion),
    )
  : undefined;
const archivePath = process.argv[2] ? resolve(process.argv[2]) : defaultArchive;

if (!archivePath) {
  console.error(
    "OpenAI plugin validation failed: pass an archive path or set NEWSJACK_VERSION",
  );
  process.exit(1);
}

try {
  const result = validateOpenAIPluginArchive({
    archivePath,
    repoRoot,
    expectedReleaseVersion,
  });
  console.log(
    `OpenAI plugin archive is valid: ${result.archivePath} (${result.skillCount} skills, ${result.entryCount} ZIP entries, manifest version ${result.version})`,
  );
} catch (error) {
  console.error(`OpenAI plugin validation failed: ${error.message}`);
  process.exit(1);
}
