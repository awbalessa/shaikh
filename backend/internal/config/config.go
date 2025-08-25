package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Config struct {
	Pool      *pgxpool.Pool
	Nats      *nats.Conn
	JetStream jetstream.JetStream
}

func Configure(ctx context.Context, cfg *Env) (*Config, error) {
	pool, err := newPostgresPool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to configure: %w", err)
	}

	nc, err := newNats(&nats.Options{
		Url:           nats.DefaultURL,
		Name:          NatsConnNameApi,
		Timeout:       NatsConnTimeoutTenSeconds,
		PingInterval:  NatsPingIntervalTwentySeconds,
		MaxPingsOut:   NatsMaxPingsOutstandingFive,
		ReconnectWait: NatsReconnectWaitTenSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to configure: %w", err)
	}

	js, err := newJetStream(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to configure: %w", err)
	}

	return &Config{
		Pool:      pool,
		Nats:      nc,
		JetStream: js,
	}, nil
}

func StartAPI(ctx context.Context, cfg *Config) error {
	slog.Info("API server starting...")
	return nil
}

// func StartWorkers(ctx context.Context, cfg *Config, cancel context.CancelFunc) error {
// 	var workers work.WorkerGroup

// 	syncer, err := work.BuildSyncer(ctx, cfg.JetStream)
// 	if err != nil {
// 		return fmt.Errorf("start worker: %w", err)
// 	}
// 	workers.Add(syncer)

// 	workers.StartAll(ctx, cancel)
// 	return nil
// }

func (c *Config) Close() {
	if c.Nats != nil {
		c.Nats.Drain()
	}
	if c.Pool != nil {
		c.Pool.Close()
	}
}

type Env struct {
	PostgresUrl      string
	VoyageAPIKey     string
	DragonFlyAddress string
	Platform         string
}

func LoadEnv() (*Env, error) {
	root, err := findDotEnv()
	if err != nil {
		return nil, err
	}

	envPath := filepath.Join(root, ".env")
	if err = godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}
	return &Env{
		PostgresUrl:      os.Getenv("POSTGRES_URL"),
		VoyageAPIKey:     os.Getenv("VOYAGE_API_KEY"),
		DragonFlyAddress: os.Getenv("DRAGONFLY_ADDR"),
		Platform:         os.Getenv("PLATFORM"),
	}, nil
}

func findDotEnv() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("error getting working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".env")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
