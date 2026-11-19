import assert from "node:assert/strict";
import { execFileSync } from "node:child_process";
import {
  copyFileSync,
  existsSync,
  mkdirSync,
  mkdtempSync,
  readFileSync,
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
  buildOpenAIPluginArchives,
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
    join(pluginRoot, "skills", "ETHICS.md"),
    "# Ethics\n\n`skills/ETHICS.md` wins.\n",
  );
  writeFileSync(
    join(pluginRoot, "skills", "WHY-NOT-SPAM.md"),
    "# Why not spam\n\nFollow `skills/WHY-NOT-SPAM.md`.\n",
  );
  writeFileSync(
    join(pluginRoot, "skills", "sample", "SKILL.md"),
    "---\nname: sample\ndescription: Exercise the sample workflow.\n---\n\nFollow `skills/ETHICS.md` and `skills/WHY-NOT-SPAM.md`.\n",
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
    manifest.extensions["com.openai"].interface.websiteURL,
    "https://newsjack.sh",
  );
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

test("generates the two portable MCP variants from the shared config", () => {
  const sourceMCP = {
    mcpServers: {
      medialyst: {
        type: "http",
        url: "https://medialyst.ai/api/mcp",
        headersHelper: "$HOME/.newsjack/bin/newsjack auth headers",
      },
    },
  };
  const chatgptMCP = generatePortableMCP(sourceMCP, "chatgpt");
  const submissionMCP = generatePortableMCP(sourceMCP, "submission");

  assert.equal(chatgptMCP.$schema, MCP_SCHEMA_URL);
  assert.deepEqual(chatgptMCP.mcpServers.medialyst, {
    type: "streamable-http",
    url: "https://medialyst.ai/api/mcp",
  });
  assert.deepEqual(Object.keys(chatgptMCP.mcpServers.medialyst).sort(), [
    "type",
    "url",
  ]);

  assert.deepEqual(submissionMCP.mcpServers.medialyst, {
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
  assert.deepEqual(Object.keys(submissionMCP.mcpServers.medialyst).sort(), [
    "extensions",
    "type",
    "url",
  ]);
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

test("requires the exact Medialyst key set for each archive variant", () => {
  const sourceMCP = {
    mcpServers: {
      medialyst: {
        type: "http",
        url: "https://medialyst.ai/api/mcp",
      },
    },
  };
  const chatgptMCP = generatePortableMCP(sourceMCP, "chatgpt");
  validateMCPShape(chatgptMCP, "chatgpt");
  const chatgptWithExtensions = structuredClone(chatgptMCP);
  chatgptWithExtensions.mcpServers.medialyst.extensions = {};
  assert.throws(
    () => validateMCPShape(chatgptWithExtensions, "chatgpt"),
    /must contain exactly type, url for the chatgpt variant/,
  );

  const submissionMCP = generatePortableMCP(sourceMCP, "submission");
  validateMCPShape(submissionMCP, "submission");
  const submissionWithoutExtensions = structuredClone(submissionMCP);
  delete submissionWithoutExtensions.mcpServers.medialyst.extensions;
  assert.throws(
    () => validateMCPShape(submissionWithoutExtensions, "submission"),
    /must contain exactly extensions, type, url for the submission variant/,
  );

  const invalidAuth = structuredClone(submissionMCP);
  invalidAuth.mcpServers.medialyst.extensions["com.openai"].auth.client.mode = "cimd";
  assert.throws(
    () => validateMCPShape(invalidAuth, "submission"),
    /must declare OpenAI OAuth with dynamic client registration/,
  );
  const extraExtensionOption = structuredClone(submissionMCP);
  extraExtensionOption.mcpServers.medialyst.extensions["com.openai"].extra = true;
  assert.throws(
    () => validateMCPShape(extraExtensionOption, "submission"),
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

test("builds and validates both OpenAI plugin archive variants", () => {
  const repoRoot = createFixtureRepo();
  const outRoot = join(repoRoot, ".tmp", "release");
  try {
    const archives = buildOpenAIPluginArchives({
      repoRoot,
      outRoot,
      releaseVersion: "v9.8.7-test.1",
    });
    const built = Object.fromEntries(archives.map((archive) => [archive.variant, archive]));
    const validated = Object.fromEntries(
      archives.map((archive) => [
        archive.variant,
        validateOpenAIPluginArchive({
          archivePath: archive.archivePath,
          repoRoot,
          expectedReleaseVersion: "v9.8.7-test.1",
          variant: archive.variant,
        }),
      ]),
    );

    assert.equal(validated.chatgpt.skillCount, 1);
    assert.equal(validated.chatgpt.version, "9.8.7");
    assert.equal(validated.submission.skillCount, 1);
    assert.equal(validated.submission.version, "9.8.7");
    assert.equal(
      built.chatgpt.archiveName,
      "newsjack-chatgpt-plugin-v9.8.7-test.1.zip",
    );
    assert.equal(
      built.submission.archiveName,
      "newsjack-openai-submission-v9.8.7-test.1.zip",
    );

    const chatgptEntries = execFileSync("unzip", ["-Z1", built.chatgpt.archivePath], {
      encoding: "utf8",
    }).split(/\r?\n/).filter(Boolean);
    const submissionEntries = execFileSync(
      "unzip",
      ["-Z1", built.submission.archivePath],
      { encoding: "utf8" },
    ).split(/\r?\n/).filter(Boolean);
    assert.deepEqual(submissionEntries, chatgptEntries);
    for (const entry of chatgptEntries) {
      if (entry.endsWith("/") || entry === "newsjack/mcp.json") {
        continue;
      }
      assert.deepEqual(
        execFileSync("unzip", ["-p", built.submission.archivePath, entry]),
        execFileSync("unzip", ["-p", built.chatgpt.archivePath, entry]),
        `${entry} must be identical across archive variants`,
      );
    }

    const extractRoot = join(repoRoot, ".tmp", "extract");
    execFileSync("unzip", ["-qq", built.chatgpt.archivePath, "-d", extractRoot]);
    const archivedSkills = join(extractRoot, "newsjack", "skills");
    assert.deepEqual(
      Object.keys(
        JSON.parse(readFileSync(join(extractRoot, "newsjack", "mcp.json"), "utf8"))
          .mcpServers.medialyst,
      ).sort(),
      ["type", "url"],
    );
    assert.equal(existsSync(join(archivedSkills, "ETHICS.md")), false);
    assert.equal(existsSync(join(archivedSkills, "WHY-NOT-SPAM.md")), false);
    assert.match(
      readFileSync(join(archivedSkills, "sample", "SKILL.md"), "utf8"),
      /`references\/ETHICS\.md` and `references\/WHY-NOT-SPAM\.md`/,
    );
    assert.match(
      readFileSync(
        join(archivedSkills, "sample", "references", "ETHICS.md"),
        "utf8",
      ),
      /`ETHICS\.md` wins/,
    );
    assert.match(
      readFileSync(
        join(archivedSkills, "sample", "references", "WHY-NOT-SPAM.md"),
        "utf8",
      ),
      /`WHY-NOT-SPAM\.md`/,
    );
    assert.match(
      readFileSync(
        join(repoRoot, "plugins", "newsjack", "skills", "sample", "SKILL.md"),
        "utf8",
      ),
      /`skills\/ETHICS\.md` and `skills\/WHY-NOT-SPAM\.md`/,
    );
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
