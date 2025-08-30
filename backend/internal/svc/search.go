package svc

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"golang.org/x/sync/errgroup"
)

type SearchSvc struct {
	Searcher dom.Searcher
	Embedder dom.Embedder
	Reranker dom.Reranker
	Logger   *slog.Logger
}

func BuildSearchSvc(se dom.Searcher, em dom.Embedder, re dom.Reranker) *SearchSvc {
	log := slog.Default().With(
		"service", "search",
	)

	return &SearchSvc{
		Searcher: se,
		Embedder: em,
		Reranker: re,
		Logger:   log,
	}
}

func (s *SearchSvc) Search(ctx context.Context, arg dom.SearchQuery) ([]dom.SearchResult, error) {
	queries, err := dom.ValidateSearchQuery(arg)
	if err != nil {
		return nil, err
	}

	log := s.Logger.With(
		"method", "Search",
	)

	numOfQueries := len(arg.QueriesWithFilters)
	s.Logger.With(
		slog.Int("number_of_queries", numOfQueries),
		slog.Int("topk", int(arg.TopK)),
	).DebugContext(ctx, "starting hybrid search...")

	results, err := s.hybridSearch(ctx, queries, dom.InitialChunks200)
	if err != nil {
		return nil, fmt.Errorf("hybrid search failed: %w", err)
	}

	log.DebugContext(ctx, "hybrid search completed, reranking...")

	docs := make([]string, 0, len(results))
	for _, chunk := range results {
		docs = append(docs, chunk.Content)
	}

	ranks, err := s.Reranker.RerankDocuments(ctx, arg.FullQuery, docs, arg.TopK)
	if err != nil {
		return nil, fmt.Errorf("reranking failed: %w", err)
	}

	final := make([]dom.SearchResult, 0, len(ranks))
	for _, rank := range ranks {
		chunk := results[rank.Index]
		final = append(final, dom.SearchResult{
			Chunk: dom.Chunk{
				Document: dom.Document{
					ID:          chunk.ID,
					Source:      chunk.Source,
					Content:     chunk.Content,
					SurahNumber: chunk.SurahNumber,
					AyahNumber:  chunk.AyahNumber,
				},
				ParentID: chunk.ParentID,
			},
			Relevance: rank.Relevance,
		})
	}

	log.With(
		"returned_chunks", len(final),
	).DebugContext(ctx, "search completed successfully")

	return final, nil
}

func (s *SearchSvc) hybridSearch(
	ctx context.Context,
	queries []dom.FullQueryContext,
	topk int,
) ([]dom.Chunk, error) {
	log := s.Logger.With(
		"method", "hybridSearch",
	)

	semChan := make(chan [][]dom.Chunk, 1)
	lexChan := make(chan [][]dom.Chunk, 1)

	numOfQueries := len(queries)
	chunksPerKind := topk / 2

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		queriesSlice := make([]string, numOfQueries)
		for i, q := range queries {
			queriesSlice[i] = q.Query
		}
		vecs, err := s.Embedder.EmbedQueries(ctx, queriesSlice)
		if err != nil {
			return err
		}

		for i := range queries {
			queries[i].Vector = vecs[i]
		}

		semRes, err := s.Searcher.ParallelSemanticSearch(ctx, queries, chunksPerKind)
		if err != nil {
			return err
		}

		semChan <- semRes
		return nil
	})

	g.Go(func() error {
		lexRes, err := s.Searcher.ParallelLexicalSearch(ctx, queries, chunksPerKind)
		if err != nil {
			return err
		}
		lexChan <- lexRes
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	semRes := <-semChan
	lexRes := <-lexChan

	fused := make([][]dom.Chunk, len(queries))
	for i := range queries {
		fused[i] = dom.RRFusion(semRes[i], lexRes[i])
	}

	var allChunks []dom.Chunk
	for _, group := range fused {
		allChunks = append(allChunks, group...)
	}

	seen := make(map[string]bool)
	deduped := make([]dom.Chunk, 0, len(allChunks))
	for _, chunk := range allChunks {
		if !seen[chunk.Content] {
			seen[chunk.Content] = true
			deduped = append(deduped, chunk)
		}
	}

	log.With(
		"fused_count", len(fused),
		"deduped_count", len(deduped),
	).DebugContext(ctx, "hybrid search completed")

	return deduped, nil
}
