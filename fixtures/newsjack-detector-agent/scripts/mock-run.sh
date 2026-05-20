#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

exec "$SCRIPT_DIR/agent-env.sh" \
  python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run \
  "AI regulation" \
  --profile profile.acme-ai.json \
  --mock \
  --emit json
