# Contributing to newsjack

Thank you for your interest in contributing. newsjack is in active early development; the CLI surface and skill format are subject to change before v1.

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

If gitleaks flags a finding, **do not** attempt to rewrite or force-push history to hide it. The secret already exists in your local repository and may have reached the remote. Rotate the credential immediately, then open an issue or contact a maintainer for help cleaning history.

## Skill contributions

If you're adding a new skill under `skills/<skill-name>/`:

- Follow the template established by `skills/meanest-editor/`
- Include a `SKILL.md` with the standard frontmatter (`name`, `description`, `when_to_use`)
- Skills must run locally with no required cloud substrate or signup
- Any Medialyst-backed functionality must remain optional — never a paywall on the base utility

## Code of conduct

Contributors are expected to engage respectfully and in good faith. Be direct, constructive, and specific in feedback; do not direct personal attacks at other contributors. Maintainers reserve the right to remove comments, commits, or contributors that violate this standard.

To report a concern, contact elvis@medialyst.ai.
