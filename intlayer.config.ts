import { Locales, type IntlayerConfig } from "intlayer";

const config: IntlayerConfig = {
  internationalization: {
    locales: [Locales.ENGLISH, Locales.ARABIC],
    defaultLocale: Locales.ARABIC,
  },
  content: {
    contentDir: ["src", "components", "app"],
  },
  routing: {
    mode: "prefix-no-default",
    storage: [{ type: "cookie", name: "INTLAYER_LOCALE" }],
  },
};

export default config;
