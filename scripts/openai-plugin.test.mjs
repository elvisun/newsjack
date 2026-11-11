import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import {
  mkdirSync,
  mkdtempSync,
  rmSync,
  writeFileSync,
} from "node:fs";
import { tmpdir } from "node:os";
import { join } from "node:path";
import test from "node:test";

import {
  MCP_SCHEMA_URL,
  PLUGIN_SCHEMA_URL,
  buildOpenAIPluginArchive,
  generateOpenAIManifest,
  generatePortableMCP,
  normalizeReleaseVersion,
  validateOpenAIPluginArchive,
} from "./lib/openai-plugin.mjs";

const claudeManifest = {
  name: "newsjack",
  displayName: "Newsjack",
  version: "1.2.3",
  description:
    "34 open-source PR and AI-visibility skills for strategy, monitoring, and research.",
  icon: "./icon.svg",
  privacyPolicyUrl: "https://newsjack.sh/privacy",
  author: {
    name: "Elvis Sun",
    email: "newsjack@elvisun.com",
    url: "https://github.com/elvisun",
  },
  homepage: "https://github.com/elvisun/newsjack",
  repository: "https://github.com/elvisun/newsjack",
  license: "MIT",
  keywords: ["pr", "newsjacking"],
};

const marketplaceEntry = {
  name: "newsjack",
  source: "./plugins/newsjack",
  displayName: "Newsjack",
  version: "1.2.3",
  category: "productivity",
};

function writeJSON(path, value) {
  writeFileSync(path, `${JSON.stringify(value, null, 2)}\n`);
}

function createFixtureRepo() {
  const repoRoot = mkdtempSync(join(tmpdir(), "newsjack-openai-test-"));
  const pluginRoot = join(repoRoot, "plugins", "newsjack");
  mkdirSync(join(repoRoot, ".claude-plugin"), { recursive: true });
  mkdirSync(join(pluginRoot, ".claude-plugin"), { recursive: true });
  mkdirSync(join(pluginRoot, "skills", "sample"), { recursive: true });
  writeJSON(join(repoRoot, ".claude-plugin", "marketplace.json"), {
    name: "newsjack",
    plugins: [marketplaceEntry],
  });
  writeJSON(join(pluginRoot, ".claude-plugin", "plugin.json"), claudeManifest);
  writeJSON(join(pluginRoot, ".mcp.json"), {
    mcpServers: {
      medialyst: {
        type: "http",
        url: "https://medialyst.ai/api/mcp",
        headersHelper: "$HOME/.newsjack/bin/newsjack auth headers",
      },
    },
  });
  writeFileSync(join(pluginRoot, "README.md"), "# Newsjack\n");
  writeFileSync(
    join(pluginRoot, "icon.svg"),
    '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 100 100"><rect width="100" height="100"/></svg>\n',
  );
  writeFileSync(
    join(pluginRoot, "skills", "sample", "SKILL.md"),
    "---\nname: sample\ndescription: Exercise the sample workflow.\n---\n\nFollow the sample workflow.\n",
  );
  return repoRoot;
}

test("normalizes a release tag to the tag core", () => {
  assert.deepEqual(normalizeReleaseVersion("v0.0.0-test.1"), {
    releaseVersion: "v0.0.0-test.1",
    coreVersion: "0.0.0",
  });
  assert.throws(() => normalizeReleaseVersion("latest"), /not a supported v\* semver/);
});

test("generates the portable OpenAI manifest from Claude metadata", () => {
  const manifest = generateOpenAIManifest({
    claudeManifest,
    marketplaceEntry,
    releaseVersion: "v9.8.7-test.1",
  });

  assert.equal(manifest.$schema, PLUGIN_SCHEMA_URL);
  assert.equal(manifest.name, "newsjack");
  assert.equal(manifest.version, "9.8.7");
  assert.equal(
    manifest.extensions["com.openai"].interface.shortDescription,
    "PR and AI-visibility skills",
  );
  assert.equal(
    manifest.extensions["com.openai"].interface.supportURL,
    "https://github.com/elvisun/newsjack/issues",
  );
  assert.equal(
    manifest.extensions["com.openai"].interface.termsOfServiceURL,
    "https://github.com/elvisun/newsjack/blob/main/LICENSE",
  );
  assert.equal(manifest.extensions["com.openai"].interface.logo, "./assets/icon.svg");
});

test("converts the shared MCP config to the portable schema", () => {
  const mcp = generatePortableMCP({
    mcpServers: {
      medialyst: {
        type: "http",
        url: "https://medialyst.ai/api/mcp",
        headersHelper: "$HOME/.newsjack/bin/newsjack auth headers",
      },
    },
  });

  assert.equal(mcp.$schema, MCP_SCHEMA_URL);
  assert.deepEqual(mcp.mcpServers.medialyst, {
    type: "streamable-http",
    url: "https://medialyst.ai/api/mcp",
  });
});

test("builds and validates a complete OpenAI plugin archive", () => {
  const repoRoot = createFixtureRepo();
  const outRoot = join(repoRoot, ".tmp", "release");
  try {
    const built = buildOpenAIPluginArchive({
      repoRoot,
      outRoot,
      releaseVersion: "v9.8.7-test.1",
    });
    const validated = validateOpenAIPluginArchive({
      archivePath: built.archivePath,
      repoRoot,
      expectedReleaseVersion: "v9.8.7-test.1",
    });

    assert.equal(validated.skillCount, 1);
    assert.equal(validated.version, "9.8.7");
    assert.equal(built.archiveName, "newsjack-openai-plugin-v9.8.7-test.1.zip");
  } finally {
    rmSync(repoRoot, { recursive: true, force: true });
  }
});

test("validator rejects stray files", () => {
  const repoRoot = createFixtureRepo();
  const outRoot = join(repoRoot, ".tmp", "release");
  const tamperRoot = join(repoRoot, ".tmp", "tamper");
  try {
    const built = buildOpenAIPluginArchive({
      repoRoot,
      outRoot,
      releaseVersion: "v9.8.7-test.1",
    });
    mkdirSync(tamperRoot, { recursive: true });
    execFileSync("unzip", ["-qq", built.archivePath, "-d", tamperRoot]);
    writeFileSync(join(tamperRoot, "newsjack", ".DS_Store"), "stray\n");
    rmSync(built.archivePath);
    execFileSync("zip", ["-q", "-X", "-r", built.archivePath, "newsjack"], {
      cwd: tamperRoot,
    });

    assert.throws(
      () =>
        validateOpenAIPluginArchive({
          archivePath: built.archivePath,
          repoRoot,
          expectedReleaseVersion: "v9.8.7-test.1",
        }),
      /forbidden path/,
    );
  } finally {
    rmSync(repoRoot, { recursive: true, force: true });
  }
});
