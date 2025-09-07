package pro

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"os"
	"time"

	db "github.com/awbalessa/shaikh/backend/internal/db/gen"
	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/awbalessa/shaikh/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"golang.org/x/sync/errgroup"
)

type Postgres struct {
	Pool *pgxpool.Pool
}

func NewPostgres(ctx context.Context) (*Postgres, error) {
	connStr := os.Getenv("POSTGRES_URL") +
		"?sslmode=disable" +
		"&pool_max_conns=8" +
		"&pool_min_conns=2" +
		"&pool_min_idle_conns=2" +
		"&pool_max_conn_lifetime=30m" +
		"&pool_max_conn_lifetime_jitter=5m" +
		"&pool_max_conn_idle_time=15m" +
		"&pool_health_check_period=30s"

	pgxCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
	}

	conn, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return nil, dom.NewTaggedError(dom.ErrUnavailable, err)
	}

	return &Postgres{Pool: conn}, nil
}

func (p *Postgres) Runner() db.Querier { return db.New(p.Pool) }

func (p *Postgres) WithTx(ctx context.Context, fn func(q db.Querier) error) error {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return dom.NewTaggedError(dom.ErrUnavailable, err)
	}

	q := db.New(tx)
	if err := fn(q); err != nil {
		_ = tx.Rollback(ctx)
		return dom.NewTaggedError(dom.ErrInternal, err)
	}
	if err := tx.Commit(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return dom.NewTaggedError(dom.ErrInternal, err)
	}
	return nil
}

func (p *Postgres) Name() string { return "db" }

func (p *Postgres) Ping(ctx context.Context) error {
	if err := p.Pool.Ping(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return dom.NewTaggedError(dom.ErrUnavailable, err)
	}
	return nil
}

func (p *Postgres) Close() {
	if p.Pool != nil {
		p.Pool.Close()
	}
}

func (p *Postgres) Begin(ctx context.Context) (dom.Tx, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return nil, dom.NewTaggedError(dom.ErrUnavailable, err)
	}
	q := db.New(tx)
	return &PostgresTx{q: q, tx: tx}, nil
}

type PostgresTx struct {
	q  db.Querier
	tx pgx.Tx
}

func (t *PostgresTx) Get(repo any) error {
	switch r := repo.(type) {
	case *dom.MessageRepo:
		*r = &PostgresMessageRepo{q: t.q}
		return nil
	default:
		return dom.NewTaggedError(dom.ErrInvalidInput, errors.New("unsupported repo type"))
	}
}

func (t *PostgresTx) Commit(ctx context.Context) error {
	if err := t.tx.Commit(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return dom.NewTaggedError(dom.ErrInternal, err)
	}
	return nil
}

func (t *PostgresTx) Rollback(ctx context.Context) error {
	if err := t.tx.Rollback(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return dom.NewTaggedError(dom.ErrTimeout, err)
		}
		return dom.NewTaggedError(dom.ErrInternal, err)
	}
	return nil
}

type PostgresUserRepo struct{ q db.Querier }

func NewPostgresUserRepo(q db.Querier) *PostgresUserRepo { return &PostgresUserRepo{q: q} }

func (u *PostgresUserRepo) CreateUser(ctx context.Context, id uuid.UUID, email, hash string) (*dom.User, error) {
	row, err := u.q.CreateUser(ctx, db.CreateUserParams{ID: id, Email: email, PasswordHash: hash})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, dom.NewTaggedError(dom.ErrConflict, err)
		}
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
	}
	return &dom.User{
		ID: row.ID, Email: row.Email, PasswordHash: row.PasswordHash,
		UpdatedAt: row.UpdatedAt, TotalMessages: row.TotalMessages,
		TotalMessagesMemorized: row.TotalMessagesMemorized,
	}, nil
}

func (u *PostgresUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*dom.User, error) {
	row, err := u.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		}
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
	}
	return &dom.User{
		ID: row.ID, Email: row.Email, PasswordHash: row.PasswordHash,
		UpdatedAt: row.UpdatedAt, TotalMessages: row.TotalMessages,
		TotalMessagesMemorized: row.TotalMessagesMemorized,
	}, nil
}

func (u *PostgresUserRepo) GetUserByEmail(ctx context.Context, email string) (*dom.User, error) {
	row, err := u.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		}
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
	}
	return &dom.User{
		ID: row.ID, Email: row.Email, PasswordHash: row.PasswordHash,
		UpdatedAt: row.UpdatedAt, TotalMessages: row.TotalMessages,
		TotalMessagesMemorized: row.TotalMessagesMemorized,
	}, nil
}

func (u *PostgresUserRepo) IncrementUserMessagesByID(
	ctx context.Context,
	id uuid.UUID,
	delta int32,
	deltaMemorized int32,
) (*dom.User, error) {
	row, err := u.q.IncrementUserMessagesByID(ctx, db.IncrementUserMessagesByIDParams{
		DeltaMessages:          delta,
		DeltaMessagesMemorized: deltaMemorized,
		ID:                     id,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		}
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
	}

	return &dom.User{
		ID:                     row.ID,
		Email:                  row.Email,
		PasswordHash:           row.PasswordHash,
		UpdatedAt:              row.UpdatedAt,
		TotalMessages:          row.TotalMessages,
		TotalMessagesMemorized: row.TotalMessagesMemorized,
	}, nil
}

func (u *PostgresUserRepo) ListUsersWithBacklog(
	ctx context.Context,
) ([]*dom.User, error) {
	rows, err := u.q.ListUsersWithBacklog(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		}
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
	}

	final := make([]*dom.User, 0, len(rows))
	for _, r := range rows {
		final = append(final, &dom.User{
			ID:                     r.ID,
			Email:                  r.Email,
			PasswordHash:           r.PasswordHash,
			UpdatedAt:              r.UpdatedAt,
			TotalMessages:          r.TotalMessages,
			TotalMessagesMemorized: r.TotalMessagesMemorized,
		})
	}

	return final, nil
}

func (u *PostgresUserRepo) DeleteUserByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	if err := u.q.DeleteUserByID(ctx, id); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return dom.NewTaggedError(dom.ErrTimeout, err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return dom.NewTaggedError(dom.ErrNoResults, err)
		}
		return dom.NewTaggedError(dom.ErrInternal, err)
	}
	return nil
}

type PostgresSessionRepo struct {
	q db.Querier
}

func NewPostgresSessionRepo(q db.Querier) *PostgresSessionRepo {
	return &PostgresSessionRepo{q: q}
}

func (s *PostgresSessionRepo) CreateSession(
	ctx context.Context,
	id, userID uuid.UUID,
) (*dom.Session, error) {
	row, err := s.q.CreateSession(ctx, db.CreateSessionParams{
		ID:     id,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		}
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
	}

	return &dom.Session{
		ID:           row.ID,
		UserID:       row.UserID,
		LastAccessed: row.UpdatedAt,
		ArchivedAt:   row.ArchivedAt,
		MaxTurn:      row.MaxTurn,
		Summary:      row.Summary,
	}, nil
}

func (s *PostgresSessionRepo) GetSessionByID(
	ctx context.Context,
	id uuid.UUID,
) (*dom.Session, error) {
	row, err := s.q.GetSessionByID(ctx, id)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		}
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
	}

	return &dom.Session{
		ID:                row.ID,
		UserID:            row.UserID,
		LastAccessed:      row.UpdatedAt,
		MaxTurn:           row.MaxTurn,
		MaxTurnSummarized: row.MaxTurnSummarized,
		ArchivedAt:        row.ArchivedAt,
		Summary:           row.Summary,
	}, nil
}

func (s *PostgresSessionRepo) GetSessionsByUserID(
	ctx context.Context,
	userID uuid.UUID,
	numberOfSessions int32,
) ([]*dom.Session, error) {
	rows, err := s.q.GetSessionsByUserID(ctx, db.GetSessionsByUserIDParams{
		UserID:           userID,
		NumberOfSessions: int64(numberOfSessions),
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		}
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
	}

	final := make([]*dom.Session, 0, len(rows))
	for _, r := range rows {
		final = append(final, &dom.Session{
			ID:           r.ID,
			UserID:       r.UserID,
			LastAccessed: r.UpdatedAt,
			ArchivedAt:   r.ArchivedAt,
			MaxTurn:      r.MaxTurn,
			Summary:      r.Summary,
		})
	}

	return final, nil
}

func (s *PostgresSessionRepo) UpdateSessionByID(
	ctx context.Context,
	id uuid.UUID,
	maxTurn *int32,
	maxTurnSummarized *int32,
	summary *string,
	archivedAt *time.Time,
) (*dom.Session, error) {
	row, err := s.q.UpdateSessionByID(ctx, db.UpdateSessionByIDParams{
		MaxTurn:           maxTurn,
		MaxTurnSummarized: maxTurnSummarized,
		ArchivedAt:        archivedAt,
		Summary:           summary,
		ID:                id,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		}
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
	}

	return &dom.Session{
		ID:           row.ID,
		UserID:       row.UserID,
		LastAccessed: row.UpdatedAt,
		ArchivedAt:   row.ArchivedAt,
		MaxTurn:      row.MaxTurn,
		Summary:      row.Summary,
	}, nil
}

func (s *PostgresSessionRepo) GetMaxTurnByID(
	ctx context.Context,
	id uuid.UUID,
) (int32, error) {
	max, err := s.q.GetMaxTurnByID(ctx, id)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return 0, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return 0, dom.NewTaggedError(dom.ErrNoResults, err)
		}
		return 0, dom.NewTaggedError(dom.ErrInternal, err)
	}

	return max, nil
}

func (s *PostgresSessionRepo) ListSessionsWithBacklog(ctx context.Context) ([]*dom.Session, error) {
	rows, err := s.q.ListSessionsWithBacklog(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		}
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
	}

	final := make([]*dom.Session, 0, len(rows))
	for _, r := range rows {
		final = append(final, &dom.Session{
			ID:                r.ID,
			UserID:            r.UserID,
			LastAccessed:      r.UpdatedAt,
			MaxTurn:           r.MaxTurn,
			MaxTurnSummarized: r.MaxTurnSummarized,
			ArchivedAt:        r.ArchivedAt,
			Summary:           r.Summary,
		})
	}

	return final, nil
}

func (s *PostgresSessionRepo) DeleteSessionByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	if err := s.q.DeleteSessionByID(ctx, id); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return dom.NewTaggedError(dom.ErrTimeout, err)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return dom.NewTaggedError(dom.ErrNoResults, err)
		}
		return dom.NewTaggedError(dom.ErrInternal, err)
	}
	return nil
}

type PostgresMessageRepo struct {
	q db.Querier
}

func NewPostgresMessageRepo(q db.Querier) *PostgresMessageRepo {
	return &PostgresMessageRepo{q: q}
}

func (m *PostgresMessageRepo) CreateMessage(
	ctx context.Context,
	msg dom.Message,
) (dom.Message, error) {
	meta := msg.Meta()
	role := msg.Role()

	row, err := m.q.CreateMessage(ctx, db.CreateMessageParams{
		SessionID:         meta.SessionID,
		UserID:            meta.UserID,
		Role:              role,
		Model:             meta.Model,
		Turn:              meta.Turn,
		TotalInputTokens:  meta.TotalInputTokens,
		TotalOutputTokens: meta.TotalOutputTokens,
		Content:           meta.Content,
		FunctionName:      meta.FunctionName,
		FunctionCall:      meta.FunctionCall,
		FunctionResponse:  meta.FunctionResponse,
	})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		case errors.Is(err, sql.ErrNoRows):
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		default:
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return nil, dom.NewTaggedError(dom.ErrConflict, err)
			}
			return nil, dom.NewTaggedError(dom.ErrInternal, err)
		}
	}

	return fromDbMessage(row), nil
}

func (m *PostgresMessageRepo) GetMessagesBySessionID(
	ctx context.Context,
	sessionID uuid.UUID,
) ([]dom.Message, error) {
	rows, err := m.q.GetMessagesBySessionIdOrdered(ctx, sessionID)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		case errors.Is(err, sql.ErrNoRows):
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		default:
			return nil, dom.NewTaggedError(dom.ErrInternal, err)
		}
	}

	final := make([]dom.Message, 0, len(rows))
	for _, r := range rows {
		final = append(final, fromDbMessage(r))
	}
	return final, nil
}

func (m *PostgresMessageRepo) GetUserMessagesByUserID(
	ctx context.Context,
	userID uuid.UUID,
	numberOfMessages int32,
) ([]dom.Message, error) {
	rows, err := m.q.GetUserMessagesByUserID(ctx, db.GetUserMessagesByUserIDParams{
		UserID:           userID,
		NumberOfMessages: int64(numberOfMessages),
	})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		case errors.Is(err, sql.ErrNoRows):
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		default:
			return nil, dom.NewTaggedError(dom.ErrInternal, err)
		}
	}

	final := make([]dom.Message, 0, len(rows))
	for _, r := range rows {
		final = append(final, fromDbMessage(r))
	}
	return final, nil
}

type PostgresMemoryRepo struct {
	q db.Querier
}

func NewPostgresMemoryRepo(q db.Querier) *PostgresMemoryRepo {
	return &PostgresMemoryRepo{q: q}
}

func (m *PostgresMemoryRepo) CreateMemory(
	ctx context.Context,
	userID uuid.UUID,
	sourceMsg string,
	confidence float32,
	uniqueKey string,
	content string,
) (*dom.Memory, error) {
	row, err := m.q.CreateMemory(ctx, db.CreateMemoryParams{
		UserID:        userID,
		SourceMessage: sourceMsg,
		Confidence:    confidence,
		UniqueKey:     uniqueKey,
		Memory:        content,
	})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		case errors.Is(err, sql.ErrNoRows):
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		default:
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return nil, dom.NewTaggedError(dom.ErrConflict, err)
			}
			return nil, dom.NewTaggedError(dom.ErrInternal, err)
		}
	}

	return &dom.Memory{
		ID:         row.ID,
		UserID:     row.UserID,
		UpdatedAt:  row.UpdatedAt,
		SourceMsg:  row.SourceMessage,
		Confidence: row.Confidence,
		UniqueKey:  row.UniqueKey,
		Content:    row.Memory,
	}, nil
}

func (m *PostgresMemoryRepo) UpsertMemory(
	ctx context.Context,
	userID uuid.UUID,
	sourceMsg string,
	confidence float32,
	uniqueKey string,
	content string,
) (*dom.Memory, error) {
	row, err := m.q.UpsertMemory(ctx, db.UpsertMemoryParams{
		UserID:        userID,
		SourceMessage: sourceMsg,
		Confidence:    confidence,
		UniqueKey:     uniqueKey,
		Memory:        content,
	})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		case errors.Is(err, sql.ErrNoRows):
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		default:
			return nil, dom.NewTaggedError(dom.ErrInternal, err)
		}
	}

	return &dom.Memory{
		ID:         row.ID,
		UserID:     row.UserID,
		UpdatedAt:  row.UpdatedAt,
		SourceMsg:  row.SourceMessage,
		Confidence: row.Confidence,
		UniqueKey:  row.UniqueKey,
		Content:    row.Memory,
	}, nil
}

func (m *PostgresMemoryRepo) GetMemoriesByUserID(
	ctx context.Context,
	userID uuid.UUID,
	numberOfMemories int32,
) ([]*dom.Memory, error) {
	rows, err := m.q.GetMemoriesByUserID(ctx, db.GetMemoriesByUserIDParams{
		UserID:           userID,
		NumberOfMemories: int64(numberOfMemories),
	})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		case errors.Is(err, sql.ErrNoRows):
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		default:
			return nil, dom.NewTaggedError(dom.ErrInternal, err)
		}
	}

	final := make([]*dom.Memory, 0, len(rows))
	for _, r := range rows {
		final = append(final, &dom.Memory{
			ID:         r.ID,
			UserID:     r.UserID,
			UpdatedAt:  r.UpdatedAt,
			SourceMsg:  r.SourceMessage,
			Confidence: r.Confidence,
			UniqueKey:  r.UniqueKey,
			Content:    r.Memory,
		})
	}
	return final, nil
}

func (m *PostgresMemoryRepo) DeleteMemoryByUserIDKey(
	ctx context.Context,
	userID uuid.UUID,
	key string,
) error {
	if err := m.q.DeleteMemoryByUserIDKey(ctx, db.DeleteMemoryByUserIDKeyParams{
		UserID: userID,
		Key:    key,
	}); err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return dom.NewTaggedError(dom.ErrTimeout, err)
		case errors.Is(err, sql.ErrNoRows):
			return dom.NewTaggedError(dom.ErrNoResults, err)
		default:
			return dom.NewTaggedError(dom.ErrInternal, err)
		}
	}
	return nil
}

type PostgresSearcher struct {
	q db.Querier
}

func NewPostgresSearcher(q db.Querier) *PostgresSearcher {
	return &PostgresSearcher{q: q}
}

func (r *PostgresSearcher) ParallelSemanticSearch(
	ctx context.Context,
	queries []dom.FullQueryContext,
	topk int,
) ([][]dom.Chunk, error) {
	if len(queries) == 0 {
		return nil, dom.NewTaggedError(dom.ErrInvalidInput, nil)
	}

	chunksPerThread := max(1, topk/len(queries))
	results := make([][]dom.Chunk, len(queries))

	g, ctx := errgroup.WithContext(ctx)

	for i, query := range queries {
		i, query := i, query
		g.Go(func() error {
			if query.Vector == nil {
				return dom.NewTaggedError(dom.ErrInvalidInput, nil)
			}
			rows, derr := r.SemanticSearch(ctx, query.VectorWithLabel, chunksPerThread)
			if derr != nil {
				return derr
			}
			results[i] = rows
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
	}
	return results, nil
}

func (r *PostgresSearcher) SemanticSearch(
	ctx context.Context,
	vector dom.VectorWithLabel,
	topk int,
) ([]dom.Chunk, error) {
	params := toSemSearchParams(vector, topk)

	rows, err := r.q.SemanticSearch(ctx, params)
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		case errors.Is(err, sql.ErrNoRows):
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		default:
			return nil, dom.NewTaggedError(dom.ErrInternal, err)
		}
	}

	final := make([]dom.Chunk, 0, len(rows))
	for _, row := range rows {
		final = append(final, dom.Chunk{
			Document: dom.Document{
				ID:          int32(row.ID),
				Source:      row.Source,
				Content:     row.EmbeddedChunk,
				SurahNumber: *row.Surah,
				AyahNumber:  *row.Ayah,
			},
			ParentID: row.ParentID,
		})
	}
	return final, nil
}

func (r *PostgresSearcher) ParallelLexicalSearch(
	ctx context.Context,
	queries []dom.FullQueryContext,
	topk int,
) ([][]dom.Chunk, error) {
	if len(queries) == 0 {
		return nil, dom.NewTaggedError(dom.ErrInvalidInput, nil)
	}

	chunksPerThread := max(1, topk/len(queries))
	results := make([][]dom.Chunk, len(queries))

	g, ctx := errgroup.WithContext(ctx)

	for i, query := range queries {
		i, query := i, query
		g.Go(func() error {
			rows, derr := r.LexicalSearch(ctx, query.QueryWithFilter, chunksPerThread)
			if derr != nil {
				return derr
			}
			results[i] = rows
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, dom.NewTaggedError(dom.ErrInternal, err)
	}
	return results, nil
}

func (r *PostgresSearcher) LexicalSearch(
	ctx context.Context,
	query dom.QueryWithFilter,
	topk int,
) ([]dom.Chunk, error) {
	tokenized, err := tokenizeQuery(query.Query)
	if err != nil {
		return nil, dom.NewTaggedError(dom.ErrInvalidInput, err)
	}
	query.Query = tokenized

	rows, err := r.q.LexicalSearch(ctx, db.LexicalSearchParams{
		NumberOfChunks: int64(topk),
		Query:          query.Query,
		ContentTypes:   query.OptionalContentTypes,
		Sources:        query.OptionalSources,
		Surahs:         query.OptionalSurahs,
		Ayahs:          query.OptionalAyahs,
	})
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return nil, dom.NewTaggedError(dom.ErrTimeout, err)
		case errors.Is(err, sql.ErrNoRows):
			return nil, dom.NewTaggedError(dom.ErrNoResults, err)
		default:
			return nil, dom.NewTaggedError(dom.ErrInternal, err)
		}
	}

	final := make([]dom.Chunk, 0, len(rows))
	for _, row := range rows {
		final = append(final, dom.Chunk{
			Document: dom.Document{
				ID:          int32(row.ID),
				Source:      row.Source,
				Content:     row.EmbeddedChunk,
				SurahNumber: *row.Surah,
				AyahNumber:  *row.Ayah,
			},
			ParentID: row.ParentID,
		})
	}
	return final, nil
}

func tokenizeQuery(query string) (string, error) {
	tokenized, err := utils.CleanAndFilterStopwords(query)
	if err != nil {
		return "", err
	}

	return tokenized, nil
}

func toSemSearchParams(vwl dom.VectorWithLabel, topk int) db.SemanticSearchParams {
	var (
		contentTypes []int16 = []int16{}
		sources      []int16 = []int16{}
		surahs       []int16 = []int16{}
		ayahs        []int16 = []int16{}
	)

	for _, ct := range vwl.OptionalContentTypeLabels {
		contentTypes = append(contentTypes, int16(ct))
	}
	for _, so := range vwl.OptionalSourceLabels {
		sources = append(sources, int16(so))
	}
	for _, sur := range vwl.OptionalSurahLabels {
		surahs = append(surahs, int16(sur))
	}
	for _, ay := range vwl.OptionalAyahLabels {
		ayahs = append(ayahs, int16(ay))
	}

	return db.SemanticSearchParams{
		NumberOfChunks:    int64(topk),
		Vector:            pgvector.NewVector(vwl.Vector),
		ContentTypeLabels: contentTypes,
		SourceLabels:      sources,
		SurahLabels:       surahs,
		AyahLabels:        ayahs,
	}
}

func fromDbMessage(row db.Message) dom.Message {
	meta := dom.MsgMeta{
		ID:                row.ID,
		SessionID:         row.SessionID,
		UserID:            row.UserID,
		Model:             row.Model,
		Turn:              row.Turn,
		TotalInputTokens:  row.TotalInputTokens,
		TotalOutputTokens: row.TotalOutputTokens,
		Content:           row.Content,
		FunctionName:      row.FunctionName,
		FunctionCall:      row.FunctionCall,
		FunctionResponse:  row.FunctionResponse,
	}

	switch row.Role {
	case dom.MessageRoleUser:
		return &dom.UserMessage{
			MsgMeta:    meta,
			MsgContent: *meta.Content,
		}
	case dom.MessageRoleFunction:
		return &dom.FunctionMessage{
			MsgMeta:          meta,
			FunctionName:     *meta.FunctionName,
			FunctionCall:     meta.FunctionCall,
			FunctionResponse: meta.FunctionResponse,
		}
	case dom.MessageRoleModel:
		return &dom.ModelMessage{
			MsgMeta:    meta,
			MsgContent: *meta.Content,
		}
	default:
		return nil
	}
}

type PostgresRefreshTokenRepo struct {
	q db.Querier
}

func NewPostgresRefreshTokenRepo(q db.Querier) *PostgresRefreshTokenRepo {
	return &PostgresRefreshTokenRepo{q: q}
}

func (r *PostgresRefreshTokenRepo) CreateRefreshToken(
	ctx context.Context,
	userID uuid.UUID,
	ttl time.Duration,
) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", dom.NewTaggedError(dom.ErrInternal, err)
	}
	refreshToken := base64.RawURLEncoding.EncodeToString(raw)
	hash := sha256.Sum256([]byte(refreshToken))
	expiry := time.Now().Add(ttl)

	if _, err := r.q.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: hex.EncodeToString(hash[:]),
		ExpiresAt: expiry,
	}); err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return "", dom.NewTaggedError(dom.ErrTimeout, err)
		default:
			return "", dom.NewTaggedError(dom.ErrInternal, err)
		}
	}

	return refreshToken, nil
}

func (r *PostgresRefreshTokenRepo) ValidateAndRotate(
	ctx context.Context,
	rawToken string,
) (uuid.UUID, error) {
	hash := sha256.Sum256([]byte(rawToken))
	row, err := r.q.GetRefreshTokenByHash(ctx, hex.EncodeToString(hash[:]))
	if err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return uuid.Nil, dom.NewTaggedError(dom.ErrTimeout, err)
		case errors.Is(err, sql.ErrNoRows):
			return uuid.Nil, dom.NewTaggedError(dom.ErrNoResults, err)
		default:
			return uuid.Nil, dom.NewTaggedError(dom.ErrInternal, err)
		}
	}

	if row.RevokedAt != nil || time.Now().After(row.ExpiresAt) {
		return uuid.Nil, dom.NewTaggedError(dom.ErrExpired, nil)
	}

	if err := r.q.RevokeRefreshTokenByID(ctx, row.ID); err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return uuid.Nil, dom.NewTaggedError(dom.ErrTimeout, err)
		case errors.Is(err, sql.ErrNoRows):
			return uuid.Nil, dom.NewTaggedError(dom.ErrNoResults, err)
		default:
			return uuid.Nil, dom.NewTaggedError(dom.ErrInternal, err)
		}
	}

	return row.UserID, nil
}

func (r *PostgresRefreshTokenRepo) Revoke(
	ctx context.Context,
	rawToken string,
) error {
	hash := sha256.Sum256([]byte(rawToken))
	if err := r.q.RevokeRefreshTokenByHash(ctx, hex.EncodeToString(hash[:])); err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return dom.NewTaggedError(dom.ErrTimeout, err)
		case errors.Is(err, sql.ErrNoRows):
			return dom.NewTaggedError(dom.ErrNoResults, err)
		default:
			return dom.NewTaggedError(dom.ErrInternal, err)
		}
	}
	return nil
}

func (r *PostgresRefreshTokenRepo) RevokeAll(
	ctx context.Context,
	userID uuid.UUID,
) error {
	if err := r.q.RevokeAllUserTokens(ctx, userID); err != nil {
		switch {
		case errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
			return dom.NewTaggedError(dom.ErrTimeout, err)
		case errors.Is(err, sql.ErrNoRows):
			return dom.NewTaggedError(dom.ErrNoResults, err)
		default:
			return dom.NewTaggedError(dom.ErrInternal, err)
		}
	}
	return nil
}
