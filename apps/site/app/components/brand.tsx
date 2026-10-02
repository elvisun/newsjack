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

// The mark: an italic Newsreader "n" with a vermilion cursor block. The
// news, mid-sentence, being written by your agent. Outlined so it renders
// without webfonts (same artwork as app/icon.svg).
const MARK_N_PATH =
  "M68 606 57 612Q126 777 189 872Q252 966 318 1006Q384 1045 462 1045Q492 1045 515 1043Q538 1040 559 1035Q580 1030 604 1022L466 582H453Q540 731 608 822Q677 914 734 963Q791 1011 843 1029Q895 1046 948 1046Q1047 1046 1089 1002Q1131 957 1131 893Q1131 857 1120 813Q1109 769 1069 674L831 88L784 127Q819 121 869 135Q920 148 990 195Q1059 242 1151 335L1161 327Q1072 202 992 125Q913 48 844 13Q775 -22 716 -22Q642 -22 615 9Q588 40 614 104L843 659Q874 733 885 764Q896 794 896 812Q896 843 875 858Q854 874 812 874Q757 874 701 838Q645 801 595 744Q545 687 508 627Q471 568 454 522L278 0H60L342 838Q349 858 346 873Q342 888 324 888Q302 888 267 865Q231 843 182 783Q133 722 68 606Z";

export function LogoMark({ size = 28 }: { size?: number }) {
  return (
    <svg aria-hidden="true" height={size} viewBox="0 0 64 64" width={size}>
      <rect fill="#1A1A1A" height="64" width="64" />
      <g transform="translate(8.18 48.29) scale(0.03182 -0.03182)">
        <path d={MARK_N_PATH} fill="#F9F8F6" />
        <rect fill="#E05A47" height="852" width="190" x="1250" />
      </g>
    </svg>
  );
}

// The mark without its tile, for oversized decorative fills.
export function MarkGlyph({ className }: { className?: string }) {
  return (
    <svg aria-hidden="true" className={className} viewBox="6 12 52 40">
      <g transform="translate(8.18 48.29) scale(0.03182 -0.03182)">
        <path d={MARK_N_PATH} fill="currentColor" />
        <rect fill="currentColor" height="852" width="190" x="1250" />
      </g>
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
      <span className="flex items-baseline">
        <span
          className={`font-serif font-medium tracking-[-0.02em] italic ${small ? "text-[22px]" : "text-[26px]"} leading-none`}
        >
          newsjack
        </span>
        <span className="ml-0.5 font-mono text-[13px] leading-none text-ink/40 transition-colors group-hover:text-accent">
          .sh
        </span>
      </span>
    </Link>
  );
}
