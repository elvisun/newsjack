import assert from "node:assert/strict";
import { readdir, readFile } from "node:fs/promises";
import path from "node:path";
import { fileURLToPath } from "node:url";
import { test } from "node:test";
import ts from "typescript";

const siteRoot = fileURLToPath(new URL("../", import.meta.url));
const contentfulModule = path.join(siteRoot, "lib", "contentful.ts");
const allowedContentTypeConstant = "NEWSJACK_BLOG_POST_CONTENT_TYPE";
const sourceExtensions = new Set([
  ".cjs",
  ".cts",
  ".js",
  ".jsx",
  ".mjs",
  ".mts",
  ".ts",
  ".tsx",
]);
const ignoredDirectories = new Set([
  ".next",
  "build",
  "coverage",
  "dist",
  "node_modules",
  "out",
]);

async function sourceFiles(directory) {
  const entries = await readdir(directory, { withFileTypes: true });
  const files = [];

  for (const entry of entries) {
    if (entry.isDirectory()) {
      if (!ignoredDirectories.has(entry.name)) {
        files.push(...(await sourceFiles(path.join(directory, entry.name))));
      }
    } else if (sourceExtensions.has(path.extname(entry.name))) {
      files.push(path.join(directory, entry.name));
    }
  }

  return files;
}

function calledName(node) {
  if (ts.isIdentifier(node)) return node.text;
  if (ts.isPropertyAccessExpression(node)) return node.name.text;
  if (
    ts.isElementAccessExpression(node) &&
    ts.isStringLiteral(node.argumentExpression)
  ) {
    return node.argumentExpression.text;
  }
  return undefined;
}

function propertyName(node) {
  if (ts.isIdentifier(node) || ts.isStringLiteral(node)) return node.text;
  return undefined;
}

test("Contentful queries use the content-type allowlist and stay in the Contentful module", async () => {
  const violations = [];

  for (const file of await sourceFiles(siteRoot)) {
    const source = await readFile(file, "utf8");
    const relativePath = path.relative(siteRoot, file);
    const sourceFile = ts.createSourceFile(
      file,
      source,
      ts.ScriptTarget.Latest,
      true,
    );

    function visit(node) {
      if (ts.isCallExpression(node)) {
        const name = calledName(node.expression);

        if (name === "createClient" && file !== contentfulModule) {
          violations.push(`${relativePath}: creates a Contentful client directly`);
        }

        if (name === "getEntries") {
          if (file !== contentfulModule) {
            violations.push(`${relativePath}: queries Contentful directly`);
          } else {
            const query = node.arguments[0];
            const contentTypeProperties =
              query && ts.isObjectLiteralExpression(query)
                ? query.properties.filter(
                    (property) =>
                      ts.isPropertyAssignment(property) &&
                      propertyName(property.name) === "content_type",
                  )
                : [];
            const contentType = contentTypeProperties[0];

            if (
              contentTypeProperties.length !== 1 ||
              !ts.isPropertyAssignment(contentType) ||
              !ts.isIdentifier(contentType.initializer) ||
              contentType.initializer.text !== allowedContentTypeConstant
            ) {
              violations.push(
                `lib/contentful.ts: every query must set content_type to ${allowedContentTypeConstant}`,
              );
            }
          }
        }
      }

      ts.forEachChild(node, visit);
    }

    visit(sourceFile);
  }

  assert.deepEqual(violations, []);
});
