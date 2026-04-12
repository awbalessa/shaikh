"use client";

import { cn } from "@/lib/utils";
import { ChatStatus, isTextUIPart, UIMessage } from "ai";
import { Streamdown } from "streamdown";
import {
  memo,
  useCallback,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import { IconArrowNarrowDown } from "@tabler/icons-react";
import { AnimatePresence, HTMLMotionProps, motion } from "motion/react";

const MIN_SPACER = 64;

type ChatThreadProps = React.ComponentPropsWithoutRef<"div"> & {
  messages: UIMessage[];
  status: ChatStatus;
};

export default function ChatThread({
  messages,
  status,
  className,
  ...props
}: ChatThreadProps) {
  const isWaiting = status === "submitted" && messages.at(-1)?.role === "user";
  const lastUserMessageId = messages.findLast((m) => m.role === "user")?.id;

  const containerRef = useRef<HTMLDivElement>(null);
  const lastUserRef = useRef<HTMLDivElement>(null);
  const spacerRef = useRef<HTMLDivElement>(null);
  const [showButton, setShowButton] = useState(false);

  const onCopy = useCallback((text: string) => {
    navigator.clipboard.writeText(text);
  }, []);
  const onEdit = useCallback((_message: UIMessage) => {}, []);
  const onRetry = useCallback((_messageID: string) => {}, []);

  const onScrollButtonClick = useCallback(() => {
    containerRef.current?.scrollTo({
      top: containerRef.current.scrollHeight,
      behavior: "smooth",
    });
  }, []);

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

    const container = containerRef.current;
    const spacer = spacerRef.current;
    const lastUser = lastUserRef.current;
    if (!container || !spacer || !lastUser) return;

    // Where are we right now?
    const prevScrollTop = container.scrollTop;

    // Reset spacer so we measure real content height
    spacer.style.minHeight = "0px";

    // Height from last user message down to the spacer
    const lastExchangeHeight = spacer.offsetTop - lastUser.offsetTop;
    const remaining = container.clientHeight - lastExchangeHeight;

    // Spacer fills remaining space but not less than MIN_SPACER
    spacer.style.minHeight = `${Math.max(MIN_SPACER, remaining)}px`;

    // Restore scroll position so the user doesn’t get yanked
    container.scrollTop = prevScrollTop;
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
              <ThreadUserMessage
                message={message}
                onCopy={onCopy}
                onEdit={onEdit}
              />
            ) : (
              <ThreadAssistantMessage
                message={message}
                onCopy={onCopy}
                onRetry={onRetry}
              />
            )}
          </div>
        ))}

        <ThreadWaitingIndicator isWaiting={isWaiting} />

        <div ref={spacerRef} className="flex-1 shrink-0" />
      </div>

      <ThreadScrollButton
        showButton={showButton}
        onClick={onScrollButtonClick}
      />
    </div>
  );
}

type ThreadUserMessageProps = {
  message: UIMessage;
  onCopy: (text: string) => void;
  onEdit: (message: UIMessage) => void;
};

const ThreadUserMessage = memo(function ThreadUserMessage({
  message,
  onCopy,
  onEdit,
}: ThreadUserMessageProps) {
  return (
    <div className="flex justify-start">
      <div dir="auto" className="bg-muted rounded-2xl px-3 py-2 max-w-[80%]">
        {message.parts.filter(isTextUIPart).map((part, i) => (
          <span key={i}>{part.text}</span>
        ))}
      </div>
    </div>
  );
});

type ThreadAssistantMessageProps = {
  message: UIMessage;
  onCopy: (text: string) => void;
  onRetry: (messageID: string) => void;
};

const ThreadAssistantMessage = memo(function ThreadAssistantMessage({
  message,
  onCopy,
  onRetry,
}: ThreadAssistantMessageProps) {
  return (
    <div dir="auto">
      {message.parts.filter(isTextUIPart).map((part, i) => (
        <Streamdown key={i}>{part.text}</Streamdown>
      ))}
    </div>
  );
});

type ThreadWaitingIndicatorProps = React.ComponentPropsWithoutRef<"div"> & {
  isWaiting: boolean;
};

const ThreadWaitingIndicator = memo(function ThreadWaitingIndicator({
  isWaiting,
  className,
  ...props
}: ThreadWaitingIndicatorProps) {
  if (!isWaiting) return null;

  return (
    <div className={cn("py-2 opacity-40 text-sm", className)} {...props}>
      ...
    </div>
  );
});

type ThreadScrollButtonProps = HTMLMotionProps<"button"> & {
  showButton: boolean;
  onClick: () => void;
};

const ThreadScrollButton = memo(function ThreadScrollButton({
  showButton,
  onClick,
  className,
  ...props
}: ThreadScrollButtonProps) {
  return (
    <AnimatePresence>
      {showButton && (
        <motion.button
          initial={{ opacity: 0, scale: 0.6 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.6 }}
          transition={{ duration: 0.15, ease: "easeOut" }}
          onClick={onClick}
          className={cn(
            "absolute bottom-2 left-1/2 -translate-x-1/2 bg-background border border-border rounded-full p-1 shadow-md",
            className,
          )}
          {...props}
        >
          <IconArrowNarrowDown className="size-4" />
        </motion.button>
      )}
    </AnimatePresence>
  );
});
