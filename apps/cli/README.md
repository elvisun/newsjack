# Newsjack CLI

The Go CLI is a thin client for Medialyst news search and journalist enrichment plus local install, auth, update, skills, and monitor plumbing. It does not run or configure an MCP server, and it does not manage hosted media lists.

## Auth

For interactive users and agents, run:

```bash
newsjack login
```

`newsjack login` starts the Medialyst OAuth device flow for the public `newsjack-cli` client. The CLI prints a Medialyst approval link, opens it in the browser when possible, then stores OAuth tokens in `~/.newsjack/credentials.json`.

API keys still work for CI and power users:

```bash
newsjack auth set-medialyst --key <mlst_...>
```

The CLI prefers saved OAuth, then API keys from `~/.newsjack/credentials.json` or `MEDIALYST_API_KEY`. `NEWSJACK_MEDIALYST_API_BASE` or `MEDIALYST_API_BASE` can point commands at another compatible API base; the default is `https://medialyst.ai/api`.

Default OAuth scopes are `news:search media_lists:manage`.

## REST Command Mapping

Commands preserve the public Newsjack UX where possible and forward requests to Medialyst without local schema re-derivation. Use `--json` or `--json-file` when you need an exact request body. Agents own any local media-list artifacts or campaign organization.

| CLI command | API endpoint |
| --- | --- |
| `newsjack credits balance` | `GET /api/v1/credits/balance` |
| `newsjack news search` | `POST /api/v1/news/search` |
| `newsjack journalists enrich` | `POST /api/v1/journalists/enrich` |
| `newsjack journalists enrich-job <job-id>` | `GET /api/v1/journalist-enrichment-jobs/{jobId}` |

The enrichment command uses the polished public enrich endpoint from Medialyst PR1024. Article URL sources are the supported path today:

```bash
newsjack journalists enrich \
  --url https://example.com/story \
  --pitch "why this journalist set fits" \
  --wait
```

## Removed MCP Surface

The CLI no longer includes `newsjack mcp`, `newsjack mcp-bridge`, MCP transport code, runtime MCP setup, or installer MCP flags. Git history is the fallback for the old implementation.

Breaking changes:

- `newsjack mcp` and `newsjack mcp-bridge` are gone.
- `newsjack media-lists ...` is not a CLI surface; agents should organize enriched journalist data in their own working files or final answer.
- Installer flags and state for `NEWSJACK_INSTALL_MCP`, `--mcp`, and `install_mcp` are gone.
- Agent skills should call `newsjack` REST commands directly or fall back to local mode.

## Tests

```bash
cd apps/cli
go test ./...
```
