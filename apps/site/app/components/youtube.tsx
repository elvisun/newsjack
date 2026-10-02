"use client";

import Image from "next/image";
import { useState } from "react";

// Click-to-load YouTube: nothing from YouTube loads until someone presses play.
function Player({
  id,
  start,
  title,
}: {
  id: string;
  start: number;
  title: string;
}) {
  return (
    <iframe
      allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture; web-share"
      allowFullScreen
      className="aspect-video w-full"
      src={`https://www.youtube-nocookie.com/embed/${id}?start=${start}&autoplay=1&rel=0`}
      title={title}
    />
  );
}

function formatTime(seconds: number): string {
  const minutes = Math.floor(seconds / 60);
  return `${minutes}:${String(seconds % 60).padStart(2, "0")}`;
}

// A YouTube video embed for "the clippings", drawn to match YouTube's card.
export function YouTubeCard({
  id,
  start,
  title,
  channel,
  channelUrl,
  channelAvatar,
  publishedAt,
  chapter,
}: {
  id: string;
  start: number;
  title: string;
  channel: string;
  channelUrl: string;
  channelAvatar: string;
  publishedAt: string;
  chapter: string;
}) {
  const [playing, setPlaying] = useState(false);
  const url = `https://www.youtube.com/watch?v=${id}&t=${start}`;
  const date = new Date(publishedAt).toLocaleDateString("en-US", {
    month: "short",
    day: "numeric",
    year: "numeric",
    timeZone: "UTC",
  });

  return (
    <article
      className="w-full max-w-[550px] overflow-hidden rounded-xl bg-white text-[#0f0f0f] shadow-[0_0_0_1px_rgba(0,0,0,0.1)]"
      style={{
        fontFamily: 'Roboto, -apple-system, "Segoe UI", Arial, sans-serif',
      }}
    >
      <div className="relative bg-black">
        {playing ? (
          <Player id={id} start={start} title={title} />
        ) : (
          <button
            aria-label={`Play "${title}" from ${formatTime(start)}`}
            className="group relative block aspect-video w-full cursor-pointer"
            onClick={() => setPlaying(true)}
            type="button"
          >
            <Image
              alt=""
              className="object-cover"
              fill
              sizes="(min-width: 1280px) 400px, (min-width: 768px) 50vw, 100vw"
              src={`https://i.ytimg.com/vi/${id}/maxresdefault.jpg`}
              unoptimized
            />
            <span className="absolute inset-0 flex items-center justify-center">
              <svg
                aria-hidden="true"
                className="h-12 w-[68px] opacity-90 transition-opacity group-hover:opacity-100"
                viewBox="0 0 68 48"
              >
                <path
                  d="M66.52 7.74c-.78-2.93-2.49-5.41-5.42-6.19C55.79.13 34 0 34 0S12.21.13 6.9 1.55c-2.93.78-4.63 3.26-5.42 6.19C.06 13.05 0 24 0 24s.06 10.95 1.48 16.26c.78 2.93 2.49 5.41 5.42 6.19C12.21 47.87 34 48 34 48s21.79-.13 27.1-1.55c2.93-.78 4.64-3.26 5.42-6.19C67.94 34.95 68 24 68 24s-.06-10.95-1.48-16.26z"
                  fill="#f00"
                />
                <path d="M45 24 27 14v20" fill="#fff" />
              </svg>
            </span>
            <span className="absolute right-2 bottom-2 rounded bg-black/80 px-1.5 py-0.5 text-xs font-medium text-white">
              Starts at {formatTime(start)}
            </span>
          </button>
        )}
      </div>
      <div className="flex gap-3 p-3">
        <a
          className="shrink-0"
          href={channelUrl}
          rel="noreferrer"
          target="_blank"
        >
          <Image
            alt=""
            className="rounded-full"
            height={36}
            src={channelAvatar}
            unoptimized
            width={36}
          />
        </a>
        <div className="min-w-0 flex-1">
          <a
            className="line-clamp-2 text-[15px] leading-snug font-medium hover:underline"
            href={url}
            rel="noreferrer"
            target="_blank"
          >
            {title}
          </a>
          <p className="mt-1 text-[13px] text-[#606060]">
            {channel} · {date}
          </p>
          <p className="mt-0.5 text-[13px] text-[#606060]">
            {formatTime(start)} — {chapter}
          </p>
        </div>
        <a href={url} rel="noreferrer" target="_blank">
          <Image
            alt="YouTube"
            className="size-6"
            height={24}
            src="/logos/youtube.svg"
            unoptimized
            width={24}
          />
        </a>
      </div>
    </article>
  );
}
