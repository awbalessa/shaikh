package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/awbalessa/shaikh/apps/server/internal/app"
	"github.com/awbalessa/shaikh/apps/server/internal/config"
)

func main() {
	loggerOpts := config.LoggerOptions{
		Level:  slog.LevelInfo,
		JSON:   true,
		Writer: os.Stdout,
	}

	slog.SetDefault(
		config.NewLogger(loggerOpts),
	)

	cfg, err := config.Load()
	if err != nil {
		slog.Error(
			"failed to load config",
			"error",
			err,
		)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	appCfg := app.AppConfig{
		Config:  cfg,
		Context: ctx,
	}

	app, err := app.Start(&appCfg)
	if err != nil {
		slog.Error(
			"failed to start app",
			"error",
			err,
		)
		os.Exit(1)
	}

	defer app.DatabasePool.Close()

}
