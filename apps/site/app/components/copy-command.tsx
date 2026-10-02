"use client";

import { Check, Copy } from "lucide-react";
import { useEffect, useRef, useState } from "react";

function useCopy(value: string) {
  const [copied, setCopied] = useState(false);
  const timer = useRef<ReturnType<typeof setTimeout>>(undefined);

  useEffect(() => () => clearTimeout(timer.current), []);

  async function copy() {
    try {
      await navigator.clipboard.writeText(value);
      setCopied(true);
      clearTimeout(timer.current);
      timer.current = setTimeout(() => setCopied(false), 1600);
    } catch (error) {
      console.error(`Copy failed. Copy it manually: ${value}`, error);
    }
  }

  return { copied, copy };
}

// Types the whole value out in TYPE_MS, like a pasted command echoing into a
// shell. Timed off animation frames, so speed doesn't depend on timer clamps.
const TYPE_MS = 200;

function useTypewriter(value: string, enabled: boolean) {
  const [length, setLength] = useState(enabled ? 0 : value.length);

  useEffect(() => {
    if (!enabled) return;
    const reduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)",
    ).matches;
    let frame = 0;
    const start = performance.now();
    const tick = (now: number) => {
      const progress = reduced ? 1 : Math.min(1, (now - start) / TYPE_MS);
      setLength(Math.round(progress * value.length));
      if (progress < 1) frame = requestAnimationFrame(tick);
    };
    frame = requestAnimationFrame(tick);
    return () => cancelAnimationFrame(frame);
  }, [enabled, value]);

  return { typed: value.slice(0, length), done: length >= value.length };
}

export function CopyCommand({
  value,
  prompt = "$",
  tone = "paper",
  label,
  wrap = false,
  typewriter = false,
}: {
  value: string;
  prompt?: string | null;
  tone?: "paper" | "ink";
  label: string;
  wrap?: boolean;
  typewriter?: boolean;
}) {
  const { copied, copy } = useCopy(value);
  const { typed, done } = useTypewriter(value, typewriter);

  return (
    <div
      className={`flex items-stretch overflow-hidden rounded-[2px] border ${
        tone === "ink"
          ? "border-page/20 bg-black/40"
          : "border-ink bg-ink shadow-lift"
      }`}
    >
      <code
        className={`flex min-w-0 flex-1 gap-3 px-4 py-3.5 font-mono text-[13px] leading-relaxed text-page sm:text-sm ${
          wrap
            ? "items-start"
            : "items-center overflow-x-auto whitespace-nowrap"
        }`}
      >
        {prompt && (
          <span aria-hidden="true" className="select-none text-accent">
            {prompt}
          </span>
        )}
        {typewriter ? (
          <span>
            <span className="sr-only">{value}</span>
            <span aria-hidden="true">{typed}</span>
            <span
              aria-hidden="true"
              className={`ml-0.5 inline-block h-[1.1em] w-2 translate-y-[0.15em] bg-accent ${done ? "nj-blink" : ""}`}
            />
          </span>
        ) : (
          value
        )}
      </code>
      <button
        aria-label={copied ? `${label} copied` : `Copy ${label}`}
        className="inline-flex shrink-0 cursor-pointer items-center gap-2 bg-accent px-4 font-mono text-[11px] tracking-[0.15em] text-white uppercase transition-colors hover:bg-accent-2 active:scale-95"
        onClick={copy}
        type="button"
      >
        {copied ? (
          <Check aria-hidden="true" size={14} />
        ) : (
          <Copy aria-hidden="true" size={14} />
        )}
        <span className="hidden sm:inline">{copied ? "Copied" : "Copy"}</span>
      </button>
    </div>
  );
}

// A compact chip that copies its value, e.g. a skill's slash command.
export function CopyChip({ value }: { value: string }) {
  const { copied, copy } = useCopy(value);

  return (
    <button
      aria-label={copied ? `Copied ${value}` : `Copy ${value}`}
      className="flex w-full min-w-0 cursor-pointer items-center justify-between gap-2 rounded-[2px] border border-ink/15 bg-page px-2.5 py-1.5 font-mono text-[11px] text-ink/80 transition-colors hover:border-accent hover:text-ink focus-visible:border-accent"
      onClick={copy}
      type="button"
    >
      <span className="truncate">
        <span className="text-accent">➜</span> {value}
      </span>
      {copied ? (
        <Check aria-hidden="true" className="shrink-0 text-accent" size={12} />
      ) : (
        <Copy aria-hidden="true" className="shrink-0 text-ink/40" size={12} />
      )}
    </button>
  );
}

// Inline command that copies on click, e.g. the terminal fallback line.
export function InlineCopy({ value }: { value: string }) {
  const { copied, copy } = useCopy(value);

  return (
    <button
      aria-label={copied ? `Copied ${value}` : `Copy ${value}`}
      className="inline-flex cursor-pointer items-center gap-1.5 font-mono text-page/70 transition-colors hover:text-page"
      onClick={copy}
      type="button"
    >
      <code>{value}</code>
      {copied ? (
        <Check aria-hidden="true" className="text-accent" size={12} />
      ) : (
        <Copy aria-hidden="true" size={12} />
      )}
      {copied && <span className="text-accent">Copied</span>}
    </button>
  );
}
