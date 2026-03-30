import "./globals.css";
import type { Metadata } from "next";
import { ThemeToggle } from "@/components/theme/theme-toggle";
import "streamdown/styles.css";
import { Geist } from "next/font/google";
import { cn } from "@/lib/utils";

const geist = Geist({subsets:['latin'],variable:'--font-sans'});

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
    <html lang="ar" dir="rtl" suppressHydrationWarning className={cn("font-sans", geist.variable)}>
      <body>
        <ThemeToggle />
        {children}
      </body>
    </html>
  );
}
