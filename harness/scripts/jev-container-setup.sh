#!/usr/bin/env bash
# Runs inside the harness container. Builds the CLI from the mounted checkout,
# installs it into an isolated home, saves the TypeSafe key, and prints a
# cheat sheet for testing the Jev coarse-filter path. Auto-update is disabled
# so the hosted main channel cannot replace the branch binary.
set -euo pipefail

export HOME="${HOME:-/tmp/newsjack-home}"
export NEWSJACK_AUTO_UPDATE=0
export PATH="$HOME/.newsjack/bin:$PATH"
rm -rf "$HOME"
mkdir -p "$HOME"

echo "== building newsjack from /repo"
mkdir -p /tmp/newsjack-build
(cd /repo/apps/cli && CGO_ENABLED=0 go build -trimpath -buildvcs=false -o /tmp/newsjack-build/newsjack ./cmd/newsjack)

echo "== installing into $HOME (runtime: ${NEWSJACK_RUNTIMES:-claude})"
(cd /repo && NEWSJACK_SOURCE_DIR=/repo \
  NEWSJACK_CLI_BINARY=/tmp/newsjack-build/newsjack \
  NEWSJACK_RUNTIMES="${NEWSJACK_RUNTIMES:-claude}" \
  NEWSJACK_INSTALL_MCP="${NEWSJACK_INSTALL_MCP:-0}" \
  NEWSJACK_RUN_SETUP=0 \
  sh ./install.sh >/tmp/newsjack-install.log 2>&1) || { cat /tmp/newsjack-install.log; exit 1; }
hash -r

if [ -n "${TYPESAFE_API_KEY:-}" ]; then
  newsjack auth set-typesafe --key "$TYPESAFE_API_KEY"
  unset TYPESAFE_API_KEY
else
  echo "!! TYPESAFE_API_KEY not in the container env; Jev calls will fail until you run: newsjack auth set-typesafe --key <key>"
fi
if [ -n "${MEDIALYST_API_KEY:-}" ]; then
  newsjack auth set-medialyst --key "$MEDIALYST_API_KEY" >/dev/null
  unset MEDIALYST_API_KEY
fi

echo
newsjack version
newsjack doctor | sed -n '/AUTH/,/SOURCES/p'

R=/repo/fixtures/newsjack-detector-agent/runs
cat <<EOF

== Jev coarse-filter test cheat sheet (binary: $(command -v newsjack))

unit tests (fake TypeSafe server, no network):
  (cd /repo/apps/cli && go test ./... -run 'Coarse|TypeSafe' -v)

what Jev is asked, and what it would cost, no calls:
  newsjack coarse-filter --print-questions | jq .
  newsjack coarse-filter --candidates $R/20260603T175505Z_simular/coarse_chunk.1.json --dry-run | jq 'del(.sample_request)'

one live chunk (18 signals, ~\$0.001):
  newsjack coarse-filter --candidates $R/20260603T175505Z_simular/coarse_chunk.1.json --output /tmp/jev.json
  jq '.engine' /tmp/jev.json
  jq -r '.decisions[] | "\(.decision)\t\(.reason)\t\(.confidence)\t\(.post_rules|join(","))\t\(.rationale)"' /tmp/jev.json

full contract through filter-apply (75 signals, ~\$0.005):
  newsjack coarse-filter --candidates $R/20260603T175505Z_simular/candidates.json --output /tmp/coarse.json
  newsjack filter-apply --candidates $R/20260603T175505Z_simular/candidates.json --decisions /tmp/coarse.json --include keep --include monitor_only --output /tmp/relevant.json
  jq '.coarse_relevance | del(.rejected_signals, .missing_signals)' /tmp/relevant.json

agreement eval against the Haiku worker (176 signals, ~\$0.013):
  NEWSJACK_BIN=newsjack /repo/eval/jev-coarse-agreement/run.sh $R/20260603T175505Z_simular $R/20260603T175505Z_slite $R/20260603T175505Z_localfalcon
  python3 /repo/eval/jev-coarse-agreement/compare.py /repo/eval/jev-coarse-agreement/runs/<timestamp>

tune wording without a rebuild:
  newsjack coarse-filter --print-questions > /tmp/q.json   # edit, then:
  newsjack coarse-filter --candidates $R/20260603T175505Z_simular/coarse_chunk.1.json --questions /tmp/q.json --output /tmp/jev2.json

full detector run through the skill (needs ANTHROPIC_API_KEY; Claude Code is installed):
  cd /repo && claude
  then: /newsjack-detector on fixtures/newsjack-detector-agent/profile.simular.json in mock mode; check the run's coarse_relevance_decisions.json has an "engine" block and the summary names jev.

EOF
