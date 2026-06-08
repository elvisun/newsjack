# newsjack.sh site

Next.js site and installer host for newsjack.sh. Browser traffic redirects to
GitHub; installer user agents on `/` receive the bundled shell installer.

## Requirements

- Node.js 22+
- pnpm 10+

## Development

```bash
pnpm install
pnpm dev
```

The dev server runs at http://localhost:3000.

## Build

```bash
pnpm build
```

## Checks

```bash
pnpm lint
pnpm test
```

## Telemetry

The installer funnel records a privacy-limited `curl_hit` event when curl, wget,
HTTPie, or another installer-style user agent requests `/`.

Set these in the Vercel project environment:

| Variable | Purpose |
| --- | --- |
| `NEWSJACK_DATABASE_URL` | Cloud Postgres connection string. Default to Neon unless an existing Newsjack Postgres database already exists. |
| `NEWSJACK_IP_HASH_SALT` | Secret salt for the daily SHA-256 IP hash. The code also includes the UTC date so hashes are only useful for same-day dedupe. Rotate the secret periodically. |
| `NEWSJACK_INSTALL_EVENT_SECRET` | Shared secret for `POST /api/install-event`, sent as `Authorization: Bearer ...` or `x-newsjack-install-event-secret`. |

Run the migration before enabling telemetry:

```bash
psql "$NEWSJACK_DATABASE_URL" -f apps/site/db/migrations/0001_install_events.sql
```

Local smoke test:

```bash
cd apps/site
pnpm dev
curl -A "curl/8.0" -i "http://localhost:3000/?utm_source=local"
```

The response should include `X-Newsjack-Install-Id`. If the database env vars
are set and the migration has run, a `curl_hit` row should appear in
`install_events`.

Query last-24h funnel counts from the repo root:

```bash
NEWSJACK_DATABASE_URL="..." node scripts/funnel-stats.mjs
```

For details on what is collected, retention, and the stage 2/3 callback
scaffolding, see [docs/telemetry.md](docs/telemetry.md).
