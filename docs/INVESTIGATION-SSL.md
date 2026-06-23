# newsjack.sh SSL Investigation

Verified on June 23, 2026 from Toronto/Canada network vantage points.

## Summary

The live Vercel deployment is healthy for modern clients. `www.newsjack.sh`
serves a valid Let's Encrypt certificate and accepts TLS 1.2 and TLS 1.3.
Clients limited to TLS 1.0 or 1.1 fail during the TLS handshake with a protocol
version alert, which can surface in Chrome as `ERR_SSL_PROTOCOL_ERROR`.

The biggest site-side risk was Vercel's default 2-year HSTS header. This PR
adds a project-level override for app-served responses:

```http
Strict-Transport-Security: max-age=300
```

It deliberately does not add `includeSubDomains` or `preload`.

## Repo Configuration Audit

Domain and host references found in this repo:

| File | Purpose | Finding |
| --- | --- | --- |
| `apps/site/next.config.ts` | Next.js project config | Now sets `Strict-Transport-Security: max-age=300` for all app paths. |
| `apps/site/proxy.ts` | Root route behavior | Installer-style user agents on `/` are rewritten to `/install.sh`; browser-style user agents are redirected to GitHub with 308. No apex/www redirect logic lives here. |
| `apps/site/app/install.sh/route.ts` | Installer response | Serves the bundled shell installer and now sets the same HSTS header. |
| `apps/site/app/layout.tsx` | Metadata | Uses `https://newsjack.sh` as `metadataBase` and Open Graph URL. |
| `apps/cli/cmd/newsjack/config.go` | CLI default install URL | Uses `https://newsjack.sh`. |
| `install.sh` | Installer default URL | Uses `https://newsjack.sh`. |
| `README.md`, `docs/*`, `harness/*` | Documentation and smoke paths | Mostly use `newsjack.sh` as the short install URL. |

No `vercel.json`, Wrangler config, Cloudflare config, Terraform, or DNS-as-code
was found in this checkout.

## Live DNS

Observed DNS:

| Host | Records |
| --- | --- |
| `newsjack.sh` | Vercel-managed IPv4 A records. Local resolver returned `216.150.1.1` and `216.150.16.1`. No AAAA record observed. |
| `www.newsjack.sh` | CNAME to `74fcbc460025be37.vercel-dns-016.com.`. IPv4-only Vercel target; no AAAA record observed. |

Different public resolvers may return nearby Vercel anycast addresses, but the
shape is the same: Vercel-managed IPv4, no IPv6.

## Live Certificate

Observed on `www.newsjack.sh:443`:

| Field | Value |
| --- | --- |
| Subject | `CN=www.newsjack.sh` |
| SAN | `DNS:www.newsjack.sh` |
| Issuer | Let's Encrypt R12 |
| Validity | May 25, 2026 01:41:53 GMT to August 23, 2026 01:41:52 GMT |

## Live Redirects

Observed redirects:

| Request | Result |
| --- | --- |
| `http://newsjack.sh/` | 308 to `https://newsjack.sh/` |
| `http://www.newsjack.sh/` | 308 to `https://www.newsjack.sh/` |
| `https://newsjack.sh/` | 307 to `https://www.newsjack.sh/` |
| `https://www.newsjack.sh/` with browser UA | 308 to `https://github.com/elvisun/newsjack` |
| `https://www.newsjack.sh/` with installer UA | 200 installer shell script |

Canonical host decision:

- The live Vercel domain configuration currently treats `www.newsjack.sh` as the
  canonical custom domain and redirects the apex to `www`.
- Repo metadata and install documentation mostly use the shorter apex
  `https://newsjack.sh`.
- This PR does not change canonical host behavior. Elvis should decide whether
  to keep apex to `www`, or switch to apex as canonical. The modern norm for
  small product/install sites is often `www` to apex, but the current setup may
  have been chosen for Vercel domain management.

Redirect chain risk:

- Browser root load currently has two hops from apex:
  `https://newsjack.sh/` -> `https://www.newsjack.sh/` -> GitHub.
- Installer root load from apex has one host-canonicalization hop:
  `https://newsjack.sh/` -> `https://www.newsjack.sh/` -> installer response.
- There is no apex/www loop.
- The apex to `www` hop is a Vercel/domain-layer 307, not repo code. A 308 would
  be more semantically correct for a permanent canonical redirect, but this PR
  leaves it unchanged.

## Live HSTS

Before this PR, both `newsjack.sh` and `www.newsjack.sh` returned:

```http
Strict-Transport-Security: max-age=63072000
```

No `includeSubDomains` and no `preload` directive were observed.

Vercel documents `strict-transport-security: max-age=63072000` as the default
for custom domains and says projects can modify it with custom response
headers:

- https://vercel.com/docs/cdn-security/encryption
- https://vercel.com/docs/headers/response-headers

This PR changes the app response header to:

```http
Strict-Transport-Security: max-age=300
```

Caveat: because the apex to `www` redirect appears to be Vercel domain-layer
behavior, Elvis should verify after deploy that `https://newsjack.sh/` also
returns `max-age=300`. If Vercel applies its default before project headers on
that redirect, the dashboard/domain setting may need a matching change.

Recommended rollout:

| Phase | HSTS value | Timing |
| --- | --- | --- |
| 1 | `max-age=300` | Now, while reliability is still being validated. |
| 2 | `max-age=86400` | After a couple weeks of cross-network validation. |
| 3 | `max-age=31536000` | Only after confidence is high and without preload. |

Do not add `includeSubDomains` or `preload` without explicit approval.

## HSTS Preload Status

Checked with:

```bash
curl 'https://hstspreload.org/api/v2/status?domain=newsjack.sh'
curl 'https://hstspreload.org/api/v2/preloadable?domain=newsjack.sh'
```

Result:

- Status: `unknown`
- Preloaded domain: empty
- Preloadability errors: missing `includeSubDomains` and missing `preload`

That means `newsjack.sh` is not currently on the preload list.

## TLS Support

Observed behavior:

| Client mode | Result |
| --- | --- |
| TLS 1.0 | Fails with TLS protocol-version alert. |
| TLS 1.1 | Fails with TLS protocol-version alert. |
| TLS 1.2 | Succeeds. |
| TLS 1.3 | Succeeds. |

Reproduction commands:

```bash
curl --tlsv1.0 --tls-max 1.0 -I https://www.newsjack.sh/
curl --tlsv1.1 --tls-max 1.1 -I https://www.newsjack.sh/
curl --tlsv1.2 --tls-max 1.2 -I https://www.newsjack.sh/
openssl s_client -tls1_3 -servername www.newsjack.sh -connect www.newsjack.sh:443 </dev/null
```

Expected:

- TLS 1.0 and 1.1 fail.
- TLS 1.2 and 1.3 connect.

## Likely User-Side Failure Modes

Most likely causes of `ERR_SSL_PROTOCOL_ERROR` on one Windows Chrome machine:

1. Corporate or antivirus TLS interception. Common products include Zscaler,
   Cisco Umbrella, Kaspersky, ESET, Bitdefender, and Sophos. The `.sh` TLD is
   often treated as suspicious or uncategorized.
2. Old client stack that does not support TLS 1.2 by default.
3. ISP or DPI filtering of `.sh` domains.
4. Stale edge state during certificate rotation. This is less likely now because
   the current certificate is valid and serving correctly.

## User Recovery Path Added

This PR adds:

- `apps/site/app/connection-help/page.tsx`
- `docs/connection-help.md`

The recovery guidance tells users to try a personal network, clear Chrome HSTS
state for both `newsjack.sh` and `www.newsjack.sh`, update Chrome/OS, and ask IT
to allow both hosts.

## Older Client Sanity Check

The marketing/root app is server-rendered and does not require client-side
application JavaScript for the basic content. The CSS uses modern niceties like
`text-wrap: balance` through Tailwind utilities; unsupported browsers should
ignore those declarations rather than fail the page. The larger compatibility
gate is TLS 1.2, which is controlled by Vercel's edge and cannot be relaxed from
this repo.

## Left For Elvis To Decide

- Whether the canonical host should remain apex to `www`, or switch to `www` to
  apex.
- Whether the apex/domain-layer redirect can be changed from 307 to 308 in
  Vercel.
- Whether a public status page is worth adding. This PR does not add one.
- When to move HSTS from 5 minutes to 1 day, then later to 1 year.
