"use client";

import { cn } from "@/lib/utils";
import { ChatStatus } from "ai";
import {
  createContext,
  useContext,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { IconArrowNarrowUp, IconPlayerStopFilled } from "@tabler/icons-react";
import { AnimatePresence, motion } from "motion/react";
import { dictionaries } from "@/lib/i18n/dictionaries";

ChatComposer.Input = ChatComposerInput;
ChatComposer.Footer = ChatComposerFooter;
ChatComposerFooter.Start = ChatComposerFooterStart;
ChatComposerFooter.End = ChatComposerFooterEnd;
ChatComposer.Action = ChatComposerAction;

type ComposerDict =
  (typeof dictionaries)[keyof typeof dictionaries]["chat"]["composer"];

type ComposerContextValue = {
  value: string;
  status: ChatStatus;
  onValueChange: (v: string) => void;
  onStop: () => Promise<void>;
  dict: ComposerDict;
  textAreaRef: React.RefObject<HTMLTextAreaElement | null>;
  setFocused: (v: boolean) => void;
};

const ComposerContext = createContext<ComposerContextValue | null>(null);

function useComposer() {
  const ctx = useContext(ComposerContext);
  if (!ctx) throw new Error("useComposer must be within <ChatComposer>");
  return ctx;
}

type ChatComposerProps = React.ComponentPropsWithoutRef<"form"> & {
  value: string;
  status: ChatStatus;
  onValueChange: (v: string) => void;
  onStop: () => Promise<void>;
  dict: ComposerDict;
};

export default function ChatComposer({
  value,
  status,
  onValueChange,
  onStop,
  dict,
  children,
  ...props
}: ChatComposerProps) {
  const [focused, setFocused] = useState(false);
  const textAreaRef = useRef<HTMLTextAreaElement | null>(null);

  const contextValue = useMemo(
    () => ({
      value,
      status,
      onValueChange,
      onStop,
      dict,
      textAreaRef,
      setFocused,
    }),
    [value, status, onValueChange, onStop, dict],
  );

  const handleMouseDown = (e: React.MouseEvent<HTMLFormElement>) => {
    if ((e.target as HTMLElement).closest("button")) return;
    e.preventDefault();
    textAreaRef.current?.focus();
  };

  return (
    <ComposerContext.Provider value={contextValue}>
      <form
        onMouseDown={handleMouseDown}
        className={cn(
          "relative transition-colors rounded-xl flex flex-col gap-1 shadow-md",
          focused
            ? "border border-transparent ring-3 ring-primary"
            : "border border-border hover:border-border-strong",
        )}
        {...props}
      >
        {children}
      </form>
    </ComposerContext.Provider>
  );
}

function ChatComposerInput({
  className,
  ...props
}: React.ComponentPropsWithoutRef<"textarea">) {
  const { value, onValueChange, dict, textAreaRef, setFocused } = useComposer();
  const isEmpty = !value.trim();

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      e.currentTarget.form?.requestSubmit();
    }
  };

  useEffect(() => {
    const el = textAreaRef.current;
    if (!el) return;
    el.style.height = "auto";
    const lineHeight = parseInt(getComputedStyle(el).lineHeight);
    const maxHeight = lineHeight * 10;
    el.style.height = Math.min(el.scrollHeight, maxHeight) + "px";
  }, [value]); // eslint-disable-line react-hooks/exhaustive-deps

  return (
    <textarea
      dir={isEmpty ? "" : "auto"}
      ref={textAreaRef}
      value={value}
      onChange={(e) => onValueChange(e.target.value)}
      onKeyDown={handleKeyDown}
      onFocus={() => setFocused(true)}
      onBlur={() => setFocused(false)}
      rows={2}
      placeholder={dict.placeholder}
      className={cn(
        "w-full resize-none outline-none composer-scroll pt-3 px-3 placeholder:text-text-neutral",
        className,
      )}
      {...props}
    />
  );
}

function ChatComposerFooter({
  children,
  ...props
}: React.ComponentPropsWithoutRef<"div">) {
  return (
    <div className="flex flex-row justify-between pb-3 px-3" {...props}>
      {children}
    </div>
  );
}

function ChatComposerFooterStart({
  children,
  ...props
}: React.ComponentPropsWithoutRef<"div">) {
  return (
    <div className="flex flex-row" {...props}>
      {children}
    </div>
  );
}

function ChatComposerFooterEnd({
  children,
  ...props
}: React.ComponentPropsWithoutRef<"div">) {
  return (
    <div className="flex flex-row" {...props}>
      {children}
    </div>
  );
}

const ChatComposerActionIcons = {
  send: <IconArrowNarrowUp className="size-4.5" />,
  stop: <IconPlayerStopFilled className="size-4.5" />,
};

function ChatComposerAction({
  className,
  ...props
}: React.ComponentPropsWithoutRef<"button">) {
  const { value, status, onStop } = useComposer();
  const ref = useRef<HTMLButtonElement>(null);

  const isStreaming = status === "streaming" || status === "submitted";
  const isReady = !isStreaming;
  const isEmpty = !value.trim();

  const actionKey = isStreaming ? "stop" : "send";

  const handleClick = () => {
    if (isStreaming) {
      onStop();
      return;
    }
    ref.current?.form?.requestSubmit();
  };

  return (
    <button
      ref={ref}
      type="button"
      onClick={handleClick}
      disabled={isReady && isEmpty}
      className={cn(
        "flex items-center justify-center size-7 rounded-full transition-colors duration-200 shrink-0",
        isStreaming && "bg-foreground text-background",
        isReady && !isEmpty && "bg-foreground text-background",
        isReady && isEmpty && "bg-muted text-muted-foreground",
        className,
      )}
      {...props}
    >
      <AnimatePresence mode="wait">
        <motion.span
          key={actionKey}
          initial={{ opacity: 0, scale: 0.6 }}
          animate={{ opacity: 1, scale: 1 }}
          exit={{ opacity: 0, scale: 0.6 }}
          transition={{ duration: 0.15, ease: "easeOut" }}
          className="flex items-center justify-center"
        >
          {ChatComposerActionIcons[actionKey]}
        </motion.span>
      </AnimatePresence>
    </button>
  );
}
