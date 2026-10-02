import type { Metadata } from "next";
import Link from "next/link";

import { getAllPosts } from "../../lib/contentful";
import { Footer } from "../components/footer";
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
    <>
      <Nav />
      <main className="mx-auto max-w-5xl px-6 pt-36 pb-32">
        <span className="nj-eyebrow">Insights</span>
        <h1 className="mt-6 font-serif text-[clamp(2.5rem,7vw,4.5rem)] leading-[0.95] tracking-[-0.03em] italic">
          From the newsjack desk.
        </h1>
        <p className="mt-6 max-w-xl text-[17px] leading-[1.7] text-ink/80">
          Ideas, guides, and dispatches on agentic PR.
        </p>

        {posts.length === 0 ? (
          <p className="mt-24 border-t border-ink/10 pt-8 text-ink/60 italic">
            No posts yet. Check back soon.
          </p>
        ) : (
          <ul className="mt-20 border-b border-ink/10">
            {posts.map((post) => (
              <li className="border-t border-ink/10" key={post.slug}>
                <Link
                  className="group grid gap-4 py-10 md:grid-cols-12 md:gap-8"
                  href={`/insights/${post.slug}`}
                >
                  <time className="nj-meta md:col-span-3 md:pt-2">
                    {formatDate(post.publishedAt)}
                  </time>
                  <div className="md:col-span-9">
                    <h2 className="font-serif text-[clamp(1.75rem,3.5vw,2.5rem)] leading-[1.1] italic transition-colors group-hover:text-accent">
                      {post.title}
                    </h2>
                    <p className="mt-3 max-w-2xl leading-[1.7] text-ink/60">
                      {post.excerpt}
                    </p>
                    <p className="nj-meta mt-5">By {post.author}</p>
                  </div>
                </Link>
              </li>
            ))}
          </ul>
        )}
      </main>
      <Footer />
    </>
  );
}
