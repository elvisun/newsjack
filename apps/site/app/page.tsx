import Link from "next/link";

import skillsData from "../lib/skills-data.json";
import { Nav } from "./components/nav";

interface Skill {
  name: string;
  description: string;
  whenToUse: string;
  category: string;
}

const CATEGORY_META: Record<string, { emoji: string; tagline: string }> = {
  Strategize: {
    emoji: "🧭",
    tagline: "Figure out what your story even is",
  },
  Act: {
    emoji: "🚀",
    tagline: "Turn signal into output",
  },
  Detect: {
    emoji: "🛰️",
    tagline: "Surface what matters in your space",
  },
};

const CATEGORY_ORDER = ["Detect", "Act", "Strategize"];

function GitHubIcon() {
  return (
    <svg
      aria-hidden="true"
      className="h-4 w-4"
      fill="currentColor"
      viewBox="0 0 24 24"
    >
      <path d="M12 .5C5.65.5.5 5.65.5 12c0 5.09 3.29 9.4 7.86 10.93.58.1.79-.25.79-.56v-2.02c-3.2.7-3.87-1.37-3.87-1.37-.53-1.33-1.29-1.69-1.29-1.69-1.05-.72.08-.7.08-.7 1.16.08 1.77 1.2 1.77 1.2 1.03 1.76 2.7 1.25 3.36.96.1-.75.4-1.25.73-1.54-2.55-.29-5.24-1.28-5.24-5.68 0-1.25.45-2.28 1.19-3.08-.12-.29-.52-1.46.11-3.04 0 0 .98-.31 3.18 1.18A11.1 11.1 0 0 1 12 6.2c.98 0 1.96.13 2.88.39 2.2-1.49 3.17-1.18 3.17-1.18.64 1.58.24 2.75.12 3.04.74.8 1.18 1.83 1.18 3.08 0 4.42-2.69 5.39-5.25 5.67.42.36.78 1.06.78 2.14v3.03c0 .31.21.67.8.56A11.52 11.52 0 0 0 23.5 12C23.5 5.65 18.35.5 12 .5Z" />
    </svg>
  );
}

function SkillCard({ skill }: { skill: Skill }) {
  return (
    <a
      href={`https://github.com/elvisun/newsjack/tree/main/skills/${skill.name}`}
      target="_blank"
      rel="noreferrer"
      className="group block rounded-lg border border-white/[0.06] bg-white/[0.02] p-5 transition hover:border-emerald-300/30 hover:bg-emerald-300/[0.04]"
    >
      <h3 className="font-mono text-sm font-semibold text-white group-hover:text-emerald-200">
        {skill.name}
      </h3>
      <p className="mt-2 text-sm leading-6 text-zinc-400">{skill.description}</p>
    </a>
  );
}

function grouped(skills: Skill[]): Record<string, Skill[]> {
  const groups: Record<string, Skill[]> = {};
  for (const s of skills) {
    (groups[s.category] ??= []).push(s);
  }
  return groups;
}

export default function Home() {
  const groups = grouped(skillsData as Skill[]);

  return (
    <div className="min-h-screen bg-[radial-gradient(circle_at_top_left,rgba(34,197,94,0.16),transparent_28rem),linear-gradient(135deg,#090b0f_0%,#101217_54%,#050607_100%)] text-zinc-50">
      <Nav />

      <main className="mx-auto w-full max-w-6xl px-5 pb-20 sm:px-8 lg:px-10">
        {/* Hero */}
        <section className="pb-20 pt-12 sm:pt-20">
          <p className="font-mono text-xs uppercase tracking-[0.22em] text-emerald-300">
            newsjack.sh
          </p>
          <h1 className="mt-5 max-w-4xl text-balance text-5xl font-semibold leading-[0.95] tracking-tight text-white sm:text-6xl lg:text-7xl">
            Open-source operating system for agentic PR.
          </h1>
          <p className="mt-7 max-w-2xl text-pretty text-lg leading-8 text-zinc-300 sm:text-xl">
            Local-first skills and CLI workflows for Claude Code, Codex,
            OpenClaw, and Hermes. Install once — your agent becomes a PR team.
          </p>

          <div className="mt-10 max-w-xl">
            <pre
              aria-label="Install command"
              className="overflow-x-auto rounded-lg border border-white/10 bg-black/60 p-4 font-mono text-sm leading-6 text-emerald-200 shadow-2xl shadow-black/30 sm:text-base"
            >
              <code>curl -fsSL newsjack.sh | bash</code>
            </pre>
            <p className="mt-3 font-mono text-xs text-zinc-500">
              Installs latest from GitHub Releases
            </p>
          </div>

          <div className="mt-9 flex flex-wrap items-center gap-4">
            <a
              className="inline-flex items-center gap-2 rounded-md border border-white/15 bg-white/[0.04] px-4 py-2.5 font-mono text-sm text-zinc-100 transition hover:border-emerald-300/50 hover:bg-emerald-300/10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-emerald-300"
              href="https://github.com/elvisun/newsjack"
              rel="noreferrer"
              target="_blank"
            >
              <GitHubIcon />
              <span>Star on GitHub</span>
            </a>
            <Link
              className="inline-flex items-center rounded-md border border-white/15 bg-white/[0.04] px-4 py-2.5 font-mono text-sm text-zinc-100 transition hover:border-emerald-300/50 hover:bg-emerald-300/10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-emerald-300"
              href="/about"
            >
              Learn more
            </Link>
          </div>
        </section>

        {/* Skills */}
        <section>
          <h2 className="text-3xl font-semibold tracking-tight text-white sm:text-4xl">
            {(skillsData as Skill[]).length} skills, three lanes
          </h2>
          <p className="mt-4 max-w-2xl text-lg text-zinc-400">
            Every skill is a plain-Markdown file your agent reads. No vendor lock-in, no API keys required for the core set.
          </p>

          <div className="mt-14 space-y-16">
            {CATEGORY_ORDER.map((cat) => {
              const meta = CATEGORY_META[cat];
              const skills = groups[cat] || [];
              return (
                <div key={cat}>
                  <div className="mb-6 flex items-center gap-3">
                    <span className="text-2xl" aria-hidden="true">
                      {meta.emoji}
                    </span>
                    <div>
                      <h3 className="text-xl font-semibold text-white">
                        {cat}
                      </h3>
                      <p className="text-sm text-zinc-400">{meta.tagline}</p>
                    </div>
                  </div>
                  <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
                    {skills.map((s) => (
                      <SkillCard key={s.name} skill={s} />
                    ))}
                  </div>
                </div>
              );
            })}
          </div>
        </section>

        {/* Who this is for */}
        <section className="mt-24">
          <h2 className="text-3xl font-semibold tracking-tight text-white sm:text-4xl">
            Who this is for
          </h2>
          <div className="mt-8 grid gap-6 sm:grid-cols-2">
            {[
              {
                title: "Founders",
                desc: "Doing their own PR because the agency quote was insane.",
              },
              {
                title: "PR agencies",
                desc: "Running more accounts than humans can babysit.",
              },
              {
                title: "Marketers",
                desc: "At small companies who need leverage, not headcount.",
              },
              {
                title: "Anyone",
                desc: "Whose agent is already running their day-to-day — and should be better at it.",
              },
            ].map((item) => (
              <div
                key={item.title}
                className="rounded-lg border border-white/[0.06] bg-white/[0.02] p-5"
              >
                <h3 className="font-semibold text-white">{item.title}</h3>
                <p className="mt-1 text-sm text-zinc-400">{item.desc}</p>
              </div>
            ))}
          </div>
        </section>
      </main>

      <footer className="mx-auto flex h-16 w-full max-w-6xl items-center px-5 font-mono text-xs text-zinc-500 sm:px-8 lg:px-10">
        MIT licensed · github.com/elvisun/newsjack
      </footer>
    </div>
  );
}
