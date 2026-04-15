"use client";

import { ThemeProvider } from "./theme-provider";
import { DirectionProvider } from "./direction-provider";
import { LocaleProvider } from "./locale-provider";
import type { Locale } from "@/lib/i18n/locale";
import { Tooltip } from "radix-ui";

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
        <ThemeProvider>
          <Tooltip.Provider delayDuration={100}>{children}</Tooltip.Provider>
        </ThemeProvider>
      </DirectionProvider>
    </LocaleProvider>
  );
}
