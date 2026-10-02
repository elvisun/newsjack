import { useEffect, useRef } from "react";
import type { Adapter, AStamps, Client, Fit, Headline } from "../engine/types";
import { AXES, LEVELS, STANDING_LEVELS, newsworthiness } from "../questions";
import { Card, logoFor } from "./Card";
import { Radar } from "./Radar";
import { Gauge } from "./Gauge";
import { Chips } from "./Chips";

export type ColumnState = {
  phase: "idle" | "A" | "B" | "done" | "stopped";
  readA: number; readB: number; placed: number; unplaced: number; judgments: number;
  inputTokens: number; outputTokens: number; cost: number; latencies: number[];
  finishedAt: number | null;
  wire: { headline: Headline; stamps: AStamps }[];
  wireAt: number;
  stamps: Record<string, AStamps>;
  tally: { desk: Record<string, number>; story_type: Record<string, number>; axes: Record<string, number>; n: number };
  lanes: Record<string, { headline: Headline; fit: Fit }[]>;
  // Desk queue in wire order. The head is the story on the sorting slot; fits
  // is null until its set-B verdict lands. lastVerdict keeps the slot filled
  // once the queue drains.
  queue: { headline: Headline; fits: Fit[] | null }[];
  lastVerdict: { headline: Headline; fits: Fit[] } | null;
  verdicts: Record<string, number>; // headline id -> desks it landed on (0 = none)
  flights: { id: string; fits: Fit[]; seq: number }[];
  flightSeq: number;
  errors: number;
  lastError: string | null;
};

export const WIRE_SLOTS = 6;

export const emptyColumn = (): ColumnState => ({
  phase: "idle", readA: 0, readB: 0, placed: 0, unplaced: 0, judgments: 0, inputTokens: 0, outputTokens: 0, cost: 0,
  latencies: [], finishedAt: null, wire: [], wireAt: 0, stamps: {}, tally: { desk: {}, story_type: {}, axes: {}, n: 0 }, lanes: {},
  queue: [], lastVerdict: null, verdicts: {}, flights: [], flightSeq: 0, errors: 0, lastError: null,
});

const fmtMoney = (v: number) => (v < 1 ? `$${v.toFixed(4)}` : `$${v.toFixed(2)}`);
const fmtInt = (v: number) => v.toLocaleString("en-US");
const p50 = (xs: number[]) => { if (!xs.length) return 0; const s = [...xs].sort((a, b) => a - b); return s[Math.floor(s.length / 2)]; };
const DESKS = ["tech", "business", "finance", "health", "science", "policy", "consumer", "culture", "world"];

// Flies a card clone to each target pile. The clone starts wherever the story
// is on screen right now: the sorting slot if it is the head of the queue,
// otherwise its card on the wire.
function useFly(flights: ColumnState["flights"], colRef: React.RefObject<HTMLElement | null>, pileRefs: React.MutableRefObject<Record<string, HTMLDivElement | null>>) {
  const done = useRef(0);
  useEffect(() => {
    const col = colRef.current;
    if (!flights.length) { done.current = 0; return; }
    if (!col) return;
    const cs = getComputedStyle(col);
    for (const fl of flights) {
      if (fl.seq <= done.current) continue;
      done.current = fl.seq;
      const origin =
        col.querySelector(`.sorting [data-id="${fl.id}"] .card-face`) ??
        col.querySelector(`.wire-strip [data-id="${fl.id}"] .card-face`) ??
        col.querySelector(".sorting .card-face");
      const from = origin?.getBoundingClientRect();
      if (!from || !from.width) continue;
      for (const f of fl.fits) {
        const target = pileRefs.current[f.clientId]?.querySelector(".pile-stack");
        if (!target) continue;
        const stack = target.getBoundingClientRect();
        const cw = Math.min(92, stack.width - 28);
        const to = { left: stack.left + 4, top: stack.top + 14, width: cw, height: cw * 0.75 };
        const el = document.createElement("div");
        el.className = `fly ${f.tier}`;
        el.style.cssText = `left:${from.left}px;top:${from.top}px;width:${from.width}px;height:${from.height}px;--fly-color:${cs.getPropertyValue("--model")};--fly-soft:${cs.getPropertyValue("--model-soft")}`;
        document.body.appendChild(el);
        requestAnimationFrame(() => requestAnimationFrame(() => {
          const sx = to.width / from.width, sy = to.height / from.height;
          el.style.transform = `translate(${to.left - from.left}px, ${to.top - from.top}px) scale(${sx}, ${sy})`;
          el.style.opacity = "0.15";
        }));
        setTimeout(() => el.remove(), 800);
      }
    }
  }, [flights]);
}

export function ModelColumn({ adapter, state, total, elapsed, clients, rival, gaugeMax }: { adapter: Adapter; state: ColumnState; total: number; elapsed: number; clients: Client[]; rival: ColumnState; gaugeMax: number }) {
  const s = state;
  const colRef = useRef<HTMLElement>(null);
  const pileRefs = useRef<Record<string, HTMLDivElement | null>>({});
  useFly(s.flights, colRef, pileRefs);
  const head = s.queue[0] ?? s.lastVerdict ?? null;
  const waiting = Math.max(0, s.queue.length - 1);
  const live = s.phase === "A" || s.phase === "B";
  const statusOf = (id: string): { text: string; pending?: boolean } | undefined => {
    const v = s.verdicts[id];
    if (v !== undefined) return { text: v ? `${v} desk${v > 1 ? "s" : ""}` : "no desk" };
    if ((s.stamps[id]?.is_news ?? 1) < 0.5) return { text: "not news" };
    return { text: live ? "queued" : "unsorted", pending: live };
  };

  const jps = elapsed > 0 ? s.judgments / elapsed : 0;
  const newsy = Object.values(s.stamps).filter((st) => st.is_news >= 0.5).length;
  const phaseLabel = { idle: "idle", A: "reading and sorting", B: "sorting", done: "done", stopped: "stopped" }[s.phase];
  const perCallCost = s.readA + s.readB ? s.cost / (s.readA + s.readB) : 0;
  const remainingCalls = Math.max(0, total - s.readA) + Math.max(0, total - s.readB);
  const projected = s.cost + perCallCost * remainingCalls;

  return (
    <section className={`column ${adapter.id}`} ref={colRef}>
      <header className="col-head">
        <div className="col-title">
          {adapter.logo.endsWith(".svg") ? <div className="model-logo model-logo-mask" style={{ maskImage: `url(${adapter.logo})`, WebkitMaskImage: `url(${adapter.logo})` }} /> : <img className="model-logo" src={adapter.logo} alt="" />}
          <div className="model-name">
            <h2>{adapter.label}</h2>
            <span className="vendor">{adapter.vendor}</span>
            <span className={`phase phase-${s.phase}`}>
              {phaseLabel}
              {s.errors > 0 && <span className="note error" title={s.lastError ?? ""}> · {s.errors} failed call{s.errors > 1 ? "s" : ""}{s.lastError ? `: ${s.lastError.replace(/^\w+ \d+: /, "").replace(/^\{"error":"(.*)"\}$/, "$1").slice(0, 60)}` : ""}</span>}
              {s.phase === "stopped" && <span className="note"> · {s.readA + s.readB} of {total * 2} calls · full run ≈ {fmtMoney(projected)}{rival.phase === "done" ? ` vs ${fmtMoney(rival.cost)}` : ""}</span>}
            </span>
          </div>
          <div className="physical">
            <Gauge value={jps} max={gaugeMax} label="judgments / sec" />
            <Chips cost={s.cost} />
          </div>
        </div>
        <dl className="folio">
          <div><dt>read</dt><dd>{s.readA}<span className="of">/{total}</span></dd></div>
          <div><dt>sorted</dt><dd>{s.readB}<span className="of">/{newsy}</span></dd></div>
          <div><dt>judgments</dt><dd>{fmtInt(s.judgments)}</dd></div>
          <div><dt>j / sec</dt><dd>{jps.toFixed(0)}</dd></div>
          <div><dt>p50 call</dt><dd>{fmtInt(p50(s.latencies))}<span className="of">ms</span></dd></div>
          <div><dt>elapsed</dt><dd>{elapsed.toFixed(1)}<span className="of">s</span></dd></div>
          <div className="money"><dt>cost</dt><dd>{fmtMoney(s.cost)}</dd></div>
        </dl>
      </header>

      <div className="layer layer-a">
        <h3><span className="layer-tag">A</span> The wire</h3>
        <div className="wire-strip">
          {Array.from({ length: WIRE_SLOTS }, (_, i) => s.wire[i]).map((w, i) => w ? <Card key={w.headline.id} h={w.headline} stamps={w.stamps} className="wire-card dealt" status={statusOf(w.headline.id)} /> : <div key={`empty-${i}`} className="card card-empty"><div className="card-face" /><div className="card-stamps" /></div>)}
        </div>
      </div>

      <div className="layer layer-b">
        <h3><span className="layer-tag">B</span> The desks <span className="sub">{s.placed} placed · {s.unplaced} landed nowhere</span></h3>
        <div className="sorting">
          {head ? (
            <>
              <Card h={head.headline} className="slot-card" />
              <div className="sorting-text">
                <span className="slot-kicker">{head.fits ? "Sorted" : "Now sorting"} · {head.headline.source}{waiting ? ` · ${waiting} waiting` : ""}</span>
                <span className="wire-title">{head.headline.title}</span>
                <span className="slot-axes">{AXES.map((k) => <span key={k}><b>{k}</b> {LEVELS[k][s.stamps[head.headline.id]?.axes[k] ?? 0]}</span>)}</span>
                <span className={`arrow${head.fits || !live ? "" : " pending"}`}>
                  {head.fits
                    ? (head.fits.length ? `→ ${head.fits.map((f) => `${clients.find((c) => c.id === f.clientId)?.name} (${STANDING_LEVELS[f.standing]})`).join(", ")}` : "→ no desk")
                    : live ? `→ asking ${clients.length} desks…` : "→ stopped before the desks answered"}
                </span>
              </div>
              <div className="slot-radar">
                <Radar title="" keys={[...AXES]} values={s.stamps[head.headline.id]?.axes ?? {}} size={118} />
                <div className="slot-score"><b>{newsworthiness(s.stamps[head.headline.id]?.axes ?? ({} as any)).toFixed(1)}</b><span>/ 10</span></div>
              </div>
            </>
          ) : (
            <>
              <div className="card slot-card card-empty"><div className="card-face" /></div>
              <p className="wire-empty">Nothing sorted yet.</p>
              <div className="slot-radar"><Radar title="" keys={[...AXES]} values={{}} size={118} /></div>
            </>
          )}
        </div>
        <div className="piles">
          {clients.map((c) => {
            const items = [...(s.lanes[c.id] ?? [])].sort((a, b) => (a.fit.tier === b.fit.tier ? 0 : a.fit.tier === "pitch_ready" ? -1 : 1));
            const pitch = items.filter((i) => i.fit.tier === "pitch_ready").length;
            return (
              <div key={c.id} className="pile" ref={(el) => { pileRefs.current[c.id] = el; }}>
                <div className="pile-head">
                  <img src={logoFor(c.domain, 32)} alt="" />
                  <span className="pile-name">{c.name}</span>
                  <span className="pile-count">{pitch ? <b>{pitch}</b> : null}{items.length - pitch ? <span>+{items.length - pitch}</span> : null}</span>
                </div>
                <div className="pile-stack">
                  {items.slice(0, 6).reverse().map(({ headline, fit }, i, arr) => (
                    <div key={headline.id} className={`pile-card ${fit.tier} ${fit.action}`} style={{ transform: `translate(${(arr.length - 1 - i) * 5}px, ${(arr.length - 1 - i) * -2}px)`, zIndex: i }} title={`${headline.title}\nstanding ${STANDING_LEVELS[fit.standing]} · ${fit.tier} · ${fit.action}`}>
                      <Card h={headline} />
                      <span className="pile-stamp">{fit.action}</span>
                    </div>
                  ))}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    </section>
  );
}
