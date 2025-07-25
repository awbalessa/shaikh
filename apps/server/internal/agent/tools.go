package agent

import (
	"google.golang.org/genai"
)

const (
	contentTypeTafsir            enumContentType = "TAFSIR"
	sourceTafsirIbnKathir        enumSource      = "TAFSIR IBN KATHIR"
	contentTypeSchemaDescription string          = "Optional filter for content types. Use only when the user's intent explicitly matches one or more of the available filter options. Otherwise, leave this filter empty to allow a broader result set."
	sourceSchemaDescription      string          = "Optional filter for sources. Use only when the user clearly refers to one more specific sources, authors, or references that match the available filter options. Otherwise, leave this filter empty to allow a broader result set."
	surahAyahVariantsDescription string          = "Exactly one of the following optional filters may be used: a single surah, a single surah with an ayah range, or a surah range. Use this filter only if the user's prompt clearly asks for or implies a specific part of the Quran. Otherwise, leave this filter empty to allow a broader result set."
)

var (
	surahNumMin float64 = 1
	surahNumMax float64 = 114
	ayahNumMin  float64 = 1
	ayahNumMax  float64 = 286
)

type enumContentType string
type enumSource string
type ToolSearch struct {
	SearchFunctions []*Function
}

type FunctionSearchChunks struct {
	Name        FunctionName
	Declaration *genai.FunctionDeclaration
}

// func BuildFunctionSearchChunks() *FunctionSearchChunks {
// 	filterCts := &genai.Schema{
// 		Title:       "Optional Content Types Filter",
// 		Type:        genai.TypeArray,
// 		Description: contentTypeSchemaDescription,
// 		Items: &genai.Schema{
// 			Title:       "Content Type",
// 			Type:        genai.TypeString,
// 			Format:      "enum",
// 			Description: "The content type to filter by.",
// 			Enum:        []string{string(contentTypeTafsir)},
// 		},
// 	}

// 	filterSrcs := &genai.Schema{
// 		Title:       "Optional Sources Filter",
// 		Type:        genai.TypeArray,
// 		Description: sourceSchemaDescription,
// 		Items: &genai.Schema{
// 			Title:       "Source",
// 			Type:        genai.TypeString,
// 			Format:      "enum",
// 			Description: "The source to fitler by.",
// 			Enum:        []string{string(sourceTafsirIbnKathir)},
// 		},
// 	}

// 	filterSurahAyahVariants := &genai.Schema{
// 		Title:       "Optional Surah and Ayah Filters",
// 		Description: surahAyahVariantsDescription,
// 		AnyOf: []*genai.Schema{
// 			{
// 				Title:       "Filter by Surah Number",
// 				Description: "Filter content by a single surah number.",
// 				Type:        genai.TypeObject,
// 				Required:    []string{"surah"},
// 				Properties: map[string]*genai.Schema{
// 					"surah": {
// 						Title:       "Surah Number",
// 						Description: "The number of the surah to filter by.",
// 						Type:        genai.TypeInteger,
// 						Format:      "int32",
// 						Minimum:     &surahNumMin,
// 						Maximum:     &surahNumMax,
// 					},
// 				},
// 			},

// 			{
// 				Title:       "Filter by Surah and Ayah Range",
// 				Description: "Filter content by specifying a surah and a range of ayahs within it.",
// 				Type:        genai.TypeObject,
// 				Required:    []string{"surah", "ayah_range"},
// 				Properties: map[string]*genai.Schema{
// 					"surah": {
// 						Title:       "Surah Number",
// 						Description: "The surah that contains the ayah range.",
// 						Type:        genai.TypeInteger,
// 						Format:      "int32",
// 						Minimum:     &surahNumMin,
// 						Maximum:     &surahNumMax,
// 					},
// 					"ayah_range": {
// 						Title:       "Ayah Range",
// 						Description: "An inclusive range of ayahs to filter within the selected surah.",
// 						Type:        genai.TypeObject,
// 						Required:    []string{"ayah_start", "ayah_end"},
// 						Properties: map[string]*genai.Schema{
// 							"ayah_start": {
// 								Title:       "Start Ayah",
// 								Description: "The first ayah in the range.",
// 								Type:        genai.TypeInteger,
// 								Format:      "int32",
// 								Minimum:     &ayahNumMin,
// 								Maximum:     &ayahNumMax,
// 							},
// 							"ayah_end": {
// 								Title:       "End Ayah",
// 								Description: "The last ayah in the range (inclusive).",
// 								Type:        genai.TypeInteger,
// 								Format:      "int32",
// 								Minimum:     &ayahNumMin,
// 								Maximum:     &ayahNumMax,
// 							},
// 						},
// 					},
// 				},
// 			},

// 			{
// 				Title:       "Filter by Surah Range",
// 				Description: "Filter content that spans multiple surahs, using an inclusive range of surah numbers.",
// 				Type:        genai.TypeObject,
// 				Required:    []string{"surah_range"},
// 				Properties: map[string]*genai.Schema{
// 					"surah_range": {
// 						Title:       "Surah Range",
// 						Description: "Defines the inclusive start and end of the surah range. The last surah number must be greater than the first surah number.",
// 						Type:        genai.TypeObject,
// 						Required:    []string{"surah_start", "surah_end"},
// 						Properties: map[string]*genai.Schema{
// 							"surah_start": {
// 								Title:       "Start Surah",
// 								Description: "The first surah in the range.",
// 								Type:        genai.TypeInteger,
// 								Format:      "int32",
// 								Minimum:     &surahNumMin,
// 								Maximum:     &surahNumMax,
// 							},
// 							"surah_end": {
// 								Title:       "End Surah",
// 								Description: "The last surah in the range (must be greater than start).",
// 								Type:        genai.TypeInteger,
// 								Format:      "int32",
// 								Minimum:     &surahNumMin,
// 								Maximum:     &surahNumMax,
// 							},
// 						},
// 					},
// 				},
// 			},
// 		},
// 	}

// 	filtersSchema :=

// 	fullSchema := &genai.Schema{
// 		Title:       "SearchChunks Parameters",
// 		Type:        genai.TypeObject,
// 		Description: "The parameters to call the SearchChunks function",
// 		Required:    []string{"full_prompt", "prompts_with_filters"},
// 		Properties: map[string]*genai.Schema{
// 			"full_prompt": {
// 				Title: "Full Prompt",
// 				Type: genai.TypeString,
// 				Format: "byte",
// 				Description: "The full user-submitted prompt (after augmenting)",
// 			},
// 			"prompts_with_filters": {
// 				Title: "Prompts With Filters",
// 				Type: genai.TypeArray,
// 				Default: any, // how to specify full_prompt
// 				Description: "logical subunits wiht their filters. Minimum one. Max 3.",
// 				MinItems: *int64(1),
// 				MaxItems: *int64(3),
// 				Items: *genai.Schema{
// 					Title: string,
// 					Type: genai.TypeObject,
// 					Description: string,
// 					Properties: map[string]*genai.Schema{
// 						"prompt": {

// 						},
// 						"filters": {

// 						}
// 					},
// 				},
// 			}
// 		},
// 	}

// 	dec := &genai.FunctionDeclaration{
// 		Description: "",
// 		Name:        "SearchChunks",
// 	}
// }

// func BuildToolSearch() *ToolSearch {
// }

func (t *FunctionSearchChunks) GetName() FunctionName {
	return t.Name
}

// func (t *FunctionSearchChunks) Call(ctx context.Context, json []byte) ([]rag.SearchResult, error) {
// 	// parse json input into the thing and call it. also log and whatev.
// }
