package rag

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/awbalessa/shaikh/apps/server/internal/config"
	"github.com/awbalessa/shaikh/apps/server/internal/database"
	"google.golang.org/genai"
)

type Pipeline struct {
	Cfg         *config.Config
	Queries     *database.Queries
	GenAIClient *genai.Client
	Logger      *slog.Logger
	Context     context.Context
}

func NewPipeline(cfg *config.Config, ctx context.Context, queries *database.Queries, logger *slog.Logger) (*Pipeline, err) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  cfg.GCPProject,
		Location: cfg.GCPRegion,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		logger.Error("error creating genai client", "err", err)
		return nil, fmt.Errorf("error creating genai client: %w", err)
	}
	return &Pipeline{
		Cfg:         cfg,
		GenAIClient: client,
		Context:     ctx,
		Queries:     queries,
		Logger:      logger,
	}, nil
}

func (p *Pipeline) logger(method string) *slog.Logger {
	return p.Logger.With("component", "rag", "method", method)
}
