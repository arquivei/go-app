package main

import (
	"context"
	stdlog "log"
	"log/slog"

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
	defer app.Recover()

	app.Bootstrap(version, &cfg)

	app.RunAndWait(func(ctx context.Context) error {
		stdlog.Println("Standard go logger. Nothing much to it.")

		log.Info().Str("logger", "zerolog").Msg("Zerolog is our main logger")

		slog.Info("Slog is also configured by go-app!", "logger", "slog")

		slogLogger := slog.Default().WithGroup("group")
		slogLogger = slogLogger.WithGroup("slog")
		slogLogger.With("foo", "bar")
		slogLogger.Warn("Everything on slog just works!", slog.Bool("cool", true))

		<-ctx.Done()

		return ctx.Err()
	})
}
