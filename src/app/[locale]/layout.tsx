import { getHTMLTextDir } from "intlayer";
import { IntlayerClientProvider, NextLayoutIntlayer } from "next-intlayer";
export { generateStaticParams } from "next-intlayer";
import type { Metadata } from "next";
import { Tooltip } from "radix-ui";
import { ThemeToggle } from "@/components/theme-toggle";
import { ThemeProvider } from "next-themes";
import DirectionProvider from "@/components/direction-provider";
import "../globals.css";
import "streamdown/styles.css";

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
            <ThemeProvider attribute="class" defaultTheme="system" enableSystem>
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
