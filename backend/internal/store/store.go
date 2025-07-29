package store

import (
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/database"
)

type postgresClient struct {
	queries *database.Queries
	logger  *slog.Logger
}

type StoreConfig struct {
	Queries *database.Queries
}

type Store struct {
	pg *postgresClient
}

func New(cfg StoreConfig) *Store {
	logger := slog.Default().With(
		"component", "postgres",
	)

	return &Store{
		pg: &postgresClient{
			queries: cfg.Queries,
			logger:  logger,
		},
	}
}
