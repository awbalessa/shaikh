"use client";

import { useEffect, useRef } from "react";
import { isTextUIPart, type UIMessage } from "ai";
import { useChat } from "@ai-sdk/react";
import { Streamdown } from "streamdown";
import {
  IconCopy,
  IconThumbUp,
  IconThumbDown,
  IconRepeat,
} from "@tabler/icons-react";
import { getIconStroke } from "@/lib/utils";

type ChatMessagesProps = React.ComponentPropsWithoutRef<"div"> & {
  messages: UIMessage[];
  status: ReturnType<typeof useChat>["status"];
};

const md: React.ComponentProps<typeof Streamdown>["components"] = {
  h1: ({ children, ...props }) => (
    <h1 className="text-4xl leading-12 font-semibold pt-10" {...props}>
      {children}
    </h1>
  ),
  h2: ({ children, ...props }) => (
    <h2 className="text-2xl leading-9 font-semibold pt-8" {...props}>
      {children}
    </h2>
  ),
  h3: ({ children, ...props }) => (
    <h3 className="text-lg leading-7 font-semibold pt-6" {...props}>
      {children}
    </h3>
  ),
  p: ({ children, ...props }) => (
    <p className="text-base leading-6 pt-4 first:pt-0" {...props}>
      {children}
    </p>
  ),
  ul: ({ children }) => (
    <ul className="pt-4 gap-3 flex flex-col list-none">{children}</ul>
  ),
  ol: ({ children }) => (
    <ol className="pt-4 gap-3 flex flex-col list-none">{children}</ol>
  ),
  li: ({ children }) => (
    <li className="flex items-baseline gap-1 text-base leading-6">
      <span aria-hidden>·</span>
      <span>{children}</span>
    </li>
  ),
  hr: () => <hr className="border-t border-divider mt-6" />,
};

function MessageActions() {
  return (
    <div className="flex gap-1 pt-3">
      {(
        [
          [IconCopy, "Copy"],
          [IconThumbUp, "Share feedback"],
          [IconThumbDown, "Share feedback"],
          [IconRepeat, "Regenerate"],
        ] as const
      ).map(([Icon, label]) => (
        <button
          key={label}
          type="button"
          aria-label={label}
          className="p-1 rounded-full text-text-muted hover:bg-surface-light transition-colors cursor-pointer"
        >
          <Icon size={16} stroke={getIconStroke(16)} />
        </button>
      ))}
    </div>
  );
}

export default function ChatMessages({
  messages,
  status,
  className,
  ...props
}: ChatMessagesProps) {
  const scrollContainerRef = useRef<HTMLDivElement | null>(null);
  const lastUserMessageRef = useRef<HTMLDivElement | null>(null);

  const isGenerating = status === "submitted" || status === "streaming";

  const lastUserIndex = messages.reduce(
    (acc, m, i) => (m.role === "user" ? i : acc),
    -1,
  );

  useEffect(() => {
    const last = messages[messages.length - 1];
    if (last?.role !== "user") return;

    requestAnimationFrame(() => {
      lastUserMessageRef.current?.scrollIntoView({
        behavior: "smooth",
        block: "start",
      });
    });
  }, [messages.length]);

  const waitingForFirstToken =
    status === "submitted" ||
    (status === "streaming" &&
      messages[messages.length - 1]?.role === "assistant" &&
      messages[messages.length - 1]?.parts
        .filter(isTextUIPart)
        .map((p) => p.text)
        .join("") === "");

  return (
    <div ref={scrollContainerRef} {...props} className={className}>
      <div className="flex flex-col px-6">
        {messages.map((message, index) => {
          const text = message.parts
            .filter(isTextUIPart)
            .map((p) => p.text)
            .join("");

          const isLastUser = message.role === "user" && index === lastUserIndex;

          const isLast = index === messages.length - 1;

          return (
            <div
              key={message.id}
              ref={isLastUser ? lastUserMessageRef : undefined}
              className="pt-4"
            >
              {message.role === "user" ? (
                <div
                  dir="auto"
                  className="w-fit max-w-[80%] bg-surface-medium text-text rounded-lg px-3 py-2 text-base leading-6"
                >
                  {text}
                </div>
              ) : (
                <div dir="auto">
                  <Streamdown components={md}>{text}</Streamdown>
                  {(!isLast || (isLast && !isGenerating)) && <MessageActions />}
                </div>
              )}
            </div>
          );
        })}

        {waitingForFirstToken && (
          <div className="pt-4">
            <span
              className="block h-2 w-2 rounded-full bg-bg-inverse animate-pulse"
              aria-hidden
            />
          </div>
        )}

        <div aria-hidden className="h-4" />
      </div>
    </div>
  );
}
