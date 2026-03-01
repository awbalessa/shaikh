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

Requires a `.env` file with `GEMINI_API_KEY` set. `ENVIRONMENT` defaults to `dev`.

## About the project

**Shaikh** is an AI-powered Quran app. The backend is in early development — the goal for the current phase is to get AI token streaming working end-to-end between the Go backend and a frontend composer built with Vercel AI SDK UI.

The overall backend architecture follows **Domain-Driven Design (DDD)**. The `internal/domain/` package is reserved for domain models as they emerge.

## Architecture

The entry point (`cmd/main.go`) initializes config and logging but does not yet start an HTTP server — wiring that up is the current next step.

### Key layers

**`internal/app/ai/`** — A deliberate port of **Vercel AI SDK Core** concepts into Go. The `Model` interface, `CallOptions`, `Part`/`Content` type hierarchy, `StreamResult`, and `Event` types mirror the Vercel AI SDK Core abstractions so that the backend streaming protocol is compatible with **Vercel AI SDK UI** on the frontend. `FakeModel` is a test double for developing without hitting the API.

**`internal/http/`** — HTTP handlers. `chat.Handler` exposes a `/chat` endpoint that streams via SSE in the Vercel AI UI message stream format (`x-vercel-ai-ui-message-stream: v1`). `sse.go` wraps `http.ResponseWriter` with JSON event flushing.

**`internal/providers/gemini/`** — Google Gemini client initialization (wraps `google.golang.org/genai`). The Gemini provider needs to be wired to implement the `ai.Model` interface.

**`config/`** — Env loading (`godotenv`) and structured logging setup. Logger uses text format in `dev`, JSON in production.

### Type system

Messages are `[]Part` (input) or `[]Content` (output). Both are interfaces satisfied by: Text, Reasoning, File, ToolCall, ToolResult, and Source variants. Streaming events use the `Event` struct with an `EventType` discriminator.

### Known issues in current code

- `internal/http/chat/chat.go:14` — references `ai.LModel` which does not exist; the correct type is `ai.Model`
- `internal/http/chat/chat.go:83` — uses `=` (assignment) instead of `==` (comparison) when checking `ev.Type == ai.EventFinish`
