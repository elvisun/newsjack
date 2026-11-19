import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import {
  copyFileSync,
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
  validateManifestShape,
  validateMCPShape,
  validateOpenAIPluginArchive,
  validatePNG,
} from "./lib/openai-plugin.mjs";

const claudeManifest = {
  name: "newsjack",
  displayName: "Newsjack",
  version: "1.2.3",
  description:
    "34 open-source PR and AI-visibility skills for strategy, monitoring, and research.",
  icon: "./icon.png",
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
  copyFileSync(
    new URL("../plugins/newsjack/icon.png", import.meta.url),
    join(pluginRoot, "icon.png"),
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
  assert.deepEqual(manifest.extensions["com.openai"].interface.defaultPrompt, [
    "Find current newsjacking opportunities for my company.",
    "Turn this company update into journalist-shaped story angles.",
    "Fact-check this press release before I send it.",
  ]);
  assert.equal(manifest.extensions["com.openai"].interface.logo, "./assets/icon.png");
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
    extensions: {
      "com.openai": {
        auth: {
          type: "oauth",
          client: {
            mode: "dcr",
          },
        },
      },
    },
  });
});

test("enforces final OpenAI submission metadata limits", () => {
  const manifest = generateOpenAIManifest({
    claudeManifest,
    marketplaceEntry,
    releaseVersion: "v9.8.7-test.1",
  });
  validateManifestShape(manifest);

  const cases = [
    ["displayName", "x".repeat(31), /displayName must be 1-30 characters/],
    ["shortDescription", "x".repeat(31), /shortDescription must be one line and 1-30 characters/],
    ["developerName", "x".repeat(81), /developerName must be 1-80 characters/],
    ["defaultPrompt", ["x".repeat(129)], /defaultPrompt entries must be 1-128 characters/],
    ["defaultPrompt", ["one", "two", "three", "four"], /defaultPrompt must contain at most 3 prompts/],
    ["supportURL", "http://example.com/support", /supportURL must be an HTTPS URL/],
  ];
  for (const [field, value, expectedError] of cases) {
    const invalid = structuredClone(manifest);
    invalid.extensions["com.openai"].interface[field] = value;
    assert.throws(() => validateManifestShape(invalid), expectedError);
  }
});

test("requires OAuth DCR for the Medialyst server", () => {
  const mcp = generatePortableMCP({
    mcpServers: {
      medialyst: {
        type: "http",
        url: "https://medialyst.ai/api/mcp",
      },
    },
  });
  validateMCPShape(mcp);
  delete mcp.mcpServers.medialyst.extensions;
  assert.throws(
    () => validateMCPShape(mcp),
    /must declare OpenAI OAuth with dynamic client registration/,
  );
});

test("validates PNG signature, IHDR dimensions, and submission size limits", () => {
  validatePNG(new URL("../plugins/newsjack/icon.png", import.meta.url));
  const root = mkdtempSync(join(tmpdir(), "newsjack-openai-png-test-"));
  try {
    const invalidSignature = join(root, "invalid-signature.png");
    writeFileSync(invalidSignature, Buffer.alloc(33));
    assert.throws(() => validatePNG(invalidSignature), /invalid PNG signature/);

    const invalidDimensions = join(root, "invalid-dimensions.png");
    const pngHeader = Buffer.alloc(33);
    Buffer.from([0x89, 0x50, 0x4e, 0x47, 0x0d, 0x0a, 0x1a, 0x0a]).copy(pngHeader);
    pngHeader.writeUInt32BE(13, 8);
    pngHeader.write("IHDR", 12, "ascii");
    pngHeader.writeUInt32BE(47, 16);
    pngHeader.writeUInt32BE(48, 20);
    writeFileSync(invalidDimensions, pngHeader);
    assert.throws(() => validatePNG(invalidDimensions), /must be square/);

    const oversized = join(root, "oversized.png");
    writeFileSync(oversized, Buffer.alloc(5 * 1024 * 1024 + 1));
    assert.throws(() => validatePNG(oversized), /exceeds 5 MiB/);
  } finally {
    rmSync(root, { recursive: true, force: true });
  }
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
