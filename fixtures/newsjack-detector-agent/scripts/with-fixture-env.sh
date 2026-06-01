#!/usr/bin/env bash
set -euo pipefail

FIXTURE_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPO_ROOT="$(cd "$FIXTURE_DIR/../.." && pwd)"

load_env() {
  local file="$1"
  if [[ -f "$file" ]]; then
    set -a
    # shellcheck disable=SC1090
    source "$file"
    set +a
  fi
}

load_env "$REPO_ROOT/.env"
load_env "$FIXTURE_DIR/.env"

cd "$FIXTURE_DIR"

if [[ "$#" -eq 0 ]]; then
  exec "${SHELL:-/bin/bash}"
fi

exec "$@"
