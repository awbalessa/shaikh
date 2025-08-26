package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

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
