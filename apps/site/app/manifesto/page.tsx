import type { Metadata } from "next";
import Link from "next/link";
import { ManifestoHero, SiteShell } from "../_components/site-shell";

export const metadata: Metadata = {
  title: "Manifesto | newsjack.sh",
  description: "Open-source operating system for agentic PR.",
};

export default function ManifestoPage() {
  return (
    <SiteShell>
      <main className="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-6xl items-center px-5 py-12 sm:px-8 lg:px-10">
        <div className="w-full">
          <div className="mb-9 flex">
            <Link
              className="inline-flex items-center rounded-md border border-emerald-300/35 bg-emerald-300/10 px-4 py-2.5 font-mono text-sm text-emerald-100 shadow-2xl shadow-emerald-950/30 transition hover:border-emerald-300/70 hover:bg-emerald-300/15 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-emerald-300"
              href="/ethics"
            >
              Read the ethical floor for AI agents in PR -&gt;
            </Link>
          </div>
          <ManifestoHero />
        </div>
      </main>
    </SiteShell>
  );
}
