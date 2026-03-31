import "server-only";
import { defaultLocale, Locale } from "./locale";

const dictionaries = {
  en: {
    chat: {
      composer: {
        placeholder: "Ask Sheikh...",
      },
    },
  },
  ar: {
    chat: {
      composer: {
        placeholder: "اسأل شيخ...",
      },
    },
  },
} as const;

export type Dictionary = (typeof dictionaries)[keyof typeof dictionaries];

export async function getDictionary(locale: Locale) {
  return (
    dictionaries[locale as keyof typeof dictionaries] ??
    dictionaries[defaultLocale]
  );
}
