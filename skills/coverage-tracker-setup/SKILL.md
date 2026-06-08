---
name: coverage-tracker-setup
description: "Set up a lightweight Google Alerts-style coverage tracker for any number of keywords. Creates a tracker config with each keyword and what it actually means, then hands recurrence to the user's agent harness."
when_to_use: "User wants to create, configure, or update coverage alerts, brand/company mention tracking, Google Alert-style monitoring, or keyword coverage tracking. Use this instead of newsjack-monitor-setup when the job is to track coverage of the user's own keywords rather than find newsjacking opportunities."
---

# Coverage Tracker Setup

Create a simple keyword tracker for `coverage-tracker`. This workflow is intentionally separate from newsjack monitor profiles: coverage tracking answers "did my keyword get real coverage?", not "can this client newsjack a broader story?"

## Inputs

Ask only for missing facts:

- tracker name
- any number of keywords
- one short `means` snippet per keyword: what entity/product/person the keyword should refer to
- optional exclusions for ambiguous terms
- cadence preference for the agent harness: daily morning, twice daily, or hourly

Do not ask for standing, spokespeople, competitors, proof assets, RSS feeds, target beats, or PR angles. Those belong to newsjacking, not coverage tracking.

## Setup Workflow

1. Build a tiny tracker JSON:

   ```json
   {
     "name": "Profound",
     "lookback_days": 2,
     "keywords": [
       {
         "keyword": "profound",
         "means": "Profound, the AI search analytics company.",
         "exclude_hints": ["generic adjective uses"]
       }
     ]
   }
   ```

   `keywords` may contain any number of entries. Keep each `means` field concrete enough that a later LLM pass can reject wrong-entity and generic mentions.

2. Save it with the CLI:

   ```bash
   newsjack coverage init <slug> --config tracker.json
   ```

   In a source checkout, prefer `bin/newsjack` from the repo root. Use `--force` only when the user explicitly wants to overwrite the existing tracker config.

3. Set up recurrence in the agent harness, not native cron. If the runtime exposes a scheduling feature, schedule this prompt:

   ```text
   Use the coverage-tracker skill for <slug>.
   ```

   If no scheduling tool is available, tell the user exactly which prompt to schedule in Claude, Codex, Hermes, OpenClaw, or their chosen harness. Do not install system cron, launchd, systemd timers, or other native schedulers.

4. Run once immediately if the user asked for a working setup, or if this is an end-to-end setup flow. Use `coverage-tracker` for the new slug and relay its first-run result to the user.

## Updating Existing Trackers

Run:

```bash
newsjack coverage status <slug>
```

Read the `config_path`, edit the tracker JSON, then re-run:

```bash
newsjack coverage init <slug> --config tracker.json --force
```

Only change the keyword aperture or meaning snippet. Alert decisions are stored in SQLite by `coverage-tracker`; do not edit those by hand.

## Output

When setup is complete, tell the user:

- tracker slug
- config path
- schedule prompt/cadence
- first-run result if you ran it
