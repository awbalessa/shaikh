package app

import (
	"fmt"
	"iter"

	"github.com/awbalessa/shaikh/internal/database"
	"github.com/awbalessa/shaikh/internal/prompt"
	"github.com/pgvector/pgvector-go"
	"google.golang.org/genai"
)

// Add PreProcess step

func (a *App) EmbedQuery(query string) ([]float32, error) {
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
	return result.Embeddings[0].Values, nil
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

func (a *App) GenerateResponseStream(query string, documents []database.CosineSimilarityRow) iter.Seq2[*genai.GenerateContentResponse, error] {
	instr := []string{
		"Always answer in the same language that the user prompt is in",
		"Respond in Markdown format and use styling to your advantage to try to convey your message as best as you can to the user. Make sure to use headings, bolded text, tables, etc.",
		"Base your answers on the documents below. Never ever refer to knowledge outside these documents. If the user prompt requires knowledge outside these documents, tell the user you don't know again in the language of their prompt. You don't have to quote the documents word for word, but always base your answer on the documents.",
		"Prioritize documents of highest similarity when you're responding.",
		"Make sure to always include any metadata that comes with the document in your response.",
		"For documents that are not Quran, make sure to cite the source in your response.",
	}
	prompt := prompt.NewBuilder().
		WithInstructions(instr []string)
}
