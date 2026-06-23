# Connection Help

Use this when `newsjack.sh` or `www.newsjack.sh` fails with
`ERR_SSL_PROTOCOL_ERROR` in Chrome.

## Quick Recovery

1. Try a non-corporate network or mobile hotspot. If that works, the likely
   cause is corporate, antivirus, or ISP filtering of the `.sh` domain.
2. Clear Chrome's HSTS state:
   - Open `chrome://net-internals/#hsts`.
   - Under "Delete domain security policies", delete `newsjack.sh`.
   - Delete `www.newsjack.sh` too, then try the site again.
3. Update Chrome and the operating system. `newsjack.sh` requires TLS 1.2 or
   newer.
4. If you are on a managed device, ask IT or the security vendor to allow both
   `newsjack.sh` and `www.newsjack.sh`.

## Useful Diagnostic Commands

```bash
curl -I https://www.newsjack.sh/
curl --tls-max 1.1 -I https://www.newsjack.sh/
```

The first command should connect. The second command should fail with a TLS
protocol-version error, which confirms that clients older than TLS 1.2 cannot
connect.
