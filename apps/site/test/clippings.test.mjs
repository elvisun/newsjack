import assert from "node:assert/strict";
import { test } from "node:test";

import {
  CLIPPINGS,
  INSTAGRAM_POSTS,
  LINKEDIN_POSTS,
} from "../lib/clippings.ts";

test("Oleksandra Naidonova clipping stays source-linked and verbatim", () => {
  const post = LINKEDIN_POSTS.oleksandra;

  assert.ok(CLIPPINGS.some((item) => item.linkedin === "oleksandra"));
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

// The wall is append-only: new clippings go at the end, in this order.
test("newest clippings are appended in order", () => {
  assert.deepEqual(CLIPPINGS.slice(-4), [
    { linkedin: "oleksandra" },
    { linkedin: "kateryna" },
    { instagram: "mariah" },
    { instagram: "tonyMa" },
  ]);
});

test("Kateryna Leliukh clipping uses her saved avatar and verbatim excerpt", () => {
  const post = LINKEDIN_POSTS.kateryna;

  assert.equal(
    post.url,
    "https://www.linkedin.com/posts/kateryna-leliukh_pr-mediarelations-ai-activity-7528044561693319168-VP6O",
  );
  assert.equal(post.avatar, "/clippings/kateryna-leliukh.jpg");
  assert.match(post.text, /^Last Friday evening I traded my PR shotgun/);
  assert.match(post.text, /an associate who never sleeps\.$/);
});

test("Mariah Brunner reel keeps local media and links to the reel", () => {
  const post = INSTAGRAM_POSTS.mariah;

  assert.equal(post.url, "https://www.instagram.com/reel/DfCOEvfgGHz/");
  assert.equal(post.avatar, "/clippings/mariah-brunner.jpg");
  assert.equal(post.cover.src, "/clippings/mariah-brunner-reel.jpg");
  assert.match(post.text, /Newsjack is a free set of PR skills/);
});

test("Tony Ma reel keeps its verified identity, exact metrics, and source", () => {
  const post = INSTAGRAM_POSTS.tonyMa;

  assert.deepEqual(CLIPPINGS.at(-1), { instagram: "tonyMa" });
  assert.equal(post.url, "https://www.instagram.com/reel/DfXdmf_srb5/");
  assert.equal(post.displayName, "Tony Ma");
  assert.equal(post.username, "tonyma.ai");
  assert.equal(post.verified, true);
  assert.deepEqual(post.stats, { likes: 26, comments: 109 });
  assert.equal(post.avatar, "/clippings/tony-ma.jpg");
  assert.equal(post.cover.src, "/clippings/tony-ma-reel.jpg");
  assert.equal(post.publishedAt, "2026-11-05T16:30:19Z");
  assert.equal(
    post.text,
    "Mit dem kostenlosen Open-Source-Projekt Newsjack kannst du deinen KI-Agenten zum persönlichen PR-Assistenten machen.",
  );
});
