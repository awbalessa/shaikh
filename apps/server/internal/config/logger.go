package config

import (
	"io"
	"log/slog"
)

type LoggerOptions struct {
	Level  slog.Level
	JSON   bool
	Writer io.Writer
}

func NewLogger(opts LoggerOptions) *slog.Logger {
	var handler slog.Handler
	if opts.JSON {
		handler = slog.NewJSONHandler(opts.Writer, &slog.HandlerOptions{
			Level: opts.Level,
		})
	} else {
		handler = slog.NewTextHandler(opts.Writer, &slog.HandlerOptions{
			Level: opts.Level,
		})
	}
	return slog.New(handler)
}
