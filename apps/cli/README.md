# Newsjack CLI

The Go CLI is a thin client for the Medialyst public REST API plus local install, auth, update, skills, and monitor plumbing. It does not run or configure an MCP server.

## Auth

Run:

```bash
newsjack login
```

The CLI loads the API key from `MEDIALYST_API_KEY` or `~/.newsjack/credentials.json`. `NEWSJACK_MEDIALYST_API_BASE` or `MEDIALYST_API_BASE` can point commands at another compatible API base; the default is `https://medialyst.ai/api`.

Recommended scopes for cloud media-list workflows are `news:search` and `media_lists:manage`.

## REST Command Mapping

Commands preserve the public Newsjack UX where possible and forward requests to Medialyst without local schema re-derivation. Use `--json` or `--json-file` when you need an exact request body.

| CLI command | API endpoint |
| --- | --- |
| `newsjack credits balance` | `GET /api/v1/credits/balance` |
| `newsjack news search` | `POST /api/v1/news/search` |
| `newsjack journalists enrich` | `POST /api/v1/journalists/enrich` |
| `newsjack journalists enrich-job <job-id>` | `GET /api/v1/journalist-enrichment-jobs/{jobId}` |
| `newsjack media-lists create` | `POST /api/v1/media-lists` |
| `newsjack media-lists create-async` | `POST /api/v1/media-lists:create-async` |
| `newsjack media-lists job <job-id>` | `GET /api/v1/jobs/{jobId}` |
| `newsjack media-lists list` | `GET /api/v1/media-lists` |
| `newsjack media-lists get <id>` | `GET /api/v1/media-lists/{mediaListId}` |
| `newsjack media-lists inspect <id>` | `POST /api/v1/media-lists/{mediaListId}/inspect` |
| `newsjack media-lists full-values <id>` | `POST /api/v1/media-lists/{mediaListId}/full-values` |
| `newsjack media-lists preview-column-render <id>` | `POST /api/v1/media-lists/{mediaListId}/column-render-preview` |
| `newsjack media-lists action <id>` | `POST /api/v1/media-lists/{mediaListId}/actions` |
| `newsjack media-lists add-urls <id>` | `POST /api/v1/media-lists/{mediaListId}/actions` (`add_articles_by_urls`) |
| `newsjack media-lists add-keywords <id>` | `POST /api/v1/media-lists/{mediaListId}/actions` (`add_articles_by_keywords`) |
| `newsjack media-lists share <id>` | `POST /api/v1/media-lists/{mediaListId}/shares` |
| `newsjack media-lists delete <id>` | `DELETE /api/v1/media-lists/{mediaListId}` |

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
- Installer flags and state for `NEWSJACK_INSTALL_MCP`, `--mcp`, and `install_mcp` are gone.
- Agent skills should call `newsjack` REST commands directly or fall back to local mode.

## Tests

```bash
cd apps/cli
go test ./...
```
