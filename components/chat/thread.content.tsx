import { t, type Dictionary } from "intlayer";

export default {
  key: "chat-thread",
  content: {
    messageActions: {
      copy: t({ en: "Copy", ar: "إنسخ" }),
      copied: t({ en: "Copied!", ar: "تم النسخ!" }),
      edit: t({ en: "Edit", ar: "عدّل" }),
    },
  },
} satisfies Dictionary;
