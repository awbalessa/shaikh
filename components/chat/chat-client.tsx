"use client";

import { useChat } from "@ai-sdk/react";
import { useState } from "react";
import ChatThread from "./chat-thread";
import ChatComposer from "./chat-composer";
import { useDictionary } from "@/lib/i18n/dictionaries";

export default function ChatClient() {
  const dict = useDictionary().chat;

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
    <div className="flex flex-col h-full">
      <ChatThread
        messages={messages}
        status={status}
        dict={dict.thread}
        className="px-4"
      />
      <div className="px-4 pb-4">
        <ChatComposer
          value={input}
          status={status}
          onSubmit={handleSubmit}
          onStop={stop}
          onValueChange={setInput}
          dict={dict.composer}
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
