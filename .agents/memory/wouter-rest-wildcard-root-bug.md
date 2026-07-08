---
name: Wouter wildcard root-path bug
description: A ":param*" catch-all wouter Route never matches the bare root path "/", causing a fully blank white screen with zero console errors and zero component mount logs.
---

# Wouter `:param*` wildcard does not match `/`

## The rule
In wouter v3 (via `regexparam`), a route like `<Route path="/:rest*" component={X} />` compiles to a regex requiring at least one path segment. It matches `/foo` but NOT the bare root `/`. If a `Switch` only has `/login` and `/:rest*` as its two routes, nothing matches `/` and the `Switch` renders nothing — a totally blank page, with no console errors and no lifecycle logs from the intended component (it never mounts).

**Why:** This pattern is commonly used to wrap all non-login routes in a layout/auth-gate component (e.g. `<Route path="/:rest*" component={AuthedRoutes} />`). It's easy to assume it covers "everything including root," but it silently excludes `/`.

**How to apply:** Whenever a wouter `Switch` uses a `:param*` wildcard as a catch-all for an authenticated route group, always add an explicit sibling route for `/` pointing at the same component, placed before the wildcard:
```tsx
<Switch>
  <Route path="/login" component={LoginPage} />
  <Route path="/" component={AuthedRoutes} />
  <Route path="/:rest*" component={AuthedRoutes} />
</Switch>
```

## Diagnosis tip
If a React app shows a pure white screen with no console errors at all (not even the app's own console.log calls placed inside the suspected component), suspect the component never mounted in the first place — check route matching before debugging auth/data-fetching logic inside the component.
