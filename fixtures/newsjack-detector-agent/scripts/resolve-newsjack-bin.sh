#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
REPO_BIN="$FIXTURE_DIR/../../bin/newsjack"
NEWSJACK_HOME="${NEWSJACK_HOME:-$HOME/.newsjack}"
INSTALLED_BIN="$NEWSJACK_HOME/bin/newsjack"

if [[ -n "${NEWSJACK_BIN:-}" ]]; then
  echo "$NEWSJACK_BIN"
  exit 0
fi

if [[ "${NEWSJACK_USE_INSTALLED:-0}" = "1" ]]; then
  if [[ -x "$INSTALLED_BIN" ]]; then
    echo "$INSTALLED_BIN"
    exit 0
  fi
  echo "NEWSJACK_USE_INSTALLED=1 but installed binary is missing: $INSTALLED_BIN" >&2
  exit 1
fi

if command -v go >/dev/null 2>&1; then
  echo "$REPO_BIN"
  exit 0
fi

if [[ -x "$INSTALLED_BIN" ]]; then
  echo "$INSTALLED_BIN"
  exit 0
fi

echo "Go is not on PATH and no installed newsjack binary was found at $INSTALLED_BIN" >&2
echo "Install Newsjack first, set NEWSJACK_BIN, or run with Go available for the source shim." >&2
exit 1
