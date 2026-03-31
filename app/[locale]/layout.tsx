import "../globals.css";
import type { Metadata } from "next";
import { DirectionProvider } from "@/components/ui/direction";
import { TooltipProvider } from "@/components/ui/tooltip";
import { ThemeToggle } from "@/components/theme/theme-toggle";
import "streamdown/styles.css";
import { isLocale, locales } from "@/lib/i18n/locale";
import { LocaleProvider } from "@/lib/i18n/locale-context";
import { notFound } from "next/navigation";

export function generateStaticParams() {
  return locales.map((locale) => ({ locale }));
}

export const metadata: Metadata = {
  title: "Shaikh",
  description: "Ask Shaikh",
};

export default async function RootLayout({
  children,
  params,
}: LayoutProps<"/[locale]">) {
  const { locale } = await params;
  if (!isLocale(locale)) notFound();

  const isRTL = locale === "ar";
  const dir = isRTL ? "rtl" : "ltr";

  return (
    <html lang={locale} dir={dir} suppressHydrationWarning>
      <body>
        <LocaleProvider locale={locale}>
          <DirectionProvider dir={dir}>
            <TooltipProvider>
              <ThemeToggle />
              {children}
            </TooltipProvider>
          </DirectionProvider>
        </LocaleProvider>
      </body>
    </html>
  );
}
