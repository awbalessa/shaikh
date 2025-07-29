package main

import (
	"log"
	"log/slog"

	"github.com/awbalessa/shaikh/server/internal/config"
	"github.com/awbalessa/shaikh/server/internal/server"
)

func main() {
	opts := config.LoggerOptions{
		JSON: false,
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
	server.Close()
}
