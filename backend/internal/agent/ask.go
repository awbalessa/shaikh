package agent

import (
	"context"
	"fmt"
	"iter"

	"github.com/awbalessa/shaikh/backend/internal/rag"
	"google.golang.org/genai"
)

// app.Agent.Ask(ctx context.Context, prompt string) iter.Seq2
// ...
// call searcher. If tool call gotten with name, call searcher.callFn(fn fname), which will receive func

// func (a *Agent) Ask(
// 	ctx context.Context,
// 	prompt string,
// ) iter.Seq2[string, error] {
// 	return iter.Seq2[string, error](func(yield func (string, errror) bool)) {
// 		userID := "testuser"
// 		sessionID := "testsession"
// 	}
// }

func (a *Agent) ask(
	ctx context.Context,
	prompt []*genai.Content,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		for resp, err := range a.searcher.gc.client.Models.GenerateContentStream(
			ctx,
			string(a.generator.model),
			prompt,
			a.generator.baseCfg,
		) {
			if err != nil {
				yield("", err)
				return
			}

			for _, part := range resp.Candidates[0].Content.Parts {
				if part.FunctionCall != nil {
					err := a.handleFunctionCall(ctx, prompt, part.FunctionCall, yield)
					if err != nil {
						yield("", err)
					}
					return
				}

				if part.Text != "" {
					if !yield(part.Text, nil) {
						return
					}
				}
			}
		}
	})
}

func (a *Agent) handleFunctionCall(
	ctx context.Context,
	prompt []*genai.Content,
	functionCall *genai.FunctionCall,
	yield func(string, error) bool,
) error {
	switch functionCall.Name {
	case string(search):
		return a.handleSearch(ctx, prompt, functionCall.Args, yield)

	default:
		return fmt.Errorf("unknown function: %s", functionCall.Name)
	}
}

func (a *Agent) handleSearch(
	ctx context.Context,
	prompt []*genai.Content,
	args map[string]any,
	yield func(string, error) bool,
) error {
	fn, err := typeFn[[]rag.SearchResult](a.searcher.functions, search)
	if err != nil {
		return err
	}

	results, err := fn.call(ctx, args)
	if err != nil {
		return err
	}

	var parts []*genai.Part
	for _, r := range results {
		parts = append(parts, genai.NewPartFromText(r.EmbeddedChunk))
	}

	contextWithResults := &genai.Content{
		Role:  genai.RoleUser,
		Parts: parts,
	}

	newPrompt := append(prompt, contextWithResults)

	for resp, err := range a.generator.gc.client.Models.GenerateContentStream(
		ctx,
		string(a.generator.model),
		newPrompt,
		a.generator.baseCfg,
	) {
		if err != nil {
			return err
		}
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				if !yield(part.Text, nil) {
					return nil
				}
			}
		}
	}

	return nil
}
