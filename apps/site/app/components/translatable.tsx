"use client";

import { useState, type ReactNode } from "react";

// X-style "Show translation" toggle. The original shows by default; the
// translation is static text styled like the tweet body it replaces.
export function Translatable({
  original,
  translation,
  language,
}: {
  original: ReactNode;
  translation: string;
  language: string;
}) {
  const [translated, setTranslated] = useState(false);

  return (
    <>
      {translated ? (
        <p
          lang="en"
          style={{
            fontSize: "var(--tweet-body-font-size)",
            fontWeight: "var(--tweet-body-font-weight)",
            lineHeight: "var(--tweet-body-line-height)",
            margin: "var(--tweet-body-margin)",
            overflowWrap: "break-word",
            whiteSpace: "pre-wrap",
          }}
        >
          {translation}
        </p>
      ) : (
        original
      )}
      <p className="-mt-1 mb-2 text-[15px]">
        {translated && (
          <span className="text-[#536471]">Translated from {language} · </span>
        )}
        <button
          className="cursor-pointer text-[#1d9bf0] hover:underline"
          onClick={() => setTranslated((value) => !value)}
          type="button"
        >
          {translated ? "Show original" : "Show translation"}
        </button>
      </p>
    </>
  );
}
