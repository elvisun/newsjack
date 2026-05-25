#!/usr/bin/env python3
"""Stdio bridge for MCP clients without dynamic HTTP header helpers."""

from __future__ import annotations

import os
import shutil
import sys
from pathlib import Path

SCRIPT_DIR = Path(__file__).resolve().parent
if str(SCRIPT_DIR) not in sys.path:
    sys.path.insert(0, str(SCRIPT_DIR))

from medialyst_auth import ENV_KEY, load_api_key  # noqa: E402


MCP_URL = "https://medialyst.ai/api/mcp"


def main() -> int:
    api_key, source = load_api_key()
    if not api_key:
        print(
            "Medialyst API key not found. Run: ~/.newsjack/bin/newsjack login",
            file=sys.stderr,
        )
        return 1

    npx = shutil.which("npx")
    if not npx:
        print(
            "npx is required for the Medialyst MCP stdio bridge. Install Node.js, then retry.",
            file=sys.stderr,
        )
        return 1

    env = os.environ.copy()
    env[ENV_KEY] = api_key
    if os.environ.get("NEWSJACK_AUTH_DEBUG"):
        print(f"Loaded Medialyst API key from {source}", file=sys.stderr)

    os.execvpe(
        npx,
        [
            npx,
            "-y",
            "mcp-remote",
            MCP_URL,
            "--header",
            f"Authorization: Bearer ${{{ENV_KEY}}}",
        ],
        env,
    )
    return 127


if __name__ == "__main__":
    raise SystemExit(main())
