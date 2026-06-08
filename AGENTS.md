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

## Minimal Memory

Keep Newsjack as stateless as practical. Durable state creates hidden baggage on users' machines, makes behavior harder to reason about, and slows iteration.

- Prefer explicit user-owned files, command inputs, and per-run artifacts over background memory.
- If something can be regenerated with one prompt or one cheap deterministic command, do not save it as durable state.
- If an LLM can derive something from live evidence in one prompt, prefer keeping that derivation in the skill/prompt layer instead of encoding a brittle approximation in Go.
- Persist only state that has a clear product owner and lifecycle: monitor profiles, client briefs, schedules, credentials, run provenance, and recurring-run seen/suppression data.
- Do not persist model-inferred topics, learned assumptions, ranking preferences, or other agent guesses unless the user explicitly promotes them into a profile or brief.
- When state is necessary, make it inspectable, portable, and easy to delete. Avoid adding new local databases, caches, or hidden files unless they are clearly justified.
- One-off commands should be as stateless as possible. Recurring monitors may use limited state for hygiene, but that state should not become the source of truth for discovery or judgment.

## Discovery Doctrine

Open discovery, strict surfacing.

- News discovery must be evidence-first, not model-prior-first. Runtime judgment should rely on collected evidence and explicit profile setup, not remembered sector trends.
- Static profile/query terms should come from explicit user input, the client's own materials, named entities such as competitors/products/regulators, or fresh current coverage. Do not persist model-inferred "hot topics" as durable search terms.
- Setup owns durable discovery aperture: a focused set of short broad beat topics, static search terms, competitors, and feeds belong in the monitor profile, not in per-run generated terms. Prefer 6-8 core topics, usually 2-3 words; one-word topics are fine when they name a real beat.
- Runtime detector skills should not invent or store retrieval terms from model memory or today's headlines. If intake is too narrow or too broad, edit the profile setup file instead of adding hidden run-specific discovery state.
- Preserve the separation between intake and judgment: widen discovery when needed, but keep freshness, story-origin, standing, brand-safety, and angle gates strict before surfacing anything as actionable.

## CLI Unix Principle

Keep the Go CLI a deterministic, composable data layer. Before adding behavior to `apps/cli`, ask whether every Newsjack user needs the same testable operation regardless of client, voice, outlet, or prompt strategy. If the answer is no, put it in a skill, fixture, or user-provided renderer instead.

CLI-owned behavior:

- Install/update/runtime/auth/MCP plumbing, local paths, local state, and health checks.
- Evidence ingestion, normalization, dedupe, clustering, source diagnostics, and local seen-store behavior.
- Mechanical scores, routing metadata, hygiene/safety flags, and deterministic thresholds used to order work.
- JSON-to-JSON transforms such as applying decisions, validating artifact contracts, attaching provenance, and computing freshness status codes.
- Machine-readable facts needed by skills to render honestly: URLs, timestamps, source provenance, status codes, counts, exclusions, and diagnostics.

Skill- or user-owned behavior:

- PR judgment, client standing, newsworthiness, brand-safety judgment, angle fit, journalist shape, handoff decisions, and final labels/framing.
- Report structure, Markdown/prose rendering, story hierarchy, "top news" wording, what to emphasize, and how to explain statuses to a human.
- Prompt-shaped taxonomies or wording that should evolve without a CLI release. Prefer adding structured JSON fields in the CLI over hardcoding prose.
- Fixture/eval reference renderers may live under `fixtures/`, but do not treat them as product CLI behavior.

Keep CLI usage details in CLI help. If agents need to know a command, flag, default, source meaning, credential recovery path, or deterministic profile behavior, expose it through `newsjack help`, command-specific help, `--help`, or `doctor`, with tests. Skill reference files may point agents to help and explain judgment doctrine, but should not duplicate flag lists or operational help text that can drift.

Do not add or expand CLI report rendering unless it is strictly a compatibility/debug path. Canonical human-facing reports should be rendered by skills from deterministic JSON artifacts.

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
  - **ATOM skills** do one judgment or transformation well. Example: `angle-generator`, `story-origin-check`, `relevance-coarse-filter`, `fact-check`, `journalist-fit-check`, `news-search`.
  - **MOLECULE skills** combine a few atoms into a bounded workflow with one clear output. Example: `newsjack-detector` (engine evidence + `story-origin-check` + `angle-generator` into a single freshness-gated opportunity report).
  - **COMPOUND skills** orchestrate multiple molecules/atoms into an end-to-end product workflow. Example (planned): a `generate-client-report` skill that runs `newsjack-detector` plus a report-generation skill.
- Keep atomic capability in the atom. Do not duplicate atom logic inside molecule or compound skills; call, reference, or delegate to the atom instead.
- When adding behavior, place it at the lowest level that owns the concept. If it is reusable judgment, make or update an ATOM. If it is orchestration, keep it in a MOLECULE or COMPOUND.
- Keep public skills runtime-agnostic and user-facing.
- Avoid fixture, beta-client, or one-off evaluation details in public skills.
- Keep prompts and rubrics concise enough for runtime use. A skill's rubric and a small number of representative examples live inline in its `SKILL.md` — do not split them into separate `rubric.md`/`examples.md` files. Only long fixture/harness examples or evaluation notes belong under `fixtures/` or other reference files.
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
