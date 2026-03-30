import ChatClient from "@/components/chat/chat-client";

export default function Page() {
  return (
    <main className="h-dvh flex flex-row bg-bg text-text overflow-hidden">
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
      className="shrink-0 basis-5/12 min-w-97.5 border-s border-border h-full py-6"
      aria-label="Chat"
    >
      <ChatClient />
    </section>
  );
}
