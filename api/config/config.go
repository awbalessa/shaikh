package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Environment string
	Port string
	GeminiAPIKey string
}

func New() (Config, error) {
	cfg := Config{
		Environment: getenv(os.Getenv("ENVIRONMENT"), "prod"),
		Port: getenv(os.Getenv("PORT"), "8080"),
		GeminiAPIKey: os.Getenv("GEMINI_API_KEY"),
	}

	if cfg.GeminiAPIKey == "" {
		return Config{}, errors.New("GEMINI_API_KEY is missing")
	}

	return cfg, nil
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}

	return def
}

func LoadEnv() error {
	if err := godotenv.Load(); err != nil {
		return fmt.Errorf("load env: %w", err)
	}
	return nil
}
