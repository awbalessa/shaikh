package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/awbalessa/shaikh/apps/server/internal/agent"
	"github.com/awbalessa/shaikh/apps/server/internal/config"
	"github.com/awbalessa/shaikh/apps/server/internal/database"
	"github.com/awbalessa/shaikh/apps/server/internal/rag"
	"github.com/awbalessa/shaikh/apps/server/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Context context.Context
	Cancel  context.CancelFunc
	Conn    *pgxpool.Pool
	Store   *store.Store
	Pipe    *rag.Pipeline
	Agent   *agent.Agent
}

func Start(cfg *config.Config) (*App, error) {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	pgxCfg, err := pgxpool.ParseConfig(cfg.PostgresConnString)
	if err != nil {
		slog.With(
			slog.Any("err", err),
			slog.String("postgres_url", cfg.PostgresConnString),
		).ErrorContext(
			ctx,
			"failed to parse postgres url",
		)
		cancel()
		return nil, fmt.Errorf("failed to start app: %w", err)
	}

	conn, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		slog.With(
			slog.Any("err", err),
			slog.String("postgres_url", cfg.PostgresConnString),
		).ErrorContext(
			ctx,
			"failed to create postgres conn",
		)
		cancel()
		return nil, fmt.Errorf("failed to start app: %w", err)
	}

	store := store.New(store.StoreConfig{
		Queries: database.New(conn),
	})

	pipe := rag.NewPipeline(rag.PipelineConfig{
		Config: cfg,
		Store:  store,
	})

	agent, err := agent.NewAgent()
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start app: %w", err)
	}

	return &App{
		Context: ctx,
		Cancel:  cancel,
		Conn:    conn,
		Store:   store,
		Pipe:    pipe,
		Agent:   agent,
	}, nil
}

func (a *App) Close() {
	a.Cancel()
	a.Conn.Close()
}
