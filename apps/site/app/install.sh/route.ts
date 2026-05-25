import { readFile } from "node:fs/promises";
import { join } from "node:path";

const installUrl =
  process.env.NEWSJACK_INSTALL_URL ??
  "https://raw.githubusercontent.com/elvisun/newsjack/main/install.sh";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

async function readBundledInstaller() {
  return readFile(join(process.cwd(), "../../install.sh"), "utf8");
}

export async function GET() {
  let body: string;

  try {
    const response = await fetch(installUrl, {
      cache: "no-store",
      headers: { accept: "text/x-shellscript,text/plain,*/*" },
    });

    if (!response.ok) {
      throw new Error(`installer fetch failed: ${response.status}`);
    }

    body = await response.text();
  } catch {
    body = await readBundledInstaller();
  }

  return new Response(body, {
    headers: {
      "cache-control": "public, max-age=60, stale-while-revalidate=300",
      "content-type": "text/x-shellscript; charset=utf-8",
    },
  });
}
