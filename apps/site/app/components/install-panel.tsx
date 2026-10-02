"use client";

import {
  MessageSquareText,
  SquareTerminal,
  type LucideIcon,
} from "lucide-react";
import { useState, type KeyboardEvent, type ReactNode } from "react";

import { CopyCommand } from "./copy-command";

type InstallOption = {
  id: string;
  label: string;
  icon: LucideIcon;
  kicker: string;
  value: string;
  prompt: string;
  copyLabel: string;
  note: ReactNode;
};

// Ask your AI (the default) or the terminal: the two install paths from the
// repo README.
export function InstallPanel({
  terminal,
  aiPrompt,
}: {
  terminal: string;
  aiPrompt: string;
}) {
  const options: InstallOption[] = [
    {
      id: "ai",
      label: "Ask your AI",
      icon: MessageSquareText,
      kicker: "Paste this into any AI",
      value: aiPrompt,
      prompt: "➜",
      copyLabel: "setup prompt",
      note: "Not technical? Claude, ChatGPT or Codex reads the repo, installs newsjack and sets up your daily monitoring. Windows included.",
    },
    {
      id: "terminal",
      label: "Terminal",
      icon: SquareTerminal,
      kicker: "Run this in your terminal",
      value: terminal,
      prompt: "$",
      copyLabel: "install command",
      note: (
        <>
          macOS and Linux. Installs the CLI and adds every skill to Claude Code,
          Codex, Hermes and OpenClaw.{" "}
          <a
            className="text-page underline decoration-accent/60 underline-offset-2 transition-colors hover:text-accent"
            href="https://github.com/elvisun/newsjack/blob/main/install.sh"
            rel="noreferrer"
            target="_blank"
          >
            Read install.sh first
          </a>
          .
        </>
      ),
    },
  ];
  const [activeId, setActiveId] = useState(options[0].id);
  const active = options.find((option) => option.id === activeId) ?? options[0];

  function onKeyDown(event: KeyboardEvent<HTMLButtonElement>, index: number) {
    const step =
      event.key === "ArrowRight" ? 1 : event.key === "ArrowLeft" ? -1 : 0;
    if (!step) return;
    event.preventDefault();
    const next = options[(index + step + options.length) % options.length];
    setActiveId(next.id);
    document.getElementById(`install-tab-${next.id}`)?.focus();
  }

  return (
    <div className="overflow-hidden rounded-lg bg-ink text-page shadow-terminal">
      <div className="flex items-center gap-6 border-b border-white/5 bg-black/40 px-5">
        <div aria-hidden="true" className="flex gap-1.5">
          {[0, 1, 2].map((dot) => (
            <span className="size-2.5 rounded-full bg-white/10" key={dot} />
          ))}
        </div>
        <div aria-label="Choose how to install" className="flex" role="tablist">
          {options.map((option, index) => {
            const selected = option.id === active.id;
            return (
              <button
                aria-controls="install-panel"
                aria-selected={selected}
                className={`-mb-px cursor-pointer border-b-2 px-4 py-4 font-mono text-[11px] tracking-[0.18em] uppercase transition-colors ${
                  selected
                    ? "border-accent text-page"
                    : "border-transparent text-page/70 hover:text-page"
                }`}
                id={`install-tab-${option.id}`}
                key={option.id}
                onClick={() => setActiveId(option.id)}
                onKeyDown={(event) => onKeyDown(event, index)}
                role="tab"
                tabIndex={selected ? 0 : -1}
                type="button"
              >
                <span className="flex items-center gap-2">
                  <option.icon aria-hidden="true" size={14} />
                  {option.id === "terminal" && !selected ? (
                    // Unopened terminal tab: a light sweep and a ping invite a click.
                    <>
                      <span className="nj-shimmer">{option.label}</span>
                      <span aria-hidden="true" className="relative flex size-2">
                        <span className="nj-ping absolute inline-flex size-full rounded-full bg-accent" />
                        <span className="relative inline-flex size-2 rounded-full bg-accent" />
                      </span>
                    </>
                  ) : (
                    option.label
                  )}
                </span>
              </button>
            );
          })}
        </div>
      </div>
      <div
        aria-labelledby={`install-tab-${active.id}`}
        className="p-6 sm:p-8"
        id="install-panel"
        role="tabpanel"
      >
        <p className="font-mono text-[10px] tracking-[0.2em] text-page/40 uppercase">
          {active.kicker}
        </p>
        <div className="mt-3">
          <CopyCommand
            key={active.id}
            label={active.copyLabel}
            prompt={active.prompt}
            tone="ink"
            typewriter={active.id === "terminal"}
            value={active.value}
            wrap={active.id === "ai"}
          />
        </div>
        <p className="mt-5 text-[13px] leading-relaxed text-page/60">
          {active.note}
        </p>
      </div>
    </div>
  );
}
