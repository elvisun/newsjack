import { readFile } from "node:fs/promises";
import { join } from "node:path";

export const runtime = "nodejs";
export const dynamic = "force-dynamic";

async function readBundledInstaller() {
  try {
    return await readFile(
      join(process.cwd(), "public/_generated/install.sh"),
      "utf8",
    );
  } catch {
    return readFile(join(process.cwd(), "../../install.sh"), "utf8");
  }
}

export async function GET() {
  const body = await readBundledInstaller();
  const headers = new Headers({
    "cache-control": "public, max-age=60, stale-while-revalidate=300",
    "content-type": "text/x-shellscript; charset=utf-8",
  });

  return new Response(body, { headers });
}
