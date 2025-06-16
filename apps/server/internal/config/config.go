package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Platform        string
	DatabaseURL     string
	EmbeddingModel  string
	GenerationModel string
	GCPProject      string
	GCPRegion       string
}

func Load() (*Config, error) {
	root, err := findRoot()
	if err != nil {
		return nil, err
	}

	envPath := filepath.Join(root, ".env")
	if err = godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("Error loading .env file: %v", err)
	}

	return &Config{
		Platform:        os.Getenv("PLATFORM"),
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		EmbeddingModel:  os.Getenv("EMBEDDING_MODEL"),
		GenerationModel: os.Getenv("GENERATION_MODEL"),
		GCPProject:      os.Getenv("GCP_PROJECT"),
		GCPRegion:       os.Getenv("GCP_REGION"),
	}, nil
}

func findRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("Error getting working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", os.ErrNotExist
		}
		dir = parent
	}
}
