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
	"google.golang.org/genai"
)

type Ayah struct {
	GlobalNumber int
	Text         string
}

func ParseAyahs(filePath string, startAyah, endAyah int) ([]Ayah, error) {
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
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  cfg.GeminiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		log.Fatal(err)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Error getting working directory: %v", err)
	}
	imlaeiSimpleQuranCom := filepath.Join(wd, "assets", "data-quran", "ayah-text", "imlaei-simple-qurancom.md")
	baqarah, err := ParseAyahs(imlaeiSimpleQuranCom, 8, 293)
	if err != nil {
		log.Fatalf("Error parsing ayahs: %v", err)
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
