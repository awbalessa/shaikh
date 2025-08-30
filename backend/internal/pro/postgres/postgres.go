package pro

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/awbalessa/shaikh/backend/internal/dom"
	db "github.com/awbalessa/shaikh/backend/internal/pro/postgres/gen"
	"github.com/awbalessa/shaikh/backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"golang.org/x/sync/errgroup"
)

const (
	postgresSSLMode               string = "sslmode=disable"
	postgresMaxConns              string = "pool_max_conns=8"
	postgresMinConns              string = "pool_min_conns=2"
	postgresMinIdleConns          string = "pool_min_idle_conns=2"
	postgresMaxConnLifetime       string = "pool_max_conn_lifetime=30m"
	postgresMaxConnLifetimeJitter string = "pool_max_conn_lifetime_jitter=5m"
	postgresMaxConnIdleTime       string = "pool_max_conn_idle_time=15m"
	postgresPoolHealthCheckPeriod string = "pool_health_check_period=30s"
)

type Postgres struct {
	Pool *pgxpool.Pool
	Log  *slog.Logger
}

func NewPostgres(ctx context.Context, log *slog.Logger, env *config.Env) (*Postgres, error) {
	connStr := fmt.Sprintf(
		"%s?%s&%s&%s&%s&%s&%s&%s&%s",
		env.PostgresUrl,
		postgresSSLMode,
		postgresMaxConns,
		postgresMinConns,
		postgresMinIdleConns,
		postgresMaxConnLifetime,
		postgresMaxConnLifetimeJitter,
		postgresMaxConnIdleTime,
		postgresPoolHealthCheckPeriod,
	)

	pgxCfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		slog.With(
			slog.Any("err", err),
			slog.String("postgres_url", connStr),
		).ErrorContext(
			ctx,
			"failed to parse postgres url",
		)
		return nil, fmt.Errorf("failed to create postgres: %w", err)
	}

	conn, err := pgxpool.NewWithConfig(ctx, pgxCfg)
	if err != nil {
		slog.With(
			slog.Any("err", err),
			slog.String("postgres_url", connStr),
		).ErrorContext(
			ctx,
			"failed to create postgres",
		)
		return nil, fmt.Errorf("failed to create postgres: %w", err)
	}

	return &Postgres{
		Pool: conn,
		Log:  log,
	}, nil
}

func (p *Postgres) Runner() db.Querier { return db.New(p.Pool) }

func (p *Postgres) WithTx(ctx context.Context, fn func(q db.Querier) error) error {
	tx, err := p.Pool.Begin(ctx)
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
		return fmt.Errorf("unsupported repo type: %T", r)
	}
}

func (t *PostgresTx) Commit(ctx context.Context) error {
	return t.tx.Commit(ctx)
}

func (t *PostgresTx) Rollback(ctx context.Context) error {
	return t.tx.Rollback(ctx)
}

func (p *Postgres) Begin(ctx context.Context) (dom.Tx, error) {
	tx, err := p.Pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin tx: %w", err)
	}

	q := db.New(tx)
	return &PostgresTx{
		q:  q,
		tx: tx,
	}, nil
}

type PostgresMessageRepo struct {
	q   db.Querier
	log *slog.Logger
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
		Role:              toDbMessageRole[role],
		Model:             toDbLargeLanguageModel(meta.Model),
		Turn:              meta.Turn,
		TotalInputTokens:  toPgtypeInt4(meta.TotalInputTokens),
		TotalOutputTokens: toPgtypeInt4(meta.TotalOutputTokens),
		Content:           toPgtypeText(meta.Content),
		FunctionName:      toPgtypeText(meta.FunctionName),
		FunctionCall:      meta.FunctionCall,
		FunctionResponse:  meta.FunctionResponse,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create message: %w", err)
	}
	return fromDbMessage(row), nil
}

func (m *PostgresMessageRepo) GetMessagesBySessionIDOrdered(
	ctx context.Context,
	sessionID uuid.UUID,
) ([]dom.Message, error) {
	rows, err := m.q.GetMessagesBySessionIdOrdered(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get messages by session id ordered: %w", err)
	}

	final := make([]dom.Message, 0, len(rows))
	for _, r := range rows {
		final = append(final, fromDbMessage(r))
	}

	return final, nil
}

type PostgresSearcher struct {
	q   db.Querier
	log *slog.Logger
}

func NewPostgresSearcher(q db.Querier, log *slog.Logger) *PostgresSearcher {
	return &PostgresSearcher{q: q, log: log}
}

func (r *PostgresSearcher) ParallelSemanticSearch(
	ctx context.Context,
	queries []dom.FullQueryContext,
	topk int,
) ([][]dom.Chunk, error) {
	chunksPerThread := topk / len(queries)
	results := make([][]dom.Chunk, len(queries))

	g, ctx := errgroup.WithContext(ctx)
	r.log.With(
		slog.String("method", "ParallelSemanticSearch"),
		slog.Int("chunks_per_thread", chunksPerThread),
		slog.Int("num_of_threads", len(queries)),
	).DebugContext(ctx, "starting parallel semantic search...")

	start := time.Now()
	for i, query := range queries {
		i, query := i, query
		g.Go(func() error {
			if query.Vector == nil {
				return fmt.Errorf("missing vector for query: %q", query.Query)
			}
			rows, err := r.SemanticSearch(ctx, query.VectorWithLabel, chunksPerThread)
			if err != nil {
				return fmt.Errorf("parallel semantic search error: %w", err)
			}

			results[i] = rows
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	r.log.With(
		slog.String("method", "ParallelSemanticSearch"),
		slog.String("duration", time.Since(start).String()),
	).DebugContext(ctx, "parallel semantic search completed: returning...")

	return results, nil
}

func (r *PostgresSearcher) SemanticSearch(
	ctx context.Context,
	vector dom.VectorWithLabel,
	topk int,
) ([]dom.Chunk, error) {
	const method = "SemanticSearch"

	r.log.With(
		slog.String("method", method),
		slog.Int("number_of_chunks", topk),
		slog.String("content_type_labels", fmt.Sprint(vector.OptionalContentTypeLabels)),
		slog.String("source_labels", fmt.Sprint(vector.OptionalSourceLabels)),
		slog.String("surah_labels", fmt.Sprint(vector.OptionalSurahLabels)),
		slog.String("ayah_labels", fmt.Sprint(vector.OptionalAyahLabels)),
	).DebugContext(ctx, "running semantic search...")

	params := toSemSearchParams(vector, topk)

	start := time.Now()
	rows, err := r.q.SemanticSearch(
		ctx,
		params,
	)
	if err != nil {
		r.log.With("err", err).ErrorContext(
			ctx,
			"failed to run semantic search",
		)
		return nil, fmt.Errorf("failed to run semantic search: %w", err)
	}

	r.log.With(
		slog.String("method", method),
		slog.String("duration", time.Since(start).String()),
		slog.Int("result_count", len(rows)),
	).DebugContext(ctx, "ran semantic search: returning...")

	returned := make([]dom.Chunk, 0, len(rows))
	for _, row := range rows {
		returned = append(returned,
			dom.Chunk{
				Document: dom.Document{
					ID:          int32(row.ID),
					Source:      RagSourceToSource[row.Source],
					Content:     row.EmbeddedChunk,
					SurahNumber: RagSurahToSurahNumber[row.Surah.RagSurah],
					AyahNumber:  RagAyahToAyahNumber[row.Ayah.RagAyah],
				},
				ParentID: row.ParentID.Int32,
			},
		)
	}

	return returned, nil
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

func (r *PostgresSearcher) ParallelLexicalSearch(
	ctx context.Context,
	queries []dom.FullQueryContext,
	topk int,
) ([][]dom.Chunk, error) {
	chunksPerThread := topk / len(queries)
	results := make([][]dom.Chunk, len(queries))

	g, ctx := errgroup.WithContext(ctx)

	r.log.With(
		slog.String("method", "parallelLexicalSearch"),
		slog.Int("chunks_per_thread", chunksPerThread),
		slog.Int("num_of_threads", len(queries)),
	).DebugContext(ctx, "starting parallel lexical search...")

	start := time.Now()
	for i, query := range queries {
		i, query := i, query
		g.Go(func() error {
			rows, err := r.LexicalSearch(ctx, query.QueryWithFilter, chunksPerThread)
			if err != nil {
				return fmt.Errorf("parallel lexical search error: %w", err)
			}

			results[i] = rows
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	r.log.With(
		slog.String("method", "parallelLexicalSearch"),
		slog.String("duration", time.Since(start).String()),
	).DebugContext(ctx, "lexical search completed: returning...")
	return results, nil
}

func (r *PostgresSearcher) LexicalSearch(
	ctx context.Context,
	query dom.QueryWithFilter,
	topk int,
) ([]dom.Chunk, error) {
	const method = "LexicalSearch"

	r.log.With(
		slog.String("method", method),
		slog.Int("number_of_chunks", topk),
		slog.String("query", query.Query),
		slog.String("content_types", fmt.Sprint(query.OptionalContentTypes)),
		slog.String("sources", fmt.Sprint(query.OptionalSources)),
		slog.String("surahs", fmt.Sprint(query.OptionalSurahs)),
		slog.String("ayahs", fmt.Sprint(query.OptionalAyahs)),
	).DebugContext(ctx, "running lexical search...")

	tokenized, err := tokenizeQuery(query.Query)
	if err != nil {
		r.log.With("err", err).ErrorContext(
			ctx,
			"failed to tokenize query",
		)
		return nil, fmt.Errorf("failed to tokenize query: %w", err)
	}

	query.Query = tokenized

	params := toLexSearchParams(query, topk)
	start := time.Now()
	rows, err := r.q.LexicalSearch(
		ctx,
		params,
	)
	if err != nil {
		r.log.With("err", err).ErrorContext(
			ctx,
			"failed to run lexical search",
		)
		return nil, fmt.Errorf("failed to run lexical search: %w", err)
	}

	r.log.With(
		slog.String("tokenized_query", params.Query),
		slog.String("duration", time.Since(start).String()),
		slog.Int("result_count", len(rows)),
	).DebugContext(ctx, "ran lexical search: returning...")

	returned := make([]dom.Chunk, 0, len(rows))
	for _, row := range rows {
		returned = append(returned,
			dom.Chunk{
				Document: dom.Document{
					ID:          int32(row.ID),
					Source:      RagSourceToSource[row.Source],
					Content:     row.EmbeddedChunk,
					SurahNumber: RagSurahToSurahNumber[row.Surah.RagSurah],
					AyahNumber:  RagAyahToAyahNumber[row.Ayah.RagAyah],
				},
				ParentID: row.ParentID.Int32,
			},
		)
	}

	return returned, nil
}

func toLexSearchParams(qwf dom.QueryWithFilter, topk int) db.LexicalSearchParams {
	var (
		contentTypes []db.RagContentType = []db.RagContentType{}
		sources      []db.RagSource      = []db.RagSource{}
		surahs       []db.RagSurah       = []db.RagSurah{}
		ayahs        []db.RagAyah        = []db.RagAyah{}
	)
	for _, ct := range qwf.OptionalContentTypes {
		contentTypes = append(contentTypes, ContentTypeToRagContentType[ct])
	}
	for _, so := range qwf.OptionalSources {
		sources = append(sources, SourceToRagSource[so])
	}
	for _, sur := range qwf.OptionalSurahs {
		surahs = append(surahs, SurahNumberToRagSurah[sur])
	}
	for _, ay := range qwf.OptionalAyahs {
		ayahs = append(ayahs, AyahNumberToRagAyah[ay])
	}

	return db.LexicalSearchParams{
		NumberOfChunks: int64(topk),
		Query:          qwf.Query,
		ContentTypes:   contentTypes,
		Sources:        sources,
		Surahs:         surahs,
		Ayahs:          ayahs,
	}
}

func tokenizeQuery(query string) (string, error) {
	tokenized, err := utils.CleanAndFilterStopwords(query)
	if err != nil {
		return "", err
	}

	return tokenized, nil
}

func toDbLargeLanguageModel(llm dom.LargeLanguageModel) db.NullLargeLanguageModel {
	if llm == "" {
		return db.NullLargeLanguageModel{
			Valid: false,
		}
	}

	return db.NullLargeLanguageModel{
		Valid:              true,
		LargeLanguageModel: toDbLargeLanguageModelMap[llm],
	}
}

var toDbLargeLanguageModelMap = map[dom.LargeLanguageModel]db.LargeLanguageModel{
	dom.GeminiV2p5Flash:     db.LargeLanguageModelGemini25Flash,
	dom.GeminiV2p5FlashLite: db.LargeLanguageModelGemini25FlashLite,
}

var fromDbLargeLanguageModel = map[db.LargeLanguageModel]dom.LargeLanguageModel{
	db.LargeLanguageModelGemini25Flash:     dom.GeminiV2p5Flash,
	db.LargeLanguageModelGemini25FlashLite: dom.GeminiV2p5FlashLite,
}

var toDbMessageRole = map[dom.MessageRole]db.MessagesRole{
	dom.UserRole:     db.MessagesRoleUser,
	dom.ModelRole:    db.MessagesRoleModel,
	dom.FunctionRole: db.MessagesRoleFunction,
}

var fromDbMessageRole = map[db.MessagesRole]dom.MessageRole{
	db.MessagesRoleUser:     dom.UserRole,
	db.MessagesRoleModel:    dom.ModelRole,
	db.MessagesRoleFunction: dom.FunctionRole,
}

func toPgtypeInt4(in *int32) pgtype.Int4 {
	if in == nil {
		return pgtype.Int4{}
	}

	return pgtype.Int4{
		Valid: true,
		Int32: int32(*in),
	}
}
func fromPgtypeInt4(v pgtype.Int4) *int32 {
	if !v.Valid {
		return nil
	}
	val := v.Int32
	return &val
}

func toPgtypeText(str *string) pgtype.Text {
	if str == nil {
		return pgtype.Text{}
	}

	return pgtype.Text{
		Valid:  true,
		String: *str,
	}
}

func fromPgtypeText(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func fromDbMessage(row db.Message) dom.Message {
	meta := dom.MsgMeta{
		ID:                row.ID,
		SessionID:         row.SessionID,
		UserID:            row.UserID,
		Model:             fromDbLargeLanguageModel[row.Model.LargeLanguageModel],
		Turn:              row.Turn,
		TotalInputTokens:  fromPgtypeInt4(row.TotalInputTokens),
		TotalOutputTokens: fromPgtypeInt4(row.TotalOutputTokens),
		Content:           fromPgtypeText(row.Content),
		FunctionName:      fromPgtypeText(row.FunctionName),
		FunctionCall:      row.FunctionCall,
		FunctionResponse:  row.FunctionResponse,
	}

	switch row.Role {
	case db.MessagesRoleUser:
		return &dom.UserMessage{
			MsgMeta:    meta,
			MsgContent: *meta.Content,
		}
	case db.MessagesRoleFunction:
		return &dom.FunctionMessage{
			MsgMeta:          meta,
			FunctionName:     *meta.FunctionName,
			FunctionCall:     meta.FunctionCall,
			FunctionResponse: meta.FunctionResponse,
		}
	case db.MessagesRoleModel:
		return &dom.ModelMessage{
			MsgMeta:    meta,
			MsgContent: *meta.Content,
		}
	default:
		return nil
	}
}

var SourceToRagSource = map[dom.Source]db.RagSource{
	dom.SourceTafsirIbnKathir: db.RagSourceTafsirIbnKathir,
}

var RagSourceToSource = map[db.RagSource]dom.Source{
	db.RagSourceTafsirIbnKathir: dom.SourceTafsirIbnKathir,
}

var ContentTypeToRagContentType = map[dom.ContentType]db.RagContentType{
	dom.ContentTypeTafsir: db.RagContentTypeTafsir,
}

var RagContentTypeToContentType = map[db.RagContentType]dom.ContentType{
	db.RagContentTypeTafsir: dom.ContentTypeTafsir,
}

var SurahNumberToRagSurah = map[dom.SurahNumber]db.RagSurah{
	dom.SurahNumber1:   db.RagSurah1,
	dom.SurahNumber2:   db.RagSurah2,
	dom.SurahNumber3:   db.RagSurah3,
	dom.SurahNumber4:   db.RagSurah4,
	dom.SurahNumber5:   db.RagSurah5,
	dom.SurahNumber6:   db.RagSurah6,
	dom.SurahNumber7:   db.RagSurah7,
	dom.SurahNumber8:   db.RagSurah8,
	dom.SurahNumber9:   db.RagSurah9,
	dom.SurahNumber10:  db.RagSurah10,
	dom.SurahNumber11:  db.RagSurah11,
	dom.SurahNumber12:  db.RagSurah12,
	dom.SurahNumber13:  db.RagSurah13,
	dom.SurahNumber14:  db.RagSurah14,
	dom.SurahNumber15:  db.RagSurah15,
	dom.SurahNumber16:  db.RagSurah16,
	dom.SurahNumber17:  db.RagSurah17,
	dom.SurahNumber18:  db.RagSurah18,
	dom.SurahNumber19:  db.RagSurah19,
	dom.SurahNumber20:  db.RagSurah20,
	dom.SurahNumber21:  db.RagSurah21,
	dom.SurahNumber22:  db.RagSurah22,
	dom.SurahNumber23:  db.RagSurah23,
	dom.SurahNumber24:  db.RagSurah24,
	dom.SurahNumber25:  db.RagSurah25,
	dom.SurahNumber26:  db.RagSurah26,
	dom.SurahNumber27:  db.RagSurah27,
	dom.SurahNumber28:  db.RagSurah28,
	dom.SurahNumber29:  db.RagSurah29,
	dom.SurahNumber30:  db.RagSurah30,
	dom.SurahNumber31:  db.RagSurah31,
	dom.SurahNumber32:  db.RagSurah32,
	dom.SurahNumber33:  db.RagSurah33,
	dom.SurahNumber34:  db.RagSurah34,
	dom.SurahNumber35:  db.RagSurah35,
	dom.SurahNumber36:  db.RagSurah36,
	dom.SurahNumber37:  db.RagSurah37,
	dom.SurahNumber38:  db.RagSurah38,
	dom.SurahNumber39:  db.RagSurah39,
	dom.SurahNumber40:  db.RagSurah40,
	dom.SurahNumber41:  db.RagSurah41,
	dom.SurahNumber42:  db.RagSurah42,
	dom.SurahNumber43:  db.RagSurah43,
	dom.SurahNumber44:  db.RagSurah44,
	dom.SurahNumber45:  db.RagSurah45,
	dom.SurahNumber46:  db.RagSurah46,
	dom.SurahNumber47:  db.RagSurah47,
	dom.SurahNumber48:  db.RagSurah48,
	dom.SurahNumber49:  db.RagSurah49,
	dom.SurahNumber50:  db.RagSurah50,
	dom.SurahNumber51:  db.RagSurah51,
	dom.SurahNumber52:  db.RagSurah52,
	dom.SurahNumber53:  db.RagSurah53,
	dom.SurahNumber54:  db.RagSurah54,
	dom.SurahNumber55:  db.RagSurah55,
	dom.SurahNumber56:  db.RagSurah56,
	dom.SurahNumber57:  db.RagSurah57,
	dom.SurahNumber58:  db.RagSurah58,
	dom.SurahNumber59:  db.RagSurah59,
	dom.SurahNumber60:  db.RagSurah60,
	dom.SurahNumber61:  db.RagSurah61,
	dom.SurahNumber62:  db.RagSurah62,
	dom.SurahNumber63:  db.RagSurah63,
	dom.SurahNumber64:  db.RagSurah64,
	dom.SurahNumber65:  db.RagSurah65,
	dom.SurahNumber66:  db.RagSurah66,
	dom.SurahNumber67:  db.RagSurah67,
	dom.SurahNumber68:  db.RagSurah68,
	dom.SurahNumber69:  db.RagSurah69,
	dom.SurahNumber70:  db.RagSurah70,
	dom.SurahNumber71:  db.RagSurah71,
	dom.SurahNumber72:  db.RagSurah72,
	dom.SurahNumber73:  db.RagSurah73,
	dom.SurahNumber74:  db.RagSurah74,
	dom.SurahNumber75:  db.RagSurah75,
	dom.SurahNumber76:  db.RagSurah76,
	dom.SurahNumber77:  db.RagSurah77,
	dom.SurahNumber78:  db.RagSurah78,
	dom.SurahNumber79:  db.RagSurah79,
	dom.SurahNumber80:  db.RagSurah80,
	dom.SurahNumber81:  db.RagSurah81,
	dom.SurahNumber82:  db.RagSurah82,
	dom.SurahNumber83:  db.RagSurah83,
	dom.SurahNumber84:  db.RagSurah84,
	dom.SurahNumber85:  db.RagSurah85,
	dom.SurahNumber86:  db.RagSurah86,
	dom.SurahNumber87:  db.RagSurah87,
	dom.SurahNumber88:  db.RagSurah88,
	dom.SurahNumber89:  db.RagSurah89,
	dom.SurahNumber90:  db.RagSurah90,
	dom.SurahNumber91:  db.RagSurah91,
	dom.SurahNumber92:  db.RagSurah92,
	dom.SurahNumber93:  db.RagSurah93,
	dom.SurahNumber94:  db.RagSurah94,
	dom.SurahNumber95:  db.RagSurah95,
	dom.SurahNumber96:  db.RagSurah96,
	dom.SurahNumber97:  db.RagSurah97,
	dom.SurahNumber98:  db.RagSurah98,
	dom.SurahNumber99:  db.RagSurah99,
	dom.SurahNumber100: db.RagSurah100,
	dom.SurahNumber101: db.RagSurah101,
	dom.SurahNumber102: db.RagSurah102,
	dom.SurahNumber103: db.RagSurah103,
	dom.SurahNumber104: db.RagSurah104,
	dom.SurahNumber105: db.RagSurah105,
	dom.SurahNumber106: db.RagSurah106,
	dom.SurahNumber107: db.RagSurah107,
	dom.SurahNumber108: db.RagSurah108,
	dom.SurahNumber109: db.RagSurah109,
	dom.SurahNumber110: db.RagSurah110,
	dom.SurahNumber111: db.RagSurah111,
	dom.SurahNumber112: db.RagSurah112,
	dom.SurahNumber113: db.RagSurah113,
	dom.SurahNumber114: db.RagSurah114,
}

var RagSurahToSurahNumber = map[db.RagSurah]dom.SurahNumber{
	db.RagSurah1:   dom.SurahNumber1,
	db.RagSurah2:   dom.SurahNumber2,
	db.RagSurah3:   dom.SurahNumber3,
	db.RagSurah4:   dom.SurahNumber4,
	db.RagSurah5:   dom.SurahNumber5,
	db.RagSurah6:   dom.SurahNumber6,
	db.RagSurah7:   dom.SurahNumber7,
	db.RagSurah8:   dom.SurahNumber8,
	db.RagSurah9:   dom.SurahNumber9,
	db.RagSurah10:  dom.SurahNumber10,
	db.RagSurah11:  dom.SurahNumber11,
	db.RagSurah12:  dom.SurahNumber12,
	db.RagSurah13:  dom.SurahNumber13,
	db.RagSurah14:  dom.SurahNumber14,
	db.RagSurah15:  dom.SurahNumber15,
	db.RagSurah16:  dom.SurahNumber16,
	db.RagSurah17:  dom.SurahNumber17,
	db.RagSurah18:  dom.SurahNumber18,
	db.RagSurah19:  dom.SurahNumber19,
	db.RagSurah20:  dom.SurahNumber20,
	db.RagSurah21:  dom.SurahNumber21,
	db.RagSurah22:  dom.SurahNumber22,
	db.RagSurah23:  dom.SurahNumber23,
	db.RagSurah24:  dom.SurahNumber24,
	db.RagSurah25:  dom.SurahNumber25,
	db.RagSurah26:  dom.SurahNumber26,
	db.RagSurah27:  dom.SurahNumber27,
	db.RagSurah28:  dom.SurahNumber28,
	db.RagSurah29:  dom.SurahNumber29,
	db.RagSurah30:  dom.SurahNumber30,
	db.RagSurah31:  dom.SurahNumber31,
	db.RagSurah32:  dom.SurahNumber32,
	db.RagSurah33:  dom.SurahNumber33,
	db.RagSurah34:  dom.SurahNumber34,
	db.RagSurah35:  dom.SurahNumber35,
	db.RagSurah36:  dom.SurahNumber36,
	db.RagSurah37:  dom.SurahNumber37,
	db.RagSurah38:  dom.SurahNumber38,
	db.RagSurah39:  dom.SurahNumber39,
	db.RagSurah40:  dom.SurahNumber40,
	db.RagSurah41:  dom.SurahNumber41,
	db.RagSurah42:  dom.SurahNumber42,
	db.RagSurah43:  dom.SurahNumber43,
	db.RagSurah44:  dom.SurahNumber44,
	db.RagSurah45:  dom.SurahNumber45,
	db.RagSurah46:  dom.SurahNumber46,
	db.RagSurah47:  dom.SurahNumber47,
	db.RagSurah48:  dom.SurahNumber48,
	db.RagSurah49:  dom.SurahNumber49,
	db.RagSurah50:  dom.SurahNumber50,
	db.RagSurah51:  dom.SurahNumber51,
	db.RagSurah52:  dom.SurahNumber52,
	db.RagSurah53:  dom.SurahNumber53,
	db.RagSurah54:  dom.SurahNumber54,
	db.RagSurah55:  dom.SurahNumber55,
	db.RagSurah56:  dom.SurahNumber56,
	db.RagSurah57:  dom.SurahNumber57,
	db.RagSurah58:  dom.SurahNumber58,
	db.RagSurah59:  dom.SurahNumber59,
	db.RagSurah60:  dom.SurahNumber60,
	db.RagSurah61:  dom.SurahNumber61,
	db.RagSurah62:  dom.SurahNumber62,
	db.RagSurah63:  dom.SurahNumber63,
	db.RagSurah64:  dom.SurahNumber64,
	db.RagSurah65:  dom.SurahNumber65,
	db.RagSurah66:  dom.SurahNumber66,
	db.RagSurah67:  dom.SurahNumber67,
	db.RagSurah68:  dom.SurahNumber68,
	db.RagSurah69:  dom.SurahNumber69,
	db.RagSurah70:  dom.SurahNumber70,
	db.RagSurah71:  dom.SurahNumber71,
	db.RagSurah72:  dom.SurahNumber72,
	db.RagSurah73:  dom.SurahNumber73,
	db.RagSurah74:  dom.SurahNumber74,
	db.RagSurah75:  dom.SurahNumber75,
	db.RagSurah76:  dom.SurahNumber76,
	db.RagSurah77:  dom.SurahNumber77,
	db.RagSurah78:  dom.SurahNumber78,
	db.RagSurah79:  dom.SurahNumber79,
	db.RagSurah80:  dom.SurahNumber80,
	db.RagSurah81:  dom.SurahNumber81,
	db.RagSurah82:  dom.SurahNumber82,
	db.RagSurah83:  dom.SurahNumber83,
	db.RagSurah84:  dom.SurahNumber84,
	db.RagSurah85:  dom.SurahNumber85,
	db.RagSurah86:  dom.SurahNumber86,
	db.RagSurah87:  dom.SurahNumber87,
	db.RagSurah88:  dom.SurahNumber88,
	db.RagSurah89:  dom.SurahNumber89,
	db.RagSurah90:  dom.SurahNumber90,
	db.RagSurah91:  dom.SurahNumber91,
	db.RagSurah92:  dom.SurahNumber92,
	db.RagSurah93:  dom.SurahNumber93,
	db.RagSurah94:  dom.SurahNumber94,
	db.RagSurah95:  dom.SurahNumber95,
	db.RagSurah96:  dom.SurahNumber96,
	db.RagSurah97:  dom.SurahNumber97,
	db.RagSurah98:  dom.SurahNumber98,
	db.RagSurah99:  dom.SurahNumber99,
	db.RagSurah100: dom.SurahNumber100,
	db.RagSurah101: dom.SurahNumber101,
	db.RagSurah102: dom.SurahNumber102,
	db.RagSurah103: dom.SurahNumber103,
	db.RagSurah104: dom.SurahNumber104,
	db.RagSurah105: dom.SurahNumber105,
	db.RagSurah106: dom.SurahNumber106,
	db.RagSurah107: dom.SurahNumber107,
	db.RagSurah108: dom.SurahNumber108,
	db.RagSurah109: dom.SurahNumber109,
	db.RagSurah110: dom.SurahNumber110,
	db.RagSurah111: dom.SurahNumber111,
	db.RagSurah112: dom.SurahNumber112,
	db.RagSurah113: dom.SurahNumber113,
	db.RagSurah114: dom.SurahNumber114,
}

var AyahNumberToRagAyah = map[dom.AyahNumber]db.RagAyah{
	dom.AyahNumber1:   db.RagAyah1,
	dom.AyahNumber2:   db.RagAyah2,
	dom.AyahNumber3:   db.RagAyah3,
	dom.AyahNumber4:   db.RagAyah4,
	dom.AyahNumber5:   db.RagAyah5,
	dom.AyahNumber6:   db.RagAyah6,
	dom.AyahNumber7:   db.RagAyah7,
	dom.AyahNumber8:   db.RagAyah8,
	dom.AyahNumber9:   db.RagAyah9,
	dom.AyahNumber10:  db.RagAyah10,
	dom.AyahNumber11:  db.RagAyah11,
	dom.AyahNumber12:  db.RagAyah12,
	dom.AyahNumber13:  db.RagAyah13,
	dom.AyahNumber14:  db.RagAyah14,
	dom.AyahNumber15:  db.RagAyah15,
	dom.AyahNumber16:  db.RagAyah16,
	dom.AyahNumber17:  db.RagAyah17,
	dom.AyahNumber18:  db.RagAyah18,
	dom.AyahNumber19:  db.RagAyah19,
	dom.AyahNumber20:  db.RagAyah20,
	dom.AyahNumber21:  db.RagAyah21,
	dom.AyahNumber22:  db.RagAyah22,
	dom.AyahNumber23:  db.RagAyah23,
	dom.AyahNumber24:  db.RagAyah24,
	dom.AyahNumber25:  db.RagAyah25,
	dom.AyahNumber26:  db.RagAyah26,
	dom.AyahNumber27:  db.RagAyah27,
	dom.AyahNumber28:  db.RagAyah28,
	dom.AyahNumber29:  db.RagAyah29,
	dom.AyahNumber30:  db.RagAyah30,
	dom.AyahNumber31:  db.RagAyah31,
	dom.AyahNumber32:  db.RagAyah32,
	dom.AyahNumber33:  db.RagAyah33,
	dom.AyahNumber34:  db.RagAyah34,
	dom.AyahNumber35:  db.RagAyah35,
	dom.AyahNumber36:  db.RagAyah36,
	dom.AyahNumber37:  db.RagAyah37,
	dom.AyahNumber38:  db.RagAyah38,
	dom.AyahNumber39:  db.RagAyah39,
	dom.AyahNumber40:  db.RagAyah40,
	dom.AyahNumber41:  db.RagAyah41,
	dom.AyahNumber42:  db.RagAyah42,
	dom.AyahNumber43:  db.RagAyah43,
	dom.AyahNumber44:  db.RagAyah44,
	dom.AyahNumber45:  db.RagAyah45,
	dom.AyahNumber46:  db.RagAyah46,
	dom.AyahNumber47:  db.RagAyah47,
	dom.AyahNumber48:  db.RagAyah48,
	dom.AyahNumber49:  db.RagAyah49,
	dom.AyahNumber50:  db.RagAyah50,
	dom.AyahNumber51:  db.RagAyah51,
	dom.AyahNumber52:  db.RagAyah52,
	dom.AyahNumber53:  db.RagAyah53,
	dom.AyahNumber54:  db.RagAyah54,
	dom.AyahNumber55:  db.RagAyah55,
	dom.AyahNumber56:  db.RagAyah56,
	dom.AyahNumber57:  db.RagAyah57,
	dom.AyahNumber58:  db.RagAyah58,
	dom.AyahNumber59:  db.RagAyah59,
	dom.AyahNumber60:  db.RagAyah60,
	dom.AyahNumber61:  db.RagAyah61,
	dom.AyahNumber62:  db.RagAyah62,
	dom.AyahNumber63:  db.RagAyah63,
	dom.AyahNumber64:  db.RagAyah64,
	dom.AyahNumber65:  db.RagAyah65,
	dom.AyahNumber66:  db.RagAyah66,
	dom.AyahNumber67:  db.RagAyah67,
	dom.AyahNumber68:  db.RagAyah68,
	dom.AyahNumber69:  db.RagAyah69,
	dom.AyahNumber70:  db.RagAyah70,
	dom.AyahNumber71:  db.RagAyah71,
	dom.AyahNumber72:  db.RagAyah72,
	dom.AyahNumber73:  db.RagAyah73,
	dom.AyahNumber74:  db.RagAyah74,
	dom.AyahNumber75:  db.RagAyah75,
	dom.AyahNumber76:  db.RagAyah76,
	dom.AyahNumber77:  db.RagAyah77,
	dom.AyahNumber78:  db.RagAyah78,
	dom.AyahNumber79:  db.RagAyah79,
	dom.AyahNumber80:  db.RagAyah80,
	dom.AyahNumber81:  db.RagAyah81,
	dom.AyahNumber82:  db.RagAyah82,
	dom.AyahNumber83:  db.RagAyah83,
	dom.AyahNumber84:  db.RagAyah84,
	dom.AyahNumber85:  db.RagAyah85,
	dom.AyahNumber86:  db.RagAyah86,
	dom.AyahNumber87:  db.RagAyah87,
	dom.AyahNumber88:  db.RagAyah88,
	dom.AyahNumber89:  db.RagAyah89,
	dom.AyahNumber90:  db.RagAyah90,
	dom.AyahNumber91:  db.RagAyah91,
	dom.AyahNumber92:  db.RagAyah92,
	dom.AyahNumber93:  db.RagAyah93,
	dom.AyahNumber94:  db.RagAyah94,
	dom.AyahNumber95:  db.RagAyah95,
	dom.AyahNumber96:  db.RagAyah96,
	dom.AyahNumber97:  db.RagAyah97,
	dom.AyahNumber98:  db.RagAyah98,
	dom.AyahNumber99:  db.RagAyah99,
	dom.AyahNumber100: db.RagAyah100,
	dom.AyahNumber101: db.RagAyah101,
	dom.AyahNumber102: db.RagAyah102,
	dom.AyahNumber103: db.RagAyah103,
	dom.AyahNumber104: db.RagAyah104,
	dom.AyahNumber105: db.RagAyah105,
	dom.AyahNumber106: db.RagAyah106,
	dom.AyahNumber107: db.RagAyah107,
	dom.AyahNumber108: db.RagAyah108,
	dom.AyahNumber109: db.RagAyah109,
	dom.AyahNumber110: db.RagAyah110,
	dom.AyahNumber111: db.RagAyah111,
	dom.AyahNumber112: db.RagAyah112,
	dom.AyahNumber113: db.RagAyah113,
	dom.AyahNumber114: db.RagAyah114,
	dom.AyahNumber115: db.RagAyah115,
	dom.AyahNumber116: db.RagAyah116,
	dom.AyahNumber117: db.RagAyah117,
	dom.AyahNumber118: db.RagAyah118,
	dom.AyahNumber119: db.RagAyah119,
	dom.AyahNumber120: db.RagAyah120,
	dom.AyahNumber121: db.RagAyah121,
	dom.AyahNumber122: db.RagAyah122,
	dom.AyahNumber123: db.RagAyah123,
	dom.AyahNumber124: db.RagAyah124,
	dom.AyahNumber125: db.RagAyah125,
	dom.AyahNumber126: db.RagAyah126,
	dom.AyahNumber127: db.RagAyah127,
	dom.AyahNumber128: db.RagAyah128,
	dom.AyahNumber129: db.RagAyah129,
	dom.AyahNumber130: db.RagAyah130,
	dom.AyahNumber131: db.RagAyah131,
	dom.AyahNumber132: db.RagAyah132,
	dom.AyahNumber133: db.RagAyah133,
	dom.AyahNumber134: db.RagAyah134,
	dom.AyahNumber135: db.RagAyah135,
	dom.AyahNumber136: db.RagAyah136,
	dom.AyahNumber137: db.RagAyah137,
	dom.AyahNumber138: db.RagAyah138,
	dom.AyahNumber139: db.RagAyah139,
	dom.AyahNumber140: db.RagAyah140,
	dom.AyahNumber141: db.RagAyah141,
	dom.AyahNumber142: db.RagAyah142,
	dom.AyahNumber143: db.RagAyah143,
	dom.AyahNumber144: db.RagAyah144,
	dom.AyahNumber145: db.RagAyah145,
	dom.AyahNumber146: db.RagAyah146,
	dom.AyahNumber147: db.RagAyah147,
	dom.AyahNumber148: db.RagAyah148,
	dom.AyahNumber149: db.RagAyah149,
	dom.AyahNumber150: db.RagAyah150,
	dom.AyahNumber151: db.RagAyah151,
	dom.AyahNumber152: db.RagAyah152,
	dom.AyahNumber153: db.RagAyah153,
	dom.AyahNumber154: db.RagAyah154,
	dom.AyahNumber155: db.RagAyah155,
	dom.AyahNumber156: db.RagAyah156,
	dom.AyahNumber157: db.RagAyah157,
	dom.AyahNumber158: db.RagAyah158,
	dom.AyahNumber159: db.RagAyah159,
	dom.AyahNumber160: db.RagAyah160,
	dom.AyahNumber161: db.RagAyah161,
	dom.AyahNumber162: db.RagAyah162,
	dom.AyahNumber163: db.RagAyah163,
	dom.AyahNumber164: db.RagAyah164,
	dom.AyahNumber165: db.RagAyah165,
	dom.AyahNumber166: db.RagAyah166,
	dom.AyahNumber167: db.RagAyah167,
	dom.AyahNumber168: db.RagAyah168,
	dom.AyahNumber169: db.RagAyah169,
	dom.AyahNumber170: db.RagAyah170,
	dom.AyahNumber171: db.RagAyah171,
	dom.AyahNumber172: db.RagAyah172,
	dom.AyahNumber173: db.RagAyah173,
	dom.AyahNumber174: db.RagAyah174,
	dom.AyahNumber175: db.RagAyah175,
	dom.AyahNumber176: db.RagAyah176,
	dom.AyahNumber177: db.RagAyah177,
	dom.AyahNumber178: db.RagAyah178,
	dom.AyahNumber179: db.RagAyah179,
	dom.AyahNumber180: db.RagAyah180,
	dom.AyahNumber181: db.RagAyah181,
	dom.AyahNumber182: db.RagAyah182,
	dom.AyahNumber183: db.RagAyah183,
	dom.AyahNumber184: db.RagAyah184,
	dom.AyahNumber185: db.RagAyah185,
	dom.AyahNumber186: db.RagAyah186,
	dom.AyahNumber187: db.RagAyah187,
	dom.AyahNumber188: db.RagAyah188,
	dom.AyahNumber189: db.RagAyah189,
	dom.AyahNumber190: db.RagAyah190,
	dom.AyahNumber191: db.RagAyah191,
	dom.AyahNumber192: db.RagAyah192,
	dom.AyahNumber193: db.RagAyah193,
	dom.AyahNumber194: db.RagAyah194,
	dom.AyahNumber195: db.RagAyah195,
	dom.AyahNumber196: db.RagAyah196,
	dom.AyahNumber197: db.RagAyah197,
	dom.AyahNumber198: db.RagAyah198,
	dom.AyahNumber199: db.RagAyah199,
	dom.AyahNumber200: db.RagAyah200,
	dom.AyahNumber201: db.RagAyah201,
	dom.AyahNumber202: db.RagAyah202,
	dom.AyahNumber203: db.RagAyah203,
	dom.AyahNumber204: db.RagAyah204,
	dom.AyahNumber205: db.RagAyah205,
	dom.AyahNumber206: db.RagAyah206,
	dom.AyahNumber207: db.RagAyah207,
	dom.AyahNumber208: db.RagAyah208,
	dom.AyahNumber209: db.RagAyah209,
	dom.AyahNumber210: db.RagAyah210,
	dom.AyahNumber211: db.RagAyah211,
	dom.AyahNumber212: db.RagAyah212,
	dom.AyahNumber213: db.RagAyah213,
	dom.AyahNumber214: db.RagAyah214,
	dom.AyahNumber215: db.RagAyah215,
	dom.AyahNumber216: db.RagAyah216,
	dom.AyahNumber217: db.RagAyah217,
	dom.AyahNumber218: db.RagAyah218,
	dom.AyahNumber219: db.RagAyah219,
	dom.AyahNumber220: db.RagAyah220,
	dom.AyahNumber221: db.RagAyah221,
	dom.AyahNumber222: db.RagAyah222,
	dom.AyahNumber223: db.RagAyah223,
	dom.AyahNumber224: db.RagAyah224,
	dom.AyahNumber225: db.RagAyah225,
	dom.AyahNumber226: db.RagAyah226,
	dom.AyahNumber227: db.RagAyah227,
	dom.AyahNumber228: db.RagAyah228,
	dom.AyahNumber229: db.RagAyah229,
	dom.AyahNumber230: db.RagAyah230,
	dom.AyahNumber231: db.RagAyah231,
	dom.AyahNumber232: db.RagAyah232,
	dom.AyahNumber233: db.RagAyah233,
	dom.AyahNumber234: db.RagAyah234,
	dom.AyahNumber235: db.RagAyah235,
	dom.AyahNumber236: db.RagAyah236,
	dom.AyahNumber237: db.RagAyah237,
	dom.AyahNumber238: db.RagAyah238,
	dom.AyahNumber239: db.RagAyah239,
	dom.AyahNumber240: db.RagAyah240,
	dom.AyahNumber241: db.RagAyah241,
	dom.AyahNumber242: db.RagAyah242,
	dom.AyahNumber243: db.RagAyah243,
	dom.AyahNumber244: db.RagAyah244,
	dom.AyahNumber245: db.RagAyah245,
	dom.AyahNumber246: db.RagAyah246,
	dom.AyahNumber247: db.RagAyah247,
	dom.AyahNumber248: db.RagAyah248,
	dom.AyahNumber249: db.RagAyah249,
	dom.AyahNumber250: db.RagAyah250,
	dom.AyahNumber251: db.RagAyah251,
	dom.AyahNumber252: db.RagAyah252,
	dom.AyahNumber253: db.RagAyah253,
	dom.AyahNumber254: db.RagAyah254,
	dom.AyahNumber255: db.RagAyah255,
	dom.AyahNumber256: db.RagAyah256,
	dom.AyahNumber257: db.RagAyah257,
	dom.AyahNumber258: db.RagAyah258,
	dom.AyahNumber259: db.RagAyah259,
	dom.AyahNumber260: db.RagAyah260,
	dom.AyahNumber261: db.RagAyah261,
	dom.AyahNumber262: db.RagAyah262,
	dom.AyahNumber263: db.RagAyah263,
	dom.AyahNumber264: db.RagAyah264,
	dom.AyahNumber265: db.RagAyah265,
	dom.AyahNumber266: db.RagAyah266,
	dom.AyahNumber267: db.RagAyah267,
	dom.AyahNumber268: db.RagAyah268,
	dom.AyahNumber269: db.RagAyah269,
	dom.AyahNumber270: db.RagAyah270,
	dom.AyahNumber271: db.RagAyah271,
	dom.AyahNumber272: db.RagAyah272,
	dom.AyahNumber273: db.RagAyah273,
	dom.AyahNumber274: db.RagAyah274,
	dom.AyahNumber275: db.RagAyah275,
	dom.AyahNumber276: db.RagAyah276,
	dom.AyahNumber277: db.RagAyah277,
	dom.AyahNumber278: db.RagAyah278,
	dom.AyahNumber279: db.RagAyah279,
	dom.AyahNumber280: db.RagAyah280,
	dom.AyahNumber281: db.RagAyah281,
	dom.AyahNumber282: db.RagAyah282,
	dom.AyahNumber283: db.RagAyah283,
	dom.AyahNumber284: db.RagAyah284,
	dom.AyahNumber285: db.RagAyah285,
	dom.AyahNumber286: db.RagAyah286,
}

var RagAyahToAyahNumber = map[db.RagAyah]dom.AyahNumber{
	db.RagAyah1:   dom.AyahNumber1,
	db.RagAyah2:   dom.AyahNumber2,
	db.RagAyah3:   dom.AyahNumber3,
	db.RagAyah4:   dom.AyahNumber4,
	db.RagAyah5:   dom.AyahNumber5,
	db.RagAyah6:   dom.AyahNumber6,
	db.RagAyah7:   dom.AyahNumber7,
	db.RagAyah8:   dom.AyahNumber8,
	db.RagAyah9:   dom.AyahNumber9,
	db.RagAyah10:  dom.AyahNumber10,
	db.RagAyah11:  dom.AyahNumber11,
	db.RagAyah12:  dom.AyahNumber12,
	db.RagAyah13:  dom.AyahNumber13,
	db.RagAyah14:  dom.AyahNumber14,
	db.RagAyah15:  dom.AyahNumber15,
	db.RagAyah16:  dom.AyahNumber16,
	db.RagAyah17:  dom.AyahNumber17,
	db.RagAyah18:  dom.AyahNumber18,
	db.RagAyah19:  dom.AyahNumber19,
	db.RagAyah20:  dom.AyahNumber20,
	db.RagAyah21:  dom.AyahNumber21,
	db.RagAyah22:  dom.AyahNumber22,
	db.RagAyah23:  dom.AyahNumber23,
	db.RagAyah24:  dom.AyahNumber24,
	db.RagAyah25:  dom.AyahNumber25,
	db.RagAyah26:  dom.AyahNumber26,
	db.RagAyah27:  dom.AyahNumber27,
	db.RagAyah28:  dom.AyahNumber28,
	db.RagAyah29:  dom.AyahNumber29,
	db.RagAyah30:  dom.AyahNumber30,
	db.RagAyah31:  dom.AyahNumber31,
	db.RagAyah32:  dom.AyahNumber32,
	db.RagAyah33:  dom.AyahNumber33,
	db.RagAyah34:  dom.AyahNumber34,
	db.RagAyah35:  dom.AyahNumber35,
	db.RagAyah36:  dom.AyahNumber36,
	db.RagAyah37:  dom.AyahNumber37,
	db.RagAyah38:  dom.AyahNumber38,
	db.RagAyah39:  dom.AyahNumber39,
	db.RagAyah40:  dom.AyahNumber40,
	db.RagAyah41:  dom.AyahNumber41,
	db.RagAyah42:  dom.AyahNumber42,
	db.RagAyah43:  dom.AyahNumber43,
	db.RagAyah44:  dom.AyahNumber44,
	db.RagAyah45:  dom.AyahNumber45,
	db.RagAyah46:  dom.AyahNumber46,
	db.RagAyah47:  dom.AyahNumber47,
	db.RagAyah48:  dom.AyahNumber48,
	db.RagAyah49:  dom.AyahNumber49,
	db.RagAyah50:  dom.AyahNumber50,
	db.RagAyah51:  dom.AyahNumber51,
	db.RagAyah52:  dom.AyahNumber52,
	db.RagAyah53:  dom.AyahNumber53,
	db.RagAyah54:  dom.AyahNumber54,
	db.RagAyah55:  dom.AyahNumber55,
	db.RagAyah56:  dom.AyahNumber56,
	db.RagAyah57:  dom.AyahNumber57,
	db.RagAyah58:  dom.AyahNumber58,
	db.RagAyah59:  dom.AyahNumber59,
	db.RagAyah60:  dom.AyahNumber60,
	db.RagAyah61:  dom.AyahNumber61,
	db.RagAyah62:  dom.AyahNumber62,
	db.RagAyah63:  dom.AyahNumber63,
	db.RagAyah64:  dom.AyahNumber64,
	db.RagAyah65:  dom.AyahNumber65,
	db.RagAyah66:  dom.AyahNumber66,
	db.RagAyah67:  dom.AyahNumber67,
	db.RagAyah68:  dom.AyahNumber68,
	db.RagAyah69:  dom.AyahNumber69,
	db.RagAyah70:  dom.AyahNumber70,
	db.RagAyah71:  dom.AyahNumber71,
	db.RagAyah72:  dom.AyahNumber72,
	db.RagAyah73:  dom.AyahNumber73,
	db.RagAyah74:  dom.AyahNumber74,
	db.RagAyah75:  dom.AyahNumber75,
	db.RagAyah76:  dom.AyahNumber76,
	db.RagAyah77:  dom.AyahNumber77,
	db.RagAyah78:  dom.AyahNumber78,
	db.RagAyah79:  dom.AyahNumber79,
	db.RagAyah80:  dom.AyahNumber80,
	db.RagAyah81:  dom.AyahNumber81,
	db.RagAyah82:  dom.AyahNumber82,
	db.RagAyah83:  dom.AyahNumber83,
	db.RagAyah84:  dom.AyahNumber84,
	db.RagAyah85:  dom.AyahNumber85,
	db.RagAyah86:  dom.AyahNumber86,
	db.RagAyah87:  dom.AyahNumber87,
	db.RagAyah88:  dom.AyahNumber88,
	db.RagAyah89:  dom.AyahNumber89,
	db.RagAyah90:  dom.AyahNumber90,
	db.RagAyah91:  dom.AyahNumber91,
	db.RagAyah92:  dom.AyahNumber92,
	db.RagAyah93:  dom.AyahNumber93,
	db.RagAyah94:  dom.AyahNumber94,
	db.RagAyah95:  dom.AyahNumber95,
	db.RagAyah96:  dom.AyahNumber96,
	db.RagAyah97:  dom.AyahNumber97,
	db.RagAyah98:  dom.AyahNumber98,
	db.RagAyah99:  dom.AyahNumber99,
	db.RagAyah100: dom.AyahNumber100,
	db.RagAyah101: dom.AyahNumber101,
	db.RagAyah102: dom.AyahNumber102,
	db.RagAyah103: dom.AyahNumber103,
	db.RagAyah104: dom.AyahNumber104,
	db.RagAyah105: dom.AyahNumber105,
	db.RagAyah106: dom.AyahNumber106,
	db.RagAyah107: dom.AyahNumber107,
	db.RagAyah108: dom.AyahNumber108,
	db.RagAyah109: dom.AyahNumber109,
	db.RagAyah110: dom.AyahNumber110,
	db.RagAyah111: dom.AyahNumber111,
	db.RagAyah112: dom.AyahNumber112,
	db.RagAyah113: dom.AyahNumber113,
	db.RagAyah114: dom.AyahNumber114,
	db.RagAyah115: dom.AyahNumber115,
	db.RagAyah116: dom.AyahNumber116,
	db.RagAyah117: dom.AyahNumber117,
	db.RagAyah118: dom.AyahNumber118,
	db.RagAyah119: dom.AyahNumber119,
	db.RagAyah120: dom.AyahNumber120,
	db.RagAyah121: dom.AyahNumber121,
	db.RagAyah122: dom.AyahNumber122,
	db.RagAyah123: dom.AyahNumber123,
	db.RagAyah124: dom.AyahNumber124,
	db.RagAyah125: dom.AyahNumber125,
	db.RagAyah126: dom.AyahNumber126,
	db.RagAyah127: dom.AyahNumber127,
	db.RagAyah128: dom.AyahNumber128,
	db.RagAyah129: dom.AyahNumber129,
	db.RagAyah130: dom.AyahNumber130,
	db.RagAyah131: dom.AyahNumber131,
	db.RagAyah132: dom.AyahNumber132,
	db.RagAyah133: dom.AyahNumber133,
	db.RagAyah134: dom.AyahNumber134,
	db.RagAyah135: dom.AyahNumber135,
	db.RagAyah136: dom.AyahNumber136,
	db.RagAyah137: dom.AyahNumber137,
	db.RagAyah138: dom.AyahNumber138,
	db.RagAyah139: dom.AyahNumber139,
	db.RagAyah140: dom.AyahNumber140,
	db.RagAyah141: dom.AyahNumber141,
	db.RagAyah142: dom.AyahNumber142,
	db.RagAyah143: dom.AyahNumber143,
	db.RagAyah144: dom.AyahNumber144,
	db.RagAyah145: dom.AyahNumber145,
	db.RagAyah146: dom.AyahNumber146,
	db.RagAyah147: dom.AyahNumber147,
	db.RagAyah148: dom.AyahNumber148,
	db.RagAyah149: dom.AyahNumber149,
	db.RagAyah150: dom.AyahNumber150,
	db.RagAyah151: dom.AyahNumber151,
	db.RagAyah152: dom.AyahNumber152,
	db.RagAyah153: dom.AyahNumber153,
	db.RagAyah154: dom.AyahNumber154,
	db.RagAyah155: dom.AyahNumber155,
	db.RagAyah156: dom.AyahNumber156,
	db.RagAyah157: dom.AyahNumber157,
	db.RagAyah158: dom.AyahNumber158,
	db.RagAyah159: dom.AyahNumber159,
	db.RagAyah160: dom.AyahNumber160,
	db.RagAyah161: dom.AyahNumber161,
	db.RagAyah162: dom.AyahNumber162,
	db.RagAyah163: dom.AyahNumber163,
	db.RagAyah164: dom.AyahNumber164,
	db.RagAyah165: dom.AyahNumber165,
	db.RagAyah166: dom.AyahNumber166,
	db.RagAyah167: dom.AyahNumber167,
	db.RagAyah168: dom.AyahNumber168,
	db.RagAyah169: dom.AyahNumber169,
	db.RagAyah170: dom.AyahNumber170,
	db.RagAyah171: dom.AyahNumber171,
	db.RagAyah172: dom.AyahNumber172,
	db.RagAyah173: dom.AyahNumber173,
	db.RagAyah174: dom.AyahNumber174,
	db.RagAyah175: dom.AyahNumber175,
	db.RagAyah176: dom.AyahNumber176,
	db.RagAyah177: dom.AyahNumber177,
	db.RagAyah178: dom.AyahNumber178,
	db.RagAyah179: dom.AyahNumber179,
	db.RagAyah180: dom.AyahNumber180,
	db.RagAyah181: dom.AyahNumber181,
	db.RagAyah182: dom.AyahNumber182,
	db.RagAyah183: dom.AyahNumber183,
	db.RagAyah184: dom.AyahNumber184,
	db.RagAyah185: dom.AyahNumber185,
	db.RagAyah186: dom.AyahNumber186,
	db.RagAyah187: dom.AyahNumber187,
	db.RagAyah188: dom.AyahNumber188,
	db.RagAyah189: dom.AyahNumber189,
	db.RagAyah190: dom.AyahNumber190,
	db.RagAyah191: dom.AyahNumber191,
	db.RagAyah192: dom.AyahNumber192,
	db.RagAyah193: dom.AyahNumber193,
	db.RagAyah194: dom.AyahNumber194,
	db.RagAyah195: dom.AyahNumber195,
	db.RagAyah196: dom.AyahNumber196,
	db.RagAyah197: dom.AyahNumber197,
	db.RagAyah198: dom.AyahNumber198,
	db.RagAyah199: dom.AyahNumber199,
	db.RagAyah200: dom.AyahNumber200,
	db.RagAyah201: dom.AyahNumber201,
	db.RagAyah202: dom.AyahNumber202,
	db.RagAyah203: dom.AyahNumber203,
	db.RagAyah204: dom.AyahNumber204,
	db.RagAyah205: dom.AyahNumber205,
	db.RagAyah206: dom.AyahNumber206,
	db.RagAyah207: dom.AyahNumber207,
	db.RagAyah208: dom.AyahNumber208,
	db.RagAyah209: dom.AyahNumber209,
	db.RagAyah210: dom.AyahNumber210,
	db.RagAyah211: dom.AyahNumber211,
	db.RagAyah212: dom.AyahNumber212,
	db.RagAyah213: dom.AyahNumber213,
	db.RagAyah214: dom.AyahNumber214,
	db.RagAyah215: dom.AyahNumber215,
	db.RagAyah216: dom.AyahNumber216,
	db.RagAyah217: dom.AyahNumber217,
	db.RagAyah218: dom.AyahNumber218,
	db.RagAyah219: dom.AyahNumber219,
	db.RagAyah220: dom.AyahNumber220,
	db.RagAyah221: dom.AyahNumber221,
	db.RagAyah222: dom.AyahNumber222,
	db.RagAyah223: dom.AyahNumber223,
	db.RagAyah224: dom.AyahNumber224,
	db.RagAyah225: dom.AyahNumber225,
	db.RagAyah226: dom.AyahNumber226,
	db.RagAyah227: dom.AyahNumber227,
	db.RagAyah228: dom.AyahNumber228,
	db.RagAyah229: dom.AyahNumber229,
	db.RagAyah230: dom.AyahNumber230,
	db.RagAyah231: dom.AyahNumber231,
	db.RagAyah232: dom.AyahNumber232,
	db.RagAyah233: dom.AyahNumber233,
	db.RagAyah234: dom.AyahNumber234,
	db.RagAyah235: dom.AyahNumber235,
	db.RagAyah236: dom.AyahNumber236,
	db.RagAyah237: dom.AyahNumber237,
	db.RagAyah238: dom.AyahNumber238,
	db.RagAyah239: dom.AyahNumber239,
	db.RagAyah240: dom.AyahNumber240,
	db.RagAyah241: dom.AyahNumber241,
	db.RagAyah242: dom.AyahNumber242,
	db.RagAyah243: dom.AyahNumber243,
	db.RagAyah244: dom.AyahNumber244,
	db.RagAyah245: dom.AyahNumber245,
	db.RagAyah246: dom.AyahNumber246,
	db.RagAyah247: dom.AyahNumber247,
	db.RagAyah248: dom.AyahNumber248,
	db.RagAyah249: dom.AyahNumber249,
	db.RagAyah250: dom.AyahNumber250,
	db.RagAyah251: dom.AyahNumber251,
	db.RagAyah252: dom.AyahNumber252,
	db.RagAyah253: dom.AyahNumber253,
	db.RagAyah254: dom.AyahNumber254,
	db.RagAyah255: dom.AyahNumber255,
	db.RagAyah256: dom.AyahNumber256,
	db.RagAyah257: dom.AyahNumber257,
	db.RagAyah258: dom.AyahNumber258,
	db.RagAyah259: dom.AyahNumber259,
	db.RagAyah260: dom.AyahNumber260,
	db.RagAyah261: dom.AyahNumber261,
	db.RagAyah262: dom.AyahNumber262,
	db.RagAyah263: dom.AyahNumber263,
	db.RagAyah264: dom.AyahNumber264,
	db.RagAyah265: dom.AyahNumber265,
	db.RagAyah266: dom.AyahNumber266,
	db.RagAyah267: dom.AyahNumber267,
	db.RagAyah268: dom.AyahNumber268,
	db.RagAyah269: dom.AyahNumber269,
	db.RagAyah270: dom.AyahNumber270,
	db.RagAyah271: dom.AyahNumber271,
	db.RagAyah272: dom.AyahNumber272,
	db.RagAyah273: dom.AyahNumber273,
	db.RagAyah274: dom.AyahNumber274,
	db.RagAyah275: dom.AyahNumber275,
	db.RagAyah276: dom.AyahNumber276,
	db.RagAyah277: dom.AyahNumber277,
	db.RagAyah278: dom.AyahNumber278,
	db.RagAyah279: dom.AyahNumber279,
	db.RagAyah280: dom.AyahNumber280,
	db.RagAyah281: dom.AyahNumber281,
	db.RagAyah282: dom.AyahNumber282,
	db.RagAyah283: dom.AyahNumber283,
	db.RagAyah284: dom.AyahNumber284,
	db.RagAyah285: dom.AyahNumber285,
	db.RagAyah286: dom.AyahNumber286,
}
