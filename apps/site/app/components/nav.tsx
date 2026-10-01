import Link from "next/link";

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

export function Nav() {
  return (
    <nav className="mx-auto flex w-full max-w-6xl items-center justify-between px-5 py-6 sm:px-8 lg:px-10">
      <Link
        href="/"
        className="font-mono text-sm text-zinc-400 transition hover:text-white"
      >
        newsjack.sh
      </Link>
      <div className="flex items-center gap-5">
        <Link
          href="/about"
          className="font-mono text-sm text-zinc-400 transition hover:text-white"
        >
          About
        </Link>
        <Link
          href="/insights"
          className="font-mono text-sm text-zinc-400 transition hover:text-white"
        >
          Insights
        </Link>
        <a
          href="https://github.com/elvisun/newsjack"
          rel="noreferrer"
          target="_blank"
          className="text-zinc-500 transition hover:text-white"
          aria-label="GitHub"
        >
          <GitHubIcon />
        </a>
      </div>
    </nav>
  );
}
