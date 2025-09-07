package svc

import (
	"context"
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
		return nil, err
	}

	hybrid, err := s.hybridSearch(ctx, queries, dom.InitialChunks200)
	if err != nil {
		return nil, err
	}

	docs := make([]string, 0, len(hybrid.results))
	for _, chunk := range hybrid.results {
		docs = append(docs, chunk.Content)
	}

	ranks, err := s.Reranker.RerankDocuments(ctx, arg.FullQuery, docs, arg.TopK)
	if err != nil {
		return nil, err
	}

	final := make([]dom.SearchResult, 0, len(ranks))
	for _, rank := range ranks {
		if int(rank.Index) < 0 || int(rank.Index) >= len(hybrid.results) {
			continue
		}
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
	if len(queries) == 0 {
		return nil, dom.NewTaggedError(dom.ErrInvalidInput, nil)
	}
	n := len(queries)
	chunksPerKind := max(1, topk/2)

	var (
		lexRes [][]dom.Chunk
		semRes [][]dom.Chunk
	)

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		r, err := s.Searcher.ParallelLexicalSearch(ctx, queries, chunksPerKind)
		if err != nil {
			return err
		}
		lexRes = r
		return nil
	})
	qStrings := make([]string, n)
	for i, q := range queries {
		qStrings[i] = q.Query
	}
	vecs, err := s.Embedder.EmbedQueries(ctx, qStrings)
	if err != nil {
		return nil, err
	}
	qs := make([]dom.FullQueryContext, n)
	copy(qs, queries)
	for i := range qs {
		qs[i].Vector = vecs[i]
	}
	sr, err := s.Searcher.ParallelSemanticSearch(ctx, qs, chunksPerKind)
	if err != nil {
		return nil, err
	}
	semRes = sr
	if err := g.Wait(); err != nil {
		return nil, err
	}

	if len(semRes) != n || len(lexRes) != n {
		return nil, dom.NewTaggedError(dom.ErrInternal, nil)
	}
	var semanticCount, lexicalCount int
	for i := range n {
		semanticCount += len(semRes[i])
		lexicalCount += len(lexRes[i])
	}

	fused := make([][]dom.Chunk, len(queries))
	var fusedCount int
	for i := range queries {
		fused[i] = dom.RRFusion(semRes[i], lexRes[i])
		fusedCount += len(fused[i])
	}

	allChunks := make([]dom.Chunk, 0, fusedCount)
	for _, group := range fused {
		allChunks = append(allChunks, group...)
	}

	seen := make(map[int32]struct{})
	deduped := make([]dom.Chunk, 0, len(allChunks))
	for _, chunk := range allChunks {
		if _, ok := seen[chunk.ID]; !ok {
			seen[chunk.ID] = struct{}{}
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