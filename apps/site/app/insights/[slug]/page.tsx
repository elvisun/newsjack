import type { Metadata } from "next";
import Link from "next/link";
import { notFound } from "next/navigation";
import { documentToReactComponents } from "@contentful/rich-text-react-renderer";
import { BLOCKS, INLINES, type Block, type Inline } from "@contentful/rich-text-types";

import { getPostBySlug, getAllSlugs } from "../../../lib/contentful";
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

const richTextOptions = {
  renderNode: {
    [BLOCKS.PARAGRAPH]: (_node: unknown, children: React.ReactNode) => (
      <p className="mt-6 leading-8 text-zinc-300">{children}</p>
    ),
    [BLOCKS.HEADING_2]: (_node: unknown, children: React.ReactNode) => (
      <h2 className="mt-10 text-2xl font-semibold text-white">{children}</h2>
    ),
    [BLOCKS.HEADING_3]: (_node: unknown, children: React.ReactNode) => (
      <h3 className="mt-8 text-xl font-semibold text-white">{children}</h3>
    ),
    [BLOCKS.UL_LIST]: (_node: unknown, children: React.ReactNode) => (
      <ul className="mt-4 list-disc space-y-2 pl-6 text-zinc-300">
        {children}
      </ul>
    ),
    [BLOCKS.OL_LIST]: (_node: unknown, children: React.ReactNode) => (
      <ol className="mt-4 list-decimal space-y-2 pl-6 text-zinc-300">
        {children}
      </ol>
    ),
    [BLOCKS.LIST_ITEM]: (_node: unknown, children: React.ReactNode) => (
      <li className="leading-7">{children}</li>
    ),
    [BLOCKS.QUOTE]: (_node: unknown, children: React.ReactNode) => (
      <blockquote className="mt-6 border-l-2 border-emerald-400/40 pl-5 text-zinc-400 italic">
        {children}
      </blockquote>
    ),
    [BLOCKS.HR]: () => (
      <hr className="my-10 border-white/10" />
    ),
    [INLINES.HYPERLINK]: (
      node: Block | Inline,
      children: React.ReactNode,
    ) => (
      <a
        href={(node.data as { uri: string }).uri}
        className="text-emerald-300 underline decoration-emerald-300/30 underline-offset-2 transition hover:decoration-emerald-300"
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
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,rgba(34,197,94,0.08),transparent_28rem),linear-gradient(135deg,#090b0f_0%,#101217_54%,#050607_100%)] text-zinc-50">
      <Nav />
      <main className="mx-auto max-w-3xl px-5 pb-20 pt-12 sm:px-8">
        <Link
          href="/insights"
          className="inline-flex items-center gap-1.5 font-mono text-xs text-zinc-500 transition hover:text-emerald-300"
        >
          <span aria-hidden="true">&larr;</span> All posts
        </Link>

        <article className="mt-8">
          <header>
            <time className="font-mono text-xs text-zinc-500">
              {formatDate(post.publishedAt)}
            </time>
            <h1 className="mt-3 text-4xl font-semibold tracking-tight text-white sm:text-5xl">
              {post.title}
            </h1>
            <p className="mt-4 font-mono text-sm text-zinc-400">
              {post.author}
            </p>
          </header>

          <div className="mt-10">
            {documentToReactComponents(post.body, richTextOptions)}
          </div>
        </article>

        <div className="mt-16 border-t border-white/10 pt-8">
          <Link
            href="/insights"
            className="inline-flex items-center gap-1.5 font-mono text-xs text-zinc-500 transition hover:text-emerald-300"
          >
            <span aria-hidden="true">&larr;</span> Back to all posts
          </Link>
        </div>
      </main>

      <footer className="mx-auto flex h-16 w-full max-w-3xl items-center px-5 font-mono text-xs text-zinc-500 sm:px-8">
        MIT licensed · github.com/elvisun/newsjack
      </footer>
    </div>
  );
}
