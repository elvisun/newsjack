import assert from "node:assert/strict";
import { createElement } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { mock, test } from "node:test";

const responses = [];
const queries = [];
const clientOptions = [];

// Mirrors Contentful's `select` projection, including the SDK only adding
// sys.id and sys.type, so tests catch queries that drop fields the code reads.
function applySelect(item, select) {
  const sys = select.includes("sys")
    ? item.sys
    : { id: item.sys.id, type: "Entry" };
  const fields = Object.fromEntries(
    Object.entries(item.fields).filter(([key]) =>
      select.includes(`fields.${key}`),
    ),
  );
  return { sys, fields };
}

const client = {
  async getEntries(query) {
    queries.push(query);
    const items = responses.shift() ?? [];
    if (!query.select) return { items };
    return { items: items.map((item) => applySelect(item, query.select)) };
  },
};

mock.module("contentful", {
  namedExports: {
    createClient(options) {
      clientOptions.push(options);
      return client;
    },
  },
});
mock.module("next/image.js", {
  defaultExport({ unoptimized, ...props }) {
    assert.equal(unoptimized, true);
    return createElement("img", props);
  },
});

process.env.CONTENTFUL_SPACE_ID = "test-space";
process.env.CONTENTFUL_ACCESS_TOKEN = "test-token";

const {
  NEWSJACK_BLOG_POST_CONTENT_TYPE,
  embeddedImage,
  firstImage,
  getAllPosts,
  getAllSlugs,
  getPostBySlug,
} = await import("../lib/contentful.ts");
const { renderPostImage } = await import(
  "../app/insights/[slug]/post-image.ts"
);

const body = { nodeType: "document", data: {}, content: [] };

function entry(contentType, slug) {
  return {
    sys: {
      id: `${contentType}-${slug}`,
      contentType: { sys: { id: contentType } },
      updatedAt: "2026-10-02T09:30:00.000Z",
    },
    fields: {
      title: `Title for ${slug}`,
      slug,
      excerpt: `Excerpt for ${slug}`,
      body,
      publishedAt: "2026-10-01T12:00:00.000Z",
      author: "Newsjack",
    },
  };
}

test("Contentful reads fail closed when another content type is returned", async () => {
  responses.push(
    [
      entry("otherContentType", "other-post"),
      entry(NEWSJACK_BLOG_POST_CONTENT_TYPE, "newsjack-post"),
    ],
    [entry("otherContentType", "other-post")],
  );

  const errors = [];
  const originalConsoleError = console.error;
  console.error = (...args) => errors.push(args);

  try {
    const posts = await getAllPosts();
    assert.deepEqual(posts.map(({ slug }) => slug), ["newsjack-post"]);

    const post = await getPostBySlug("other-post");
    assert.equal(post, undefined);
  } finally {
    console.error = originalConsoleError;
  }

  assert.deepEqual(clientOptions, [
    {
      space: "test-space",
      accessToken: "test-token",
      environment: "master",
    },
  ]);
  assert.equal(queries[0].content_type, NEWSJACK_BLOG_POST_CONTENT_TYPE);
  assert.deepEqual(queries[0].order, ["-fields.publishedAt"]);
  assert.equal(queries[1].content_type, NEWSJACK_BLOG_POST_CONTENT_TYPE);
  assert.equal(queries[1]["fields.slug"], "other-post");
  assert.equal(errors.length, 2);
});

test("getAllSlugs keeps blog post slugs through the content-type guard", async () => {
  responses.push([
    entry(NEWSJACK_BLOG_POST_CONTENT_TYPE, "newsjack-post"),
    entry("otherContentType", "other-post"),
  ]);

  const originalConsoleError = console.error;
  console.error = () => {};

  try {
    assert.deepEqual(await getAllSlugs(), ["newsjack-post"]);
  } finally {
    console.error = originalConsoleError;
  }
});

test("embedded images use the local proxy in data and rendered post HTML", () => {
  const asset = (fields) => ({
    nodeType: "embedded-asset-block",
    data: { target: { fields } },
    content: [],
  });
  const doc = {
    nodeType: "document",
    data: {},
    content: [
      { nodeType: "paragraph", data: {}, content: [] },
      // A PDF has no image dimensions and can't be a social card.
      asset({ title: "Deck", file: { url: "//assets.ctfassets.net/deck.pdf" } }),
      asset({
        title: "Medialyst homepage",
        description: "Product screenshot",
        file: {
          url: `//images.ctfassets.net/${process.env.CONTENTFUL_SPACE_ID}/asset-id/token/medialyst.jpg`,
          details: { image: { width: 800, height: 600 } },
        },
      }),
    ],
  };

  const image = embeddedImage(doc.content[2]);
  assert.deepEqual(image, {
    src: "/insights/media/asset-id/token/medialyst.jpg",
    width: 800,
    height: 600,
    alt: "Medialyst homepage",
    description: "Product screenshot",
  });
  assert.deepEqual(firstImage(doc), image);

  const html = renderToStaticMarkup(
    createElement("article", null, renderPostImage(image)),
  );
  for (const leakedValue of [
    "ctfassets",
    process.env.CONTENTFUL_SPACE_ID,
  ]) {
    assert.doesNotMatch(image.src, new RegExp(leakedValue, "i"));
    assert.doesNotMatch(html, new RegExp(leakedValue, "i"));
  }
  assert.match(html, /\/insights\/media\/asset-id\/token\/medialyst\.jpg/);
  assert.equal(firstImage(body), undefined);
});
