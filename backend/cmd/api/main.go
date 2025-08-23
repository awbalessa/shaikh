package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/app"
	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/awbalessa/shaikh/backend/internal/observe"
)

func main() {
	opts := observe.LoggerOptions{
		Level: slog.LevelInfo,
		JSON:  true,
	}

	slog.SetDefault(
		observe.NewLogger(opts),
	)

	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	cfg, err := config.Load()
	if err != nil {
		cancel()
		log.Fatal(err)
	}

	app, err := app.StartAPI(ctx, cfg)
	if err != nil {
		cancel()
		log.Fatal(err)
	}
	defer app.Close()
}
