---
name: Express 5 wildcard route syntax
description: path-to-regexp v8 breaking change — bare * no longer valid in route patterns
---

## Rule

In Express 5 + path-to-regexp v8, use named wildcards:

```ts
// WRONG — throws PathError: Missing parameter name
router.all("*", handler)

// CORRECT
router.all("/{*path}", handler)
```

**Why:** path-to-regexp v8 requires all wildcards to be named parameters. This broke the api-server catch-all proxy at startup with `PathError [TypeError]: Missing parameter name at index 1: *`.

**How to apply:** Whenever writing a catch-all route in this project's Express server, use `/{*path}` (Express 5 syntax) not `*`.
