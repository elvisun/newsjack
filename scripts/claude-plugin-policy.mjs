// File rules the Claude plugin directory applies to organizations that are not
// Verified: a package may hold only Markdown, text, JSON, YAML and a few image
// formats, with no scripts or programs. The package sync script uses these rules
// to choose what to copy; the validation script uses them to check the result.

import { basename, extname } from "node:path";

const TEXT_EXTENSIONS = new Set([".md", ".txt", ".json", ".yaml", ".yml"]);
const IMAGE_SIGNATURES = {
  ".png": [[0x89, 0x50, 0x4e, 0x47]],
  ".jpg": [[0xff, 0xd8, 0xff]],
  ".jpeg": [[0xff, 0xd8, 0xff]],
  ".gif": [[0x47, 0x49, 0x46, 0x38]],
  ".webp": [[0x52, 0x49, 0x46, 0x46]],
  ".ico": [[0x00, 0x00, 0x01, 0x00]],
};
// Extensionless text files the directory names as allowed.
const TEXT_BASENAMES = new Set(["LICENSE", "README", ".gitignore"]);
const PACKAGE_MANAGER_FILES = new Set([
  "package.json",
  "package-lock.json",
  "pnpm-lock.yaml",
  "pnpm-workspace.yaml",
  "yarn.lock",
  ".npmrc",
  ".yarnrc.yml",
]);

function startsWith(bytes, signature) {
  return signature.every((value, index) => bytes[index] === value);
}

/**
 * Returns null when the directory accepts the file, otherwise the reason it does not.
 * `path` is relative to the plugin root; `bytes` is the file content as a Buffer.
 */
export function directoryPolicyViolation(path, bytes) {
  const name = basename(path);
  const extension = extname(name).toLowerCase();

  if (!/^[\x20-\x7e]+$/.test(path) || /[<>:"|?*\\]/.test(path)) {
    return "path has a character the directory does not accept";
  }
  if (PACKAGE_MANAGER_FILES.has(name)) {
    return "package manager configuration file";
  }
  if (bytes.length >= 2 && bytes[0] === 0x23 && bytes[1] === 0x21) {
    return "starts with #!";
  }

  const signatures = IMAGE_SIGNATURES[extension];
  if (signatures) {
    return signatures.some((signature) => startsWith(bytes, signature))
      ? null
      : `content is not a real ${extension} image`;
  }

  if (!TEXT_EXTENSIONS.has(extension) && !TEXT_BASENAMES.has(name)) {
    return "not Markdown, text, JSON, YAML, or a PNG, JPEG, GIF, WebP, or ICO image";
  }
  if (bytes.includes(0)) {
    return "contains a NUL byte";
  }
  try {
    new TextDecoder("utf-8", { fatal: true }).decode(bytes);
  } catch {
    return "is not valid UTF-8";
  }
  return null;
}
