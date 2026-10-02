import { ArrowUpRight } from "lucide-react";

import { REPO_URL } from "../../lib/social-proof";
import { SectionHeading } from "./section-heading";

const ETHICS_URL = `${REPO_URL}/blob/main/skills/ETHICS.md`;

// "the standard": the file name is the hook; the doctrine lives in the repo.
export function Standard({ number }: { number: string }) {
  return (
    <section
      aria-labelledby="standard"
      className="overflow-x-clip border-b border-ink/10 bg-accent/5 px-6 py-32"
    >
      <div className="mx-auto max-w-7xl">
        <SectionHeading id="standard" number={number} title="the standard" />

        <a
          aria-label="Read ETHICS.md on GitHub"
          className="group block w-fit max-w-full"
          href={ETHICS_URL}
          rel="noreferrer"
          target="_blank"
        >
          <span className="nj-px-wordmark block font-mono text-[clamp(3.5rem,15vw,12rem)] leading-none font-bold tracking-[-0.05em] transition-colors duration-300 group-hover:text-accent">
            ETHICS
            <span className="text-ink/25 transition-colors duration-300 group-hover:text-accent/60">
              .md
            </span>
          </span>
        </a>

        <div className="mt-12 grid items-end gap-10 lg:grid-cols-12">
          <p className="font-serif text-[clamp(1.5rem,3vw,2.25rem)] leading-[1.15] italic lg:col-span-8">
            The open standard for PR agents. Every agent with newsjack installed
            loads it before doing anything else.
          </p>
          <div className="lg:col-span-4 lg:justify-self-end">
            <a
              className="nj-btn-ghost border-ink/25 px-6 py-4 text-xs"
              href={ETHICS_URL}
              rel="noreferrer"
              target="_blank"
            >
              Read ETHICS.md <ArrowUpRight aria-hidden="true" size={14} />
            </a>
          </div>
        </div>
      </div>
    </section>
  );
}
