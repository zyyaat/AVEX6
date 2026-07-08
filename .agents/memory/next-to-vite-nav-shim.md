---
name: Next.js → Vite navigation shim
description: Rules for the @/lib/navigation.tsx drop-in shim that replaces next/navigation and next/link in Vite ports
---

## Rules

1. **useSearchParams() must return URLSearchParams directly** (not a tuple). Next.js 13+ App Router returns the object, not [params, setter]. Returning a tuple causes `.get()` calls to fail at runtime.

2. **router.push/replace must prepend BASE_URL**. When Vite deploys at a sub-path (e.g., `/admin/`), calling `history.pushState('/login', ...)` sets the URL to `/login` which is outside wouter's base. Fix:
   ```ts
   const BASE = (import.meta.env.BASE_URL || '/').replace(/\/$/, '');
   function toFullPath(href: string) { return href.startsWith('/') ? BASE + href : href; }
   ```

3. **usePathname() returns the wouter-relative path** (after base stripping). This is correct for nav highlighting — matches how Next.js App Router reports pathname inside sub-path deploys.

4. **useParams() re-exports wouter's useParams** with a generic type cast. Direct re-export is sufficient.

5. **Link component must also use toFullPath()** for href and history.pushState.

**Why:** The shim is used in 5 apps each at different base paths (`/`, `/admin/`, `/driver/`, `/merchant/`, `/support/`). Without BASE_URL prepending, all router.push calls target the wrong URL and wouter never sees the intended route.

**How to apply:** Copy the canonical shim from any `artifacts/*/src/lib/navigation.tsx` when porting additional apps.
