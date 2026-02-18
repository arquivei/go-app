# go-app

[![PkgGoDev](https://pkg.go.dev/badge/github.com/arquivei/go-app)](https://pkg.go.dev/github.com/arquivei/go-app)
[![Go Report Card](https://goreportcard.com/badge/github.com/arquivei/go-app)](https://goreportcard.com/report/github.com/arquivei/go-app)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)

A very opinionated app with life cycle and graceful shutdown.

## Quickstart

Here is an example of a program that does nothing but it compiles and runs with all app features:

```go
package main

import (
	"context"

	"github.com/arquivei/go-app"
)

var (
	version = "v0.0.0-dev"
	cfg     struct {
		app.Config
	}
)

func main() {
	defer app.Recover()
	app.Bootstrap(version, &cfg)
	app.RunAndWait(func(ctx context.Context) error {
		<-ctx.Done()
		return ctx.Err()
	})
}
```

Check all available options with `go run ./ -h`

```text
Supported Fields:
FIELD                             FLAG                               ENV                               DEFAULT           USAGE
-----                             -----                              -----                             -------           -----
App.Log.Level                     -app-log-level                     APP_LOG_LEVEL                     info              The log level. Possible values are: trace, debug, info, warn, error, fatal and panic.
App.Log.Human                     -app-log-human                     APP_LOG_HUMAN                     false             Whether to use human friendly log output. If true, the log will be printed in a human friendly format. If false, the log will be printed in JSON format.
App.AdminServer.Enabled           -app-adminserver-enabled           APP_ADMINSERVER_ENABLED           true              Enables the admin server
App.AdminServer.Addr              -app-adminserver-addr              APP_ADMINSERVER_ADDR              localhost:9000    The address the admin server will bind to. To bind to all interfaces, use :9000.
App.AdminServer.With.DebugURLs    -app-adminserver-with-debugurls    APP_ADMINSERVER_WITH_DEBUGURLS    true              Enables the /debug URLs in the admin server.
App.AdminServer.With.Metrics      -app-adminserver-with-metrics      APP_ADMINSERVER_WITH_METRICS      true              Enables the /metrics endpoint in the admin server.
App.AdminServer.With.Probes       -app-adminserver-with-probes       APP_ADMINSERVER_WITH_PROBES       true              Enables the /ready and /healthy endpoints in the admin server.
App.Shutdown.GracePeriod          -app-shutdown-graceperiod          APP_SHUTDOWN_GRACEPERIOD          3s                The grace period for the shutdown procedure. During normal shutdown procedures, the shutdown function will wait this amount of time before actually starting calling the shutdown handlers.
App.Shutdown.Timeout              -app-shutdown-timeout              APP_SHUTDOWN_TIMEOUT              5s                The timeout for the shutdown procedure. If the shutdown procedure takes longer than this value, the application will force exit.
App.Config.Output                 -app-config-output                 APP_CONFIG_OUTPUT                                   Prints the configuration in the desired format and exit. Possible values are: env, yaml and json.
App.Check.Ready                   -app-check-ready                   APP_CHECK_READY                   false             Whether to execute a ready check on the application. If true, the application will execute the ready check and exit with code 0 if the check is successful or with code 1 if the check fails.
App.Check.Healthy                 -app-check-healthy                 APP_CHECK_HEALTHY                 false             Whether to execute a healthy check on the application. If true, the application will execute the healthy check and exit with code 0 if the check is successful or with code 1 if the check fails.
```

There is a special option to print out the default configuration in `env` or `yaml` format: `go run . -app-config-output=env`.

``` text
APP_LOG_LEVEL=info
APP_LOG_HUMAN=
APP_ADMINSERVER_ENABLED=true
APP_ADMINSERVER_ADDR=localhost:9000
APP_ADMINSERVER_WITH_DEBUGURLS=true
APP_ADMINSERVER_WITH_METRICS=true
APP_ADMINSERVER_WITH_PROBES=true
APP_SHUTDOWN_GRACEPERIOD=3s
APP_SHUTDOWN_TIMEOUT=5s
APP_CONFIG_OUTPUT=
```

Version can be overwritten in compile time using `-ldflags`:

```sh
-ldflags="-X main.version=v0.0.1"
```

More information on the [presentation](docs/presentation.slide) slides.

Use the [present](https://pkg.go.dev/golang.org/x/tools/present) tool to render the slides or you can check it online at https://go-talks.appspot.com/github.com/arquivei/go-app/docs/presentation.slide

## Minimal dependencies

- `prometheus/client_golang`: Metrics
- `rs/zerolog`: Structured log in JSON format
- `stretchr/testify`: Better unit testing asserts

## Getting Started

These instructions will give you a copy of the project up and running on
your local machine for development and testing purposes. See deployment
for notes on deploying the project on a live system.

### Prerequisites

Requirements for the software and other tools to build, test and push:

- [go](https://go.dev/)
- [golangci-lint](https://golangci-lint.run/): for linting.

### Linting

Please run `golangci-lint run` before submitting code.

### Godoc

To read the godoc documentation run:

```sh
godoc -http=localhost:6060
```

and open `http://localhost:6060` on your browser.

## Contributing

Please read [CONTRIBUTING.md](CONTRIBUTING.md) for details on our code
of conduct, and the process for submitting pull requests to us.

## Versioning

We use [Semantic Versioning](http://semver.org/) for versioning. For the versions
available, see the [tags on this
repository](https://github.com/arquivei/go-app/tags).

## License

This project is licensed under the _BSD 3-Clause_ - see the [LICENSE.txt](LICENSE.txt) file for
details.

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=arquivei/go-app&type=Date)](https://www.star-history.com/#arquivei/go-app&Date)
