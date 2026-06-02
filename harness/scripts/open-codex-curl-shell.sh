#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="${NEWSJACK_HARNESS_IMAGE:-newsjack-agent-harness:codex}"
DIST_DIR="${NEWSJACK_RELEASE_DIST:-.tmp/newsjack-codex-curl-dist}"
PORT="${NEWSJACK_LOCAL_INSTALL_PORT:-}"
CLEAN_NO_ENV=0

usage() {
  cat <<'USAGE'
Usage: harness/scripts/open-codex-curl-shell.sh [options]

Options:
  --clean-no-env   Open a Codex-only container with no repo mount and all
                   Newsjack/API-key env vars blanked. Use this to test the
                   first-run user experience when no .env is present.
  --help, -h       Show this help.
USAGE
}

while [[ "$#" -gt 0 ]]; do
  case "$1" in
    --clean-no-env)
      CLEAN_NO_ENV=1
      shift
      ;;
    --help|-h)
      usage
      exit 0
      ;;
    *)
      usage >&2
      exit 2
      ;;
  esac
done

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

if [[ "$CLEAN_NO_ENV" = "1" ]]; then
  cat <<MSG
Local Newsjack release server:
  http://127.0.0.1:$PORT

Opening a clean no-env Codex harness:
  - no repo mount
  - no .env
  - no ~/.newsjack credentials
  - Medialyst/X/Twitter/OpenAI env vars blank

Inside the container, run:
  curl -fsSL $release_base/install.sh | NEWSJACK_RELEASE_BASE=$release_base NEWSJACK_AUTO_UPDATE=0 bash

Then try:
  newsjack doctor --json | jq '{root_ok, auth, runtimes: .runtimes}'
  newsjack setup --schedule-runtime codex

Setup should start automatically during the installer when this shell has a TTY.
Use NEWSJACK_RUN_SETUP=0 on the installer command only when you want install-only debugging.

MSG

  docker run --rm -it \
    --workdir /home/newsjack \
    --env HOME=/home/newsjack \
    --env XDG_CONFIG_HOME=/home/newsjack/.config \
    --env XDG_CACHE_HOME=/home/newsjack/.cache \
    --env XDG_DATA_HOME=/home/newsjack/.local/share \
    --env PATH=/home/newsjack/.newsjack/bin:/usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin \
    --env NEWSJACK_AUTO_UPDATE=0 \
    --env OPENAI_API_KEY= \
    --env MEDIALYST_API_KEY= \
    --env MEDIALYST_API_BASE= \
    --env MEDIALYST_NEWS_PATH= \
    --env X_BEARER_TOKEN= \
    --env TWITTER_BEARER_TOKEN= \
    --env X_API_BEARER_TOKEN= \
    --env TWITTER_API_BEARER_TOKEN= \
    "$IMAGE" \
    bash -c "
      cat <<'INNER'
Clean no-env Codex harness is ready.

Preflight:
  test ! -e .env
  test ! -e /repo
  test ! -e \"\$HOME/.newsjack/credentials.json\"

Run the installer:

  curl -fsSL $release_base/install.sh | NEWSJACK_RELEASE_BASE=$release_base NEWSJACK_AUTO_UPDATE=0 bash

Setup should start automatically. Then inspect or rerun setup:

  newsjack doctor --json | jq '{root_ok, auth, runtimes: .runtimes}'
  newsjack setup --schedule-runtime codex

INNER
      exec bash
    "
  exit 0
fi

cat <<MSG
Local Newsjack release server:
  http://127.0.0.1:$PORT

Inside the container, run:
  curl -fsSL $release_base/install.sh | NEWSJACK_RELEASE_BASE=$release_base NEWSJACK_RUNTIMES=codex NEWSJACK_INSTALL_MCP=0 NEWSJACK_AUTO_UPDATE=0 bash

Then try:
  newsjack doctor --json | jq .
  NEWSJACK_USE_INSTALLED=1 NEWSJACK_RUN_DIR=/tmp/newsjack-bluebottle-mock fixtures/newsjack-detector-agent/scripts/run-one-profile.sh bluebottle "specialty coffee" profile.bluebottle.json --mock
  codex

Setup should start automatically during the installer when this shell has a TTY.
Use NEWSJACK_RUN_SETUP=0 on the installer command only when you want install-only debugging.

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

Setup should start automatically. Then try:

  newsjack doctor --json | jq .
  NEWSJACK_USE_INSTALLED=1 NEWSJACK_RUN_DIR=/tmp/newsjack-bluebottle-mock fixtures/newsjack-detector-agent/scripts/run-one-profile.sh bluebottle \"specialty coffee\" profile.bluebottle.json --mock
  codex

INNER
    exec bash
  "
