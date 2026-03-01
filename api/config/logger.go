package config

import (
	"io"
	"log/slog"
	"os"
)

func NewLogger(env string) *slog.Logger {
	var out io.Writer = os.Stdout

	opts := &slog.HandlerOptions{
		AddSource: env == "dev",
		Level:     slog.LevelInfo,
	}

	if env == "dev" {
		opts.Level = slog.LevelDebug
	}

	var handler slog.Handler

	if env == "dev" {
		handler = slog.NewTextHandler(out, opts)
	} else {
		handler = slog.NewJSONHandler(out, opts)
	}

	logger := slog.New(handler)
	return logger
}
