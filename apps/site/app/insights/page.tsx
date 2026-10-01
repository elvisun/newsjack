import type { Metadata } from "next";
import Link from "next/link";

import { getAllPosts } from "../../lib/contentful";
import { Nav } from "../components/nav";

export const revalidate = 60;

export const metadata: Metadata = {
  title: "Insights | newsjack.sh",
  description:
    "Ideas, guides, and dispatches on agentic PR from the newsjack team.",
};

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
  });
}

export default async function InsightsPage() {
  const posts = await getAllPosts();

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,rgba(34,197,94,0.08),transparent_28rem),linear-gradient(135deg,#090b0f_0%,#101217_54%,#050607_100%)] text-zinc-50">
      <Nav />
      <main className="mx-auto max-w-3xl px-5 pb-20 pt-12 sm:px-8">
        <p className="font-mono text-xs uppercase tracking-[0.22em] text-emerald-300">
          Insights
        </p>
        <h1 className="mt-5 text-4xl font-semibold tracking-tight text-white sm:text-5xl">
          From the newsjack team
        </h1>
        <p className="mt-5 text-lg leading-8 text-zinc-300">
          Ideas, guides, and dispatches on agentic PR.
        </p>

        {posts.length === 0 ? (
          <p className="mt-16 text-center text-zinc-500">
            No posts yet. Check back soon.
          </p>
        ) : (
          <ul className="mt-12 space-y-10">
            {posts.map((post) => (
              <li key={post.slug}>
                <Link
                  href={`/insights/${post.slug}`}
                  className="group block rounded-lg border border-white/[0.06] bg-white/[0.02] p-6 transition hover:border-emerald-300/30 hover:bg-emerald-300/[0.04]"
                >
                  <time className="font-mono text-xs text-zinc-500">
                    {formatDate(post.publishedAt)}
                  </time>
                  <h2 className="mt-2 text-xl font-semibold text-white group-hover:text-emerald-200">
                    {post.title}
                  </h2>
                  <p className="mt-2 leading-7 text-zinc-400">
                    {post.excerpt}
                  </p>
                  <p className="mt-4 font-mono text-xs text-zinc-500">
                    {post.author}
                  </p>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </main>

      <footer className="mx-auto flex h-16 w-full max-w-3xl items-center px-5 font-mono text-xs text-zinc-500 sm:px-8">
        MIT licensed · github.com/elvisun/newsjack
      </footer>
    </div>
  );
}
