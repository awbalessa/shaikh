package app

import (
	"context"
	"fmt"

	"github.com/awbalessa/shaikh/apps/server/internal/agent"
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
	Gc    *agent.GeminiClient
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

	gc, err := agent.NewGeminiClient(app.Context, &agent.GeminiClientConfig{
		MaxRetries:     agent.GeminiMaxRetries,
		Timeout:        agent.GeminiTimeout,
		GCPProjectID:   agent.GCPProjectID,
		GeminiBackend:  agent.GeminiBackend,
		GeminiLocation: agent.GeminiLocationGlobal,
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
