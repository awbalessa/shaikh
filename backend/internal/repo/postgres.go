package repo

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/repo"
	db "github.com/awbalessa/shaikh/backend/internal/repo/postgres/gen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPostgresRepo creates a new PostgresRepo instance.
func NewPostgresRepo(pool *pgxpool.Pool, logger *slog.Logger) *PostgresRepo {
	return &PostgresRepo{
		queries: db.New(pool),
		pool:    pool,
		logger:  logger,
	}
}

// Ensure PostgresRepo implements the repo.Store interface.
var _ repo.Store = (*PostgresRepo)(nil)

func (pg *PostgresRepo) RunSemanticSearch(
	ctx context.Context,
	arg db.SemanticSearchParams,
) ([]db.SemanticSearchRow, error) {
	const method = "RunSemanticSearch"
	log := pg.logger.With(
		slog.String("method", method),
		slog.Int("number_of_chunks", int(arg.NumberOfChunks)),
		slog.Int("vector_len", len(arg.Vector.Slice())),
		slog.String("content_type_labels", fmt.Sprint(arg.ContentTypeLabels)),
		slog.String("source_labels", fmt.Sprint(arg.SourceLabels)),
		slog.String("surah_labels", fmt.Sprint(arg.SurahLabels)),
		slog.String("ayah_labels", fmt.Sprint(arg.AyahLabels)),
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
	arg db.LexicalSearchParams,
) ([]db.LexicalSearchRow, error) {
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
	tokenized, err := utils.CleanAndFilterStopwords(query)
	if err != nil {
		return "", err
	}

	return tokenized, nil
}

func (pg *postgresClient) GetAyatByKeys(ctx context.Context, surah int32, ayat []int32) ([]db.RagAyat, error) {
	rows, err := pg.queries.GetAyatByKeys(ctx, db.GetAyatByKeysParams{
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
	arg db.GetMemoriesByUserIDParams,
) ([]db.Memory, error) {
	rows, err := pg.queries.GetMemoriesByUserID(ctx, db.GetMemoriesByUserIDParams{
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
	arg db.GetSessionsByUserIDParams,
) ([]db.Session, error) {
	rows, err := pg.queries.GetSessionsByUserID(ctx, db.GetSessionsByUserIDParams{
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
) ([]db.Message, error) {
	rows, err := pg.queries.GetMessagesBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by session id: %w", err)
	}

	return rows, nil
}

func (pg *postgresClient) GetMessagesBySessionIDAsc(
	ctx context.Context,
	sessionID pgtype.UUID,
) ([]db.Message, error) {
	rows, err := pg.queries.GetMessagesBySessionIDAsc(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by session id ascending: %w", err)
	}

	return rows, nil
}

func (pg *postgresClient) GetMessagesBySessionIDOrdered(
	ctx context.Context,
	sessionID pgtype.UUID,
) ([]db.Message, error) {
	rows, err := pg.queries.GetMessagesBySessionIdOrdered(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by session id ascending: %w", err)
	}

	return rows, nil
}

func (pg *postgresClient) CreateMessage(
	ctx context.Context,
	arg db.CreateMessageParams,
) (db.Message, error) {
	rows, err := pg.queries.CreateMessage(ctx, arg)
	if err != nil {
		return db.Message{}, fmt.Errorf("failed to create message: %w", err)
	}

	return rows, nil
}

func (pg *postgresClient) CreateMessageTx(
	ctx context.Context,
	tx pgx.Tx,
	arg db.CreateMessageParams,
) (db.Message, error) {
	q := db.New(tx)

	msg, err := q.CreateMessage(ctx, arg)
	if err != nil {
		return db.Message{}, fmt.Errorf("failed to create message tx: %w", err)
	}

	return msg, nil
}
