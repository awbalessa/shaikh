package store

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/arabic"
	"github.com/awbalessa/shaikh/backend/internal/database"
	"github.com/jackc/pgx/v5/pgtype"
)

func (pg *postgresClient) RunSemanticSearch(
	ctx context.Context,
	arg database.SemanticSearchParams,
) ([]database.SemanticSearchRow, error) {
	const method = "RunSemanticSearch"
	log := pg.logger.With(
		slog.String("method", method),
		slog.Int("number_of_chunks", int(arg.NumberOfChunks)),
		slog.Int("vector_len", len(arg.Vector.Slice())),
		slog.String("label_filters", fmt.Sprint(arg.LabelFilters)),
	)

	log.DebugContext(ctx, "running semantic search...")

	start := time.Now()
	rows, err := pg.queries.SemanticSearch(
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

func (pg *postgresClient) RunLexicalSearch(
	ctx context.Context,
	arg database.LexicalSearchParams,
) ([]database.LexicalSearchRow, error) {
	const method = "RunLexicalSearch"
	log := pg.logger.With(
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
	rows, err := pg.queries.LexicalSearch(
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

func (pg *postgresClient) GetAyatByKeys(ctx context.Context, surah int32, ayat []int32) ([]database.Ayat, error) {
	rows, err := pg.queries.GetAyatByKeys(ctx, database.GetAyatByKeysParams{
		Surah: surah,
		Ayat:  ayat,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get ayat by keys: %w", err)
	}

	return rows, nil
}

func (pg *postgresClient) GetMemoriesByUserID(
	ctx context.Context,
	arg database.GetMemoriesByUserIDParams,
) ([]database.Memory, error) {
	rows, err := pg.queries.GetMemoriesByUserID(ctx, database.GetMemoriesByUserIDParams{
		NumberOfMemories: arg.NumberOfMemories,
		UserID:           arg.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get memories by user id: %w", err)
	}

	return rows, nil
}

func (pg *postgresClient) GetSessionsByUserID(
	ctx context.Context,
	arg database.GetSessionsByUserIDParams,
) ([]database.Session, error) {
	rows, err := pg.queries.GetSessionsByUserID(ctx, database.GetSessionsByUserIDParams{
		NumberOfSessions: arg.NumberOfSessions,
		UserID:           arg.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by user id: %w", err)
	}

	return rows, nil
}

func (pg *postgresClient) GetMessagesBySessionID(
	ctx context.Context,
	sessionID pgtype.UUID,
) ([]database.Message, error) {
	rows, err := pg.queries.GetMessagesBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by session id: %w", err)
	}

	return rows, nil
}

func (pg *postgresClient) GetMessagesBySessionIDAsc(
	ctx context.Context,
	sessionID pgtype.UUID,
) ([]database.Message, error) {
	rows, err := pg.queries.GetMessagesBySessionIDAsc(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by session id ascending: %w", err)
	}

	return rows, nil
}
