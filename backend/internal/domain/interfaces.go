package dom

import (
	"context"
	"time"
)

type Vector []float32

type Embedder interface {
	EmbedQueries(ctx context.Context, queries []string) ([]Vector, error)
}

type Rank struct {
	Index     int
	Relevance float64
}

type Reranker interface {
	RerankDocuments(
		ctx context.Context,
		query string,
		documents []string,
		topk TopK,
	) ([]Rank, error)
}

type SemanticSearcher interface {
	SemanticSearch(
		ctx context.Context,
		vector VectorWithFilter,
		topk TopK,
	) ([]Chunk, error)
}

type LexicalSearcher interface {
	LexicalSearch(
		ctx context.Context,
		query QueryWithFilter,
		topk TopK,
	) ([]Chunk, error)
}

type Searcher interface {
	SemanticSearcher
	LexicalSearcher
}

type Cache interface {
	Set(ctx context.Context, key string, value []byte, expr time.Duration) error
	Get(ctx context.Context, key string) ([]byte, error)
}
