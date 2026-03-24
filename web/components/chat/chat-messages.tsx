"use client";

import { useEffect, useMemo, useRef, useState } from "react";
import { isTextUIPart, type UIMessage } from "ai";
import { useChat } from "@ai-sdk/react";
import { Streamdown } from "streamdown";

type ChatMessagesProps = React.ComponentPropsWithoutRef<"div"> & {
  messages: UIMessage[];
  status: ReturnType<typeof useChat>["status"];
};

const md: React.ComponentProps<typeof Streamdown>["components"] = {
  h1: ({ children, ...props }) => (
    <h1 className="text-4xl leading-12 font-semibold pt-10" {...props}>
      {children}
    </h1>
  ),
  h2: ({ children, ...props }) => (
    <h2 className="text-2xl leading-9 font-semibold pt-8" {...props}>
      {children}
    </h2>
  ),
  h3: ({ children, ...props }) => (
    <h3 className="text-lg leading-7 font-semibold pt-6" {...props}>
      {children}
    </h3>
  ),
  p: ({ children, ...props }) => (
    <p className="text-base leading-6 pt-4 first:pt-0" {...props}>
      {children}
    </p>
  ),
  ul: ({ children }) => (
    <ul className="pt-4 gap-3 flex flex-col list-none">{children}</ul>
  ),
  ol: ({ children }) => (
    <ol className="pt-4 gap-3 flex flex-col list-none">{children}</ol>
  ),
  li: ({ children }) => (
    <li className="flex items-baseline gap-1 text-base leading-6">
      <span aria-hidden>·</span>
      <span>{children}</span>
    </li>
  ),
  hr: () => <hr className="border-t border-divider mt-8" />,
};

export default function ChatMessages({
  messages,
  status,
  className,
  ...props
}: ChatMessagesProps) {
  return <div></div>;
}
