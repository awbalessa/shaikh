package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/awbalessa/shaikh/internal/database"
	"github.com/joho/godotenv"
)

type Config struct {
	GeminiKey      string
	Platform       string
	DatabaseURL    string
	EmbeddingModel string
	Queries        *database.Queries
}

func Load() (*Config, error) {
	// Find the project root directory
	root, err := findRoot()
	if err != nil {
		return nil, err
	}

	// Construct the path to the .env file
	envPath := filepath.Join(root, ".env")

	// Load the .env file
	if err = godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("Error loading .env file: %v", err)
	}

	return &Config{
		GeminiKey:      os.Getenv("GEMINI_API_KEY"),
		DatabaseURL:    os.Getenv("DB_URL"),
		EmbeddingModel: os.Getenv("EMBEDDING_MODEL"),
	}, nil
}

func findRoot() (string, error) {
	// Get current working directory
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("Error getting working directory: %v", err)
	}

	// Recursively search upwards for a directory including go.mod
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
