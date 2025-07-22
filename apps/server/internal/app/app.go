package app

import (
	"context"
	"fmt"

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
	Gc    *rag.GeminiClient
	Pipe  *rag.Pipeline
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

	gc, err := rag.NewGeminiClient(app.Context, &rag.GeminiClientConfig{
		MaxRetries:     rag.GeminiMaxRetries,
		Timeout:        rag.GeminiTimeout,
		GCPProjectID:   rag.GCPProjectID,
		GeminiBackend:  rag.GeminiBackend,
		GeminiLocation: rag.GeminiLocation,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create new app: %w", err)
	}

	pipe := rag.NewPipeline(store, vc)

	return &App{
		Store: store,
		Vc:    vc,
		Gc:    gc,
		Pipe:  pipe,
	}, nil
}
