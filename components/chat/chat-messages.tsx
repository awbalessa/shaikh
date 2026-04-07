"use client";

import { useEffect, useRef } from "react";
import { ChatStatus, isTextUIPart, type UIMessage } from "ai";
import { Streamdown } from "streamdown";
import { cn } from "@/lib/utils";

const md: React.ComponentProps<typeof Streamdown>["components"] = {
  h1: ({ children, ...props }) => (
    <h1 className="text-3xl leading-12 font-semibold pt-10 m-0" {...props}>
      {children}
    </h1>
  ),
  h2: ({ children, ...props }) => (
    <h2 className="text-2xl leading-9 font-semibold pt-8 m-0" {...props}>
      {children}
    </h2>
  ),
  h3: ({ children, ...props }) => (
    <h3 className="text-xl leading-7 font-semibold pt-6 m-0" {...props}>
      {children}
    </h3>
  ),
  h4: ({ children, ...props }) => (
    <h4 className="text-lg leading-7 font-semibold pt-6 m-0" {...props}>
      {children}
    </h4>
  ),
  p: ({ children, ...props }) => (
    <p className="text-base pt-5 first:pt-0 m-0 leading-6.5" {...props}>
      {children}
    </p>
  ),
  ul: ({ children }) => (
    <ul className="flex flex-col gap-4 pt-5 ps-5 list-disc m-0">{children}</ul>
  ),

  ol: ({ children }) => (
    <ol className="flex flex-col gap-4 pt-5 ps-5 list-decimal m-0">
      {children}
    </ol>
  ),

  li: ({ children }) => <li className="leading-6 text-base m-0">{children}</li>,

  hr: () => <hr className="border-t border-divider m-0 mt-6" />,
};

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
  const scrollContainerRef = useRef<HTMLDivElement | null>(null);
  const lastUserMessageRef = useRef<HTMLDivElement | null>(null);

  const isGenerating = status === "submitted" || status === "streaming";

  const lastUserIndex = messages.reduce(
    (acc, m, i) => (m.role === "user" ? i : acc),
    -1,
  );

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
              className={message.role === "user" ? "pt-8 first:pt-0" : ""}
            >
              {message.role === "user" ? (
                <div className="group">
                  <div
                    dir="auto"
                    className="w-fit max-w-[80%] bg-surface-light dark:bg-surface-medium text-text rounded-lg px-3 py-2 text-base leading-6"
                  >
                    {text}
                  </div>
                </div>
              ) : (
                <div dir="auto">
                  <Streamdown components={md}>{text}</Streamdown>
                </div>
              )}
            </div>
          );
        })}

        {waitingForFirstToken && (
          <div className="pt-2">
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
