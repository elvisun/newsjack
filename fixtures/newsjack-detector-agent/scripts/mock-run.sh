#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
QUERY="${NEWSJACK_QUERY:-${1:-AI search visibility}}"
PROFILE="${NEWSJACK_PROFILE:-${2:-profile.localfalcon.json}}"
if [[ "$#" -gt 0 ]]; then shift; fi
if [[ "$#" -gt 0 ]]; then shift; fi

exec "$SCRIPT_DIR/agent-env.sh" \
  python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run \
  "$QUERY" \
  --profile "$PROFILE" \
  "$@" \
  --mock \
  --emit json
