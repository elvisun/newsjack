# Contributing to newsjack

Thanks for the interest. This is early — the API surface and skill format will change.

## Setup

```bash
git clone https://github.com/elvisun/newsjack
cd newsjack
# (Install steps will land once the CLI / site builds are in.)
```

## Workflow

- Branch: `feat/<short-name>` for features, `fix/<short-name>` for fixes
- Commits: prefer conventional commits (`feat:`, `fix:`, `docs:`, `chore:`, `refactor:`)
- One feature per PR. Small PRs review faster.
- PRs against `main`. Squash on merge.

## Pre-commit checks

This repo runs **gitleaks** in CI on every PR to catch accidental secret leaks. To run locally before pushing:

```bash
brew install gitleaks
gitleaks detect --staged
```

If gitleaks flags something, **do not** force-push to bury it — the secret is already in your local history and may be in the remote if pushed. Rotate the secret immediately, then ask for help cleaning history.

## Skill contributions

If you're adding a new skill under `skills/<skill-name>/`:

- Follow the template established by `skills/meanest-editor/`
- Include `SKILL.md` with the standard frontmatter (`name`, `description`, `when_to_use`)
- Skills should work locally with no required substrate / signup
- Premium / Medialyst-backed behavior is an optional upgrade, never a paywall on the base utility

## Code of conduct

Be honest, be direct, don't be cruel. Same energy as the meanest-editor skill: cuts, but doesn't insult.
