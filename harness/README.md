# Newsjack Agent Runtime Harness

The previous checked-in harness runner was removed during the Go-only CLI cleanup.
Do not add a second product implementation path here.

Current verification lives in the Go CLI package and direct smoke commands:

```bash
(cd apps/cli && go test ./...)
./bin/newsjack detector run "AI search visibility" --mock --limit 1 --emit json
```

If runtime-container coverage is reintroduced, implement the runner as Go or shell
around the `newsjack` binary. Keep detector, auth, MCP, filtering, and summary
behavior behind the Go CLI.
