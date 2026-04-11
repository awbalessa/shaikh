"use client";

import { cn } from "@/lib/utils";
import { ChatStatus, isTextUIPart, UIMessage } from "ai";
import { Streamdown } from "streamdown";
import { useEffect, useLayoutEffect, useRef, useState } from "react";
import { IconArrowNarrowDown } from "@tabler/icons-react";

const MIN_SPACER = 64;

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
  const lastUserMessageId = messages.findLast((m) => m.role === "user")?.id;

  const containerRef = useRef<HTMLDivElement>(null);
  const lastUserRef = useRef<HTMLDivElement>(null);
  const spacerRef = useRef<HTMLDivElement>(null);
  const [showButton, setShowButton] = useState(false);

  useEffect(() => {
    const el = containerRef.current;
    if (!el) return;
    const onScroll = () => {
      setShowButton(
        el.scrollHeight - el.scrollTop - el.clientHeight > MIN_SPACER * 2,
      );
    };
    el.addEventListener("scroll", onScroll, { passive: true });
    return () => el.removeEventListener("scroll", onScroll);
  }, []);

  useLayoutEffect(() => {
    const last = messages[messages.length - 1];
    if (last?.role !== "user") return;
    const container = containerRef.current;
    const el = lastUserRef.current;
    const spacer = spacerRef.current;
    if (!container || !el || !spacer) return;
    spacer.style.minHeight = `${Math.max(
      MIN_SPACER,
      container.clientHeight - el.offsetHeight,
    )}px`;
    requestAnimationFrame(() => {
      container.scrollTo({
        top: el.offsetTop - container.offsetTop,
        behavior: "smooth",
      });
    });
  }, [messages.length]); // eslint-disable-line

  useLayoutEffect(() => {
    if (status !== "ready" && status !== "error") return;
    if (spacerRef.current)
      spacerRef.current.style.minHeight = `${MIN_SPACER}px`;
  }, [status]);

  return (
    <div className="relative flex-1 min-h-0">
      <div
        ref={containerRef}
        className={cn(
          "h-full overflow-y-auto messages-scroll flex flex-col",
          className,
        )}
        {...props}
      >
        {messages.map((message) => (
          <div
            key={message.id}
            ref={message.id === lastUserMessageId ? lastUserRef : undefined}
            className="py-2"
          >
            {message.role === "user" ? (
              <div className="flex justify-start">
                <div
                  dir="auto"
                  className="bg-muted rounded-2xl px-3 py-1 max-w-[80%]"
                >
                  {message.parts.filter(isTextUIPart).map((part, i) => (
                    <span key={i}>{part.text}</span>
                  ))}
                </div>
              </div>
            ) : (
              <div dir="auto">
                {message.parts.filter(isTextUIPart).map((part, i) => (
                  <Streamdown key={i}>{part.text}</Streamdown>
                ))}
              </div>
            )}
          </div>
        ))}

        {isWaiting && <div className="py-2 opacity-40 text-sm">...</div>}

        <div ref={spacerRef} className="flex-1 shrink-0" />
      </div>

      {showButton && status !== "submitted" && (
        <button
          onClick={() =>
            containerRef.current?.scrollTo({
              top: containerRef.current.scrollHeight,
              behavior: "smooth",
            })
          }
          className="absolute bottom-2 left-1/2 -translate-x-1/2 bg-background border
  border-border rounded-full p-1 shadow-md"
        >
          <IconArrowNarrowDown className="size-4" />
        </button>
      )}
    </div>
  );
}
