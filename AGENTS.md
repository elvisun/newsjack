# AGENTS.md

Guidance for coding agents working in this repo.

## Repo Shape

- `apps/cli/` - Go CLI source for `newsjack`. Commands live in `apps/cli/cmd/newsjack/`.
- `apps/site/` - Next.js site and distribution build. `proxy.ts` routes curl/wget to `install.sh`; browser traffic gets the site.
- `bin/newsjack` - source-checkout shim that runs the Go CLI with `go run`. Working in this repo, prefer this shim; `~/.newsjack/bin/newsjack` is the end-user install path that public skills reference.
- `skills/` - public Newsjack skills. These are user-facing runtime instructions.
- `fixtures/` - local fixtures and smoke harnesses. Fixture-specific docs, scripts, prompts, and generated run folders belong here.
- `eval/` - evaluation artifacts, including reverse-newsjack methodology and dated run results.
- `harness/` - local agent/runtime harness support.
- `docs/` - planning and architecture notes.

## Canonical Boundaries

- Public skill behavior belongs in `skills/*/SKILL.md` and skill-local support files.
- Fixture usage belongs in the relevant `fixtures/*/README.md`, scripts, or prompt files, not in public skills.
- `skills/newsjack-detector/SKILL.md` is the canonical detector pipeline contract.
- `fixtures/newsjack-detector-agent/PROMPT.md` should only point agents to the canonical detector skill and the fixture profile files. Do not duplicate run commands or profile lists there.
- `fixtures/newsjack-detector-agent/README.md` owns fixture script usage.

## Common Commands

CLI tests:

```bash
cd apps/cli
go test ./...
```

Site checks:

```bash
cd apps/site
pnpm lint
pnpm test
pnpm build
```

Fixture smoke test without live credentials:

```bash
NEWSJACK_RUN_DIR=/tmp/newsjack-fixture-smoke NEWSJACK_INCLUDE_ALL_SCORED=0 \
  fixtures/newsjack-detector-agent/scripts/run-one-profile.sh simular "computer-use agents" profile.simular.json --mock
```

## Generated Files

- Do not commit `.tmp/`, fixture `runs/`, local databases, `.claude/`, `.next/`, `node_modules/`, or compiled binaries.
- `apps/cli/newsjack` is a local compiled binary and is ignored.
- Dated eval results under `eval/` may be committed when they are intentional evaluation artifacts.
- Check `git status --short` before and after edits. Do not remove unrelated user changes.

## Skill Editing Rules

- Skills are compositional:
  - **ATOM skills** do one judgment or transformation well. Example: `angle-generator`, `story-origin-check`, `relevance-coarse-filter`, `fact-check`, `journalist-fit-check`.
  - **MOLECULE skills** combine a few atoms into a bounded workflow with one clear output. Example: `newsjack-detector` (engine evidence + `story-origin-check` + `angle-generator` into a single freshness-gated opportunity report).
  - **COMPOUND skills** orchestrate multiple molecules/atoms into an end-to-end product workflow. Example (planned): a `generate-client-report` skill that runs `newsjack-detector` plus a report-generation skill.
- Keep atomic capability in the atom. Do not duplicate atom logic inside molecule or compound skills; call, reference, or delegate to the atom instead.
- When adding behavior, place it at the lowest level that owns the concept. If it is reusable judgment, make or update an ATOM. If it is orchestration, keep it in a MOLECULE or COMPOUND.
- Keep public skills runtime-agnostic and user-facing.
- Avoid fixture, beta-client, or one-off evaluation details in public skills.
- Keep prompts and rubrics concise enough for runtime use. Put long local examples or harness notes in fixture docs or reference files.
- Follow `skills/ETHICS.md` and `skills/WHY-NOT-SPAM.md` for all PR/pitch/newsjack workflows.
- For detector changes, preserve the separation:
  - Go CLI owns ingestion, deterministic scoring, filtering application, and freshness gate application.
  - Skills own PR judgment, story-origin reasoning, angle fit, brand safety, and handoff.

## Implementation Notes

- Prefer existing repo patterns over new abstractions.
- Use `rg`/`rg --files` for searches.
- Make minimal, targeted edits; do not rewrite whole files. Use whatever patch/edit tooling your harness provides.
- Keep changes scoped. Do not reformat unrelated files.
- If adding or renaming fixture scripts, update the fixture README and any internal script references.
- If changing CLI behavior, update or add focused Go tests in `apps/cli/cmd/newsjack/`.
- If changing report format, check golden fixtures and run `go test ./...` from `apps/cli`.
