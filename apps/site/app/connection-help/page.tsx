import type { Metadata } from "next";

import { pageMetadata } from "../../lib/site";

export const metadata: Metadata = pageMetadata({
  title: "Connection help",
  description:
    "Fixes for Chrome ERR_SSL_PROTOCOL_ERROR and other connection problems with newsjack.sh.",
  path: "/connection-help",
});

export default function ConnectionHelp() {
  return (
    <main className="min-h-screen px-6 pt-24 pb-32">
      <div className="mx-auto max-w-3xl">
        <span className="nj-eyebrow">newsjack.sh</span>
        <h1 className="mt-6 font-serif text-[clamp(2.5rem,7vw,4.5rem)] leading-[0.95] tracking-[-0.03em] italic">
          Connection help
        </h1>
        <p className="mt-8 text-[17px] leading-[1.7] text-ink/80">
          If Chrome shows <code>ERR_SSL_PROTOCOL_ERROR</code>, the most common
          causes are a corporate or antivirus TLS proxy, an old browser or
          operating system, or local HSTS state left behind after a failed
          connection.
        </p>

        <section className="mt-16 space-y-14">
          <div>
            <h2 className="border-t border-ink/10 pt-8 font-serif text-[2rem] leading-[1.1] lowercase italic">
              Quick recovery steps
            </h2>
            <ol className="mt-6 list-decimal space-y-4 pl-5 text-[17px] leading-[1.7] text-ink/80 marker:font-mono marker:text-sm marker:text-accent">
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
            <h2 className="border-t border-ink/10 pt-8 font-serif text-[2rem] leading-[1.1] lowercase italic">
              Still blocked?
            </h2>
            <p className="mt-6 text-[17px] leading-[1.7] text-ink/80">
              Install from GitHub or share the error with your network
              administrator:
            </p>
            <a
              className="nj-btn-ghost mt-6"
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
