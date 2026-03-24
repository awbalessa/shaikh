"use client";

import { useEffect, useRef, useState } from "react";
import {
  IconArrowNarrowUp,
  IconAdjustmentsHorizontal,
} from "@tabler/icons-react";
import { useChat } from "@ai-sdk/react";
import { cn } from "@/lib/utils";
import { BASE_DIR } from "@/lib/config";

type ChatComposerProps = React.ComponentPropsWithoutRef<"div"> & {
  sendMessage: ReturnType<typeof useChat>["sendMessage"];
  status: ReturnType<typeof useChat>["status"];
};

export default function ChatComposer({
  sendMessage,
  status,
  className,
  ...props
}: ChatComposerProps) {
  const [value, setValue] = useState<string>("");
  const [isTextAreaFocused, setIsTextAreaFocused] = useState<boolean>(false);
  const textAreaRef = useRef<HTMLTextAreaElement | null>(null);

  const MAX_LINES = 10;
  const LINE_HEIGHT_PX = 24;
  const maxHeight = MAX_LINES * LINE_HEIGHT_PX;

  const resize = (): void => {
    const el = textAreaRef.current;
    if (!el) return;

    el.style.height = "auto";
    const desired = el.scrollHeight;
    const next = Math.min(desired, maxHeight);

    el.style.height = `${next}px`;
    el.style.overflowY = desired > maxHeight ? "auto" : "hidden";
  };

  useEffect(() => {
    resize();
  }, [value]);

  const isStreaming = status === "streaming" || status === "submitted";

  const send = (): void => {
    const trimmed = value.trim();
    if (!trimmed || isStreaming) return;

    sendMessage({ text: trimmed });
    setValue("");
  };

  const isEmpty = value.trim().length === 0;

  const focusTextArea = (): void => {
    textAreaRef.current?.focus();
  };

  return (
    <div
      {...props}
      onMouseDown={(e) => {
        e.preventDefault();
        focusTextArea();
      }}
      className={cn(
        "w-full mx-auto flex flex-col py-3 border border-border rounded-lg bg-highlight dark:bg-surface-light shadow-shadow-md transition-colors",
        !isTextAreaFocused && "hover:border-border-strong",
        isTextAreaFocused && "ring-2 ring-inset ring-primary",
        className,
      )}
    >
      <div className="w-full px-4">
        <textarea
          dir={isEmpty ? BASE_DIR : "auto"}
          ref={textAreaRef}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          rows={2}
          className="composer-scroll w-full text-text text-base leading-6 bg-transparent resize-none outline-none caret-text cursor-text placeholder:text-text-muted placeholder:opacity-100"
          placeholder="اسأل شيخ..."
          onFocus={() => setIsTextAreaFocused(true)}
          onBlur={() => setIsTextAreaFocused(false)}
          onKeyDown={(e) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              send();
            }
          }}
        />
      </div>

      <div className="flex flex-row items-center justify-between w-full ps-4 pe-4">
        <button
          type="button"
          onMouseDown={(e) => e.stopPropagation()}
          className="inline-flex items-center justify-center p-1 rounded-lg transition-colors cursor-pointer hover:bg-surface-medium focus:outline-none focus-visible:ring-2 focus-visible:ring-primary"
        >
          <IconAdjustmentsHorizontal
            size={20}
            stroke={2}
            className="text-text-muted"
          />
        </button>

        <button
          type="button"
          disabled={isEmpty || isStreaming}
          onMouseDown={(e) => e.stopPropagation()}
          onClick={() => !isEmpty && !isStreaming && send()}
          className={cn(
            "p-1 rounded-full transition-colors cursor-pointer focus:outline-none focus-visible:ring-2 focus-visible:ring-primary",
            isEmpty || isStreaming
              ? "bg-surface-light dark:bg-surface-medium"
              : "bg-primary hover:bg-primary-hover",
          )}
        >
          <IconArrowNarrowUp
            size={20}
            stroke={2}
            className={cn(
              "text-text-on-primary",
              isEmpty || isStreaming
                ? "dark:text-text-muted"
                : "dark:text-text-on-primary",
            )}
          />
        </button>
      </div>
    </div>
  );
}
