import {
  createClient,
  type ContentfulClientApi,
  type Entry,
  type EntryFieldTypes,
  type EntrySkeletonType,
} from "contentful";
import { BLOCKS, type Document, type Node } from "@contentful/rich-text-types";

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
  updatedAt: string;
  author: string;
}

export interface PostImage {
  src: string;
  width: number;
  height: number;
  alt: string;
  description?: string;
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

function toBlogPost(item: Entry<BlogPostSkeleton, undefined>): BlogPost {
  return {
    title: item.fields.title,
    slug: item.fields.slug,
    excerpt: item.fields.excerpt,
    body: item.fields.body,
    publishedAt: item.fields.publishedAt,
    updatedAt: item.sys.updatedAt,
    author: item.fields.author,
  };
}

export async function getAllPosts(): Promise<BlogPost[]> {
  const client = getClient();
  if (!client) return [];

  const entries = await client.getEntries<BlogPostSkeleton>({
    content_type: NEWSJACK_BLOG_POST_CONTENT_TYPE,
    order: ["-fields.publishedAt"],
  });

  return onlyNewsjackBlogPosts(entries.items).map(toBlogPost);
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
  return item && toBlogPost(item);
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

type AssetFields = {
  file?: {
    url?: string;
    details?: { image?: { width: number; height: number } };
  };
  title?: string;
  description?: string;
};

const CONTENTFUL_IMAGE_HOST = "images.ctfassets.net";
const INSIGHTS_MEDIA_PREFIX = "/insights/media";

function proxiedImageUrl(rawUrl: string): string | undefined {
  let url: URL;
  try {
    url = new URL(rawUrl.startsWith("//") ? `https:${rawUrl}` : rawUrl);
  } catch {
    return undefined;
  }

  if (url.protocol !== "https:" || url.hostname !== CONTENTFUL_IMAGE_HOST) {
    return undefined;
  }

  const [, spaceId, assetId, token, filename, ...extra] =
    url.pathname.split("/");
  if (
    extra.length > 0 ||
    !spaceId ||
    spaceId !== process.env.CONTENTFUL_SPACE_ID ||
    !assetId ||
    !token ||
    !filename
  ) {
    return undefined;
  }

  return `${INSIGHTS_MEDIA_PREFIX}/${assetId}/${token}/${filename}${url.search}`;
}

// Image fields of an embedded-asset node. Non-image assets (PDFs) have no
// dimensions and return undefined.
export function embeddedImage(node: Node): PostImage | undefined {
  const fields = (node.data?.target as { fields?: AssetFields } | undefined)
    ?.fields;
  const url = fields?.file?.url;
  const image = fields?.file?.details?.image;
  if (!url || !image) return undefined;

  const src = proxiedImageUrl(url);
  if (!src) return undefined;

  return {
    src,
    width: image.width,
    height: image.height,
    alt: fields.title ?? "",
    description: fields.description,
  };
}

// First embedded image in a post, used as its social card and schema image.
export function firstImage(body: Document): PostImage | undefined {
  for (const node of body.content) {
    if (node.nodeType !== BLOCKS.EMBEDDED_ASSET) continue;
    const image = embeddedImage(node);
    if (image) return image;
  }
  return undefined;
}
