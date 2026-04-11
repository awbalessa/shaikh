"use client";

import { ThemeProvider } from "./theme-provider";
import { DirectionProvider } from "./direction-provider";
import { LocaleProvider } from "./locale-provider";
import type { Locale } from "@/lib/i18n/locale";

export function AppProviders({
  locale,
  dir,
  children,
}: {
  locale: Locale;
  dir: "ltr" | "rtl";
  children: React.ReactNode;
}) {
  return (
    <LocaleProvider locale={locale}>
      <DirectionProvider dir={dir}>
        <ThemeProvider>{children}</ThemeProvider>
      </DirectionProvider>
    </LocaleProvider>
  );
}
