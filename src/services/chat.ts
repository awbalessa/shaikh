import {
  convertToModelMessages,
  smoothStream,
  streamText,
  UIMessage,
} from "ai";

export async function Chat(messages: UIMessage[]) {
  return streamText({
    model: "anthropic/claude-haiku-4.5",
    messages: await convertToModelMessages(messages),
    experimental_transform: smoothStream(),
  });
}
