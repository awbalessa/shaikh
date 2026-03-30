"use client";

import { useChat } from "@ai-sdk/react";
import { DefaultChatTransport } from "ai";
import ChatMessages from "./chat-messages";
import ChatComposer from "./chat-composer";

export default function ChatClient() {
  const { messages, status, sendMessage } = useChat({
    transport: new DefaultChatTransport({
      api: "/api/chat",
    }),
  });

  return (
    <div className="w-full h-full max-w-150 mx-auto flex flex-col">
      <ChatMessages
        messages={messages}
        status={status}
        className="messages-scroll flex-1 min-h-0 overflow-y-auto pb-4"
      />
      <div
        className="composer-fade pointer-events-none h-6 w-full"
        aria-hidden
      />
      <ChatComposer sendMessage={sendMessage} status={status} />
    </div>
  );
}
