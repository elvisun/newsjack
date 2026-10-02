import { ArrowRight, ArrowUpRight, SquareTerminal } from "lucide-react";

import { REPO_URL } from "../../lib/social-proof";

export interface Skill {
  name: string;
  title: string;
  description: string;
  whenToUse: string;
  category: string;
}

// Square editorial card; hovering (or focusing) slides up an ink panel with
// the invocation and when-to-use guidance.
export function SkillCard({ skill, number }: { skill: Skill; number: number }) {
  return (
    <a
      className="group relative flex min-h-[320px] flex-col overflow-hidden border border-ink/10 bg-white p-8 shadow-editorial focus-visible:outline-offset-0 sm:aspect-square sm:min-h-0"
      href={`${REPO_URL}/blob/main/skills/${skill.name}/SKILL.md`}
      rel="noreferrer"
      target="_blank"
    >
      <div className="relative transition-opacity duration-200 group-hover:opacity-0 group-focus-visible:opacity-0">
        <div className="mb-5 flex items-start justify-between">
          <SquareTerminal aria-hidden="true" className="text-accent" size={20} />
          <span className="nj-meta">No. {String(number).padStart(2, "0")}</span>
        </div>
        <h4 className="font-serif text-[26px] leading-[1.15] italic">
          {skill.title}
        </h4>
        <p className="mt-3 line-clamp-5 text-sm leading-relaxed text-ink/60">
          {skill.description}
        </p>
      </div>
      <span className="nj-link relative mt-auto pt-6 text-ink/60 transition-opacity duration-200 group-hover:opacity-0 group-focus-visible:opacity-0">
        {skill.name} <ArrowRight aria-hidden="true" size={12} />
      </span>

      <div className="absolute inset-0 flex translate-y-full flex-col justify-center bg-ink p-8 text-page transition-transform duration-500 ease-editorial group-hover:translate-y-0 group-focus-visible:translate-y-0">
        <span className="nj-meta mb-2 text-page/40">Invoke</span>
        <span className="overflow-x-auto whitespace-nowrap rounded-[2px] border border-page/20 bg-black/40 p-3 font-mono text-[13px]">
          <span className="text-accent">➜</span> /{skill.name}
        </span>
        {skill.whenToUse && (
          <>
            <span className="nj-meta mt-6 mb-2 text-page/40">Use it when</span>
            <span className="line-clamp-5 text-sm leading-relaxed text-page/70">
              {skill.whenToUse}
            </span>
          </>
        )}
        <span className="nj-link mt-6 text-page/60">
          Read SKILL.md <ArrowUpRight aria-hidden="true" size={12} />
        </span>
      </div>
    </a>
  );
}
