import { Wordmark } from "./brand";
import { Credits } from "./credits";

export function Footer() {
  return (
    <footer className="border-t border-ink/10 px-6 py-12">
      <div className="mx-auto flex max-w-7xl flex-wrap items-center justify-between gap-6">
        <Wordmark small />
        <div className="sm:text-right">
          <p className="nj-meta">
            An open-source project by <Credits />
          </p>
          <p className="nj-meta mt-1.5 text-ink/20">
            © {new Date().getFullYear()} newsjack.sh — MIT licensed
          </p>
        </div>
      </div>
    </footer>
  );
}
