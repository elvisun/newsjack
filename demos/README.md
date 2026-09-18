# Demos

Fun, self-contained demo apps built on Newsjack that we want to open source.

Each demo lives in its own subfolder with its own README, dependencies, and run
instructions. Demos are not part of the core CLI or site and are not published
as packages. Treat them as examples and playgrounds, not supported products.

## Adding a demo

1. Create a folder: `demos/<demo-name>/`
2. Add a `README.md` that explains what the demo does and how to run it.
3. Keep it self-contained. Do not import from `apps/` internals; use the
   published CLI, plugin skills, or public APIs instead.
4. Never commit credentials. Use `.env.example` with placeholders.

## Demos

- `news-desk-dealer/` — "What should I newsjack today?" Deals today's Google News headlines onto 15 company desks, Jev and Claude Opus 5 side by side, animated. Mock engine runs with no keys; live engine wired behind a switch. See `news-desk-dealer/README.md`.
