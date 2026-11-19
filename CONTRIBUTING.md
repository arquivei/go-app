# Contributing

Use issues to report problems, suggest improvements, or ask questions about
the library. In pull requests, explain the problem, the public behavior
affected, and how you verified the change. Keep changes focused on common
service and worker needs, including configuration, logging, metrics, and
lifecycle management.

## Run checks

The module requires Go 1.26.0 or later. From the repository root, run these
checks with Go and golangci-lint installed.

```sh
go vet ./...
go build ./...
go test ./...
go test -v -race ./...
golangci-lint run
```

With Task installed, `task checks` runs vet, build, tests, and lint. Run the
race tests separately. CI uses Go 1.27 for build, vet, and race tests.

## Update documentation and examples

Write documentation, code comments, configuration help, logs, and test
messages in English. Update the [guides](README.md#guides-and-examples),
package Godoc, and [runnable examples](examples/) when relevant behavior
changes. Keep examples in `example_test.go` deterministic so `go test ./...`
can verify their output. If you change code featured in the
[presentation](docs/presentation.slide), check that its code references still
show the intended snippets.

When configuration fields, tags, or help output change, regenerate the
[configuration reference](docs/configuration-reference.md) from the repository
root. Include the generated diff in the pull request.

```sh
task docs:generate
```

Follow the [code of conduct](CODE_OF_CONDUCT.md) in project discussions and
contributions.
