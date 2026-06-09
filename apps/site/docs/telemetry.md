# Traffic Telemetry

Newsjack logs privacy-limited request events for `newsjack.sh`. The goal is to
answer two launch questions:

- how many people visited the URL
- what kind of users or agents requested the installer

## Events

| Event | Meaning |
| --- | --- |
| `site_visit` | A browser, bot, agent, or unknown client requested `/` and was redirected to GitHub. |
| `install_request` | curl, wget, HTTPie, or another installer-style client requested `/` and was served `install.sh`. |

There are no install-started, install-completed, or install-failed callbacks.
The shell installer does not send telemetry.

## What Is Collected

| Field | Notes |
| --- | --- |
| `event_type` | `site_visit` or `install_request`. |
| `created_at` | Server timestamp. |
| `ip_hash` | SHA-256 of IP plus a server-only salt plus the UTC date. Raw IPs are not stored. |
| `country`, `region` | Vercel geolocation headers. |
| `user_agent` | Raw user agent. |
| `client_kind` | Parsed user-agent family, such as `browser`, `curl`, `wget`, `codex`, `claude`, or `bot`. |
| `referer`, `accept_language` | Request headers, when present. |
| `query_params` | Full query string, including UTMs and other tracking params. |
| `metadata` | Operational context such as path, method, host, and Vercel request id. |

The site does not store raw IP addresses. Query strings can carry personal data
if someone puts it there, so do not add email addresses or names to launch URLs.

## Storage

Use the Neon Postgres database connected through Vercel.

The code accepts any of these connection-string variables:

| Variable | Notes |
| --- | --- |
| `NEWSJACK_DATABASE_URL` | Newsjack-specific override when set. |
| `DATABASE_URL` | Standard Neon/Vercel connection string. |
| `POSTGRES_URL` | Standard Vercel Postgres connection string. |
| `POSTGRES_PRISMA_URL` | Alternate pooled Vercel Postgres connection string. |

`NEWSJACK_IP_HASH_SALT` is optional. If it is missing, the server-only database
URL is used as the hash secret. The UTC date is still included so hashes are only
useful for same-day dedupe.

Migration:

```bash
psql "$DATABASE_URL" -f apps/site/db/migrations/0001_install_events.sql
```

Stats from a local Vercel env file:

```bash
NEWSJACK_ENV_FILE="/Users/elvissun/Documents/GitHub/newsjack-worktrees/newsjack-main/.env" \
  node scripts/funnel-stats.mjs
```
