# AGENTS.md

Context for AI agents (and future-you) picking up this repo cold.

## What this repo is

`grr-gin` is a single-function Gin middleware adapter for the `grr`/`gold` three-repo family:

- **[`grr`](https://github.com/arpaad/grr)** — Go Registry Resolver. Untyped factory/value registry with scope lifecycle. Read its AGENTS.md if you're touching anything scope-related.
- **[`gold`](https://github.com/arpaad/gold)** — Go Logic Dependency. Typed domain logic layer on top of `grr`. This repo does not import `gold`, but the middleware enables `gold.Logic.Do` to work in a Gin context.
- **`grr-gin`** (this repo) — `Middleware(r *grr.Registry) gin.HandlerFunc`. That's the entire public API.

## Where the reasoning lives

- **[ARCHITECTURE.md](ARCHITECTURE.md)** — explains why the repo exists as a separate module (dependency isolation), what the middleware does step-by-step, and the scope lifecycle guarantee. Read this before adding any state, error handling, or additional parameters to `Middleware`.
- **[plan.md](plan.md)** — phase-2 roadmap (Echo/Fiber/gRPC adapters, benchmarks). Nothing here is implemented.
- **[README.md](README.md)** — public-facing docs including the Gin+gold wiring pattern and testing example.

## Hard constraints — don't violate these without a conversation first

- **This repo must not import `gold`.** The middleware works with `gold` indirectly (via `grr.WithRegistry` + `grr.BeginScope`), but importing `gold` here would create a circular dependency and defeat the separation of concerns.
- **`Middleware` must stay a single function with a single responsibility.** It opens a scope and attaches a registry. Don't add routing, error handling, logging, or recovery — those belong in Gin's own middleware ecosystem or in the application layer.
- **`endScope` must be called synchronously via `defer`, not in a goroutine.** The synchronous cleanup guarantee is what makes `TestMiddlewareScopeEndsAfterRequest` pass and what makes stale-context access reliably panic. If you ever feel tempted to move this to a goroutine, read `grr/ARCHITECTURE.md` first — the distinction between sync and async endScope was a bug that had to be fixed.
- **No `reflect` package.** `grr-gin` itself uses none; don't add any.

## Running things

```sh
make test     # go test ./... -race
make lint     # golangci-lint run
make vet      # go vet ./...
make tidy     # go mod tidy, fails if it would change go.mod/go.sum
make ci       # everything CI runs, locally
```

Note: `go.mod` has a `replace github.com/arpaad/grr => ../grr` directive for local development. The CI workflow checks out `grr` as a sibling directory so this resolves correctly on GitHub Actions.

## Conventions

- This is a single-file package. Keep it that way unless there's a very strong reason to split.
- The test file (`middleware_test.go`) has two tests: one proving the scope reuse works within a request, one proving the scope ends after the request. Any new behavior in `Middleware` needs a test proving its contract.
- Runnable examples (`Example*` with `// Output:`) belong in `example_test.go`.
