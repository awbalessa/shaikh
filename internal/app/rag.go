package app

import (
	"fmt"

	"github.com/awbalessa/shaikh/internal/database"
	"github.com/pgvector/pgvector-go"
	"google.golang.org/genai"
)

// Add PreProcess step

func (a *App) EmbedQuery(query string) (*genai.EmbedContentResponse, error) {
	content := []*genai.Content{
		{
			Parts: []*genai.Part{
				{
					Text: query,
				},
			},
		},
	}
	embedCfg := &genai.EmbedContentConfig{
		TaskType:     "RETRIEVAL_QUERY",
		AutoTruncate: false,
	}

	result, err := a.GenAIClient.Models.EmbedContent(
		a.Context,
		a.Cfg.EmbeddingModel,
		content,
		embedCfg,
	)
	if err != nil {
		return nil, fmt.Errorf("Error embedding content: %v", err)
	}

	a.Logger.Println("Successfully embedded user query: %s", query)
	return result, nil
}

func (a *App) RetrieveDocuments(vector []float32, numberOfDocuments int) ([]database.CosineSimilarityRow, error) {
	params := database.CosineSimilarityParams{
		Embedding: pgvector.NewVector(vector),
		Limit:     int32(numberOfDocuments),
	}
	documents, err := a.Queries.CosineSimilarity(a.Context, params)
	if err != nil {
		return nil, fmt.Errorf("Error retrieving documents from database: %v", err)
	}
	a.Logger.Println("Successfully retrieved %d documents", len(documents))
	return documents, nil
}

func (a *App) GenerateResponse(query string, documents []database.CosineSimilarityRow)
