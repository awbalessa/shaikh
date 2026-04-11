"use client";

import { dictionaries } from "@/lib/i18n/dictionaries";
import { useLocale } from "@/lib/i18n/locale-context";
import { useDirection } from "../ui/direction";
import { useChat } from "@ai-sdk/react";
import { useState } from "react";
import ChatMessages from "./chat-messages";
import ChatComposer from "./chat-composer";

export default function ChatClient() {
  const dir = useDirection();
  const { locale } = useLocale();
  const dict = dictionaries[locale];

  const { messages, status, sendMessage, stop } = useChat();
  const [input, setInput] = useState("");

  const handleSubmit = (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    const trimmed = input.trim();
    if (!trimmed || status === "streaming" || status === "submitted") return;

    sendMessage({ text: trimmed });
    setInput("");
  };

  return (
    <div dir={dir} className="flex flex-col h-full">
      <ChatMessages messages={messages} status={status} className="px-4" />
      <div className="composer-fade h-6" />
      <div className="px-4 pb-4">
        <ChatComposer
          value={input}
          status={status}
          onSubmit={handleSubmit}
          onStop={stop}
          onValueChange={setInput}
          dict={dict.chat.composer}
        >
          <ChatComposer.Input />
          <ChatComposer.Footer>
            <ChatComposer.Footer.Start />
            <ChatComposer.Footer.End>
              <ChatComposer.Action />
            </ChatComposer.Footer.End>
          </ChatComposer.Footer>
        </ChatComposer>
      </div>
    </div>
  );
}
