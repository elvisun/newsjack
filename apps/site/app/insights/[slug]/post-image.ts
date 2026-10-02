import { createElement } from "react";
import Image from "next/image.js";

import type { PostImage } from "../../../lib/contentful";

export function renderPostImage(image: PostImage | undefined) {
  // next/image needs intrinsic dimensions; Contentful only sets them on images.
  if (!image) return null;

  return createElement(
    "figure",
    { className: "mt-8" },
    createElement(Image, {
      src: image.src,
      alt: image.alt,
      width: image.width,
      height: image.height,
      sizes: "(min-width: 768px) 720px, 100vw",
      unoptimized: true,
      className: "h-auto w-full rounded-lg border border-ink/10",
    }),
    image.description
      ? createElement(
          "figcaption",
          {
            className: "mt-2 text-center font-mono text-xs text-ink/40",
          },
          image.description,
        )
      : null,
  );
}
