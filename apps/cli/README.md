# Newsjack CLI

The Go CLI is a thin client for Medialyst news search, PR calendar lookup, journalist enrichment, and asynchronous media-list research plus local install, auth, update, skills, and monitor plumbing. It does not run or configure an MCP server, and it does not expose the hosted spreadsheet-management surface.

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

Default OAuth scopes are `news:search media_lists:manage projects:manage`. Older compatible servers that reject `projects:manage` with `invalid_scope` are retried once with the two legacy scopes. Existing users must run `newsjack login` again to grant access to Medialyst project tools; `newsjack auth status` reports the scopes on the stored grant.

## REST Command Mapping

Commands preserve the public Newsjack UX where possible and forward requests to Medialyst without local schema re-derivation. Use `--json` or `--json-file` when you need an exact request body. Agents own campaign judgment, user approval, incremental progress reporting, and the final fit-checked send wave.

| CLI command | API endpoint |
| --- | --- |
| `newsjack credits balance` | `GET /api/v1/credits/balance` |
| `newsjack news search` | `POST /api/v1/news/search` |
| `newsjack pr-calendar query` | `POST /api/v1/pr-calendar/query` |
| `newsjack journalists enrich` | `POST /api/v1/journalists/enrich` |
| `newsjack journalists enrich-job <job-id>` | `GET /api/v1/journalist-enrichment-jobs/{jobId}` |
| `newsjack media-lists create` | `POST /api/v1/media-lists:create-async` |
| `newsjack media-lists job <job-id>` | `GET /api/v1/jobs/{jobId}` |

The enrichment command uses the polished public enrich endpoint from Medialyst PR1024. Article URL sources are the supported path today:

```bash
newsjack journalists enrich \
  --url https://example.com/story \
  --pitch "why this journalist set fits" \
  --wait
```

[Media-list research](https://medialyst.ai/docs/developers) is asynchronous and credit-bearing. Agents must get explicit user approval for the exact campaign prompt and calculated research target before creation, then poll the returned job and surface normalized rows as they become ready:

```bash
newsjack media-lists create \
  --prompt "Find Canadian journalists covering PR technology and AI media tools" \
  --target-list-size 50 \
  --idempotency-key campaign-2026-09-14

newsjack media-lists job <job-id> --include-results --limit 50
```

The user's desired count means good fits; it is not normally the API target. `find-journalists` recommends a research target of 5x the desired count, increasing toward 10x for unusually constrained briefs. It shows that calculated target and maximum credit exposure for approval before creating the job. The CLI itself sends the exact supplied value and never applies a hidden multiplier.

The target size sets the requested research size and credit budget. A discovery or beat sweep can temporarily expose more candidate rows than that target, and neither the target nor `ready_rows` guarantees the same number of complete, unique journalist profiles.

The CLI intentionally exposes no media-list list/get/inspect/action/share/delete commands. Those are app workflows, not agent mechanics.

The public target range is 1–1,000. Some organizations enable multi-angle planning, which can require a practical minimum of 3 so every angle retains a retrieval lane; the API reports a retryable `PROMPT_ANGLE_ARTICLE_LIMIT_TOO_SMALL` error when that applies.

## Removed MCP Surface

The CLI no longer includes `newsjack mcp`, `newsjack mcp-bridge`, MCP transport code, runtime MCP setup, or installer MCP flags. Git history is the fallback for the old implementation.

Breaking changes:

- `newsjack mcp` and `newsjack mcp-bridge` are gone.
- The old spreadsheet-oriented `newsjack media-lists` CRUD/action surface remains removed. Only asynchronous `create` and incremental `job` reads are exposed for agent-driven list creation.
- Installer flags and state for `NEWSJACK_INSTALL_MCP`, `--mcp`, and `install_mcp` are gone.
- Agent skills should call `newsjack` REST commands directly or fall back to local mode.

## Tests

```bash
cd apps/cli
go test ./...
```
