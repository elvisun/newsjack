export const CHATGPT_INSTALL_PROMPT = [
  "Install the Newsjack plugin for me. It is not in the plugin directory, so do not search for it there. Install it from its public GitHub repo by running these commands in the terminal, in order:",
  "",
  "codex plugin marketplace add elvisun/newsjack",
  "codex plugin add newsjack@newsjack",
  "codex mcp login medialyst",
  "",
  "If the `codex` command is not found, use the Codex CLI that ships inside the ChatGPT desktop app. The last command opens a browser window so I can sign in to Medialyst (free), which gives Newsjack live news search. If you cannot run terminal commands in this chat, stop and tell me to switch ChatGPT to Work mode in the desktop app. When everything succeeds, tell me to quit and reopen ChatGPT, then start a new chat and type @Newsjack.",
].join("\n");
