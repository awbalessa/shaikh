package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/awbalessa/shaikh/api/config"
	"github.com/awbalessa/shaikh/api/internal/http/chat"
	"github.com/awbalessa/shaikh/api/internal/providers/fake"
	"github.com/go-chi/chi/v5"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := config.LoadEnv(); err != nil {
		slog.WarnContext(ctx, "failed to load .env file", "err", err)
	}

	cfg, err := config.New()
	if err != nil {
		slog.ErrorContext(ctx, "env var missing", "err", err)
		os.Exit(1)
	}

	log := config.NewLogger(cfg.Environment)
	slog.SetDefault(log)

	chatHandler := chat.New(fake.Model{})

	r := chi.NewRouter()
	r.Post("/chat", chatHandler.Stream)

	srv := &http.Server{
		Addr: ":" + cfg.Port,
		Handler: r,
	}

	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	srv.Shutdown(context.Background())
}
