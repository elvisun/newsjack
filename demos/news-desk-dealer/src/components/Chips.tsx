// Cost as poker chips. Each chip is worth CHIP_VALUE dollars; chips stack up
// to STACK_MAX high, then a new stack starts. Past the visible stacks the
// remainder is written as +N so the header never grows.
export const CHIP_VALUE = 0.25;
export const CHIP_LABEL = "25¢";
const STACK_MAX = 10;
const STACKS_MAX = 8;

export function chipCount(cost: number) { return Math.floor(cost / CHIP_VALUE + 1e-9); }

export function Chips({ cost }: { cost: number }) {
  const n = chipCount(cost);
  const visible = Math.min(n, STACK_MAX * STACKS_MAX);
  const stacks = Array.from({ length: STACKS_MAX }, (_, i) => Math.max(0, Math.min(STACK_MAX, visible - i * STACK_MAX)));
  const hidden = n - visible;
  return (
    <div className="chips">
      <div className="chip-stacks">
        {stacks.map((h, i) => (
          <div key={i} className="chip-stack">
            {Array.from({ length: h }, (_, j) => <div key={j} className="chip" style={{ bottom: j * 4 }} />)}
            {i === 0 && n === 0 && cost > 0 ? <div className="chip ghost" style={{ bottom: 0 }} /> : null}
          </div>
        ))}
      </div>
      <div className="chips-read">
        <b>{cost < 1 ? `$${cost.toFixed(4)}` : `$${cost.toFixed(2)}`}</b>
        <span>{n === 0 && cost > 0 ? `under one ${CHIP_LABEL} chip` : `${n.toLocaleString("en-US")} chip${n === 1 ? "" : "s"} at ${CHIP_LABEL}`}{hidden > 0 ? ` · +${hidden.toLocaleString("en-US")} off table` : ""}</span>
      </div>
    </div>
  );
}
