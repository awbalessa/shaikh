package rag

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/awbalessa/shaikh/apps/server/internal/database"
	"github.com/awbalessa/shaikh/apps/server/internal/models"
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

type parallelSemanticSearchParams struct {
	totalChunks int
	items       []vectorWithLabels
}

type parallelSemanticSearchRow struct {
	rowsPerQuery []database.SemanticSearchRow
}

type vectorWithLabels struct {
	vector       pgvector.Vector
	labelFilters []int16
}

type queryWithLabels struct {
	query        string
	labelFilters []int16
}

type parallelLexicalSearchParams struct {
	totalChunks int
	items       []queryWithFilters
}

type queryWithFilters struct {
	query       string
	contentType database.NullContentType
	source      database.NullSource
	surahStart  pgtype.Int4
	surahEnd    pgtype.Int4
	surah       pgtype.Int4
	ayahStart   pgtype.Int4
	ayahEnd     pgtype.Int4
}

type parallelLexicalSearchRow struct {
	rowsPerQuery []database.LexicalSearchRow
}

type hybridSearchParams struct {
	totalChunks int
	items       []queryWithFilters
}

type hybridSearchResult struct {
	semRowsPerQuery []parallelSemanticSearchRow
	lexRowsPerQuery []parallelLexicalSearchRow
}

type rankedLists [][]int64
type rrfConstant int
type rrfScore float64

type searchRow struct {
	ID            int64
	Score         float64
	EmbeddedChunk string
	Source        string
	Surah         int32
	Ayah          int32
}

type fusedSearchResult struct {
	searchRowsPerQuery [][]searchRow
}

type idScorePair struct {
	id    int64
	score float64
}

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

func (p *Pipeline) parallelSemanticSearch(
	ctx context.Context,
	arg parallelSemanticSearchParams,
) ([]parallelSemanticSearchRow, error) {
	chunksPerThread := arg.totalChunks / len(arg.items)

	results := make([]parallelSemanticSearchRow, len(arg.items))
	g, ctx := errgroup.WithContext(ctx)

	log := p.logger.With(
		slog.String("method", "parallelSemanticSearch"),
		slog.Int("chunks_per_thread", chunksPerThread),
		slog.Int("num_of_threads", len(arg.items)),
	)
	log.InfoContext(ctx, "starting parallel semantic search...")

	start := time.Now()
	for i, item := range arg.items {
		i, item := i, item
		g.Go(func() error {
			rows, err := p.store.RunSemanticSearch(
				ctx,
				database.SemanticSearchParams{
					NumberOfChunks: int32(chunksPerThread),
					Vector:         item.vector,
					LabelFilters:   item.labelFilters,
				},
			)
			if err != nil {
				return fmt.Errorf("parallel semantic search error: %w", err)
			}

			results[i] = parallelSemanticSearchRow{
				rowsPerQuery: rows,
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("parallel execution error: %w", err)
	}

	duration := time.Since(start)
	log.With(
		slog.Int64("duration_ms", duration.Milliseconds()),
	).InfoContext(ctx, "semantic search completed: returning...")

	return results, nil
}

func (p *Pipeline) parallelLexicalSearch(
	ctx context.Context,
	arg parallelLexicalSearchParams,
) ([]parallelLexicalSearchRow, error) {
	chunksPerThread := arg.totalChunks / len(arg.items)

	results := make([]parallelLexicalSearchRow, len(arg.items))
	g, ctx := errgroup.WithContext(ctx)

	log := p.logger.With(
		slog.String("method", "parallelLexicalSearch"),
		slog.Int("chunks_per_thread", chunksPerThread),
		slog.Int("num_of_threads", len(arg.items)),
	)
	log.InfoContext(ctx, "starting parallel lexical search...")

	start := time.Now()
	for i, item := range arg.items {
		i, item := i, item
		g.Go(func() error {
			rows, err := p.store.RunLexicalSearch(
				ctx,
				database.LexicalSearchParams{
					NumberOfChunks: int32(chunksPerThread),
					Query:          item.query,
					ContentType:    item.contentType,
					Source:         item.source,
					SurahStart:     item.surahStart,
					SurahEnd:       item.surahEnd,
					Surah:          item.surah,
					AyahStart:      item.ayahStart,
					AyahEnd:        item.ayahEnd,
				},
			)
			if err != nil {
				return fmt.Errorf("parallel lexical search error: %w", err)
			}

			results[i] = parallelLexicalSearchRow{
				rowsPerQuery: rows,
			}
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, fmt.Errorf("parallel execution error: %w", err)
	}

	duration := time.Since(start)
	log.With(
		slog.Int64("duration_ms", duration.Milliseconds()),
	).InfoContext(ctx, "lexical search completed: returning...")
	return results, nil
}

func (p *Pipeline) hybridSearch(
	ctx context.Context,
	arg hybridSearchParams,
) (hybridSearchResult, error) {
	chunksPerThread := arg.totalChunks / 2
	numOfQueries := len(arg.items)

	var result hybridSearchResult
	g, ctx := errgroup.WithContext(ctx)
	log := p.logger.With(
		slog.String("method", "hybridSearch"),
		slog.Int("chunks_per_thread", chunksPerThread),
		slog.Int("num_of_threads", 2),
	)
	log.InfoContext(ctx, "starting hybrid search...")
	start := time.Now()

	g.Go(func() error {
		qwl := make([]queryWithLabels, 0, numOfQueries)
		for _, item := range arg.items {
			qwl = append(qwl, filtersToLabels(item))
		}
		queries := make([]string, 0, numOfQueries)
		for _, item := range qwl {
			queries = append(queries, item.query)
		}
		vecs, err := p.vc.EmbedQueries(ctx, queries)
		if err != nil {
			return fmt.Errorf("hybrid search error: %w", err)
		}
		vwl := make([]vectorWithLabels, 0, len(vecs))
		for i := range vecs {
			vwl = append(vwl, vectorWithLabels{
				vector:       vecs[i],
				labelFilters: qwl[i].labelFilters,
			},
			)
		}

		semRows, err := p.parallelSemanticSearch(
			ctx,
			parallelSemanticSearchParams{
				totalChunks: chunksPerThread,
				items:       vwl,
			},
		)
		if err != nil {
			return fmt.Errorf("hybrid search error: %w", err)
		}

		result.semRowsPerQuery = semRows
		return nil
	})

	g.Go(func() error {
		lexRows, err := p.parallelLexicalSearch(
			ctx,
			parallelLexicalSearchParams{
				totalChunks: chunksPerThread,
				items:       arg.items,
			},
		)
		if err != nil {
			return fmt.Errorf("hybrid search error: %w", err)
		}

		result.lexRowsPerQuery = lexRows
		return nil
	})

	if err := g.Wait(); err != nil {
		return hybridSearchResult{}, err
	}

	log.InfoContext(
		ctx,
		"hybrid search completed: returning...",
		slog.Duration("duration", time.Since(start)),
	)
	return result, nil
}

// func (p *Pipeline) Search() -> top level full. hybrid then parallel then rrf then reranker. Scrutinize all filters heavily. Don't allow any bs. This is the top level tool call that was called after the agent returns the JSON. Make it heavily strongly typed. Enumerate everything. Sends the main original user prompt to reranker. Add guard on number of items. Check everything is good before calling the wrapped tool.

func filtersToLabels(qwf queryWithFilters) queryWithLabels {
	var qwl queryWithLabels
	qwl.query = qwf.query
	if qwf.contentType.Valid {
		qwl.labelFilters = append(
			qwl.labelFilters,
			int16(models.ContentTypeToLabel[qwf.contentType.ContentType]),
		)
	}

	if qwf.source.Valid {
		qwl.labelFilters = append(
			qwl.labelFilters,
			int16(models.SourceToLabel[qwf.source.Source]),
		)
	}

	if qwf.surahStart.Valid && qwf.surahEnd.Valid {
		var surahRange []int16
		for i := int32(0); qwf.surahStart.Int32+i <= qwf.surahEnd.Int32; i++ {
			surahRange = append(
				surahRange,
				int16(models.SurahNumberToLabel[qwf.surahStart.Int32+i]),
			)
		}

		qwl.labelFilters = append(qwl.labelFilters, surahRange...)
	} else if qwf.surah.Valid && qwf.ayahStart.Valid && qwf.ayahEnd.Valid {
		qwl.labelFilters = append(
			qwl.labelFilters,
			int16(models.SurahNumberToLabel[qwf.surah.Int32]),
		)
		var ayahRange []int16
		for i := int32(0); qwf.ayahStart.Int32+i <= qwf.ayahEnd.Int32; i++ {
			ayahRange = append(
				ayahRange,
				int16(models.AyahNumberToLabel[qwf.ayahStart.Int32+i]),
			)
		}

		qwl.labelFilters = append(qwl.labelFilters, ayahRange...)
	} else if qwf.surah.Valid {
		qwl.labelFilters = append(
			qwl.labelFilters,
			int16(models.SurahNumberToLabel[qwf.surah.Int32]),
		)
	}

	return qwl
}

func (p *Pipeline) parallelRRFusion(
	ctx context.Context,
	arg hybridSearchResult,
) (fusedSearchResult, error) {
	numQueries := len(arg.semRowsPerQuery)
	resultRows := make([][]searchRow, numQueries)

	g, ctx := errgroup.WithContext(ctx)

	log := p.logger.With(
		slog.String("method", "parallelRRFusion"),
		slog.Int("number_of_threads", numQueries),
	)

	log.InfoContext(ctx, "starting parallel RRFusion...")
	start := time.Now()
	for i := range numQueries {
		i := i
		g.Go(func() error {
			sem := arg.semRowsPerQuery[i].rowsPerQuery
			lex := arg.lexRowsPerQuery[i].rowsPerQuery

			fused := p.runRRFusion(sem, lex)
			resultRows[i] = fused
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return fusedSearchResult{}, fmt.Errorf("parallel RRFusion error: %w", err)
	}

	log.With(
		slog.Duration("duration", time.Since(start)),
	).InfoContext(ctx, "parallel RRFusion completed: returning...")
	return fusedSearchResult{
		searchRowsPerQuery: resultRows,
	}, nil
}

func (p *Pipeline) runRRFusion(
	sem []database.SemanticSearchRow,
	lex []database.LexicalSearchRow,
) []searchRow {
	lists := make(rankedLists, 2)
	rowMap := make(map[int64]searchRow, len(sem)+len(lex))
	semList := make([]int64, 0, len(sem))

	log := p.logger.With(
		slog.String("method", "runRRFusion"),
		slog.Int("semantic_rows", len(sem)),
		slog.Int("lexical_rows", len(lex)),
	)

	start := time.Now()
	log.Info("running RRFusion...")
	for _, row := range sem {
		semList = append(
			semList,
			row.ID,
		)

		rowMap[row.ID] = searchRow{
			ID:            row.ID,
			Score:         0,
			EmbeddedChunk: row.EmbeddedChunk,
			Source:        string(row.Source),
			Surah:         row.Surah.Int32,
			Ayah:          row.Ayah.Int32,
		}
	}
	lists = append(lists, semList)

	lexList := make([]int64, 0, len(lex))
	for _, row := range lex {
		lexList = append(
			lexList,
			row.ID,
		)

		rowMap[row.ID] = searchRow{
			ID:            row.ID,
			Score:         0,
			EmbeddedChunk: row.EmbeddedChunk,
			Source:        string(row.Source),
			Surah:         row.Surah.Int32,
			Ayah:          row.Ayah.Int32,
		}
	}
	lists = append(lists, lexList)

	scores := rrfusion(lists, rrf60)
	pairs := make([]idScorePair, 0, len(scores))
	for id, score := range scores {
		pairs = append(pairs, idScorePair{id: id, score: float64(score)})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].score > pairs[j].score
	})

	half := len(pairs) / 2
	top := pairs[:half]
	fused := make([]searchRow, 0, half)
	for _, pair := range top {
		if row, ok := rowMap[pair.id]; ok {
			row.Score = pair.score
			fused = append(fused, row)
		}
	}

	log.With(
		slog.Duration("duration", time.Since(start)),
	).Info("RRFusion completed: returning...")

	return fused
}

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
