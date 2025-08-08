package agent

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/service/rag"
	"github.com/awbalessa/shaikh/backend/internal/repo/postgres"
	"github.com/nats-io/nats.go/jetstream"
	"google.golang.org/genai"
)

const (
	AgentStream      string        = "AGENT"
	SyncerSubject    string        = "agent.context.sync"
	SyncIdleTime     time.Duration = 2 * time.Minute
	SyncMaxBatchSize int           = 5
)

type AgentConfig struct {
	Context  context.Context
	Pipeline *rag.Pipeline
	Store    *repo.Store
	Stream   jetstream.JetStream
}

type Agent struct {
	agents    map[agentName]*agentProfile
	functions map[functionName]function
	logger    *slog.Logger
	store     *repo.Store
	gc        *geminiClient
	js        jetstream.JetStream
}

func NewAgent(cfg AgentConfig) (*Agent, error) {
	log := slog.Default().With(
		"component", "agent",
	)

	gc, err := newGeminiClient(geminiClientConfig{
		context:    cfg.Context,
		maxRetries: geminiMaxRetriesThree,
		timeout:    geminiTimeoutFifteenSeconds,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build new agent: %w", err)
	}

	se := buildSearcher(searcherConfig{
		pipe:   cfg.Pipeline,
		logger: log,
	})

	g := buildGenerator()

	fmap := map[functionName]function{
		search: buildFunctionSearch(log),
	}

	amap := map[agentName]*agentProfile{
		searcherAgent: {
			name:   searcherAgent,
			model:  se.model,
			config: se.baseCfg,
		},
		generatorAgent: {
			name:   generatorAgent,
			model:  g.model,
			config: g.baseCfg,
		},
	}

	_, err = cfg.Stream.CreateStream(cfg.Context, jetstream.StreamConfig{
		Name:        AgentStream,
		Subjects:    []string{"agent.context.*"},
		Retention:   jetstream.WorkQueuePolicy,
		Storage:     jetstream.FileStorage,
		MaxAge:      jsMsgsMaxAge,
		MaxMsgSize:  1 * 1024 * 1024,
		DenyDelete:  true,
		DenyPurge:   false,
		AllowRollup: false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build new agent: %w", err)
	}

	return &Agent{
		agents:    amap,
		gc:        gc,
		functions: fmap,
		logger:    log,
		store:     cfg.Store,
		js:        cfg.Stream,
	}, nil
}

const (
	searcherAgent  agentName     = "searcher"
	generatorAgent agentName     = "generator"
	jsMsgsMaxAge   time.Duration = 24 * time.Hour
)

type agentName string

type agentProfile struct {
	name   agentName
	model  geminiModel
	config *genai.GenerateContentConfig
}

type searcherConfig struct {
	pipe   *rag.Pipeline
	logger *slog.Logger
}

type searcher struct {
	model   geminiModel
	baseCfg *genai.GenerateContentConfig
}

type generator struct {
	model   geminiModel
	baseCfg *genai.GenerateContentConfig
}

func buildSearcher(cfg searcherConfig) *searcher {
	fsearch := buildFunctionSearch(cfg.logger)

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
- When asked a question about the Quran, you must **only answer based on the retrieved documents provided in the conversation history**. If the documents do **not** sufficiently answer the question, initiate a tool call using the 'Search()' function.
- Never guess or answer without evidence. If unsure, search.

## 🧠 Prompt Context

You receive:
- A long conversation history (up to 200,000 tokens) that may include prior questions, retrieved documents, and ayat.
- A new user prompt at the end.

Use all available history to decide whether to answer or search.

## 🛠 Function Usage: Search()

When calling the 'Search()' function, follow the structure defined in its schema.

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
		model:   geminiFlashLiteV2p5,
		baseCfg: generationConfig,
	}
}

func buildGenerator() *generator {
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
		model:   geminiFlashLiteV2p5,
		baseCfg: generationConfig,
	}
}

func (a *Agent) getProfile(ag agentName) (*agentProfile, error) {
	profile, ok := a.agents[ag]
	if !ok {
		return nil, fmt.Errorf("unknown agent: %s", ag)
	}

	return profile, nil
}

func (a *Agent) getFunction(fn functionName) (function, error) {
	function, ok := a.functions[fn]
	if !ok {
		return nil, fmt.Errorf("unknown function: %s", fn)
	}

	return function, nil
}
