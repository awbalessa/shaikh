package app

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/queue"
	repo "github.com/awbalessa/shaikh/backend/internal/repo"
	"github.com/awbalessa/shaikh/backend/internal/service/agent"
	"github.com/awbalessa/shaikh/backend/internal/service/rag"
	"github.com/awbalessa/shaikh/backend/internal/worker"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type App struct {
	Pool      *pgxpool.Pool
	Nats      *nats.Conn
	JetStream jetstream.JetStream
	Store     repo.Store
	Pipeline  *rag.Pipeline
	Agent     *agent.Agent
}

func NewApp(ctx context.Context, cfg *Config) (*App, error) {
	pool, err := NewPostgres(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	nc, err := NewNats(&nats.Options{
		Url:           nats.DefaultURL,
		Name:          queue.NatsConnNameApi,
		Timeout:       queue.NatsConnTimeoutTenSeconds,
		PingInterval:  queue.NatsPingIntervalTwentySeconds,
		MaxPingsOut:   queue.NatsMaxPingsOutstandingFive,
		ReconnectWait: queue.NatsReconnectWaitTenSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create nats connection: %w", err)
	}

	js, err := NewJetStream(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream context: %w", err)
	}

	store := NewStore(cfg, pool)
	pipeline := NewPipeline(cfg, store)
	agent, err := NewAgent(agent.AgentConfig{
		Context:  ctx,
		Pipeline: pipeline,
		Store:    store,
		Stream:   js,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return &App{
		Pool:      pool,
		Nats:      nc,
		JetStream: js,
		Store:     store,
		Pipeline:  pipeline,
		Agent:     agent,
	}, nil
}

func StartAPI(ctx context.Context, app *App) error {
	// TODO: Implement API server startup using app.Agent, app.Store etc.
	// For now, just a placeholder.
	slog.Info("API server starting...")
	return nil
}

func StartWorker(ctx context.Context, app *App, cancel context.CancelFunc) error {
	var workers worker.WorkerGroup

	syncer, err := worker.BuildSyncer(ctx, app.JetStream, app.Store)
	if err != nil {
		return fmt.Errorf("start worker: %w", err)
	}
	workers.Add(syncer)

	workers.StartAll(ctx, cancel)
	return nil
}

func (a *App) Close() {
	if a.Nats != nil {
		a.Nats.Drain()
	}
	if a.Pool != nil {
		a.Pool.Close()
	}
}