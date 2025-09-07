package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/awbalessa/shaikh/backend/internal/pro"
	"github.com/awbalessa/shaikh/backend/internal/svc"
	"github.com/awbalessa/shaikh/backend/rest"
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
	defer pg.Close()

	fly := pro.NewDragonflyCache()
	defer fly.Close()

	voy := pro.NewVoyageEmbedderReranker()
	defer voy.Close()

	gem, err := pro.NewGeminiLLM(ctx)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to create gemini")
		os.Exit(1)
	}
	defer gem.Close()

	nc, err := pro.NewNats(ServiceName)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to create nats")
		os.Exit(1)
	}
	defer nc.Close()

	js, err := pro.NewJS(nc)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to create jetstream")
		os.Exit(1)
	}

	natsps := pro.NewNatsPubSub(nc, js)

	q := pg.Runner()
	pgSearcher := pro.NewPostgresSearcher(q)
	pgUserRepo := pro.NewPostgresUserRepo(q)
	pgSessionRepo := pro.NewPostgresSessionRepo(q)
	pgMessageRepo := pro.NewPostgresMessageRepo(q)
	pgMemoryRepo := pro.NewPostgresMemoryRepo(q)
	pgRefreshRepo := pro.NewPostgresRefreshTokenRepo(q)

	agent := dom.BuildAgent(gem)

	usersvc := svc.BuildUserSvc(pgUserRepo)

	sessionsvc := svc.BuildSessionSvc(pgSessionRepo)

	searchsvc := svc.BuildSearchSvc(pgSearcher, voy, voy)

	asksvc, err := svc.BuildAskSvc(
		ctx, agent,
		fly, pgMemoryRepo,
		pgSessionRepo, pgMessageRepo,
		natsps, searchsvc,
	)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to build ask service")
		os.Exit(1)
	}

		iss, err := svc.NewJWTIssuer(30 * time.Minute)
	if err != nil {
		slog.With(
			"err", err,
		).ErrorContext(ctx, "failed to create jwt issuer")
		os.Exit(1)
	}
	authsvc := svc.BuildAuthSvc(iss, pgRefreshRepo)
	healthsvc := svc.BuildHealthReadinessSvc([]dom.Probe{
		pg, fly, voy, gem, nc,
		svc.NewWorkerProbe(svc.SyncerDurableName, svc.SyncerPingSubject, natsps),
		svc.NewWorkerProbe(svc.SummarizerDurableName, svc.SummarizerPingSubject, natsps),
		svc.NewWorkerProbe(svc.MemorizerDurableName, svc.MemorizerPingSubject, natsps),
	})

	jwtval := rest.NewJWTValidator()

	router := rest.CreateRouter(&rest.Deps{
		UserSvc:    usersvc,
		SessionSvc: sessionsvc,
		AskSvc:     asksvc,
		AuthSvc:    authsvc,
		HealthSvc:  healthsvc,
		JWTValid:   jwtval,
	})

	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	go func() {
		slog.With("addr", srv.Addr).InfoContext(ctx, "server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.With("err", err).ErrorContext(ctx, "server failed")
			os.Exit(1)
		}
	}()

	<-ctx.Done()
	slog.Info("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.With("err", err).ErrorContext(shutdownCtx, "graceful shutdown failed")
	}
}
