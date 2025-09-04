package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/awbalessa/shaikh/backend/internal/pro"
	"github.com/awbalessa/shaikh/backend/internal/svc"
)

const (
	ServiceName string = "shaikh-workers"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	os.Setenv("SERVICE_NAME", ServiceName)
	if err := config.LoadEnv(); err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to load environment")
		os.Exit(1)
	}

	slog.SetDefault(
		config.NewLogger(os.Getenv("PLATFORM")),
	)

	pg, err := pro.NewPostgres(ctx)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to create postgres")
		os.Exit(1)
	}
	defer pg.Pool.Close()

	gem, err := pro.NewGeminiLLM(ctx)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to create gemini")
		os.Exit(1)
	}

	nc, err := pro.NewNats(ServiceName)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to create nats")
		os.Exit(1)
	}
	defer nc.Conn.Drain()

	js, err := pro.NewJS(nc)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to create jetstream")
		os.Exit(1)
	}

	natsps := pro.NewNatsPubSub(nc, js)

	q := pg.Runner()
	pgUserRepo := pro.NewPostgresUserRepo(q)
	pgSessionRepo := pro.NewPostgresSessionRepo(q)
	pgMessageRepo := pro.NewPostgresMessageRepo(q)
	pgMemoryRepo := pro.NewPostgresMemoryRepo(q)

	agent := dom.BuildAgent(gem)

	syncer, err := svc.BuildSyncer(ctx, natsps, pg, pgSessionRepo, pgUserRepo)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to build syncer")
		os.Exit(1)
	}

	summarizer, err := svc.BuildSummarizer(ctx, natsps, agent, pgSessionRepo, pgMessageRepo)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to build summarizer")
		os.Exit(1)
	}

	memorizer, err := svc.BuildMemorizer(ctx, natsps, agent, pgMessageRepo, pgMemoryRepo)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to build memorizer")
		os.Exit(1)
	}

	var group dom.WorkerGroup
	group.Add(syncer)
	group.Add(summarizer)
	group.Add(memorizer)
	group.StartAll(ctx, stop)
}
