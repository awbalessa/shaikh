package dom

import (
	"context"
	"sort"
	"time"
)

type Vector []float32

type Embedder interface {
	EmbedQueries(ctx context.Context, queries []string) ([]Vector, error)
}

type Rank struct {
	Index     int32
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
		vector VectorWithLabel,
		topk int,
	) ([]Chunk, error)
}

type LexicalSearcher interface {
	LexicalSearch(
		ctx context.Context,
		query QueryWithFilter,
		topk int,
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

type rankedLists [][]int32

func RRFusion(sem []Chunk, lex []Chunk) []Chunk {
	ranked := rankedLists{}
	rowMap := make(map[int32]Chunk)
	semIDs := make([]int32, 0, len(sem))
	lexIDs := make([]int32, 0, len(lex))

	for _, row := range sem {
		semIDs = append(semIDs, row.ID)
		rowMap[row.ID] = row
	}
	ranked = append(ranked, semIDs)

	for _, row := range lex {
		lexIDs = append(lexIDs, row.ID)
		rowMap[row.ID] = row
	}
	ranked = append(ranked, lexIDs)

	scores := rrfusion(ranked)

	pairs := make([]Rank, 0, len(scores))
	for id, score := range scores {
		pairs = append(pairs, Rank{Index: id, Relevance: float64(score)})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].Relevance > pairs[j].Relevance
	})

	total := len(pairs)
	cutoff := total
	if total > 100 {
		cutoff = total / 2
	}
	top := pairs[:cutoff]

	fused := make([]Chunk, 0, cutoff)
	for _, pair := range top {
		fused = append(fused, rowMap[pair.Index])
	}

	return fused
}

func rrfusion(rankings rankedLists) map[int32]float64 {
	scores := make(map[int32]float64)

	for _, ranking := range rankings {
		for rank, docID := range ranking {
			score := 1.0 / float64(
				RRFConstant+rank,
			)
			scores[docID] += score
		}
	}
	return scores
}

func ValidateSearchQuery(arg SearchQuery) ([]FullQueryContext, error) {
	if len(arg.QueriesWithFilters) == 0 {
		return nil, ErrNoSubqueries
	}
	if len(arg.QueriesWithFilters) > MaxSubqueries {
		return nil, ErrTooManySubqueries
	}

	out := make([]FullQueryContext, 0, len(arg.QueriesWithFilters))

	for _, item := range arg.QueriesWithFilters {
		f := item.FilterContext // start with the user-provided filters

		switch {
		case len(f.OptionalAyahs) > 0 && len(f.OptionalSurahs) != 1:
			return nil, ErrAyahNeedsSingleSurah

		case len(f.OptionalSurahs) > 1:
			f.OptionalAyahs = nil
		}

		labels := FiltersToLabels(f)

		out = append(out, FullQueryContext{
			QueryWithFilter: QueryWithFilter{
				Query:         item.Query,
				FilterContext: f,
			},
			VectorWithLabel: VectorWithLabel{
				LabelContext: labels,
				Vector:       nil,
			},
		})
	}

	return out, nil
}

func FiltersToLabels(f FilterContext) LabelContext {
	var (
		contentTypes []LabelContentType = []LabelContentType{}
		sources      []LabelSource      = []LabelSource{}
		surahs       []LabelSurahNumber = []LabelSurahNumber{}
		ayahs        []LabelAyahNumber  = []LabelAyahNumber{}
	)

	if len(f.OptionalContentTypes) > 0 {
		for _, ct := range f.OptionalContentTypes {
			contentTypes = append(contentTypes, ContentTypeToLabel[ct])
		}
	}

	if len(f.OptionalSources) > 0 {
		for _, src := range f.OptionalSources {
			sources = append(sources, SourceToLabel[src])
		}
	}

	if len(f.OptionalSurahs) > 0 {
		for _, sur := range f.OptionalSurahs {
			surahs = append(surahs,
				LabelSurahNumber(sur+SurahNumber(SurahNumberToLabelOffset)),
			)
		}
	}

	if len(f.OptionalAyahs) > 0 {
		for _, aya := range f.OptionalAyahs {
			ayahs = append(ayahs,
				LabelAyahNumber(aya+AyahNumber(AyahNumberToLabelOffset)),
			)
		}
	}

	return LabelContext{
		OptionalContentTypeLabels: contentTypes,
		OptionalSourceLabels:      sources,
		OptionalSurahLabels:       surahs,
		OptionalAyahLabels:        ayahs,
	}
}
