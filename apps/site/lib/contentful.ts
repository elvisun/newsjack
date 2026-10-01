import { createClient, type ContentfulClientApi, type EntrySkeletonType } from "contentful";
import type { Document } from "@contentful/rich-text-types";

let _client: ContentfulClientApi<undefined> | null = null;

function getClient(): ContentfulClientApi<undefined> {
  if (!_client) {
    if (!process.env.CONTENTFUL_SPACE_ID || !process.env.CONTENTFUL_ACCESS_TOKEN) {
      throw new Error("CONTENTFUL_SPACE_ID and CONTENTFUL_ACCESS_TOKEN must be set");
    }
    _client = createClient({
      space: process.env.CONTENTFUL_SPACE_ID,
      accessToken: process.env.CONTENTFUL_ACCESS_TOKEN,
    });
  }
  return _client;
}

export interface BlogPost {
  title: string;
  slug: string;
  excerpt: string;
  body: Document;
  publishedAt: string;
  author: string;
}

type BlogPostSkeleton = EntrySkeletonType<
  {
    title: string;
    slug: string;
    excerpt: string;
    body: Document;
    publishedAt: string;
    author: string;
  },
  "blogPost"
>;

export async function getAllPosts(): Promise<BlogPost[]> {
  const entries = await getClient().getEntries<BlogPostSkeleton>({
    content_type: "blogPost",
    order: ["-sys.createdAt"],
  });

  return entries.items.map((item) => ({
    title: item.fields.title,
    slug: item.fields.slug,
    excerpt: item.fields.excerpt,
    body: item.fields.body,
    publishedAt: item.fields.publishedAt,
    author: item.fields.author,
  }));
}

export async function getPostBySlug(
  slug: string,
): Promise<BlogPost | undefined> {
  const entries = await getClient().getEntries<BlogPostSkeleton>({
    content_type: "blogPost",
    limit: 1,
    ...({ "fields.slug": slug } as Record<string, string>),
  });

  const item = entries.items[0];
  if (!item) return undefined;

  return {
    title: item.fields.title,
    slug: item.fields.slug,
    excerpt: item.fields.excerpt,
    body: item.fields.body,
    publishedAt: item.fields.publishedAt,
    author: item.fields.author,
  };
}

export async function getAllSlugs(): Promise<string[]> {
  const entries = await getClient().getEntries<BlogPostSkeleton>({
    content_type: "blogPost",
    select: ["fields.slug"],
  });

  return entries.items.map((item) => item.fields.slug);
}
