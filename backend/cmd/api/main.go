package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/awbalessa/shaikh/backend/internal/config"
)

func main() {
	os.Setenv("SERVICE_NAME", "shaikh-api")
	_, cancel := context.WithCancel(
		context.Background(),
	)

	if err := config.LoadEnv(); err != nil {
		cancel()
		slog.With(
			"err", err,
		).Error("failed to start")
		os.Exit(1)
	}

	slog.SetDefault(
		config.NewLogger(os.Getenv("PLATFORM")),
	)
}
