import type { NextConfig } from "next";

import { HSTS_HEADER_NAME, HSTS_HEADER_VALUE } from "./lib/security-headers";

const nextConfig: NextConfig = {
  async redirects() {
    return [
      {
        source: "/claude-install",
        destination:
          "https://docs.google.com/document/d/1SAgTqLqgeWvpGctnjfv9m6wp_eU8_bc3WAJeZfXjNtI/edit",
        permanent: false,
      },
      {
        source: "/chatgpt-install",
        destination:
          "https://docs.google.com/document/d/1J7bX1frZ_aZBcRh-w-aJ3MCh-isSMSwVMMEaLIe8hrg/edit",
        permanent: false,
      },
    ];
  },
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
