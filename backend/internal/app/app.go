package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/agent"
	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/awbalessa/shaikh/backend/internal/database"
	"github.com/awbalessa/shaikh/backend/internal/rag"
	"github.com/awbalessa/shaikh/backend/internal/store"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type App struct {
	Agent  *agent.Agent
	Nats   *nats.Conn
	Stream jetstream.JetStream
	Pool   *pgxpool.Pool
	Store  *store.Store
}

func Start(ctx context.Context, cfg *config.Config) (*App, error) {
	pgxCfg, err := pgxpool.ParseConfig(cfg.PostgresConnString)
	if err != nil {
		slog.With(
			slog.Any("err", err),
			slog.String("postgres_url", cfg.PostgresConnString),
		).ErrorContext(
			ctx,
			"failed to parse postgres url",
		)
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
		return nil, fmt.Errorf("failed to start app: %w", err)
	}

	store := store.New(store.StoreConfig{
		Config:  cfg,
		Pool:    conn,
		Queries: database.New(conn),
	})

	pipe := rag.NewPipeline(rag.PipelineConfig{
		Config: cfg,
		Store:  store,
	})

	nc, err := NewNats(&nats.Options{
		Url:          nats.DefaultURL,
		Name:         natsConnNameMain,
		Timeout:      natsConnTimeoutTenSeconds,
		PingInterval: natsPingIntervalTwentySeconds,
		MaxPingsOut:  natsMaxPingsOutstandingFive,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start app: %w", err)
	}

	js, err := NewJetStream(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to start app: %w", err)
	}

	agent, err := agent.NewAgent(agent.AgentConfig{
		Context:  ctx,
		Pipeline: pipe,
		Store:    store,
		Stream:   js,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start app: %w", err)
	}

	return &App{
		Agent:  agent,
		Nats:   nc,
		Stream: js,
		Pool:   conn,
		Store:  store,
	}, nil
}

func (s *App) Close() {
	s.Pool.Close()
	s.Nats.Drain()
}
