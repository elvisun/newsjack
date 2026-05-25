#!/usr/bin/env python3
"""Local Medialyst auth helper for Newsjack MCP clients."""

from __future__ import annotations

import argparse
import getpass
import json
import os
import stat
import sys
from datetime import datetime, timezone
from pathlib import Path
from typing import Iterable


APP_DIR_NAME = ".newsjack"
CREDENTIALS_FILE = "credentials.json"
ENV_KEY = "MEDIALYST_API_KEY"


def _newsjack_home() -> Path:
    override = os.environ.get("NEWSJACK_HOME")
    if override:
        return Path(override).expanduser()
    return Path.home() / APP_DIR_NAME


def _credentials_path() -> Path:
    return _newsjack_home() / CREDENTIALS_FILE


def _candidate_env_paths() -> Iterable[Path]:
    cwd = Path.cwd().resolve()
    for directory in (cwd, *cwd.parents):
        yield directory / ".env"
    yield _newsjack_home() / ".env"


def _read_dotenv_key(path: Path) -> str | None:
    try:
        text = path.read_text(encoding="utf-8")
    except OSError:
        return None

    for raw_line in text.splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        if key.strip() != ENV_KEY:
            continue
        value = value.strip().strip("\"'")
        return value or None
    return None


def _read_saved_key(path: Path) -> str | None:
    try:
        data = json.loads(path.read_text(encoding="utf-8"))
    except (OSError, json.JSONDecodeError):
        return None

    medialyst = data.get("medialyst")
    if isinstance(medialyst, dict):
        value = medialyst.get("api_key")
        if isinstance(value, str) and value.strip():
            return value.strip()

    value = data.get("MEDIALYST_API_KEY")
    if isinstance(value, str) and value.strip():
        return value.strip()

    return None


def load_api_key() -> tuple[str | None, str | None]:
    env_value = os.environ.get(ENV_KEY)
    if env_value:
        return env_value.strip(), f"environment:{ENV_KEY}"

    credentials = _credentials_path()
    value = _read_saved_key(credentials)
    if value:
        return value, f"credentials:{credentials}"

    for path in _candidate_env_paths():
        value = _read_dotenv_key(path)
        if value:
            return value, f"dotenv:{path}"

    return None, None


def _validate_key(api_key: str) -> None:
    if not api_key.startswith("mlst_"):
        raise ValueError("Medialyst API keys should start with 'mlst_'.")
    if len(api_key) < 12:
        raise ValueError("Medialyst API key is too short.")


def _write_credentials(api_key: str) -> Path:
    path = _credentials_path()
    path.parent.mkdir(mode=0o700, parents=True, exist_ok=True)
    payload = {
        "medialyst": {
            "api_key": api_key,
            "created_at": datetime.now(timezone.utc).isoformat(),
            "source": "newsjack-local-login",
        }
    }
    path.write_text(json.dumps(payload, indent=2) + "\n", encoding="utf-8")
    path.chmod(stat.S_IRUSR | stat.S_IWUSR)
    return path


def cmd_headers(_: argparse.Namespace) -> int:
    api_key, source = load_api_key()
    if not api_key:
        print(
            "Medialyst API key not found. Run: ~/.newsjack/bin/newsjack login",
            file=sys.stderr,
        )
        return 1
    print(json.dumps({"Authorization": f"Bearer {api_key}"}))
    if os.environ.get("NEWSJACK_AUTH_DEBUG"):
        print(f"Loaded Medialyst API key from {source}", file=sys.stderr)
    return 0


def cmd_login(args: argparse.Namespace) -> int:
    api_key = args.key
    if not api_key:
        api_key = getpass.getpass("Medialyst API key: ").strip()
    try:
        _validate_key(api_key)
    except ValueError as exc:
        print(f"Invalid key: {exc}", file=sys.stderr)
        return 2

    path = _write_credentials(api_key)
    print(f"Saved Medialyst credentials to {path}")
    print("Claude Code can now use the repo .mcp.json without MEDIALYST_API_KEY exports.")
    return 0


def cmd_status(_: argparse.Namespace) -> int:
    api_key, source = load_api_key()
    if not api_key:
        print(json.dumps({"configured": False, "source": None}, indent=2))
        return 1
    print(json.dumps({"configured": True, "source": source}, indent=2))
    return 0


def cmd_logout(_: argparse.Namespace) -> int:
    path = _credentials_path()
    try:
        path.unlink()
    except FileNotFoundError:
        print("No saved Medialyst credentials found.")
        return 0
    print(f"Removed saved Medialyst credentials at {path}")
    return 0


def build_parser() -> argparse.ArgumentParser:
    parser = argparse.ArgumentParser(description="Manage local Medialyst auth for Newsjack.")
    subparsers = parser.add_subparsers(dest="command", required=True)

    login = subparsers.add_parser("login", help="Save a Medialyst API key locally.")
    login.add_argument(
        "--key",
        help="API key to save. Prefer the interactive prompt to keep keys out of shell history.",
    )
    login.set_defaults(func=cmd_login)

    status = subparsers.add_parser("status", help="Show whether Medialyst auth is configured.")
    status.set_defaults(func=cmd_status)

    headers = subparsers.add_parser("headers", help="Print MCP headers as JSON for Claude Code.")
    headers.set_defaults(func=cmd_headers)

    logout = subparsers.add_parser("logout", help="Remove saved Medialyst credentials.")
    logout.set_defaults(func=cmd_logout)

    return parser


def main() -> int:
    parser = build_parser()
    args = parser.parse_args()
    return args.func(args)


if __name__ == "__main__":
    raise SystemExit(main())
