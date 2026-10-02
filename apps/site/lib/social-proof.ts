import { unstable_cache } from "next/cache";

import { countInstalls } from "./install-telemetry";

// Live public numbers for the landing page. Every loader returns null (or an
// empty list) on failure so a GitHub, npm, or database outage never breaks a
// render.

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

async function getJson<T>(
  url: string,
  headers?: HeadersInit,
): Promise<T | null> {
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

// Cached for an hour across all requests, so page traffic never reaches the
// analytics database directly.
// Only real numbers are cached: a missing database or failed query throws, and
// unstable_cache never stores a thrown result, so the next render retries.
const cachedInstallCount = unstable_cache(
  async (): Promise<number> => {
    const installs = await countInstalls();
    if (installs === null) throw new Error("install count unavailable");
    return installs;
  },
  ["install-count"],
  { revalidate: 3600 },
);

export async function loadInstallCount(): Promise<number | null> {
  try {
    return await cachedInstallCount();
  } catch {
    return null;
  }
}

export async function loadReleaseCount(): Promise<number | null> {
  const releases = await getJson<unknown[]>(
    `${GITHUB_API}/releases?per_page=100`,
    githubHeaders(),
  );
  return Array.isArray(releases) ? releases.length : null;
}

// First publish of the newsjack package on npm. Only the main package is
// counted: its per-platform binaries install alongside it and would double up.
export const NPM_FIRST_PUBLISHED = "2026-06-08";

// All-time npm downloads, summed from the range endpoint in windows of at most
// 18 months (the API's limit). Includes updates, CI and mirrors.
export async function loadNpmTotalDownloads(): Promise<number | null> {
  const day = (date: Date) => date.toISOString().slice(0, 10);
  const windowMs = 540 * 86_400_000;
  const today = new Date();
  let total = 0;

  for (
    let from = new Date(NPM_FIRST_PUBLISHED);
    from <= today;
    from = new Date(from.getTime() + windowMs + 86_400_000)
  ) {
    const to = new Date(Math.min(today.getTime(), from.getTime() + windowMs));
    const json = await getJson<{ downloads?: { downloads: number }[] }>(
      `https://api.npmjs.org/downloads/range/${day(from)}:${day(to)}/${NPM_PACKAGE}`,
    );
    if (!json?.downloads) return null;
    total += json.downloads.reduce((sum, entry) => sum + entry.downloads, 0);
  }

  return total;
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
