package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/app"
	"github.com/awbalessa/shaikh/backend/internal/config"
)

func main() {
	opts := config.LoggerOptions{
		Level: slog.LevelInfo,
		JSON:  true,
	}

	slog.SetDefault(
		config.NewLogger(opts),
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	cfg, err := config.Load()
	if err != nil {
		cancel()
		log.Fatal(err)
	}

	if err := app.StartWorker(ctx, cfg, cancel); err != nil {
		log.Fatal(err)
	}
}
