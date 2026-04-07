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
