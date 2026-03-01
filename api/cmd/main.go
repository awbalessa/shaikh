package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/awbalessa/shaikh/api/config"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := config.LoadEnv(); err != nil {
		slog.WarnContext(ctx, "failed to load .env file", "err", err)
	}

	cfg, err := config.New()
	if err != nil {
		slog.ErrorContext(ctx, "env var missing", "err", err)
		os.Exit(1)
	}

	log := config.NewLogger(cfg.Platform)
	slog.SetDefault(log)
}
