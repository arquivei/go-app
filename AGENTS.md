# Project Context

## What this service is

`go-app` is a Go library (not a standalone executable service) that provides a highly opinionated application framework with a ready-to-use lifecycle and *graceful shutdown*. It solves the problem of repeating the same basic infrastructure in every new Go service: configuration bootstrap, structured logging, metrics, health/readiness probes, and admin endpoints.

Main capabilities:
- **Configuration management** extensible via `uconfig` (vendored in `internal/thirdparty/uconfig`), with support for flags, environment variables, and files.
- **Structured logging** with `rs/zerolog` (default) and `log/slog` support.
- **Metrics** integrated with `prometheus/client_golang`.
- **Admin server**: exposes `/ready`, `/healthy` (probes), and `/debug` (debugging URLs) on port 9000 by default.
- **Shutdown handlers** with configurable priority and error policies.
- **Memory protection** (`pkg/memguard`): monitors process memory usage and signals overload once a configurable percentage of `GOMEMLIMIT` is exceeded. Automatically bootstrapped by `app.New`/`app.Bootstrap` with production defaults (see `Config.App.MemGuard`); `app.IsOverloaded()`/`App.IsOverloaded()` is always available (e.g. for middleware that returns HTTP 429) regardless of whether the readiness/liveness probes are enabled.

## Commands

```bash
# Run all checks (vet, build, test, lint) — uses Taskfile.yml
task checks

# Individually
go vet ./...
go build ./...
go test ./...
go test -race -v ./...   # as run in CI
golangci-lint run

# See all configuration options for the example app
go run ./ -h

# Print the default configuration in a specific format
go run . -app-config-output=env   # or yaml, json

# Documentation (godoc)
godoc -http=localhost:6060
```

CI (`.github/workflows/go.yml`) runs in parallel: `build`, `test` (with `-race`), `vet`, and `golangci-lint` on every push/PR to `main`/`master`, using Go 1.26.

## Architecture

```
.
├── app.go, default.go, config.go     # core: lifecycle, global default app, root config
├── probe.go                          # Probe / ProbeGroup (readiness/liveness)
├── shutdownhandler.go                # registration and execution of shutdown handlers with priority
├── httphandler_dump.go               # debug handlers (/debug)
├── recover.go, errors.go, doc.go     # utilities and documentation for the `app` package
├── logger/                           # log config, field flattening, level hook, slog handler
├── pkg/memguard/                     # memory protection (Guard) — package with no dependency on `app`, used by app.go
├── internal/thirdparty/uconfig/      # internal fork of the config lib (flags, env, file, defaults)
└── examples/                         # runnable examples: quickstart, probes, panic, servefiles,
                                       # shutdown-handlers, slog
```

- The root package is `app` (module `github.com/arquivei/go-app`). The recommended usage pattern is the "default app": public functions in the `app` package operate on a global instance (`app.Bootstrap`, `app.RunAndWait`, `app.RegisterShutdownHandler`, etc.).
- `internal/` contains code that isn't publicly exported: a fork/vendoring of the `uconfig` lib used internally for configuration.
- `pkg/memguard` is a leaf package (it does not import `app`, to avoid an import cycle) that exposes only `Guard`/`Config`/`Overloader`. The `app` package itself does the `ProbeGroup` integration (in `app.go`, `startMemGuard` function, called by `New`), so no consumer needs to do any manual wiring.
- Versioning is injected at compile time via `-ldflags="-X main.version=..."`.

## Conventions

- **Go version**: 1.25.0+ (set in `go.mod`); CI uses Go 1.26.
- **Linting**: code must pass `golangci-lint run` (config in `.golangci.yml`) before submitting. Enabled linters include `bodyclose`, `dupl`, `errcheck`, `copyloopvar`, `goconst`, `gocritic`, `gocyclo`, `gosec`, `govet`, `ineffassign`, `misspell`, `nakedret`, `noctx`, `prealloc`, `rowserrcheck`, `staticcheck`, `unconvert`, `unparam`, `unused`, `whitespace`. Don't use hacks to bypass linters unless explicitly necessary.
- **Tests**: use `stretchr/testify` for assertions. Always add or update tests when changing behavior. Test files follow the `_test.go` pattern next to the code under test.
- **Logging**: prefer structured logging. The application standardizes logs in JSON or human-readable format via configuration (`APP_LOG_HUMAN`).
- **Configuration**: config fields use `default` and `usage` tags (e.g. `pkg/memguard/memguard.go`), which automatically feed flags, env vars, and help (`-h`).
- **Examples first**: before implementing new features, check `examples/` (`panic`, `probes`, `quickstart`, `servefiles`, `shutdown-handlers`, `slog`) to align with existing patterns.
- **Minimal dependencies**: minimize external dependencies; prefer libraries already present in `go.mod` (`zerolog`, `prometheus/client_golang`, `echo/v4`, `testify`).
- The whole codebase, including `pkg/memguard`, is documented in English — the project is public, so comments, package/type docs, config `usage` strings, and log/test messages must be in English.

## Relevant dependencies

- `github.com/labstack/echo/v4`: used for the admin server's HTTP handlers.
- `github.com/prometheus/client_golang`: metrics.
- `github.com/rs/zerolog`: structured logging (JSON by default).
- `github.com/stretchr/testify`: test assertions.
- `golang.org/x/text`: text/i18n support.
- `internal/thirdparty/uconfig`: internal fork (not an external dependency) responsible for resolving configuration from flags, env vars, and files.

## Extra information

- Detailed library presentation in slides: `docs/presentation.slide` (renderable with the [`present`](https://pkg.go.dev/golang.org/x/tools/present) tool, or online at https://go-talks.appspot.com/github.com/arquivei/go-app/docs/presentation.slide).
- Package documentation (godoc) embedded in `doc.go` explains the application lifecycle (`Bootstrap` → shutdown handler registration → `RunAndWait`) with examples.
- `CONTRIBUTING.md` describes the PR process and acceptance criteria (issues, clear intent, quality).
- License: BSD 3-Clause (`LICENSE.txt`).
