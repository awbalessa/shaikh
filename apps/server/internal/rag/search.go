package rag

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/awbalessa/shaikh/apps/server/internal/database"
	"github.com/awbalessa/shaikh/apps/server/internal/models"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/pgvector/pgvector-go"
	"golang.org/x/sync/errgroup"
)

type FilterContentType models.NullableContentType
type FilterSource models.NullableSource
type FilterSurahNumber models.NullableSurahNumber
type FilterSurahRange struct {
	SurahStart FilterSurahNumber
	SurahEnd   FilterSurahNumber
}
type FilterAyahNumber models.NullableAyahNumber
type FilterAyahRange struct {
	AyahStart FilterAyahNumber
	AyahEnd   FilterAyahNumber
}

type PromptWithFilters struct {
	Prompt              string
	NullableContentType *FilterContentType
	NullableSource      *FilterSource
	NullableSurahRange  *FilterSurahRange
	NullableSurah       *FilterSurahNumber
	NullableAyahRange   *FilterAyahRange
}

type SearchParameters struct {
	RawPrompt          string
	PromptsWithFilters []PromptWithFilters
	FinalChunks        TopK
}

type SearchResult struct {
	ID            int64
	Relevance     float64
	EmbeddedChunk string
	Source        string
	Surah         *int32
	Ayah          *int32
}

func (p *Pipeline) SearchChunks(ctx context.Context, arg SearchParameters) ([]SearchResult, error) {
	queries, initialChunks, err := validateSearchParams(arg)
	if err != nil {
		return nil, err
	}

	numOfPrompts := len(arg.PromptsWithFilters)
	log := p.logger.With(
		slog.String("method", "Search"),
		slog.Int("num_of_prompts", numOfPrompts),
		slog.Int("final_chunks", int(arg.FinalChunks)),
	)

	log.InfoContext(ctx, "starting search...")
	start := time.Now()
	results, err := p.hybridSearch(ctx, queries, initialChunks)
	if err != nil {
		return nil, fmt.Errorf("failed search: %w", err)
	}

	log.With(
		slog.Duration("duration", time.Since(start)),
	).InfoContext(ctx, "search completed: sending to reranker...")

	var (
		flat []resultChunks
		docs []string
	)
	for _, group := range results {
		for _, chunk := range group {
			docs = append(docs, chunk.embeddedChunk)
			flat = append(flat, chunk)
		}
	}

	ranks, err := p.vc.RerankDocuments(ctx, arg.RawPrompt, docs, arg.FinalChunks)
	if err != nil {
		return nil, fmt.Errorf("failed search: %w", err)
	}

	final := make([]SearchResult, 0, len(ranks))
	for _, rank := range ranks {
		chunk := flat[rank.Index]
		final = append(final, SearchResult{
			ID:            chunk.id,
			Relevance:     rank.RelevanceScore,
			EmbeddedChunk: chunk.embeddedChunk,
			Source:        chunk.source,
			Surah:         chunk.surah,
			Ayah:          chunk.ayah,
		})
	}

	log.With(
		slog.Duration("duration", time.Since(start)),
	).InfoContext(ctx, "reranking completed: returning...")

	return final, nil
}

const (
	surahRange     surahAyahFilterMode = 1
	surahAyahRange surahAyahFilterMode = 2
	surahOnly      surahAyahFilterMode = 3
	rrf60          rrfConstant         = 60
	max3           maxSubPrompts       = 3
)

type maxSubPrompts int
type surahAyahFilterMode int

type queryFilters struct {
	contentType database.NullContentType
	source      database.NullSource
	surahStart  pgtype.Int4
	surahEnd    pgtype.Int4
	surah       pgtype.Int4
	ayahStart   pgtype.Int4
	ayahEnd     pgtype.Int4
}

type queryContext struct {
	query        string
	filters      *queryFilters
	vector       *pgvector.Vector
	labelFilters []int16
}

type resultChunks struct {
	id            int64
	score         float64
	embeddedChunk string
	source        string
	surah         *int32
	ayah          *int32
}

type rankedLists [][]int64
type rrfConstant int
type rrfScore float64

type idScorePair struct {
	id    int64
	score float64
}

func filtersToLabels(f queryFilters) []int16 {
	var labels []int16

	if f.contentType.Valid {
		labels = append(labels, int16(models.ContentTypeToLabel[f.contentType.ContentType]))
	}

	if f.source.Valid {
		labels = append(labels, int16(models.SourceToLabel[f.source.Source]))
	}

	if f.surahStart.Valid && f.surahEnd.Valid {
		for i := f.surahStart.Int32; i <= f.surahEnd.Int32; i++ {
			labels = append(labels, int16(models.SurahNumberToLabel[i]))
		}
	} else if f.surah.Valid && f.ayahStart.Valid && f.ayahEnd.Valid {
		labels = append(labels, int16(models.SurahNumberToLabel[f.surah.Int32]))
		for i := f.ayahStart.Int32; i <= f.ayahEnd.Int32; i++ {
			labels = append(labels, int16(models.AyahNumberToLabel[i]))
		}
	} else if f.surah.Valid {
		labels = append(labels, int16(models.SurahNumberToLabel[f.surah.Int32]))
	}

	return labels
}

func validateSearchParams(arg SearchParameters) ([]queryContext, int, error) {
	if arg.PromptsWithFilters == nil || len(arg.PromptsWithFilters) == 0 {
		return nil, 0, errors.New("must pass in at least one prompt")
	}
	if len(arg.PromptsWithFilters) > int(max3) {
		return nil, 0, errors.New("cannot pass in more than 3 sub-prompts")
	}

	initialChunks := int(10 * arg.FinalChunks)
	queries := make([]queryContext, 0, len(arg.PromptsWithFilters))

	for _, item := range arg.PromptsWithFilters {
		var mode surahAyahFilterMode

		// Determine filtering mode
		if item.NullableSurahRange != nil {
			if item.NullableSurah != nil || item.NullableAyahRange != nil {
				return nil, 0, errors.New("cannot combine surah range with surah/ayah filters")
			}
			if item.NullableSurahRange.SurahStart.SurahNumber >= item.NullableSurahRange.SurahEnd.SurahNumber {
				return nil, 0, errors.New("SurahEnd must be greater than SurahStart")
			}
			mode = surahRange
		} else if item.NullableAyahRange != nil {
			if item.NullableSurah == nil {
				return nil, 0, errors.New("AyahRange filter must be paired with a Surah filter")
			}
			if item.NullableAyahRange.AyahStart.AyahNumber > item.NullableAyahRange.AyahEnd.AyahNumber {
				return nil, 0, errors.New("AyahEnd must be >= AyahStart")
			}
			mode = surahAyahRange
		} else {
			mode = surahOnly
		}

		// Build database-compatible filters
		ct := database.NullContentType{Valid: false}
		if item.NullableContentType != nil {
			ct = database.NullContentType{
				ContentType: item.NullableContentType.ContentType,
				Valid:       true,
			}
		}

		src := database.NullSource{Valid: false}
		if item.NullableSource != nil {
			src = database.NullSource{
				Source: item.NullableSource.Source,
				Valid:  true,
			}
		}

		var f queryFilters
		switch mode {
		case surahRange:
			f = queryFilters{
				contentType: ct,
				source:      src,
				surahStart:  pgtype.Int4{Int32: int32(item.NullableSurahRange.SurahStart.SurahNumber), Valid: true},
				surahEnd:    pgtype.Int4{Int32: int32(item.NullableSurahRange.SurahEnd.SurahNumber), Valid: true},
			}
		case surahAyahRange:
			f = queryFilters{
				contentType: ct,
				source:      src,
				surah:       pgtype.Int4{Int32: int32(item.NullableSurah.SurahNumber), Valid: true},
				ayahStart:   pgtype.Int4{Int32: int32(item.NullableAyahRange.AyahStart.AyahNumber), Valid: true},
				ayahEnd:     pgtype.Int4{Int32: int32(item.NullableAyahRange.AyahEnd.AyahNumber), Valid: true},
			}
		case surahOnly:
			f = queryFilters{
				contentType: ct,
				source:      src,
				surah:       pgtype.Int4{Int32: int32(item.NullableSurah.SurahNumber), Valid: true},
			}
		}

		queries = append(queries, queryContext{
			query:        item.Prompt,
			filters:      &f,
			labelFilters: filtersToLabels(f),
			vector:       nil,
		})
	}

	return queries, initialChunks, nil
}

func semChunksToResultChunks(rows []database.SemanticSearchRow) []resultChunks {
	results := make([]resultChunks, len(rows))
	for i, row := range rows {
		results[i] = resultChunks{
			id:            row.ID,
			score:         0,
			embeddedChunk: row.EmbeddedChunk,
			source:        string(row.Source),
			surah:         &(row.Surah.Int32),
			ayah:          &(row.Ayah.Int32),
		}
	}
	return results
}

func (p *Pipeline) parallelSemanticSearch(
	ctx context.Context,
	queries []queryContext,
	totalChunks int,
) ([][]resultChunks, error) {
	chunksPerThread := totalChunks / len(queries)
	results := make([][]resultChunks, len(queries))

	g, ctx := errgroup.WithContext(ctx)
	log := p.logger.With(
		slog.String("method", "parallelSemanticSearch"),
		slog.Int("chunks_per_thread", chunksPerThread),
		slog.Int("num_of_threads", len(queries)),
	)
	log.InfoContext(ctx, "starting parallel semantic search...")

	start := time.Now()
	for i, query := range queries {
		i, query := i, query
		g.Go(func() error {
			if query.vector == nil {
				return fmt.Errorf("missing vector for query: %q", query.query)
			}
			rows, err := p.store.RunSemanticSearch(ctx, database.SemanticSearchParams{
				NumberOfChunks: int32(chunksPerThread),
				Vector:         *query.vector,
				LabelFilters:   query.labelFilters,
			},
			)
			if err != nil {
				return fmt.Errorf("parallel semantic search error: %w", err)
			}

			results[i] = semChunksToResultChunks(rows)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	log.With(
		slog.Duration("duration", time.Since(start)),
	).InfoContext(ctx, "semantic search completed: returning...")

	return results, nil
}

func lexChunksToResultChunks(rows []database.LexicalSearchRow) []resultChunks {
	results := make([]resultChunks, len(rows))
	for i, row := range rows {
		results[i] = resultChunks{
			id:            row.ID,
			score:         0,
			embeddedChunk: row.EmbeddedChunk,
			source:        string(row.Source),
			surah:         &(row.Surah.Int32),
			ayah:          &(row.Ayah.Int32),
		}
	}
	return results
}

func (p *Pipeline) parallelLexicalSearch(
	ctx context.Context,
	queries []queryContext,
	totalChunks int,
) ([][]resultChunks, error) {
	chunksPerThread := totalChunks / len(queries)
	results := make([][]resultChunks, len(queries))

	g, ctx := errgroup.WithContext(ctx)

	log := p.logger.With(
		slog.String("method", "parallelLexicalSearch"),
		slog.Int("chunks_per_thread", chunksPerThread),
		slog.Int("num_of_threads", len(queries)),
	)
	log.InfoContext(ctx, "starting parallel lexical search...")

	start := time.Now()
	for i, query := range queries {
		i, query := i, query
		g.Go(func() error {
			rows, err := p.store.RunLexicalSearch(ctx, database.LexicalSearchParams{
				NumberOfChunks: int32(chunksPerThread),
				Query:          query.query,
				ContentType:    query.filters.contentType,
				Source:         query.filters.source,
				SurahStart:     query.filters.surahStart,
				SurahEnd:       query.filters.surahEnd,
				Surah:          query.filters.surah,
				AyahStart:      query.filters.ayahStart,
				AyahEnd:        query.filters.ayahEnd,
			},
			)
			if err != nil {
				return fmt.Errorf("parallel lexical search error: %w", err)
			}

			results[i] = lexChunksToResultChunks(rows)
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	log.With(
		slog.Duration("duration", time.Since(start)),
	).InfoContext(ctx, "lexical search completed: returning...")
	return results, nil
}

func (p *Pipeline) hybridSearch(
	ctx context.Context,
	queries []queryContext,
	totalChunks int,
) ([][]resultChunks, error) {
	semChan := make(chan [][]resultChunks, 1)
	lexChan := make(chan [][]resultChunks, 1)

	numOfQueries := len(queries)
	chunksPerKind := totalChunks / 2

	g, ctx := errgroup.WithContext(ctx)
	log := p.logger.With(
		slog.String("method", "hybridSearch"),
		slog.Int("chunks_per_thread", chunksPerKind),
		slog.Int("num_of_threads", 2),
	)
	log.InfoContext(ctx, "starting hybrid search...")
	start := time.Now()

	g.Go(func() error {
		queriesSlice := make([]string, 0, numOfQueries)
		for i, q := range queries {
			queriesSlice[i] = q.query
		}
		vecs, err := p.vc.EmbedQueries(ctx, queriesSlice)
		if err != nil {
			return fmt.Errorf("hybrid search error: %w", err)
		}

		semQueries := make([]queryContext, numOfQueries)
		for i := range queries {
			semQueries[i] = queries[i]
			semQueries[i].vector = &vecs[i]
		}

		semRes, err := p.parallelSemanticSearch(ctx, semQueries, chunksPerKind)
		if err != nil {
			return fmt.Errorf("hybrid search error: %w", err)
		}

		semChan <- semRes
		return nil
	})

	g.Go(func() error {
		lexRes, err := p.parallelLexicalSearch(ctx, queries, chunksPerKind)
		if err != nil {
			return fmt.Errorf("hybrid search error: %w", err)
		}
		lexChan <- lexRes
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	semRes := <-semChan
	lexRes := <-lexChan

	fused := make([][]resultChunks, len(queries))
	for i := range queries {
		fused[i] = p.runRRFusion(semRes[i], lexRes[i])
	}

	log.InfoContext(
		ctx,
		"hybrid search completed: returning...",
		slog.Duration("duration", time.Since(start)),
	)
	return fused, nil
}

func (p *Pipeline) runRRFusion(
	sem []resultChunks,
	lex []resultChunks,
) []resultChunks {
	ranked := rankedLists{}
	rowMap := make(map[int64]resultChunks)
	semIDs := make([]int64, 0, len(sem))
	lexIDs := make([]int64, 0, len(lex))

	for _, row := range sem {
		semIDs = append(semIDs, row.id)
		rowMap[row.id] = row
	}
	ranked = append(ranked, semIDs)

	for _, row := range lex {
		lexIDs = append(lexIDs, row.id)
		rowMap[row.id] = row
	}
	ranked = append(ranked, lexIDs)

	scores := rrfusion(ranked, rrf60)

	pairs := make([]idScorePair, 0, len(scores))
	for id, score := range scores {
		pairs = append(pairs, idScorePair{id: id, score: float64(score)})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].score > pairs[j].score
	})

	half := len(pairs) / 2
	top := pairs[:half]

	fused := make([]resultChunks, 0, half)
	for _, pair := range top {
		row := rowMap[pair.id]
		row.score = pair.score
		fused = append(fused, row)
	}

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
