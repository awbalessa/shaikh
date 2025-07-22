package rag

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"time"

	"github.com/awbalessa/shaikh/apps/server/internal/database"
	"github.com/awbalessa/shaikh/apps/server/internal/store"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
	"golang.org/x/sync/errgroup"
)

const (
	rrf60 rrfConstant = 60
)

type Pipeline struct {
	store  *store.Store
	vc     *VoyageClient
	logger *slog.Logger
}

type ParallelSemanticSearchParams struct {
	TotalChunks int32
	Items       []VectorWithFilters
}

type ParallelSemanticSearchRow struct {
	RowsPerThread []database.SemanticSearchRow
}

type VectorWithFilters struct {
	Vector       pgvector.Vector
	LabelFilters []int16
}

type ParallelLexicalSearchParams struct {
	TotalChunks int32
	Items       []QueryWithFilters
}

type QueryWithFilters struct {
	Query       string
	ContentType database.NullContentType
	Source      database.NullSource
	Surah       pgtype.Int4
	AyahStart   pgtype.Int4
	AyahEnd     pgtype.Int4
}

type ParallelLexicalSearchRow struct {
	RowsPerThread []database.LexicalSearchRow
}

type HybridSearchParams struct {
	TotalChunks int32
	Items       []QueryWithFilters
}

type HybridSearchRow struct {
	sem []ParallelSemanticSearchRow
	lex []ParallelLexicalSearchRow
}

type rankedLists [][]int64
type rrfConstant int
type rrfScore float64

func NewPipeline(store *store.Store, vc *VoyageClient) *Pipeline {
	log := slog.Default().With(
		"component", "pipeline",
	)

	return &Pipeline{
		store:  store,
		vc:     vc,
		logger: log,
	}
}

func (p *Pipeline) ParallelSemanticSearch(
	ctx context.Context,
	arg ParallelSemanticSearchParams,
) ([]ParallelSemanticSearchRow, error) {
	type result struct {
		rows []ParallelSemanticSearchRow
		err  error
	}

	results := make(chan result, len(arg.Items))
	chunksPerThread := math.Floor(
		float64(arg.TotalChunks) / float64(len(arg.Items)),
	)
	g, ctx := errgroup.WithContext(ctx)

	log := p.logger.With(
		slog.String("method", "ParallelSemanticSearch"),
		slog.Int("chunks_per_thread", int(chunksPerThread)),
		slog.Int("num_of_threads", len(arg.Items)),
	)
	log.InfoContext(ctx, "starting parallel semantic search...")

	start := time.Now()
	for _, item := range arg.Items {
		item := item
		g.Go(func() error {
			rows, err := p.store.RunSemanticSearch(
				ctx,
				database.SemanticSearchParams{
					NumberOfChunks: int32(chunksPerThread),
					Vector:         item.Vector,
					LabelFilters:   item.LabelFilters,
				},
			)
			results <- result{
				rows: []ParallelSemanticSearchRow{
					{RowsPerThread: rows},
				},
				err: err,
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("parallel execution error: %w", err)
	}
	close(results)

	duration := time.Since(start)
	log.With(
		slog.Int64("duration_ms", duration.Milliseconds()),
	).InfoContext(ctx, "done searching: parsing results...")

	var allRows []ParallelSemanticSearchRow
	var searchErr error

	for res := range results {
		if res.err != nil && searchErr == nil {
			searchErr = res.err
		}
		allRows = append(allRows, res.rows...)
	}

	if searchErr != nil {
		return nil, fmt.Errorf("semantic search error: %w", searchErr)
	}

	duration = time.Since(start)
	log.With(
		slog.Int64("duration_ms", duration.Milliseconds()),
	).InfoContext(ctx, "done parsing: returning...")
	return allRows, nil
}

func (p *Pipeline) ParallelLexicalSearch(
	ctx context.Context,
	arg ParallelLexicalSearchParams,
) ([]ParallelLexicalSearchRow, error) {
	type result struct {
		rows []ParallelLexicalSearchRow
		err  error
	}

	results := make(chan result, len(arg.Items))
	chunksPerThread := math.Floor(
		float64(arg.TotalChunks) / float64(len(arg.Items)),
	)
	g, ctx := errgroup.WithContext(ctx)

	log := p.logger.With(
		slog.String("method", "ParallelLexicalSearch"),
		slog.Int("chunks_per_thread", int(chunksPerThread)),
		slog.Int("num_of_threads", len(arg.Items)),
	)
	log.InfoContext(ctx, "starting parallel lexical search...")

	start := time.Now()
	for _, item := range arg.Items {
		item := item
		g.Go(func() error {
			rows, err := p.store.RunLexicalSearch(
				ctx,
				database.LexicalSearchParams{
					NumberOfChunks: int32(chunksPerThread),
					Query:          item.Query,
					ContentType:    item.ContentType,
					Source:         item.Source,
					Surah:          item.Surah,
					AyahStart:      item.AyahStart,
					AyahEnd:        item.AyahEnd,
				},
			)
			results <- result{
				rows: []ParallelLexicalSearchRow{
					{RowsPerThread: rows},
				},
				err: err,
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("parallel execution error: %w", err)
	}
	close(results)

	duration := time.Since(start)
	log.With(
		slog.Int64("duration_ms", duration.Milliseconds()),
	).InfoContext(ctx, "done searching: parsing results...")

	var allRows []ParallelLexicalSearchRow
	var searchErr error

	for res := range results {
		if res.err != nil && searchErr == nil {
			searchErr = res.err
		}
		allRows = append(allRows, res.rows...)
	}

	if searchErr != nil {
		return nil, fmt.Errorf("lexical search error: %w", searchErr)
	}

	duration = time.Since(start)
	log.With(
		slog.Int64("duration_ms", duration.Milliseconds()),
	).InfoContext(ctx, "done parsing: returning...")
	return allRows, nil
}

func (p *Pipeline) HybridSearch(
	ctx context.Context,
	arg HybridSearchParams,
) (HybridSearchRow, error) {
	var (
		semRows []ParallelSemanticSearchRow
		lexRows []ParallelLexicalSearchRow
	)

	chunksPerThread := math.Floor(
		float64(arg.TotalChunks) / float64(2),
	)
	g, ctx := errgroup.WithContext(ctx)

	log := p.logger.With(
		slog.String("method", "HybridSearch"),
		slog.Int("chunks_per_thread", int(chunksPerThread)),
		slog.Int("num_of_threads", 2),
	)
	log.InfoContext(ctx, "starting hybrid search...")

	start := time.Now()
	g.Go(func() error {
		queries := make([]string, len(arg.Items))
		for i, item := range arg.Items {
			queries[i] = item.Query
		}

		vecs, err := p.vc.EmbedQueries(
			ctx,
			queries,
		)
		if err != nil {
			return fmt.Errorf("semantic search failed: %w", err)
		}

		vecsWithFilters := make([]VectorWithFilters, len(vecs))
		for _, vec := range vecs {
			vecsWithFilters = append(
				VectorWithFilters{
					Vector: vec,
					LabelFilters: []int16{

					},
				}
			)
		}
	})
}

// func (p *Pipeline) Search() -> top level full. hybrid then parallel then rrf then reranker.

func rrfusion(rankings rankedLists, k rrfConstant) map[int64]rrfScore {
	scores := make(map[int64]rrfScore)

	for _, ranking := range rankings {
		for rank, docID := range ranking {
			score := 1.0 / float64(
				int(k)+rank,
			)
			scores[docID] += rrfScore(score)
		}
	}

	return scores
}
