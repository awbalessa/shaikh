package agent

import (
	"context"
	"fmt"
	"iter"
	"strings"

	"google.golang.org/genai"
)

func (a *Agent) Ask(
	ctx context.Context,
	prompt string,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		cc, win, err := a.getContext(ctx)
		if err != nil {
			yield("", err)
			return
		}

		userIn := genai.NewPartFromText(prompt)
		var fnOut *genai.Part
		var modelOut strings.Builder

		for resp, err := range a.ask(ctx, win, userIn, &fnOut) {
			if err != nil {
				yield("", err)
				return
			}
			modelOut.WriteString(resp)
			if !yieldOk(ctx, yield, resp) {
				return
			}
		}

		modelOutPart := genai.NewPartFromText(modelOut.String())
		cc.Window.history = append(cc.Window.history, interaction{
			input: inputPrompt{
				functionResponse: fnOut,
				userInput:        userIn,
			},
			modelOutput: modelOutPart,
		})
	})
}

func (a *Agent) ask(
	ctx context.Context,
	win []*genai.Content,
	prompt *genai.Part,
	fnResOut **genai.Part,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		prof, err := a.getProfile(searcherAgent)
		if err != nil {
			yield("", err)
			return
		}

		full := append(win, &genai.Content{
			Role:  genai.RoleUser,
			Parts: []*genai.Part{prompt},
		})

		for resp, err := range a.gc.client.Models.GenerateContentStream(
			ctx,
			string(prof.model),
			full,
			prof.config,
		) {
			if ctx.Err() != nil {
				yield("", ctx.Err())
				return
			}
			if err != nil {
				yield("", err)
				return
			}
			if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
				continue
			}

			for _, part := range resp.Candidates[0].Content.Parts {
				if ctx.Err() != nil {
					yield("", ctx.Err())
					return
				}

				switch {
				case part.FunctionCall != nil:
					fnRes, err := a.handleFunctionCall(
						ctx,
						win,
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
					if !yieldOk(ctx, yield, part.Text) {
						return
					}
				}
			}
		}
	})
}

func (a *Agent) handleFunctionCall(
	ctx context.Context,
	win []*genai.Content,
	prompt *genai.Part,
	fnCall *genai.FunctionCall,
	yield func(string, error) bool,
) (*genai.Part, error) {
	prof, err := a.getProfile(generatorAgent)
	if err != nil {
		return nil, fmt.Errorf("failed to handle function %s: %w", fnCall.Name, err)
	}

	fn, err := a.getFunction(functionName(fnCall.Name))
	if err != nil {
		return nil, fmt.Errorf("failed to handle function %s: %w", fnCall.Name, err)
	}

	results, err := fn.call(ctx, fnCall.Args)
	if err != nil {
		return nil, fmt.Errorf("failed to handle function %s: %w", fnCall.Name, err)
	}

	fnPart := genai.NewPartFromFunctionResponse(string(fn.name()), results)
	full := append(win, &genai.Content{
		Role:  genai.RoleUser,
		Parts: []*genai.Part{fnPart, prompt},
	})

	for resp, err := range a.gc.client.Models.GenerateContentStream(
		ctx,
		string(prof.model),
		full,
		prof.config,
	) {
		if ctx.Err() != nil {
			return fnPart, ctx.Err()
		}
		if err != nil {
			return fnPart, err
		}
		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			continue
		}
		for _, part := range resp.Candidates[0].Content.Parts {
			if ctx.Err() != nil {
				return fnPart, ctx.Err()
			}
			if part.Text != "" {
				if !yieldOk(ctx, yield, part.Text) {
					return fnPart, nil
				}
			}
		}
	}

	return fnPart, nil
}

func yieldOk(ctx context.Context, yield func(string, error) bool, s string) bool {
	if ctx.Err() != nil {
		yield("", ctx.Err())
		return false
	}
	if !yield(s, nil) {
		return false
	}
	return true
}
