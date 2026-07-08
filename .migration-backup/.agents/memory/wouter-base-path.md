---
name: Wouter base path routing in Replit multi-artifact setup
description: How to wire wouter correctly when each artifact lives at a different preview path
---

## Setup

Each Vite artifact has `BASE_PATH` env var (e.g., `/admin/`). Configure wouter's Router:

```tsx
<WouterRouter base={import.meta.env.BASE_URL.replace(/\/$/, '')}>
```

This strips the base from all internal route matching, so routes are `/login`, `/orders`, etc. (not `/admin/login`).

## Key behaviors

- `useLocation()[0]` inside the router returns path **relative to base** — use this for nav highlighting
- `history.pushState` must include the full path with base prefix or wouter won't see the route change
- The screenshot tool constructs URLs as `localhost:80{previewPath}{path}` — use `path="/"` not `path="/admin/"` when screenshotting

**Why:** Without the base, wouter matches absolute paths. After auth redirect to `/login`, wouter can't find the route and renders nothing (silent blank page).

**How to apply:** Always wrap the top-level `<Switch>` in `<WouterRouter base={...}>` for sub-path artifacts. Customer app at `/` still needs the wrapper (base becomes empty string).
