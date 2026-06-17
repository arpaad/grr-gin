# `grr-gin` – Architecture

> Module path: `github.com/arpaad/grr-gin` | Status: v0.1 implemented  
> This file is the design document and decision log for v0.1. Future work lives in [plan.md](plan.md).

---

## Why a separate repo

`grr` and `gold` are deliberately framework-free — their only dependencies are the Go standard library and each other. Adding Gin support directly to `grr` would pull `github.com/gin-gonic/gin` (and its transitive dependency tree) into every project that uses `grr`, even projects that use a different HTTP framework or no HTTP layer at all.

The pattern follows what Go's own `database/sql` ecosystem does: the driver lives in a separate import (`github.com/lib/pq`, not `database/sql`), and `database/sql` stays clean.

## What the middleware does

```
incoming request
     │
     ▼
grrgin.Middleware(r)
     │
     ├── grr.WithRegistry(ctx, r)   — binds r to the request context
     ├── r.BeginScope(ctx)          — opens a per-request scope
     │                                (globally unique scope ID via atomic counter in grr)
     │
     ▼
  c.Next()                          — handler runs; any r.Resolve(ctx, key)
     │                                for a scoped key returns the same instance
     │                                for the duration of this request
     ▼
defer endScope()                   — synchronously releases the scope, panics
                                     any stale ctx.Resolve after the request
```

The middleware calls `c.Request = c.Request.WithContext(ctx)` so that downstream handlers receive the enriched context via `c.Request.Context()`. Gin's `c.Request.Context()` is the canonical way to access `context.Context` in Gin handlers; `c.Request.WithContext` is the standard Go pattern for propagating a derived context.

## Scope lifecycle guarantee

`endScope` is called synchronously in a `defer`, not in a goroutine. This means:
- By the time `ServeHTTP` returns (i.e., after all middleware and handlers finish), the scope is closed.
- Any attempt to resolve a scoped key on a saved/leaked context after the request panics with a clear message from `grr`.
- There is no window where a background goroutine outlives its scope silently.

## What this repo is NOT

- It does not wrap or re-export any Gin APIs.
- It does not add routing, error handling, or any other framework features.
- It does not support `gold.Logic` directly — `gold.Logic.Do` works via `grr.RegistryFromCtx`, which is framework-agnostic; the middleware just makes sure `ctx` carries a registry and a scope, which is all `gold` needs.
