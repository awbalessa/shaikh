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

	conn, err := pgxpool.New(ctx, cfg.PostgresURL)
	if err != nil {
		slog.Error(
			"failed to create pgxpool",
			"error",
			err,
			"postgres_url",
			cfg.PostgresURL,
		)
		os.Exit(1)
	}
	defer conn.Close()

	appCfg := app.AppConfig{
		Config:  cfg,
		Context: ctx,
		Pool:    conn,
	}

	app, err := app.New(&appCfg)
	if err != nil {
		slog.Error(
			"failed to start app",
			"error",
			err,
		)
		os.Exit(1)
	}

	// Write RPF. Init gemini stuff. Write out reranking logic. Start writing agent logic following MCP guidelines. Build out full RAG pipeline including Cache and convo and everything. Wrap external code in loggers and other stuff. Create a pretty logger - it'll go a long way. Start writing tests for functions or components that require tests. Look into interface abstractions for tests.
}
