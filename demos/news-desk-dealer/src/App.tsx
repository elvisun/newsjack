import { useEffect, useMemo, useRef, useState } from "react";
import headlinesDoc from "../data/headlines.json";
import { CLIENTS } from "./clients";
import { currentEngine, makeAdapters, setSpeed } from "./engine/models";
import { runModel, type RunEvent } from "./engine/runner";
import type { Adapter, Headline } from "./engine/types";
import { ModelColumn, WIRE_SLOTS, emptyColumn, type ColumnState } from "./components/ModelColumn";
import { chipCount } from "./components/Chips";
import { clink, setSoundOn, snap, thud } from "./sound";

const HEADLINES = (headlinesDoc as { headlines: Headline[]; pulled_at: string }).headlines;
const PULLED_AT = (headlinesDoc as { pulled_at: string }).pulled_at;

function reduce(s: ColumnState, e: RunEvent, adapter: Adapter): ColumnState {
  const cost = (inTok: number, outTok: number) => (inTok * adapter.pricing.input_per_mtok + outTok * adapter.pricing.output_per_mtok) / 1e6;
  if (e.kind === "phase") return { ...s, phase: e.phase, finishedAt: e.phase === "done" || e.phase === "stopped" ? performance.now() : s.finishedAt };
  if (e.kind === "error") return { ...s, errors: s.errors + 1, lastError: e.message, queue: s.queue.filter((q) => q.headline.id !== e.headline.id) };
  if (e.kind === "a") {
    const tally = { ...s.tally, desk: { ...s.tally.desk }, story_type: { ...s.tally.story_type }, axes: { ...s.tally.axes }, n: s.tally.n + 1 };
    tally.desk[e.stamps.desk] = (tally.desk[e.stamps.desk] ?? 0) + 1;
    tally.story_type[e.stamps.story_type] = (tally.story_type[e.stamps.story_type] ?? 0) + 1;
    for (const [k, v] of Object.entries(e.stamps.axes)) tally.axes[k] = (tally.axes[k] ?? 0) + v;
    return {
      ...s, tally,
      readA: s.readA + 1,
      judgments: s.judgments + Object.keys(e.result.answers).length,
      inputTokens: s.inputTokens + e.result.usage.input_tokens,
      outputTokens: s.outputTokens + e.result.usage.output_tokens,
      cost: s.cost + cost(e.result.usage.input_tokens, e.result.usage.output_tokens),
      latencies: [...s.latencies, e.result.latency_ms],
      wire: [{ headline: e.headline, stamps: e.stamps }, ...s.wire].slice(0, WIRE_SLOTS),
      stamps: { ...s.stamps, [e.headline.id]: e.stamps },
    };
  }
  if (e.kind === "sorting") {
    // The story joins the desk queue in wire order. The slot shows the head of
    // the queue, so the story on the desk is always the oldest one still waiting.
    return { ...s, queue: [...s.queue, { headline: e.headline, fits: null }] };
  }
  const lanes = { ...s.lanes };
  let placed = 0;
  for (const f of e.fits) {
    if (f.tier === "watch") continue;
    lanes[f.clientId] = [{ headline: e.headline, fit: f }, ...(lanes[f.clientId] ?? [])];
    placed++;
  }
  const fits = e.fits.filter((f) => f.tier !== "watch");
  const flightSeq = s.flightSeq + 1;
  return {
    ...s, lanes,
    queue: s.queue.map((q) => (q.headline.id === e.headline.id ? { ...q, fits } : q)),
    verdicts: { ...s.verdicts, [e.headline.id]: fits.length },
    flights: [...s.flights, { id: e.headline.id, fits, seq: flightSeq }].slice(-24),
    flightSeq,
    readB: s.readB + 1,
    placed: s.placed + placed,
    unplaced: s.unplaced + (placed === 0 ? 1 : 0),
    judgments: s.judgments + Object.keys(e.result.answers).length,
    inputTokens: s.inputTokens + e.result.usage.input_tokens,
    outputTokens: s.outputTokens + e.result.usage.output_tokens,
    cost: s.cost + cost(e.result.usage.input_tokens, e.result.usage.output_tokens),
    latencies: [...s.latencies, e.result.latency_ms],
  };
}

// Runs once per render batch. A head whose verdict has already been on screen
// for a batch leaves the slot so the next story in wire order takes its place.
function advance(s: ColumnState): ColumnState {
  const head = s.queue[0];
  if (!head || head.fits === null) return s;
  return { ...s, queue: s.queue.slice(1), lastVerdict: { headline: head.headline, fits: head.fits } };
}

export default function App() {
  const engine = useMemo(currentEngine, []);
  const adapters = useMemo(() => makeAdapters(engine), [engine]);
  // Dial ceiling for the judgments-per-second gauge. Real Jev over the
  // WebSocket path sits around 1,000-1,300; Opus around 20.
  const gaugeMax = 2500;
  const [jev, setJev] = useState<ColumnState>(emptyColumn);
  const [opus, setOpus] = useState<ColumnState>(emptyColumn);
  const [running, setRunning] = useState(false);
  const [startedAt, setStartedAt] = useState<number | null>(null);
  const [now, setNow] = useState(performance.now());
  const [speed, setSpeedState] = useState(1);
  const [limit, setLimit] = useState(HEADLINES.length);
  const [stoppedReason, setStoppedReason] = useState<string | null>(null);
  const [sound, setSound] = useState(false);
  const chipsSeen = useRef({ jev: 0, opus: 0 });
  useEffect(() => { const n = chipCount(jev.cost); if (n > chipsSeen.current.jev) clink(); chipsSeen.current.jev = n; }, [jev.cost]);
  useEffect(() => { const n = chipCount(opus.cost); if (n > chipsSeen.current.opus) clink(); chipsSeen.current.opus = n; }, [opus.cost]);
  const ctrl = useRef<{ jev: AbortController; opus: AbortController } | null>(null);

  useEffect(() => {
    if (!running) return;
    const t = setInterval(() => setNow(performance.now()), 100);
    return () => clearInterval(t);
  }, [running]);

  const headlines = HEADLINES.slice(0, limit);

  async function start() {
    setJev(emptyColumn());
    setOpus(emptyColumn());
    setStoppedReason(null);
    const c = { jev: new AbortController(), opus: new AbortController() };
    ctrl.current = c;
    const t0 = performance.now();
    setStartedAt(t0);
    setRunning(true);
    // Events are queued and folded into state once per animation frame so the
    // UI renders at frame rate no matter how fast a model answers.
    const queues = { jev: [] as RunEvent[], opus: [] as RunEvent[] };
    let raf = 0;
    const cue = (batch: RunEvent[]) => {
      if (batch.some((e) => e.kind === "a")) snap();
      if (batch.some((e) => e.kind === "b" && e.fits.some((f) => f.tier !== "watch"))) setTimeout(thud, 650);
    };
    const flush = () => {
      raf = 0;
      cue(queues.jev); cue(queues.opus);
      if (queues.jev.length) { const batch = queues.jev.splice(0); setJev((s) => batch.reduce((acc, e) => reduce(acc, e, adapters.jev), advance(s))); }
      if (queues.opus.length) { const batch = queues.opus.splice(0); setOpus((s) => batch.reduce((acc, e) => reduce(acc, e, adapters.opus), advance(s))); }
    };
    const enqueue = (k: "jev" | "opus", e: RunEvent) => { queues[k].push(e); if (!raf) raf = requestAnimationFrame(flush); };
    const jevRun = runModel(adapters.jev, headlines, CLIENTS, { concurrency: adapters.jev.concurrency, signal: c.jev.signal, onEvent: (e) => enqueue("jev", e) });
    const opusRun = runModel(adapters.opus, headlines, CLIENTS, { concurrency: adapters.opus.concurrency, signal: c.opus.signal, onEvent: (e) => enqueue("opus", e) });
    await jevRun;
    // Jev is done: stop Opus so no more tokens are spent.
    if (!c.opus.signal.aborted) {
      c.opus.abort();
      setStoppedReason("Stopped when Jev finished");
    }
    await opusRun;
    flush();
    setRunning(false);
  }
  function stop() {
    ctrl.current?.jev.abort();
    ctrl.current?.opus.abort();
    setStoppedReason("Stopped by hand");
    setRunning(false);
  }

  const elapsed = (s: ColumnState) => startedAt == null ? 0 : ((s.finishedAt ?? now) - startedAt) / 1000;

  return (
    <div className="page">
      <header className="masthead">
        <div className="masthead-row">
          <div className="masthead-left">
            <h1>News Desk Dealer</h1>
            <a className="byline" href="https://medialyst.com" target="_blank" rel="noreferrer"><span>by</span><b>Medialyst</b></a>
            <p className="dateline">
              Edition of {new Date(PULLED_AT).toLocaleDateString("en-US", { day: "numeric", month: "long", year: "numeric" })} ·
              {" "}{headlines.length} headlines off Google News · {CLIENTS.length} desks · {engine === "live" ? "live: Jev direct, Opus via OpenRouter" : "mock engine, no keys wired"}
              {stoppedReason ? <span className="stopped"> · {stoppedReason}</span> : null}
            </p>
          </div>
          <div className="controls">
            <label>Headlines <input type="number" min={10} max={HEADLINES.length} value={limit} disabled={running} onChange={(e) => setLimit(Math.max(10, Math.min(HEADLINES.length, +e.target.value || 10)))} /></label>
            <label>Speed
              <select value={speed} disabled={running} onChange={(e) => { const v = +e.target.value; setSpeedState(v); setSpeed(v); }}>
                <option value={1}>1× real</option><option value={4}>4×</option><option value={10}>10×</option>
              </select>
            </label>
            <button className={`btn ghost${sound ? " on" : ""}`} onClick={() => { const v = !sound; setSound(v); setSoundOn(v); if (v) snap(); }} title="Card snaps, chip clinks, pile thuds">{sound ? "Sound on" : "Sound off"}</button>
            {running ? <button className="btn stop" onClick={stop}>Stop</button> : <button className="btn" onClick={start}>Deal the wire</button>}
          </div>
        </div>
      </header>

      <main className="columns">
        <ModelColumn adapter={adapters.jev} state={jev} total={headlines.length} elapsed={elapsed(jev)} clients={CLIENTS} rival={opus} gaugeMax={gaugeMax} />
        <ModelColumn adapter={adapters.opus} state={opus} total={headlines.length} elapsed={elapsed(opus)} clients={CLIENTS} rival={jev} gaugeMax={gaugeMax} />
      </main>

      <footer className="colophon">
        Set A stamps every headline with a desk, a story type, and six scores that draw its radar: magnitude, velocity, novelty, window, heat, risk. Set B scores each story per desk on standing and journalist shape, then tiers it. A story can land on several desks or none. Rubrics follow the public Newsjack skills. The desks are public companies used for illustration, not clients.
      </footer>
    </div>
  );
}
