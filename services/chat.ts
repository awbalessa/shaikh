import { convertToModelMessages, streamText, UIMessage } from "ai";
import { models, ModelTier } from "@/lib/ai";

export async function streamReply(
  tier: ModelTier = "medium",
  messages: UIMessage[],
) {
  return streamText({
    model: models[tier],
    messages: await convertToModelMessages(messages),
  });
}
