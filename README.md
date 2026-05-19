# newsjack

> **Open-source operating system for agentic PR.**
>
> The skill layer that turns Claude, ChatGPT, and Cursor into PR operators.

```bash
curl newsjack.sh | sh
```

## What is this?

newsjack is the open-source operating system for **agentic PR** — a skill layer that installs into Claude, ChatGPT, Cursor, or any agent runtime that supports skills. Install it once and your agent becomes a PR operator: briefing, newsjacking, pitch critique, angle generation, media list workflow.

Local skills:
- **newsjack-detector** — find newsworthy moments before the wave breaks
- **meanest-editor** — roast your pitch against the rubric pros use
- **angle-generator** — turn one update into ten pitchable hooks
- _(more skills shipping in v1)_

Optional cloud substrate via `newsjack login` unlocks Medialyst-backed media list generation.

## Install

```bash
curl newsjack.sh | sh
```

Don't trust pipe-to-shell? Inspect first:

```bash
curl -fsSL https://newsjack.sh/install.sh
```

Or install via npm:

```bash
npm install -g newsjack
```

## Repo layout

```
apps/
  site/      # marketing site + edge function (UA-sniffs curl vs browser)
  cli/       # the newsjack CLI (install target, skill loader, login)
skills/      # the OSS skill files
install.sh   # what `curl newsjack.sh` serves to curl/wget clients
```

## Why is `curl newsjack.sh | sh` even possible?

The marketing site at `newsjack.sh` and the install script live at the same URL. An edge function sniffs the `User-Agent` header: curl/wget gets the shell script, browsers get the marketing page. The trick (and the script) are both in this repo — see `apps/site/`.

## License

MIT. See [LICENSE](./LICENSE).

## Credits

Built by [@elvissun](https://x.com/elvissun). Powered by [Medialyst](https://medialyst.com).
