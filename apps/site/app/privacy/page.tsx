import type { Metadata } from "next";

import { pageMetadata } from "../../lib/site";
import { Footer } from "../components/footer";
import { Nav } from "../components/nav";

const LAST_UPDATED = "2026-10-17";
const CONTACT_EMAIL = "elvis@medialyst.ai";

export const metadata: Metadata = pageMetadata({
  title: "Privacy policy",
  description:
    "What the newsjack.sh site, the open-source newsjack CLI, and its agent skills store, what leaves your machine, and how journalist contact data is handled.",
  path: "/privacy",
});

function formatDate(iso: string): string {
  return new Date(`${iso}T00:00:00Z`).toLocaleDateString("en-US", {
    year: "numeric",
    month: "long",
    day: "numeric",
    timeZone: "UTC",
  });
}

function Section({
  id,
  title,
  children,
}: {
  id: string;
  title: string;
  children: React.ReactNode;
}) {
  return (
    <section id={id} className="scroll-mt-28">
      <h2 className="mt-16 border-t border-ink/10 pt-10 font-serif text-[2rem] leading-[1.15] italic">
        {title}
      </h2>
      {children}
    </section>
  );
}

function H3({ children }: { children: React.ReactNode }) {
  return (
    <h3 className="mt-10 font-serif text-2xl leading-[1.2] font-medium">
      {children}
    </h3>
  );
}

function P({ children }: { children: React.ReactNode }) {
  return <p className="mt-6 text-[17px] leading-[1.7] text-ink/80">{children}</p>;
}

function List({ children }: { children: React.ReactNode }) {
  return (
    <ul className="mt-4 list-disc space-y-2 pl-6 text-[17px] leading-[1.7] text-ink/80 marker:text-accent">
      {children}
    </ul>
  );
}

function A({ href, children }: { href: string; children: React.ReactNode }) {
  const external = href.startsWith("http");
  return (
    <a
      className="text-ink underline decoration-accent/40 underline-offset-2 transition-colors hover:text-accent hover:decoration-accent"
      href={href}
      {...(external && { rel: "noreferrer", target: "_blank" })}
    >
      {children}
    </a>
  );
}

function Code({ children }: { children: React.ReactNode }) {
  return <code className="font-mono text-[0.85em] text-ink">{children}</code>;
}

const Contact = () => <A href={`mailto:${CONTACT_EMAIL}`}>{CONTACT_EMAIL}</A>;

export default function PrivacyPage() {
  return (
    <>
      <Nav />
      <main className="mx-auto max-w-3xl px-6 pt-36 pb-32">
        <span className="nj-eyebrow">newsjack.sh</span>
        <h1 className="mt-6 font-serif text-[clamp(2.5rem,7vw,4.5rem)] leading-[0.95] tracking-[-0.03em] italic">
          Privacy policy
        </h1>
        <p className="nj-meta mt-6">
          Last updated{" "}
          <time dateTime={LAST_UPDATED}>{formatDate(LAST_UPDATED)}</time>
        </p>

        <P>
          Newsjack is open-source software that runs on your own computer,
          inside the AI agent you already use. Most of what it does never
          leaves your machine. This page explains, in plain terms, what it
          stores, what it sends and to whom, how journalist contact details
          are handled, and how to delete everything.
        </P>

        <Section id="scope" title="Who we are and what this covers">
          <P>
            Newsjack is an open-source project (MIT licensed) maintained by
            Medialyst. In this policy, &ldquo;we&rdquo; means the Newsjack
            maintainers. The policy covers:
          </P>
          <List>
            <li>
              the website at <Code>newsjack.sh</Code>, including the installer
              it serves;
            </li>
            <li>
              the <Code>newsjack</Code> command-line tool (CLI), whether
              installed from <Code>newsjack.sh</Code>, GitHub, or npm;
            </li>
            <li>
              the Newsjack agent skills and plugin, wherever you install them.
            </li>
          </List>
          <P>
            It does not cover third-party services you choose to connect, such
            as Medialyst, X, or Slack, or the AI agent and model provider you
            run Newsjack in. Each of those has its own privacy policy. The
            source code is public at{" "}
            <A href="https://github.com/elvisun/newsjack">
              github.com/elvisun/newsjack
            </A>
            , so you can check every statement here against it.
          </P>
        </Section>

        <Section id="local" title="Your data stays on your machine by default">
          <P>
            Newsjack has no Newsjack account and no Newsjack server that your
            work is uploaded to. The CLI and skills keep their state in files
            on your computer. By default:
          </P>
          <List>
            <li>
              <Code>~/.newsjack/</Code> is the main folder (change it with the{" "}
              <Code>NEWSJACK_HOME</Code> environment variable). It holds:
              <List>
                <li>
                  <Code>bin/</Code> and <Code>newsjack/</Code>: the CLI and
                  the bundled skills, plus the previous version for rollback,
                  and <Code>install.json</Code> recording the installed
                  version.
                </li>
                <li>
                  <Code>credentials.json</Code>: your Medialyst API key or
                  sign-in tokens, if you connect Medialyst (readable only by
                  your user account).
                </li>
                <li>
                  <Code>.env</Code>: optional keys you add, such as an X API
                  token or a TypeSafe API key (readable only by your user
                  account).
                </li>
                <li>
                  <Code>monitors/&lt;name&gt;/</Code>: each monitor profile
                  (company, topics, competitors, proof points, spokespeople,
                  feeds), its brief and schedule notes, and a{" "}
                  <Code>runs/</Code> folder with the news evidence and reports
                  from each run.
                </li>
                <li>
                  <Code>monitors/&lt;name&gt;/delivery.json</Code>: your Slack
                  webhook URL, if you set up Slack delivery (readable only by
                  your user account), plus small delivery receipts so the same
                  report is not posted twice.
                </li>
                <li>
                  <Code>coverage/&lt;name&gt;/</Code>: coverage tracker
                  keywords and runs.
                </li>
                <li>
                  <Code>voice/</Code>: voice profiles (for example{" "}
                  <Code>voice/active.yaml</Code>) describing how you write. The
                  skill is told never to store your raw writing samples in
                  them.
                </li>
              </List>
            </li>
            <li>
              <Code>~/.local/share/newsjack/monitor.db</Code> is a local SQLite
              database (change it with <Code>NEWSJACK_STORE</Code>). It
              remembers which links a monitor or coverage tracker has already
              seen, along with run history and the news evidence collected.
              That evidence can include public article bylines and public X
              posts and handles.
            </li>
            <li>
              Agent skill folders: the installer copies Newsjack skills into
              the folders your agents read, such as{" "}
              <Code>~/.claude/skills</Code>, <Code>~/.agents/skills</Code>,{" "}
              <Code>~/.openclaw/skills</Code>, or{" "}
              <Code>~/.hermes/skills</Code>.
            </li>
            <li>
              Files you ask for: press clips, coverage reports, media lists,
              drafts, and other outputs are saved wherever you or your agent
              choose.
            </li>
          </List>
          <P>
            On macOS and Linux the installer does not edit your shell
            configuration. On Windows, the CLI adds{" "}
            <Code>%USERPROFILE%\.newsjack\bin</Code> to your user PATH unless
            you set <Code>NEWSJACK_NO_PATH_UPDATE=1</Code>.
          </P>
          <P>
            The CLI contains no analytics, telemetry, or crash reporting, and
            it does not call any AI model itself. Your AI agent reads and
            writes these files when you run a skill, so whatever you share
            with your agent is also handled by the agent and the model provider
            it uses, under their terms.
          </P>
        </Section>

        <Section id="third-parties" title="What leaves your machine, and to whom">
          <P>
            Newsjack only contacts outside services to do what you ask. Here
            is every service it can talk to and what it sends.
          </P>

          <H3>GitHub (installs and updates)</H3>
          <P>
            The installer and the CLI download release files from GitHub.
            Before most commands, an installed CLI also checks GitHub for a
            newer release. That check is a plain download of a public file and
            sends no account, version, or usage information beyond what any
            web request carries (such as your IP address). Turn it off with{" "}
            <Code>NEWSJACK_AUTO_UPDATE=0</Code>. GitHub&rsquo;s own privacy
            policy applies to these requests.
          </P>

          <H3>newsjack.sh (the installer)</H3>
          <P>
            Running <Code>curl -fsSL https://newsjack.sh | sh</Code> downloads
            the installer from this site. On macOS and Linux,{" "}
            <Code>newsjack update</Code> and automatic updates fetch the
            installer the same way. These requests are logged like any other
            visit to the site (see below).
          </P>

          <H3>Medialyst (optional)</H3>
          <P>
            Medialyst is an optional service you can connect for news search,
            journalist research, media lists, and the PR calendar. Nothing is
            sent to Medialyst unless you sign in or add a Medialyst key. Once
            connected, requests go to <Code>medialyst.ai</Code> with your key
            or sign-in token, and include what each feature needs: search
            terms (for a monitor, its topics, competitors, and search terms),
            the campaign brief you approve for a media list, article links to
            enrich, and calendar queries. The plugin can also connect your
            agent to Medialyst&rsquo;s MCP server at{" "}
            <Code>medialyst.ai/api/mcp</Code>, which receives whatever your
            agent sends through those tools. Medialyst handles this data under
            its own{" "}
            <A href="https://medialyst.ai/legal/privacy-policy">
              privacy policy
            </A>
            .
          </P>

          <H3>News sources</H3>
          <P>
            Monitors read RSS feeds (by default Techmeme&rsquo;s, or the feeds
            in your profile) with plain downloads that send no search terms.
            If you turn on the Reddit or Hacker News sources, your search terms
            are sent to Reddit&rsquo;s and Hacker News&rsquo;s (Algolia&rsquo;s)
            public search. The press-clip skill opens the article links you
            give it in a fresh headless browser, with no cookies or browser
            profile of yours.
          </P>

          <H3>X (optional)</H3>
          <P>
            If you add an X API token, monitors send your search terms to the
            X API to find recent posts, news, and trends, using your token.
          </P>

          <H3>Slack (optional)</H3>
          <P>
            If you set up Slack delivery, the CLI posts each report message to
            the Slack incoming webhook you provide, and nowhere else. The
            detector skill writes that message as a short summary with source
            links and the local path to the full report, and is told never to
            include credentials or private data. You can remove it any time with{" "}
            <Code>newsjack monitor delivery remove-slack &lt;name&gt;</Code>.
          </P>

          <H3>TypeSafe (optional)</H3>
          <P>
            If you add a TypeSafe API key, the relevance filter sends your
            monitor profile (company, description, website, topics,
            competitors, search terms, and exclusions) and short excerpts of
            the collected news to TypeSafe to score relevance. Without a key,
            nothing is sent.
          </P>
        </Section>

        <Section id="site" title="The newsjack.sh website">
          <P>
            The site sets no cookies and runs no third-party analytics or
            advertising scripts. When someone requests the home page (or the
            installer at <Code>newsjack.sh</Code>), we log one record so we can
            count visits and installs. Each record holds:
          </P>
          <List>
            <li>whether it was a page visit or an installer download, and when;</li>
            <li>
              a one-way hash of your IP address, salted and changed every day,
              so the same visitor can be counted once per day. We do not store
              your raw IP address;
            </li>
            <li>approximate country and region;</li>
            <li>
              your browser or tool&rsquo;s user agent, the referring page, and
              your language preference;
            </li>
            <li>the query string in the link you followed (such as UTM tags).</li>
          </List>
          <P>
            We use these records only to understand traffic and show the
            install count. We do not sell them or use them for advertising.
            The site is hosted on Vercel, which processes requests to serve it.
          </P>
          <P>
            Some pages show posts from X and LinkedIn. Their images load
            directly from those companies&rsquo; servers, which can see your
            IP address. YouTube videos load from{" "}
            <Code>youtube-nocookie.com</Code> only when you press play.
          </P>
        </Section>

        <Section id="journalists" title="Journalist contact data">
          <P>
            Newsjack helps people find journalists who actually cover a story.
            Journalist names, outlets, email addresses, and social handles
            reach a Newsjack user in one of two ways.
          </P>

          <H3>(a) Through the optional Medialyst connection</H3>
          <P>
            If the user chooses to connect Medialyst, contact details come
            from Medialyst&rsquo;s journalist research (the find-journalists
            skill, media lists, and journalist enrichment). Medialyst&rsquo;s{" "}
            <A href="https://medialyst.ai/legal/privacy-policy">
              privacy policy
            </A>{" "}
            (section 6) says it builds journalist records from:
          </P>
          <List>
            <li>
              publicly available information: published articles, bylines and
              author bios, official author pages on publication websites,
              professional bios, press releases, and public speaking;
            </li>
            <li>
              professional social profiles, such as public LinkedIn profiles
              and X accounts used for professional purposes;
            </li>
            <li>
              a business contact enrichment provider (Hunter.io) and email
              validation services (DeBounce, Instantly);
            </li>
            <li>media lists uploaded by its users, and corrections.</li>
          </List>
          <P>
            Medialyst says it keeps professional information only (name,
            professional email and work phone if public, outlet, role, beat,
            published work, and professional profiles), and does not collect
            home addresses, personal email addresses, sensitive data, or
            details of private or family life.
          </P>
          <P>
            Medialyst processes this data on the basis of legitimate interest
            under GDPR Article 6(1)(f), for professional media-relations
            communication. Its stated balancing points are that the data is
            limited to professional information journalists have made public,
            journalists expect professional communications as part of their
            role, the processing offers them relevant story opportunities, and
            safeguards include an easy opt-out.
          </P>
          <P>
            <strong className="font-medium text-ink">Removal.</strong> A
            journalist can opt out or ask for deletion at{" "}
            <A href="https://medialyst.ai/journalist-optout">
              medialyst.ai/journalist-optout
            </A>{" "}
            or by emailing{" "}
            <A href="mailto:privacy@medialyst.ai?subject=Journalist%20Opt-Out%20Request">
              privacy@medialyst.ai
            </A>{" "}
            with the subject &ldquo;Journalist Opt-Out Request&rdquo;
            (include your name, professional email addresses, and
            publications). Medialyst says it acknowledges requests within 5
            business days, completes them within 30 days, and adds you to a
            permanent suppression list. Once you are removed from Medialyst,
            you stop appearing in Newsjack results that come through
            Medialyst. Newsjack keeps no separate copy of Medialyst results;
            anything a user already saved stays in that user&rsquo;s own files.
          </P>

          <H3>(b) Without Medialyst: your own agent&rsquo;s lookup</H3>
          <P>
            Without Medialyst, the user&rsquo;s own AI agent does a
            best-effort lookup for an email address or handle that the
            journalist or their publication has listed publicly, such as an
            author or staff page, a masthead, or the journalist&rsquo;s own
            website or public profile. The skill tells the agent to cite the
            page each contact came from, never to guess an address (for
            example from a name pattern), and never to use unofficial contact
            databases. If nothing is publicly listed, the contact is left
            blank.
          </P>
          <P>
            This lookup runs on the user&rsquo;s machine. Newsjack operates no
            server or database for it and keeps nothing. Results stay in the
            user&rsquo;s local files. The user decides what to keep and is
            responsible for it as its controller, and must use it only for
            relevant, professional media-relations outreach (legitimate
            interest).
          </P>
          <P>
            <strong className="font-medium text-ink">Removal.</strong> There
            is no Newsjack-held copy to delete. Removing or changing the public
            listing stops future lookups from finding it. Journalists can
            still contact us at <Contact /> with any question.
          </P>

          <H3>Newsjack never sends outreach</H3>
          <P>
            The CLI has no email-sending code, and the skills only research and
            draft. Every pitch or reply is reviewed and sent by the user, from
            their own accounts, if they decide to send it. Newsjack&rsquo;s{" "}
            <A href="https://github.com/elvisun/newsjack/blob/main/skills/ETHICS.md">
              ethics rules
            </A>{" "}
            also tell the agent to respect opt-outs and to refuse mass,
            untargeted pitching.
          </P>
          <P>
            Journalists who want to be removed can also email us at{" "}
            <Contact />. We will forward Medialyst-related requests to
            Medialyst and answer any other questions.
          </P>
        </Section>

        <Section id="your-data" title="Your choices and deleting your data">
          <P>
            Because your Newsjack data lives on your computer, you control it
            directly. To remove it:
          </P>
          <List>
            <li>
              Delete <Code>~/.newsjack</Code> (or your{" "}
              <Code>NEWSJACK_HOME</Code>). This removes the CLI, credentials,
              keys, monitor profiles, run folders, Slack settings, coverage
              trackers, and voice profiles.
            </li>
            <li>
              Delete <Code>~/.local/share/newsjack</Code> (or your{" "}
              <Code>NEWSJACK_STORE</Code> file) to remove the local database.
            </li>
            <li>
              Delete the Newsjack skill folders from your agents&rsquo; skill
              directories (listed above). Each one Newsjack installed contains
              a <Code>.newsjack-installed</Code> file.
            </li>
            <li>
              If you installed with npm, run{" "}
              <Code>npm uninstall -g newsjack</Code>. On Windows, also remove{" "}
              <Code>%USERPROFILE%\.newsjack\bin</Code> from your user PATH.
            </li>
          </List>
          <P>
            For narrower changes, <Code>newsjack auth logout</Code> removes
            your stored Medialyst credentials, and{" "}
            <Code>newsjack monitor delivery remove-slack &lt;name&gt;</Code>{" "}
            removes a Slack webhook. To stop using a connected service, also
            revoke the key or webhook in that service. For data held by
            Medialyst, X, Slack, GitHub, or TypeSafe, use their privacy
            controls. For the website&rsquo;s visit records, email us.
          </P>
        </Section>

        <Section id="children" title="Children">
          <P>
            Newsjack is a professional tool and is not directed at children.
            We do not knowingly collect personal information from children.
          </P>
        </Section>

        <Section id="changes" title="Changes to this policy">
          <P>
            We will update this page when Newsjack&rsquo;s data handling
            changes, and change the &ldquo;Last updated&rdquo; date above. The
            full history of this page is public in the GitHub repository.
          </P>
        </Section>

        <Section id="contact" title="Contact">
          <P>
            Questions, requests, or concerns about privacy: <Contact />.
          </P>
        </Section>
      </main>
      <Footer />
    </>
  );
}
