import assert from "node:assert/strict";
import { test } from "node:test";

import { CLIPPINGS, LINKEDIN_POSTS } from "../lib/clippings.ts";

test("Oleksandra Naidonova clipping stays source-linked and verbatim", () => {
  const post = LINKEDIN_POSTS.oleksandra;

  assert.deepEqual(CLIPPINGS.at(-1), { linkedin: "oleksandra" });
  assert.equal(
    post.url,
    "https://www.linkedin.com/posts/alex-naidonova_i-asked-what-pr-tools-you-actually-use-and-activity-7517922322368897024-V-XU",
  );
  assert.equal(post.avatar, "/clippings/oleksandra-naidonova.webp");
  assert.equal(post.avatarAlt, "Oleksandra Naidonova");
  assert.equal(
    post.text,
    "… Media databases that also send your pitches:\nMedialyst, from its founder Elvis Sun and recommended by Beth Trier\n\n… Newsjack, open-source PR skills from Elvis Sun\n\n… => Newsjack is the one I can't wait to play with, because it's free and you can plug it straight into Claude / ChatGPT setup.",
  );
});
