package agent

import (
	"context"
	"fmt"
	"iter"
	"log/slog"
	"strings"
	"time"

	"google.golang.org/genai"
)

func (a *Agent) Ask(
	ctx context.Context,
	prompt string,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		const method = "Ask"
		start := time.Now()
		cc, win, err := a.getContext(ctx)
		if err != nil {
			yield("", err)
			return
		}

		userIn := genai.NewPartFromText(prompt)
		var fnOut *genai.Part
		var modelOut strings.Builder

		log := a.logger.With(
			slog.String("method", method),
			slog.String("userID", cc.UserID.String()),
			slog.String("sessionID", cc.SessionID.String()),
			slog.String("created_at", cc.CreatedAt.Format(time.RFC822)),
			slog.String("updated_at", cc.UpdatedAt.Format(time.RFC822)),
			slog.Int("turn", cc.Window.turns+1),
		)

		log.DebugContext(ctx, "asking agent...")
		gotFirst := false
		var ttft time.Duration

		for resp, err := range a.ask(ctx, win, userIn, &fnOut) {
			if err != nil {
				yield("", err)
				return
			}

			if !gotFirst {
				ttft = time.Since(start)
				gotFirst = true
			}
			modelOut.WriteString(resp)
			if !yieldOk(ctx, yield, resp) {
				return
			}
		}

		totalTime := time.Since(start)

		log.With(
			slog.String("ttft", ttft.String()),
			slog.String("total_time", totalTime.String()),
		).DebugContext(ctx, "response recieved: updating context...")

		modelOutPart := genai.NewPartFromText(modelOut.String())
		lastInteraction := &Interaction{
			Input: inputPrompt{
				FunctionResponse: fnOut,
				UserInput:        userIn,
			},
			ModelOutput: modelOutPart,
			TurnNumber:  cc.Window.turns + 1,
		}
		cc.Window.history = append(cc.Window.history, lastInteraction)

		if err = a.setContext(ctx, cc, lastInteraction); err != nil {
			yield("", err)
			return
		}

		log.DebugContext(ctx, "context updated: returning...")
	})
}

func (a *Agent) ask(
	ctx context.Context,
	win []*genai.Content,
	prompt *genai.Part,
	fnRes **genai.Part,
) iter.Seq2[string, error] {
	return iter.Seq2[string, error](func(yield func(string, error) bool) {
		const method = "ask"
		log := a.logger.With(
			"method", method,
		)
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
					log.With(
						"name", part.FunctionCall.Name,
					).DebugContext(ctx, "agent called function")
					fnResponse, err := a.handleFunctionCall(
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

					*fnRes = fnResponse
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
	const method = "handleFunctionCall"
	log := a.logger.With(
		"method", method,
	)

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
		log.With(
			"name", fnCall.Name,
			"args", fnCall.Args,
		).ErrorContext(ctx, "failed to handle function")
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
		if err != nil {
			return fnPart, err
		}

		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			continue
		}

		for _, part := range resp.Candidates[0].Content.Parts {
			if !yieldOk(ctx, yield, part.Text) {
				return fnPart, nil
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
