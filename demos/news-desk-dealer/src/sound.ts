// Synthesised table sounds, no audio files. Off by default; the toggle creates
// the AudioContext on a user gesture. Each sound is rate-limited so Jev's
// speed turns into a fast riffle rather than a wall of noise.

let ctx: AudioContext | null = null;
let noise: AudioBuffer | null = null;
let on = false;
const last: Record<string, number> = {};

export function setSoundOn(v: boolean) {
  on = v;
  if (v) { ctx ??= new AudioContext(); void ctx.resume(); }
}
export const soundOn = () => on;

function ready(kind: string, gapMs: number): AudioContext | null {
  if (!on || !ctx) return null;
  const t = performance.now();
  if (t - (last[kind] ?? 0) < gapMs) return null;
  last[kind] = t;
  return ctx;
}
function noiseBuffer(c: AudioContext) {
  if (noise) return noise;
  noise = c.createBuffer(1, Math.floor(c.sampleRate * 0.12), c.sampleRate);
  const d = noise.getChannelData(0);
  for (let i = 0; i < d.length; i++) d[i] = Math.random() * 2 - 1;
  return noise;
}

// A card snapped onto the table: a short band-passed noise burst.
export function snap() {
  const c = ready("snap", 40); if (!c) return;
  const t = c.currentTime;
  const src = c.createBufferSource(); src.buffer = noiseBuffer(c);
  const f = c.createBiquadFilter(); f.type = "bandpass"; f.frequency.value = 1500 + Math.random() * 900; f.Q.value = 1.4;
  const g = c.createGain(); g.gain.setValueAtTime(0.5, t); g.gain.exponentialRampToValueAtTime(0.001, t + 0.06);
  src.connect(f).connect(g).connect(c.destination); src.start(t); src.stop(t + 0.08);
}

// A chip dropped on a stack: two bright partials with a fast decay.
export function clink() {
  const c = ready("clink", 70); if (!c) return;
  const t = c.currentTime;
  const g = c.createGain(); g.gain.setValueAtTime(0.12, t); g.gain.exponentialRampToValueAtTime(0.001, t + 0.1);
  g.connect(c.destination);
  for (const base of [2500, 3700]) {
    const o = c.createOscillator(); o.type = "sine"; o.frequency.value = base * (0.97 + Math.random() * 0.06);
    o.connect(g); o.start(t); o.stop(t + 0.1);
  }
}

// A card landing on a pile: a low, damped thump.
export function thud() {
  const c = ready("thud", 60); if (!c) return;
  const t = c.currentTime;
  const o = c.createOscillator(); o.type = "sine";
  o.frequency.setValueAtTime(150, t); o.frequency.exponentialRampToValueAtTime(55, t + 0.1);
  const g = c.createGain(); g.gain.setValueAtTime(0.3, t); g.gain.exponentialRampToValueAtTime(0.001, t + 0.13);
  o.connect(g).connect(c.destination); o.start(t); o.stop(t + 0.15);
}
