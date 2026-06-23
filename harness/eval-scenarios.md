# Newsjack Eval Scenarios

These are model/runtime dogfood scenarios, not default CI tests. The agent or
reviewer running them owns the harness choice, model choice, mock services, and
transcript capture. Keep this file to prompts and expected behavior so the same
case can run across Hermes, Codex, Claude Code, OpenClaw, or future harnesses.

## First-Run Setup Uses Medialyst OAuth

Prompt:

```text
help me install https://newsjack.sh from the github repo and setup a daily newsjack monitoring for my company.
```

Expected output:

- The agent installs or verifies Newsjack from the GitHub repo/current checkout,
  then starts the normal setup flow for a local agent runtime.
- If live Medialyst-backed news search or journalist enrichment is useful and
  Medialyst is not configured, the agent uses `newsjack login`.
- The agent relays the exact Medialyst approval URL printed by `newsjack login`
  and tells the user to approve `newsjack CLI`.
- The agent does not ask the user to paste an `mlst_...` API key.
- The agent does not run `newsjack auth set-medialyst --key ...` or
  `newsjack login --key ...` unless the prompt explicitly asks for CI,
  automation, or API-key setup.
- After approval, the agent reports that Medialyst is connected through OAuth.
  A credential inspection should show `medialyst.oauth` and
  `source: newsjack-oauth-device-flow`, with no new `medialyst.api_key`.

## Medialyst Already Connected

Prompt:

```text
help me install https://newsjack.sh from the github repo and setup a daily newsjack monitoring for my company.
```

Initial state:

- Newsjack is installed or available from the current checkout.
- `newsjack auth status` reports Medialyst already configured through OAuth.

Expected output:

- The agent does not run `newsjack login`.
- The agent does not ask for an API key.
- The agent proceeds with monitor setup and may mention that Medialyst is
  already connected for live news search and journalist enrichment.

## API-Key Setup Is Explicitly Requested

Prompt:

```text
Set up Newsjack for CI using a Medialyst API key.
```

Expected output:

- The agent treats API-key setup as an automation-specific path.
- The agent asks for or uses the provided `mlst_...` key only for the CI setup.
- The agent uses `newsjack auth set-medialyst --key <mlst_...>` or documents
  the equivalent environment variable path.
- The agent does not present API keys as the default interactive user setup.
