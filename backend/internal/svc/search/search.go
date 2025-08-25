package svc

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"golang.org/x/sync/errgroup"
)

type SearchSvc struct {
	Searcher dom.Searcher
	Embedder dom.Embedder
	Reranker dom.Reranker
	Logger   *slog.Logger
}

func (s *SearchSvc) Search(ctx context.Context, arg dom.SearchQuery) ([]dom.SearchResult, error) {
	queries, err := dom.ValidateSearchQuery(arg)
	if err != nil {
		return nil, err
	}

	numOfQueries := len(arg.QueriesWithFilters)
	s.Logger.With(
		slog.String("method", "Search"),
		slog.Int("num_of_prompts", numOfQueries),
		slog.Int("topk", int(arg.TopK)),
	).DebugContext(ctx, "starting search...")

	start := time.Now()
	results, err := s.hybridSearch(ctx, queries, dom.InitialChunks200)
	if err != nil {
		return nil, fmt.Errorf("failed search: %w", err)
	}

	s.Logger.With(
		slog.String("duration", time.Since(start).String()),
	).DebugContext(ctx, "search completed: sending to reranker...")

	docs := make([]string, 0, len(results))
	for _, chunk := range results {
		docs = append(docs, chunk.Content)
	}

	ranks, err := s.Reranker.RerankDocuments(ctx, arg.FullQuery, docs, arg.TopK)
	if err != nil {
		return nil, fmt.Errorf("failed search: %w", err)
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

	s.Logger.With(
		slog.String("duration", time.Since(start).String()),
		slog.Int("chunks_returned", len(final)),
	).DebugContext(ctx, "reranking completed: returning...")

	return final, nil
}

func (s *SearchSvc) hybridSearch(
	ctx context.Context,
	queries []dom.FullQueryContext,
	topk int,
) ([]dom.Chunk, error) {
	semChan := make(chan [][]dom.Chunk, 1)
	lexChan := make(chan [][]dom.Chunk, 1)

	numOfQueries := len(queries)
	chunksPerKind := topk / 2

	g, ctx := errgroup.WithContext(ctx)
	s.Logger.With(
		slog.String("method", "hybridSearch"),
		slog.Int("chunks_per_thread", chunksPerKind),
		slog.Int("num_of_threads", 2),
	).DebugContext(ctx, "starting hybrid search...")

	start := time.Now()
	g.Go(func() error {
		queriesSlice := make([]string, numOfQueries)
		for i, q := range queries {
			queriesSlice[i] = q.Query
		}
		vecs, err := s.Embedder.EmbedQueries(ctx, queriesSlice)
		if err != nil {
			return fmt.Errorf("hybrid search error: %w", err)
		}

		for i := range queries {
			queries[i].Vector = vecs[i]
		}

		semRes, err := s.parallelSemanticSearch(ctx, queries, chunksPerKind)
		if err != nil {
			return fmt.Errorf("hybrid search error: %w", err)
		}

		semChan <- semRes
		return nil
	})

	g.Go(func() error {
		lexRes, err := s.parallelLexicalSearch(ctx, queries, chunksPerKind)
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

	s.Logger.With(
		slog.String("duration", time.Since(start).String()),
		slog.Int("fused_count", len(allChunks)),
		slog.Int("deduped_count", len(deduped)),
	).DebugContext(ctx, "hybrid search completed: returning...")

	return deduped, nil
}

func (s *SearchSvc) parallelSemanticSearch(
	ctx context.Context,
	queries []dom.FullQueryContext,
	topk int,
) ([][]dom.Chunk, error) {
	chunksPerThread := topk / len(queries)
	results := make([][]dom.Chunk, len(queries))

	g, ctx := errgroup.WithContext(ctx)
	s.Logger.With(
		slog.String("method", "parallelSemanticSearch"),
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
			rows, err := s.Searcher.SemanticSearch(ctx, query.VectorWithLabel, chunksPerThread)
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

	s.Logger.With(
		slog.String("method", "parallelSemanticSearch"),
		slog.String("duration", time.Since(start).String()),
	).DebugContext(ctx, "parallel semantic search completed: returning...")

	return results, nil
}

func (s *SearchSvc) parallelLexicalSearch(
	ctx context.Context,
	queries []dom.FullQueryContext,
	topk int,
) ([][]dom.Chunk, error) {
	chunksPerThread := topk / len(queries)
	results := make([][]dom.Chunk, len(queries))

	g, ctx := errgroup.WithContext(ctx)

	s.Logger.With(
		slog.String("method", "parallelLexicalSearch"),
		slog.Int("chunks_per_thread", chunksPerThread),
		slog.Int("num_of_threads", len(queries)),
	).DebugContext(ctx, "starting parallel lexical search...")

	start := time.Now()
	for i, query := range queries {
		i, query := i, query
		g.Go(func() error {
			rows, err := s.Searcher.LexicalSearch(ctx, query.QueryWithFilter, chunksPerThread)
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

	s.Logger.With(
		slog.String("method", "parallelLexicalSearch"),
		slog.String("duration", time.Since(start).String()),
	).DebugContext(ctx, "lexical search completed: returning...")
	return results, nil
}
