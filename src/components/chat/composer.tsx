"use client";

import { cn } from "@/lib/utils";
import { ChatStatus } from "ai";
import { useEffect, useRef, useState } from "react";
import { IconArrowNarrowUp, IconPlayerStopFilled } from "@tabler/icons-react";
import { AnimatePresence, motion } from "motion/react";
import { useIntlayer } from "next-intlayer";

type ComposerProps = Omit<
  React.ComponentPropsWithoutRef<"form">,
  "onSubmit"
> & {
  value: string;
  status: ChatStatus;
  onValueChange: (v: string) => void;
  onStop: () => Promise<void>;
  onSubmit: (e: React.SubmitEvent<HTMLFormElement>) => void;
};

export default function Composer({
  value,
  status,
  onValueChange,
  onStop,
  onSubmit,
  className,
  ...props
}: ComposerProps) {
  const content = useIntlayer("chat-composer");
  const [focused, setFocused] = useState(false);
  const textAreaRef = useRef<HTMLTextAreaElement | null>(null);

  const isStreaming = status === "streaming" || status === "submitted";

  const handleMouseDown = (e: React.MouseEvent<HTMLFormElement>) => {
    if ((e.target as HTMLElement).closest("button")) return;
    e.preventDefault();
    textAreaRef.current?.focus();
  };

  const handleActionClick = () => {
    if (isStreaming) {
      onStop();
      return;
    }
    textAreaRef.current?.form?.requestSubmit();
  };

  useEffect(() => {
    const el = textAreaRef.current;
    if (!el) return;
    el.style.height = "auto";
    const lineHeight = parseInt(getComputedStyle(el).lineHeight);
    const maxHeight = lineHeight * 10;
    el.style.height = Math.min(el.scrollHeight, maxHeight) + "px";
  }, [value]);

  return (
    <form
      onMouseDown={handleMouseDown}
      onSubmit={onSubmit}
      className={cn(
        "dark:bg-surface relative transition-colors rounded-xl flex flex-col gap-1 shadow",
        focused
          ? "border border-transparent ring-3 ring-primary"
          : "border border-border hover:border-border-strong",
        className,
      )}
      {...props}
    >
      <Input
        ref={textAreaRef}
        value={value}
        onValueChange={onValueChange}
        onFocus={() => setFocused(true)}
        onBlur={() => setFocused(false)}
        placeholder={content.placeholder.value}
      />
      <Footer>
        <div />
        <Action
          isStreaming={isStreaming}
          isEmpty={!value.trim()}
          onClick={handleActionClick}
        />
      </Footer>
    </form>
  );
}

type InputProps = {
  ref: React.RefObject<HTMLTextAreaElement | null>;
  value: string;
  onValueChange: (v: string) => void;
  onFocus: () => void;
  onBlur: () => void;
  placeholder: string;
};

function Input({
  ref,
  value,
  onValueChange,
  onFocus,
  onBlur,
  placeholder,
}: InputProps) {
  const isEmpty = !value.trim();

  const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      e.currentTarget.form?.requestSubmit();
    }
  };

  return (
    <textarea
      dir={isEmpty ? undefined : "auto"}
      ref={ref}
      value={value}
      onChange={(e) => onValueChange(e.target.value)}
      onKeyDown={handleKeyDown}
      onFocus={onFocus}
      onBlur={onBlur}
      rows={2}
      placeholder={placeholder}
      className="w-full resize-none outline-none composer-scroll pt-3 px-3 placeholder:text-text-muted"
    />
  );
}

function Footer({ children }: { children: React.ReactNode }) {
  return (
    <div className="flex flex-row justify-between pb-3 px-3">{children}</div>
  );
}

type ActionProps = {
  isStreaming: boolean;
  isEmpty: boolean;
  onClick: () => void;
};

function Action({ isStreaming, isEmpty, onClick }: ActionProps) {
  const actionKey = isStreaming ? "stop" : "send";

  const arrowClassName = cn(
    "size-4.5",
    isStreaming && "text-on-primary",
    !isStreaming && !isEmpty && "text-on-primary",
    !isStreaming && isEmpty && "text-foreground opacity-25",
  );

  return (
    <button
      type="button"
      onClick={onClick}
      disabled={!isStreaming && isEmpty}
      className={cn(
        "flex items-center justify-center size-7 rounded-full transition-colors duration-200 shrink-0",
        isStreaming && "bg-primary",
        !isStreaming && !isEmpty && "bg-primary",
        !isStreaming && isEmpty && "bg-foreground/8",
      )}
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
          {isStreaming ? (
            <IconPlayerStopFilled className="size-4.5" />
          ) : (
            <IconArrowNarrowUp className={arrowClassName} />
          )}
        </motion.span>
      </AnimatePresence>
    </button>
  );
}
