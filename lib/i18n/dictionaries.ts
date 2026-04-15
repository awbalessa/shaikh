import { useLocale } from "@/components/providers/locale-provider";

export const dictionaries = {
  en: {
    chat: {
      thread: {
        userMessage: {
          actions: {
            copy: "Copy",
            copied: "Copied!",
            edit: "Edit",
          },
        },
      },
      composer: {
        placeholder: "Ask Shaikh...",
      },
    },
  },
  ar: {
    chat: {
      thread: {
        userMessage: {
          actions: {
            copy: "إنسخ",
            copied: "تم النسخ!",
            edit: "عدّل",
          },
        },
      },
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
