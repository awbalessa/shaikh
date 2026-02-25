import ChatClient from "./chat-client";

type ChatPaneProps = React.ComponentPropsWithoutRef<"section">;

export default function ChatPane({ className, ...props }: ChatPaneProps) {
  return (
    <section
      {...props}
      className={["flex flex-col h-full min-h-0", className ?? ""].join(" ")}
    >
      <div className="flex flex-col mx-auto w-full h-full px-8 pb-8 pt-6">
        <ChatClient />
      </div>
    </section>
  );
}
