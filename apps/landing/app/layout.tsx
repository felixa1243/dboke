import type { Metadata } from 'next';
import "./globals.css";

export const metadata: Metadata = {
  title: "Dboke - Visual Database Explorer",
  description: "Dboke is a secure, universal visual database explorer and management tool. Experience a minimalist aesthetic with native support for PostgreSQL, MySQL, and MongoDB.",
  keywords: ["Visual Database Explorer", "Database Manager", "PostgreSQL", "MySQL", "MongoDB", "Database GUI", "Go", "Next.js"],
  openGraph: {
    title: "Dboke - Visual Database Explorer",
    description: "A secure, universal visual database explorer built with Go and Next.js.",
    type: "website",
    url: "https://felixa1243.github.io/dboke",
  },
  twitter: {
    card: "summary_large_image",
    title: "Dboke - Visual Database Explorer",
    description: "A secure, universal visual database explorer built with Go and Next.js.",
  },
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html lang="en">
      <body className="antialiased bg-white text-black dark:bg-black dark:text-white font-sans selection:bg-black selection:text-white dark:selection:bg-white dark:selection:text-black">
        {children}
      </body>
    </html>
  );
}
