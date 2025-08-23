package repo

import (
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/awbalessa/shaikh/backend/internal/repo/postgres/gen"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

const (
	flyDialTimout      time.Duration = 1 * time.Second
	flyPoolTimeout     time.Duration = 1 * time.Second
	flyReadTimeout     time.Duration = 500 * time.Millisecond
	flyWriteTimeout    time.Duration = 500 * time.Millisecond
	flyConnMaxIdleTime time.Duration = 5 * time.Minute
	flyConnMaxLifetime time.Duration = 30 * time.Minute
	flyPoolSize        int           = 10
	flyMinIdleConns    int           = 2
)

type postgresClient struct {
	Pool    *pgxpool.Pool
	queries *gen.Queries
	logger  *slog.Logger
}

type dragonflyClient struct {
	cli    *redis.Client
	logger *slog.Logger
}

type StoreConfig struct {
	Config  *config.Config
	Pool    *pgxpool.Pool
	Queries *gen.Queries
}

type Store struct {
	Pg  *postgresClient
	Fly *dragonflyClient
}

func New(cfg StoreConfig) *Store {
	pg_log := slog.Default().With(
		"component", "postgres",
	)

	fly_log := slog.Default().With(
		"component", "dragonfly",
	)

	fly := redis.NewClient(&redis.Options{
		Addr:                  cfg.Config.DragonFlyAddress,
		ContextTimeoutEnabled: true,
		DialTimeout:           flyDialTimout,
		PoolTimeout:           flyPoolTimeout,
		ReadTimeout:           flyReadTimeout,
		WriteTimeout:          flyWriteTimeout,
		ConnMaxIdleTime:       flyConnMaxIdleTime,
		ConnMaxLifetime:       flyConnMaxLifetime,
		PoolSize:              flyPoolSize,
		MinIdleConns:          flyMinIdleConns,
		ClientName:            "shaikh",
	})

	return &Store{
		Pg: &postgresClient{
			Pool:    cfg.Pool,
			queries: cfg.Queries,
			logger:  pg_log,
		},
		Fly: &dragonflyClient{
			cli:    fly,
			logger: fly_log,
		},
	}
}
