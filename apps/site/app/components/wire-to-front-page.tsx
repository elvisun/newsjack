"use client";

import { useEffect, useRef, useState } from "react";

// Hero motion: the news wire as a flood of greeked type, one line catching
// fire, fanning into angles, matching a reporter, then every line of type
// settling into a front page with your story as the lead.

const W = 520;
const H = 600;
const LOOP = 15;

const INK = [26, 26, 26];
const PAPER = [249, 248, 246];
const NEWSPRINT = [243, 241, 236];
const ACCENT = "#E05A47";

const SCENE_TOP = 48;
const SCENE_BOTTOM = 544;
const LOCK_Y = 300;
const SIGNAL_CENTER = { x: 260, y: 236 };
const NODE_Y = 372;
const NODE_COUNT = 7;
const MATCH = 4;

// Beats drive the HTML chrome; the canvas reads the clock directly.
const BEATS = [
  { at: 0, step: 0, status: "Reading the wire" },
  { at: 3.6, step: 1, status: "Found a story worth riding" },
  { at: 6.4, step: 2, status: "Fit-checking reporters" },
  { at: 9.0, step: 3, status: "You hit send" },
  { at: 10.6, step: 3, status: "Your story, above the fold" },
  { at: 13.8, step: 3, status: "Back to the wire" },
];
const STEPS = [
  { label: "Detect", seconds: 3.6 },
  { label: "Angle", seconds: 2.8 },
  { label: "Fit-check", seconds: 2.6 },
  { label: "Coverage", seconds: 4.8 },
];

type Target = { x: number; y: number; w: number; h: number; alpha: number };
type Bar = {
  x: number;
  w: number;
  y0: number;
  speed: number;
  alpha: number;
  target: Target | null;
  delay: number;
  back: number;
};

const clamp = (v: number, lo = 0, hi = 1) => Math.min(hi, Math.max(lo, v));
const lerp = (a: number, b: number, t: number) => a + (b - a) * t;
const easeOut = (t: number) => 1 - (1 - t) ** 3;
const easeInOut = (t: number) =>
  t < 0.5 ? 4 * t * t * t : 1 - (-2 * t + 2) ** 3 / 2;
const window01 = (t: number, start: number, length: number) =>
  clamp((t - start) / length);

function seeded(seed: number) {
  return () => {
    seed |= 0;
    seed = (seed + 0x6d2b79f5) | 0;
    let r = Math.imul(seed ^ (seed >>> 15), 1 | seed);
    r = (r + Math.imul(r ^ (r >>> 7), 61 | r)) ^ r;
    return ((r ^ (r >>> 14)) >>> 0) / 4294967296;
  };
}

function frontPage(): Target[] {
  const targets: Target[] = [
    { x: 40, y: 58, w: 70, h: 2, alpha: 0.35 },
    { x: 410, y: 58, w: 70, h: 2, alpha: 0.35 },
    { x: 150, y: 68, w: 220, h: 12, alpha: 0.9 },
    { x: 40, y: 88, w: 440, h: 1.5, alpha: 0.8 },
    { x: 40, y: 92, w: 440, h: 1, alpha: 0.5 },
    { x: 40, y: 128, w: 250, h: 14, alpha: 0.88 },
    { x: 56, y: 158.5, w: 110, h: 3, alpha: 0.4 },
  ];
  const rand = seeded(7);
  for (let col = 0; col < 3; col += 1) {
    let line = 0;
    for (let y = 178; y <= SCENE_BOTTOM - 8; y += 9) {
      line += 1;
      if (line % 8 === 0) continue;
      const short = line % 8 === 7;
      targets.push({
        x: 40 + col * 150,
        y,
        w: short ? 50 + rand() * 50 : 140,
        h: 3,
        alpha: 0.2,
      });
    }
  }
  return targets;
}

function buildScene() {
  const rand = seeded(42);
  const columns = [
    { x: 36, w: 96, speed: 16 },
    { x: 152, w: 96, speed: 25 },
    { x: 268, w: 96, speed: 13 },
    { x: 384, w: 100, speed: 20 },
  ];
  const bars: Bar[] = [];
  for (const column of columns) {
    for (let y = 0; y < SCENE_BOTTOM - SCENE_TOP + 24; y += 13) {
      bars.push({
        x: column.x,
        w: column.w * (0.35 + rand() * 0.65),
        y0: y,
        speed: column.speed,
        alpha: 0.12 + rand() * 0.22,
        target: null,
        delay: 0,
        back: rand(),
      });
    }
  }
  const targets = frontPage();
  const order = bars.map((_, i) => i).sort(() => rand() - 0.5);
  targets.forEach((target, i) => {
    const bar = bars[order[i]];
    bar.target = target;
    bar.delay = ((target.y - SCENE_TOP) / 480) * 0.9 + rand() * 0.15;
  });
  return bars;
}

function fieldY(bar: Bar, clock: number) {
  const range = SCENE_BOTTOM - SCENE_TOP + 24;
  const y = (((bar.y0 - bar.speed * clock) % range) + range) % range;
  return SCENE_TOP - 12 + y;
}

function edgeFade(y: number) {
  return clamp((y - SCENE_TOP) / 28) * clamp((SCENE_BOTTOM - y) / 28);
}

function nodeX(k: number) {
  return SIGNAL_CENTER.x + (k - (NODE_COUNT - 1) / 2) * 56;
}

function draw(
  ctx: CanvasRenderingContext2D,
  bars: Bar[],
  clock: number,
  t: number,
  chosen: { index: number; x: number; y: number; w: number } | null,
) {
  // A press roller: newsprint rolls down over the ink on the way in, and
  // ink rolls back down over it on the way out.
  const printIn = easeInOut(window01(t, 9, 1));
  const printOut = easeInOut(window01(t, 13.8, 0.9));
  const paperTop = printOut * (H + 20);
  const paperBottom = printIn * (H + 20);
  const paperAt = (y: number) =>
    printIn === 0 ? 0 : clamp((paperBottom - y) / 16) * clamp((y - paperTop) / 16);
  const toneAt = (y: number, alpha: number) => {
    const p = paperAt(y);
    return `rgba(${INK.map((c, i) => Math.round(lerp(PAPER[i], c, p))).join(",")},${alpha})`;
  };

  ctx.fillStyle = `rgb(${INK.join(",")})`;
  ctx.fillRect(0, 0, W, H);
  if (paperBottom > paperTop) {
    ctx.fillStyle = `rgb(${NEWSPRINT.join(",")})`;
    ctx.fillRect(0, paperTop, W, paperBottom - paperTop);
  }
  const rollerY = printOut > 0 ? paperTop : paperBottom;
  if ((printIn > 0 && printIn < 1) || (printOut > 0 && printOut < 1)) {
    ctx.fillStyle = ACCENT;
    ctx.fillRect(0, rollerY - 1, W, 2);
  }

  const scanning = t >= 0.4 && t < 3;
  const scanY = lerp(SCENE_TOP, LOCK_Y, easeOut(window01(t, 0.4, 2.2)));
  const settled = t >= 2.6 && t < 13.8;
  const dimAfterLock = lerp(1, 0.32, window01(t, 2.6, 0.6));
  const unassignedFade =
    (1 - window01(t, 9, 0.5)) + window01(t, 13.8, 0.8);

  bars.forEach((bar, index) => {
    if (chosen && index === chosen.index && t >= 2.6 && t < 14.2) return;
    const fy = fieldY(bar, clock);
    let fieldAlpha = bar.alpha * edgeFade(fy);
    if (t < 2.6) {
      if (scanning && fy < scanY) fieldAlpha *= 0.55;
      if (scanning && Math.abs(fy - scanY) < 7) fieldAlpha = Math.min(1, fieldAlpha + 0.45);
    } else if (settled) {
      fieldAlpha *= dimAfterLock;
    }

    if (!bar.target) {
      ctx.fillStyle = toneAt(fy, fieldAlpha * clamp(unassignedFade));
      ctx.fillRect(bar.x, fy, bar.w, 3);
      return;
    }

    const forward = easeOut(window01(t, 9.2 + bar.delay, 0.9));
    const back = easeInOut(window01(t, 13.8 + bar.back * 0.3, 0.8));
    const e = forward * (1 - back);
    const target = bar.target;
    const y = lerp(fy, target.y, e);
    ctx.fillStyle = toneAt(y, lerp(fieldAlpha, target.alpha, e));
    ctx.fillRect(
      lerp(bar.x, target.x, e),
      y,
      lerp(bar.w, target.w, e),
      lerp(3, target.h, e),
    );
  });

  if (scanning) {
    const alpha = 1 - window01(t, 2.6, 0.4);
    ctx.fillStyle = `rgba(224,90,71,${0.9 * alpha})`;
    ctx.fillRect(20, scanY + 1, W - 40, 1);
    ctx.fillRect(14, scanY - 2, 6, 7);
  }

  // Fit-check: reporters as nodes, one line out to each.
  if (t >= 6.4 && t < 10.4) {
    const fadeRejects = 1 - window01(t, 9, 0.4);
    for (let k = 0; k < NODE_COUNT; k += 1) {
      const appear = easeOut(window01(t, 6.4 + k * 0.07, 0.35));
      const checkAt = 7 + k * 0.17;
      const reach = window01(t, checkAt, 0.12);
      const decided = t >= checkAt + 0.2;
      const x = nodeX(k);
      const isMatch = k === MATCH;
      if (isMatch && t >= 9) continue;

      if (reach > 0) {
        const lineAlpha = isMatch
          ? 0.8
          : 0.6 * (1 - window01(t, checkAt + 0.2, 0.3));
        ctx.strokeStyle = `rgba(224,90,71,${lineAlpha * fadeRejects})`;
        ctx.lineWidth = 1;
        ctx.beginPath();
        ctx.moveTo(SIGNAL_CENTER.x, SIGNAL_CENTER.y + 6);
        ctx.lineTo(
          lerp(SIGNAL_CENTER.x, x, reach),
          lerp(SIGNAL_CENTER.y + 6, NODE_Y - 10, reach),
        );
        ctx.stroke();
      }

      const alpha = appear * (isMatch ? 1 : fadeRejects);
      ctx.lineWidth = 1.5;
      if (decided && isMatch) {
        ctx.fillStyle = ACCENT;
        ctx.beginPath();
        ctx.arc(x, NODE_Y, 9, 0, Math.PI * 2);
        ctx.fill();
        const pulse = window01(t, checkAt + 0.2, 0.8);
        ctx.strokeStyle = `rgba(224,90,71,${0.6 * (1 - pulse)})`;
        ctx.beginPath();
        ctx.arc(x, NODE_Y, 9 + pulse * 14, 0, Math.PI * 2);
        ctx.stroke();
      } else {
        const dim = decided ? 0.35 : 1;
        ctx.strokeStyle = `rgba(249,248,246,${0.55 * alpha * dim})`;
        ctx.beginPath();
        ctx.arc(x, NODE_Y, 9 * appear, 0, Math.PI * 2);
        ctx.stroke();
        if (decided) {
          ctx.beginPath();
          ctx.moveTo(x - 3.5, NODE_Y - 3.5);
          ctx.lineTo(x + 3.5, NODE_Y + 3.5);
          ctx.moveTo(x + 3.5, NODE_Y - 3.5);
          ctx.lineTo(x - 3.5, NODE_Y + 3.5);
          ctx.stroke();
        }
      }
    }
  }

  // The matched reporter becomes the lead story's byline.
  if (t >= 9 && t < 14.6) {
    const move = easeInOut(window01(t, 9.4, 1));
    const fade = 1 - window01(t, 13.8, 0.8);
    ctx.fillStyle = `rgba(224,90,71,${fade})`;
    ctx.beginPath();
    ctx.arc(
      lerp(nodeX(MATCH), 45, move),
      lerp(NODE_Y, 160, move),
      lerp(9, 4.5, move),
      0,
      Math.PI * 2,
    );
    ctx.fill();
  }

  // The signal: one line of the wire, lit, then turned into the lead.
  if (chosen && t >= 2.6 && t < 14.6) {
    const ignite = window01(t, 2.6, 0.4);
    const toCenter = easeInOut(window01(t, 3.6, 0.8));
    const toLead = easeInOut(window01(t, 9.2, 1));
    const fade = 1 - window01(t, 13.8, 0.8);
    const base = {
      x: lerp(chosen.x, SIGNAL_CENTER.x - 130, toCenter),
      y: lerp(chosen.y, SIGNAL_CENTER.y, toCenter),
      w: lerp(chosen.w, 260, toCenter),
      h: lerp(3, 8, Math.max(ignite * 0.4, toCenter)),
    };
    const lead = {
      x: lerp(base.x, 40, toLead),
      y: lerp(base.y, 108, toLead),
      w: lerp(base.w, 380, toLead),
      h: lerp(base.h, 14, toLead),
    };

    // Angles: the line fans into three takes; only one survives.
    const fan = easeOut(window01(t, 4.4, 0.6));
    const cull = easeInOut(window01(t, 5.4, 0.6));
    if (fan > 0 && cull < 1) {
      [-1, 1].forEach((side) => {
        const alpha = fan * (1 - cull);
        ctx.save();
        ctx.translate(
          SIGNAL_CENTER.x + side * cull * 40,
          SIGNAL_CENTER.y + side * (30 * fan + cull * 30),
        );
        ctx.rotate(((side * 7 * Math.PI) / 180) * fan);
        ctx.fillStyle = `rgba(224,90,71,${0.75 * alpha})`;
        const w = side < 0 ? 236 : 214;
        ctx.fillRect(-w / 2, -4, w, 8);
        ctx.restore();
      });
    }

    const signalTone =
      ignite < 1
        ? `rgba(${Math.round(lerp(PAPER[0], 224, ignite))},${Math.round(lerp(PAPER[1], 90, ignite))},${Math.round(lerp(PAPER[2], 71, ignite))},${lerp(0.4, 1, ignite) * fade})`
        : `rgba(224,90,71,${fade})`;
    ctx.fillStyle = signalTone;
    ctx.fillRect(lead.x, lead.y - lead.h / 2 + 1.5, lead.w, lead.h);
  }
}

export function WireToFrontPage() {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const [beat, setBeat] = useState(0);
  const [loop, setLoop] = useState(0);

  useEffect(() => {
    const canvas = canvasRef.current;
    const ctx = canvas?.getContext("2d");
    if (!canvas || !ctx) return;

    const bars = buildScene();
    const reduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    let chosen: { index: number; x: number; y: number; w: number } | null = null;
    let lastLoop = -1;
    let lastBeat = -1;
    let elapsed = 0;
    let previous = performance.now();
    let frame = 0;
    let visible = true;

    const resize = () => {
      const dpr = Math.min(window.devicePixelRatio || 1, 2);
      const width = canvas.clientWidth;
      canvas.width = Math.round(width * dpr);
      canvas.height = Math.round(width * (H / W) * dpr);
      ctx.setTransform((canvas.width / W), 0, 0, canvas.height / H, 0, 0);
    };

    const pick = (clock: number) => {
      let best = -1;
      let distance = Infinity;
      bars.forEach((bar, index) => {
        if (bar.x < 100 || bar.x > 300) return;
        const d = Math.abs(fieldY(bar, clock) - LOCK_Y);
        if (d < distance) {
          distance = d;
          best = index;
        }
      });
      const bar = bars[best];
      return { index: best, x: bar.x, y: fieldY(bar, clock), w: bar.w };
    };

    const render = (clock: number) => {
      const t = clock % LOOP;
      const loopIndex = Math.floor(clock / LOOP);
      if (loopIndex !== lastLoop) {
        lastLoop = loopIndex;
        chosen = null;
        setLoop(loopIndex);
      }
      if (!chosen && t >= 2.6) chosen = pick(clock);
      let current = 0;
      BEATS.forEach((b, i) => {
        if (t >= b.at) current = i;
      });
      if (current !== lastBeat) {
        lastBeat = current;
        setBeat(current);
      }
      draw(ctx, bars, clock, t, chosen);
    };

    const tick = (now: number) => {
      elapsed += Math.min(now - previous, 100) / 1000;
      previous = now;
      render(elapsed);
      frame = requestAnimationFrame(tick);
    };

    const start = () => {
      if (reduced || frame || !visible || document.hidden) return;
      previous = performance.now();
      frame = requestAnimationFrame(tick);
    };
    const stop = () => {
      cancelAnimationFrame(frame);
      frame = 0;
    };

    resize();
    const observer = new ResizeObserver(() => {
      resize();
      if (reduced) render(12);
      else render(elapsed);
    });
    observer.observe(canvas);
    const io = new IntersectionObserver(([entry]) => {
      visible = entry.isIntersecting;
      if (visible) start();
      else stop();
    });
    io.observe(canvas);
    const onVisibility = () => (document.hidden ? stop() : start());
    document.addEventListener("visibilitychange", onVisibility);

    if (reduced) {
      render(12);
    } else {
      start();
    }

    return () => {
      stop();
      observer.disconnect();
      io.disconnect();
      document.removeEventListener("visibilitychange", onVisibility);
    };
  }, []);

  const { step, status } = BEATS[beat];
  const onPaper = beat === 3 || beat === 4;

  return (
    <div className="nj-rise relative overflow-hidden rounded-lg bg-ink shadow-terminal">
      <canvas
        aria-label="Animation: newsjack reads the news wire, finds a story worth riding, drafts angles, fit-checks reporters, and lands your story on the front page."
        className="block aspect-[26/30] w-full"
        ref={canvasRef}
        role="img"
      />
      <div
        className={`pointer-events-none absolute inset-x-0 top-0 flex items-center justify-between px-6 pt-5 font-mono text-[10px] tracking-[0.2em] uppercase transition-colors duration-700 ${
          onPaper ? "text-ink/60" : "text-page/60"
        }`}
      >
        <span className="flex items-center gap-2">
          <span className="text-accent">➜</span>
          {status}
          <span className="nj-blink inline-block h-3 w-1.5 bg-accent" />
        </span>
      </div>
      {beat === 4 && (
        <div className="nj-stamp pointer-events-none absolute top-[27%] right-[7%] border-2 border-accent px-3 py-1.5 font-mono text-[10px] font-bold tracking-[0.25em] text-accent uppercase">
          Above the fold
        </div>
      )}
      <ol
        className={`pointer-events-none absolute inset-x-0 bottom-0 grid grid-cols-4 gap-4 px-6 pb-5 font-mono text-[9px] tracking-[0.2em] uppercase transition-colors delay-700 duration-500 ${
          onPaper ? "text-ink/50" : "text-page/50"
        }`}
      >
        {STEPS.map((item, index) => (
          <li key={item.label}>
            <span className="relative block h-px overflow-hidden bg-current/30">
              <span
                className={`absolute inset-y-0 left-0 bg-accent ${index === step ? "nj-fill" : ""}`}
                key={`${loop}-${index}-${index === step}`}
                style={{
                  width: index < step ? "100%" : index === step ? undefined : 0,
                  animationDuration: `${item.seconds}s`,
                }}
              />
            </span>
            <span
              className={`mt-2 block ${index === step ? "text-accent" : ""}`}
            >
              {item.label}
            </span>
          </li>
        ))}
      </ol>
    </div>
  );
}
