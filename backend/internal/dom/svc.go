package dom

import (
	"errors"
	"sort"
)

var (
	ErrQueriesVectorsNot1to1 = errors.New("vectors and queries are not one-to-one")
	ErrNoSubqueries          = errors.New("must pass in at least one subquery")
	ErrTooManySubqueries     = errors.New("cannot pass in more than 3 sub-queries")
	ErrAyahNeedsSingleSurah  = errors.New("must specify exactly one surah when specifying ayah filters")
)

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

type FilterContext struct {
	OptionalContentTypes []ContentType
	OptionalSources      []Source
	OptionalSurahs       []SurahNumber
	OptionalAyahs        []AyahNumber
}

type LabelContext struct {
	OptionalContentTypeLabels []LabelContentType
	OptionalSourceLabels      []LabelSource
	OptionalSurahLabels       []LabelSurahNumber
	OptionalAyahLabels        []LabelAyahNumber
}

type QueryWithFilter struct {
	Query string
	FilterContext
}

type VectorWithLabel struct {
	Vector Vector
	LabelContext
}

type FullQueryContext struct {
	QueryWithFilter
	VectorWithLabel
}

type SearchQuery struct {
	FullQuery          string
	TopK               TopK
	QueriesWithFilters []QueryWithFilter
}

type SearchResult struct {
	Chunk
	Relevance float64
}

type InputPrompt struct {
	Text             string
	FunctionResponse *LLMFunctionResponse
}

type ModelOutput struct {
	Text         string
	FunctionCall *LLMFunctionCall
}

type Interaction struct {
	Input      InputPrompt
	Output     ModelOutput
	TurnNumber int32
}

type ContextWindow struct {
	UserMemories     []Memory
	PreviousSessions []Session
	History          []Interaction
	Turns            int32
}
