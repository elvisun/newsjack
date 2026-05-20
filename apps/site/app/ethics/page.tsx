import type { Metadata } from "next";
import { readFileSync } from "node:fs";
import path from "node:path";
import Link from "next/link";
import ReactMarkdown from "react-markdown";
import { SiteShell } from "../_components/site-shell";

const articleMarkdown = readFileSync(
  path.join(process.cwd(), "app", "ethics", "article.md"),
  "utf8",
);

export const metadata: Metadata = {
  title: "The Ethical Floor for AI Agents in PR | newsjack.sh",
  description:
    "Shared constraints for how AI agents should behave when PR outreach may land in a journalist's inbox.",
};

export default function EthicsPage() {
  return (
    <SiteShell>
      <main className="mx-auto w-full max-w-6xl px-5 py-10 sm:px-8 lg:px-10">
        <nav
          aria-label="Article navigation"
          className="mb-12 flex items-center justify-between gap-4 font-mono text-xs uppercase tracking-[0.18em] text-zinc-500"
        >
          <Link
            className="text-emerald-300 transition hover:text-emerald-200 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-emerald-300"
            href="/manifesto"
          >
            newsjack.sh
          </Link>
          <Link
            className="transition hover:text-emerald-300 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-emerald-300"
            href="/manifesto"
          >
            Manifesto
          </Link>
        </nav>

        <article className="mx-auto max-w-3xl pb-20">
          <ReactMarkdown
            components={{
              h1: ({ children }) => (
                <h1 className="text-balance text-4xl font-semibold leading-tight tracking-tight text-white sm:text-5xl">
                  {children}
                </h1>
              ),
              h2: ({ children }) => (
                <h2 className="mt-14 border-t border-white/10 pt-9 text-balance text-2xl font-semibold leading-snug text-white sm:text-3xl">
                  {children}
                </h2>
              ),
              p: ({ children }) => (
                <p className="mt-6 text-lg leading-9 text-zinc-300">
                  {children}
                </p>
              ),
              blockquote: ({ children }) => (
                <blockquote className="mt-8 border-l-2 border-emerald-300/70 bg-emerald-300/[0.06] px-5 py-4 text-xl leading-9 text-emerald-100 [&>p]:mt-0 [&>p]:text-xl [&>p]:leading-9 [&>p]:text-emerald-100">
                  {children}
                </blockquote>
              ),
              ul: ({ children }) => (
                <ul className="mt-6 list-disc space-y-3 pl-6 text-lg leading-8 text-zinc-300">
                  {children}
                </ul>
              ),
              li: ({ children }) => <li>{children}</li>,
              strong: ({ children }) => (
                <strong className="font-semibold text-zinc-50">
                  {children}
                </strong>
              ),
              em: ({ children }) => (
                <em className="text-zinc-100">{children}</em>
              ),
            }}
          >
            {articleMarkdown}
          </ReactMarkdown>
        </article>
      </main>
    </SiteShell>
  );
}
