import { t, type Dictionary } from "intlayer";

export interface Content {
  user: {
    tooltips: {
      copy: string;
      copied: string;
      edit: string;
    };
    editor: {
      send: string;
      cancel: string;
      warning: string;
    };
  };
}

export default {
  key: "chat-message",
  content: {
    user: {
      tooltips: {
        copy: t({ en: "Copy", ar: "إنسخ" }),
        copied: t({ en: "Copied!", ar: "تم النسخ!" }),
        edit: t({ en: "Edit", ar: "عدّل" }),
      },
      editor: {
        send: t({ en: "Send", ar: "أرسل" }),
        cancel: t({ en: "Cancel", ar: "ألغِ" }),
        warning: t({
          en: "Editing this message will overwrite all subsequent messages.",
          ar: "تعديل هذه الرسالة سيحذف الرسائل التالية.",
        }),
      },
    },
  },
} satisfies Dictionary<Content>;
