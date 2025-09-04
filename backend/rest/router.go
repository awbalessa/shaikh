package rest

import (
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/svc"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httplog/v3"
)

type Deps struct {
	UserSvc    *svc.UserSvc
	SessionSvc *svc.SessionSvc
	AskSvc     *svc.AskSvc
	AuthSvc    *svc.AuthSvc
	HealthSvc  *svc.HealthReadinessSvc
	JWTValid   *JWTValidator
}

func CreateRouter(d *Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)

	logger, opts := NewLogger()
	r.Use(httplog.RequestLogger(logger, opts))

	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	RegisterSystemRoutes(r, d)
	RegisterAdminRoutes(r, d)
	RegisterAuthRoutes(r, d)
	RegisterAppRoutes(r, d)

	return r
}

func RegisterSystemRoutes(r chi.Router, d *Deps) {
	r.Get("/healthz", healthzHandler())
	r.Get("/readyz", readyzHandler(d.HealthSvc))
}

func RegisterAdminRoutes(r chi.Router, d *Deps) {
	r.Route("/admin", func(ar chi.Router) {
		ar.Use(d.JWTValid.Middleware)
		ar.Use(AdminOnlyMiddleware)

		ar.Delete("/users/{id}", adminDeleteUserHandler(d.UserSvc))
	})
}

func RegisterAuthRoutes(r chi.Router, d *Deps) {
	r.Post("/register", registerHandler(d.UserSvc))
	r.Route("/auth", func(sr chi.Router) {
		sr.Post("/login", loginHandler(d.UserSvc, d.AuthSvc))
		sr.Post("/refresh", refreshHandler(d.AuthSvc))
		sr.Post("/logout", logoutHandler(d.AuthSvc))
		sr.Post("/logout_all", logoutAllHandler(d.AuthSvc))
	})
}

func RegisterAppRoutes(r chi.Router, d *Deps) {
	r.Route("/v1", func(v1 chi.Router) {
		v1.Use(d.JWTValid.Middleware)

		v1.Delete("/users/me", deleteUserHandler(d.UserSvc))

		v1.Post("/sessions", createSessionHandler(d.SessionSvc))

		v1.Route("/sessions/{sessionID}", func(sr chi.Router) {
			sr.Use(SessionAuthMiddleware)
			sr.Delete("", deleteSessionHandler(d.SessionSvc))
			sr.Post("/ask", askHandler(d.AskSvc))
			sr.Patch("/archive", archiveSessionHandler(d.SessionSvc))
		})
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
