# Newsjack Agent Runtime Harness

This harness is for observing the real installer inside a disposable Linux
container. It intentionally does not contain a second product implementation.
The installer should always exercise the compiled `newsjack` binary and the
real runtime skill install paths.

## Build The Harness Image

From the repo root:

```bash
harness/scripts/build-image.sh --harness all
```

Preinstall only the harness CLIs needed for a scenario:

```bash
harness/scripts/build-image.sh --harness claude --image newsjack-agent-harness:claude
harness/scripts/build-image.sh --harness claude --harness openclaw --image newsjack-agent-harness:claude-openclaw
harness/scripts/build-image.sh --harness none --image newsjack-agent-harness:none
```

The equivalent raw Docker build arg is:

```bash
docker build \
  -f harness/Dockerfile \
  --build-arg NEWSJACK_HARNESS_RUNTIMES=claude,openclaw \
  -t newsjack-agent-harness:claude-openclaw \
  .
```

Open a shell with the repo mounted:

```bash
harness/scripts/docker-shell.sh --image newsjack-agent-harness:local
```

To load provider/API keys from the ignored local harness env file:

```bash
harness/scripts/docker-shell.sh \
  --image newsjack-agent-harness:local \
  --with-local-env
```

Inside the container, isolate runtime state:

```bash
export HOME=/tmp/newsjack-home
export XDG_CONFIG_HOME="$HOME/.config"
export XDG_CACHE_HOME="$HOME/.cache"
export XDG_DATA_HOME="$HOME/.local/share"
export PATH="$HOME/.newsjack/bin:$PATH"
rm -rf "$HOME"
```

`curl | sh` runs in a child shell, so it cannot update the parent shell's
`PATH`. Keep the `PATH` export above in the interactive shell, or run the
binary by absolute path:

```bash
"$HOME/.newsjack/bin/newsjack" version
```

## Local Source Path

Use this when iterating on `install.sh`, skills, or the CLI before deploying.
The public installer no longer builds Go on user machines, so local source
installs must provide a compiled binary explicitly.

Inside the container:

```bash
mkdir -p /tmp/newsjack-build
(cd apps/cli && CGO_ENABLED=0 go build -trimpath -buildvcs=false -o /tmp/newsjack-build/newsjack ./cmd/newsjack)

NEWSJACK_SOURCE_DIR=/repo \
NEWSJACK_CLI_BINARY=/tmp/newsjack-build/newsjack \
NEWSJACK_RUNTIMES=all \
NEWSJACK_INSTALL_MCP=1 \
sh ./install.sh

hash -r
command -v newsjack
file "$(command -v newsjack)"
newsjack version
newsjack setup
newsjack skills list
newsjack doctor | jq .
```

Expected:

- `file "$(command -v newsjack)"` reports an ELF executable, not a shell script.
- Skills land under the temp home runtime dirs, for example
  `$HOME/.agents/skills`, `$HOME/.claude/skills`, and `$HOME/.openclaw/skills`.
- MCP setup either configures detected runtimes or logs non-blocking warnings.

## Local Hosted-Dist Path

Use this to test the same shape as production before pushing: the site serves
`/install.sh` and `/dist`, and the container installs via HTTP.

For an interactive Codex-only install test where you manually run the
`curl | bash` command inside a clean container:

```bash
harness/scripts/open-codex-curl-shell.sh
```

The script builds a local release bundle from the current checkout, serves it on
localhost, and opens a container with Codex preinstalled. It prints the exact
installer command to run inside the container. In an interactive container, the
installer should hand off to `newsjack setup` automatically. Set
`NEWSJACK_RUN_SETUP=0` on the installer command for install-only debugging.

To test the first-run user experience with Codex installed but no repo `.env`,
no mounted source checkout, and no API-key environment variables:

```bash
harness/scripts/open-codex-curl-shell.sh --clean-no-env
```

On the host, from the repo root:

```bash
pnpm --dir apps/site run build
pnpm --dir apps/site exec next start --port 3010
```

In another terminal, start the harness container. On Docker Desktop for macOS,
the host is reachable as `host.docker.internal`:

```bash
docker run --rm -it \
  -v "$PWD:/repo" \
  -w /repo \
  newsjack-agent-harness:local \
  bash
```

Inside the container:

```bash
export HOME=/tmp/newsjack-home
export XDG_CONFIG_HOME="$HOME/.config"
export XDG_CACHE_HOME="$HOME/.cache"
export XDG_DATA_HOME="$HOME/.local/share"
export PATH="$HOME/.newsjack/bin:$PATH"
rm -rf "$HOME"

curl -fsSL http://host.docker.internal:3010 | \
  NEWSJACK_DIST_BASE=http://host.docker.internal:3010/dist \
  NEWSJACK_RUNTIMES=all \
  NEWSJACK_INSTALL_MCP=1 \
  sh

hash -r
file "$(command -v newsjack)"
newsjack version
newsjack setup
newsjack doctor | jq .
```

On Linux hosts, add Docker's host gateway mapping when starting the container:

```bash
docker run --rm -it \
  --add-host=host.docker.internal:host-gateway \
  -v "$PWD:/repo" \
  -w /repo \
  newsjack-agent-harness:local \
  bash
```

## Production Path

Use this after a push/deploy to verify the live domain end to end:

```bash
export HOME=/tmp/newsjack-home
export XDG_CONFIG_HOME="$HOME/.config"
export XDG_CACHE_HOME="$HOME/.cache"
export XDG_DATA_HOME="$HOME/.local/share"
export PATH="$HOME/.newsjack/bin:$PATH"
rm -rf "$HOME"

curl -fsSL newsjack.sh | \
  NEWSJACK_RUNTIMES=all \
  NEWSJACK_INSTALL_MCP=1 \
  sh

hash -r
command -v newsjack
file "$(command -v newsjack)"
newsjack version
newsjack setup
newsjack skills list
newsjack doctor | jq .
```

To inspect the deployed channel:

```bash
curl -fsSL https://newsjack.sh/dist/channels/main.txt
curl -fsSL https://newsjack.sh/dist/manifest.json | jq .
```

Scripted production smoke with local credentials:

```bash
harness/scripts/run-ci-installer.sh \
  --image newsjack-agent-harness:all \
  --runtime all \
  --production-path \
  --with-local-env
```

`--with-local-env` loads `harness/.env.local` into the container with Docker's
`--env-file`. The script validates that repo-local env files are ignored by git
and never prints secret values. CI should keep using the default no-token path.

## Auto-Update Observation

Installed binaries auto-update from the hosted `main` channel before normal
user-facing commands. To force that path in the container:

```bash
printf 'stale-version\n' > "$HOME/.newsjack/newsjack/VERSION"
newsjack doctor > /tmp/newsjack-doctor.json 2> /tmp/newsjack-update.log

cat /tmp/newsjack-update.log
jq . /tmp/newsjack-doctor.json
cat "$HOME/.newsjack/newsjack/VERSION"
```

Expected:

- stderr shows `[info] auto-updating ...`.
- stdout remains valid JSON for `doctor`.
- `VERSION` is rewritten to the live channel commit.

Disable auto-update for deterministic debugging:

```bash
NEWSJACK_AUTO_UPDATE=0 newsjack doctor | jq .
```

## Notes

- Put installer environment variables on the `sh` side of the pipe:
  `curl -fsSL newsjack.sh | NEWSJACK_RUNTIMES=all sh`.
- Do not use `NEWSJACK_RUNTIMES=all curl ... | sh`; that only sets the
  variable for `curl`, not for the installer shell.
- The old `harness/run.py` workflow is gone. If a scripted harness returns,
  keep it as a thin shell or Go wrapper around these commands and the compiled
  `newsjack` binary.
