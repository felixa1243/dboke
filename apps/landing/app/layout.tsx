import "./globals.css";

export const metadata = {
  title: "Dboke - Minimalist Universal Database Manager",
  description: "A secure, universal database management tool with a minimalist, elegant aesthetic.",
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
