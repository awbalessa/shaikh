"use client";

import { cn } from "@/lib/utils";
import { ChatStatus, isTextUIPart, UIMessage } from "ai";
import { Streamdown } from "streamdown";

type ChatMessagesProps = React.ComponentPropsWithoutRef<"div"> & {
  messages: UIMessage[];
  status: ChatStatus;
};

export default function ChatMessages({
  messages,
  status,
  className,
  ...props
}: ChatMessagesProps) {
  const isWaiting = status === "submitted" && messages.at(-1)?.role === "user";

  return (
    <div className={cn("flex-1 overflow-y-auto", className)} {...props}>
      {messages.map((message) => (
        <div
          key={message.id}
          className={cn(
            "py-4 px-2",
            message.role === "user" ? "text-right" : "text-left",
          )}
        >
          {message.parts
            .filter(isTextUIPart)
            .map((part, i) =>
              message.role === "assistant" ? (
                <Streamdown key={i}>{part.text}</Streamdown>
              ) : (
                <span key={i}>{part.text}</span>
              ),
            )}
        </div>
      ))}

      {isWaiting && <div className="py-4 px-2 opacity-40 text-sm">...</div>}
    </div>
  );
}
