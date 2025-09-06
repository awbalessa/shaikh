package svc

import (
	"context"
	"fmt"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"golang.org/x/sync/errgroup"
)

type SearchSvc struct {
	Searcher dom.Searcher
	Embedder dom.Embedder
	Reranker dom.Reranker
}

func BuildSearchSvc(se dom.Searcher, em dom.Embedder, re dom.Reranker) *SearchSvc {
	return &SearchSvc{
		Searcher: se,
		Embedder: em,
		Reranker: re,
	}
}

type SearchResult struct {
	Results  []dom.SearchResult
	Metadata map[string]any
}

func (s *SearchSvc) Search(ctx context.Context, arg dom.SearchQuery) (*SearchResult, error) {
	start := time.Now()

	queries, err := dom.ValidateSearchQuery(arg)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	hybrid, err := s.hybridSearch(ctx, queries, dom.InitialChunks200)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	docs := make([]string, 0, len(hybrid.results))
	for _, chunk := range hybrid.results {
		docs = append(docs, chunk.Content)
	}

	ranks, err := s.Reranker.RerankDocuments(ctx, arg.FullQuery, docs, arg.TopK)
	if err != nil {
		return nil, fmt.Errorf("search: %w", err)
	}

	final := make([]dom.SearchResult, 0, len(ranks))
	for _, rank := range ranks {
		chunk := hybrid.results[rank.Index]
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

	metadata := map[string]any{
		"duration_ms":           time.Since(start).Milliseconds(),
		"semantic_result_count": hybrid.metadata["semantic_result_count"],
		"lexical_result_count":  hybrid.metadata["lexical_result_count"],
		"fused_result_count":    hybrid.metadata["fused_result_count"],
		"deduped_result_count":  hybrid.metadata["deduped_result_count"],
		"final_result_count":    len(final),
	}

	return &SearchResult{
		Results:  final,
		Metadata: metadata,
	}, nil
}

type hybridSearchResult struct {
	results  []dom.Chunk
	metadata map[string]any
}

func (s *SearchSvc) hybridSearch(
	ctx context.Context,
	queries []dom.FullQueryContext,
	topk int,
) (*hybridSearchResult, error) {
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

	var semanticCount, lexicalCount int
	for i := range queries {
		semanticCount += len(semRes[i])
		lexicalCount += len(lexRes[i])
	}

	fused := make([][]dom.Chunk, len(queries))
	var fusedCount int
	for i := range queries {
		fused[i] = dom.RRFusion(semRes[i], lexRes[i])
		fusedCount += len(fused[i])
	}

	var allChunks []dom.Chunk
	for _, group := range fused {
		allChunks = append(allChunks, group...)
	}

	seen := make(map[int32]bool)
	deduped := make([]dom.Chunk, 0, len(allChunks))
	for _, chunk := range allChunks {
		if !seen[chunk.ID] {
			seen[chunk.ID] = true
			deduped = append(deduped, chunk)
		}
	}

	return &hybridSearchResult{
		results: deduped,
		metadata: map[string]any{
			"semantic_result_count": semanticCount,
			"lexical_result_count":  lexicalCount,
			"fused_result_count":    fusedCount,
			"deduped_result_count":  len(deduped),
		},
	}, nil
}
