package infra

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	dom "github.com/awbalessa/shaikh/backend/internal/domain"
	db "github.com/awbalessa/shaikh/backend/internal/infra/postgres/gen"
	"github.com/awbalessa/shaikh/backend/pkg/utils"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	pool *pgxpool.Pool
	log  *slog.Logger
}

func NewPostgres(pool *pgxpool.Pool, log *slog.Logger) *Postgres {
	return &Postgres{pool: pool, log: log}
}

func (p *Postgres) Runner() db.Querier { return db.New(p.pool) }

func (p *Postgres) WithTx(ctx context.Context, fn func(q db.Querier) error) error {
	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to open pool with tx: %w", err)
	}

	q := db.New(tx)
	if err := fn(q); err != nil {
		_ = tx.Rollback(ctx)
		return fmt.Errorf("failed to run query with tx: %w", err)
	}

	return tx.Commit(ctx)
}

type SearchRepo struct {
	q   db.Querier
	log *slog.Logger
}

func NewSearchRepo(q db.Querier, log *slog.Logger) *SearchRepo {
	return &SearchRepo{q: q, log: log}
}

func (r *SearchRepo) SemanticSearch(
	ctx context.Context,
	vector dom.VectorWithFilter,
	topk dom.TopK,
) ([]dom.Chunk, error) {
	const method = "SemanticSearch"
	log := r.log.With(
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
	rows, err := r.q.SemanticSearch(
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

func (r *SearchRepo) LexicalSearch(
	ctx context.Context,
	arg db.LexicalSearchParams,
) ([]db.LexicalSearchRow, error) {
	const method = "RunLexicalSearch"
	log := r.log.With(
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
	rows, err := r.q.LexicalSearch(
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

type MemoryRepo struct {
	q   db.Querier
	log *slog.Logger
}

func (r *MemoryRepo) GetMemoriesByUserID(
	ctx context.Context,
	arg db.GetMemoriesByUserIDParams,
) ([]db.Memory, error) {
	rows, err := r.q.GetMemoriesByUserID(ctx, db.GetMemoriesByUserIDParams{
		NumberOfMemories: arg.NumberOfMemories,
		UserID:           arg.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get memories by user id: %w", err)
	}

	return rows, nil
}

func (pg *PostgresRepo) GetSessionsByUserID(
	ctx context.Context,
	arg db.GetSessionsByUserIDParams,
) ([]db.Session, error) {
	rows, err := pg.Queries.GetSessionsByUserID(ctx, db.GetSessionsByUserIDParams{
		NumberOfSessions: arg.NumberOfSessions,
		UserID:           arg.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions by user id: %w", err)
	}

	return rows, nil
}

func (pg *PostgresRepo) GetMessagesBySessionID(
	ctx context.Context,
	sessionID pgtype.UUID,
) ([]db.Message, error) {
	rows, err := pg.Queries.GetMessagesBySessionID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by session id: %w", err)
	}

	return rows, nil
}

func (pg *PostgresRepo) GetMessagesBySessionIDAsc(
	ctx context.Context,
	sessionID pgtype.UUID,
) ([]db.Message, error) {
	rows, err := pg.Queries.GetMessagesBySessionIDAsc(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by session id ascending: %w", err)
	}

	return rows, nil
}

func (pg *PostgresRepo) GetMessagesBySessionIDOrdered(
	ctx context.Context,
	sessionID pgtype.UUID,
) ([]db.Message, error) {
	rows, err := pg.Queries.GetMessagesBySessionIdOrdered(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by session id ascending: %w", err)
	}

	return rows, nil
}

func (pg *PostgresRepo) CreateMessage(
	ctx context.Context,
	arg db.CreateMessageParams,
) (db.Message, error) {
	rows, err := pg.Queries.CreateMessage(ctx, arg)
	if err != nil {
		return db.Message{}, fmt.Errorf("failed to create message: %w", err)
	}

	return rows, nil
}

func (pg *PostgresRepo) CreateMessageTx(
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
