"use client";
import { useEffect, useRef, useState } from "react";
import {
  IconArrowNarrowUp,
  IconAdjustmentsHorizontal,
} from "@tabler/icons-react";

export default function Page() {
  return (
    <main className="h-dvh flex bg-bg text-text min-h-0">
      <section className="basis-7/12 min-w-[590px]">
        <div className="mx-auto h-full px-8 pb-8 pt-6"></div>
      </section>
      <ChatPane className="basis-5/12 border-s-[0.5px] border-border min-w-[390px] max-w-[850px]"></ChatPane>
    </main>
  );
}

type ChatPaneProps = React.ComponentPropsWithoutRef<"section">;

function ChatPane({ className, ...props }: ChatPaneProps) {
  return (
    <section
      {...props}
      className={["flex flex-col h-full min-h-0", className ?? ""].join(" ")}
    >
      <div className="flex flex-col mx-auto w-full h-full px-8 pb-8 pt-6">
        <ChatMessages className="flex-1 min-h-0 overflow-y-auto pb-4" />
        <ChatComposer />
      </div>
    </section>
  );
}

type ChatMessagesProps = React.ComponentPropsWithoutRef<"div">;

function ChatMessages({ className, ...props }: ChatMessagesProps) {
  return <div {...props} className={className}></div>;
}

type ChatComposerProps = React.ComponentPropsWithoutRef<"div">;

type AppLang = "ar" | "en";
const APP_LANG: AppLang = "ar";

type Dir = "rtl" | "ltr";
const baseDir: Dir = APP_LANG === "ar" ? "rtl" : "ltr";

function ChatComposer({ className, ...props }: ChatComposerProps) {
  const [value, setValue] = useState<string>("");
  const textAreaRef = useRef<HTMLTextAreaElement | null>(null);

  const MAX_LINES = 8;
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

  const autoDir = (): void => {
    const el = textAreaRef.current;
    if (!el) return;

    const currentDir = el?.dir;
    el.dir = el.value === "" ? currentDir : "auto";
  };

  useEffect(() => {
    resize();
  }, [value]);

  useEffect(() => {
    autoDir();
  }, [value]);

  const send = (): void => {
    const trimmed = value.trim();
    if (!trimmed) return;

    console.log("send:", trimmed);
    setValue("");
  };

  const isEmpty = value.trim().length === 0;

  return (
    <div
      {...props}
      className={[
        "w-full flex flex-col gap-5 py-3 border border-border rounded-md bg-highlight shadow-md focus-within:border-primary focus-within:border-2",
        className ?? "",
      ].join(" ")}
    >
      <div className="w-full px-4">
        <textarea
          dir={isEmpty ? baseDir : "auto"}
          ref={textAreaRef}
          value={value}
          onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
            setValue(e.target.value)
          }
          rows={1}
          className="w-full text-text text-base leading-6 bg-transparent resize-none outline-none caret-primary placeholder:text-text-muted placeholder:opacity-100"
          placeholder="اسأل شيخ..."
          onKeyDown={(e: React.KeyboardEvent<HTMLTextAreaElement>) => {
            if (e.key === "Enter" && !e.shiftKey) {
              e.preventDefault();
              send();
            }
          }}
        ></textarea>
      </div>
      <div className="flex flex-row justify-between w-full ps-3 pe-4">
        <button
          type="button"
          className="p-1 rounded-sm transition-colors hover:bg-surface focus:outline-none focus:ring-2 focus:ring-primary"
          aria-label="Options"
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
          onClick={() => !isEmpty && send()}
          className={[
            "p-1 bg-primary rounded-sm transition-colors hover:bg-surface focus:outline-none focus:ring-2 focus:ring-primary",
            isEmpty
              ? "bg-primary-off cursor-not-allowed"
              : "bg-primary hover:bg-primary-hover",
          ].join(" ")}
          aria-label="Send"
        >
          <IconArrowNarrowUp size={24} stroke={2} className="text-bg" />
        </button>
      </div>
    </div>
  );
}
