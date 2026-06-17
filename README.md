# grr-gin

[![CI](https://github.com/arpaad/grr-gin/actions/workflows/ci.yml/badge.svg)](https://github.com/arpaad/grr-gin/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/arpaad/grr-gin.svg)](https://pkg.go.dev/github.com/arpaad/grr-gin)

Gin middleware for [grr](https://github.com/arpaad/grr): injects a registry into each request's context and manages a per-request scope. One function, 20 lines.

```go
import (
    "github.com/arpaad/grr"
    grrgin "github.com/arpaad/grr-gin"
)

r := grr.New()
// ... register your dependencies on r ...

router := gin.Default()
router.Use(grrgin.Middleware(r))

router.GET("/users/:id", func(c *gin.Context) {
    ctx := c.Request.Context() // carries the registry + open scope
    resp, err := logic.UserInfo.Do(ctx, req)
    // ...
})
```

Per request, the middleware:

1. Binds `r` to the request context via `grr.WithRegistry`.
2. Opens a scope (`r.BeginScope`) so any scoped dependency is built at most once per request.
3. Defers `endScope()` — when the handler chain returns, the scope is synchronously closed.

## Why a separate repo

`grr` and `gold` have no framework dependencies. Adding Gin support inside `grr` would pull Gin's full dependency tree into every project that imports `grr`, even ones using a different HTTP framework. A separate module keeps the core clean — same reason Go's `database/sql` drivers live in separate import paths.

## Install

```sh
go get github.com/arpaad/grr-gin
```

## Usage with gold

`gold.Logic.Do` reads the registry from `ctx` via `grr.RegistryFromCtx` — there's no Gin-specific code needed in your business logic layer. The middleware just ensures `ctx` has the right registry and an open scope before your handler runs.

```go
// di/userinfo.go — wired once at startup
func init() {
    logic.UserInfo.RegisterScopedIn(appRegistry, func(ctx context.Context) userinfo.Model {
        return userinfo.NewModel(db.FromCtx(ctx))
    })
}

// handler — knows nothing about grr or Gin internals
func GetUser(c *gin.Context) {
    resp, err := logic.UserInfo.Do(c.Request.Context(), userinfo.Request{UserID: id})
    // ...
}
```

## Testing

Test handlers without a real server — `httptest` works directly:

```go
func TestGetUser(t *testing.T) {
    gin.SetMode(gin.TestMode)

    r := grr.New()
    logic.UserInfo.RegisterScopedIn(r, func(ctx context.Context) userinfo.Model {
        return &mockModel{name: "Alice"}
    })

    router := gin.New()
    router.Use(grrgin.Middleware(r))
    router.GET("/users/:id", GetUser)

    req := httptest.NewRequest(http.MethodGet, "/users/1", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)
}
```

Runnable examples live in [example_test.go](example_test.go).

## Status

v0.1 — `Middleware`. See [plan.md](plan.md) for what's coming next (Echo/Fiber/gRPC adapters, benchmarks) and [ARCHITECTURE.md](ARCHITECTURE.md) for the design rationale.

## License

MIT — see [LICENSE](LICENSE).
