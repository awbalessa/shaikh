package main

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/awbalessa/shaikh/backend/internal/models"
	"github.com/awbalessa/shaikh/backend/internal/rag"
	"github.com/awbalessa/shaikh/backend/internal/server"
)

func main() {
	opts := config.LoggerOptions{
		Level: slog.LevelDebug,
		JSON:  true,
	}

	slog.SetDefault(
		config.NewLogger(opts),
	)

	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	server, err := server.Serve(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer server.Close()

	res, err := server.Pipe.Search(server.Context, rag.SearchParameters{
		RawPrompt:  "لماذا قال يخرج الحي وثم قال ومخرج الميت من الحي",
		ChunkLimit: rag.Top20Documents,
		PromptsWithFilters: []rag.PromptWithFilters{
			{
				Prompt: "لماذا قال يخرج الحي وثم قال ومخرج الميت من الحي",
				NullableSurahs: []models.SurahNumber{
					models.SurahNumberSix,
				},
				NullableAyahs: []models.AyahNumber{
					models.AyahNumberNinetyFour,
					models.AyahNumberNinetyFive,
					models.AyahNumberNinetySix,
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, r := range res {
		fmt.Printf("Relevance: %.2f\n\n%s\n\n", r.Relevance, r.EmbeddedChunk)
	}
}
