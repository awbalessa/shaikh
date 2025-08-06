package main

import (
	"log"
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/awbalessa/shaikh/backend/internal/server"
)

func main() {
	opts := config.LoggerOptions{
		Level: slog.LevelDebug,
		JSON:  true,
	}

	slog.SetDefault(
		config.NewLogger(opts),
	)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	server, err := server.Serve(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()
}
