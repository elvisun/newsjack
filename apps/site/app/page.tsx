import { ArrowUpRight, CircleCheck, Play } from "lucide-react";
import type { Metadata } from "next";
import Image from "next/image";

import {
  REPO_URL,
  formatDaysAgo,
  formatFull,
  loadNpmTotalDownloads,
  NPM_FIRST_PUBLISHED,
  loadAllForkers,
  loadInstallCount,
  loadReleaseCount,
  loadRepoStats,
} from "../lib/social-proof";
import {
  SITE_DESCRIPTION,
  SITE_NAME,
  SITE_URL,
  absoluteUrl,
  pageMetadata,
} from "../lib/site";
import installData from "../lib/install-data.json";
import skillsData from "../lib/skills-data.json";
import { Clippings } from "./components/clippings";
import { CopyCommand, InlineCopy } from "./components/copy-command";
import { MarkGlyph } from "./components/brand";
import { Credits } from "./components/credits";
import { Footer } from "./components/footer";
import { InstallPanel } from "./components/install-panel";
import { JsonLd } from "./components/json-ld";
import { Nav } from "./components/nav";
import { Reveal } from "./components/reveal";
import { SectionHeading } from "./components/section-heading";
import { SkillCard, type Skill } from "./components/skill-card";
import { Standard } from "./components/standard";
import { StarButton } from "./components/star-button";

const skills = skillsData as Skill[];

const LANES: { id: string; tagline: string }[] = [
  { id: "Detect", tagline: "surface what matters in your space" },
  { id: "Act", tagline: "turn signal into output" },
  { id: "Strategize", tagline: "figure out what your story even is" },
  { id: "AI visibility", tagline: "get your facts into AI answers" },
  { id: "More", tagline: "everything else in the kit" },
];

// Brand marks from each product's own site, brand page, or Wikipedia infobox.
const WORKS_WITH = [
  { name: "ChatGPT", logo: "/logos/openai.svg" },
  { name: "Claude", logo: "/logos/claude.svg" },
  { name: "Grok", logo: "/logos/grok.svg" },
  { name: "Gemini", logo: "/logos/gemini.png" },
  { name: "Muse", logo: "/logos/muse.svg" },
  { name: "Claude Code", logo: "/logos/claude-code.svg" },
  { name: "Codex", logo: "/logos/openai.svg" },
  { name: "Cursor", logo: "/logos/cursor.svg" },
  { name: "OpenClaw", logo: "/logos/openclaw.svg" },
  { name: "Hermes", logo: "/logos/hermes-agent.png" },
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

// Trial order: catalog first, circulation last. Section numbers follow this
// list, so reverting is just reordering it (previously circulation,
// clippings, standard, skills).
const SECTION_ORDER = [
  "skills",
  "standard",
  "clippings",
  "circulation",
] as const;

const WALKTHROUGH_URL = "https://www.youtube.com/watch?v=1tg6E6ZYGCk";

function avatarSrc(url: string): string {
  const src = new URL(url);
  src.searchParams.set("s", "128");
  return src.toString();
}

function Hero({ stars }: { stars: number | null }) {
  return (
    <section className="px-6 pt-32 pb-24 sm:pt-36" id="install">
      <div className="mx-auto grid max-w-7xl gap-x-16 gap-y-12 lg:grid-cols-12">
        <div className="min-w-0 lg:col-span-6 lg:row-start-1 lg:self-end">
          <span className="nj-eyebrow">
            Vol. 01 — Open-source skills for agentic PR
          </span>
          <h1 className="mt-6 font-serif text-[clamp(2.75rem,6.5vw,5.25rem)] leading-[0.9] tracking-[-0.03em] italic">
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
            Paste one prompt and your AI picks up {skills.length} open-source PR
            skills. It watches the news for stories you can ride, finds the
            angle, checks which reporters actually cover it, and roasts your
            pitch before you hit send.
          </p>
        </div>

        <div className="min-w-0 lg:col-span-6 lg:col-start-7 lg:row-span-2 lg:row-start-1 lg:self-center">
          <div className="nj-rise">
            <InstallPanel
              aiPrompt={installData.aiPrompt}
              skillCount={skills.length}
              terminal={installData.terminal}
            />
          </div>
          <a
            className="nj-link mt-5 text-ink/50"
            href={WALKTHROUGH_URL}
            rel="noreferrer"
            target="_blank"
          >
            <Play aria-hidden="true" size={12} />
            Prefer video? Watch the install walkthrough · 26:47
            <ArrowUpRight aria-hidden="true" size={12} />
          </a>
        </div>

        <div className="flex min-w-0 flex-wrap items-center gap-x-8 gap-y-6 lg:col-span-6 lg:row-start-2 lg:self-start">
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
    </section>
  );
}

function RunsIn() {
  return (
    <div className="border-y border-ink/10 px-6 py-6">
      <div className="mx-auto flex max-w-7xl flex-wrap items-center gap-x-6 gap-y-4 xl:flex-nowrap xl:justify-between">
        <span className="nj-meta shrink-0">Works with</span>
        {WORKS_WITH.map((item) => (
          <span className="flex shrink-0 items-center gap-2" key={item.name}>
            <Image
              alt=""
              className="size-5 object-contain"
              height={20}
              src={item.logo}
              unoptimized
              width={20}
            />
            <span className="text-sm font-medium text-ink/80">{item.name}</span>
          </span>
        ))}
      </div>
    </div>
  );
}

async function Circulation({
  number,
  repo,
}: {
  number: string;
  repo: Awaited<ReturnType<typeof loadRepoStats>>;
}) {
  const [forkers, releases, npmDownloads, installs] = await Promise.all([
    loadAllForkers(),
    loadReleaseCount(),
    loadNpmTotalDownloads(),
    loadInstallCount(),
  ]);
  const lastPush = repo ? formatDaysAgo(repo.pushedAt) : null;
  const npmSince = new Date(NPM_FIRST_PUBLISHED).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    timeZone: "UTC",
  });

  // Separate channels: install.sh never touches npm, so the two add up.
  const channels = [
    installs !== null && {
      label: "Direct installs",
      detail:
        "curl, wget and PowerShell runs of the installer at newsjack.sh, once per IP per day",
      value: installs,
    },
    npmDownloads !== null && {
      label: "npm downloads",
      detail: `The newsjack package since ${npmSince}. Includes updates and CI.`,
      value: npmDownloads,
    },
  ].filter(
    (channel): channel is { label: string; detail: string; value: number } =>
      Boolean(channel),
  );
  const total = channels.reduce((sum, channel) => sum + channel.value, 0);

  const stats = [
    repo && { label: "GitHub stars", value: formatFull(repo.stars) },
    repo && { label: "Forks", value: formatFull(repo.forks) },
    releases !== null && {
      label: "Releases shipped",
      value: releases >= 100 ? "100+" : String(releases),
    },
    lastPush && { label: "Last push", value: lastPush },
  ].filter((stat): stat is { label: string; value: string } => Boolean(stat));

  if (channels.length === 0 && stats.length === 0 && forkers.length === 0) {
    return null;
  }

  return (
    <section aria-labelledby="circulation" className="px-6 py-32">
      <div className="mx-auto max-w-7xl">
        <SectionHeading
          id="circulation"
          number={number}
          subtitle="Open source, built in public. Every number here is live."
          title="the circulation"
        />

        {channels.length > 0 && (
          <Reveal>
            <div className="grid gap-10 border-t border-accent pt-6 lg:grid-cols-12 lg:items-end">
              <div className="lg:col-span-7">
                <p className="nj-meta">
                  {channels.length > 1
                    ? "Installs + downloads, all time"
                    : `${channels[0].label}, all time`}
                </p>
                <p className="mt-3 font-serif text-[clamp(4rem,11vw,9rem)] leading-[0.9] tracking-[-0.03em] italic">
                  {formatFull(total)}
                </p>
              </div>
              <dl className="divide-y divide-ink/10 border-y border-ink/10 lg:col-span-5">
                {channels.map((channel) => (
                  <div
                    className="flex items-baseline justify-between gap-6 py-4"
                    key={channel.label}
                  >
                    <dt>
                      <span className="block font-medium">{channel.label}</span>
                      <span className="mt-0.5 block text-[13px] leading-snug text-ink/50">
                        {channel.detail}
                      </span>
                    </dt>
                    <dd className="shrink-0 font-serif text-3xl italic">
                      {formatFull(channel.value)}
                    </dd>
                  </div>
                ))}
              </dl>
            </div>
            <p className="mt-6 text-[13px] text-ink/40 italic">
              Direct installs from the Claude marketplace or the ChatGPT
              marketplace aren&apos;t tracked yet.
            </p>
          </Reveal>
        )}

        <dl className="mt-20 grid grid-cols-2 gap-x-6 gap-y-12 md:grid-cols-4">
          {stats.map((stat, index) => (
            <Reveal delayMs={index * 100} key={stat.label}>
              <div className="border-t border-ink pt-3">
                <dt className="nj-meta">{stat.label}</dt>
                <dd className="mt-2 font-serif text-[clamp(2rem,3.5vw,2.75rem)] leading-none italic">
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

function SkillCatalog({ number }: { number: string }) {
  const lanes = LANES.map((lane) => ({
    ...lane,
    skills: skills.filter((skill) => skill.category === lane.id),
  })).filter((lane) => lane.skills.length > 0);

  return (
    <section
      aria-labelledby="skills"
      className="border-b border-ink/10 bg-ink/5 px-6 py-32"
    >
      <div className="mx-auto max-w-7xl">
        <SectionHeading
          id="skills"
          number={number}
          subtitle={`${skills.length} skills in ${lanes.length} lanes. Each one is a plain-Markdown file your agent reads, so you can open it, fork it, or rewrite it.`}
          title="the skill catalog"
        />

        <div className="flex flex-col gap-16">
          {lanes.map((lane) => (
            <div key={lane.id}>
              <div className="mb-6 flex flex-wrap items-baseline justify-between gap-4 border-b border-ink/10 pb-4">
                <div className="flex flex-wrap items-baseline gap-x-5 gap-y-2">
                  <span className="nj-eyebrow text-[13px] font-bold tracking-[0.22em]">
                    {lane.id}
                  </span>
                  <h3 className="font-serif text-[clamp(1.5rem,3vw,2rem)] leading-tight lowercase italic">
                    {lane.tagline}
                  </h3>
                </div>
                <span className="nj-meta">
                  {String(lane.skills.length).padStart(2, "0")} skills
                </span>
              </div>
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
                {lane.skills.map((skill) => (
                  <SkillCard key={skill.name} skill={skill} />
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
          One prompt installs all {skills.length} skills. Your agent does the
          research; every pitch still goes out under your name, sent by you.
        </p>
        <div className="mx-auto mt-12 max-w-xl text-left">
          <CopyCommand
            label="setup prompt"
            prompt="➜"
            tone="ink"
            value={installData.aiPrompt}
            wrap
          />
          <p className="mt-4 text-center font-mono text-xs text-page/40">
            Prefer the terminal? <InlineCopy value={installData.terminal} />
          </p>
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

export const metadata: Metadata = pageMetadata({
  description: SITE_DESCRIPTION,
  path: "/",
});

const HOME_JSON_LD = {
  "@context": "https://schema.org",
  "@graph": [
    {
      "@type": "WebSite",
      name: SITE_NAME,
      url: SITE_URL,
      description: SITE_DESCRIPTION,
    },
    {
      "@type": "SoftwareApplication",
      name: "newsjack",
      url: SITE_URL,
      description: SITE_DESCRIPTION,
      applicationCategory: "BusinessApplication",
      operatingSystem: "macOS, Linux, Windows",
      installUrl: absoluteUrl("/install.sh"),
      license: "https://opensource.org/licenses/MIT",
      isAccessibleForFree: true,
      offers: { "@type": "Offer", price: "0", priceCurrency: "USD" },
      sameAs: [REPO_URL],
    },
  ],
};

export default async function Home() {
  const repo = await loadRepoStats();
  const stars = repo?.stars ?? null;

  return (
    <>
      <JsonLd data={HOME_JSON_LD} />
      <Nav />
      <main>
        <Hero stars={stars} />
        <RunsIn />
        {SECTION_ORDER.map((section, index) => {
          const number = String(index + 1).padStart(2, "0");
          if (section === "skills")
            return <SkillCatalog key={section} number={number} />;
          if (section === "standard")
            return <Standard key={section} number={number} />;
          if (section === "clippings")
            return <Clippings key={section} number={number} />;
          return <Circulation key={section} number={number} repo={repo} />;
        })}
        <ClosingCta stars={stars} />
      </main>
      <Footer />
    </>
  );
}
