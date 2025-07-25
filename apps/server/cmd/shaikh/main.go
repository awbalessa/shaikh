package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/awbalessa/shaikh/apps/server/internal/app"
	"github.com/awbalessa/shaikh/apps/server/internal/config"
	"github.com/awbalessa/shaikh/apps/server/internal/models"
	"github.com/awbalessa/shaikh/apps/server/internal/rag"
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

	app, err := app.New(&appCfg)
	if err != nil {
		slog.Error(
			"failed to start app",
			"error",
			err,
		)
		os.Exit(1)
	}

	res, err := app.Pipe.SearchChunks(ctx, rag.SearchParameters{
		RawPrompt:  "من هو ذو القرنين",
		ChunkLimit: rag.Top20Documents,
		PromptsWithFilters: []rag.PromptWithFilters{
			{
				Prompt: "من هو ذو القرنين",
				NullableSurahs: []models.SurahNumber{
					models.SurahNumberEighteen,
					models.SurahNumberFourtyFour,
				},
			},
		},
	})
	if err != nil {
		slog.Error(
			"failed to search",
			"error",
			err,
		)
		os.Exit(1)
	}

	for _, r := range res {
		fmt.Printf("Relevance: %.2f\n\n%s\n\n", r.Relevance, r.EmbeddedChunk)
	}
}
