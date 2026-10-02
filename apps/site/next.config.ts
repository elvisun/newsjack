import type { NextConfig } from "next";

import { HSTS_HEADER_NAME, HSTS_HEADER_VALUE } from "./lib/security-headers";

const nextConfig: NextConfig = {
  async rewrites() {
    const spaceId = process.env.CONTENTFUL_SPACE_ID;
    if (!spaceId) return [];

    return [
      {
        source: "/insights/media/:assetId/:token/:filename",
        destination: `https://images.ctfassets.net/${encodeURIComponent(spaceId)}/:assetId/:token/:filename`,
      },
    ];
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
