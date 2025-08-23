package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

const (
	PostgresSSLMode               string = "sslmode=disable"
	PostgresMaxConns              string = "pool_max_conns=8"
	PostgresMinConns              string = "pool_min_conns=2"
	PostgresMinIdleConns          string = "pool_min_idle_conns=2"
	PostgresMaxConnLifetime       string = "pool_max_conn_lifetime=30m"
	PostgresMaxConnLifetimeJitter string = "pool_max_conn_lifetime_jitter=5m"
	PostgresMaxConnIdleTime       string = "pool_max_conn_idle_time=15m"
	PostgresPoolHealthCheckPeriod string = "pool_health_check_period=30s"
)

type Config struct {
	PostgresConnString string
	VoyageAPIKey       string
	DragonFlyAddress   string
}

func LoadConfig() (*Config, error) {
	root, err := findDotEnv()
	if err != nil {
		return nil, err
	}

	envPath := filepath.Join(root, ".env")
	if err = godotenv.Load(envPath); err != nil {
		slog.With(
			"envPath", envPath,
		).Error("error loading .env file")
		return nil, fmt.Errorf("error loading .env file: %v", err)
	}

	connStr := fmt.Sprintf(
		"%s?%s&%s&%s&%s&%s&%s&%s&%s",
		os.Getenv("POSTGRES_URL"),
		PostgresSSLMode,
		PostgresMaxConns,
		PostgresMinConns,
		PostgresMinIdleConns,
		PostgresMaxConnLifetime,
		PostgresMaxConnLifetimeJitter,
		PostgresMaxConnIdleTime,
		PostgresPoolHealthCheckPeriod,
	)

	return &Config{
		PostgresConnString: connStr,
		VoyageAPIKey:       os.Getenv("VOYAGE_API_KEY"),
		DragonFlyAddress:   os.Getenv("DRAGONFLY_ADDR"),
	}, nil
}

func findDotEnv() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		slog.With(
			"working_directory", dir,
			"err", err,
		).Error("error getting working directory")
		return "", fmt.Errorf("error getting working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, ".env")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			slog.With(
				"parent", parent,
			).Error(".env file does not exist")
			return "", os.ErrNotExist
		}
		dir = parent
	}
}