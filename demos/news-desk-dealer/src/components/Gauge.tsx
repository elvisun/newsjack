// A tachometer for judgments per second. The needle is the only thing that
// moves; the dial is static so the header never shifts.
export function Gauge({ value, max, label }: { value: number; max: number; label: string }) {
  const frac = Math.max(0, Math.min(1, value / max));
  const angle = -90 + frac * 180;
  const ticks = 6;
  return (
    <div className="gauge">
      <svg viewBox="4 4 92 50" width="88" height="48">
        <path className="gauge-arc" d="M 8 50 A 42 42 0 0 1 92 50" />
        <path className="gauge-arc gauge-hot" d="M 66.7 8.3 A 42 42 0 0 1 92 50" />
        {Array.from({ length: ticks }, (_, i) => {
          const a = (-90 + (i / (ticks - 1)) * 180) * Math.PI / 180;
          const x1 = 50 + Math.sin(a) * 42, y1 = 50 - Math.cos(a) * 42, x2 = 50 + Math.sin(a) * 36, y2 = 50 - Math.cos(a) * 36;
          return <line key={i} className="gauge-tick" x1={x1} y1={y1} x2={x2} y2={y2} />;
        })}
        <g className="gauge-needle" style={{ transform: `rotate(${angle}deg)`, transformOrigin: "50px 50px" }}>
          <line x1="50" y1="50" x2="50" y2="12" />
          <circle cx="50" cy="50" r="3" />
        </g>
        <text className="gauge-max" x="50" y="34" textAnchor="middle">0 – {max >= 1000 ? `${max / 1000}k` : max}</text>
      </svg>
      <div className="gauge-read"><b>{Math.round(value).toLocaleString("en-US")}</b><span>{label}</span></div>
    </div>
  );
}
