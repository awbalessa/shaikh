# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Working Mode

**Claude's role in this project is director, not implementer.**

Default behavior:
- Teach, guide, strategize, and plan — in conversation
- Write code as text in responses (for the user to copy and apply)
- **Never write or edit files** unless explicitly told to ("write it", "apply it", "go ahead and implement")

The user implements. Claude explains everything.

## External SDK Documentation

**Always use context7 MCP for any external SDK docs. Never rely on training data.**

- Vercel AI SDK → context7 library ID: `/vercel/ai`
- Gemini SDK, or any other SDK → resolve via `mcp__context7__resolve-library-id` first

This is especially important for Vercel AI SDK because this project mirrors its abstractions closely. When in doubt about an API, event format, hook option, or protocol detail — query context7 before writing any code.

## AI Feature Design Principle

**All AI features — frontend and backend — must mirror the Vercel AI SDK paradigm.**

The backend is in Go, so Vercel AI SDK Core concepts are ported manually. Any new AI feature should:
1. Follow the Vercel AI SDK Core abstraction layer (`internal/app/ai/`) for the Go side
2. Use Vercel AI SDK UI (`@ai-sdk/react`, `useChat`, etc.) for the frontend
3. Stream via the **UI message stream protocol** (`x-vercel-ai-ui-message-stream: v1`)
4. Use `smoothStream` or equivalent word-boundary buffering on the **backend** — never on the frontend
5. Use `streamdown` for rendering streaming markdown — not `react-markdown` with custom word-reveal hacks

## Commands

```bash
# Run the server (from api/)
go run ./cmd/api/main.go

# Build
go build ./...

# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/app/ai/...

# Run a single test
go test ./internal/app/ai/... -run TestName

# Vet and check for issues
go vet ./...
```

Requires a `.env` file with `GEMINI_API_KEY` set. `ENVIRONMENT` defaults to `dev`. `PORT` defaults to `8080`.

## About the project

**Shaikh** is an AI-powered Quran app. The backend streams AI responses to a Next.js frontend using the Vercel AI SDK UI message stream protocol.

The overall backend architecture follows **Domain-Driven Design (DDD)**. The `internal/domain/` package is reserved for domain models as they emerge.

## Current status

End-to-end streaming is fully working: Go backend (Gemini) → BFF route at `web/app/api/chat/route.ts` → `useChat` hook in Next.js. Both backend and frontend are connected and streaming tokens in real time.

## Architecture

`cmd/main.go` starts an HTTP server on `cfg.Port` (default 8080) using chi router. It wires `gemini.Model` into `chat.Handler`.

### Key layers

**`internal/app/ai/`** — A deliberate port of **Vercel AI SDK Core** concepts into Go. The `Model` interface, `CallOptions`, `Part`/`Content` type hierarchy, `StreamResult`, and `Event` types mirror the Vercel AI SDK Core abstractions so that the backend streaming protocol is compatible with **Vercel AI SDK UI** on the frontend.

**`internal/http/`** — HTTP handlers. `chat.Handler` exposes `POST /chat` that streams via SSE in the Vercel AI UI message stream format (`x-vercel-ai-ui-message-stream: v1`). `sse.go` wraps `http.ResponseWriter` with JSON event flushing.

**`internal/providers/gemini/`** — Google Gemini client and `Model` implementation. `Model.Stream` converts `ai.CallOptions` to Gemini API calls and maps response chunks to `ai.Event` values.

**`config/`** — Env loading (`godotenv`) and structured logging setup. Logger uses text format in `dev`, JSON in production.

### Type system

Messages are `[]Part` (input) or `[]Content` (output). Both are interfaces satisfied by: Text, Reasoning, File, ToolCall, ToolResult, and Source variants. Streaming events use the `Event` struct with an `EventType` discriminator. `Model.Stream` returns `StreamResult{Stream}` — callers use `result.Stream` for the event loop.

### Frontend integration

The Next.js BFF route (`web/app/api/chat/route.ts`) proxies requests from `useChat` to the Go backend. It extracts the user message text from `UIMessage.parts` (AI SDK v3 format) and forwards the SSE response stream directly to the client. The frontend uses `DefaultChatTransport` pointing at `/api/chat`.
