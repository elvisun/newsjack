---
name: coverage-tracker-setup
description: "Set up a lightweight Google Alerts-style coverage tracker for any number of keywords. Creates a tracker config with each keyword and what it actually means, then hands recurrence to the user's agent harness."
when_to_use: "User wants to create, configure, or update coverage alerts, brand/company mention tracking, Google Alert-style monitoring, or keyword coverage tracking. Use this instead of newsjack-monitor-setup when the job is to track coverage of the user's own keywords rather than find newsjacking opportunities."
---

# Coverage Tracker Setup

This sets up a simple keyword tracker that the `coverage-tracker` skill runs for you. Think of it like a Google Alert: it watches for new coverage of the keywords you care about and tells you when something real shows up.

It is deliberately a different job from setting up newsjack monitoring. A coverage tracker answers "did my keyword get real coverage?" Newsjack monitoring (`newsjack-monitor-setup`) answers a bigger question: "can my company jump into a broader news story?" If the user wants the second one, use that skill instead.

## Where you're running this

Two situations, and they change what you can finish here:

- **You can save files and run commands** (Claude Code, Codex, OpenClaw, Hermes, or a similar setup with shell, file, and network access). Here you can do the whole thing: save the tracker, run it once, and hand the recurring schedule to your agent. **Before you decide you can't run the CLI, check whether `newsjack` is installed.** It ships as a prebuilt, bundled binary — you do **not** need Go, a compiler, or any build/install step. Never look for a Go toolchain or tell the user they need a "Go environment" without checking first: run `newsjack --version`; if that's not on `PATH`, try the bundled location `~/.newsjack/bin/newsjack --version` and use that full path everywhere if it works. The bundled binary is almost always already installed — verify, don't assume it's missing.
- **You're in a plain chat window** (Claude.ai, ChatGPT, Claude Cowork, or anything without file/command access). Here you can only draft. Write out the tracker config and the schedule prompt, then tell the user to move to a setup that can save and run things. Don't try to install or run anything, and don't tell the user it was saved or scheduled when it wasn't.

## What to ask the user

Only ask for what you don't already know:

- a name for the tracker
- the keywords to watch (as many as they want)
- for each keyword, one short note on what it actually means: which company, product, or person should it point to? This is what later filters out wrong matches.
- optionally, words or uses to ignore when a keyword is ambiguous (for example, a common word that also happens to be a brand name)
- how often to check: every morning, twice a day, or hourly

Don't ask about company standing, spokespeople, competitors, proof assets, news feeds, target reporters, or story angles. All of that is for newsjacking, not for tracking coverage of your own keywords.

## Steps to set it up

1. **Write the tracker config.** This is a small file that the `coverage-tracker` skill reads back every time it runs, so keep its shape exactly as shown:

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

   You can list as many keywords as you want. The `means` line for each one matters: make it specific (which company, product, or person) so the tracker can later throw out mentions of the wrong thing or generic uses of the word.

   If you're in a plain chat window, stop here. Hand the user this config plus the schedule prompt from step 3, and tell them to finish in a setup that can save and run things. Don't claim it was saved or scheduled.

2. **Save the config** with this command (replace `<slug>` with a short name for the tracker):

   ```bash
   newsjack coverage init <slug> --config tracker.json
   ```

   Working inside a copy of the project source, use `bin/newsjack` from the project root instead. Only add `--force` if the user clearly wants to overwrite a tracker that already exists.

3. **Schedule the recurring check through your agent, not your computer's built-in scheduler.** If your agent has a way to schedule prompts, schedule this one:

   ```text
   Use the coverage-tracker skill for <slug>.
   ```

   If your agent can't schedule things itself, tell the user exactly which prompt to set up on a schedule in Claude, Codex, Hermes, OpenClaw, or whatever they use. Do not set up system-level schedulers (cron, launchd, systemd timers, or similar) on the user's machine.

4. **Run it once right away** if the user wanted a working setup, or if you're doing the whole setup end to end. Use the `coverage-tracker` skill for the new slug and tell the user what the first run found.

## Updating a tracker you already have

First, look up where the config lives:

```bash
newsjack coverage status <slug>
```

That shows a `config_path`. Open the config file there, make your edits, then save it again with:

```bash
newsjack coverage init <slug> --config tracker.json --force
```

Only change the keywords or their meaning notes. The record of what's already been seen and alerted on is kept separately by `coverage-tracker` (in its own database) — don't edit that by hand.

## What to tell the user when you're done

Wrap up by giving them:

- the tracker slug (its short name)
- where the config file is saved
- the schedule prompt and how often it'll run
- what the first run found, if you ran it
