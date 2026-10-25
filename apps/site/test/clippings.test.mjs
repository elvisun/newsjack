import assert from "node:assert/strict";
import { test } from "node:test";

import {
  CLIPPINGS,
  FEATURED_LINKEDIN_CLIP,
} from "../lib/clippings.ts";

test("Oleksandra Naidonova clipping stays source-linked and verbatim", () => {
  const clip = FEATURED_LINKEDIN_CLIP;

  assert.equal(
    clip.url,
    "https://www.linkedin.com/posts/alex-naidonova_i-asked-what-pr-tools-you-actually-use-and-activity-7517922322368897024-V-XU",
  );
  assert.equal(clip.image.src, "/clippings/oleksandra-naidonova-linkedin.webp");
  assert.equal(
    clip.mobileImage.src,
    "/clippings/oleksandra-naidonova-linkedin-mobile.webp",
  );
  assert.deepEqual(clip.excerpts, [
    "Media databases that also send your pitches:\nMedialyst, from its founder Elvis Sun and recommended by Beth Trier",
    "Newsjack, open-source PR skills from Elvis Sun",
    "=> Newsjack is the one I can't wait to play with, because it's free and you can plug it straight into Claude / ChatGPT setup.",
  ]);
  assert.equal(
    clip.comment,
    "already planning my first campaign with your Newsjack Elvis Sun!\n\ni'll report back on the results!",
  );
  assert.equal(
    CLIPPINGS.some((item) => "linkedin" in item && item.linkedin === "oleksandra"),
    false,
  );
});
