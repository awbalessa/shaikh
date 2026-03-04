"use client";

import { useChat } from "@ai-sdk/react";
import { DefaultChatTransport } from "ai";
import ChatMessages from "./chat-messages";
import ChatComposer from "./chat-composer";

export default function ChatClient() {
  const { messages, status, sendMessage } = useChat({
    transport: new DefaultChatTransport({ api: "/api/chat" }),
  });

  return (
    <>
      <ChatMessages
        messages={messages}
        status={status}
        className="flex-1 min-h-0 overflow-y-auto pb-4"
      />
      <ChatComposer sendMessage={sendMessage} status={status} />
    </>
  );
}
