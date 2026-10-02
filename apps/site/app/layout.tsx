import type { Metadata } from "next";
import { DM_Sans, IBM_Plex_Mono, Newsreader } from "next/font/google";
import {
  OG_IMAGE,
  SITE_DESCRIPTION,
  SITE_NAME,
  SITE_TITLE,
  SITE_URL,
} from "../lib/site";
import { DevAnnotations } from "./components/dev-annotations";
import "./globals.css";

const newsreader = Newsreader({
  subsets: ["latin"],
  style: ["normal", "italic"],
  axes: ["opsz"],
  variable: "--font-newsreader",
});
const dmSans = DM_Sans({
  subsets: ["latin"],
  axes: ["opsz"],
  variable: "--font-dm-sans",
});
const plexMono = IBM_Plex_Mono({
  subsets: ["latin"],
  weight: ["400", "500", "700"],
  variable: "--font-plex-mono",
});

// Pages set their own canonical and openGraph via pageMetadata(). X falls back
// to the og: tags, so twitter only needs the card type.
export const metadata: Metadata = {
  metadataBase: new URL(SITE_URL),
  title: { default: SITE_TITLE, template: `%s | ${SITE_NAME}` },
  description: SITE_DESCRIPTION,
  openGraph: {
    title: SITE_TITLE,
    description: SITE_DESCRIPTION,
    siteName: SITE_NAME,
    images: [OG_IMAGE],
    type: "website",
  },
  twitter: { card: "summary_large_image" },
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html
      lang="en"
      className={`${newsreader.variable} ${dmSans.variable} ${plexMono.variable}`}
      suppressHydrationWarning
    >
      <head>
        {/* Lets CSS hide scroll-reveal content only when JS can reveal it. */}
        <script
          dangerouslySetInnerHTML={{
            __html: "document.documentElement.classList.add('js')",
          }}
        />
      </head>
      <body>
        {children}
        <DevAnnotations />
      </body>
    </html>
  );
}
