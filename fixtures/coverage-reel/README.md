# Coverage Reel fixture

This fixture exercises the public `coverage-reel` skill without live articles or publisher material. Every outlet, headline, logo, preview, metric, URL, and review record is synthetic.

Render the dashboard and optional local preview/video from the repository root:

```bash
node skills/coverage-reel/render.mjs \
  --csv fixtures/coverage-reel/coverage.csv \
  --brand fixtures/coverage-reel/brand.json \
  --clips fixtures/coverage-reel/clips.json \
  --out /tmp/newsjack-coverage-reel-fixture \
  --preview --video
```

Run the focused tests:

```bash
node --test fixtures/coverage-reel/coverage-reel.test.mjs
```

The `review` values demonstrate the renderer handoff contract only. They are not evidence of a real press-clip review.
