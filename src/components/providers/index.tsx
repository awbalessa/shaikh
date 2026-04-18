"use client";

import { ThemeProvider } from "./theme-provider";
import { DirectionProvider } from "./direction-provider";
import { Tooltip } from "radix-ui";

export function AppProviders({
  dir,
  children,
}: {
  dir: "ltr" | "rtl";
  children: React.ReactNode;
}) {
  return (
    <DirectionProvider dir={dir}>
      <ThemeProvider>
        <Tooltip.Provider delayDuration={100}>{children}</Tooltip.Provider>
      </ThemeProvider>
    </DirectionProvider>
  );
}
