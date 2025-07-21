package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/awbalessa/shaikh/apps/server/internal/config"
	"github.com/awbalessa/shaikh/apps/server/internal/database"
	"github.com/awbalessa/shaikh/apps/server/internal/rag"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppConfig struct {
	Config  *config.Config
	Context context.Context
}

type App struct {
	Pool    *pgxpool.Pool
	Queries *database.Queries
	Vc      *rag.VoyageClient
}

func New(app *AppConfig) (*App, error) {
	conn, err := pgxpool.New(app.Context, app.Config.PostgresURL)
	if err != nil {
		slog.Error(
			"failed to create pgxpool",
			"error",
			err,
			"postgres_url",
			app.Config.PostgresURL,
		)
		return nil, fmt.Errorf("error creating pgxpool: %w", err)
	}

	vc := rag.NewVoyageClient(rag.VoyageClientConfig{
		Config:     app.Config,
		MaxRetries: rag.VoyageMaxRetries,
		Timeout:    rag.VoyageTimeout,
	})

	queries := database.New(conn)

	return &App{
		Pool:    conn,
		Queries: queries,
		Vc:      vc,
	}, nil
}
