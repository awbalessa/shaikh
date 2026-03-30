import "./globals.css";
import type { Metadata } from "next";
import { ThemeToggle } from "@/components/theme/theme-toggle";
import "streamdown/styles.css";

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
    <html lang="ar" dir="rtl" suppressHydrationWarning>
      <body>
        <ThemeToggle />
        {children}
      </body>
    </html>
  );
}
