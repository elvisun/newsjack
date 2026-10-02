const AUTHORS = [
  { handle: "elvissun", href: "https://x.com/elvissun" },
  { handle: "prcarly", href: "https://www.linkedin.com/in/prcarly" },
];

// "@elvissun & @prcarly", each linked to where that person posts.
export function Credits() {
  return (
    <>
      {AUTHORS.map((author, index) => (
        <span key={author.handle}>
          {index > 0 && " & "}
          <a
            className="transition-colors hover:text-accent"
            href={author.href}
            rel="noreferrer"
            target="_blank"
          >
            @{author.handle}
          </a>
        </span>
      ))}
    </>
  );
}
