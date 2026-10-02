import { ArrowUpRight } from "lucide-react";

import { REPO_URL } from "../../lib/social-proof";
import { CopyChip } from "./copy-command";

export interface Skill {
  name: string;
  title: string;
  description: string;
  whenToUse: string;
  category: string;
}

// Compact card: title, short description, and a copyable slash command. Only
// the SKILL.md link leaves the page.
export function SkillCard({ skill }: { skill: Skill }) {
  return (
    <article className="flex flex-col border border-ink/10 bg-white p-5 shadow-editorial transition-colors hover:border-ink/25">
      <div className="flex items-start justify-between gap-3">
        <h4 className="font-serif text-[19px] leading-[1.2] italic">
          {skill.title}
        </h4>
        <a
          aria-label={`Read the ${skill.title} SKILL.md on GitHub`}
          className="nj-link shrink-0 pt-1 text-ink/50"
          href={`${REPO_URL}/blob/main/skills/${skill.name}/SKILL.md`}
          rel="noreferrer"
          target="_blank"
        >
          SKILL.md <ArrowUpRight aria-hidden="true" size={12} />
        </a>
      </div>
      <p className="mt-2 line-clamp-3 text-[13px] leading-relaxed text-ink/60">
        {skill.description}
      </p>
      <div className="mt-auto pt-4">
        <CopyChip value={`/${skill.name}`} />
      </div>
    </article>
  );
}
