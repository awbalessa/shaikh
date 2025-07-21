package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/awbalessa/shaikh/apps/server/internal/app"
	"github.com/awbalessa/shaikh/apps/server/internal/config"
	"github.com/awbalessa/shaikh/apps/server/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
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

	cfg, err := config.Load()
	if err != nil {
		slog.Error(
			"failed to load config",
			"err",
			err,
		)
		os.Exit(1)
	}

	ctx, cancel := context.WithCancel(
		context.Background(),
	)
	defer cancel()

	appCfg := app.AppConfig{
		Config:  cfg,
		Context: ctx,
	}

	app, err := app.New(&appCfg)
	if err != nil {
		slog.Error(
			"failed to start app",
			"error",
			err,
		)
		os.Exit(1)
	}

	defer app.Pool.Close()

	vecs, err := app.Vc.EmbedQuery(
		ctx,
		[]string{"معنى ختار كفور"},
	)
	if err != nil {
		slog.Error(
			"failed to embed query",
			"error",
			err,
		)
		os.Exit(1)
	}

	semantic := database.SemanticSearchParams{
		NumberOfChunks: 10,
		Vector:         vecs[0],
		LabelFilters:   []int16{},
	}
	semanticRows, err := app.Queries.SemanticSearch(ctx, semantic)
	if err != nil {
		slog.Error(
			"failed to run semantic search",
			"error",
			err,
		)
		os.Exit(1)
	}

	lexical := database.LexicalSearchParams{
		NumberOfChunks: 10,
		Query:          "معنى ختار كفور",
		ContentType:    database.NullContentType{},
		Source:         database.NullSource{},
		Surah:          pgtype.Int4{},
		AyahStart:      pgtype.Int4{},
		AyahEnd:        pgtype.Int4{},
	}

	lexicalRows, err := app.Queries.LexicalSearch(ctx, lexical)
	for i, _ := range semanticRows {
		fmt.Printf("\n")
		fmt.Printf(
			"ID: %d\nScore: %.2f\nChunk: %s\nSource: %v\nSurah: %d\nAyah: %d\n",
			semanticRows[i].ID,
			semanticRows[i].Score,
			semanticRows[i].EmbeddedChunk,
			semanticRows[i].Source,
			semanticRows[i].Surah.Int32,
			semanticRows[i].Ayah.Int32,
		)
		fmt.Printf("\n")
		fmt.Printf(
			"ID: %d\nScore: %.2f\nChunk: %s\nSource: %v\nSurah: %d\nAyah: %d\n",
			lexicalRows[i].ID,
			lexicalRows[i].Score,
			lexicalRows[i].EmbeddedChunk,
			lexicalRows[i].Source,
			lexicalRows[i].Surah.Int32,
			lexicalRows[i].Ayah.Int32,
		)
	}
}
