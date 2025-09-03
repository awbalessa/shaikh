package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/awbalessa/shaikh/backend/internal/pro"
)

const (
	ServiceName string = "shaikh-api"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	os.Setenv("SERVICE_NAME", ServiceName)
	if err := config.LoadEnv(); err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to load environment")
		stop()
	}

	slog.SetDefault(
		config.NewLogger(os.Getenv("PLATFORM")),
	)

	pg, err := pro.NewPostgres(ctx)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to create postgres")
		stop()
	}

	fly := pro.NewDragonflyCache()
	voy := pro.NewVoyageEmbedderReranker()
	gem, err := pro.NewGeminiLLM(ctx)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to create gemini")
	}
	nc, err := pro.NewNats(ServiceName)
	js, err := pro.NewJS(nc)
}
