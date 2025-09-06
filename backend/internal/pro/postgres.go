package pro

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
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
	connStr := fmt.Sprintf(
		"%s?%s&%s&%s&%s&%s&%s&%s&%s",
		os.Getenv("POSTGRES_URL"),
		"sslmode=disable",
		"pool_max_conns=8",
		"pool_min_conns=2",
		"pool_min_idle_conns=2",
		"pool_max_conn_lifetime=30m",
		"pool_max_conn_lifetime_jitter=5m",
		"pool_max_conn_idle_time=15m",
		"pool_health_check_period=30s",
	)

	pgxCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("postgres parse config: %w", dom.ErrInternal)
	}

	conn, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("postgres connect: %w", dom.ErrTimeout)
		}
		return nil, fmt.Errorf("postgres connect: %w", dom.ErrUnavailable)
	}

	return &Postgres{
		Pool: conn,
	}, nil
}

func (p *Postgres) Runner() db.Querier { return db.New(p.Pool) }

func (p *Postgres) WithTx(ctx context.Context, fn func(q db.Querier) error) error {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("begin tx: %w", dom.ErrTimeout)
		}
		return fmt.Errorf("begin tx: %w", dom.ErrUnavailable)
	}

	q := db.New(tx)
	if err := fn(q); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	if err := tx.Commit(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("tx commit: %w", dom.ErrTimeout)
		}
		return fmt.Errorf("tx commit: %w", dom.ErrInternal)
	}
	return nil
}

func (p *Postgres) Name() string {
	return "db"
}

func (p *Postgres) Ping(ctx context.Context) error {
	if err := p.Pool.Ping(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("postgres ping: %w", dom.ErrTimeout)
		}
		return fmt.Errorf("postgres ping: %w", dom.ErrUnavailable)
	}
	return nil
}

func (p *Postgres) Close() error {
	if p.Pool != nil {
		p.Pool.Close()
	}
	return nil
}

func (p *Postgres) Begin(ctx context.Context) (dom.Tx, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("begin tx: %w", dom.ErrTimeout)
		}
		return nil, fmt.Errorf("begin tx: %w", dom.ErrUnavailable)
	}

	q := db.New(tx)
	return &PostgresTx{
		q:  q,
		tx: tx,
	}, nil
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
		return fmt.Errorf("unsupported repo type: %T, %w", r, dom.ErrInvalidInput)
	}
}

func (t *PostgresTx) Commit(ctx context.Context) error {
	if err := t.tx.Commit(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("tx commit: %w", dom.ErrTimeout)
		}
		return fmt.Errorf("tx commit: %w", dom.ErrInternal)
	}
	return nil
}

func (t *PostgresTx) Rollback(ctx context.Context) error {
	if err := t.tx.Rollback(ctx); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("tx rollback: %w", dom.ErrTimeout)
		}
		return fmt.Errorf("tx rollback: %w", dom.ErrInternal)
	}
	return nil
}

type PostgresUserRepo struct {
	q db.Querier
}

func NewPostgresUserRepo(q db.Querier) *PostgresUserRepo {
	return &PostgresUserRepo{q: q}
}

func (u *PostgresUserRepo) CreateUser(
	ctx context.Context,
	id uuid.UUID,
	email, hash string,
) (*dom.User, error) {
	row, err := u.q.CreateUser(ctx, db.CreateUserParams{
		ID:           id,
		Email:        email,
		PasswordHash: hash,
	})
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("create user: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("create user: %w", dom.ErrNoResults)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("create user: %w", dom.ErrConflict)
		}
		return nil, fmt.Errorf("create user: %w", dom.ErrInternal)
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

func (u *PostgresUserRepo) GetUserByID(
	ctx context.Context,
	id uuid.UUID,
) (*dom.User, error) {
	row, err := u.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("get user by id: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get user by id: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("get user by id: %w", dom.ErrInternal)
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

func (u *PostgresUserRepo) GetUserByEmail(
	ctx context.Context,
	email string,
) (*dom.User, error) {
	row, err := u.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("get user by email: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get user by email: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("get user by email: %w", dom.ErrInternal)
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
			return nil, fmt.Errorf("increment user messages: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("increment user messages: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("increment user messages: %w", dom.ErrInternal)
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
			return nil, fmt.Errorf("list users with backlog: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("list users with backlog: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("list users with backlog: %w", dom.ErrInternal)
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
			return fmt.Errorf("delete user: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("delete user: %w", dom.ErrNoResults)
		}
		return fmt.Errorf("delete user: %w", dom.ErrInternal)
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
			return nil, fmt.Errorf("create session: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("create session: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("create session: %w", dom.ErrInternal)
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
			return nil, fmt.Errorf("get session by id: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get session by id: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("get session by id: %w", dom.ErrInternal)
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
			return nil, fmt.Errorf("get sessions by user id: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get sessions by user id: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("get sessions by user id: %w", dom.ErrInternal)
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
			return nil, fmt.Errorf("update session by id: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("update session by id: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("update session by id: %w", dom.ErrInternal)
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
			return 0, fmt.Errorf("get max turn by id: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return 0, fmt.Errorf("get max turn by id: %w", dom.ErrNoResults)
		}
		return 0, fmt.Errorf("get max turn by id: %w", dom.ErrInternal)
	}

	return max, nil
}

func (s *PostgresSessionRepo) ListSessionsWithBacklog(ctx context.Context) ([]*dom.Session, error) {
	rows, err := s.q.ListSessionsWithBacklog(ctx)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("list sessions with backlog: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("list sessions with backlog: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("list sessions with backlog: %w", dom.ErrInternal)
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
			return fmt.Errorf("delete session: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("delete session: %w", dom.ErrNoResults)
		}
		return fmt.Errorf("delete session: %w", dom.ErrInternal)
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
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("create message: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("create message: %w", dom.ErrNoResults)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation
			return nil, fmt.Errorf("create message: %w", dom.ErrConflict)
		}
		return nil, fmt.Errorf("create message: %w", dom.ErrInternal)
	}

	return fromDbMessage(row), nil
}

func (m *PostgresMessageRepo) GetMessagesBySessionID(
	ctx context.Context,
	sessionID uuid.UUID,
) ([]dom.Message, error) {
	rows, err := m.q.GetMessagesBySessionIdOrdered(ctx, sessionID)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("get messages by session id: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get messages by session id: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("get messages by session id: %w", dom.ErrInternal)
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
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("get user messages by user id: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get user messages by user id: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("get user messages by user id: %w", dom.ErrInternal)
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
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("create memory: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("create memory: %w", dom.ErrNoResults)
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation
			return nil, fmt.Errorf("create memory: %w", dom.ErrConflict)
		}
		return nil, fmt.Errorf("create memory: %w", dom.ErrInternal)
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
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("upsert memory: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("upsert memory: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("upsert memory: %w", dom.ErrInternal)
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
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("get memories by user id: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("get memories by user id: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("get memories by user id: %w", dom.ErrInternal)
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
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("delete memory by user id key: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("delete memory by user id key: %w", dom.ErrNoResults)
		}
		return fmt.Errorf("delete memory by user id key: %w", dom.ErrInternal)
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
		return nil, fmt.Errorf("parallel semantic search: %w", dom.ErrInvalidInput)
	}
	chunksPerThread := max(1, topk/len(queries))

	results := make([][]dom.Chunk, len(queries))

	g, ctx := errgroup.WithContext(ctx)

	for i, query := range queries {
		i, query := i, query
		g.Go(func() error {
			if query.Vector == nil {
				return fmt.Errorf("parallel semantic search: %w", dom.ErrInvalidInput)
			}
			rows, err := r.SemanticSearch(ctx, query.VectorWithLabel, chunksPerThread)
			if err != nil {
				return fmt.Errorf("parallel semantic search: %w", err)
			}
			results[i] = rows
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
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
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("semantic search: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("semantic search: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("semantic search: %w", dom.ErrInternal)
	}

	returned := make([]dom.Chunk, 0, len(rows))
	for _, row := range rows {
		returned = append(returned,
			dom.Chunk{
				Document: dom.Document{
					ID:          int32(row.ID),
					Source:      row.Source,
					Content:     row.EmbeddedChunk,
					SurahNumber: *row.Surah,
					AyahNumber:  *row.Ayah,
				},
				ParentID: row.ParentID,
			},
		)
	}
	return returned, nil
}

func (r *PostgresSearcher) ParallelLexicalSearch(
	ctx context.Context,
	queries []dom.FullQueryContext,
	topk int,
) ([][]dom.Chunk, error) {
	if len(queries) == 0 {
		return nil, fmt.Errorf("parallel lexical search: %w", dom.ErrInvalidInput)
	}
	chunksPerThread := max(1, topk/len(queries))
	results := make([][]dom.Chunk, len(queries))

	g, ctx := errgroup.WithContext(ctx)

	for i, query := range queries {
		i, query := i, query
		g.Go(func() error {
			rows, err := r.LexicalSearch(ctx, query.QueryWithFilter, chunksPerThread)
			if err != nil {
				return fmt.Errorf("parallel lexical search: %w", err)
			}
			results[i] = rows
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
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
		return nil, fmt.Errorf("lexical search tokenize: %w", dom.ErrInvalidInput)
	}
	query.Query = tokenized

	params := db.LexicalSearchParams{
		NumberOfChunks: int64(topk),
		Query:          query.Query,
		ContentTypes:   query.OptionalContentTypes,
		Sources:        query.OptionalSources,
		Surahs:         query.OptionalSurahs,
		Ayahs:          query.OptionalAyahs,
	}
	rows, err := r.q.LexicalSearch(ctx, params)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("lexical search: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("lexical search: %w", dom.ErrNoResults)
		}
		return nil, fmt.Errorf("lexical search: %w", dom.ErrInternal)
	}

	returned := make([]dom.Chunk, 0, len(rows))
	for _, row := range rows {
		returned = append(returned,
			dom.Chunk{
				Document: dom.Document{
					ID:          int32(row.ID),
					Source:      row.Source,
					Content:     row.EmbeddedChunk,
					SurahNumber: *row.Surah,
					AyahNumber:  *row.Ayah,
				},
				ParentID: row.ParentID,
			},
		)
	}
	return returned, nil
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
		return "", fmt.Errorf("create refresh token rand: %w", dom.ErrInternal)
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
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return "", fmt.Errorf("create refresh token: %w", dom.ErrTimeout)
		}
		return "", fmt.Errorf("create refresh token: %w", dom.ErrInternal)
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
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return uuid.Nil, fmt.Errorf("validate refresh token: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, fmt.Errorf("validate refresh token: %w", dom.ErrNoResults)
		}
		return uuid.Nil, fmt.Errorf("validate refresh token: %w", dom.ErrInternal)
	}

	if row.RevokedAt != nil || time.Now().After(row.ExpiresAt) {
		return uuid.Nil, fmt.Errorf("validate refresh token: %w", dom.ErrExpired)
	}

	if err := r.q.RevokeRefreshTokenByID(ctx, row.ID); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return uuid.Nil, fmt.Errorf("revoke refresh token: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return uuid.Nil, fmt.Errorf("revoke refresh token: %w", dom.ErrNoResults)
		}
		return uuid.Nil, fmt.Errorf("revoke refresh token: %w", dom.ErrInternal)
	}

	return row.UserID, nil
}

func (r *PostgresRefreshTokenRepo) Revoke(
	ctx context.Context,
	rawToken string,
) error {
	hash := sha256.Sum256([]byte(rawToken))
	if err := r.q.RevokeRefreshTokenByHash(ctx, hex.EncodeToString(hash[:])); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("revoke refresh token: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("revoke refresh token: %w", dom.ErrNoResults)
		}
		return fmt.Errorf("revoke refresh token: %w", dom.ErrInternal)
	}
	return nil
}

func (r *PostgresRefreshTokenRepo) RevokeAll(
	ctx context.Context,
	userID uuid.UUID,
) error {
	if err := r.q.RevokeAllUserTokens(ctx, userID); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("revoke all refresh tokens: %w", dom.ErrTimeout)
		}
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("revoke all refresh tokens: %w", dom.ErrNoResults)
		}
		return fmt.Errorf("revoke all refresh tokens: %w", dom.ErrInternal)
	}
	return nil
}
