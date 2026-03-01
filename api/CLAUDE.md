# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Run the server
go run ./cmd/main.go

# Build
go build ./cmd/main.go

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

**Shaikh** is an AI-powered Quran app. The backend is in early development — the goal for the current phase is to get AI token streaming working end-to-end between the Go backend and a Next.js frontend using Vercel AI SDK UI.

The overall backend architecture follows **Domain-Driven Design (DDD)**. The `internal/domain/` package is reserved for domain models as they emerge.

## Current status

SSE streaming is working end-to-end with `fake.Model`. `curl -X POST http://localhost:8080/chat` streams Vercel AI UI message stream format events correctly. Next steps: implement `ai.Model` on the Gemini provider, then wire up the Next.js BFF route and `useChat` frontend.

## Architecture

`cmd/main.go` starts an HTTP server on `cfg.Port` (default 8080) using chi router. It currently wires `fake.Model` into `chat.Handler`.

### Key layers

**`internal/app/ai/`** — A deliberate port of **Vercel AI SDK Core** concepts into Go. The `Model` interface, `CallOptions`, `Part`/`Content` type hierarchy, `StreamResult`, and `Event` types mirror the Vercel AI SDK Core abstractions so that the backend streaming protocol is compatible with **Vercel AI SDK UI** on the frontend.

**`internal/http/`** — HTTP handlers. `chat.Handler` exposes `POST /chat` that streams via SSE in the Vercel AI UI message stream format (`x-vercel-ai-ui-message-stream: v1`). `sse.go` wraps `http.ResponseWriter` with JSON event flushing.

**`internal/providers/fake/`** — `fake.Model` implements `ai.Model`, streaming a hardcoded "Hello world" message across 6 events with 200ms delays. Used for local development without hitting the API.

**`internal/providers/gemini/`** — Google Gemini client initialization only (wraps `google.golang.org/genai`). Does not yet implement `ai.Model`.

**`config/`** — Env loading (`godotenv`) and structured logging setup. Logger uses text format in `dev`, JSON in production.

### Type system

Messages are `[]Part` (input) or `[]Content` (output). Both are interfaces satisfied by: Text, Reasoning, File, ToolCall, ToolResult, and Source variants. Streaming events use the `Event` struct with an `EventType` discriminator. `Model.Stream` returns `StreamResult{Stream}` — callers use `result.Stream` for the event loop.
