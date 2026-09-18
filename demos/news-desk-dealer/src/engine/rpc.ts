// One call to a keyed upstream API. In dev the call rides Vite's HMR
// WebSocket (no per-host connection cap, key stays on the server); without
// HMR it falls back to the HTTP proxy at /api/<route>.
export type ApiRoute = "/api/openrouter" | "/api/typesafe";
type Result = { status: number; text: string };

const hot = import.meta.hot;
const pending = new Map<string, (r: Result) => void>();
hot?.on("ndd:result", (r: Result & { id: string }) => { pending.get(r.id)?.(r); pending.delete(r.id); });

export function callApi(route: ApiRoute, path: string, body: unknown, signal?: AbortSignal): Promise<Result> {
  if (!hot) {
    return fetch(route + path, { method: "POST", headers: { "content-type": "application/json" }, body: JSON.stringify(body), signal })
      .then(async (res) => ({ status: res.status, text: await res.text() }));
  }
  const id = crypto.randomUUID();
  return new Promise<Result>((resolve, reject) => {
    const abort = () => { pending.delete(id); hot.send("ndd:abort", { id }); reject(new DOMException("aborted", "AbortError")); };
    if (signal?.aborted) return abort();
    signal?.addEventListener("abort", abort, { once: true });
    pending.set(id, (r) => { signal?.removeEventListener("abort", abort); resolve(r); });
    hot.send("ndd:call", { id, route, path, body });
  });
}
