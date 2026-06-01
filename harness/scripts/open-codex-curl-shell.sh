#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="${NEWSJACK_HARNESS_IMAGE:-newsjack-agent-harness:codex}"
DIST_DIR="${NEWSJACK_RELEASE_DIST:-.tmp/newsjack-codex-curl-dist}"
PORT="${NEWSJACK_LOCAL_INSTALL_PORT:-}"

if ! command -v docker >/dev/null 2>&1; then
  echo "docker is required" >&2
  exit 1
fi

if ! docker image inspect "$IMAGE" >/dev/null 2>&1; then
  "$REPO_DIR/harness/scripts/build-image.sh" --harness codex --image "$IMAGE"
fi

if [[ -z "$PORT" ]]; then
  PORT="$(python3 - <<'PY'
import socket
s = socket.socket()
s.bind(("127.0.0.1", 0))
print(s.getsockname()[1])
s.close()
PY
)"
fi

(
  cd "$REPO_DIR"
  NEWSJACK_VERSION="v0.1.0-local-codex" \
  NEWSJACK_RELEASE_DIST="$DIST_DIR" \
    node scripts/build-release-dist.mjs
)

python3 -m http.server "$PORT" --bind 127.0.0.1 --directory "$REPO_DIR/$DIST_DIR" >/tmp/newsjack-codex-curl-server.log 2>&1 &
server_pid=$!
trap 'kill "$server_pid" >/dev/null 2>&1 || true' EXIT

release_base="http://host.docker.internal:$PORT"

cat <<MSG
Local Newsjack release server:
  http://127.0.0.1:$PORT

Inside the container, run:
  curl -fsSL $release_base/install.sh | NEWSJACK_RELEASE_BASE=$release_base NEWSJACK_RUNTIMES=codex NEWSJACK_INSTALL_MCP=0 NEWSJACK_AUTO_UPDATE=0 bash

Then try:
  newsjack doctor --json | jq .
  NEWSJACK_USE_INSTALLED=1 NEWSJACK_RUN_DIR=/tmp/newsjack-bluebottle-mock fixtures/newsjack-detector-agent/scripts/run-one-profile.sh bluebottle "specialty coffee" profile.bluebottle.json --mock
  codex

MSG

docker run --rm -it \
  --mount "type=bind,src=${REPO_DIR},dst=/repo,readonly" \
  --workdir /repo \
  --env HOME=/home/newsjack \
  --env XDG_CONFIG_HOME=/home/newsjack/.config \
  --env XDG_CACHE_HOME=/home/newsjack/.cache \
  --env XDG_DATA_HOME=/home/newsjack/.local/share \
  --env PATH=/home/newsjack/.newsjack/bin:/usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin \
  --env NEWSJACK_AUTO_UPDATE=0 \
  --env OPENAI_API_KEY \
  "$IMAGE" \
  bash -c "
    cat <<'INNER'
Codex harness is ready. Run this install command yourself:

  curl -fsSL $release_base/install.sh | NEWSJACK_RELEASE_BASE=$release_base NEWSJACK_RUNTIMES=codex NEWSJACK_INSTALL_MCP=0 NEWSJACK_AUTO_UPDATE=0 bash

Then try:

  newsjack doctor --json | jq .
  NEWSJACK_USE_INSTALLED=1 NEWSJACK_RUN_DIR=/tmp/newsjack-bluebottle-mock fixtures/newsjack-detector-agent/scripts/run-one-profile.sh bluebottle \"specialty coffee\" profile.bluebottle.json --mock
  codex

INNER
    exec bash
  "
