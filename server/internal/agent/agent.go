package agent

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/awbalessa/shaikh/server/internal/rag"
	"google.golang.org/genai"
)

type Agent struct {
	// state     *stateStore
	searcher  *searcher
	generator *generator
}

func NewAgent(ctx context.Context, p *rag.Pipeline) (*Agent, error) {
	gc, err := newGeminiClient(geminiClientConfig{
		context:    ctx,
		maxRetries: geminiMaxRetriesThree,
		timeout:    geminiTimeoutFifteenSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build new agent: %w", err)
	}

	s := buildSearcher(searcherConfig{
		pipe: p,
		gc:   gc,
	})

	g := buildGenerator(generatorConfig{
		gc: gc,
	})

	return &Agent{
		searcher:  s,
		generator: g,
	}, nil
}

// type stateStore interface {
// 	Load(ctx context.Context)
// }

type function[T any] interface {
	name() functionName
	call(ctx context.Context, args map[string]any) (T, error)
}

type searcherConfig struct {
	pipe *rag.Pipeline
	gc   *geminiClient
}

type searcher struct {
	functions map[functionName]any
	gc        *geminiClient
	model     geminiModel
	logger    *slog.Logger
	baseCfg   *genai.GenerateContentConfig
}

type generatorConfig struct {
	gc *geminiClient
}

type generator struct {
	gc      *geminiClient
	model   geminiModel
	logger  *slog.Logger
	baseCfg *genai.GenerateContentConfig
}

func typeFn[T any](fns map[functionName]any, fname functionName) (function[T], error) {
	raw, ok := fns[fname]
	if !ok {
		return nil, fmt.Errorf("function %s not registered", fname)
	}
	typed, ok := raw.(function[T])
	if !ok {
		return nil, fmt.Errorf("function %s has wrong return type", fname)
	}

	return typed, nil
}

func buildSearcher(cfg searcherConfig) *searcher {
	log := slog.Default().With(
		"component", "searcher",
	)

	fsearch := buildFunctionSearch(log)

	fmap := map[functionName]any{
		search: fsearch,
	}

	tools := []*genai.Tool{
		{
			FunctionDeclarations: []*genai.FunctionDeclaration{
				fsearch.declaration,
			},
		},
	}

	resSchema := &genai.Schema{
		Type:        genai.TypeString,
		Description: "A Markdown-formatted answer in the user's original language. Use rich formatting like headers, lists, bold text, and tables to visually illustrate your answers.",
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
- When asked a question about the Quran, you must **only answer based on the retrieved documents provided in the conversation history**. If the documents do **not** sufficiently answer the question, initiate a tool call using the 'Search' function.
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
		Labels: map[string]string{
			"agent": "searcher",
		},
		Tools: tools,
	}

	return &searcher{
		functions: fmap,
		gc:        cfg.gc,
		model:     geminiFlashLiteV2p5,
		logger:    log,
		baseCfg:   generationConfig,
	}
}

func buildGenerator(cfg generatorConfig) *generator {
	log := slog.Default().With(
		"component", "generator",
	)

	resSchema := &genai.Schema{
		Type:        genai.TypeString,
		Description: "A Markdown-formatted answer in the user's original language. Use rich formatting like headers, lists, bold text, and tables to visually illustrate your answers.",
	}

	instr := &genai.Content{
		Parts: []*genai.Part{
			genai.NewPartFromText(`
You are Shaikh — a helpful, multilingual, scholarly AI assistant designed to make learning about the Quran more accessible, structured, and insightful for users of all backgrounds.

Your goal is to assist users in understanding Quranic content deeply, drawing only from the documents provided in the conversation history. You are not allowed to use external tools or data sources. If the documents do not provide enough information to answer, respond humbly and transparently.

## 🔍 Role and Behavior

- Always respond in the **same language** as the user's prompt.
- Your response should be **visually illustrative and educational**, using rich **Markdown formatting**:
  - Use **headers**, **bold text**, bullet points, **numbered lists**, and **tables** to clarify your response.
- You must **only answer based on the retrieved documents provided in the conversation history**.
- **Do not guess or fabricate answers.** If the context is insufficient, say so clearly and humbly.

## 🧠 Prompt Context

You receive:
- A long conversation history (up to 200,000 tokens) that may include prior questions, retrieved documents, and ayat.
- A new user prompt at the end.

Your job is to generate a high-quality, evidence-based answer using only the provided context.
`),
		},
	}

	generationConfig := &genai.GenerateContentConfig{
		SystemInstruction: instr,
		Temperature:       ptr(geminiTemperatureZero),
		CandidateCount:    1,
		ResponseSchema:    resSchema,
		Labels: map[string]string{
			"agent": "generator",
		},
	}

	return &generator{
		gc:      cfg.gc,
		model:   geminiFlashLiteV2p5,
		logger:  log,
		baseCfg: generationConfig,
	}
}
