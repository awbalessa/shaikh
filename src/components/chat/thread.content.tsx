import { t, type Dictionary } from "intlayer";

interface Content {
  messageActions: {
    copy: string;
    copied: string;
    edit: string;
  };
}

export default {
  key: "chat-thread",
  content: {
    messageActions: {
      copy: t({ en: "Copy", ar: "إنسخ" }),
      copied: t({ en: "Copied!", ar: "تم النسخ!" }),
      edit: t({ en: "Edit", ar: "عدّل" }),
    },
  },
} satisfies Dictionary<Content>;
