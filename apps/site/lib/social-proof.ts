// Live public numbers for the landing page. Every loader returns null (or an
// empty list) on failure so a GitHub or npm outage never breaks a render.

export const REPO = "elvisun/newsjack";
export const REPO_URL = `https://github.com/${REPO}`;
const GITHUB_API = `https://api.github.com/repos/${REPO}`;
const NPM_PACKAGE = "newsjack";
const REVALIDATE_SECONDS = 900;

export interface RepoStats {
  stars: number;
  forks: number;
  pushedAt: string;
}

export interface Forker {
  login: string;
  avatarUrl: string;
  forkUrl: string;
}

function githubHeaders(): HeadersInit {
  const headers: Record<string, string> = {
    Accept: "application/vnd.github+json",
    "X-GitHub-Api-Version": "2022-11-28",
  };
  // Optional: unauthenticated calls share a 60/hour limit per egress IP.
  if (process.env.GITHUB_TOKEN) {
    headers.Authorization = `Bearer ${process.env.GITHUB_TOKEN}`;
  }
  return headers;
}

async function getJson<T>(url: string, headers?: HeadersInit): Promise<T | null> {
  try {
    const response = await fetch(url, {
      headers,
      next: { revalidate: REVALIDATE_SECONDS },
    });
    if (!response.ok) return null;
    return (await response.json()) as T;
  } catch {
    return null;
  }
}

export async function loadRepoStats(): Promise<RepoStats | null> {
  const json = await getJson<{
    stargazers_count?: number;
    forks_count?: number;
    pushed_at?: string;
  }>(GITHUB_API, githubHeaders());
  if (typeof json?.stargazers_count !== "number") return null;
  return {
    stars: json.stargazers_count,
    forks: json.forks_count ?? 0,
    pushedAt: json.pushed_at ?? "",
  };
}

export async function loadStarCount(): Promise<number | null> {
  return (await loadRepoStats())?.stars ?? null;
}

const MAX_FORK_PAGES = 10;

// Listing stargazers now needs an authenticated token with starring access;
// forks stay public, and a fork is the stronger signal anyway.
export async function loadAllForkers(): Promise<Forker[]> {
  const perPage = 100;
  const forkers: Forker[] = [];

  for (let page = 1; page <= MAX_FORK_PAGES; page += 1) {
    const forks = await getJson<
      { html_url: string; owner?: { login: string; avatar_url: string } }[]
    >(
      `${GITHUB_API}/forks?sort=newest&per_page=${perPage}&page=${page}`,
      githubHeaders(),
    );
    if (!forks) break;
    for (const fork of forks) {
      if (fork.owner) {
        forkers.push({
          login: fork.owner.login,
          avatarUrl: fork.owner.avatar_url,
          forkUrl: fork.html_url,
        });
      }
    }
    if (forks.length < perPage) break;
  }

  return forkers;
}

export async function loadReleaseCount(): Promise<number | null> {
  const releases = await getJson<unknown[]>(
    `${GITHUB_API}/releases?per_page=100`,
    githubHeaders(),
  );
  return Array.isArray(releases) ? releases.length : null;
}

export async function loadNpmMonthlyDownloads(): Promise<number | null> {
  const json = await getJson<{ downloads?: number }>(
    `https://api.npmjs.org/downloads/point/last-month/${NPM_PACKAGE}`,
  );
  return typeof json?.downloads === "number" ? json.downloads : null;
}

export function formatCompact(count: number): string {
  if (count < 1000) return String(count);
  const thousands = count / 1000;
  const formatted =
    thousands >= 10
      ? String(Math.round(thousands))
      : thousands.toFixed(1).replace(/\.0$/, "");
  return `${formatted}k`;
}

export function formatFull(count: number): string {
  return count.toLocaleString("en-US");
}

export function formatDaysAgo(iso: string, now = new Date()): string | null {
  const then = new Date(iso);
  if (Number.isNaN(then.getTime())) return null;
  const days = Math.floor((now.getTime() - then.getTime()) / 86_400_000);
  if (days <= 0) return "Today";
  if (days === 1) return "Yesterday";
  return `${days} days ago`;
}
