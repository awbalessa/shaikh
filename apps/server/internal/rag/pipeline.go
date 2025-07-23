package rag

import (
	"log/slog"

	"github.com/awbalessa/shaikh/apps/server/internal/store"
)

type Pipeline struct {
	store  *store.Store
	vc     *VoyageClient
	logger *slog.Logger
}

func NewPipeline(store *store.Store, vc *VoyageClient) *Pipeline {
	log := slog.Default().With(
		"component", "pipeline",
	)

	return &Pipeline{
		store:  store,
		vc:     vc,
		logger: log,
	}
}
