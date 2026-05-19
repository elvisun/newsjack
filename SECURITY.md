# Security Policy

## Reporting a vulnerability

If you discover a security vulnerability in newsjack — in the CLI, the hosted substrate, the install flow, or any related code — **please report it privately** before public disclosure.

**Email:** elvis@medialyst.com
**Subject prefix:** `[newsjack security]`

Please include:
- A description of the vulnerability and its impact
- Steps to reproduce
- Affected versions or commit SHAs if known
- Your suggested remediation, if any

## What to expect

- **Acknowledgement** within 48 hours
- **Initial assessment** within 5 business days
- **Coordinated disclosure** — we'll work with you on timing, typically 30-90 days depending on severity and fix complexity

## Scope

In scope:
- The newsjack CLI and OSS skills in this repository
- The install flow at `newsjack.sh` and `newsjack.sh/install.sh`
- Any auth flows that bridge to the Medialyst substrate

Out of scope:
- Issues in third-party agent runtimes (Claude, ChatGPT, Cursor, etc.) — report those to the respective vendors
- Issues in Medialyst that are not exposed through the newsjack interface — report to security@medialyst.com

## Acknowledgements

Reporters of valid issues will be credited in release notes (with permission).
