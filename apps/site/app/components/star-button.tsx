import { Star } from "lucide-react";

import { REPO_URL, formatCompact, formatFull } from "../../lib/social-proof";
import { GitHubMark } from "./brand";

type Variant = "nav" | "hero" | "ink";

const variantClass: Record<Variant, string> = {
  nav: "border-ink/10 px-3.5 py-[7px] text-[10px] hover:border-ink hover:bg-ink hover:text-page",
  hero: "border-ink/25 px-6 py-4 text-xs hover:border-ink hover:bg-ink hover:text-page",
  ink: "border-page/20 px-6 py-4 text-xs text-page hover:border-page hover:bg-page hover:text-ink",
};

export function StarButton({
  stars,
  variant = "hero",
}: {
  stars: number | null;
  variant?: Variant;
}) {
  const label = variant === "nav" ? "GitHub" : "Star on GitHub";

  return (
    <a
      aria-label={
        stars === null ? label : `${label}, ${formatFull(stars)} stars`
      }
      className={`group inline-flex items-center gap-3 border font-mono uppercase tracking-[0.2em] transition-colors duration-300 ease-editorial ${variantClass[variant]}`}
      href={REPO_URL}
      rel="noreferrer"
      target="_blank"
    >
      <span className="inline-flex items-center gap-2">
        <GitHubMark size={variant === "nav" ? 12 : 14} />
        {label}
      </span>
      {stars !== null && (
        <span className="inline-flex items-center gap-1.5 border-l border-current/20 pl-3">
          <Star aria-hidden="true" className="text-accent" size={12} />
          {variant === "nav" ? formatCompact(stars) : formatFull(stars)}
        </span>
      )}
    </a>
  );
}
