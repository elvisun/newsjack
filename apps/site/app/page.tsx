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

export default function Home() {
  return (
    <div className="min-h-screen overflow-hidden bg-[radial-gradient(circle_at_top_left,rgba(34,197,94,0.16),transparent_28rem),linear-gradient(135deg,#090b0f_0%,#101217_54%,#050607_100%)] text-zinc-50">
      <main className="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-6xl items-center px-5 py-12 sm:px-8 lg:px-10">
        <section
          aria-labelledby="hero-title"
          className="grid w-full gap-10 lg:grid-cols-[minmax(0,1fr)_24rem] lg:items-center"
        >
          <div className="max-w-3xl">
            <p className="mb-5 font-mono text-xs uppercase tracking-[0.22em] text-emerald-300">
              newsjack.sh
            </p>
            <h1
              id="hero-title"
              className="max-w-4xl text-balance text-5xl font-semibold leading-[0.95] tracking-tight text-white sm:text-6xl lg:text-7xl"
            >
              Open-source operating system for agentic PR.
            </h1>
            <p className="mt-7 max-w-2xl text-pretty text-lg leading-8 text-zinc-300 sm:text-xl">
              The skill layer that turns Claude, ChatGPT, and Cursor into PR
              operators.
            </p>

            <div className="mt-10 max-w-xl">
              <pre
                aria-label="Install command"
                className="overflow-x-auto rounded-lg border border-white/10 bg-black/60 p-4 font-mono text-sm leading-6 text-emerald-200 shadow-2xl shadow-black/30 sm:text-base"
              >
                <code>curl newsjack.sh | sh</code>
              </pre>
              <p className="mt-3 font-mono text-xs text-zinc-500">
                Installs latest from GitHub main
              </p>
            </div>

            <a
              className="mt-9 inline-flex items-center gap-2 rounded-md border border-white/15 bg-white/[0.04] px-4 py-2.5 font-mono text-sm text-zinc-100 transition hover:border-emerald-300/50 hover:bg-emerald-300/10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-emerald-300"
              href="https://github.com/elvisun/newsjack"
              rel="noreferrer"
              target="_blank"
            >
              <GitHubIcon />
              <span>Star on GitHub</span>
            </a>
          </div>

          <div
            aria-hidden="true"
            className="hidden rounded-lg border border-white/10 bg-zinc-950/70 p-4 font-mono text-xs text-zinc-400 shadow-2xl shadow-black/30 lg:block"
          >
            <div className="mb-4 flex items-center gap-2 border-b border-white/10 pb-3">
              <span className="h-2.5 w-2.5 rounded-full bg-red-400" />
              <span className="h-2.5 w-2.5 rounded-full bg-amber-300" />
              <span className="h-2.5 w-2.5 rounded-full bg-emerald-400" />
            </div>
            <div className="space-y-2">
              <p>
                <span className="text-emerald-300">$</span> newsjack install
              </p>
              <p className="text-zinc-500">resolving agent runtime...</p>
              <p className="text-zinc-500">syncing pr operator skills...</p>
              <p className="text-zinc-500">configuring optional mcp...</p>
              <p className="text-zinc-100">ready</p>
            </div>
          </div>
        </section>
      </main>

      <footer className="mx-auto flex h-16 w-full max-w-6xl items-center px-5 font-mono text-xs text-zinc-500 sm:px-8 lg:px-10">
        MIT licensed · github.com/elvisun/newsjack
      </footer>
    </div>
  );
}
