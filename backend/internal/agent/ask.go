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
// 	return iter.Seq2[string, error](func(yield func(string, error) bool) {
// 		testUser := uuid.New()
// 		testSesh := uuid.New()

// 		key := createContextCacheKey(testUser, testSesh)
// 		// getting: pull from gcc. if miss, pull from fly, if miss, pull from pg, if miss, create empty to pass to ask().
// 		// setting: set to gcc and to fly. gcc first to store gcc name in fly, then pass msg to nats broker to sync with postgres a bit later. use defer() a bunch to do stuff at the end of the function.
// 	})
// }

func (a *Agent) ask(
	ctx context.Context,
	prompt string,
	cw []*genai.Content,
	fnResOut **genai.Part,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		prof, err := a.getProfile(searcherAgent)
		if err != nil {
			yield("", err)
			return
		}

		full := append(cw, &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				genai.NewPartFromText(prompt),
			},
		})

		for resp, err := range a.gc.client.Models.GenerateContentStream(
			ctx,
			string(prof.model),
			full,
			prof.config,
		) {
			if err != nil {
				yield("", err)
				return
			}

			for _, part := range resp.Candidates[0].Content.Parts {
				if part.FunctionCall != nil {
					fnRes, err := a.handleFunctionCall(ctx, prompt, cw, part.FunctionCall, yield)
					if err != nil {
						yield("", err)
					}
					*fnResOut = fnRes
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
	prompt string,
	cw []*genai.Content,
	functionCall *genai.FunctionCall,
	yield func(string, error) bool,
) (*genai.Part, error) {
	fn, err := a.getFunction(functionName(functionCall.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to handle function %s: %w", fn.name(), err)
	}

	results, err := fn.call(ctx, functionCall.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to handle function %s: %w", fn.name(), err)
	}

	prof, err := a.getProfile(generatorAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to handle function %s: %w", fn.name(), err)
	}

	fnRes := genai.NewPartFromFunctionResponse(
		string(fn.name()),
		results,
	)
	promptPart := genai.NewPartFromText(prompt)

	newContent := &genai.Content{
		Role: genai.RoleUser,
		Parts: []*genai.Part{
			fnRes,
			promptPart,
		},
	}

	full := append(cw, newContent)

	for resp, err := range a.gc.client.Models.GenerateContentStream(
		ctx,
		string(prof.model),
		full,
		prof.config,
	) {
		if err != nil {
			return nil, err
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
