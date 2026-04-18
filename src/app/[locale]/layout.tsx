import { getHTMLTextDir } from "intlayer";
import { IntlayerClientProvider, NextLayoutIntlayer } from "next-intlayer";
export { generateStaticParams } from "next-intlayer";
import type { Metadata } from "next";
import { DirectionProvider } from "@/components/providers/direction-provider";
import { ThemeProvider } from "@/components/providers/theme-provider";
import { Tooltip } from "radix-ui";
import { ThemeToggle } from "@/components/theme-toggle";

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
          <DirectionProvider dir={dir as "rtl" | "ltr"}>
            <ThemeProvider>
              <ThemeToggle />
              <Tooltip.Provider delayDuration={100}>
                {children}
              </Tooltip.Provider>
            </ThemeProvider>
          </DirectionProvider>
        </IntlayerClientProvider>
      </body>
    </html>
  );
};

export default LocaleLayout;
