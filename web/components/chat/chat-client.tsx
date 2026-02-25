"use client";

import ChatMessages from "./chat-messages";
import ChatComposer from "./chat-composer";

export default function ChatClient() {
  return (
    <>
      <ChatMessages className="flex-1 min-h-0 overflow-y-auto pb-4" />
      <ChatComposer />
    </>
  );
}
