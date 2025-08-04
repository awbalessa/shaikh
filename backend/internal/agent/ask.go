package agent

import (
	"context"
	"fmt"
	"iter"

	"google.golang.org/genai"
)

// func (a *Agent) Ask(
// 	ctx context.Context,
// 	prompt string,
// ) iter.Seq2[string, error] {

// }

func (a *Agent) ask(
	ctx context.Context,
	sc *sessionContext,
	cw []*genai.Content,
	prompt string,
	fnResOut **genai.Part,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		promptPart := genai.NewPartFromText(prompt)
		parts := []*genai.Part{promptPart}

		prof, fullContext, err := a.applygcc(searcherAgent, sc, cw, parts)
		if err != nil {
			yield("", err)
			return
		}

		for resp, err := range a.gc.client.Models.GenerateContentStream(
			ctx,
			string(prof.model),
			fullContext,
			prof.config,
		) {
			if err != nil {
				yield("", err)
				return
			}

			if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
				continue
			}

			for _, part := range resp.Candidates[0].Content.Parts {
				switch {
				case part.FunctionCall != nil:
					fnRes, err := a.handleFunctionCall(
						ctx,
						sc,
						cw,
						prompt,
						part.FunctionCall,
						yield,
					)
					if err != nil {
						yield("", err)
						return
					}
					*fnResOut = fnRes
					return

				case part.Text != "":
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
	sc *sessionContext,
	cw []*genai.Content,
	prompt string,
	fnCall *genai.FunctionCall,
	yield func(string, error) bool,
) (*genai.Part, error) {
	fn, err := a.getFunction(functionName(fnCall.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to handle function %s: %w", fnCall.Name, err)
	}

	results, err := fn.call(ctx, fnCall.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to handle function %s: %w", fnCall.Name, err)
	}

	fnRes := genai.NewPartFromFunctionResponse(string(fn.name()), results)
	parts := []*genai.Part{fnRes, genai.NewPartFromText(prompt)}

	prof, fullContext, err := a.applygcc(generatorAgent, sc, cw, parts)
	if err != nil {
		return nil, fmt.Errorf("failed to handle function %s: %w", fnCall.Name, err)
	}

	for resp, err := range a.gc.client.Models.GenerateContentStream(
		ctx,
		string(prof.model),
		fullContext,
		prof.config,
	) {
		if err != nil {
			return nil, err
		}
		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			continue
		}
		for _, part := range resp.Candidates[0].Content.Parts {
			if part.Text != "" {
				if !yield(part.Text, nil) {
					return fnRes, nil
				}
			}
		}
	}

	return fnRes, nil
}
