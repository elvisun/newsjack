import type { Metadata } from "next";

// Canonical origin for metadata, the sitemap, and structured data. Must match
// the primary domain in Vercel; any other host should redirect here.
export const SITE_URL = "https://newsjack.sh";
export const SITE_NAME = "newsjack.sh";

export const SITE_TITLE = "newsjack.sh — open-source PR skills for AI agents";
export const SITE_DESCRIPTION =
  "Open-source PR skills for Claude Code, Codex, and other AI agents. Spot newsjacking openings, find the angle, vet reporters, and roast pitches before you send.";

export const OG_IMAGE = {
  url: "/newsjack-og-image.png",
  width: 1497,
  height: 789,
  alt: "newsjack.sh - The open-source skills that turn your agent into a PR operator.",
};

type OpenGraphImage = { url: string; width?: number; height?: number; alt?: string };

// Per-page metadata. Next.js replaces (not merges) a parent's openGraph, so
// every page sets its own; canonical stays per-page so no page inherits "/".
export function pageMetadata({
  title,
  description,
  path,
  image = OG_IMAGE,
  article,
}: {
  title?: string;
  description: string;
  path: string;
  image?: OpenGraphImage;
  article?: { publishedTime: string; modifiedTime?: string; authors: string[] };
}): Metadata {
  const openGraph = {
    // Share cards for section pages need the site name; article titles stand alone.
    title: !title ? SITE_TITLE : article ? title : `${title} | ${SITE_NAME}`,
    description,
    url: path,
    siteName: SITE_NAME,
    images: [image],
  };

  return {
    ...(title && { title }),
    description,
    alternates: { canonical: path },
    openGraph: article
      ? { ...openGraph, type: "article", ...article }
      : { ...openGraph, type: "website" },
  };
}

export function absoluteUrl(path: string): string {
  return new URL(path, SITE_URL).toString();
}
