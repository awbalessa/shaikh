"use client";

import { useChat } from "@ai-sdk/react";
import { useCallback, useState } from "react";
import Thread from "./thread";
import Composer from "./composer";

export default function Client() {
  const { messages, status, sendMessage, stop } = useChat();
  const [input, setInput] = useState("");
  const [editingMessageID, setEditingMessageID] = useState<string | null>(null);
  const handleStartEditing = useCallback(
    (id: string) => setEditingMessageID(id),
    [],
  );
  const handleStopEditing = useCallback(() => setEditingMessageID(null), []);

  const handleEditMessage = useCallback(
    (id: string, text: string) => {
      sendMessage({ text, messageId: id });
      setEditingMessageID(null);
    },
    [sendMessage],
  );

  const handleSubmit = (e: React.SubmitEvent<HTMLFormElement>) => {
    e.preventDefault();
    const trimmed = input.trim();
    if (!trimmed || status === "streaming" || status === "submitted") return;
    setEditingMessageID(null);
    sendMessage({ text: trimmed });
    setInput("");
  };

  return (
    <div className="flex flex-col h-full">
      <Thread
        messages={messages}
        status={status}
        editingMessageID={editingMessageID}
        onStartEditing={handleStartEditing}
        onStopEditing={handleStopEditing}
        onEditMessage={handleEditMessage}
        className="px-4"
      />

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
