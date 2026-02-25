import ChatPane from "@/components/chat/chat-pane";

export default function Page() {
  return (
    <main className="h-dvh flex bg-bg text-text min-h-0">
      <section className="basis-7/12 min-w-[590px]">
        <div className="mx-auto h-full px-8 pb-8 pt-6" />
      </section>

      <ChatPane className="basis-5/12 border-s-[0.5px] border-border" />
    </main>
  );
}
