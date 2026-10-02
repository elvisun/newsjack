import shots from "../../data/shots.json";
import type { AStamps, Headline } from "../engine/types";
import { LEVELS } from "../questions";

const SHOT_IDS = new Set((shots as { ids: string[] }).ids);
export const logoFor = (domain: string, size = 64) => `https://www.google.com/s2/favicons?domain=${domain}&sz=${size}`;

// A headline as a small press card: the page screenshot when we have one,
// otherwise a typeset stand-in with the outlet's logo.
export function Card({ h, stamps, className = "", status }: { h: Headline; stamps?: AStamps; className?: string; status?: { text: string; pending?: boolean } }) {
  const shot = SHOT_IDS.has(h.id);
  return (
    <figure className={`card ${className}${stamps && stamps.axes.risk >= 4 ? " unsafe" : ""}`} data-id={h.id}>
      <div className="card-face">
        {status ? <span className={`card-status${status.pending ? " pending" : ""}`}>{status.text}</span> : null}
        {shot ? <img src={`/shots/${h.id}.jpg`} alt="" loading="lazy" /> : (
          <div className="card-type">
            <div className="card-mast"><img src={logoFor(h.domain, 32)} alt="" /><span>{h.source}</span></div>
            <div className="card-head">{h.title}</div>
          </div>
        )}
        <img className="card-logo" src={logoFor(h.domain, 32)} alt="" />
      </div>
      {stamps && (
        <figcaption className="card-stamps">
          <b>{stamps.desk}</b><em>{LEVELS.magnitude[stamps.axes.magnitude]}</em><u>{LEVELS.window[stamps.axes.window]}</u>
          {stamps.axes.risk >= 4 ? <s>avoid</s> : null}
        </figcaption>
      )}
    </figure>
  );
}
