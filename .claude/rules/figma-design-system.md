# Figma Design System Rules

Rules for implementing Figma designs in the Shaikh codebase using the Figma MCP server.

---

## Project Stack

- **Framework:** Next.js 16 App Router, React 19, TypeScript
- **Styling:** Tailwind CSS v4 with `@theme inline` token mapping
- **Package manager:** Bun
- **Path alias:** `@/` → `src/`

---

## Design Tokens

All tokens are CSS custom properties defined in `src/app/globals.css`.

### Primitive ramps (do not use directly in components)
- Neutral: `--neutral-50` through `--neutral-950` (OKLCH, warm-tinted)
- Primary: `--primary-100` through `--primary-900` (OKLCH, hue 35 — warm amber/brown)

### Semantic aliases (use these in all component code)
| Category | Tokens |
|----------|--------|
| Base | `--background`, `--foreground` |
| Text | `--text-primary`, `--text-secondary`, `--text-tertiary`, `--text-muted`, `--text-inverse`, `--text-link`, `--text-link-hover` |
| Surfaces | `--surface`, `--surface-raised`, `--surface-overlay`, `--surface-tooltip` |
| Borders | `--border-light`, `--border`, `--border-strong`, `--border-focus` |
| Primary | `--primary-subtle`, `--primary-muted`, `--primary`, `--primary-hover`, `--primary-active`, `--on-primary` |
| Component | `--muted`, `--muted-foreground`, `--popover`, `--popover-foreground`, `--ring` |
| Radius | `--radius-xs/sm/md/lg/xl/2xl/3xl/4xl/full` (derived from `--radius: 0.625rem`) |
| Shadow | `--shadow` |

### Tailwind usage
All semantic tokens are mapped via `@theme inline` and available as Tailwind classes:
```
bg-background       text-foreground
bg-surface          text-text-primary
bg-surface-raised   text-text-secondary
border-border       text-text-muted
bg-primary          text-on-primary
ring-primary        text-primary
```

**IMPORTANT:** Never hardcode hex or oklch colors in component code. Always use the semantic token Tailwind classes above.

### Dark mode
Dark mode tokens are defined under `.dark` in `globals.css`. In Tailwind, use the `dark:` variant:
```tsx
className="bg-surface dark:bg-surface-raised"
```

---

## Component Organization

| Directory | Purpose |
|-----------|---------|
| `src/components/chat/` | Chat UI — Thread, Composer, Message, ChatClient |
| `src/components/ui/` | Radix UI primitives (hand-written wrappers, not shadcn) |
| `src/components/` | App-level providers and one-off components |
| `src/app/[locale]/` | Page-level layouts and page components |

**IMPORTANT:** When implementing a Figma design, always check `src/components/ui/` for an existing primitive before creating a new one. Check `src/components/chat/` for chat-specific components.

---

## Component Patterns

### Styling utility
Always use `cn()` from `@/lib/utils` for conditional class merging:
```tsx
import { cn } from "@/lib/utils";
className={cn("base-classes", conditional && "conditional-class", className)}
```

### Radix UI primitives
Import from the unified `radix-ui` package (not `@radix-ui/react-*` individual packages):
```tsx
import { Tooltip } from "radix-ui";
```
Wrap and re-export from `src/components/ui/`. See `tooltip.tsx` for the pattern.

### `className` prop
All components must accept and forward a `className` prop for composition.

### Variants
Use union types for variants, not boolean flags:
```tsx
variant: 'primary' | 'secondary' | 'ghost'
```

---

## Icons

**Primary icon library:** `@tabler/icons-react`
```tsx
import { IconArrowNarrowUp, IconPlayerStopFilled } from "@tabler/icons-react";
<IconArrowNarrowUp className="size-4.5" />
```

**IMPORTANT:** Do not install new icon libraries. Use `@tabler/icons-react` for all new icons. Size icons with Tailwind `size-*` classes.

---

## Animation

Use `motion/react` (Framer Motion v12) for all animations:
```tsx
import { motion, AnimatePresence } from "motion/react";
```
Use `AnimatePresence` for enter/exit transitions. Keep durations short (0.1–0.2s for UI micro-interactions).

---

## Styling Rules

1. **Tailwind v4** — use utility classes; no inline styles
2. **Token-only colors** — no hardcoded values (no hex, no raw oklch)
3. **Semantic tokens** — always prefer semantic aliases over primitive ramps
4. **Dark mode** — use `dark:` variant classes; the `.dark` class is applied to `<html>`
5. **Radius** — use `rounded-*` Tailwind classes backed by `--radius-*` tokens (e.g., `rounded-xl`, `rounded-full`)
6. **Shadows** — use `shadow` class which maps to `--shadow` token
7. **Scrollbars** — custom scrollbar styling is in `globals.css` via `.messages-scroll` and `textarea.composer-scroll`

---

## i18n / String Content

User-visible strings go through `next-intlayer`, not hardcoded:
```tsx
const content = useIntlayer("component-name");
// content.someKey.value
```
When implementing Figma designs, use placeholder strings with `useIntlayer` — do not hardcode English text.

---

## Figma-to-Code Workflow

Follow this sequence for every Figma implementation task:

1. **`get_design_context`** — fetch structured representation for the target node
2. **`get_screenshot`** — get visual reference for the node
3. If response is too large, use `get_metadata` to get the node map, then re-fetch only required nodes
4. Translate the Figma MCP output (React + Tailwind) into this project's conventions:
   - Replace raw hex/oklch with semantic token Tailwind classes
   - Replace generic components with existing ones from `src/components/`
   - Replace inline styles with utility classes
   - Replace non-tabler icons with `@tabler/icons-react` equivalents
5. Validate the implementation against the Figma screenshot before completing

---

## Asset Handling

- **IMPORTANT:** If the Figma MCP server returns a `localhost` source for an image or SVG, use that source directly — do not create placeholders
- Store any downloaded static assets in `public/`
- Do not install new icon packages — all icons come from `@tabler/icons-react`

---

## What Never to Do

- Hardcode colors (hex, rgb, oklch) in component code
- Install new icon libraries
- Use `@radix-ui/react-*` individual packages (use unified `radix-ui` instead)
- Write inline styles
- Create components outside `src/components/` or `src/app/`
- Use relative imports beyond one level (`../../` is a red flag — use `@/` alias)
