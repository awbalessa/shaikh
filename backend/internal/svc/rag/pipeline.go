package service

import (
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/nats-io/nsc/v2/cmd/store"
)


type PipelineConfig struct {
	Config *config.Config
	Store  *store.Store
}

type Pipeline struct {
	store  *store.Store
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
		store:  cfg.Store,
		vc:     vc,
		logger: log,
	}
}
