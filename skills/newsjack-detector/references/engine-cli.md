# Engine CLI Reference

The Go monitoring engine collects evidence, computes mechanical scores, and emits JSON. The `newsjack-detector` skill owns PR judgment, standing, story-origin reasoning, brand safety, and human-facing rendering.

Do not duplicate the CLI surface in this file. Agents should discover current commands, flags, defaults, source availability, and credential recovery from the installed CLI:

```bash
newsjack help
newsjack help login
newsjack help detector
newsjack detector run --help
newsjack doctor
```

In this repo, prefer the source shim:

```bash
./bin/newsjack help detector
./bin/newsjack help login
./bin/newsjack detector run --help
```

For routine profile runs, rely on the profile:

```bash
newsjack detector run --profile profile.json --save
```

Use `--topic` only when the user explicitly asks for a one-off retrieval topic. Durable monitor discovery belongs in `profile.json`, not in ad hoc run terms.

Profile caveat for agents: when `search_terms` are present, retrieval uses them instead of raw `topics + competitors`. Keep `topics`, `competitors`, and `standing` as matching/judgment context; put broad retrieval terms plus named platforms, products, regulators, and competitors in `search_terms` when they must drive collection. Terms must be static, explicit, and provenance-safe: user input, client materials, named entities, or current coverage, not model-remembered sector trends.
