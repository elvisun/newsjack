# Coverage Reel fixture

This fixture exercises the public `coverage-reel` skill without live articles or publisher material. Every outlet, headline, byline, sentence, logo, metric, URL and review record is synthetic.

- `coverage.csv`, `brand.json`, `clips.json`: the renderer's inputs. Row 2 has no headline (filled from the page and labelled), row 3 has no publish date (reported, never filled), and row 4 stands in for a page that blocked automated capture (left out of the reel and reported).
- `clips/<slug>/`: synthetic press-clip captures (`top.png`, `strip-1.png`, `mention-1.png`, `masthead.png`) at 3x, as press-clip `--scale 3` writes them for the reel's phone formats, and their `clip.json` sidecars, with real mention rectangles so the whole reel renders offline. Regenerate them with `node fixtures/coverage-reel/make-fixture.mjs`; the tests check the committed files match the generator byte for byte.
- `synthetic-article.html`: a made-up article page for the press-clip `--assets` integration test.
- `assets/`: the synthetic brand logo and a placeholder PDF.

Render the dashboard, the reel player, a storyboard and a short MP4 from the repository root (needs `playwright-core` and Chrome or Edge; no ffmpeg):

```bash
node skills/coverage-reel/render.mjs \
  --csv fixtures/coverage-reel/coverage.csv \
  --brand fixtures/coverage-reel/brand.json \
  --clips fixtures/coverage-reel/clips.json \
  --out /tmp/newsjack-coverage-reel-fixture \
  --format landscape --video
```

Run the focused tests (the two browser tests skip themselves when Chrome or Edge is missing):

```bash
node --test fixtures/coverage-reel/coverage-reel.test.mjs
```

The `review` values demonstrate the renderer handoff contract only. They are not evidence of a real press-clip review.
