import Link from "next/link";

// Lucide dropped brand marks, so the GitHub glyph stays inline.
export function GitHubMark({ size = 14 }: { size?: number }) {
  return (
    <svg
      aria-hidden="true"
      fill="currentColor"
      height={size}
      viewBox="0 0 24 24"
      width={size}
    >
      <path d="M12 .5C5.65.5.5 5.65.5 12c0 5.09 3.29 9.4 7.86 10.93.58.1.79-.25.79-.56v-2.02c-3.2.7-3.87-1.37-3.87-1.37-.53-1.33-1.29-1.69-1.29-1.69-1.05-.72.08-.7.08-.7 1.16.08 1.77 1.2 1.77 1.2 1.03 1.76 2.7 1.25 3.36.96.1-.75.4-1.25.73-1.54-2.55-.29-5.24-1.28-5.24-5.68 0-1.25.45-2.28 1.19-3.08-.12-.29-.52-1.46.11-3.04 0 0 .98-.31 3.18 1.18A11.1 11.1 0 0 1 12 6.2c.98 0 1.96.13 2.88.39 2.2-1.49 3.17-1.18 3.17-1.18.64 1.58.24 2.75.12 3.04.74.8 1.18 1.83 1.18 3.08 0 4.42-2.69 5.39-5.25 5.67.42.36.78 1.06.78 2.14v3.03c0 .31.21.67.8.56A11.52 11.52 0 0 0 23.5 12C23.5 5.65 18.35.5 12 .5Z" />
    </svg>
  );
}

// The newsjack.sh mark: "N" plus a vermilion cursor underscore, redrawn on a
// 100-unit grid from the brand logo (same artwork as app/icon.svg).
function MarkShapes({ letter, cursor }: { letter: string; cursor: string }) {
  return (
    <>
      <path
        d="M25 36h5v31h-5zM41 36h5v31h-5zM25 36h8l13 31h-8z"
        fill={letter}
      />
      <rect fill={cursor} height="5" width="22" x="53" y="71" />
    </>
  );
}

export function LogoMark({ size = 28 }: { size?: number }) {
  return (
    <svg aria-hidden="true" height={size} viewBox="0 0 100 100" width={size}>
      <rect fill="#1A1A1A" height="100" width="100" />
      <MarkShapes cursor="#E05A47" letter="#F9F8F6" />
    </svg>
  );
}

// The mark without its tile, for oversized decorative fills.
export function MarkGlyph({ className }: { className?: string }) {
  return (
    <svg aria-hidden="true" className={className} viewBox="22 33 56 46">
      <MarkShapes cursor="currentColor" letter="currentColor" />
    </svg>
  );
}

export function Wordmark({ small = false }: { small?: boolean }) {
  return (
    <Link
      aria-label="newsjack.sh home"
      className="group flex items-center gap-2.5"
      href="/"
    >
      <LogoMark size={small ? 24 : 28} />
      <span
        className={`font-mono font-bold tracking-[-0.02em] ${small ? "text-sm" : "text-lg"}`}
      >
        newsjack
        <span className="opacity-40 transition-opacity group-hover:opacity-70">
          .sh
        </span>
      </span>
    </Link>
  );
}
