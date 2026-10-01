import type { Metadata } from "next";

import { Nav } from "../components/nav";

export const metadata: Metadata = {
  title: "About | newsjack.sh",
  description:
    "What newsjack is, how it works, and how to install it on any agent platform.",
};

function CheckIcon() {
  return (
    <svg
      className="mt-0.5 h-4 w-4 shrink-0 text-emerald-400"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth={2.5}
    >
      <path strokeLinecap="round" strokeLinejoin="round" d="M5 13l4 4L19 7" />
    </svg>
  );
}

export default function AboutPage() {
  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,rgba(34,197,94,0.08),transparent_28rem),linear-gradient(135deg,#090b0f_0%,#101217_54%,#050607_100%)] text-zinc-50">
      <Nav />
      <main className="mx-auto max-w-3xl px-5 pb-20 pt-12 sm:px-8">
        <p className="font-mono text-xs uppercase tracking-[0.22em] text-emerald-300">
          About
        </p>
        <h1 className="mt-5 text-4xl font-semibold tracking-tight text-white sm:text-5xl">
          Your agent, but better at PR
        </h1>
        <p className="mt-6 text-lg leading-8 text-zinc-300">
          Newsjack is an open-source set of skills — plain-Markdown instructions
          your agent reads — plus a small CLI. Install once, and your agent can
          monitor your industry, generate story angles, fact-check pitches,
          build media lists, and more.
        </p>

        <section className="mt-14">
          <h2 className="text-2xl font-semibold text-white">How it works</h2>
          <p className="mt-4 leading-8 text-zinc-300">
            Each skill is a Markdown file that tells your agent exactly what to do
            and how to do it — step by step, with guardrails built in. No API keys
            are required for the core set. Some skills reach for a live news index
            or journalist enrichment through the Medialyst API, but most run
            anywhere your agent does.
          </p>
          <p className="mt-4 leading-8 text-zinc-300">
            Skills work across platforms: Claude Code, Codex, Hermes, OpenClaw,
            Claude.ai, Cowork, and ChatGPT. Local agents get the full experience
            with scheduling and saved state. Browser agents get a best-effort,
            one-shot pass.
          </p>
        </section>

        <section className="mt-14">
          <h2 className="text-2xl font-semibold text-white">Install</h2>

          <div className="mt-6 space-y-6">
            <div>
              <h3 className="font-semibold text-white">
                Local agents (Claude Code, Codex, Hermes, OpenClaw)
              </h3>
              <pre className="mt-3 overflow-x-auto rounded-lg border border-white/10 bg-black/60 p-4 font-mono text-sm text-emerald-200">
                <code>curl -fsSL newsjack.sh | bash</code>
              </pre>
              <p className="mt-2 text-sm text-zinc-400">
                Or copy this to any AI:{" "}
                <code className="text-emerald-300/70">
                  help me setup https://newsjack.sh
                </code>
              </p>
            </div>

            <div>
              <h3 className="font-semibold text-white">Claude.ai &amp; Cowork</h3>
              <p className="mt-2 leading-7 text-zinc-300">
                Install the{" "}
                <a
                  href="https://claude.ai/customize"
                  className="text-emerald-300 underline decoration-emerald-300/30 underline-offset-2 hover:decoration-emerald-300"
                  target="_blank"
                  rel="noreferrer"
                >
                  newsjack plugin
                </a>{" "}
                from the Anthropic marketplace. Go to Customize → Personal plugins →
                Add marketplace → Add from repository → enter{" "}
                <code className="text-emerald-300/70">elvisun/newsjack</code>.
              </p>
            </div>
          </div>
        </section>

        <section className="mt-14">
          <h2 className="text-2xl font-semibold text-white">Principles</h2>
          <ul className="mt-6 space-y-4 text-zinc-300">
            {[
              "Open source — MIT licensed, no vendor lock-in",
              "Skills are plain Markdown, not compiled plugins",
              "Local-first — your data stays on your machine",
              "Works across agent platforms",
              "Earned media only — never sends on your behalf",
            ].map((item) => (
              <li key={item} className="flex items-start gap-3">
                <CheckIcon />
                <span className="leading-6">{item}</span>
              </li>
            ))}
          </ul>
        </section>

        <section className="mt-14">
          <h2 className="text-2xl font-semibold text-white">Links</h2>
          <ul className="mt-6 space-y-3 font-mono text-sm">
            <li>
              <a
                href="https://github.com/elvisun/newsjack"
                className="text-emerald-300 underline decoration-emerald-300/30 underline-offset-2 hover:decoration-emerald-300"
                target="_blank"
                rel="noreferrer"
              >
                GitHub repository
              </a>
            </li>
            <li>
              <a
                href="https://github.com/elvisun/newsjack/tree/main/docs/getting-started.md"
                className="text-emerald-300 underline decoration-emerald-300/30 underline-offset-2 hover:decoration-emerald-300"
                target="_blank"
                rel="noreferrer"
              >
                Getting started guide
              </a>
            </li>
            <li>
              <a
                href="https://medialyst.ai"
                className="text-emerald-300 underline decoration-emerald-300/30 underline-offset-2 hover:decoration-emerald-300"
                target="_blank"
                rel="noreferrer"
              >
                Medialyst (news API)
              </a>
            </li>
          </ul>
        </section>
      </main>

      <footer className="mx-auto flex h-16 w-full max-w-3xl items-center px-5 font-mono text-xs text-zinc-500 sm:px-8">
        MIT licensed · github.com/elvisun/newsjack
      </footer>
    </div>
  );
}
