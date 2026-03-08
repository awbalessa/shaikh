"use client";

import { useEffect, useRef, useState } from "react";
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
  const scrollContainerRef = useRef<HTMLDivElement | null>(null);
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
      const container = scrollContainerRef.current;
      const lastEl = lastMessageRef.current;
      if (container && lastEl) {
        const scrollToMessageTop = () => {
          const msgTop = lastEl.getBoundingClientRect().top;
          const containerTop = container.getBoundingClientRect().top;
          container.scrollTop += msgTop - containerTop;
        };
        requestAnimationFrame(scrollToMessageTop);
      }
    } else {
      bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [messages]);

  return (
    <div ref={scrollContainerRef} {...props} className={className}>
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
                  isLastMessage ? (
                    <StreamingWordReveal text={text} isStreaming={status === "streaming"} />
                  ) : (
                    <ReactMarkdown remarkPlugins={[remarkGfm, remarkBreaks]}>
                      {text}
                    </ReactMarkdown>
                  )
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

function StreamingWordReveal({ text, isStreaming }: { text: string; isStreaming: boolean }) {
  const words = text.split(/\s+/).filter(Boolean);
  const targetRef = useRef(words.length);
  const [displayedCount, setDisplayedCount] = useState(0);

  useEffect(() => {
    targetRef.current = words.length;
  }, [words.length]);

  useEffect(() => {
    const t = setInterval(() => {
      setDisplayedCount((c) => Math.min(c + 1, targetRef.current));
    }, 28);
    return () => clearInterval(t);
  }, []);

  const done = displayedCount >= words.length && !isStreaming;

  if (done && text.length > 0) {
    return (
      <ReactMarkdown remarkPlugins={[remarkGfm, remarkBreaks]}>
        {text}
      </ReactMarkdown>
    );
  }

  if (words.length === 0) {
    return null;
  }

  const visible = words.slice(0, displayedCount);

  return (
    <span className="inline">
      {visible.map((word, i) => (
        <span
          key={i}
          className={i === visible.length - 1 ? "streaming-word-fade-in" : undefined}
        >
          {i > 0 ? " " : ""}
          {word}
        </span>
      ))}
    </span>
  );
}