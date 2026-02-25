import ChatClient from "./chat-client";
import { cn } from "@/lib/utils";

type ChatPaneProps = React.ComponentPropsWithoutRef<"section">;

export default function ChatPane({ className, ...props }: ChatPaneProps) {
  return (
    <section
      {...props}
      className={cn("flex flex-col h-full min-h-0", className)}
    >
      <div className="flex flex-col mx-auto w-full h-full px-8 pb-8 pt-6 min-w-[390px] max-w-[850px]">
        <ChatClient />
      </div>
    </section>
  );
}
