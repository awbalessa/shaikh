package app

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/repo"
	"github.com/awbalessa/shaikh/backend/internal/repo/postgres/gen"
	"github.com/awbalessa/shaikh/backend/internal/service/agent"
	"github.com/awbalessa/shaikh/backend/internal/service/rag"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func NewPostgres(ctx context.Context, cfg *Config) (*pgxpool.Pool, error) {
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
		return nil, fmt.Errorf("failed to create postgres: %w", err)
	}

	return conn, nil
}

func NewNats(opts *nats.Options) (*nats.Conn, error) {
	nc, err := nats.Connect(
		opts.Url,
		nats.Name(opts.Name),
		nats.Timeout(opts.Timeout),
		nats.PingInterval(opts.PingInterval),
		nats.MaxPingsOutstanding(opts.MaxPingsOut),
		nats.ReconnectWait(opts.ReconnectWait),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create nats conn: %w", err)
	}

	return nc, nil
}

func NewJetStream(nc *nats.Conn) (jetstream.JetStream, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create new jetstream: %w", err)
	}

	return js, err
}

func NewStore(cfg *Config, pool *pgxpool.Pool) *repo.Store {
	return repo.New(repo.StoreConfig{
		Config:  cfg,
		Pool:    pool,
		Queries: gen.New(pool),
	})
}

func NewPipeline(cfg *Config, store *repo.Store) *rag.Pipeline {
	return rag.NewPipeline(rag.PipelineConfig{
		Config: cfg,
		Store:  store,
	})
}

func NewAgent(cfg agent.AgentConfig) (*agent.Agent, error) {
	agent, err := agent.NewAgent(agent.AgentConfig{
		Context:  cfg.Context,
		Pipeline: cfg.Pipeline,
		Store:    cfg.Store,
		Stream:   cfg.Stream,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return agent, nil
}
