// This is an example program that does nothing but it executes the full
// app bootstrap and waits for app to be gracefully shutdown.
//
// To run this program in the commandline you could use:
//   go run ./examples/quickstart/ -app-log-human -app-log-level=trace

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
