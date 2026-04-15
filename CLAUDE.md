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
- AI Elements → context7 or invoke `ai-elements` skill before working with its components
- Any other SDK → resolve via `mcp__context7__resolve-library-id` first

When in doubt about any API, hook option, component prop, or protocol detail — query context7 before writing any code.

## AI Feature Design Principle

**All AI features must follow the Vercel AI SDK paradigm end-to-end.**

The stack is: **custom chat UI → `useChat` → BFF route → `streamText` + Vercel Gateway → Gemini**

Any new AI feature should:
1. Use Vercel AI SDK UI (`@ai-sdk/react`, `useChat`, etc.) for hook-level state
2. Build chat UI as custom components in `components/chat/` — no UI framework dependency
3. Route through the BFF at `app/api/chat/route.ts` using `streamText().toUIMessageStreamResponse()`
4. Use `streamdown` for rendering streaming markdown in assistant messages
5. Use `smoothStream` word-boundary buffering in `streamText` options — never on the frontend

## Commands

```bash
# Dev server (from project root)
bun run dev

# Build
bun run build

# Lint
bun run lint
```

## About the project

**Shaikh** is an AI-powered Quran app. The frontend is a root-level Next.js 16 app using the App Router.

No separate backend. All AI routing goes through the Vercel AI Gateway via `streamText` in the BFF route.

## Current status

Building a custom chat UI from scratch. No shadcn, no AI Elements. The stack is fully Vercel: custom components + `useChat` + Vercel AI Gateway + Gemini. The old Go backend has been removed.

## Architecture

The project is a single Next.js app at the repository root.

### Key layers

**`app/api/chat/route.ts`** — BFF route. Accepts `UIMessage[]` from `useChat`, calls `services/chat.ts`, returns SSE via `.toUIMessageStreamResponse()`.

**`services/chat.ts`** — AI logic. Calls `streamText` with the Vercel Gateway model string. This is where system prompts, tools, and model selection live.

**`components/chat/`** — All chat UI components wired to `useChat`. Custom-built, no UI framework dependency.

**`components/ui/`** — Custom Radix UI wrappers (hand-written, not shadcn-scaffolded). Uses the `radix-ui` unified package. Build primitives here as needed (Tooltip, Popover, Dialog, etc.).

**`components/providers/`** — App-level providers (theme, locale, direction).

**`lib/`** — Utilities (`utils.ts`, `config.ts`).

### Frontend integration

`useChat` in `components/chat/chat-client.tsx` points at `/api/chat` via `DefaultChatTransport`. The BFF route streams back using the Vercel AI UI message stream protocol. AI Elements components consume the `messages` and `status` from `useChat`.

### Rendering

Streaming markdown is rendered with `streamdown` inside AI Elements message slots. Never use `react-markdown` with custom word-reveal hacks.
