# Contentful blog workflow

The `/insights` blog reads the `newsjackBlogPost` content type from Contentful.
The Contentful space may contain other content types, but the site does not
render them.

## Content model

In the Contentful web app, the content type is named **Newsjack - Blog post**.
Its ID is `newsjackBlogPost`. Every field is required.

| Field | Type | Rules |
| --- | --- | --- |
| `title` | Short text (Symbol) | Required. |
| `slug` | Short text (Symbol) | Required and unique. Must match `^[a-z0-9]+(?:-[a-z0-9]+)*$`. |
| `excerpt` | Long text | Required. Used on the insights index and for page metadata. |
| `body` | Rich text | Required. See the limits below. |
| `publishedAt` | Date and time | Required. Controls the displayed date and newest-first ordering. |
| `author` | Short text (Symbol) | Required. |

The `body` field supports paragraphs, level-two and level-three headings,
ordered and unordered lists, quotes, dividers, and hyperlinks. It supports
bold, italic, underline, and code marks. Embedded images, assets, and entries
are not supported.

## Create and publish a post

1. Open the Contentful web app and go to **Content**.
2. Create an entry and choose **Newsjack - Blog post**.
3. Complete every field. Use a lowercase, hyphen-separated slug that matches
   the required pattern.
4. Review the rendered structure against the supported rich-text options.
5. Select **Publish**.

Draft entries never render on the site because it uses the Content Delivery
API, which returns published content only.

## Local development

Copy `.env.example` to `.env.local` in `apps/site` and set:

| Variable | Value |
| --- | --- |
| `CONTENTFUL_SPACE_ID` | The Contentful space ID. |
| `CONTENTFUL_ACCESS_TOKEN` | A Content Delivery API token for that space. |

Then start the site:

```bash
cd apps/site
pnpm install
pnpm dev
```

Open `http://localhost:3000/insights`. Keep real credentials in `.env.local`;
never add them to `.env.example` or commit them.

## Delivery and caching

The insights index and known post pages are statically generated at build time.
Both routes use incremental static regeneration with a 60-second revalidation
interval. A post published after a deployment can therefore appear without a
redeploy: after the cached page becomes eligible for revalidation, the next
request starts a refresh, and a following request receives the updated page.

Post slugs not generated during the build are rendered on demand. A newly
published post can be opened directly without waiting for another deployment.
Unpublishing or editing a post follows the same revalidation behavior.

## Content-type guard

All Contentful queries use the exported `newsjackBlogPost` content-type
constant. Each returned entry is checked again at runtime. An entry with any
other content type is dropped and logged with `console.error`. If a single-post
slug lookup receives an entry of another content type, the lookup returns no
post and the route responds with 404.
