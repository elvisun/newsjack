"use client";

import { useState, type KeyboardEvent, type ReactNode } from "react";

import { CopyCommand, INSTALL_COMMAND } from "./copy-command";

type InstallOption = {
  id: string;
  label: string;
  value: string;
  prompt: string | null;
  copyLabel: string;
  note: ReactNode;
};

const noteLink =
  "text-ink underline decoration-accent/40 underline-offset-2 transition-colors hover:text-accent";

const OPTIONS: InstallOption[] = [
  {
    id: "terminal",
    label: "Terminal",
    value: INSTALL_COMMAND,
    prompt: "$",
    copyLabel: "install command",
    note: (
      <>
        macOS and Linux. Installs the CLI and adds every skill to Claude Code,
        Codex, Hermes and OpenClaw.{" "}
        <a
          className={noteLink}
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
  {
    id: "ai",
    label: "Ask your AI",
    value: "help me setup https://newsjack.sh",
    prompt: "➜",
    copyLabel: "setup prompt",
    note: "Not technical? Paste this into Claude, ChatGPT or Codex. Your agent reads the guide and handles the rest, Windows included.",
  },
  {
    id: "claude",
    label: "Claude.ai & Cowork",
    value: "elvisun/newsjack",
    prompt: null,
    copyLabel: "plugin repository",
    note: (
      <>
        Open{" "}
        <a
          className={noteLink}
          href="https://claude.ai/customize"
          rel="noreferrer"
          target="_blank"
        >
          Customize
        </a>{" "}
        → Personal plugins → Create plugin → Add marketplace → Add from a
        repository. Paste this, then install <code>newsjack@newsjack</code>.
      </>
    ),
  },
  {
    id: "npm",
    label: "npm",
    value: "npm i -g newsjack@latest",
    prompt: "$",
    copyLabel: "npm install command",
    note: "The same CLI and skills bundle, with newsjack on your PATH.",
  },
];

export function InstallPanel() {
  const [activeId, setActiveId] = useState(OPTIONS[0].id);
  const active = OPTIONS.find((option) => option.id === activeId) ?? OPTIONS[0];

  function onKeyDown(event: KeyboardEvent<HTMLButtonElement>, index: number) {
    const step =
      event.key === "ArrowRight" ? 1 : event.key === "ArrowLeft" ? -1 : 0;
    if (!step) return;
    event.preventDefault();
    const next = OPTIONS[(index + step + OPTIONS.length) % OPTIONS.length];
    setActiveId(next.id);
    document.getElementById(`install-tab-${next.id}`)?.focus();
  }

  return (
    <div className="border border-ink/10 bg-white shadow-editorial">
      <div
        aria-label="Choose how to install"
        className="flex overflow-x-auto border-b border-ink/10"
        role="tablist"
      >
        {OPTIONS.map((option, index) => {
          const selected = option.id === active.id;
          return (
            <button
              aria-controls="install-panel"
              aria-selected={selected}
              className={`-mb-px shrink-0 cursor-pointer border-b-2 px-4 py-3.5 font-mono text-[10px] tracking-[0.2em] uppercase transition-colors sm:px-5 ${
                selected
                  ? "border-accent text-ink"
                  : "border-transparent text-ink/40 hover:text-ink"
              }`}
              id={`install-tab-${option.id}`}
              key={option.id}
              onClick={() => setActiveId(option.id)}
              onKeyDown={(event) => onKeyDown(event, index)}
              role="tab"
              tabIndex={selected ? 0 : -1}
              type="button"
            >
              {option.label}
            </button>
          );
        })}
      </div>
      <div
        aria-labelledby={`install-tab-${active.id}`}
        className="p-5 sm:p-6"
        id="install-panel"
        role="tabpanel"
      >
        <CopyCommand
          key={active.id}
          label={active.copyLabel}
          prompt={active.prompt}
          value={active.value}
        />
        <p className="mt-4 text-[13px] leading-relaxed text-ink/60">
          {active.note}
        </p>
      </div>
    </div>
  );
}
