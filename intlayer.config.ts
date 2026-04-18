import { Locales, type IntlayerConfig } from "intlayer";

const config: IntlayerConfig = {
  internationalization: {
    locales: [Locales.ARABIC, Locales.ENGLISH],
    defaultLocale: Locales.ARABIC,
  },
  content: {
    contentDir: ["src"],
  },
  routing: {
    mode: "prefix-no-default",
    storage: [{ type: "cookie" }],
  },
};

export default config;
