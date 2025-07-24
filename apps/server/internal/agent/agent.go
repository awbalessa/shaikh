package agent

import "context"

const (
	Search ToolName = "Search"
)

type ToolName string

type Tool interface {
	GetName() ToolName
	Call(ctx context.Context, json []byte) (any, error)
}

type Agent struct {
	gc    *GeminiClient
	model GeminiModel
	tools []Tool
}

func NewAgent(gc *GeminiClient, model GeminiModel, tools []Tool) *Agent {
	for _, tool := range tools {
		switch tool.Name() {
		case Search:

		}
	}
}
