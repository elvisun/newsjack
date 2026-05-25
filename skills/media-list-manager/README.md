# media-list-manager

Build and manage small, fit-checked media lists for Newsjack campaigns.

This skill can use the optional Medialyst MCP server for live list creation, inspection, enrichment, views, and share links. Without Medialyst credentials, it still returns a local media-list artifact for review or later import.

## Use it now

### Claude / ChatGPT / Cursor

Load `SKILL.md` as a skill or project file, then provide the campaign angle, target beat, and any source articles.

If your runtime supports project MCP config, the repo root `.mcp.json` points to the Medialyst MCP server. For Claude Code, run the local login helper once:

```bash
newsjack login
claude --strict-mcp-config --mcp-config .mcp.json
```

The helper stores the API key under `~/.newsjack/credentials.json` and the `.mcp.json` `headersHelper` reads it automatically. `MEDIALYST_API_KEY` and repo `.env` still work for advanced users and automation.

### Codex

Codex does not use Claude Code's `headersHelper`. Use the stdio bridge:

```bash
newsjack login
codex mcp add medialyst -- ~/.newsjack/bin/newsjack mcp-bridge
```

### OpenClaw

OpenClaw supports configured MCP stdio servers. Use the same bridge:

```bash
newsjack login
openclaw mcp set medialyst "{\"command\":\"$HOME/.newsjack/bin/newsjack\",\"args\":[\"mcp-bridge\"]}"
```

The bridge requires Node.js because it launches `npx -y mcp-remote`.

## Natural-language invocations

- "Build a media list for this angle"
- "Create a first-wave list for this newsjack"
- "Find journalists who recently covered this topic"
- "Turn these articles into a media list"
- "Add a notes column and share this Medialyst list"
- "Inspect this media list and cut weak fits"
- "Create a saved view for the first wave"

## What you get back

- a small first-wave media list, not a broad database
- per-recipient fit status and anchor evidence
- cut reasons for weak or unsafe rows
- optional Medialyst IDs, views, and share URL when MCP is available
- a local artifact when MCP is not configured

## Files in this skill

| File | Purpose |
| --- | --- |
| `SKILL.md` | Agent instructions and output format. |
| `rubric.md` | Fit, evidence, anti-spam, and MCP management checks. |
| `examples.md` | Worked examples for live and local modes. |
| `scripts/medialyst_auth.py` | Local login and Claude Code MCP header helper. |
| `scripts/medialyst_mcp_bridge.py` | Stdio MCP bridge for Codex, OpenClaw, and other clients without `headersHelper`. |
| `README.md` | This file. |

## License

MIT
