package config

import (
	"log/slog"
	"os"
)

type LoggerOptions struct {
	Level slog.Level
	JSON  bool
}

func NewLogger(opts LoggerOptions) *slog.Logger {
	var handler slog.Handler
	if opts.JSON {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: opts.Level,
		})
	} else {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: opts.Level,
		})
	}
	return slog.New(handler)
}
