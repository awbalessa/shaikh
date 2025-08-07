package main

import (
	"context"
	"log"
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/app"
	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/awbalessa/shaikh/backend/internal/worker"
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
	app, err := app.Start(ctx, cfg)
	if err != nil {
		cancel()
		log.Fatal(err)
	}
	defer app.Close()

	var workers worker.WorkerGroup

	syncer, err := worker.BuildSyncer(ctx, app.Stream, app.Store)
	if err != nil {
		cancel()
		log.Fatal(err)
	}

	workers.Add(syncer)
	workers.StartAll(ctx, cancel)
}
