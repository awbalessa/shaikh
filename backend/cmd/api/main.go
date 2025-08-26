package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/awbalessa/shaikh/backend/internal/config"
)

func main() {
	_, cancel := context.WithCancel(
		context.Background(),
	)

	env, err := config.LoadEnv()
	if err != nil {
		cancel()
		slog.With(
			"err", err,
		).Error("failed to start")
		os.Exit(1)
	}

	slog.SetDefault(
		config.NewLogger(env.Platform),
	)
}
