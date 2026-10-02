export function SectionHeading({
  id,
  number,
  title,
  subtitle,
}: {
  id?: string;
  number: string;
  title: string;
  subtitle?: string;
}) {
  return (
    <div className="relative mb-16 flex flex-wrap items-start justify-between gap-4 pt-8">
      <span
        aria-hidden="true"
        className="nj-px-rule absolute inset-x-0 top-0 h-px bg-ink/10"
      />
      <div className="nj-px-title flex items-baseline gap-4">
        <span className="font-mono text-xs text-ink/40">{number}</span>
        <h2
          className="font-serif text-[clamp(2rem,5vw,3.5rem)] leading-none tracking-[-0.02em] lowercase italic"
          id={id}
        >
          {title}
        </h2>
      </div>
      {subtitle && (
        <p className="mt-3 max-w-[420px] leading-relaxed text-ink/60 italic">
          {subtitle}
        </p>
      )}
    </div>
  );
}
