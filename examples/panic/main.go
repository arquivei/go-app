// This example shows how panics are captured and gracefully printed.
//
// To run this example
// go run -ldflags="-X main.version=v0.0.1" ./examples/panic/ -app-log-human -app-log-level=trace
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

	// Comment this next line to see the other panic
	thisWillPanic()

	app.RunAndWait(func(_ context.Context) error {
		panic("panics inside run and wait will trigger a graceful shutdown")
	})
}

func thisWillPanic() {
	panic("panics outside RunAndWait should be caught by  app.Recover()")
}
