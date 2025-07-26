package agent

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

const (
	Router AgentName = "router"
)

type AgentName string

type Agent interface {
	GetName() AgentName
	GetTool(toolName) (tool, error)
}

type AgentRouter struct {
	name             AgentName
	tools            map[toolName]tool
	gc               *geminiClient
	model            geminiModel
	generationConfig *genai.GenerateContentConfig
}

func (r *AgentRouter) GetName() AgentName {
	return r.name
}

func (r *AgentRouter) GetTool(t toolName) (tool, error) {
	tool, ok := r.tools[t]
	if !ok {
		return nil, fmt.Errorf("tool %s does not exist", string(t))
	}

	return tool, nil
}

func BuildRouter(ctx context.Context) (*AgentRouter, error) {
	gc, err := newGeminiClient(ctx, geminiClientConfig{
		maxRetries:     geminiMaxRetriesThree,
		timeout:        geminiTimeoutFifteenSeconds,
		gcpProjectID:   gcpProjectID,
		geminiBackend:  geminiBackend,
		geminiLocation: geminiLocationGlobal,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to build router agent: %w", err)
	}

	return nil, nil
}

const (
	RAG    toolName     = "RAG"
	Search functionName = "Search"
)

type toolName string
type functionName string

type tool interface {
	getName() toolName
	getFunction(functionName) (function, error)
}

type function interface {
	getName() functionName
	call(ctx context.Context, jsonBytes []byte) (any, error)
}
