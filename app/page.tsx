"use client";
import { useEffect, useRef, useState } from "react";

export default function Page() {
  return (
    <main className="h-dvh flex flex-col bg-bg text-text">
      <div className="flex items-center h-12 bg-surface px-8 border-b-[0.5px] border-border">
        هذا اهو الهيدر
      </div>

      <div className="flex flex-1 min-h-0">
        <section className="basis-7/12 min-w-[590px]">
          <div className="mx-auto h-full px-8 pb-8 pt-6"></div>
        </section>
        <ChatPane className="basis-5/12 border-s-[0.5px] border-border min-w-[390px] max-w-[850px]"></ChatPane>
      </div>
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

function ChatComposer({ className, ...props }: ChatComposerProps) {
  const [value, setValue] = useState<string>("");
  const textAreaRef = useRef<HTMLTextAreaElement | null>(null);

  const MAX_LINES = 8;
  const LINE_HEIGHT_PX = 24;
  const max_height = MAX_LINES * LINE_HEIGHT_PX;

  const resize = (): void => {
    const el = textAreaRef.current;
    if (!el) return;

    el.style.height = "auto";
    const desired = el.scrollHeight;
    const next = Math.min(desired, max_height);

    el.style.height = `${next}px`;
    el.style.overflowY = desired > max_height ? "auto" : "hidden";
  };

  useEffect(() => {
    resize();
  }, [value]);

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
          ref={textAreaRef}
          value={value}
          onChange={(e: React.ChangeEvent<HTMLTextAreaElement>) =>
            setValue(e.target.value)
          }
          rows={1}
          className="w-full py-0 text-text text-base leading-6 bg-transparent resize-none outline-none caret-primary placeholder:text-muted placeholder:opacity-100"
          placeholder="اسأل شيخ"
        ></textarea>
      </div>
    </div>
  );
}
