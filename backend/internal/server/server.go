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
)

type Server struct {
	Cancel context.CancelFunc
	Conn   *pgxpool.Pool
	Store  *store.Store
	Pipe   *rag.Pipeline
	Agent  *agent.Agent
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

	agent, err := agent.NewAgent(ctx, pipe)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("failed to start app: %w", err)
	}

	return &Server{
		Cancel: cancel,
		Conn:   conn,
		Store:  store,
		Pipe:   pipe,
		Agent:  agent,
	}, nil
}

func (s *Server) Close() {
	s.Cancel()
	s.Conn.Close()
}
