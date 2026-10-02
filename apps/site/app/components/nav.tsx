import Link from "next/link";

import { loadStarCount } from "../../lib/social-proof";
import { Wordmark } from "./brand";
import { StarButton } from "./star-button";

export async function Nav() {
  const stars = await loadStarCount();

  return (
    <nav className="fixed inset-x-0 top-0 z-50 border-b border-ink/10 bg-page/80 backdrop-blur-md">
      <div className="mx-auto flex max-w-7xl items-center justify-between gap-6 px-6 py-4">
        <Wordmark />
        <div className="flex items-center gap-6 sm:gap-8">
          <Link className="nj-link hidden md:inline-flex" href="/#install">
            Install
          </Link>
          <Link className="nj-link hidden md:inline-flex" href="/#skills">
            Skills
          </Link>
          <Link className="nj-link hidden sm:inline-flex" href="/insights">
            Insights
          </Link>
          <Link className="nj-link hidden sm:inline-flex" href="/about">
            About
          </Link>
          <StarButton stars={stars} variant="nav" />
        </div>
      </div>
    </nav>
  );
}
