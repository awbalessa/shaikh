import ChatClient from "@/components/chat/chat-client";
import ChatPane from "@/components/chat/chat-pane";

export default function Page() {
  return (
    <main className="h-dvh flex flex-row bg-bg text-text min-h-0 overflow-hidden">
      {/* RTL: first = right, second = left → Quran right (7/12), AI left (5/12) */}
      <section
        className="flex-1 min-w-[560px] flex flex-col"
        aria-label="Quran"
      >
        <div className="w-full max-w-[720px] mx-auto" />
      </section>
      <section
        className="shrink-0 basis-5/12 min-w-[390px] flex flex-col border-s border-border h-full min-h-0 py-6"
        aria-label="Chat"
      >
        <ChatClient />
      </section>
    </main>
  );
}
