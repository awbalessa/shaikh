"use client";

import { Direction } from "radix-ui";
import { PropsWithChildren } from "react";

export default function DirectionProvider({
  dir,
  children,
}: PropsWithChildren<{ dir: "rtl" | "ltr" }>) {
  return <Direction.Provider dir={dir}>{children}</Direction.Provider>;
}
