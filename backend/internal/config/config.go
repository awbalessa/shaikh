package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/svc/agent"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Config struct {
	Pool      *pgxpool.Pool
	Nats      *nats.Conn
	JetStream jetstream.JetStream
	Agent     *agent.Agent
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

const (
	postgresSSLMode               string = "sslmode=disable"
	postgresMaxConns              string = "pool_max_conns=8"
	postgresMinConns              string = "pool_min_conns=2"
	postgresMinIdleConns          string = "pool_min_idle_conns=2"
	postgresMaxConnLifetime       string = "pool_max_conn_lifetime=30m"
	postgresMaxConnLifetimeJitter string = "pool_max_conn_lifetime_jitter=5m"
	postgresMaxConnIdleTime       string = "pool_max_conn_idle_time=15m"
	postgresPoolHealthCheckPeriod string = "pool_health_check_period=30s"
)

func newPostgresPool(ctx context.Context, env *Env) (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf(
		"%s?%s&%s&%s&%s&%s&%s&%s&%s",
		env.PostgresUrl,
		postgresSSLMode,
		postgresMaxConns,
		postgresMinConns,
		postgresMinIdleConns,
		postgresMaxConnLifetime,
		postgresMaxConnLifetimeJitter,
		postgresMaxConnIdleTime,
		postgresPoolHealthCheckPeriod,
	)

	pgxCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		slog.With(
			slog.Any("err", err),
			slog.String("postgres_url", connStr),
		).ErrorContext(
			ctx,
			"failed to parse postgres url",
		)
		return nil, fmt.Errorf("failed to start cfg: %w", err)
	}

	conn, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		slog.With(
			slog.Any("err", err),
			slog.String("postgres_url", connStr),
		).ErrorContext(
			ctx,
			"failed to create postgres conn",
		)
		return nil, fmt.Errorf("failed to create postgres: %w", err)
	}

	return conn, nil
}

const (
	NatsConnNameApi               string        = "shaikh-api"
	NatsConnNameWorker            string        = "shaikh-worker"
	NatsConnTimeoutTenSeconds     time.Duration = 10 * time.Second
	NatsPingIntervalTwentySeconds time.Duration = 20 * time.Second
	NatsMaxPingsOutstandingFive   int           = 5
	NatsReconnectWaitTenSeconds   time.Duration = 10 * time.Second
)

func newNats(opts *nats.Options) (*nats.Conn, error) {
	nc, err := nats.Connect(
		opts.Url,
		nats.Name(opts.Name),
		nats.Timeout(opts.Timeout),
		nats.PingInterval(opts.PingInterval),
		nats.MaxPingsOutstanding(opts.MaxPingsOut),
		nats.ReconnectWait(opts.ReconnectWait),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create nats conn: %w", err)
	}

	return nc, nil
}

func newJetStream(nc *nats.Conn) (jetstream.JetStream, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create new jetstream: %w", err)
	}

	return js, err
}
