import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import { documentToReactComponents } from "@contentful/rich-text-react-renderer";
import { BLOCKS, INLINES, type Block, type Inline, type Node } from "@contentful/rich-text-types";

import {
  embeddedImage,
  firstImage,
  getAllSlugs,
  getPostBySlug,
  type BlogPost,
} from "../../../lib/contentful";
import {
  OG_IMAGE,
  SITE_NAME,
  SITE_URL,
  absoluteUrl,
  pageMetadata,
} from "../../../lib/site";
import { Footer } from "../../components/footer";
import { JsonLd } from "../../components/json-ld";
import { Nav } from "../../components/nav";

export const revalidate = 60;

type Params = { slug: string };

export async function generateStaticParams(): Promise<Params[]> {
  const slugs = await getAllSlugs();
  return slugs.map((slug) => ({ slug }));
}

export async function generateMetadata({
  params,
}: {
  params: Promise<Params>;
}): Promise<Metadata> {
  const { slug } = await params;
  const post = await getPostBySlug(slug);
  if (!post) return {};

  const image = firstImage(post.body);
  return pageMetadata({
    title: post.title,
    description: post.excerpt,
    path: `/insights/${post.slug}`,
    image: image
      ? { url: image.src, width: image.width, height: image.height, alt: image.alt }
      : OG_IMAGE,
    article: {
      publishedTime: post.publishedAt,
      modifiedTime: post.updatedAt,
      authors: [post.author],
    },
  });
}

function postJsonLd(post: BlogPost) {
  const url = absoluteUrl(`/insights/${post.slug}`);
  // author is free text in Contentful: the team byline is the org itself.
  const author = /newsjack/i.test(post.author)
    ? { "@type": "Organization", name: post.author, url: SITE_URL }
    : { "@type": "Person", name: post.author };

  return {
    "@context": "https://schema.org",
    "@type": "BlogPosting",
    headline: post.title,
    description: post.excerpt,
    datePublished: post.publishedAt,
    dateModified: post.updatedAt,
    author,
    publisher: {
      "@type": "Organization",
      name: SITE_NAME,
      url: SITE_URL,
      logo: { "@type": "ImageObject", url: absoluteUrl("/newsjack-logo.png") },
    },
    image: absoluteUrl(firstImage(post.body)?.src ?? OG_IMAGE.url),
    mainEntityOfPage: url,
    url,
  };
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

const richTextOptions = {
  renderNode: {
    [BLOCKS.PARAGRAPH]: (_node: unknown, children: React.ReactNode) => (
      <p className="mt-6 text-[17px] leading-[1.7] text-ink/80">{children}</p>
    ),
    [BLOCKS.HEADING_2]: (_node: unknown, children: React.ReactNode) => (
      <h2 className="mt-14 font-serif text-[2rem] leading-[1.15] italic">
        {children}
      </h2>
    ),
    [BLOCKS.HEADING_3]: (_node: unknown, children: React.ReactNode) => (
      <h3 className="mt-10 font-serif text-2xl leading-[1.2] font-medium">
        {children}
      </h3>
    ),
    [BLOCKS.UL_LIST]: (_node: unknown, children: React.ReactNode) => (
      <ul className="mt-4 list-disc space-y-2 pl-6 text-ink/80 marker:text-accent">
        {children}
      </ul>
    ),
    [BLOCKS.OL_LIST]: (_node: unknown, children: React.ReactNode) => (
      <ol className="mt-4 list-decimal space-y-2 pl-6 text-ink/80 marker:font-mono marker:text-sm marker:text-ink/40">
        {children}
      </ol>
    ),
    [BLOCKS.LIST_ITEM]: (_node: unknown, children: React.ReactNode) => (
      <li className="text-[17px] leading-[1.7] [&>p]:mt-0">{children}</li>
    ),
    [BLOCKS.QUOTE]: (_node: unknown, children: React.ReactNode) => (
      <blockquote className="my-12 border-l-4 border-accent pl-8 [&_p]:font-serif [&_p]:text-[clamp(1.5rem,3vw,2rem)] [&_p]:leading-[1.25] [&_p]:text-ink [&_p]:italic [&_p:first-child]:mt-0">
        {children}
      </blockquote>
    ),
    [BLOCKS.HR]: () => <hr className="my-14 border-ink/10" />,
    [BLOCKS.EMBEDDED_ASSET]: (node: Node) => {
      // next/image needs intrinsic dimensions; Contentful only sets them on images.
      const image = embeddedImage(node);
      if (!image) return null;
      return (
        <figure className="mt-8">
          <Image
            src={image.src}
            alt={image.alt}
            width={image.width}
            height={image.height}
            sizes="(min-width: 768px) 720px, 100vw"
            className="h-auto w-full rounded-lg border border-ink/10"
          />
          {image.description && (
            <figcaption className="mt-2 text-center font-mono text-xs text-ink/40">
              {image.description}
            </figcaption>
          )}
        </figure>
      );
    },
    [INLINES.HYPERLINK]: (
      node: Block | Inline,
      children: React.ReactNode,
    ) => (
      <a
        href={(node.data as { uri: string }).uri}
        className="text-ink underline decoration-accent/40 underline-offset-2 transition-colors hover:text-accent hover:decoration-accent"
        target="_blank"
        rel="noreferrer"
      >
        {children}
      </a>
    ),
  },
};

export default async function PostPage({
  params,
}: {
  params: Promise<Params>;
}) {
  const { slug } = await params;
  const post = await getPostBySlug(slug);
  if (!post) notFound();

  return (
    <>
      <JsonLd data={postJsonLd(post)} />
      <Nav />
      <main className="mx-auto max-w-3xl px-6 pt-36 pb-32">
        <Link className="nj-link text-ink/60" href="/insights">
          <span aria-hidden="true">&larr;</span> All posts
        </Link>

        <article className="mt-12">
          <header className="border-b border-ink/10 pb-10">
            <time className="nj-eyebrow" dateTime={post.publishedAt}>
              {formatDate(post.publishedAt)}
            </time>
            <h1 className="mt-6 font-serif text-[clamp(2.5rem,6vw,4rem)] leading-[0.95] tracking-[-0.03em] italic">
              {post.title}
            </h1>
            <p className="nj-meta mt-6">By {post.author}</p>
          </header>

          <div className="mt-4">
            {documentToReactComponents(post.body, richTextOptions)}
          </div>
        </article>

        <div className="mt-20 border-t border-ink/10 pt-8">
          <Link className="nj-link text-ink/60" href="/insights">
            <span aria-hidden="true">&larr;</span> Back to all posts
          </Link>
        </div>
      </main>
      <Footer />
    </>
  );
}
