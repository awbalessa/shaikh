import ChatClient from "@/components/chat";
import { NextPageIntlayer } from "next-intlayer";
import { IntlayerServerProvider } from "next-intlayer/server";

const Page: NextPageIntlayer = async ({ params }) => {
  const { locale } = await params;

  return (
    <IntlayerServerProvider locale={locale}>
      <PageContent />
    </IntlayerServerProvider>
  );
};

export default Page;

function PageContent() {
  return (
    <main className="h-dvh flex flex-row overflow-hidden bg-surface-composer">
      <QuranPane />
      <ChatPane />
    </main>
  );
}

function QuranPane() {
  return (
    <section className="flex-1 min-w-140 flex flex-col" aria-label="Quran">
      <div className="w-full max-w-180 mx-auto" />
    </section>
  );
}

function ChatPane() {
  return (
    <section
      className="shrink-0 basis-5/12 min-w-97.5 border-s border-border-light size-full"
      aria-label="Chat"
    >
      <ChatClient />
    </section>
  );
}
