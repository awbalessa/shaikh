"use client";

import { useChat } from "@ai-sdk/react";
import { DefaultChatTransport, isTextUIPart } from "ai";
import ChatMessages from "./chat-messages";
import ChatComposer from "./chat-composer";

export default function ChatClient() {
  const { messages, status, sendMessage } = useChat({
    experimental_throttle: 50,
    transport: new DefaultChatTransport({
      api: "/api/chat",
      prepareSendMessagesRequest: ({ messages: msgs }) => {
        const lastUser = msgs.findLast((m) => m.role === "user");
        const text = (lastUser?.parts ?? [])
          .filter(isTextUIPart)
          .map((p) => p.text)
          .join("");
        return { body: { message: text } };
      },
    }),
  });

  return (
    <>
      <ChatMessages
        messages={messages}
        status={status}
        className="messages-scroll flex-1 min-h-0 overflow-y-auto pb-4"
      />
      <div className="composer-zone shrink-0">
        <div
          className="composer-fade pointer-events-none h-6 w-full"
          aria-hidden
        />
        <ChatComposer sendMessage={sendMessage} status={status} />
      </div>
    </>
  );
}
