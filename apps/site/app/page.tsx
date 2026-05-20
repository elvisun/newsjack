import { ManifestoHero, SiteShell } from "./_components/site-shell";

export default function Home() {
  return (
    <SiteShell>
      <main className="mx-auto flex min-h-[calc(100vh-4rem)] w-full max-w-6xl items-center px-5 py-12 sm:px-8 lg:px-10">
        <ManifestoHero />
      </main>
    </SiteShell>
  );
}
