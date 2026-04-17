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
import {
  IconArrowNarrowDown,
  IconCheck,
  IconCopy,
  IconEdit,
} from "@tabler/icons-react";
import { AnimatePresence, HTMLMotionProps, motion } from "motion/react";
import { Tooltip, TooltipContent, TooltipTrigger } from "../ui/tooltip";
import { dictionaries } from "@/lib/i18n/dictionaries";

const MIN_SPACER = 64;

type ThreadDict =
  (typeof dictionaries)[keyof typeof dictionaries]["chat"]["thread"];

type ChatThreadProps = React.ComponentPropsWithoutRef<"div"> & {
  messages: UIMessage[];
  status: ChatStatus;
  dict: ThreadDict;
};

export default function ChatThread({
  messages,
  status,
  dict,
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
        {messages.map((message) => (
          <div
            key={message.id}
            ref={message.id === lastUserMessageId ? lastUserRef : undefined}
          >
            {message.role === "user" ? (
              <UserMessage
                message={message}
                onMessageCopy={onCopy}
                onMessageEdit={onEdit}
                dict={dict}
              />
            ) : (
              <AssistantMessage
                message={message}
                onMessageCopy={onCopy}
                onMessageRetry={onRetry}
              />
            )}
          </div>
        ))}

        <WaitingIndicator isWaiting={isWaiting} />

        <Spacer ref={spacerRef} />
      </div>

      <ScrollButton showButton={showButton} onClick={onScrollButtonClick} />
      <div className="absolute bottom-0 inset-x-0 h-6 pointer-events-none composer-fade" />
    </div>
  );
}

type UserMessageProps = React.ComponentPropsWithoutRef<"div"> & {
  message: UIMessage;
  onMessageCopy: (text: string) => void;
  onMessageEdit: (message: UIMessage) => void;
  dict: ThreadDict;
};

const UserMessage = memo(function UserMessage({
  message,
  onMessageCopy,
  onMessageEdit,
  dict,
  className,
}: UserMessageProps) {
  return (
    <div
      dir="ltr"
      className={cn("group flex flex-col w-full items-end gap-1", className)}
    >
      <div dir="auto" className="bg-muted rounded-2xl px-4 py-2 max-w-[80%]">
        {message.parts.filter(isTextUIPart).map((part, i) => (
          <span key={i}>{part.text}</span>
        ))}
      </div>
      <UserMessageActions
        message={message}
        onMessageCopy={onMessageCopy}
        onMessageEdit={onMessageEdit}
        dict={dict}
      />
    </div>
  );
});

type AssistantMessageProps = React.ComponentPropsWithoutRef<"div"> & {
  message: UIMessage;
  onMessageCopy: (text: string) => void;
  onMessageRetry: (messageID: string) => void;
};

const AssistantMessage = memo(function AssistantMessage({
  message,
  onMessageCopy,
  onMessageRetry,
  className,
}: AssistantMessageProps) {
  return (
    <div dir="auto" className={cn("", className)}>
      {message.parts.filter(isTextUIPart).map((part, i) => (
        <Streamdown key={i}>{part.text}</Streamdown>
      ))}
    </div>
  );
});

type MessageActionButtonProps = React.ComponentPropsWithoutRef<"button"> & {
  label: React.ReactNode;
  open?: boolean;
  onOpenChange?: (open: boolean) => void;
  children: React.ReactNode;
};

function MessageActionButton({
  label,
  open,
  onOpenChange,
  children,
  ...buttonProps
}: MessageActionButtonProps) {
  return (
    <Tooltip open={open} onOpenChange={onOpenChange}>
      <TooltipTrigger asChild>
        <button
          className="rounded-md p-1 text-muted-foreground hover:text-foreground hover:bg-muted
          transition-colors"
          {...buttonProps}
        >
          {children}
        </button>
      </TooltipTrigger>
      <TooltipContent side="bottom">{label}</TooltipContent>
    </Tooltip>
  );
}

function UserMessageActions({
  message,
  onMessageCopy,
  onMessageEdit,
  dict,
}: UserMessageProps) {
  const [copied, setCopied] = useState(false);
  const [hovered, setHovered] = useState(false);

  const handleCopy = useCallback(() => {
    const text = message.parts
      .filter(isTextUIPart)
      .map((p) => p.text)
      .join("");
    onMessageCopy(text);
    setCopied(true);
    setTimeout(() => setCopied(false), 1000);
  }, [message.parts, onMessageCopy]);

  return (
    <div
      className="flex items-center gap-0.5 py-1 opacity-0 group-hover:opacity-100
transition-opacity duration-100"
    >
      <MessageActionButton
        label={
          copied
            ? dict.userMessage.actions.copied
            : dict.userMessage.actions.copy
        }
        open={hovered || copied}
        onOpenChange={setHovered}
        onClick={handleCopy}
        tabIndex={-1}
      >
        <AnimatePresence mode="wait" initial={false}>
          {copied ? (
            <motion.span
              key="check"
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              exit={{ scale: 0 }}
              transition={{ duration: 0.1 }}
              className="block"
            >
              <IconCheck className="size-4" />
            </motion.span>
          ) : (
            <motion.span
              key="copy"
              initial={{ scale: 0 }}
              animate={{ scale: 1 }}
              exit={{ scale: 0 }}
              transition={{ duration: 0.1 }}
              className="block"
            >
              <IconCopy className="size-4" />
            </motion.span>
          )}
        </AnimatePresence>
      </MessageActionButton>

      <MessageActionButton
        label={dict.userMessage.actions.edit}
        onClick={() => onMessageEdit(message)}
        tabIndex={-1}
      >
        <IconEdit className="size-4" />
      </MessageActionButton>
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

function Spacer({
  ref,
  className,
  ...props
}: React.ComponentPropsWithRef<"div">) {
  return <div ref={ref} className={cn("flex-1", className)} {...props} />;
}

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
