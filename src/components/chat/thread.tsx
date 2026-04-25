"use client";

import { cn } from "@/lib/utils";
import { ChatStatus, UIMessage } from "ai";
import {
  forwardRef,
  memo,
  useCallback,
  useDeferredValue,
  useEffect,
  useLayoutEffect,
  useRef,
  useState,
} from "react";
import { IconArrowNarrowDown } from "@tabler/icons-react";
import { AnimatePresence, HTMLMotionProps, motion } from "motion/react";
import Message from "./message";

const MIN_SPACER = 64;

type ThreadProps = React.ComponentPropsWithoutRef<"div"> & {
  messages: UIMessage[];
  status: ChatStatus;
  editingMessageID: string | null;
  onStartEditing: (id: string) => void;
  onStopEditing: () => void;
  onEditMessage: (id: string, text: string) => void;
};

export default function Thread({
  messages,
  status,
  editingMessageID,
  onStartEditing,
  onStopEditing,
  onEditMessage,
  className,
  ...props
}: ThreadProps) {
  const deferredMessages = useDeferredValue(messages);

  const isWaiting = status === "submitted" && messages.at(-1)?.role === "user";
  const lastUserMessageId = messages.findLast((m) => m.role === "user")?.id;

  const containerRef = useRef<HTMLDivElement>(null);
  const lastUserRef = useRef<HTMLDivElement>(null);
  const spacerRef = useRef<HTMLDivElement>(null);
  const [showButton, setShowButton] = useState(false);

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
  }, [deferredMessages.length]); // eslint-disable-line

  useLayoutEffect(() => {
    if (status !== "ready" && status !== "error") return;
    const container = containerRef.current;
    const spacer = spacerRef.current;
    const lastUser = lastUserRef.current;
    if (!container || !spacer || !lastUser) return;
    const prevScrollTop = container.scrollTop;
    spacer.style.minHeight = "0px";
    const lastExchangeHeight = spacer.offsetTop - lastUser.offsetTop;
    const remaining = container.clientHeight - lastExchangeHeight;
    spacer.style.minHeight = `${Math.max(MIN_SPACER, remaining)}px`;
    container.scrollTop = prevScrollTop;
  }, [status]);

  return (
    <div className="relative flex-1 min-h-0">
      <div
        ref={containerRef}
        className={cn(
          "h-full overflow-y-auto outline-none focus-visible:outline-none messages-scroll flex flex-col",
          className,
        )}
        {...props}
      >
        {deferredMessages.map((message) => (
          <div
            key={message.id}
            ref={message.id === lastUserMessageId ? lastUserRef : undefined}
          >
            <Message
              message={message}
              editingMessageID={editingMessageID}
              onStartEditing={onStartEditing}
              onStopEditing={onStopEditing}
              onEditMessage={onEditMessage}
            />
          </div>
        ))}
        <WaitingIndicator isWaiting={isWaiting} />
        <Spacer ref={spacerRef} />
      </div>

      <ScrollButton showButton={showButton} onClick={onScrollButtonClick} />
      <div className="absolute bottom-0 inset-x-0 h-6 pointer-events-none thread-fade" />
    </div>
  );
}

type WaitingIndicatorProps = HTMLMotionProps<"div"> & {
  isWaiting: boolean;
};

export const WaitingIndicator = memo(function WaitingIndicator({
  isWaiting,
  className,
  ...props
}: WaitingIndicatorProps) {
  if (!isWaiting) return null;

  return (
    <motion.div
      initial={{ opacity: 1, scale: 1 }}
      animate={{ opacity: 0.6, scale: 0.8 }}
      transition={{
        duration: 1.2,
        repeat: Infinity,
        repeatType: "mirror",
        ease: "easeInOut",
      }}
      className={cn(
        "size-3 shrink-0 rounded-full bg-foreground my-2",
        className,
      )}
      {...props}
    ></motion.div>
  );
});

const Spacer = forwardRef<
  HTMLDivElement,
  React.ComponentPropsWithoutRef<"div">
>(function Spacer({ className, ...props }, ref) {
  return <div ref={ref} className={cn("flex-1", className)} {...props} />;
});

type ScrollButtonProps = HTMLMotionProps<"button"> & {
  showButton: boolean;
  onClick: () => void;
};

const ScrollButton = memo(function ScrollButton({
  showButton,
  onClick,
  className,
  ...props
}: ScrollButtonProps) {
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
            "absolute z-50 bottom-2 left-1/2 -translate-x-1/2 bg-surface-floating border border-border focus-visible:border-transparent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary rounded-full p-1.5",
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
