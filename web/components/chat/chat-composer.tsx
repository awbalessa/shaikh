"use client";

import { useEffect, useRef, useState } from "react";
import {
  IconArrowNarrowUp,
  IconAdjustmentsHorizontal,
} from "@tabler/icons-react";

type ChatComposerProps = React.ComponentPropsWithoutRef<"div">;

type AppLang = "ar" | "en";
const APP_LANG: AppLang = "ar";

type Dir = "rtl" | "ltr";
const baseDir: Dir = APP_LANG === "ar" ? "rtl" : "ltr";

export default function ChatComposer({
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

  const send = (): void => {
    const trimmed = value.trim();
    if (!trimmed) return;

    console.log("send:", trimmed);
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
      className={[
        "w-full flex flex-col gap-1 py-3 border border-border rounded-md bg-surface-light shadow-md",
        isTextAreaFocused ? "border-2 border-primary" : "",
        className ?? "",
      ].join(" ")}
    >
      <div className="w-full px-4">
        <textarea
          dir={isEmpty ? baseDir : "auto"}
          ref={textAreaRef}
          value={value}
          onChange={(e) => setValue(e.target.value)}
          rows={1}
          className="w-full text-text text-base leading-6 bg-transparent resize-none outline-none caret-text cursor-text placeholder:text-text-muted placeholder:opacity-100"
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

      <div className="flex flex-row items-center justify-between w-full ps-3 pe-4">
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
          disabled={isEmpty}
          onMouseDown={(e) => e.stopPropagation()}
          onClick={() => !isEmpty && send()}
          className={[
            "p-1.5 bg-primary rounded-sm transition-colors cursor-pointer focus:outline-none focus-visible:ring-2 focus-visible:ring-primary",
            isEmpty ? "bg-primary-off" : "hover:bg-primary-hover",
          ].join(" ")}
        >
          <IconArrowNarrowUp
            size={24}
            stroke={2}
            className={[
              "text-text-on-primary",
              isEmpty ? "dark:text-text-muted" : "dark:text-text-on-primary",
            ].join(" ")}
          />
        </button>
      </div>
    </div>
  );
}
