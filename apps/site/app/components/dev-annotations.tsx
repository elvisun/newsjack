"use client";

import dynamic from "next/dynamic";

// Agentation is a dev-only toolbar for annotating UI and handing the notes to
// a coding agent. It never loads outside `next dev`.
const Agentation = dynamic(
  () => import("agentation").then((mod) => mod.Agentation),
  { ssr: false },
);

export function DevAnnotations() {
  if (process.env.NODE_ENV !== "development") return null;
  return <Agentation appName="newsjack.sh" />;
}
