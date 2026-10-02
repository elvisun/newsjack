import type { Metadata } from "next";
import { ArrowUpRight, CircleCheck } from "lucide-react";

import { REPO_URL } from "../../lib/social-proof";
import { Footer } from "../components/footer";
import { InstallPanel } from "../components/install-panel";
import { Nav } from "../components/nav";
import { SectionHeading } from "../components/section-heading";

export const metadata: Metadata = {
  title: "About | newsjack.sh",
  description:
    "What newsjack is, how it works, and how to install it on any agent platform.",
};

const PRINCIPLES = [
  "Open source — MIT licensed, no vendor lock-in",
  "Skills are plain Markdown, not compiled plugins",
  "Local-first — your data stays on your machine",
  "Works across agent platforms",
  "Earned media only — never sends on your behalf",
];

const LINKS = [
  { label: "GitHub repository", href: REPO_URL },
  {
    label: "Getting started guide",
    href: `${REPO_URL}/tree/main/docs/getting-started.md`,
  },
  {
    label: "Installation walkthrough (video)",
    href: "https://www.youtube.com/watch?v=1tg6E6ZYGCk",
  },
  { label: "Medialyst (news API)", href: "https://medialyst.ai/agents" },
];

const prose = "text-[17px] leading-[1.7] text-ink/80";

export default function AboutPage() {
  return (
    <>
      <Nav />
      <main className="mx-auto max-w-3xl px-6 pt-36 pb-32">
        <span className="nj-eyebrow">About</span>
        <h1 className="mt-6 font-serif text-[clamp(2.5rem,7vw,4.5rem)] leading-[0.95] tracking-[-0.03em] italic">
          Your agent, but better at PR.
        </h1>
        <p className={`mt-8 ${prose}`}>
          Newsjack is an open-source set of skills — plain-Markdown instructions
          your agent reads — plus a small CLI. Install once, and your agent can
          monitor your industry, generate story angles, fact-check pitches,
          build media lists, and more.
        </p>

        <section className="mt-24">
          <SectionHeading number="01" title="how it works" />
          <div className="space-y-5">
            <p className={prose}>
              Each skill is a Markdown file that tells your agent exactly what
              to do and how to do it — step by step, with guardrails built in.
              No API keys are required for the core set. Some skills reach for a
              live news index or journalist enrichment through the Medialyst
              API, but most run anywhere your agent does.
            </p>
            <p className={prose}>
              Skills work across platforms: Claude Code, Codex, Hermes,
              OpenClaw, Claude.ai, Cowork, and ChatGPT. Local agents get the
              full experience with scheduling and saved state. Browser agents
              get a best-effort, one-shot pass.
            </p>
          </div>
        </section>

        <section className="mt-24">
          <SectionHeading number="02" title="install" />
          <div>
            <InstallPanel />
          </div>
        </section>

        <section className="mt-24">
          <SectionHeading number="03" title="principles" />
          <ul className="space-y-4">
            {PRINCIPLES.map((item) => (
              <li className={`flex items-start gap-3 ${prose}`} key={item}>
                <CircleCheck
                  aria-hidden="true"
                  className="mt-1.5 shrink-0 text-accent"
                  size={16}
                />
                <span>{item}</span>
              </li>
            ))}
          </ul>
        </section>

        <section className="mt-24">
          <SectionHeading number="04" title="links" />
          <ul className="divide-y divide-ink/10 border-y border-ink/10">
            {LINKS.map((link) => (
              <li key={link.href}>
                <a
                  className="group flex items-center justify-between py-4 font-serif text-xl italic transition-colors hover:text-accent"
                  href={link.href}
                  rel="noreferrer"
                  target="_blank"
                >
                  {link.label}
                  <ArrowUpRight
                    aria-hidden="true"
                    className="text-ink/40 transition-colors group-hover:text-accent"
                    size={16}
                  />
                </a>
              </li>
            ))}
          </ul>
        </section>
      </main>
      <Footer />
    </>
  );
}
