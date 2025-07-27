package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/awbalessa/shaikh/apps/server/internal/rag"
	"google.golang.org/genai"
)

type Agent struct {
	searcher  *searcher
	generator *generator
}

func NewAgent() (*Agent, error) {
	return &Agent{
		searcher:  nil,
		generator: nil,
	}, nil
}

type searcherConfig struct {
	ctx  context.Context
	pipe *rag.Pipeline
	gc   *geminiClient
}

type searcher struct {
	rag     *toolRAG
	gc      *geminiClient
	model   geminiModel
	logger  *slog.Logger
	baseCfg *genai.GenerateContentConfig
}

type generator struct {
	gc      *geminiClient
	model   geminiModel
	logger  *slog.Logger
	baseCfg *genai.GenerateContentConfig
}

func buildSearcher(cfg searcherConfig) (*searcher, error) {
	log := slog.Default().With(
		"component", "searcher",
	)

	rag := buildToolRAG(cfg.pipe, log)

	tools := []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				rag.search.declaration,
			},
		},
	}

	resSchema := &genai.Schema{
		Title:       "Searcher Agent Output",
		Type:        genai.TypeObject,
		Description: "Structured JSON output for a natural language response from the searcher agent. The 'answer' field should be written in the same language as the user's original query and formatted using rich Markdown to enhance clarity and readability.",
		Required:    []string{"answer"},
		Properties: map[string]*genai.Schema{
			"answer": {
				Title:       "Answer",
				Type:        genai.TypeString,
				Description: "The natural language response to the user's query. It should be expressive, clear, and use Markdown formatting features (e.g., headers, bullet points, bold text, numbered lists, and tables) to organize the information in a structured and visually informative way.",
			},
		},
	}

	instr := &genai.Content{
		Parts: []*genai.Part{
			genai.NewPartFromText(`
You are Shaikh — a helpful, multilingual, scholarly AI assistant designed to make learning about the Quran more accessible, structured, and insightful for users of all backgrounds.

Your goal is to assist users in understanding Quranic content deeply, drawing only from the documents provided in the conversation history unless explicitly instructed otherwise.

## 🔍 Role and Behavior

- Always respond in the **same language** as the user's prompt.
- Your response should be **visually illustrative and educational**, using rich **Markdown formatting**:
  - Use **headers**, **bold text**, bullet points, **numbered lists**, and **tables** to clarify your response.
- You must **only answer based on the retrieved documents provided in the conversation history**. If the documents do **not** sufficiently answer the question, initiate a tool call using the 'Search' function.
- Never guess or answer without evidence. If unsure, search.

## 🧠 Prompt Context

You receive:
- A long conversation history (up to 200,000 tokens) that may include prior questions, retrieved documents, and ayat.
- A new user prompt at the end.

Use all available history to decide whether to answer or search.

## 🛠 Tool Usage: Search

When calling the 'Search' function, follow the structure defined in its schema.

You must:
- Provide a 'full_prompt': a semantically coherent, self-contained version of the user’s query, translated into **Arabic** regardless of the user's input language.
  - Include relevant context from earlier in the conversation if it improves clarity or precision.
  - This improves both **vector search and keyword retrieval**.
- Provide at least one 'prompt_with_filter' block:
  - If the prompt is simple or unified, you may reuse the 'full_prompt' as a single sub-prompt, optionally including filters.
  - If the prompt is complex or multi-part, break it into multiple focused sub-prompts, each with its own optional filters (e.g. surah, source, ayah, content type).
  - Use filters and prompt splitting **only when the user's intent clearly supports it** and it improves search accuracy.
`),
		},
	}

	generationConfig := &genai.GenerateContentConfig{
		SystemInstruction: instr,
		Temperature:       ptr(geminiTemperatureZero),
		CandidateCount:    1,
		ResponseSchema:    resSchema,
		ResponseMIMEType:  "application/json",
		Labels: map[string]string{
			"agent": "searcher",
		},
		Tools: tools,
	}

	gc, err := newGeminiClient(geminiClientConfig{
		context:    cfg.ctx,
		maxRetries: geminiMaxRetriesThree,
		timeout:    geminiTimeoutFifteenSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build searcher agent: %w", err)
	}

	return &searcher{
		rag:     rag,
		gc:      gc,
		model:   geminiFlashLiteV2p5,
		logger:  log,
		baseCfg: generationConfig,
	}, nil
}
