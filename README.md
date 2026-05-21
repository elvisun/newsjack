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
- **newsjack-setup** — create a company monitor profile and choose RSS feeds
- **newsjack-detector** — find newsworthy moments before the wave breaks
- **newsworthiness-check** — score whether an event or pitch is actually newsworthy
- **meanest-editor** — roast your pitch against the rubric pros use
- **angle-generator** — turn one update into ten pitchable hooks
- _(more skills shipping in v1)_

Optional cloud substrate via `newsjack login` unlocks Medialyst-backed media list generation.

## Install

```bash
curl newsjack.sh | sh
```

To review the install script before executing it:

```bash
curl -fsSL https://newsjack.sh/install.sh
```

Alternatively, install via npm:

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

## How `curl newsjack.sh | sh` works

The marketing site at `newsjack.sh` and the install script are served from the same URL. An edge function inspects the `User-Agent` header: curl and wget receive the shell script, while browsers receive the marketing page. Both the routing logic and the install script live in this repository under `apps/site/`.

## License

MIT. See [LICENSE](./LICENSE).

## Credits

Built by [@elvissun](https://x.com/elvissun). Powered by [Medialyst](https://medialyst.com).
