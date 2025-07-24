package agent

import (
	"context"

	"github.com/awbalessa/shaikh/apps/server/internal/rag"
	"google.golang.org/genai"
)

type ToolSearch struct {
	Name        ToolName
	Declaration *genai.FunctionDeclaration
}

func (t *ToolSearch) GetName() ToolName {
	return t.Name
}

func (t *ToolSearch) Call(ctx context.Context, json []byte) ([]rag.SearchResult, error) {
	// parse json input into the thing and call it. also log and whatev.
}
