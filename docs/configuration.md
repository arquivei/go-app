# Configuration

`app.Bootstrap(version, &cfg)` loads configuration, sets up the global logger,
and creates the default app. Use `app.Config` directly or embed it in a custom
configuration struct. Its promoted `GetAppConfig` method allows the custom
struct to satisfy `app.AppConfig`.

```go
type Config struct {
	app.Config

	Service struct {
		Addr string `default:":8080" usage:"Address for the service HTTP listener."`
	}
}

var cfg Config
app.Bootstrap(version, &cfg)
```

`app.SetupConfig` applies struct tag defaults first, then environment
variables, then explicitly supplied command-line flags. In the example,
`SERVICE_ADDR` overrides the default and `-service-addr=:8081` overrides the
environment variable. The default names come from the dotted field path
(`Service.Addr`), using uppercase and underscores for environment variables
and lowercase and hyphens for flags. The parser reads defaults, environment
variables, and flags only. Load files separately if your application uses them.

For an explicit `*app.App`, use this alternative to `Bootstrap` with the same
custom `Config` type.

```go
func newApp() *app.App {
	var cfg Config
	app.SetupConfig(&cfg)
	return app.New(cfg.Config)
}
```

`app.SetupConfig` applies the tag defaults and overrides. `app.New` uses the
`app.Config` values you pass directly and leaves global logger setup to the
caller. `app.New(app.Config{})` uses Go zero values, so the admin server and
MemGuard are disabled.

See the [secrets guide](secrets.md) for the rules that mask configuration
values in the bootstrap log.

## Inspect options and defaults

Run these commands from the repository root.

```sh
go run ./examples/quickstart/ -h
go run ./examples/quickstart/ -app-config-output=env
go run ./examples/quickstart/ -app-config-output=yaml
```

Help lists fields, flags, environment variables, and usage descriptions. It
currently prints usage and exits with status 1. The env and YAML commands
print tag defaults to stderr, exit with status 0, and skip application startup.
The YAML uses environment variable names as flat keys and string values.
Help advertises JSON, but `go run ./examples/quickstart/ -app-config-output=json`
currently fails with an invalid-format error and exits with status 1.

Regenerate the [configuration reference](configuration-reference.md) from the
quickstart with Go and Task installed.

```sh
task docs:generate
```

## Set the application version at build time

The quickstart passes its `main.version` string to `app.Bootstrap`, which
includes it in structured logs. Build the quickstart with a different version
from the repository root.

```sh
build_dir=$(mktemp -d)
go build -ldflags="-X main.version=v0.0.1" -o "$build_dir/quickstart" ./examples/quickstart/
```
