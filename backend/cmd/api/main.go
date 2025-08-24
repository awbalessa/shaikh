package main

import (
	"context"
	"log/slog"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(
		context.Background(),
	)

	cfg, err := configure()
	if err != nil {
		cancel()
		slog.With(
			"err", err,
		).Error("failed to configure")
		os.Exit(1)
	}

	slog.SetDefault(
		newLogger(cfg.platform),
	)
}
