package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/awbalessa/shaikh/apps/server/internal/app"
	"github.com/awbalessa/shaikh/apps/server/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
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
			"err",
			err,
		)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	pgxCfg, err := pgxpool.ParseConfig(cfg.PostgresConnString)
	if err != nil {
		slog.Error(
			"failed to create pgxpool",
			"error",
			err,
			"postgres_url",
			cfg.PostgresConnString,
		)
		os.Exit(1)
	}

	conn, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		slog.Error(
			"failed to create pgxpool",
			"error",
			err,
			"postgres_url",
			cfg.PostgresConnString,
		)
		os.Exit(1)
	}
	defer conn.Close()

	appCfg := app.AppConfig{
		Config:  cfg,
		Context: ctx,
		Pool:    conn,
	}

	_, err = app.New(&appCfg)
	if err != nil {
		slog.Error(
			"failed to start app",
			"error",
			err,
		)
		os.Exit(1)
	}
}
