package main

import (
	"log/slog"
	"os"

	"github.com/awbalessa/shaikh/apps/server/internal/config"
)

func main() {
	loggerOpts := config.LoggerOptions{
		Level:  slog.LevelInfo,
		JSON:   false,
		Writer: os.Stdout,
	}

	slog.SetDefault(
		config.NewLogger(loggerOpts),
	)

	dude := "duuuuuude"
	man := "maaaaaaan"

	slog.Info(
		"here's my message",
		"dude", dude,
		"man", man,
	)

	// cfg, err := config.Load()
	// if err != nil {
	// 	slog.Error(
	// 		"failed to load config",
	// 		"err",
	// 		err,
	// 	)
	// 	os.Exit(1)
	// }

	// ctx, cancel := context.WithCancel(
	// 	context.Background(),
	// )
	// defer cancel()

	// conn, err := pgxpool.New(ctx, cfg.PostgresURL)
	// if err != nil {
	// 	slog.Error(
	// 		"failed to create pgxpool",
	// 		"error",
	// 		err,
	// 		"postgres_url",
	// 		cfg.PostgresURL,
	// 	)
	// 	os.Exit(1)
	// }
	// defer conn.Close()

	// appCfg := app.AppConfig{
	// 	Config:  cfg,
	// 	Context: ctx,
	// 	Pool:    conn,
	// }

	// _, err = app.New(&appCfg)
	// if err != nil {
	// 	slog.Error(
	// 		"failed to start app",
	// 		"error",
	// 		err,
	// 	)
	// 	os.Exit(1)
	// }
}
