#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FIXTURE_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
NEWSJACK_BIN="$("$SCRIPT_DIR/resolve-newsjack-bin.sh")"
QUERY="${NEWSJACK_QUERY:-${1:-AI search visibility}}"
PROFILE="${NEWSJACK_PROFILE:-${2:-profile.localfalcon.json}}"
if [[ "$#" -gt 0 ]]; then shift; fi
if [[ "$#" -gt 0 ]]; then shift; fi

if [[ "$PROFILE" != /* ]]; then
  if [[ -f "$FIXTURE_DIR/$PROFILE" ]]; then
    PROFILE="$FIXTURE_DIR/$PROFILE"
  elif [[ -f "$PROFILE" ]]; then
    PROFILE="$(cd "$(dirname "$PROFILE")" && pwd)/$(basename "$PROFILE")"
  fi
fi

exec "$SCRIPT_DIR/with-fixture-env.sh" \
  "$NEWSJACK_BIN" detector run \
  "$QUERY" \
  --profile "$PROFILE" \
  "$@" \
  --mock \
  --emit json
