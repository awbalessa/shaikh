package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/config"
	repo "github.com/awbalessa/shaikh/backend/internal/repo"
	database "github.com/awbalessa/shaikh/backend/internal/repo/postgres/gen"
	"github.com/awbalessa/shaikh/backend/internal/service/agent"
	"github.com/awbalessa/shaikh/backend/internal/service/rag"
	"github.com/awbalessa/shaikh/backend/internal/worker"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type App struct {
	Pool  *pgxpool.Pool
	Nats  *nats.Conn
	Agent *agent.Agent
}

func NewPostgres(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
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

func NewStore(cfg *config.Config, pool *pgxpool.Pool) *repo.Store {
	return repo.New(repo.StoreConfig{
		Config:  cfg,
		Pool:    pool,
		Queries: database.New(pool),
	})
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

func NewPipeline(cfg *config.Config, store *store.Store) *rag.Pipeline {
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

func StartAPI(ctx context.Context, cfg *config.Config) (*App, error) {
	pool, err := NewPostgres(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("start api: %w", err)
	}

	store := NewStore(cfg, pool)

	nc, err := NewNats(&nats.Options{
		Url:           nats.DefaultURL,
		Name:          natsConnNameApi,
		Timeout:       natsConnTimeoutTenSeconds,
		PingInterval:  natsPingIntervalTwentySeconds,
		MaxPingsOut:   natsMaxPingsOutstandingFive,
		ReconnectWait: natsReconnectWaitTenSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("start api: %w", err)
	}

	js, err := NewJetStream(nc)
	if err != nil {
		return nil, fmt.Errorf("start api: %w", err)
	}

	pipe := NewPipeline(cfg, store)

	a, err := NewAgent(agent.AgentConfig{
		Context:  ctx,
		Pipeline: pipe,
		Store:    store,
		Stream:   js,
	})
	if err != nil {
		return nil, fmt.Errorf("start api: %w", err)
	}

	return &App{
		Agent: a,
		Nats:  nc,
		Pool:  pool,
	}, nil
}

func StartWorker(ctx context.Context, cfg *config.Config, cancel context.CancelFunc) error {
	pool, err := NewPostgres(ctx, cfg)
	if err != nil {
		return fmt.Errorf("start worker: %w", err)
	}

	store := NewStore(cfg, pool)

	nc, err := NewNats(&nats.Options{
		Url:           nats.DefaultURL,
		Name:          natsConnNameWorker,
		Timeout:       natsConnTimeoutTenSeconds,
		PingInterval:  natsPingIntervalTwentySeconds,
		MaxPingsOut:   natsMaxPingsOutstandingFive,
		ReconnectWait: natsReconnectWaitTenSeconds,
	})
	if err != nil {
		return fmt.Errorf("start worker: %w", err)
	}

	js, err := NewJetStream(nc)
	if err != nil {
		return fmt.Errorf("start worker: %w", err)
	}

	var workers worker.WorkerGroup

	syncer, err := worker.BuildSyncer(ctx, js, store)
	if err != nil {
		return fmt.Errorf("start worker: %w", err)
	}
	workers.Add(syncer)

	workers.StartAll(ctx, cancel)
	return nil
}

const (
	natsConnNameApi               string        = "shaikh-api"
	natsConnNameWorker            string        = "shaikh-worker"
	natsConnTimeoutTenSeconds     time.Duration = 10 * time.Second
	natsPingIntervalTwentySeconds time.Duration = 20 * time.Second
	natsMaxPingsOutstandingFive   int           = 5
	natsReconnectWaitTenSeconds   time.Duration = 10 * time.Second
)

func (a *App) Close() {
	if a.Nats != nil {
		a.Nats.Drain()
	}
	if a.Pool != nil {
		a.Pool.Close()
	}
}
