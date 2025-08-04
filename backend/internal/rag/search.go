package rag

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sort"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/database"
	"github.com/awbalessa/shaikh/backend/internal/models"
	"github.com/pgvector/pgvector-go"
	"golang.org/x/sync/errgroup"
)

type PromptWithFilters struct {
	Prompt               string
	NullableContentTypes []database.ContentType
	NullableSources      []database.Source
	NullableSurahs       []models.SurahNumber
	NullableAyahs        []models.AyahNumber
}

type SearchParameters struct {
	RawPrompt          string
	ChunkLimit         TopK
	PromptsWithFilters []PromptWithFilters
}

type SearchResult struct {
	ID            int64
	Relevance     float64
	EmbeddedChunk string
	Source        string
	Surah         *int32
	Ayah          *int32
}

func (p *Pipeline) Search(ctx context.Context, arg SearchParameters) ([]SearchResult, error) {
	queries, err := validateSearchParams(arg)
	if err != nil {
		return nil, err
	}

	numOfPrompts := len(arg.PromptsWithFilters)
	log := p.logger.With(
		slog.String("method", "Search"),
		slog.Int("num_of_prompts", numOfPrompts),
		slog.Int("chunk_limit", int(arg.ChunkLimit)),
	)

	log.DebugContext(ctx, "starting search...")
	start := time.Now()
	results, err := p.hybridSearch(ctx, queries, int(initial200))
	if err != nil {
		return nil, fmt.Errorf("failed search: %w", err)
	}

	log.With(
		slog.String("duration", time.Since(start).String()),
	).DebugContext(ctx, "search completed: sending to reranker...")

	docs := make([]string, 0, len(results))
	for _, chunk := range results {
		docs = append(docs, chunk.embeddedChunk)
	}

	ranks, err := p.vc.rerankDocuments(ctx, arg.RawPrompt, docs, arg.ChunkLimit)
	if err != nil {
		return nil, fmt.Errorf("failed search: %w", err)
	}

	final := make([]SearchResult, 0, len(ranks))
	for _, rank := range ranks {
		chunk := results[rank.Index]
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
		slog.String("duration", time.Since(start).String()),
		slog.Int("chunks_returned", len(final)),
	).DebugContext(ctx, "reranking completed: returning...")

	return final, nil
}

const (
	initial200 initialChunks = 200
	rrf60      rrfConstant   = 60
	max3       maxSubPrompts = 3
)

type initialChunks int
type maxSubPrompts int
type surahAyahFilterMode int

type queryFilters struct {
	contentTypes []database.ContentType
	sources      []database.Source
	surahs       []int32
	ayahs        []int32
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

func validateSearchParams(arg SearchParameters) ([]queryContext, error) {
	if arg.PromptsWithFilters == nil || len(arg.PromptsWithFilters) == 0 {
		return nil, errors.New("must pass in at least one prompt")
	}
	if len(arg.PromptsWithFilters) > int(max3) {
		return nil, errors.New("cannot pass in more than 3 sub-prompts")
	}

	queries := make([]queryContext, 0, len(arg.PromptsWithFilters))

	for _, item := range arg.PromptsWithFilters {
		var cts []database.ContentType
		var srcs []database.Source
		var surahs []int32
		var ayahs []int32

		if len(item.NullableContentTypes) > 0 {
			cts = append(cts, item.NullableContentTypes...)
		} else {
			cts = nil
		}

		if len(item.NullableSources) > 0 {
			srcs = append(srcs, item.NullableSources...)
		} else {
			srcs = nil
		}

		if len(item.NullableAyahs) > 0 && len(item.NullableSurahs) != 1 {
			return nil, errors.New("must specify exactly one surah to specify ayah filters")
		} else if len(item.NullableAyahs) > 0 && len(item.NullableSurahs) == 1 {
			for _, s := range item.NullableSurahs {
				surahs = append(surahs, int32(s))
			}
			for _, a := range item.NullableAyahs {
				ayahs = append(ayahs, int32(a))
			}
		} else if len(item.NullableSurahs) > 1 {
			for _, s := range item.NullableSurahs {
				surahs = append(surahs, int32(s))
			}
			ayahs = nil
		}

		var f queryFilters
		f = queryFilters{
			contentTypes: cts,
			sources:      srcs,
			surahs:       surahs,
			ayahs:        ayahs,
		}

		queries = append(queries, queryContext{
			query:        item.Prompt,
			filters:      &f,
			labelFilters: filtersToLabels(f),
			vector:       nil,
		})
	}

	return queries, nil
}

func filtersToLabels(f queryFilters) []int16 {
	var labels []int16 = []int16{}

	if len(f.contentTypes) > 0 {
		for _, ct := range f.contentTypes {
			labels = append(labels, int16(models.ContentTypeToLabel[ct]))
		}
	}

	if len(f.sources) > 0 {
		for _, src := range f.sources {
			labels = append(labels, int16(models.SourceToLabel[src]))
		}
	}

	if len(f.surahs) > 0 {
		for _, sur := range f.surahs {
			labels = append(labels, int16(models.SurahNumberToLabel[sur]))
		}
	}

	if len(f.ayahs) > 0 {
		for _, aya := range f.ayahs {
			labels = append(labels, int16(models.AyahNumberToLabel[aya]))
		}
	}

	return labels
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
	log.DebugContext(ctx, "starting parallel semantic search...")

	start := time.Now()
	for i, query := range queries {
		i, query := i, query
		g.Go(func() error {
			if query.vector == nil {
				return fmt.Errorf("missing vector for query: %q", query.query)
			}
			rows, err := p.store.Pg.RunSemanticSearch(ctx, database.SemanticSearchParams{
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
		slog.String("duration", time.Since(start).String()),
	).DebugContext(ctx, "semantic search completed: returning...")

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
	log.DebugContext(ctx, "starting parallel lexical search...")

	start := time.Now()
	for i, query := range queries {
		i, query := i, query
		g.Go(func() error {
			rows, err := p.store.Pg.RunLexicalSearch(ctx, database.LexicalSearchParams{
				NumberOfChunks: int32(chunksPerThread),
				Query:          query.query,
				ContentTypes:   query.filters.contentTypes,
				Sources:        query.filters.sources,
				Surahs:         query.filters.surahs,
				Ayahs:          query.filters.ayahs,
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
		slog.String("duration", time.Since(start).String()),
	).DebugContext(ctx, "lexical search completed: returning...")
	return results, nil
}

func (p *Pipeline) hybridSearch(
	ctx context.Context,
	queries []queryContext,
	totalChunks int,
) ([]resultChunks, error) {
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
	log.DebugContext(ctx, "starting hybrid search...")
	start := time.Now()

	g.Go(func() error {
		queriesSlice := make([]string, numOfQueries)
		for i, q := range queries {
			queriesSlice[i] = q.query
		}
		vecs, err := p.vc.embedQueries(ctx, queriesSlice)
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
		fused[i] = runRRFusion(semRes[i], lexRes[i])
	}

	// Flatten all fused results into one slice
	var allChunks []resultChunks
	for _, group := range fused {
		allChunks = append(allChunks, group...)
	}

	// Deduplicate based on rawChunk content
	seen := make(map[string]bool)
	deduped := make([]resultChunks, 0, len(allChunks))
	for _, chunk := range allChunks {
		if !seen[chunk.embeddedChunk] {
			seen[chunk.embeddedChunk] = true
			deduped = append(deduped, chunk)
		}
	}

	// Log final deduped result count
	log.With(
		slog.String("duration", time.Since(start).String()),
		slog.Int("fused_count", len(allChunks)),
		slog.Int("deduped_count", len(deduped)),
	).DebugContext(ctx, "hybrid search completed: returning...")

	// Return final deduped slice (wrapped in [][]resultChunks to match method signature)
	return deduped, nil
}

func runRRFusion(
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
