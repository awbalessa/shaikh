package rag

import (
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/awbalessa/shaikh/backend/internal/repo"
)

const (
	Top5Documents  TopK = 5
	Top10Documents TopK = 10
	Top15Documents TopK = 15
	Top20Documents TopK = 20
)

type TopK int

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
