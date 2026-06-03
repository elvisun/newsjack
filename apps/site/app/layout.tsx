import type { Metadata } from "next";
import "./globals.css";

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
    <html lang="en">
      <body>{children}</body>
    </html>
  );
}
