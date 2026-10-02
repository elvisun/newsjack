import type { MetadataRoute } from "next";

import { getAllPosts } from "../lib/contentful";
import { SITE_URL, absoluteUrl } from "../lib/site";

// Matches /insights so newly published posts are listed within a minute.
export const revalidate = 60;

export default async function sitemap(): Promise<MetadataRoute.Sitemap> {
  const posts = await getAllPosts();
  const latestUpdate = posts
    .map((post) => post.updatedAt)
    .sort()
    .at(-1);

  return [
    { url: SITE_URL },
    { url: absoluteUrl("/insights"), lastModified: latestUpdate },
    ...posts.map((post) => ({
      url: absoluteUrl(`/insights/${post.slug}`),
      lastModified: post.updatedAt,
    })),
    { url: absoluteUrl("/connection-help") },
  ];
}
