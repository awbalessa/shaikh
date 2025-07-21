package app

import (
	"context"

	"github.com/awbalessa/shaikh/apps/server/internal/config"
	"github.com/awbalessa/shaikh/apps/server/internal/database"
	"github.com/awbalessa/shaikh/apps/server/internal/rag"
	"github.com/awbalessa/shaikh/apps/server/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AppConfig struct {
	Config  *config.Config
	Context context.Context
	Pool    *pgxpool.Pool
}

type App struct {
	Store *store.Store
	Vc    *rag.VoyageClient
}

func New(app *AppConfig) (*App, error) {
	store := store.New(store.StoreConfig{
		Queries: database.New(app.Pool),
	})

	vc := rag.NewVoyageClient(rag.VoyageClientConfig{
		Config:     app.Config,
		MaxRetries: rag.VoyageMaxRetries,
		Timeout:    rag.VoyageTimeout,
	})

	return &App{
		Store: store,
		Vc:    vc,
	}, nil
}
