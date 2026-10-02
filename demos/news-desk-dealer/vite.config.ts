import { defineConfig, loadEnv, type Plugin } from "vite";
import react from "@vitejs/plugin-react";

// Key-holding proxy. The browser calls /api/openrouter/* and /api/typesafe/*;
// the dev server adds the Authorization header from .env and forwards the
// request. Keys never reach the client bundle (they are not VITE_-prefixed).
function keyProxy(env: Record<string, string>): Plugin {
  const routes = {
    "/api/openrouter": { target: "https://openrouter.ai/api/v1", key: env.OPENROUTER_API_KEY, headers: (k: string) => ({ Authorization: `Bearer ${k}`, "HTTP-Referer": "https://github.com/elvisun/newsjack", "X-Title": "News Desk Dealer" }) },
    "/api/typesafe": { target: "https://api.typesafe.ai", key: env.TYPESAFE_API_KEY, headers: (k: string) => ({ Authorization: `Bearer ${k}` }) },
  };
  return {
    name: "news-desk-dealer-key-proxy",
    configureServer(server) {
      // Primary path: calls multiplexed over Vite's HMR WebSocket. Browsers cap
      // HTTP/1.1 at six connections per host, which starved the fast model
      // while the slow model's long calls held the sockets.
      const inflight = new Map<string, AbortController>();
      server.ws.on("ndd:call", async (msg: { id: string; route: string; path: string; body: unknown }, client) => {
        const r = routes[msg.route as keyof typeof routes];
        const reply = (status: number, text: string) => client.send("ndd:result", { id: msg.id, status, text });
        if (!r) return reply(404, JSON.stringify({ error: `unknown route ${msg.route}` }));
        if (!r.key) return reply(401, JSON.stringify({ error: `${msg.route}: no key in .env` }));
        const ac = new AbortController();
        inflight.set(msg.id, ac);
        try {
          const up = await fetch(r.target + msg.path, { method: "POST", headers: { "content-type": "application/json", ...r.headers(r.key) }, body: JSON.stringify(msg.body), signal: ac.signal });
          reply(up.status, await up.text());
        } catch (e) {
          if (!ac.signal.aborted) reply(0, JSON.stringify({ error: String(e) }));
        } finally {
          inflight.delete(msg.id);
        }
      });
      server.ws.on("ndd:abort", (msg: { id: string }) => inflight.get(msg.id)?.abort());

      // Fallback path: plain HTTP, handy for curl.
      for (const [prefix, r] of Object.entries(routes)) {
        server.middlewares.use(prefix, async (req, res) => {
          if (!r.key) { res.statusCode = 401; res.end(JSON.stringify({ error: `${prefix}: no key in .env` })); return; }
          const chunks: Buffer[] = [];
          for await (const c of req) chunks.push(c as Buffer);
          const ac = new AbortController();
          res.on("close", () => { if (!res.writableFinished) ac.abort(); });
          try {
            const up = await fetch(r.target + (req.url ?? ""), {
              method: req.method, headers: { "content-type": "application/json", ...r.headers(r.key) },
              body: req.method === "GET" || req.method === "HEAD" ? undefined : Buffer.concat(chunks), signal: ac.signal,
            });
            res.statusCode = up.status;
            res.setHeader("content-type", up.headers.get("content-type") ?? "application/json");
            res.end(Buffer.from(await up.arrayBuffer()));
          } catch (e) {
            if (!res.headersSent) { res.statusCode = 502; res.end(JSON.stringify({ error: String(e) })); }
          }
        });
      }
    },
  };
}

export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd(), "");
  return { plugins: [react(), keyProxy(env)] };
});
