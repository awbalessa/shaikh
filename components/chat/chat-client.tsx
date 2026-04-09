"use client";

import { dictionaries } from "@/lib/i18n/dictionaries";
import { useLocale } from "@/lib/i18n/locale-context";
import { useDirection } from "../ui/direction";
import { useChat } from "@ai-sdk/react";
import { useState } from "react";
import ChatMessages from "./chat-messages";
import {
  ChatComposer,
  ChatComposerFooter,
  ChatComposerInput,
  ChatComposerAction,
} from "./chat-composer";

export default function ChatClient() {
  const { locale } = useLocale();
  const dict = dictionaries[locale];
  const dir = useDirection();

  const { messages, status, sendMessage, stop, regenerate } = useChat();
  const [input, setInput] = useState("");

  const handleSubmit = (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    const trimmed = input.trim();
    if (!trimmed || status === "streaming" || status === "submitted") return;

    sendMessage({ text: trimmed });
    setInput("");
  };

  return (
    <div dir={dir} className="flex flex-col h-full max-w-2xl mx-auto p-2 pb-4">
      <ChatMessages messages={messages} status={status} className="flex-1" />
      <ChatComposer
        onSubmit={handleSubmit}
        onStop={stop}
        onRetry={regenerate}
        value={input}
        onValueChange={setInput}
        status={status}
      >
        <ChatComposerInput placeholder={dict.chat.composer.placeholder} />
        <ChatComposerFooter>
          <div></div>
          <ChatComposerAction />
        </ChatComposerFooter>
      </ChatComposer>
    </div>
  );
}
