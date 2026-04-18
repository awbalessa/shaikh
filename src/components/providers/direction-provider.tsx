"use client";

import { Direction } from "radix-ui";

export function DirectionProvider({
  dir,
  children,
}: {
  dir: "rtl" | "ltr";
  children: React.ReactNode;
}) {
  return <Direction.Provider dir={dir}>{children}</Direction.Provider>;
}
