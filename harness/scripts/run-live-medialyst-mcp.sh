#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="${NEWSJACK_HARNESS_IMAGE:-newsjack-agent-harness:all}"
ENV_FILE="${NEWSJACK_HARNESS_ENV_FILE:-}"
RUNTIMES=()

usage() {
  cat <<'USAGE'
Usage: harness/scripts/run-live-medialyst-mcp.sh [options]

Options:
  --runtime <runtime>  Runtime target: all, codex, claude, openclaw, hermes.
                       Repeat to test multiple runtimes. Default: all.
  --image <image>      Harness container image (default: newsjack-agent-harness:all)
  --with-local-env     Load harness/.env.local into the container
  --env-file <path>    Load an explicit Docker env file into the container
  --help, -h           Show this help

This is an opt-in live test. It requires MEDIALYST_API_KEY through the host
environment or an ignored env file. The key is passed into Docker at runtime only.
USAGE
}

log() {
  printf '%s\n' "live-medialyst-mcp: $*" >&2
}

abs_path() {
  local path="$1"
  local dir
  dir="$(cd "$(dirname "$path")" && pwd)"
  printf '%s/%s\n' "$dir" "$(basename "$path")"
}

validate_env_file() {
  local file="$1"
  [ -f "$file" ] || { log "env file not found: $file"; exit 1; }

  local abs
  abs="$(abs_path "$file")"
  case "$abs" in
    "$REPO_DIR"/*)
      local rel="${abs#"$REPO_DIR"/}"
      if ! git -C "$REPO_DIR" check-ignore -q "$rel"; then
        log "refusing to pass repo env file that is not git-ignored: $rel"
        exit 1
      fi
      ;;
  esac
}

add_runtime() {
  local raw="$1"
  raw="${raw//claude-code/claude}"
  raw="${raw//claude_code/claude}"
  raw="${raw//cladue/claude}"
  case "$raw" in
    all)
      RUNTIMES=(codex claude openclaw hermes)
      ;;
    codex|claude|openclaw|hermes)
      RUNTIMES+=("$raw")
      ;;
    *)
      log "unsupported runtime: $raw"
      exit 2
      ;;
  esac
}

while [ "$#" -gt 0 ]; do
  case "$1" in
    --runtime)
      [ "$#" -ge 2 ] || { usage >&2; exit 2; }
      add_runtime "$2"
      shift 2
      ;;
    --image)
      [ "$#" -ge 2 ] || { usage >&2; exit 2; }
      IMAGE="$2"
      shift 2
      ;;
    --with-local-env)
      ENV_FILE="$REPO_DIR/harness/.env.local"
      shift
      ;;
    --env-file)
      [ "$#" -ge 2 ] || { usage >&2; exit 2; }
      ENV_FILE="$2"
      shift 2
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

if [ "${#RUNTIMES[@]}" -eq 0 ]; then
  add_runtime all
fi

if ! command -v docker >/dev/null 2>&1; then
  log "docker is required to run the live harness"
  exit 1
fi

base_docker_args=(
  --rm
  --interactive
  --mount "type=bind,src=${REPO_DIR},dst=/repo,readonly"
  --workdir /tmp
)

if [ -n "$ENV_FILE" ]; then
  validate_env_file "$ENV_FILE"
  base_docker_args+=(--env-file "$ENV_FILE")
  log "loading env file into container: ${ENV_FILE#$REPO_DIR/}"
else
  base_docker_args+=(--env MEDIALYST_API_KEY)
fi

for runtime in "${RUNTIMES[@]}"; do
  log "running clean live Medialyst/MCP check for ${runtime}"
  docker run \
    "${base_docker_args[@]}" \
    --env "HOME=/tmp/newsjack-home" \
    --env "XDG_CONFIG_HOME=/tmp/newsjack-home/.config" \
    --env "XDG_CACHE_HOME=/tmp/newsjack-home/.cache" \
    --env "XDG_DATA_HOME=/tmp/newsjack-home/.local/share" \
    --env "PATH=/tmp/newsjack-home/.newsjack/bin:/usr/local/go/bin:/usr/local/bin:/usr/local/sbin:/usr/sbin:/usr/bin:/sbin:/bin" \
    --env "NEWSJACK_RUNTIMES=${runtime}" \
    --env "NEWSJACK_NO_AUTO_UPDATE=1" \
    --env "NEWSJACK_IGNORE_DOTENV=1" \
    --env "NEWSJACK_RUN_SETUP=0" \
    "${IMAGE}" bash -s <<'EOF'
set -euo pipefail

runtime="${NEWSJACK_RUNTIMES}"
if [ -z "${MEDIALYST_API_KEY:-}" ]; then
  printf '%s\n' "missing MEDIALYST_API_KEY for live harness" >&2
  exit 1
fi

log() {
  printf '%s\n' "live-medialyst-mcp(container:${runtime}): $*" >&2
}

assert_runtime_available() {
  command -v "$runtime" >/dev/null 2>&1 || {
    log "runtime CLI not available: $runtime"
    exit 1
  }
}

sanitize_source() {
  mkdir -p /tmp/newsjack-source
  tar \
    --exclude .git \
    --exclude .tmp \
    --exclude .mcp.json \
    --exclude .env \
    --exclude '.env.*' \
    --exclude node_modules \
    --exclude .next \
    -C /repo -cf - . | tar -C /tmp/newsjack-source -xf -
  test ! -e /tmp/newsjack-source/.mcp.json
}

verify_runtime_mcp() {
  case "$runtime" in
    codex)
      codex mcp list > /tmp/newsjack-mcp.txt
      grep -Eq '^medialyst[[:space:]]+' /tmp/newsjack-mcp.txt
      grep -q 'newsjack' /tmp/newsjack-mcp.txt
      grep -q 'mcp-bridge' /tmp/newsjack-mcp.txt
      ;;
    claude)
      claude mcp list > /tmp/newsjack-mcp.txt
      grep -Eq '^medialyst:' /tmp/newsjack-mcp.txt
      grep -q 'newsjack mcp-bridge' /tmp/newsjack-mcp.txt
      ;;
    openclaw)
      openclaw mcp list > /tmp/newsjack-mcp.txt
      grep -q 'medialyst' /tmp/newsjack-mcp.txt
      openclaw mcp show medialyst > /tmp/newsjack-mcp-show.txt
      grep -q 'newsjack' /tmp/newsjack-mcp-show.txt
      grep -q 'mcp-bridge' /tmp/newsjack-mcp-show.txt
      ;;
    hermes)
      test -f "$HOME/.hermes/config.yaml"
      grep -q 'mcp_servers:' "$HOME/.hermes/config.yaml"
      grep -q 'medialyst:' "$HOME/.hermes/config.yaml"
      grep -q 'newsjack' "$HOME/.hermes/config.yaml"
      grep -q 'mcp-bridge' "$HOME/.hermes/config.yaml"
      ;;
  esac
}

assert_runtime_available
test "$(pwd)" = "/tmp"
test ! -e /tmp/.mcp.json

log "building local CLI and sanitized source bundle"
sanitize_source
(cd /repo/apps/cli && go build -buildvcs=false -o /tmp/newsjack ./cmd/newsjack)

log "installing without install-time MCP"
NEWSJACK_SOURCE_DIR=/tmp/newsjack-source \
NEWSJACK_CLI_BINARY=/tmp/newsjack \
NEWSJACK_INSTALL_MCP=0 \
NEWSJACK_FORCE=1 \
/tmp/newsjack-source/install.sh

hash -r
cd /tmp
test ! -e .mcp.json
test ! -e "$HOME/.newsjack/credentials.json"

log "running setup from clean cwd with key supplied on stdin"
printf '%s\n\n%s\n' "$runtime" "$MEDIALYST_API_KEY" | \
  env -u MEDIALYST_API_KEY \
  newsjack setup --no-launch

log "checking saved credentials without environment key"
env -u MEDIALYST_API_KEY newsjack auth status > /tmp/newsjack-auth.json
jq -e '.medialyst_configured == true and (.medialyst_source | startswith("credentials:"))' /tmp/newsjack-auth.json >/dev/null

log "checking runtime MCP registration"
verify_runtime_mcp

log "running live Medialyst news_search without environment key"
env -u MEDIALYST_API_KEY \
  newsjack detector run "artificial intelligence" \
    --sources news_search \
    --limit 1 \
    --max-age-hours 0 \
  > /tmp/newsjack-live-search.json
jq -e '
  .diagnostics.source_status.news_search.status == "used" and
  (.diagnostics.evidence_by_source.news_search // 0) > 0
' /tmp/newsjack-live-search.json >/dev/null

log "clean live Medialyst/MCP check passed"
EOF
done

log "all clean live Medialyst/MCP checks passed"
