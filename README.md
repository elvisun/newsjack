# newsjack

> **Open-source operating system for agentic PR.**
>
> The skill layer that turns Claude, ChatGPT, and Cursor into PR operators.

```bash
curl -fsSL newsjack.sh | bash
newsjack setup
```

`newsjack setup` is interactive: choose where to install skills, choose the
agent harness that owns scheduled runs, optionally save X and Medialyst API
keys, then launch the selected harness with the `newsjack-setup` prompt.

The install path tracks the latest GitHub Release. `newsjack.sh` serves the installer to curl/wget clients and redirects browser traffic to this repository.

The installer:
- downloads the latest GitHub Release artifact for the current OS/arch
- verifies the artifact checksum
- installs a managed copy at `~/.newsjack/newsjack`
- installs `newsjack` at `~/.newsjack/bin/newsjack`
- generates instruction-only skill folders into detected runtimes
- keeps executable Newsjack code in `~/.newsjack/bin/newsjack` and the managed bundle under `~/.newsjack/newsjack`
- configures the optional `medialyst` MCP server through the Go CLI when the runtime exposes a noninteractive setup path

The curl installer uses prebuilt binaries for macOS and Linux on arm64/amd64. Users do not need Go installed.

Installed binaries auto-update from the latest GitHub Release before normal user-facing commands. If the hosted release version differs from `~/.newsjack/newsjack/VERSION`, Newsjack re-runs the installer and then re-runs the original command on the new binary. Set `NEWSJACK_AUTO_UPDATE=0` to disable this for CI or debugging.

Runtime detection is additive. If a user has multiple supported runtimes, the installer configures all of them.

Supported targets:
- **Codex:** skills to `~/.agents/skills`, MCP via `codex mcp add`
- **Claude Code:** skills to `~/.claude/skills`, MCP via `claude mcp add-json --scope user`
- **OpenClaw:** skills to `~/.openclaw/skills`, MCP via `openclaw mcp set`
- **Hermes:** skills to `~/.hermes/skills`, MCP by adding `mcp_servers.medialyst` to `~/.hermes/config.yaml`

Override auto-detection when needed:

```bash
curl -fsSL newsjack.sh | NEWSJACK_RUNTIMES=codex,claude bash
curl -fsSL newsjack.sh | NEWSJACK_RUNTIMES=all bash
curl -fsSL newsjack.sh | NEWSJACK_INSTALL_MCP=0 bash
curl -fsSL newsjack.sh | NEWSJACK_VERSION=v0.1.0 bash
curl -fsSL newsjack.sh | NEWSJACK_INSTALL_SKILLS=0 bash
NEWSJACK_AUTO_UPDATE=0 newsjack doctor
```

See [Distribution Roadmap](./docs/2026-05-24-distribution-roadmap.md) for the curl-v1 and npm-later plan. Release steps live in the [Release Playbook](./docs/release-playbook.md).

## What is this?

newsjack is the open-source operating system for **agentic PR** — a skill layer that installs into Claude, ChatGPT, Cursor, or any agent runtime that supports skills. Install it once and your agent becomes a PR operator: briefing, newsjacking, pitch critique, angle generation, media list workflow.

Local skills:
- **newsjack-setup** — create a company monitor profile and choose RSS feeds
- **newsjack-detector** — find newsworthy moments before the wave breaks
- **story-origin-check** — verify first-public freshness and canonical major coverage for a signal
- **newsworthiness-check** — score whether an event or pitch is actually newsworthy
- **meanest-editor** — roast your pitch against the rubric pros use
- **angle-generator** — turn one update into ten pitchable hooks
- **media-list-manager** — build and manage small, fit-checked media lists
- _(more skills shipping in v1)_

Optional cloud substrate via `newsjack login` or an MCP-compatible runtime unlocks Medialyst-backed media list generation.

## Agent Harness Setup

All Newsjack skills run as local skill files. Medialyst-backed media list generation is optional: when an MCP client is configured, `media-list-manager` can create, inspect, enrich, and share real Medialyst lists; without MCP it returns a local artifact for review or later import.

### Medialyst Login

Use the smallest API-key scopes needed for media-list workflows:

```text
news:search
media_lists:read
media_lists:write
```

Save the key once:

```bash
newsjack login
```

The helper stores credentials in `~/.newsjack/credentials.json` with user-only file permissions. Repo `.env` and `MEDIALYST_API_KEY` still work for CI and advanced local setups.

### Claude Code

Claude Code supports project MCP config and `headersHelper`, so it can use the repo `.mcp.json` directly:

```bash
newsjack login
claude --strict-mcp-config --mcp-config .mcp.json
```

Smoke test:

```bash
claude mcp list
```

Expected result: `medialyst` is connected.

### Codex

Codex supports MCP servers, but not Claude Code's `headersHelper`. Use the stdio bridge, which reads the same saved Newsjack credential and runs `mcp-remote` under the hood:

```bash
newsjack login
codex mcp add medialyst -- ~/.newsjack/bin/newsjack mcp-bridge
codex mcp list
```

For one-off advanced sessions, Codex can also use its native bearer-token env-var config:

```bash
codex mcp add medialyst --url https://medialyst.ai/api/mcp --bearer-token-env-var MEDIALYST_API_KEY
```

That native path requires `MEDIALYST_API_KEY` to be present in the Codex process environment.

### OpenClaw

OpenClaw supports configured MCP stdio servers. Use the same stdio bridge:

```bash
newsjack login
openclaw mcp set medialyst "{\"command\":\"$HOME/.newsjack/bin/newsjack\",\"args\":[\"mcp-bridge\"]}"
openclaw mcp list
```

The bridge requires Node.js because it launches `npx -y mcp-remote`.

### Other MCP Clients

If your client supports Claude-style `headersHelper`, point it at:

```bash
~/.newsjack/bin/newsjack auth headers
```

If your client supports only stdio MCP servers, configure:

```json
{
  "command": "/Users/alice/.newsjack/bin/newsjack",
  "args": ["mcp-bridge"]
}
```

If your client supports direct HTTP MCP with environment interpolation, configure:

```json
{
  "type": "http",
  "url": "https://medialyst.ai/api/mcp",
  "headers": {
    "Authorization": "Bearer ${MEDIALYST_API_KEY}"
  }
}
```

### Skills-Only Runtimes

For ChatGPT, Claude web, Cursor without MCP enabled, or any runtime that only supports local instructions, load the relevant `skills/*/SKILL.md` files. `media-list-manager` will operate in `local_artifact` mode and will not claim to create a live Medialyst list.

## Install

```bash
curl -fsSL newsjack.sh | bash
```

To review the install script before executing it:

```bash
curl -fsSL https://newsjack.sh
```

Alternatively, install via npm:

```bash
npm install -g newsjack
```

## Repo layout

```
apps/
  site/      # marketing site + edge function (UA-sniffs curl vs browser)
  cli/       # the newsjack CLI (install target, skill loader, login)
skills/      # the OSS skill files
.mcp.json    # optional Medialyst MCP project config
install.sh   # what `curl -fsSL newsjack.sh` serves to curl/wget clients
```

## How `curl -fsSL newsjack.sh | bash` works

The public `newsjack.sh` endpoint and the install script are served from the same URL. The Next.js `proxy.ts` in `apps/site/` inspects the `User-Agent` header: curl and wget are rewritten to `/install.sh`, while browsers are redirected to the GitHub repository. The root `install.sh` downloads versioned artifacts from GitHub Releases.

## License

MIT. See [LICENSE](./LICENSE).

## Credits

Built by [@elvissun](https://x.com/elvissun). Powered by [Medialyst](https://medialyst.com).
