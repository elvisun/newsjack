import type { Metadata } from "next";

import { CHATGPT_INSTALL_PROMPT } from "../../lib/chatgpt-install";
import { pageMetadata } from "../../lib/site";
import { CopyCommand } from "../components/copy-command";

export const metadata: Metadata = pageMetadata({
  title: "Install Newsjack in ChatGPT",
  description:
    "Install the Newsjack plugin in the ChatGPT desktop app with one prompt.",
  path: "/chatgpt",
});

export default function ChatGPTInstall() {
  return (
    <main className="min-h-screen px-6 pt-24 pb-32">
      <div className="mx-auto max-w-3xl">
        <span className="nj-eyebrow">newsjack.sh</span>
        <h1 className="mt-6 font-serif text-[clamp(2.5rem,7vw,4.5rem)] leading-[0.95] tracking-[-0.03em] italic">
          Install Newsjack in ChatGPT
        </h1>
        <p className="mt-8 text-[17px] leading-[1.7] text-ink/80">
          One prompt installs Newsjack in the ChatGPT desktop app.
        </p>

        <ol className="mt-16 space-y-14">
          <li>
            <p className="font-mono text-xs tracking-[0.15em] text-accent uppercase">
              Step 1
            </p>
            <h2 className="mt-3 border-t border-ink/10 pt-8 font-serif text-[2rem] leading-[1.1] italic">
              Open ChatGPT in Work mode
            </h2>
            <p className="mt-6 text-[17px] leading-[1.7] text-ink/80">
              Open the{" "}
              <a
                className="underline decoration-accent/60 underline-offset-2 transition-colors hover:text-accent"
                href="https://chatgpt.com/download"
                rel="noreferrer"
              >
                ChatGPT desktop app
              </a>{" "}
              on Mac or Windows. Switch to <strong>Work mode</strong>. The
              ChatGPT website and phone apps cannot install Newsjack.
            </p>
          </li>

          <li>
            <p className="font-mono text-xs tracking-[0.15em] text-accent uppercase">
              Step 2
            </p>
            <h2 className="mt-3 border-t border-ink/10 pt-8 font-serif text-[2rem] leading-[1.1] italic">
              Copy the prompt and send it
            </h2>
            <div className="mt-6">
              <CopyCommand
                label="ChatGPT install prompt"
                prompt={null}
                value={CHATGPT_INSTALL_PROMPT}
                wrap
              />
            </div>
          </li>

          <li>
            <p className="font-mono text-xs tracking-[0.15em] text-accent uppercase">
              Step 3
            </p>
            <h2 className="mt-3 border-t border-ink/10 pt-8 font-serif text-[2rem] leading-[1.1] italic">
              Approve twice and sign in
            </h2>
            <p className="mt-6 text-[17px] leading-[1.7] text-ink/80">
              ChatGPT asks for permission twice: first to download from GitHub,
              then to open the Medialyst sign-in. Approve both. Sign in to
              Medialyst in the browser tab that opens. A free account is fine.
            </p>
          </li>

          <li>
            <p className="font-mono text-xs tracking-[0.15em] text-accent uppercase">
              Step 4
            </p>
            <h2 className="mt-3 border-t border-ink/10 pt-8 font-serif text-[2rem] leading-[1.1] italic">
              Restart ChatGPT
            </h2>
            <p className="mt-6 text-[17px] leading-[1.7] text-ink/80">
              Quit and reopen ChatGPT. Start a new chat and type{" "}
              <code>@Newsjack</code>. For example: “@Newsjack find newsjacking
              opportunities for my company”.
            </p>
          </li>
        </ol>

        <section className="mt-20">
          <h2 className="border-t border-ink/10 pt-8 font-serif text-[2rem] leading-[1.1] italic">
            If it does not work
          </h2>
          <ul className="mt-6 list-disc space-y-4 pl-5 text-[17px] leading-[1.7] text-ink/80 marker:text-accent">
            <li>
              If ChatGPT says Newsjack is not in the plugin directory or that
              it cannot run commands, you are in normal chat. Switch to Work
              mode and send the prompt again.
            </li>
            <li>
              Company-managed ChatGPT accounts may block plugins from outside
              the directory. Use a personal account or ask your administrator.
            </li>
            <li>
              If the Medialyst sign-in times out, ask ChatGPT to “run{" "}
              <code>codex mcp login medialyst</code> again”.
            </li>
            <li>
              To update Newsjack later, ask ChatGPT to run{" "}
              <code>codex plugin marketplace upgrade newsjack</code>.
            </li>
          </ul>
          <p className="mt-8 text-[15px] leading-[1.7] text-ink/60">
            We tested these steps on macOS. Windows uses the same steps, but we
            have not verified it yet.
          </p>
        </section>
      </div>
    </main>
  );
}
