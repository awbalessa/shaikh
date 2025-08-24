package config

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/awbalessa/shaikh/backend/internal/queue"
	repo "github.com/awbalessa/shaikh/backend/internal/repo"
	"github.com/awbalessa/shaikh/backend/internal/repo/postgres/gen"
	"github.com/awbalessa/shaikh/backend/internal/service/agent"
	"github.com/awbalessa/shaikh/backend/internal/work"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Config struct {
	Pool      *pgxpool.Pool
	Nats      *nats.Conn
	JetStream jetstream.JetStream
	Pipeline  *rag.Pipeline
	Agent     *agent.Agent
}

func Configure(ctx context.Context, cfg *config) (*Config, error) {
	pool, err := NewPostgresPool(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres pool: %w", err)
	}

	nc, err := NewNatsClient(&nats.Options{
		Url:           nats.DefaultURL,
		Name:          queue.NatsConnNameApi,
		Timeout:       queue.NatsConnTimeoutTenSeconds,
		PingInterval:  queue.NatsPingIntervalTwentySeconds,
		MaxPingsOut:   queue.NatsMaxPingsOutstandingFive,
		ReconnectWait: queue.NatsReconnectWaitTenSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create nats connection: %w", err)
	}

	js, err := NewJetStreamClient(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream context: %w", err)
	}

	pipeline := NewPipeline(cfg, store)
	agent, err := NewAgent(agent.AgentConfig{
		Context:  ctx,
		Pipeline: pipeline,
		Store:    store,
		Stream:   js,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return &Config{
		Pool:      pool,
		Nats:      nc,
		JetStream: js,
		Pipeline:  pipeline,
		Agent:     agent,
	}, nil
}

func StartAPI(ctx context.Context, cfg *Config) error {
	slog.Info("API server starting...")
	return nil
}

func StartWorkers(ctx context.Context, cfg *Config, cancel context.CancelFunc) error {
	var workers work.WorkerGroup

	syncer, err := work.BuildSyncer(ctx, cfg.JetStream, cfg.Store)
	if err != nil {
		return fmt.Errorf("start worker: %w", err)
	}
	workers.Add(syncer)

	workers.StartAll(ctx, cancel)
	return nil
}

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

func loadEnv() (*Env, error) {
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

func newStore(cfg *Config, pool *pgxpool.Pool) *repo.Store {
	return repo.New(repo.StoreConfig{
		Config:  cfg,
		Pool:    pool,
		Queries: gen.New(pool),
	})
}

func NewPipeline(cfg *Config, store *repo.Store) *rag.Pipeline {
	return rag.NewPipeline(rag.PipelineConfig{
		Config: cfg,
		Store:  store,
	})
}

func NewAgent(cfg agent.AgentConfig) (*agent.Agent, error) {
	agent, err := agent.NewAgent(agent.AgentConfig{
		Context:  cfg.Context,
		Pipeline: cfg.Pipeline,
		Store:    cfg.Store,
		Stream:   cfg.Stream,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return agent, nil
}
