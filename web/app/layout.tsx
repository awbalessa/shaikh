import "./globals.css";
import type { Metadata } from "next";
import { Rubik } from "next/font/google";
import { ThemeToggle } from "@/components/theme/theme-toggle";

const rubik = Rubik({
  subsets: ["arabic", "latin"],
  variable: "--font-sans",
  display: "swap",
});

export const metadata: Metadata = {
  title: "Shaikh",
  description: "Ask Shaikh",
};

export default function RootLayout({
  children,
}: {
  children: React.ReactNode;
}) {
  return (
    <html
      lang="ar"
      dir="rtl"
      className={rubik.variable}
      suppressHydrationWarning
    >
      <body>
        <ThemeToggle />
        {children}
      </body>
    </html>
  );
}
