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

The site records privacy-limited request events for `newsjack.sh`:

- `site_visit` when browser-style traffic visits `/` and is redirected to GitHub
- `install_request` when curl, wget, HTTPie, or another installer-style user
  agent requests `/` and receives `install.sh`

The code accepts these Vercel/Neon connection variables:

| Variable | Purpose |
| --- | --- |
| `NEWSJACK_DATABASE_URL` | Newsjack-specific override when set. |
| `DATABASE_URL` | Standard Neon/Vercel connection string. |
| `POSTGRES_URL` | Standard Vercel Postgres connection string. |
| `POSTGRES_PRISMA_URL` | Alternate pooled Vercel Postgres connection string. |

`NEWSJACK_IP_HASH_SALT` is optional. If it is missing, the server-only database
URL is used as the hash secret for same-day IP dedupe. Raw IP addresses are not
stored.

Run the migration before enabling telemetry:

```bash
psql "$DATABASE_URL" -f apps/site/db/migrations/0001_install_events.sql
```

Local smoke test:

```bash
cd apps/site
pnpm dev
curl -A "curl/8.0" -i "http://localhost:3000/?utm_source=local"
```

If the database env vars are set and the migration has run, an
`install_request` row should appear in `install_events`. Browser-style requests
to `/` should create `site_visit` rows.

Query last-24h counts from the repo root:

```bash
NEWSJACK_ENV_FILE="/Users/elvissun/Documents/GitHub/newsjack-worktrees/newsjack-main/.env" \
  node scripts/funnel-stats.mjs
```

For details on what is collected, see [docs/telemetry.md](docs/telemetry.md).
