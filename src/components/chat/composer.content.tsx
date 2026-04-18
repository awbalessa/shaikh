import { t, type Dictionary } from "intlayer";

interface Content {
  placeholder: string;
}

export default {
  key: "chat-composer",
  content: {
    placeholder: t({ en: "Ask Shaikh...", ar: "اسأل شيخ..." }),
  },
} satisfies Dictionary<Content>;
