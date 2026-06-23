import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Connection help | newsjack.sh",
  description:
    "Fixes for Chrome ERR_SSL_PROTOCOL_ERROR and other connection problems with newsjack.sh.",
};

export default function ConnectionHelp() {
  return (
    <main className="min-h-screen bg-zinc-950 px-5 py-12 text-zinc-100 sm:px-8">
      <div className="mx-auto max-w-3xl">
        <p className="font-mono text-xs uppercase tracking-[0.22em] text-emerald-300">
          newsjack.sh
        </p>
        <h1 className="mt-5 text-4xl font-semibold tracking-tight text-white sm:text-5xl">
          Connection help
        </h1>
        <p className="mt-5 text-lg leading-8 text-zinc-300">
          If Chrome shows <code>ERR_SSL_PROTOCOL_ERROR</code>, the most common
          causes are a corporate or antivirus TLS proxy, an old browser or
          operating system, or local HSTS state left behind after a failed
          connection.
        </p>

        <section className="mt-10 space-y-8">
          <div>
            <h2 className="text-xl font-semibold text-white">
              Quick recovery steps
            </h2>
            <ol className="mt-4 list-decimal space-y-4 pl-5 text-zinc-300">
              <li>
                Try a personal network or mobile hotspot. If that works, ask
                IT or your security vendor to allow <code>newsjack.sh</code>{" "}
                and <code>www.newsjack.sh</code>.
              </li>
              <li>
                In Chrome, open <code>chrome://net-internals/#hsts</code>.
                Under <strong>Delete domain security policies</strong>, delete{" "}
                <code>newsjack.sh</code> and <code>www.newsjack.sh</code>, then
                try the site again.
              </li>
              <li>
                Update Chrome and the operating system. The site requires TLS
                1.2 or newer, which very old clients do not support by default.
              </li>
            </ol>
          </div>

          <div>
            <h2 className="text-xl font-semibold text-white">
              Still blocked?
            </h2>
            <p className="mt-4 text-zinc-300">
              Install from GitHub or share the error with your network
              administrator:
            </p>
            <a
              className="mt-5 inline-flex rounded-md border border-white/15 bg-white/[0.04] px-4 py-2.5 font-mono text-sm text-zinc-100 transition hover:border-emerald-300/50 hover:bg-emerald-300/10 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-4 focus-visible:outline-emerald-300"
              href="https://github.com/elvisun/newsjack"
              rel="noreferrer"
            >
              github.com/elvisun/newsjack
            </a>
          </div>
        </section>
      </div>
    </main>
  );
}
