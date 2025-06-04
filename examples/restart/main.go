// Example of an app using the restart policy.
//
// Example of how you could run this demo:
// go run ./examples/restart/ -app-log-human -app-log-level=info -app-restart-policy=on_error -app-restart-max=2 -errorcounter=3
// This example will run the main loop, which will sleep for 1 second and then return an error.
// It will restart on errors. The main loop is programmed to return an error 3 times.
// The app is configured to retry only 2 times, so it will stop after the second error.
// You can play with the flags to see how the behavior changes.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/arquivei/go-app"
)

var (
	version = "v0.0.0"
	cfg     struct {
		app.Config
		Timer        time.Duration `default:"1s"`
		ErrorCounter int           `default:"0"`
	}
)

func main() {
	app.Bootstrap(version, &cfg)
	defer app.Recover()

	errCounter := cfg.ErrorCounter

	app.RunAndWait(func(ctx context.Context) error {
		time.Sleep(cfg.Timer)

		if errCounter > 0 {
			errCounter--
			return fmt.Errorf("simulated error")
		}
		return nil
	})
}
