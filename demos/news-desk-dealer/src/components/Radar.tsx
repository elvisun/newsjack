// Radar (spider) chart. Draws the given values directly, no tween.
export function Radar({ title, keys, values, max = 4, size = 104, labels = true }: { title: string; keys: string[]; values: Record<string, number>; max?: number; size?: number; labels?: boolean }) {
  const shown = keys.map((k) => values[k] ?? 0);
  const cx = size / 2, cy = size / 2, r = size / 2 - (labels ? 22 : 4);
  const n = keys.length;
  const angle = (i: number) => -Math.PI / 2 + (i * 2 * Math.PI) / n;
  const pt = (i: number, v: number) => [cx + Math.cos(angle(i)) * r * (v / max), cy + Math.sin(angle(i)) * r * (v / max)];
  const poly = shown.map((v, i) => pt(i, v).map((x) => x.toFixed(1)).join(",")).join(" ");
  const rings = [0.25, 0.5, 0.75, 1];

  return (
    <div className="radar">
      <h4>{title}</h4>
      <svg viewBox={`0 0 ${size} ${size}`} width={size} height={size}>
        {rings.map((f) => <polygon key={f} className="radar-ring" points={keys.map((_, i) => pt(i, max * f).map((x) => x.toFixed(1)).join(",")).join(" ")} />)}
        {keys.map((_, i) => { const [x, y] = pt(i, max); return <line key={i} className="radar-axis" x1={cx} y1={cy} x2={x} y2={y} />; })}
        <polygon className="radar-area" points={poly} />
        {shown.map((v, i) => { const [x, y] = pt(i, v); return <circle key={i} className="radar-dot" cx={x} cy={y} r={1.6} />; })}
        {labels && keys.map((k, i) => {
          const [x, y] = pt(i, max * 1.22);
          const anchor = Math.abs(Math.cos(angle(i))) < 0.2 ? "middle" : Math.cos(angle(i)) > 0 ? "start" : "end";
          return <text key={k} className="radar-label" x={x} y={y + 2.5} textAnchor={anchor}>{k.replace(/_/g, " ")}</text>;
        })}
      </svg>
    </div>
  );
}
