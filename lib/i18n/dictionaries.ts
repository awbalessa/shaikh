import { useLocale } from "@/components/providers/locale-provider";

export const dictionaries = {
  en: {
    chat: {
      composer: {
        placeholder: "Ask Shaikh...",
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

export function useDictionary() {
  const { locale } = useLocale();
  return dictionaries[locale];
}
