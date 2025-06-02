// This is an example program on how to use the app package. It uses to FileServer
// handler to serve all files on a given directory.
//
// To run this program in the commandline you could use:
//
//	go run ./examples/servefiles/ -app-log-human -app-log-level=trace
package main

import (
	"context"
	"net/http"
	"time"

	"github.com/arquivei/go-app"
	"github.com/rs/zerolog/log"
)

var (
	version = "v0.0.0-dev"
	cfg     struct {
		// App is the app specific configuration
		app.Config

		// Programs can have any configuration the want.

		HTTP struct {
			Addr string `default:"localhost:8000"`
		}
		Dir string `default:"."`
	}
)

func main() {
	// Bootstrap will:
	// - Set cfg values
	// - Start admin server
	// - Start Healthiness prove as Healthy
	// - Start Readiness probe as Unready
	app.Bootstrap(version, &cfg)

	// Initialize app dependencies, this could take a while.
	// It's recommended to initialize everything before calling RunAndWait because
	// readiness probe is already up and reporting the app as not ready yet.
	// App will become ready when RunAndWait is called.
	httpServer := newHTTPServer()

	// Run the main loop until it finishes or receives termination signal
	// On this point the readiness probe starts returning success.
	app.RunAndWait(func(_ context.Context) error {
		log.Info().
			Str("dir", cfg.Dir).
			Str("port", cfg.HTTP.Addr).
			Msg("Serving directory.")
		return httpServer.ListenAndServe()
	})
}

func newHTTPServer() *http.Server {
	httpServer := &http.Server{
		Addr:              cfg.HTTP.Addr,
		Handler:           http.FileServer(http.Dir(cfg.Dir)),
		ReadHeaderTimeout: 10 * time.Second,
	}

	// You can register the shutdown handlers at any order, but do it before starting the app
	app.RegisterShutdownHandler(
		&app.ShutdownHandler{
			Name:     "http_server",
			Priority: app.ShutdownPriority(100),
			Handler:  httpServer.Shutdown,
			Policy:   app.ErrorPolicyAbort,
		},
	)
	return httpServer
}
