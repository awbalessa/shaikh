// package main

// import (
// 	"context"
// 	"log"
// 	"log/slog"

// 	"github.com/awbalessa/shaikh/backend/internal/app"
// 	"github.com/awbalessa/shaikh/backend/internal/config"
// )

// func main() {
// 	opts := observe.LoggerOptions{
// 		Level: slog.LevelInfo,
// 		JSON:  true,
// 	}

// 	slog.SetDefault(
// 		observe.NewLogger(opts),
// 	)

// 	ctx, cancel := context.WithCancel(
// 		context.Background(),
// 	)

// 	cfg, err := config.Load()
// 	if err != nil {
// 		cancel()
// 		log.Fatal(err)
// 	}

// 	if err := app.StartWorker(ctx, cfg, cancel); err != nil {
// 		log.Fatal(err)
// 	}
// }
