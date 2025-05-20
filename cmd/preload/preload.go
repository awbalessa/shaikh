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

	"github.com/awbalessa/shaikh/internal/config"
	"github.com/awbalessa/shaikh/internal/database"
	"github.com/awbalessa/shaikh/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
	"google.golang.org/genai"
)

type Ayah struct {
	GlobalNumber int
	Text         string
}

func ParseFromMarkdown(filePath string, startAyah, endAyah int) ([]Ayah, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("Error opening file at %s: %w", filePath, err)
	}
	defer file.Close()

	var ayahs []Ayah
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

		// Check if the line is an Ayah number marker
		if strings.HasPrefix(line, "#") {
			// If we were collecting lines for a previous Ayah AND we were capturing it,
			// process the collected lines and append the Ayah
			if currentAyahNumber != 0 && len(currentAyahLines) > 0 && captureAyah {
				fullAyahText := strings.Join(currentAyahLines, " ")
				ayahs = append(ayahs, Ayah{GlobalNumber: currentAyahNumber, Text: fullAyahText})
			}

			// Parse the new Ayah number
			ayahNumberStr := strings.TrimSpace(strings.TrimPrefix(line, "#"))
			ayahNumber, err := strconv.Atoi(ayahNumberStr)
			if err != nil {
				// Return an error if parsing fails, including the problematic line
				return nil, fmt.Errorf("error parsing Ayah number from line '%s': %w", line, err)
			}

			// Reset for the new Ayah
			currentAyahNumber = ayahNumber
			currentAyahLines = []string{} // Start collecting lines for the new Ayah

			// Determine if we should capture this Ayah based on the range
			// The range check is done here when we encounter the Ayah number marker
			if (startAyah == 0 || ayahNumber >= startAyah) && (endAyah == 0 || ayahNumber <= endAyah) {
				captureAyah = true
			} else {
				captureAyah = false
			}

		} else {
			// This line is part of the current Ayah text
			// Only append the line if we have an Ayah number established
			// and we are currently capturing Ayahs
			if currentAyahNumber != 0 && captureAyah {
				currentAyahLines = append(currentAyahLines, line)
			}
		}
	}

	// After the loop, add the last collected Ayah if applicable and if it was being captured
	if currentAyahNumber != 0 && len(currentAyahLines) > 0 && captureAyah {
		fullAyahText := strings.Join(currentAyahLines, " ")
		ayahs = append(ayahs, Ayah{GlobalNumber: currentAyahNumber, Text: fullAyahText})
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	return ayahs, nil
}
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
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
	batch, err := ParseFromMarkdown(imlaeiSimpleQuranCom, 208, 293)
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
