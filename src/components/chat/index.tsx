"use client";

import { useChat } from "@ai-sdk/react";
import { useState } from "react";
import Thread from "./thread";
import Composer from "./composer";

export default function Client() {
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
      <Thread messages={messages} status={status} className="px-4" />

      <div className="px-4 pb-4">
        <Composer
          value={input}
          status={status}
          onSubmit={handleSubmit}
          onStop={stop}
          onValueChange={setInput}
        />
      </div>
    </div>
  );
}
