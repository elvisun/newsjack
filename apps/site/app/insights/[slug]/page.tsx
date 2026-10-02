import type { Metadata } from "next";
import Image from "next/image";
import Link from "next/link";
import { notFound } from "next/navigation";
import { documentToReactComponents } from "@contentful/rich-text-react-renderer";
import { BLOCKS, INLINES, type Block, type Inline, type Node } from "@contentful/rich-text-types";

import { getPostBySlug, getAllSlugs } from "../../../lib/contentful";
import { Footer } from "../../components/footer";
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

  return {
    title: `${post.title} | newsjack.sh`,
    description: post.excerpt,
  };
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

type EmbeddedAssetFields = {
  file?: {
    url?: string;
    details?: { image?: { width: number; height: number } };
  };
  title?: string;
  description?: string;
};

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
      const fields = (node.data?.target as { fields?: EmbeddedAssetFields })?.fields;
      const image = fields?.file?.details?.image;
      // next/image needs intrinsic dimensions; Contentful only sets them on images.
      if (!fields?.file?.url || !image) return null;
      const src = fields.file.url.startsWith("//") ? `https:${fields.file.url}` : fields.file.url;
      return (
        <figure className="mt-8">
          <Image
            src={src}
            alt={fields.title ?? ""}
            width={image.width}
            height={image.height}
            sizes="(min-width: 768px) 720px, 100vw"
            className="h-auto w-full rounded-lg border border-ink/10"
          />
          {fields.description && (
            <figcaption className="mt-2 text-center font-mono text-xs text-ink/40">
              {fields.description}
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
      <Nav />
      <main className="mx-auto max-w-3xl px-6 pt-36 pb-32">
        <Link className="nj-link text-ink/60" href="/insights">
          <span aria-hidden="true">&larr;</span> All posts
        </Link>

        <article className="mt-12">
          <header className="border-b border-ink/10 pb-10">
            <time className="nj-eyebrow">{formatDate(post.publishedAt)}</time>
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
