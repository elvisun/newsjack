"use client";

import { Check, Copy } from "lucide-react";
import { useEffect, useRef, useState } from "react";

export const INSTALL_COMMAND = "curl -fsSL newsjack.sh | bash";

export function CopyCommand({
  value,
  prompt = "$",
  tone = "paper",
  label,
}: {
  value: string;
  prompt?: string | null;
  tone?: "paper" | "ink";
  label: string;
}) {
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

  return (
    <div
      className={`flex items-stretch overflow-hidden rounded-[2px] border ${
        tone === "ink"
          ? "border-page/20 bg-black/40"
          : "border-ink bg-ink shadow-lift"
      }`}
    >
      <code className="flex min-w-0 flex-1 items-center gap-3 overflow-x-auto whitespace-nowrap px-4 py-3.5 font-mono text-[13px] text-page sm:text-sm">
        {prompt && (
          <span aria-hidden="true" className="select-none text-accent">
            {prompt}
          </span>
        )}
        {value}
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
