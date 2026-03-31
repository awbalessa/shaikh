"use client";

import { createContext, useContext } from "react";
import { defaultLocale, Locale } from "./locale";

const LocaleContext = createContext<{ locale: Locale }>({
  locale: defaultLocale,
});

export function LocaleProvider({
  locale,
  children,
}: {
  locale: Locale;
  children: React.ReactNode;
}) {
  return (
    <LocaleContext.Provider value={{ locale }}>
      {children}
    </LocaleContext.Provider>
  );
}

export function useLocale() {
  return useContext(LocaleContext);
}
