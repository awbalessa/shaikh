import { Locales, type IntlayerConfig } from "intlayer";

const config: IntlayerConfig = {
  internationalization: {
    locales: [Locales.ARABIC, Locales.ENGLISH],
    defaultLocale: Locales.ARABIC,
  },
  content: {
    contentDir: ["src"],
  },
  log: {
    mode: "verbose",
  },
};

export default config;
