package app

import (
	"context"
	"log"

	"github.com/awbalessa/shaikh/internal/config"
	"github.com/awbalessa/shaikh/internal/database"
	"google.golang.org/genai"
)

type App struct {
	Cfg         *config.Config
	Queries     *database.Queries
	GenAIClient *genai.Client
	Context     context.Context
	Logger      *log.Logger
}

func New(cfg *config.Config, queries *database.Queries, genaiClient *genai.Client, logger *log.Logger) *App {
	// Look into instantiating all your wiring here. Probably shouldn't pass in all these things, should create them here.
	return &App{
		Cfg:         cfg,
		Queries:     queries,
		GenAIClient: genaiClient,
		Context:     context.Background(),
		Logger:      logger,
	}
}
