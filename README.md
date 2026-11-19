# go-app

[![PkgGoDev](https://pkg.go.dev/badge/github.com/arquivei/go-app)](https://pkg.go.dev/github.com/arquivei/go-app)
[![Go Report Card](https://goreportcard.com/badge/github.com/arquivei/go-app)](https://goreportcard.com/report/github.com/arquivei/go-app)
[![License](https://img.shields.io/badge/License-BSD%203--Clause-blue.svg)](https://opensource.org/licenses/BSD-3-Clause)

go-app is a Go library for service lifecycle and graceful shutdown. It provides
configuration, structured logging with zerolog and `log/slog`, Prometheus
metrics, readiness and liveness probes, and memory-overload detection.

## Quickstart

go-app requires Go 1.26.0 or later. In your own Go module, install the library
and import `github.com/arquivei/go-app` as `app`.

```sh
go get github.com/arquivei/go-app
```

Create a `main.go` in your module.

```go
package main

import (
	"context"

	"github.com/arquivei/go-app"
)

func main() {
	var cfg app.Config
	app.Bootstrap("v0.0.0-dev", &cfg)
	app.RunAndWait(func(ctx context.Context) {
		<-ctx.Done()
	})
}
```

To run the [repository quickstart](examples/quickstart/main.go), use this
command from the repository root. Press Ctrl+C to begin graceful shutdown.

```sh
go run ./examples/quickstart/ -app-log-human
```

## Guides and examples

- [Configuration](docs/configuration.md) and [generated defaults](docs/configuration-reference.md)
- [Operations and shutdown](docs/operations.md)
- [Memory protection](docs/memory-protection.md)
- [Secrets in configuration logs](docs/secrets.md)
- [Runnable examples](examples/) for probes, shutdown handlers, slog, panic recovery, and serving files
- [Project presentation](docs/presentation.slide)

## Contributing

See the [contribution guide](CONTRIBUTING.md) for the pull request process.
CI uses Go 1.27 for build, test, and vet, and the stable Go toolchain for lint.

## Versioning

The project follows [Semantic Versioning](https://semver.org/). See the
[repository tags](https://github.com/arquivei/go-app/tags) for available versions.

## License

go-app is licensed under the BSD 3-Clause license. See [LICENSE.txt](LICENSE.txt).

## Star History

[![Star History Chart](https://api.star-history.com/svg?repos=arquivei/go-app&type=Date)](https://www.star-history.com/#arquivei/go-app&Date)
