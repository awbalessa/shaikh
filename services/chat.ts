import { convertToModelMessages, streamText, UIMessage } from "ai";

export async function Chat(messages: UIMessage[]) {
  return streamText({
    model: "google/gemini-3-flash",
    messages: await convertToModelMessages(messages),
  });
}
