import type { Metadata } from "next";
import { DM_Sans, IBM_Plex_Mono, Newsreader } from "next/font/google";
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

const title = "newsjack.sh";
const description =
  "The open-source skills that turn your agent into a PR operator.";
const ogImage = {
  url: "/newsjack-og-image.png",
  width: 1497,
  height: 789,
  alt: "newsjack.sh - The open-source skills that turn your agent into a PR operator.",
};

export const metadata: Metadata = {
  metadataBase: new URL("https://newsjack.sh"),
  title,
  description,
  openGraph: {
    title,
    description,
    url: "https://newsjack.sh",
    siteName: "newsjack.sh",
    images: [ogImage],
    type: "website",
  },
  twitter: {
    card: "summary_large_image",
    title,
    description,
    images: [ogImage],
  },
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
