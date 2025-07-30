package agent

import (
	"context"
	"crypto/rand"
	"fmt"
	"iter"
	"time"

	"github.com/google/uuid"
	"github.com/oklog/ulid"
	"google.golang.org/genai"
)

func (a *Agent) Ask(
	ctx context.Context,
	prompt string,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		t := time.Now().UTC()
		entropy := ulid.Monotonic(rand.Reader, 0)
		testUser := uuid.New()
		testSesh, err := ulid.New(ulid.Timestamp(t), entropy)
		if err != nil {
			yield("", err)
			return
		}
		key := createContextCacheKey(testUser, testSesh)
		// getting: pull from gcc. if miss, pull from fly, if miss, pull from pg, if miss, create empty to pass to ask().
		// setting: set to gcc and to fly. gcc first to store gcc name in fly, then pass msg to nats broker to sync with postgres a bit later. use defer() a bunch to do stuff at the end of the function.
	})
}

func (a *Agent) ask(
	ctx context.Context,
	prompt string,
	cw []*genai.Content,
	fnResOut **genai.Part,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		full := append(cw, &genai.Content{
			Role: genai.RoleUser,
			Parts: []*genai.Part{
				genai.NewPartFromText(prompt),
			},
		})

		for resp, err := range a.searcher.gc.client.Models.GenerateContentStream(
			ctx,
			string(a.generator.model),
			full,
			a.generator.baseCfg,
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
	fn, ok := a.functions[functionName(functionCall.Name)]
	if !ok {
		return nil, fmt.Errorf("unknown function: %s", functionCall.Name)
	}

	results, err := fn.call(ctx, functionCall.Args)
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

	for resp, err := range a.generator.gc.client.Models.GenerateContentStream(
		ctx,
		string(a.generator.model),
		full,
		a.generator.baseCfg,
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
