package store

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/apps/server/internal/arabic"
	"github.com/awbalessa/shaikh/apps/server/internal/database"
)

func (s *Store) RunSemanticSearch(
	ctx context.Context,
	arg database.SemanticSearchParams,
) ([]database.SemanticSearchRow, error) {
	const method = "RunSemanticSearch"
	log := s.pg.logger.With(
		slog.String("method", method),
		slog.Int("number_of_chunks", int(arg.NumberOfChunks)),
		slog.Int("vector_len", len(arg.Vector.Slice())),
		slog.Any("label_filters", arg.LabelFilters),
	)

	log.InfoContext(ctx, "running semantic search...")

	start := time.Now()
	rows, err := s.pg.queries.SemanticSearch(
		ctx,
		arg,
	)
	if err != nil {
		log.With("err", err).ErrorContext(
			ctx,
			"failed to run semantic search",
		)
		return nil, fmt.Errorf("failed to run semantic search: %w", err)
	}

	duration := time.Since(start)
	log.With(
		slog.Int64("duration_ms", duration.Milliseconds()),
		slog.Int("result_count", len(rows)),
	).InfoContext(ctx, "ran semantic search: returning...")

	return rows, nil
}

func (s *Store) RunLexicalSearch(
	ctx context.Context,
	arg database.LexicalSearchParams,
) ([]database.LexicalSearchRow, error) {
	const method = "RunLexicalSearch"
	log := s.pg.logger.With(
		slog.String("method", method),
		slog.Int("number_of_chunks", int(arg.NumberOfChunks)),
		slog.String("query", arg.Query),
		slog.String("content_type", string(arg.ContentType.ContentType)),
		slog.String("source", string(arg.Source.Source)),
		slog.Int("surah_start", int(arg.SurahStart.Int32)),
		slog.Int("surah_end", int(arg.SurahEnd.Int32)),
		slog.Int("surah", int(arg.Surah.Int32)),
		slog.Int("ayah_start", int(arg.AyahStart.Int32)),
		slog.Int("ayah_end", int(arg.AyahEnd.Int32)),
	)

	tokenized, err := tokenizeQuery(arg.Query)
	if err != nil {
		log.With("err", err).ErrorContext(
			ctx,
			"failed to tokenize query",
		)
		return nil, fmt.Errorf("failed to tokenize query: %w", err)
	}

	arg.Query = tokenized
	log.With(
		"tokenized_query", arg.Query,
	).InfoContext(ctx, "tokenized query: running lexical search...")

	start := time.Now()
	rows, err := s.pg.queries.LexicalSearch(
		ctx,
		arg,
	)
	if err != nil {
		log.With("err", err).ErrorContext(
			ctx,
			"failed to run lexical search",
		)
		return nil, fmt.Errorf("failed to run lexical search: %w", err)
	}

	duration := time.Since(start)
	log.With(
		slog.Int64("duration_ms", duration.Milliseconds()),
		slog.Int("result_count", len(rows)),
	).InfoContext(ctx, "ran lexical search: returning...")

	return rows, nil
}

func tokenizeQuery(query string) (string, error) {
	tokenized, err := arabic.CleanAndFilterStopwords(query)
	if err != nil {
		return "", err
	}

	return tokenized, nil
}
