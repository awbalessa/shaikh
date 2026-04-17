import { getHTMLTextDir } from "intlayer";
import { IntlayerClientProvider, NextLayoutIntlayer } from "next-intlayer";
export { generateStaticParams } from "next-intlayer";
import { AppProviders } from "@/components/providers";
import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Shaikh",
  description: "Ask Shaikh",
};

const LocaleLayout: NextLayoutIntlayer = async ({ children, params }) => {
  const { locale } = await params;
  const dir = getHTMLTextDir(locale);

  return (
    <html lang={locale} dir={dir} suppressHydrationWarning>
      <body>
        <IntlayerClientProvider locale={locale}>
          <AppProviders dir={dir as "ltr" | "rtl"}>{children}</AppProviders>
        </IntlayerClientProvider>
      </body>
    </html>
  );
};

export default LocaleLayout;
