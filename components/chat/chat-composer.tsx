"use client";

import { cn } from "@/lib/utils";
import { ChatStatus } from "ai";
import { createContext, useContext, useRef } from "react";
import {
  IconArrowNarrowUp,
  IconPlayerStopFilled,
  IconRefresh,
} from "@tabler/icons-react";
import { AnimatePresence, motion } from "motion/react";

type ComposerContextValue = {
  value: string;
  status: ChatStatus;
  onValueChange: (v: string) => void;
  onStop: () => Promise<void>;
  onRetry: () => Promise<void>;
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
  onRetry: () => Promise<void>;
};

function ChatComposer({
  value,
  status,
  onValueChange,
  onStop,
  onRetry,
  children,
  className,
  ...props
}: ChatComposerProps) {
  return (
    <ComposerContext.Provider
      value={{ value, status, onValueChange, onStop, onRetry }}
    >
      <form
        className={cn(
          "border border-border rounded-xl flex flex-col",
          className,
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
  const { value, onValueChange } = useComposer();

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      e.currentTarget.form?.requestSubmit();
    }
  };

  return (
    <textarea
      value={value}
      onChange={(e) => onValueChange(e.target.value)}
      onKeyDown={handleKeyDown}
      rows={2}
      className={cn(
        "w-full resize-none bg-transparent outline-none",
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
    <div className="flex flex-row justify-between" {...props}>
      {children}
    </div>
  );
}

function ChatComposerAction({
  className,
  ...props
}: React.ComponentPropsWithoutRef<"button">) {
  const { value, status, onStop, onRetry } = useComposer();
  const ref = useRef<HTMLButtonElement>(null);

  const isStreaming = status === "streaming" || status === "submitted";
  const isError = status === "error";
  const isReady = !isStreaming && !isError;
  const isEmpty = !value.trim();

  const actionKey = isStreaming ? "stop" : isError ? "retry" : "send";

  const icons = {
    send: <IconArrowNarrowUp className="size-5" />,
    stop: <IconPlayerStopFilled className="size-5" />,
    retry: <IconRefresh className="size-5" />,
  };

  const handleClick = () => {
    if (isStreaming) {
      onStop();
      return;
    }
    if (isError) {
      onRetry();
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
        "flex items-center justify-center p-1 rounded-full transition-colors duration-200 shrink-0",
        isStreaming && "bg-foreground text-background",
        isError && "bg-destructive text-destructive-foreground",
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
          {icons[actionKey]}
        </motion.span>
      </AnimatePresence>
    </button>
  );
}

export {
  ChatComposer,
  ChatComposerFooter,
  ChatComposerInput,
  ChatComposerAction,
};
