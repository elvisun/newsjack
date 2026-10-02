import {
  createClient,
  type ContentfulClientApi,
  type EntryFieldTypes,
  type EntrySkeletonType,
} from "contentful";
import type { Document } from "@contentful/rich-text-types";

export const NEWSJACK_BLOG_POST_CONTENT_TYPE = "newsjackBlogPost";

let _client: ContentfulClientApi<undefined> | null = null;

function getClient(): ContentfulClientApi<undefined> | null {
  if (!process.env.CONTENTFUL_SPACE_ID || !process.env.CONTENTFUL_ACCESS_TOKEN) {
    return null;
  }
  if (!_client) {
    _client = createClient({
      space: process.env.CONTENTFUL_SPACE_ID,
      accessToken: process.env.CONTENTFUL_ACCESS_TOKEN,
      environment: "master",
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
    title: EntryFieldTypes.Symbol;
    slug: EntryFieldTypes.Symbol;
    excerpt: EntryFieldTypes.Text;
    body: EntryFieldTypes.RichText;
    publishedAt: EntryFieldTypes.Date;
    author: EntryFieldTypes.Symbol;
  },
  typeof NEWSJACK_BLOG_POST_CONTENT_TYPE
>;

type EntryWithContentType = {
  sys: {
    id: string;
    contentType: {
      sys: {
        id: string;
      };
    };
  };
};

function onlyNewsjackBlogPosts<T extends EntryWithContentType>(
  items: T[],
): T[] {
  return items.filter((item) => {
    const contentType = item.sys.contentType?.sys?.id;
    if (contentType === NEWSJACK_BLOG_POST_CONTENT_TYPE) return true;

    if (contentType) {
      console.error("Dropped Contentful entry with unexpected content type", {
        entryId: item.sys.id,
      });
    }
    return false;
  });
}

export async function getAllPosts(): Promise<BlogPost[]> {
  const client = getClient();
  if (!client) return [];

  const entries = await client.getEntries<BlogPostSkeleton>({
    content_type: NEWSJACK_BLOG_POST_CONTENT_TYPE,
    order: ["-fields.publishedAt"],
  });

  return onlyNewsjackBlogPosts(entries.items).map((item) => ({
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
  const client = getClient();
  if (!client) return undefined;

  const entries = await client.getEntries<BlogPostSkeleton>({
    content_type: NEWSJACK_BLOG_POST_CONTENT_TYPE,
    limit: 1,
    "fields.slug": slug,
  });

  const item = onlyNewsjackBlogPosts(entries.items)[0];
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
  const client = getClient();
  if (!client) return [];

  const entries = await client.getEntries<BlogPostSkeleton>({
    content_type: NEWSJACK_BLOG_POST_CONTENT_TYPE,
    // The SDK only adds sys.id and sys.type to a select; the content-type
    // guard needs sys.contentType too.
    select: ["sys", "fields.slug"],
  });

  return onlyNewsjackBlogPosts(entries.items).map((item) => item.fields.slug);
}
