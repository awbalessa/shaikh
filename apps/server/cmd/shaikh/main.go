package main

import (
	"log"
	"log/slog"

	"github.com/awbalessa/shaikh/apps/server/internal/app"
	"github.com/awbalessa/shaikh/apps/server/internal/config"
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

	app, err := app.Start(cfg)
	if err != nil {
		log.Fatal(err)
	}
	app.Close()
}
