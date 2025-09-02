package api

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/svc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/httplog/v3"
)

type Deps struct {
	Health *svc.HealthReadinessSvc
	Ask    *svc.AskSvc
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	logger, opts := NewLogger()
	r.Use(httplog.RequestLogger(logger, opts))

	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	RegisterSystemRoutes(r, d)

	return r
}

func RegisterSystemRoutes(r chi.Router, d Deps) {
	r.Get("/healthz", healthzHandler())
	r.Get("/readyz", readyzHandler(d.Health))
}

func RegisterAppRoutes(r chi.Router, d Deps) {
	r.Route("/v1/sessions/{sessionID}", func(sr chi.Router) {
		sr.Use(UserAuthMiddleware)
		sr.Use(SessionAuthMiddleware)
		sr.Post("/ask", askHandler(d.Ask))
	})
}

func NewLogger() (*slog.Logger, *httplog.Options) {
	platform := os.Getenv("PLATFORM")
	sname := os.Getenv("SERVICE_NAME")

	var level slog.Level
	var lbct []string
	var lbml int
	if platform == "dev" {
		level = slog.LevelDebug
		lbct = []string{
			"application/json", "application/xml", "text/plain", "text/csv",
			"application/x-www-form-urlencoded", "",
		}
		lbml = 1024
	} else {
		level = slog.LevelInfo
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{})).With(
		slog.String("deployment.environment", platform),
		slog.String("service.name", sname),
	)

	opts := &httplog.Options{
		Level:               level,
		Schema:              httplog.SchemaOTEL,
		RecoverPanics:       true,
		LogRequestHeaders:   []string{"Content-Type", "Origin"},
		LogResponseHeaders:  []string{"Content-Type"},
		LogBodyContentTypes: lbct,
		LogBodyMaxLen:       lbml,
	}

	return logger, opts
}
