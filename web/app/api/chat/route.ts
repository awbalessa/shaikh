type ChatRequestBody = {
  message: string;
  conversationId?: string;
};

export async function POST(req: Request) {
  const body = (await req.json()) as ChatRequestBody;
  const { message } = body;

  if (typeof message !== "string" || !message.trim()) {
    return new Response("bad request", { status: 400 });
  }

  let upstream: Response;
  try {
    upstream = await fetch("http://localhost:8080/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ message: message.trim() }),
    });
  } catch {
    return new Response("upstream unavailable", { status: 502 });
  }

  return new Response(upstream.body, {
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "x-vercel-ai-ui-message-stream": "v1",
    },
  });
}
