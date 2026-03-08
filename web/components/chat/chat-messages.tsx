"use client";

import { useEffect, useRef } from "react";
import ReactMarkdown from "react-markdown";
import remarkGfm from "remark-gfm";
import remarkBreaks from "remark-breaks";
import { isTextUIPart, type UIMessage } from "ai";
import { useChat } from "@ai-sdk/react";
import { BaseDir } from "./chat-client";

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
  const lastMessageRef = useRef<HTMLDivElement | null>(null);

  const lastMessage = messages[messages.length - 1];
  const lastIsAssistantEmpty =
    lastMessage?.role === "assistant" &&
    lastMessage.parts
      .filter(isTextUIPart)
      .map((p) => p.text)
      .join("") === "";
  const waitingForFirstToken =
    status === "submitted" || (status === "streaming" && lastIsAssistantEmpty);

  useEffect(() => {
    const last = messages[messages.length - 1];
    if (last?.role === "user") {
      lastMessageRef.current?.scrollIntoView({
        behavior: "smooth",
        block: "start",
      });
    } else {
      bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages]);

  return (
    <div {...props} className={className}>
      <div className="flex flex-col gap-4 px-6 max-w-[850px] mx-auto">
        {messages.map((message, index) => {
          const text = message.parts
            .filter(isTextUIPart)
            .map((p) => p.text)
            .join("");

          const isLastMessage = index === messages.length - 1;

          return (
            <div
              key={message.id}
              ref={isLastMessage ? lastMessageRef : undefined}
              dir={message.role === "user" ? BaseDir : "auto"}
              className={message.role === "user" ? "max-w-[80%]" : ""}
            >
              <div
                className={
                  message.role === "user"
                    ? "w-fit justify-start bg-surface-light text-text rounded-lg px-4 py-2"
                    : ""
                }
              >
                {message.role === "assistant" ? (
                  <ReactMarkdown remarkPlugins={[remarkGfm, remarkBreaks]}>
                    {text}
                  </ReactMarkdown>
                ) : (
                  <span>{text}</span>
                )}
              </div>
            </div>
          );
        })}
        {waitingForFirstToken && (
          <div className="flex justify-start">
            <span
              className="h-2 w-2 rounded-full bg-bg-inverse animate-pulse"
              aria-hidden
            />
          </div>
        )}
      </div>
      <div ref={bottomRef} />
    </div>
  );
}
