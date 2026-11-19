# Project context

`go-app` is a Go library for configuration, structured logging, Prometheus
metrics, readiness and liveness probes, and graceful shutdown. The root
package is `app` (module `github.com/arquivei/go-app`).

## Commands

Run these commands from the repository root with Go 1.26.0 or later.

```sh
task checks
go vet ./...
go build ./...
go test ./...
go test -v -race ./...
golangci-lint run

go run ./examples/quickstart/ -app-log-human
go run ./examples/quickstart/ -h
go run ./examples/quickstart/ -app-config-output=env
go run ./examples/quickstart/ -app-config-output=yaml
task docs:generate
```

`task checks` runs vet, build, tests, and lint. The build, vet, and race-test
CI jobs in [.github/workflows/go.yml](.github/workflows/go.yml) use Go 1.27.
The lint job uses stable Go and the latest golangci-lint.

Help currently prints usage and exits with status 1. Configuration export
supports env and YAML. The advertised JSON format currently fails.

## Architecture

- `app.go`, `default.go`, and `config.go` implement the application lifecycle
  and configuration. `app.Bootstrap` sets up the global logger and default app.
  `app.New` creates an explicit instance using the supplied values directly.
- `internal/thirdparty/uconfig/` is the vendored configuration parser.
  `Bootstrap` and `SetupConfig` load struct tag defaults, environment variables,
  and flags, in that order. Applications that need configuration files load
  them separately.
- The admin server uses `net/http` and defaults to `localhost:9000`. When
  enabled, it always registers `/metrics`, `/ready`, and `/healthy`.
  `App.AdminServer.With.Metrics` and `App.AdminServer.With.Probes` currently have
  no effect on route registration. Debug routes are configurable.
- `probe.go` implements probe groups. `shutdownhandler.go` implements handler
  priority and error policy. `httphandler_dump.go` implements debug dumps.
- `pkg/memguard/` monitors Go runtime memory against the configured
  `GOMEMLIMIT`. The root package integrates overload with probes and exposes
  it via `IsOverloaded` for application middleware.
- `logger/` implements zerolog setup, slog integration, and configuration
  value redaction. Prometheus metrics use `prometheus/client_golang`.

## Conventions

- Write documentation, code comments, configuration help, logs, and test
  messages in English.
- Keep dependencies minimal. Check the [runnable examples](examples/) before
  adding public APIs. Add or update tests when changing behavior, and use
  `stretchr/testify` for assertions in unit tests.
- Run `gofmt` and `golangci-lint run` when changing Go files.
- Keep `logger.Flatten` and the [secrets guide](docs/secrets.md) in sync when
  changing how configuration values are masked.
- Update package documentation in `doc.go` and executable Godoc examples in
  `example_test.go` when public behavior changes. Run example tests with
  `go test ./...` and check slide source references when moving code.
- Regenerate the [configuration reference](docs/configuration-reference.md)
  with `task docs:generate` when configuration fields, tags, or output change.

## Guides

- [Configuration](docs/configuration.md)
- [Operations and shutdown](docs/operations.md)
- [Memory protection](docs/memory-protection.md)
- [Project presentation](docs/presentation.slide)
- [Contribution requirements](CONTRIBUTING.md)
