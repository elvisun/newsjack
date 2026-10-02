import type { NextConfig } from "next";

import { HSTS_HEADER_NAME, HSTS_HEADER_VALUE } from "./lib/security-headers";

const nextConfig: NextConfig = {
  images: {
    // Images embedded in Contentful-backed /insights posts.
    remotePatterns: [{ protocol: "https", hostname: "images.ctfassets.net" }],
  },
  async headers() {
    return [
      {
        source: "/:path*",
        headers: [
          {
            key: HSTS_HEADER_NAME,
            value: HSTS_HEADER_VALUE,
          },
        ],
      },
    ];
  },
};

export default nextConfig;
