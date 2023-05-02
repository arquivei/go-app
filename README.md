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
FIELD                             FLAG                               ENV                               DEFAULT
-----                             -----                              -----                             -------
App.Log.Level                     -app-log-level                     APP_LOG_LEVEL                     info
App.Log.Human                     -app-log-human                     APP_LOG_HUMAN
App.AdminServer.Enabled           -app-adminserver-enabled           APP_ADMINSERVER_ENABLED           true
App.AdminServer.Port              -app-adminserver-port              APP_ADMINSERVER_PORT              9000
App.AdminServer.With.DebugURLs    -app-adminserver-with-debugurls    APP_ADMINSERVER_WITH_DEBUGURLS    true
App.AdminServer.With.Metrics      -app-adminserver-with-metrics      APP_ADMINSERVER_WITH_METRICS      true
App.AdminServer.With.Probes       -app-adminserver-with-probes       APP_ADMINSERVER_WITH_PROBES       true
App.Shutdown.GracePeriod          -app-shutdown-graceperiod          APP_SHUTDOWN_GRACEPERIOD          3s
App.Shutdown.Timeout              -app-shutdown-timeout              APP_SHUTDOWN_TIMEOUT              5s
```

Version can be overwritten in compile time using `-ldflags`:

```sh
-ldflags="-X main.version=v0.0.1"
```

More information on the [presentation](docs/presentation.slide) slides.

Use the [present](https://pkg.go.dev/golang.org/x/tools/present) tool to render the slides or you can check it online at https://go-talks.appspot.com/github.com/arquivei/go-app/docs/presentation.slide

## Minimal dependencies

- `omeid/uconfig`: Bind flag and environment to struct
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
