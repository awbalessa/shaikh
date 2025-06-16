package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/awbalessa/shaikh/apps/server/internal/config"
	"github.com/awbalessa/shaikh/apps/server/internal/database"
	"github.com/awbalessa/shaikh/apps/server/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"google.golang.org/genai"
)

func ptr[T any](v T) *T {
	return &v
}

func main() {

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  "shaikh-460416",
		Location: "europe-west4",
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		log.Fatalf("Error creating genai client: %v", err)
	}

	userQuery := "بني إسرائيل خالفوا أوامر الله وتعلموا السحر في عهد سليمان"
	embedContentResponse, err := embedQuery(ctx, cfg, client, userQuery)
	if err != nil {
		log.Fatalf("Error generating embedding: %v", err)
	}
	fmt.Println("Embedded query")

	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()
	Queries := database.New(pool)
	// Pull top 10 most relevant documents using Cosine similarity
	queryVec := pgvector.NewVector(embedContentResponse.Embeddings[0].Values)
	topFive, err := Queries.CosineSimilarity(ctx, queryVec)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Fetched top five documents")

	promptStr := fmt.Sprintf(`
			User Question:
			%s

			Ayat:
			%s
			%s
			%s
			%s
			%s
			`,
		userQuery,
		topFive[0].Content,
		topFive[1].Content,
		topFive[2].Content,
		topFive[3].Content,
		topFive[4].Content)

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{
					Text: promptStr,
				},
			},
		},
	}
	fmt.Println("Sending prompt...")
	generateConfig := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{
				{
					Text: `
						You are a helpful assistant answering a user's question using only the provided ayat.

						Instructions:
						- Answer in the same language and dialect as the user. Always.
						- Respond clearly and in a way the user can actually understand.
						- Use only the documents below. Do not use outside knowledge. Ever.
						- If the answer isn't found in the documents, say "I don't know" in the user's language.
						- Do not summarize unless asked.
						- If quoting a verse, mention its surah and ayah number.
	`,
				},
			},
		},
		Temperature: ptr(float32(0.2)),
		TopP:        ptr(float32(0.8)),
	}

	response, err := client.Models.GenerateContent(

		ctx,
		cfg.GenerationModel,
		contents,
		generateConfig,
	)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Sent prompt to Gemini")
	fmt.Printf("Vector:\n\n")
	fmt.Println(embedContentResponse.Embeddings[0].Values[:5])
	fmt.Printf("Documents:\n\n")
	fmt.Println(topFive[0].Content)
	fmt.Println(topFive[1].Content)
	fmt.Println(topFive[2].Content)
	fmt.Println(topFive[3].Content)
	fmt.Println(topFive[4].Content)
	fmt.Printf("Gemini:\n\n")
	fmt.Println(response.Candidates[0].Content.Parts[0].Text)
}

func embedQuery(ctx context.Context, cfg *config.Config, client *genai.Client, query string) (*genai.EmbedContentResponse, error) {
	content := make([]*genai.Content, 1)
	parts := make([]*genai.Part, 1)
	parts[0] = &genai.Part{Text: query}
	content[0] = &genai.Content{Parts: parts}
	embedCfg := &genai.EmbedContentConfig{
		TaskType:     "RETRIEVAL_QUERY",
		AutoTruncate: false,
	}

	result, err := client.Models.EmbedContent(
		ctx,
		cfg.EmbeddingModel,
		content,
		embedCfg,
	)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func loadVectorDB() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	defer pool.Close()
	Queries := database.New(pool)
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  "shaikh-460416",
		Location: "me-central1",
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		log.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	imlaeiSimpleQuranCom := filepath.Join(wd, "assets", "data-quran", "ayah-text", "imlaei-simple-qurancom.md")
	batch, err := parseAyatFromMarkdown(imlaeiSimpleQuranCom, 208, 293)
	if err != nil {
		log.Fatal(err)
	}

	contents := make([]*genai.Content, len(batch))
	parts := make([]*genai.Part, len(batch))
	for i := range batch {
		parts[i] = &genai.Part{
			Text: batch[i].Text,
		}
		contents[i] = &genai.Content{
			Parts: parts,
		}
	}

	fmt.Printf("Sending embedding request for %d ayat...\n", len(batch))
	embedCfg := &genai.EmbedContentConfig{
		TaskType:     "RETRIEVAL_DOCUMENT",
		Title:        "Surat Al Imran",
		AutoTruncate: false,
	}

	result, err := client.Models.EmbedContent(
		ctx,
		cfg.EmbeddingModel,
		contents,
		embedCfg,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Retreived %d embeddings successfully!\n", len(result.Embeddings))
	for i := range result.Embeddings {
		meta := models.AyahLvlMetadata{
			SurahNumber: models.SurahNumberBaqarah,
			AyahNumber:  i + 1,
		}

		bytes, err := json.Marshal(meta)
		if err != nil {
			log.Fatal(err)
		}
		config := database.CreateEmbeddingParams{
			Granularity:      database.GranularityAyah,
			ContentType:      database.ContentTypeQuran,
			Content:          batch[i].Text,
			Lang:             database.LangAr,
			LiteratureSource: database.LiteratureSourceQuran,
			EmbeddingTitle:   embedCfg.Title,
			Embedding:        pgvector.NewVector(result.Embeddings[i].Values),
			Metadata:         bytes,
		}

		fmt.Printf("Inserting embedding #%d...\n", i+1)
		_, err = Queries.CreateEmbedding(ctx, config)
		if err != nil {
			log.Fatal(err)
		}
	}
}

type ayah struct {
	GlobalNumber int
	Text         string
}

func parseAyatFromMarkdown(filePath string, startAyah, endAyah int) ([]ayah, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening file at %s: %w", filePath, err)
	}
	defer file.Close()

	var ayahs []ayah
	scanner := bufio.NewScanner(file)

	var currentAyahNumber int
	var currentAyahLines []string // Slice to accumulate lines for the current Ayah
	captureAyah := false          // Flag to indicate if we should capture the current Ayah

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comment/header lines
		if line == "" || strings.HasPrefix(line, "<!--") || strings.HasPrefix(line, "-->") || strings.HasPrefix(line, "Source") || strings.HasPrefix(line, "Text type") {
			continue
		}

		// Check if the line is an ayah number marker
		if strings.HasPrefix(line, "#") {
			// If we were collecting lines for a previous ayah AND we were capturing it,
			// process the collected lines and append the Ayah
			if currentAyahNumber != 0 && len(currentAyahLines) > 0 && captureAyah {
				fullAyahText := strings.Join(currentAyahLines, " ")
				ayahs = append(ayahs, ayah{GlobalNumber: currentAyahNumber, Text: fullAyahText})
			}

			// Parse the new ayah number
			ayahNumberStr := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			ayahNumber, err := strconv.Atoi(ayahNumberStr)
			if err != nil {
				// Return an error if parsing fails, including the problematic line
				return nil, fmt.Errorf("error parsing ayah number from line '%s': %w", line, err)
			}

			// Reset for the new Ayah
			currentAyahNumber = ayahNumber
			currentAyahLines = []string{} // Start collecting lines for the new Ayah

			// Determine if we should capture this ayah based on the range
			// The range check is done here when we encounter the ayah number marker
			if (startAyah == 0 || ayahNumber >= startAyah) && (endAyah == 0 || ayahNumber <= endAyah) {
				captureAyah = true
			} else {
				captureAyah = false
			}

		} else {
			// This line is part of the current ayah text
			// Only append the line if we have an ayah number established
			// and we are currently capturing Ayahs
			if currentAyahNumber != 0 && captureAyah {
				currentAyahLines = append(currentAyahLines, line)
			}
		}
	}

	// After the loop, add the last collected ayah if applicable and if it was being captured
	if currentAyahNumber != 0 && len(currentAyahLines) > 0 && captureAyah {
		fullAyahText := strings.Join(currentAyahLines, " ")
		ayahs = append(ayahs, ayah{GlobalNumber: currentAyahNumber, Text: fullAyahText})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return ayahs, nil
}
