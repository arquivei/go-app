package logger

import (
	"context"
	"log/slog"

	"github.com/rs/zerolog"
)

var _ slog.Handler = (*slogHandler)(nil)

type slogHandler struct {
	logger *zerolog.Logger
	group  string
}

func (h *slogHandler) Enabled(ctx context.Context, l slog.Level) bool {
	return h.logger.WithLevel(zLevel(l)).Enabled()
}

func (h *slogHandler) Handle(ctx context.Context, r slog.Record) error {
	c := h.logger.WithLevel(zLevel(r.Level))

	if r.NumAttrs() > 0 {
		r.Attrs(func(attr slog.Attr) bool {
			c = c.Interface(h.group+attr.Key, attr.Value.Any())
			return true
		})
	}

	c.Msg(r.Message)

	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	c := h.logger.With()
	for _, attr := range attrs {
		c = c.Interface(h.group+attr.Key, attr.Value.Any())
	}

	logger := c.Logger()

	return &slogHandler{logger: &logger}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}

	return &slogHandler{
		logger: h.logger,
		group:  name + "." + h.group,
	}
}

func zLevel(l slog.Level) zerolog.Level {
	switch l {
	case slog.LevelDebug:
		return zerolog.DebugLevel
	case slog.LevelInfo:
		return zerolog.InfoLevel
	case slog.LevelWarn:
		return zerolog.WarnLevel
	case slog.LevelError:
		return zerolog.ErrorLevel
	default:
		return zerolog.WarnLevel
	}
}
