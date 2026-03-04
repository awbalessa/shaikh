export async function POST(req: Request) {
  const { messages } = await req.json();
  const last = messages.findLast((m: any) => m.role === "user");
  const text = (last?.parts ?? [])
    .filter((p: any) => p.type === "text")
    .map((p: any) => p.text as string)
    .join("");

  let upstream: Response;
  try {
    upstream = await fetch("http://localhost:8080/chat", {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ message: text }),
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
