package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/awbalessa/shaikh/internal/config"
	"google.golang.org/genai"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.GeminiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	contents := []*genai.Content{
		genai.NewContentFromText("Are you sure that's the meaning of life?", genai.RoleUser),
	}

	fmt.Println("Sending embedding request...")
	dim := int32(1536)
	embedCfg := &genai.EmbedContentConfig{
		TaskType:             "RETRIEVAL_QUERY",
		OutputDimensionality: &dim,
	}
	result, err := client.Models.EmbedContent(
		ctx,
		"gemini-embedding-exp-03-07",
		contents,
		embedCfg,
	)
	if err != nil {
		log.Fatal(err)
	}

	dims := len(result.Embeddings[0].Values)
	fmt.Printf("Genereted an emedding of %d dimensions", dims)
	embeddings, err := json.MarshalIndent(result.Embeddings, "", "  ")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(embeddings))
}
