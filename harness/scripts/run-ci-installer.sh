#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
IMAGE="${NEWSJACK_HARNESS_IMAGE:-newsjack-agent-harness:local}"
RUNTIMES="${NEWSJACK_RUNTIMES:-all}"
SOURCE_MODE="local"
INSTALLER_URL="${NEWSJACK_INSTALLER_URL:-https://newsjack.sh}"
ENV_FILE="${NEWSJACK_HARNESS_ENV_FILE:-}"

usage() {
  cat <<'USAGE'
Usage: harness/scripts/run-ci-installer.sh [options]

Options:
  --runtime <runtime>      Runtime target: all, codex, claude, openclaw, hermes
  --image <image>          Harness container image (default: newsjack-agent-harness:local)
  --local-source           Build and install from the checked-out source tree (default)
  --production-path        Install through the hosted newsjack.sh path
  --installer-url <url>    Hosted installer URL for --production-path
  --with-local-env         Load harness/.env.local into the container
  --env-file <path>        Load an explicit Docker env file into the container
USAGE
}

log() {
  printf '%s\n' "ci-installer: $*" >&2
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

while [ "$#" -gt 0 ]; do
  case "$1" in
    --runtime)
      [ "$#" -ge 2 ] || { usage >&2; exit 2; }
      RUNTIMES="$2"
      shift 2
      ;;
    --image)
      [ "$#" -ge 2 ] || { usage >&2; exit 2; }
      IMAGE="$2"
      shift 2
      ;;
    --local-source)
      SOURCE_MODE="local"
      shift
      ;;
    --production-path)
      SOURCE_MODE="production"
      shift
      ;;
    --installer-url)
      [ "$#" -ge 2 ] || { usage >&2; exit 2; }
      INSTALLER_URL="$2"
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

if ! command -v docker >/dev/null 2>&1; then
  log "docker is required to run the installer harness"
  exit 1
fi

log "running ${SOURCE_MODE} installer smoke in ${IMAGE} for runtimes: ${RUNTIMES}"

docker_args=(
  --rm
  --interactive
  --mount "type=bind,src=${REPO_DIR},dst=/repo,readonly"
  --workdir /repo
)

if [ -n "$ENV_FILE" ]; then
  validate_env_file "$ENV_FILE"
  docker_args+=(--env-file "$ENV_FILE")
  log "loading env file into container: ${ENV_FILE#$REPO_DIR/}"
else
  docker_args+=(
    --env "MEDIALYST_API_KEY="
    --env "X_BEARER_TOKEN="
    --env "TWITTER_BEARER_TOKEN="
    --env "X_API_BEARER_TOKEN="
    --env "TWITTER_API_BEARER_TOKEN="
  )
fi

docker_args+=(
  --env "HOME=/tmp/newsjack-home"
  --env "XDG_CONFIG_HOME=/tmp/newsjack-home/.config"
  --env "XDG_CACHE_HOME=/tmp/newsjack-home/.cache"
  --env "XDG_DATA_HOME=/tmp/newsjack-home/.local/share"
  --env "NEWSJACK_RUNTIMES=${RUNTIMES}"
  --env "NEWSJACK_NO_AUTO_UPDATE=1"
  --env "NEWSJACK_HARNESS_SOURCE_MODE=${SOURCE_MODE}"
  --env "NEWSJACK_INSTALLER_URL=${INSTALLER_URL}"
)

docker run "${docker_args[@]}" "${IMAGE}" bash -s <<'EOF'
set -euo pipefail

log() {
  printf '%s\n' "ci-installer(container): $*" >&2
}

selected_skill_dirs() {
  local raw="${NEWSJACK_RUNTIMES:-all}"
  raw="${raw//claude-code/claude}"
  raw="${raw//claude_code/claude}"
  raw="${raw//cladue/claude}"
  raw="${raw//,/ }"

  local dirs=()
  local rt
  for rt in $raw; do
    case "$rt" in
      all|auto)
        printf '%s\n' \
          "$HOME/.agents/skills" \
          "$HOME/.claude/skills" \
          "$HOME/.openclaw/skills" \
          "$HOME/.hermes/skills"
        return
        ;;
      none)
        return
        ;;
      codex)
        dirs+=("$HOME/.agents/skills")
        ;;
      claude)
        dirs+=("$HOME/.claude/skills")
        ;;
      openclaw)
        dirs+=("$HOME/.openclaw/skills")
        ;;
      hermes)
        dirs+=("$HOME/.hermes/skills")
        ;;
      "")
        ;;
      *)
        log "unsupported runtime selection: $rt"
        exit 1
        ;;
    esac
  done

  printf '%s\n' "${dirs[@]}"
}

export PATH="$HOME/.newsjack/bin:$PATH"
export NEWSJACK_NO_AUTO_UPDATE=1

case "${NEWSJACK_HARNESS_SOURCE_MODE:-local}" in
  local)
    log "building and testing local CLI"
    (cd /repo/apps/cli && go test ./... && go build -buildvcs=false -o /tmp/newsjack ./cmd/newsjack)

    log "running local-source installer"
    NEWSJACK_SOURCE_DIR=/repo \
    NEWSJACK_CLI_BINARY=/tmp/newsjack \
    NEWSJACK_INSTALL_MCP=1 \
    NEWSJACK_FORCE=1 \
    /repo/install.sh
    ;;
  production)
    log "running hosted installer: ${NEWSJACK_INSTALLER_URL}"
    curl -fsSL "${NEWSJACK_INSTALLER_URL}" | \
      NEWSJACK_RUNTIMES="${NEWSJACK_RUNTIMES}" \
      NEWSJACK_INSTALL_MCP=1 \
      NEWSJACK_FORCE=1 \
      sh
    ;;
  *)
    log "unsupported source mode: ${NEWSJACK_HARNESS_SOURCE_MODE}"
    exit 1
    ;;
esac

cd /tmp

log "checking installed CLI"
newsjack version
newsjack setup --json | tee /tmp/newsjack-setup.json
jq -e '.monitors_dir and .agent_prompt and .recommended_runtime and .recommended_scheduler and .agent_command' /tmp/newsjack-setup.json >/dev/null
newsjack doctor --json | tee /tmp/newsjack-doctor.json
jq -e '.root_ok == true and .mcp_bridge.transport == "native"' /tmp/newsjack-doctor.json >/dev/null

log "running native MCP bridge smoke against mock server"
(cd /repo/harness/mock-mcp && exec go run . --addr 127.0.0.1:8970 --key mock-key) &
mock_mcp_pid=$!
trap 'kill "$mock_mcp_pid" 2>/dev/null || true' EXIT
for _ in $(seq 1 50); do
  status="$(curl -s -o /dev/null -w '%{http_code}' http://127.0.0.1:8970/mcp || true)"
  if [ -n "$status" ] && [ "$status" != "000" ]; then
    break
  fi
  sleep 0.2
done

# Restrict PATH so the smoke proves the bridge needs no Node/npx on the
# machine; only the installed CLI and core utilities stay visible.
printf '%s\n' \
  '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-03-26","capabilities":{},"clientInfo":{"name":"harness-smoke","version":"0"}}}' \
  '{"jsonrpc":"2.0","method":"notifications/initialized"}' \
  '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | \
  env PATH="$HOME/.newsjack/bin:/usr/bin:/bin" \
    MEDIALYST_API_KEY=mock-key \
    NEWSJACK_MEDIALYST_MCP_URL=http://127.0.0.1:8970/mcp \
    "$HOME/.newsjack/bin/newsjack" mcp-bridge > /tmp/newsjack-bridge.out
grep -q '"protocolVersion"' /tmp/newsjack-bridge.out
grep -q 'mock_search' /tmp/newsjack-bridge.out
kill "$mock_mcp_pid" 2>/dev/null || true

log "checking installed runtime skills"
mapfile -t skill_dirs < <(selected_skill_dirs)
for dir in "${skill_dirs[@]}"; do
  test -f "$dir/newsjack-detector/SKILL.md"
  test -f "$dir/newsjack-monitor-setup/SKILL.md"
  test ! -d "$dir/newsjack-detector/scripts"
done

log "running mock detector smoke"
newsjack detector run "AI search visibility" --mock --limit 1 \
  | tee /tmp/newsjack-detector.json \
  | jq -e '.monitor.mock == true and (.monitor.queries | index("AI search visibility")) and (.signals | length) >= 1' >/dev/null

log "running monitor lifecycle smoke"
cat >/tmp/newsjack-profile.json <<'JSON'
{
  "company": "Harness Coffee",
  "description": "Specialty coffee company used for installer verification.",
  "topics": ["coffee supply chain"],
  "search_terms": ["coffee supply chain"],
  "feed_urls": ["https://example.com/feed.xml"],
  "x_news": {"enabled": true},
  "x_trends": {"mode": "none", "woeids": [], "locations": []},
  "standing": ["coffee sourcing"],
  "proof_assets": ["sourcing data"]
}
JSON

newsjack monitor init harness-coffee --profile /tmp/newsjack-profile.json | tee /tmp/newsjack-monitor-init.json
jq -e '.slug == "harness-coffee" and .profile_path' /tmp/newsjack-monitor-init.json >/dev/null

newsjack monitor test harness-coffee --mock --limit 2 | tee /tmp/newsjack-monitor-test.json
jq -e '.candidates and .summary and .report_target' /tmp/newsjack-monitor-test.json >/dev/null
test -f "$(jq -r '.candidates' /tmp/newsjack-monitor-test.json)"
test -f "$(jq -r '.summary' /tmp/newsjack-monitor-test.json)"
test ! -f "$(jq -r '.report_target' /tmp/newsjack-monitor-test.json)"

newsjack monitor schedule harness-coffee --runtime claude --every 1h | tee /tmp/newsjack-monitor-schedule.json
jq -e '.system_cron == false and .runtime == "claude" and .schedule_path and (.suggested_minute >= 1) and (.suggested_minute <= 59)' /tmp/newsjack-monitor-schedule.json >/dev/null
schedule_md="$(jq -r '.schedule_path' /tmp/newsjack-monitor-schedule.json)"
test -f "$schedule_md"
! grep -Eq 'crontab|launchd|systemd' "$schedule_md"
grep -q 'never minute 0' "$schedule_md"

newsjack monitor status harness-coffee | tee /tmp/newsjack-monitor-status.json
jq -e '.exists == true and .run_count == 1 and (.latest_report_path == null)' /tmp/newsjack-monitor-status.json >/dev/null

log "installer smoke complete"
EOF
