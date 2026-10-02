import { CircleCheck, Cpu } from "lucide-react";
import Image from "next/image";

import {
  REPO_URL,
  formatDaysAgo,
  formatFull,
  loadNpmMonthlyDownloads,
  loadAllForkers,
  loadReleaseCount,
  loadRepoStats,
} from "../lib/social-proof";
import skillsData from "../lib/skills-data.json";
import { CopyCommand, INSTALL_COMMAND } from "./components/copy-command";
import { MarkGlyph } from "./components/brand";
import { Credits } from "./components/credits";
import { Footer } from "./components/footer";
import { InstallPanel } from "./components/install-panel";
import { Nav } from "./components/nav";
import { Reveal } from "./components/reveal";
import { SectionHeading } from "./components/section-heading";
import { SkillCard, type Skill } from "./components/skill-card";
import { StarButton } from "./components/star-button";
import { WireToFrontPage } from "./components/wire-to-front-page";

const skills = skillsData as Skill[];

const LANES: { id: string; tagline: string }[] = [
  { id: "Detect", tagline: "surface what matters in your space" },
  { id: "Act", tagline: "turn signal into output" },
  { id: "Strategize", tagline: "figure out what your story even is" },
  { id: "More", tagline: "everything else in the kit" },
];

const RUNTIMES = [
  "Claude Code",
  "Codex",
  "Claude.ai",
  "Cowork",
  "Hermes",
  "OpenClaw",
  "ChatGPT (limited)",
];

const TRUST = [
  { key: "license", label: "MIT licensed" },
  {
    key: "authors",
    label: (
      <>
        Built by <Credits />
      </>
    ),
  },
  { key: "sending", label: "Never sends on your behalf" },
];

const HEADLINE = "Turn your agent into a full PR team.";

function avatarSrc(url: string): string {
  const src = new URL(url);
  src.searchParams.set("s", "128");
  return src.toString();
}

function Hero({ stars }: { stars: number | null }) {
  return (
    <section className="px-6 pt-32 pb-24 sm:pt-36" id="install">
      <div className="mx-auto grid max-w-7xl gap-x-12 gap-y-20 lg:grid-cols-12">
        <div className="min-w-0 lg:col-span-7">
          <span className="nj-eyebrow">
            Vol. 01 — Open-source skills for agentic PR
          </span>
          <h1 className="mt-6 font-serif text-[clamp(2.75rem,7vw,5.25rem)] leading-[0.9] tracking-[-0.03em] italic">
            {HEADLINE.split(" ").map((word, index) => (
              <span
                className="nj-word mr-[0.25em]"
                key={index}
                style={{ animationDelay: `${index * 50}ms` }}
              >
                {word}
              </span>
            ))}
          </h1>
          <p className="mt-8 max-w-xl text-[clamp(1rem,1.2vw,1.2rem)] leading-[1.7] text-ink/80">
            {skills.length} open-source skills that teach Claude, Codex and
            friends to spot a story worth riding, fit-check reporters, and roast
            your pitch before it leaves your outbox.
          </p>

          <div className="mt-10">
            <InstallPanel />
          </div>

          <div className="mt-8 flex flex-wrap items-center gap-x-8 gap-y-6">
            <StarButton stars={stars} variant="hero" />
            <ul className="flex flex-col gap-1.5 font-mono text-[10px] tracking-[0.2em] text-ink/60 uppercase">
              {TRUST.map((item) => (
                <li className="inline-flex items-center gap-2" key={item.key}>
                  <CircleCheck
                    aria-hidden="true"
                    className="text-accent"
                    size={12}
                  />
                  <span>{item.label}</span>
                </li>
              ))}
            </ul>
          </div>
        </div>

        <div className="min-w-0 lg:col-span-5 lg:pt-28">
          <div className="relative">
            <WireToFrontPage />
            <a
              className="absolute -top-7 right-0 flex items-center gap-3 border border-ink/10 bg-page p-4 shadow-lift transition-colors hover:border-ink/25 sm:-right-4"
              href="https://medialyst.ai/agents"
              rel="noreferrer"
              target="_blank"
            >
              <span className="flex rounded-[2px] bg-ink p-2">
                <Cpu aria-hidden="true" className="text-accent" size={16} />
              </span>
              <span className="font-mono text-[10px] leading-snug tracking-[0.18em] uppercase">
                Live data by
                <br />
                <strong className="font-bold">Medialyst</strong>
              </span>
            </a>
          </div>
        </div>
      </div>
    </section>
  );
}

function RunsIn() {
  return (
    <div className="border-y border-ink/10 px-6 py-6">
      <div className="mx-auto flex max-w-7xl flex-wrap items-center gap-x-10 gap-y-3">
        <span className="nj-meta">Runs in</span>
        {RUNTIMES.map((runtime) => (
          <span
            className="font-mono text-[11px] tracking-[0.2em] text-ink/70 uppercase"
            key={runtime}
          >
            {runtime}
          </span>
        ))}
      </div>
    </div>
  );
}

async function Circulation({
  repo,
}: {
  repo: Awaited<ReturnType<typeof loadRepoStats>>;
}) {
  const [forkers, releases, npmDownloads] = await Promise.all([
    loadAllForkers(),
    loadReleaseCount(),
    loadNpmMonthlyDownloads(),
  ]);
  const lastPush = repo ? formatDaysAgo(repo.pushedAt) : null;

  const stats = [
    repo && { label: "GitHub stars", value: formatFull(repo.stars) },
    repo && { label: "Forks", value: formatFull(repo.forks) },
    npmDownloads !== null && {
      label: "npm installs, 30 days",
      value: formatFull(npmDownloads),
    },
    releases !== null && {
      label: "Releases shipped",
      value: releases >= 100 ? "100+" : String(releases),
    },
    lastPush && { label: "Last push", value: lastPush },
  ].filter((stat): stat is { label: string; value: string } => Boolean(stat));

  if (stats.length === 0 && forkers.length === 0) return null;

  return (
    <section aria-labelledby="circulation" className="px-6 py-32">
      <div className="mx-auto max-w-7xl">
        <SectionHeading
          id="circulation"
          number="01"
          subtitle="Open source, built in public, and pulled live from GitHub and npm."
          title="the circulation"
        />

        <dl className="grid grid-cols-2 gap-x-6 gap-y-12 md:grid-cols-3 lg:grid-cols-5">
          {stats.map((stat, index) => (
            <Reveal delayMs={index * 100} key={stat.label}>
              <div
                className={`border-t pt-3 ${index === 0 ? "border-accent" : "border-ink"}`}
              >
                <dt className="nj-meta">{stat.label}</dt>
                <dd className="mt-2 font-serif text-[clamp(2.25rem,4vw,3.25rem)] leading-none italic">
                  {stat.value}
                </dd>
              </div>
            </Reveal>
          ))}
        </dl>

        {forkers.length > 0 && (
          <Reveal className="mt-24">
            <div className="mb-6 flex flex-wrap items-baseline justify-between gap-4">
              <p className="font-serif text-2xl italic">
                Forked by founders, PR people and builders
              </p>
              <a
                className="nj-link text-ink/60"
                href={`${REPO_URL}/forks`}
                rel="noreferrer"
                target="_blank"
              >
                All {formatFull(forkers.length)} forks →
              </a>
            </div>
            <ul className="grid grid-cols-10 gap-1 sm:grid-cols-16 lg:grid-cols-22">
              {forkers.map((user) => (
                <li key={user.forkUrl}>
                  <a
                    className="block"
                    href={user.forkUrl}
                    rel="noreferrer"
                    target="_blank"
                    title={`@${user.login}`}
                  >
                    <Image
                      alt={`@${user.login}`}
                      className="aspect-square w-full mix-blend-multiply grayscale transition duration-300 hover:mix-blend-normal hover:grayscale-0"
                      height={64}
                      src={avatarSrc(user.avatarUrl)}
                      unoptimized
                      width={64}
                    />
                  </a>
                </li>
              ))}
            </ul>
          </Reveal>
        )}
      </div>
    </section>
  );
}

function SkillCatalog() {
  const lanes = LANES.map((lane) => ({
    ...lane,
    skills: skills.filter((skill) => skill.category === lane.id),
  })).filter((lane) => lane.skills.length > 0);

  return (
    <section
      aria-labelledby="skills"
      className="border-y border-ink/10 bg-ink/5 px-6 py-32"
    >
      <div className="mx-auto max-w-7xl">
        <SectionHeading
          id="skills"
          number="02"
          subtitle={`${skills.length} skills in ${lanes.length} lanes. Each one is a plain-Markdown file your agent reads, so you can open it, fork it, or rewrite it.`}
          title="the skill catalog"
        />

        <div className="flex flex-col gap-24">
          {lanes.map((lane) => (
            <div key={lane.id}>
              <div className="mb-8 flex flex-wrap items-baseline justify-between gap-4 border-b border-ink/10 pb-5">
                <div className="flex flex-wrap items-baseline gap-x-5 gap-y-2">
                  <span className="nj-eyebrow">{lane.id}</span>
                  <h3 className="font-serif text-[clamp(1.5rem,3vw,2rem)] leading-tight lowercase italic">
                    {lane.tagline}
                  </h3>
                </div>
                <span className="nj-meta">
                  {String(lane.skills.length).padStart(2, "0")} skills
                </span>
              </div>
              <div className="grid gap-6 md:grid-cols-2 xl:grid-cols-3">
                {lane.skills.map((skill) => (
                  <SkillCard
                    key={skill.name}
                    number={skills.indexOf(skill) + 1}
                    skill={skill}
                  />
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </section>
  );
}

function ClosingCta({ stars }: { stars: number | null }) {
  return (
    <section className="relative overflow-hidden bg-ink px-6 py-40 text-page">
      <div className="absolute inset-x-0 top-0 h-1 bg-accent" />
      <div className="relative z-10 mx-auto max-w-3xl text-center">
        <h2 className="font-serif text-[clamp(2.5rem,8vw,5rem)] leading-[0.95] italic">
          Ready to exit the spray-and-pray?
        </h2>
        <p className="mx-auto mt-7 max-w-xl font-serif text-xl leading-normal text-page/60 italic">
          One command installs all {skills.length} skills. Your agent does the
          research; every pitch still goes out under your name, sent by you.
        </p>
        <div className="mx-auto mt-12 max-w-xl text-left">
          <CopyCommand
            label="install command"
            tone="ink"
            value={INSTALL_COMMAND}
          />
        </div>
        <div className="mt-8 flex justify-center">
          <StarButton stars={stars} variant="ink" />
        </div>
        <div className="mt-12 flex flex-wrap justify-center gap-x-8 gap-y-3 text-page/40">
          <a
            className="nj-link"
            href={`${REPO_URL}/blob/main/docs/getting-started.md`}
            rel="noreferrer"
            target="_blank"
          >
            Getting started guide
          </a>
          <a
            className="nj-link"
            href={`${REPO_URL}/tree/main/skills`}
            rel="noreferrer"
            target="_blank"
          >
            Read every skill
          </a>
          <a
            className="nj-link"
            href={`${REPO_URL}/blob/main/LICENSE`}
            rel="noreferrer"
            target="_blank"
          >
            License (MIT)
          </a>
        </div>
      </div>
      <MarkGlyph className="pointer-events-none absolute -right-16 -bottom-24 w-[560px] text-page opacity-5" />
    </section>
  );
}

export default async function Home() {
  const repo = await loadRepoStats();
  const stars = repo?.stars ?? null;

  return (
    <>
      <Nav />
      <main>
        <Hero stars={stars} />
        <RunsIn />
        <Circulation repo={repo} />
        <SkillCatalog />
        <ClosingCta stars={stars} />
      </main>
      <Footer />
    </>
  );
}
