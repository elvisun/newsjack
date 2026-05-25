# Newsjack Agent Runtime Harness

Container-first harness for validating the Newsjack installer against real agent runtime binaries.

## No-Token CI Lane

This mode must not receive model-provider credentials and must not call agents:

```bash
python3 harness/run.py --build-image --mode ci-installer --runtime all --local-source
```

It runs one fresh container per runtime and validates:

- real runtime binary is installed in the image
- Newsjack installs into an isolated container `HOME`
- expected runtime skill directories exist
- `newsjack skills` works
- direct detector mock run works

## Token-Burning Integration Lane

Create:

```bash
cp harness/.env.example harness/.env.local
```

Fill in:

```bash
OPENAI_API_KEY=...
ANTHROPIC_API_KEY=...
MEDIALYST_API_KEY=...
NEWSJACK_HARNESS_ALLOW_MODEL_CALLS=1
```

Then run one runtime first:

```bash
python3 harness/run.py --build-image --mode native-smoke --runtime codex --env-file harness/.env.local --local-source
```

Full setup-flow mode should be run manually until spend and adapter behavior are stable.

Current control-plane coverage:

- Codex: native smoke and ACP setup-flow use `OPENAI_API_KEY`.
- Claude Code: native smoke and ACP setup-flow use `ANTHROPIC_API_KEY`.
- Hermes: native smoke/setup-flow use `ANTHROPIC_API_KEY`; `hermes acp` currently advertises OpenRouter/Hermes setup auth, so ACP smoke needs `OPENROUTER_API_KEY`.
- OpenClaw: native smoke uses `OPENAI_API_KEY`; ACP smoke/setup-flow starts a local Gateway inside the container.

OpenClaw injects a large default context even for trivial prompts. Prefer running it once per integration pass.

## Production Installer Path

Local-source mode tests the current checkout as installed source. Production-path mode validates the public installer and tarball:

```bash
python3 harness/run.py --mode ci-installer --runtime all --production-path
```
