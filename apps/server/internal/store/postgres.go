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
		slog.String("label_filters", fmt.Sprint(arg.LabelFilters)),
	)

	log.DebugContext(ctx, "running semantic search...")

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

	log.With(
		slog.String("duration", time.Since(start).String()),
		slog.Int("result_count", len(rows)),
	).DebugContext(ctx, "ran semantic search: returning...")

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
		slog.String("content_types", fmt.Sprint(arg.ContentTypes)),
		slog.String("sources", fmt.Sprint(arg.Sources)),
		slog.String("surahs", fmt.Sprint(arg.Surahs)),
		slog.String("ayahs", fmt.Sprint(arg.Ayahs)),
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
	).DebugContext(ctx, "tokenized query: running lexical search...")

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

	log.With(
		slog.String("tokenized_query", arg.Query),
		slog.String("duration", time.Since(start).String()),
		slog.Int("result_count", len(rows)),
	).DebugContext(ctx, "ran lexical search: returning...")

	return rows, nil
}

func tokenizeQuery(query string) (string, error) {
	tokenized, err := arabic.CleanAndFilterStopwords(query)
	if err != nil {
		return "", err
	}

	return tokenized, nil
}
