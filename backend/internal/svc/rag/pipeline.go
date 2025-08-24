package svc

import (
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/config"
)

type PipelineConfig struct {
	Config *config.Config
}

type Pipeline struct {
	vc     *voyageClient
	logger *slog.Logger
}

func NewPipeline(cfg PipelineConfig) *Pipeline {
	vc := newVoyageClient(voyageClientConfig{
		config:     cfg.Config,
		maxRetries: voyageMaxRetriesThree,
		timeout:    voyageTimeoutTenSeconds,
	})

	log := slog.Default().With(
		"component", "pipeline",
	)

	return &Pipeline{
		vc:     vc,
		logger: log,
	}
}
