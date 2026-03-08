import ChatClient from "./chat-client";
import { cn } from "@/lib/utils";

type ChatPaneProps = React.ComponentPropsWithoutRef<"section">;

export default function ChatPane({ className, ...props }: ChatPaneProps) {
  return (
    <section
      {...props}
      className={cn("flex flex-col h-full min-h-0 w-full", className)}
    >
      <div className="flex flex-col flex-1 min-h-0 w-full py-6">
        <ChatClient />
      </div>
    </section>
  );
}
