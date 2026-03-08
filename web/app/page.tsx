import ChatPane from "@/components/chat/chat-pane";

export default function Page() {
  return (
    <main className="h-dvh flex flex-row bg-bg text-text min-h-0 overflow-hidden">
      {/* RTL: first = right, second = left → Quran right (7/12), AI left (5/12) */}
      <section className="basis-7/12 min-w-0 flex flex-col bg-bg" aria-label="Quran">
        <div className="flex-1 min-h-0 px-8 pb-8 pt-6" />
      </section>
      <ChatPane className="basis-5/12 min-w-[390px] flex flex-col border-s border-border" aria-label="Chat" />
    </main>
  );
}
