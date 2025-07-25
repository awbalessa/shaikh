package agent

import (
	"context"
	"log/slog"

	"google.golang.org/genai"
)

const (
	Search FunctionName = "SearchChunks"
)

type FunctionName string

type Function interface {
	GetName() FunctionName
	Call(ctx context.Context, json []byte) (any, error)
}

type Temperature float32

type AgentConfig struct {
	Gc           *GeminiClient
	Model        GeminiModel
	Tools        map[FunctionName]Function
	Instructions genai.Content
	Temperature  Temperature
}

type Agent struct {
	gc     *GeminiClient
	model  GeminiModel
	tools  map[FunctionName]Function
	logger *slog.Logger
}

// func NewAgent(gc *GeminiClient, model GeminiModel, tools []Function) *Agent {
// 	for _, tool := range tools {
// 		if tool.GetName() == Search {

// 		}
// 	}
// }
