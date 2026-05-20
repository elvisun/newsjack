You are testing the local newsjack-detector skill in this repo.

First read:

- `../../skills/newsjack-detector/SKILL.md`
- `../../skills/newsjack-detector/rubric.md`

Use this monitor profile:

- `profile.acme-ai.json`

Run the monitoring engine from this fixture directory. Start with a mock pass:

```bash
python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run "AI regulation" --profile profile.acme-ai.json --mock --emit json
```

Then, if `MEDIALYST_API_KEY` is available and `xurl whoami` succeeds, run a live quick pass:

```bash
python3 ../../skills/newsjack-detector/scripts/newsjack_detector.py run "AI regulation" --profile profile.acme-ai.json --sources news_search,x --lookback-days 7 --depth quick --save --emit json
```

Apply the skill rubric to the returned `signals`. Return the skill's required JSON object with `opportunities`, `rejected_signals`, `brand_safety_blocks`, and `monitor_notes`.
