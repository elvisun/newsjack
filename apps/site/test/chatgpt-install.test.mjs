import assert from "node:assert/strict";
import { readFile, stat } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "node:test";

import { CHATGPT_INSTALL_PROMPT } from "../lib/chatgpt-install.ts";

const siteRoot = fileURLToPath(new URL("../", import.meta.url));
const repoRoot = path.resolve(siteRoot, "../..");

async function readJson(relativePath) {
  return JSON.parse(await readFile(path.join(repoRoot, relativePath), "utf8"));
}

test("ChatGPT install prompt stays in the README", async () => {
  const readme = await readFile(path.join(repoRoot, "README.md"), "utf8");

  assert.ok(
    readme.includes(CHATGPT_INSTALL_PROMPT),
    "README.md must contain the exact ChatGPT install prompt",
  );
});

test("newsjack@newsjack still resolves to an existing plugin directory", async () => {
  const marketplace = await readJson(".claude-plugin/marketplace.json");
  const plugin = marketplace.plugins?.find(({ name }) => name === "newsjack");

  assert.equal(marketplace.name, "newsjack");
  assert.ok(plugin, "marketplace must contain a plugin named newsjack");
  assert.equal(
    (await stat(path.resolve(repoRoot, plugin.source))).isDirectory(),
    true,
    "the newsjack plugin source must be an existing directory",
  );
});

test("Newsjack plugin still registers the Medialyst MCP server", async () => {
  const mcp = await readJson("plugins/newsjack/.mcp.json");

  assert.ok(
    mcp.mcpServers?.medialyst,
    "plugins/newsjack/.mcp.json must contain a medialyst server",
  );
});
