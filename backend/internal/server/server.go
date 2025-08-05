package server

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

type Server struct {
	Context context.Context
	Cancel  context.CancelFunc
	Nc      *nats.Conn
	Js      jetstream.JetStream
	Conn    *pgxpool.Pool
	Store   *store.Store
	Pipe    *rag.Pipeline
	Agent   *agent.Agent
}

func Serve(cfg *config.Config) (*Server, error) {
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
		return nil, fmt.Errorf("failed to start server: %w", err)
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
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	store := store.New(store.StoreConfig{
		Config:  cfg,
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
		cancel()
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	js, err := NewJetStream(nc)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	agent, err := agent.NewAgent(agent.AgentConfig{
		Context:  ctx,
		Pipeline: pipe,
		Store:    store,
		Stream:   js,
	})
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start server: %w", err)
	}

	return &Server{
		Context: ctx,
		Cancel:  cancel,
		Nc:      nc,
		Js:      js,
		Conn:    conn,
		Store:   store,
		Pipe:    pipe,
		Agent:   agent,
	}, nil
}

func (s *Server) Close() {
	s.Cancel()
	s.Conn.Close()
	s.Nc.Drain()
}
