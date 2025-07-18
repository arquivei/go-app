// This is an example program that shows how probes work.
// It will fetch and log it's own probe and it will swap
// the readiness probe state every 5 seconds.
//
// To run this program in the commandline you could use:
//   go run ./examples/probes/main.go -app-log-human --app-log-level=trace
//
// To check the probes statuses from outside the process you could use:
//   go run ./examples/probes/main.go -app-log-human -app-check-ready
//     or
//   go run ./examples/probes/main.go -app-log-human -app-check-healthy

package main

import (
	"context"
	"io"
	"net/http"
	"time"

	"github.com/arquivei/go-app"
	"github.com/rs/zerolog/log"
)

var (
	version = "v0.0.0-dev"
	cfg     struct {
		app.Config
	}
)

func main() {
	app.Bootstrap(version, &cfg)

	exampleProbe, err := app.ReadinessProbeGoup().NewProbe("example_probe", true) // HL
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create probe")
	}

	// Invert probe status every 5 seconds
	go func() {
		for {
			time.Sleep(5 * time.Second)
			exampleProbe.Set(!exampleProbe.IsOk()) // HL
		}
	}()

	app.RunAndWait(func(ctx context.Context) error {
		readinessURL := "http://" + cfg.App.AdminServer.Addr + "/ready"
		for ctx.Err() == nil {
			time.Sleep(time.Second)

			// This is just an example, we are not using TLS here.
			//nolint: gosec, noctx
			resp, err := http.Get(readinessURL)
			if err != nil {
				log.Error().Err(err).Msg("Failed to fetch probe")
				continue
			}
			b, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Error().Err(err).Msg("Failed to read probe")
				continue
			}
			// nolint: errcheck, gosec
			resp.Body.Close()

			log.Info().Int("http_status", resp.StatusCode).Msgf("Probe says: %s", b)
		}
		return ctx.Err()
	})
}
