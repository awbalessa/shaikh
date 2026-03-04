"use client";

import { useEffect, useRef } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import { isTextUIPart, type UIMessage } from "ai";
import { useChat } from "@ai-sdk/react";
import { cn } from "@/lib/utils";

type ChatMessagesProps = React.ComponentPropsWithoutRef<"div"> & {
  messages: UIMessage[];
  status: ReturnType<typeof useChat>["status"];
};

export default function ChatMessages({
  messages,
  status,
  className,
  ...props
}: ChatMessagesProps) {
  const bottomRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  return (
    <div {...props} className={className}>
      <div className="flex flex-col gap-4 px-4 py-4">
        {messages.map((message) => {
          const text = message.parts.filter(isTextUIPart).map((p) => p.text).join("");

          return (
            <div
              key={message.id}
              className={cn(
                message.role === "user" ? "flex justify-end" : "flex justify-start",
              )}
            >
              <div
                className={cn(
                  "max-w-[80%] rounded-lg px-4 py-2",
                  message.role === "user"
                    ? "bg-primary text-text-on-primary"
                    : "bg-surface-light text-text",
                )}
              >
                {message.role === "assistant" ? (
                  <ReactMarkdown remarkPlugins={[remarkGfm]}>
                    {text}
                  </ReactMarkdown>
                ) : (
                  <span>{text}</span>
                )}
              </div>
            </div>
          );
        })}
        {status === "submitted" && (
          <div className="flex justify-start">
            <div className="bg-surface-light text-text-muted rounded-lg px-4 py-2 text-sm">
              ...
            </div>
          </div>
        )}
      </div>
      <div ref={bottomRef} />
    </div>
  );
}
