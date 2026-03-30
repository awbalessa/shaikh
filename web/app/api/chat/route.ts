import { streamReply } from "@/services/chat";
import { UIMessage } from "ai";

export async function POST(req: Request) {
  const { messages }: { messages: UIMessage[] } = await req.json();

  const result = await streamReply("low", messages);

  return result.toUIMessageStreamResponse();
}
