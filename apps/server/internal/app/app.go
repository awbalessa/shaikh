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
	Pipe  *rag.Pipeline
	Ags   map[agent.AgentName]agent.Agent
}

func New(app AppConfig) (*App, error) {
	store := store.New(store.StoreConfig{
		Queries: database.New(app.Pool),
	})

	pipe := rag.NewPipeline(rag.PipelineConfig{
		Config: app.Config,
		Store:  store,
	})

	router, err := agent.BuildRouter(app.Context)
	if err != nil {
		return nil, fmt.Errorf("failed to start app: %w", err)
	}

	agents := map[agent.AgentName]agent.Agent{
		agent.Router: router,
	}

	return &App{
		Store: store,
		Pipe:  pipe,
		Ags:   agents,
	}, nil
}
