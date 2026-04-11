import "../globals.css";
import "streamdown/styles.css";
import type { Metadata } from "next";
import { AppProviders } from "@/components/providers";
import { ThemeToggle } from "@/components/theme-toggle";
import { isLocale, locales } from "@/lib/i18n/locale";
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
        <AppProviders locale={locale} dir={dir}>
          <ThemeToggle />
          {children}
        </AppProviders>
      </body>
    </html>
  );
}
