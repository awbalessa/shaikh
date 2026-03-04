package gemini

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/awbalessa/shaikh/api/internal/app/ai"
	"google.golang.org/genai"
)

type Model struct {
	client *genai.Client
	id     string
}

func NewModel(client *genai.Client, id string) *Model {
	return &Model{client: client, id: id}
}

func NewClient(ctx context.Context, key string) (*genai.Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: key,
	})
	if err != nil {
		return nil, fmt.Errorf("new gemini client: %w", err)
	}
	return client, nil
}

func (m *Model) ID() string       { return m.id }
func (m *Model) Provider() string { return "google" }

func (m *Model) Generate(ctx context.Context, call ai.CallOptions) (ai.GenerateResult, error) {
	return ai.GenerateResult{}, fmt.Errorf("not implemented")
}

func (m *Model) Stream(ctx context.Context, call ai.CallOptions) (ai.StreamResult, error) {
	contents, config := convertCallOptions(call)

	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan streamResult, 16)

	go func() {
		defer close(ch)
		runStream(ctx, m.client, m.id, contents, config, ch)
	}()

	return ai.StreamResult{Stream: &stream{ch: ch, cancel: cancel}}, nil
}

type streamResult struct {
	event ai.Event
	err   error
}

type stream struct {
	ch     <-chan streamResult
	cancel context.CancelFunc
}

func (s *stream) Recv() (ai.Event, error) {
	r, ok := <-s.ch
	if !ok {
		return ai.Event{}, io.EOF
	}
	return r.event, r.err
}

func (s *stream) Close() error {
	s.cancel()
	return nil
}

func runStream(ctx context.Context, client *genai.Client, modelID string, contents []*genai.Content, config *genai.GenerateContentConfig, ch chan<- streamResult) {
	send := func(e ai.Event) bool {
		select {
		case ch <- streamResult{event: e}:
			return true
		case <-ctx.Done():
			return false
		}
	}

	send(ai.Event{Type: ai.EventStreamStart})

	var textBlockID string
	var hasTextBlock bool
	var lastUsage *genai.GenerateContentResponseUsageMetadata
	var lastFinishReason genai.FinishReason

	for chunk, err := range client.Models.GenerateContentStream(ctx, modelID, contents, config) {
		if err != nil {
			ch <- streamResult{err: err}
			return
		}

		if chunk.UsageMetadata != nil {
			lastUsage = chunk.UsageMetadata
		}

		if len(chunk.Candidates) == 0 {
			continue
		}

		cand := chunk.Candidates[0]

		if cand.FinishReason != "" {
			lastFinishReason = cand.FinishReason
		}

		if cand.Content == nil {
			continue
		}

		for _, part := range cand.Content.Parts {
			if part.Text == "" || part.Thought {
				continue
			}

			if !hasTextBlock {
				textBlockID = "0"
				hasTextBlock = true
				if !send(ai.Event{Type: ai.EventTextStart, ID: textBlockID}) {
					return
				}
			}

			if !send(ai.Event{Type: ai.EventTextDelta, ID: textBlockID, Delta: part.Text}) {
				return
			}
		}
	}

	if hasTextBlock {
		send(ai.Event{Type: ai.EventTextEnd, ID: textBlockID})
	}

	usage := &ai.Usage{}
	if lastUsage != nil {
		usage.InputTokens = int(lastUsage.PromptTokenCount)
		usage.OutputTokens = int(lastUsage.CandidatesTokenCount)
		usage.TotalTokens = int(lastUsage.TotalTokenCount)
	}

	send(ai.Event{
		Type:   ai.EventFinish,
		Reason: convertFinishReason(lastFinishReason),
		Usage:  usage,
	})
}

func convertCallOptions(call ai.CallOptions) ([]*genai.Content, *genai.GenerateContentConfig) {
	config := &genai.GenerateContentConfig{}
	if call.MaxOutputTokens != nil {
		config.MaxOutputTokens = *call.MaxOutputTokens
	}
	if call.Temperature != nil {
		config.Temperature = call.Temperature
	}
	if call.PresencePenalty != nil {
		config.PresencePenalty = call.PresencePenalty
	}
	if call.FrequencyPenalty != nil {
		config.FrequencyPenalty = call.FrequencyPenalty
	}

	if len(call.Tools) > 0 && call.ToolChoice != nil {
		config.Tools, config.ToolConfig = convertTools(call.Tools, call.ToolChoice)
	}

	var contents []*genai.Content
	for _, m := range call.Prompt {
		if m.Role == ai.RoleSystem {
			if config.SystemInstruction == nil {
				config.SystemInstruction = &genai.Content{}
			}
			config.SystemInstruction.Parts = append(config.SystemInstruction.Parts, convertSystemMessage(m.Parts)...)
			continue
		}
		contents = append(contents, &genai.Content{
			Role:  convertRole(m.Role),
			Parts: convertParts(m.Parts),
		})
	}

	return contents, config
}

func convertTools(tools []*ai.Tool, toolChoice *ai.ToolChoice) ([]*genai.Tool, *genai.ToolConfig) {
	gtool := &genai.Tool{}
	for _, t := range tools {
		gtool.FunctionDeclarations = append(gtool.FunctionDeclarations, &genai.FunctionDeclaration{
			Description:          t.Description,
			Name:                 t.Name,
			ParametersJsonSchema: t.InputSchema,
		})
	}

	config := &genai.ToolConfig{FunctionCallingConfig: &genai.FunctionCallingConfig{}}
	if toolChoice != nil {
		switch toolChoice.Type {
		case ai.ToolChoiceAuto:
			config.FunctionCallingConfig.Mode = genai.FunctionCallingConfigModeAuto
		case ai.ToolChoiceNone:
			config.FunctionCallingConfig.Mode = genai.FunctionCallingConfigModeNone
		case ai.ToolChoiceRequired:
			config.FunctionCallingConfig.Mode = genai.FunctionCallingConfigModeAny
		default:
			config.FunctionCallingConfig.Mode = genai.FunctionCallingConfigModeAny
			config.FunctionCallingConfig.AllowedFunctionNames = append(config.FunctionCallingConfig.AllowedFunctionNames, toolChoice.ToolName)
		}
	}

	return []*genai.Tool{gtool}, config
}

func convertSystemMessage(parts []ai.Part) []*genai.Part {
	var instructions []*genai.Part
	for _, p := range parts {
		if p.Type() != ai.PartText {
			continue
		}
		instructions = append(instructions, genai.NewPartFromText(p.(ai.TextPart).Text))
	}
	return instructions
}

func convertParts(parts []ai.Part) []*genai.Part {
	var gparts []*genai.Part
	for _, p := range parts {
		switch v := p.(type) {
		case ai.TextPart:
			gparts = append(gparts, genai.NewPartFromText(v.Text))
		case ai.ToolCallPart:
			var args map[string]any
			json.Unmarshal(v.Input, &args)
			gparts = append(gparts, &genai.Part{
				FunctionCall: &genai.FunctionCall{
					ID:   v.ToolCallID,
					Name: v.ToolName,
					Args: args,
				},
			})
		case ai.ToolResultPart:
			var response map[string]any
			json.Unmarshal(v.Result, &response)
			gparts = append(gparts, &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					ID:       v.ToolCallID,
					Name:     v.ToolName,
					Response: response,
				},
			})
		case ai.FilePart:
			gparts = append(gparts, genai.NewPartFromBytes(v.Data, v.MediaType))
		}
	}
	return gparts
}

func convertRole(r ai.Role) string {
	switch r {
	case ai.RoleUser:
		return genai.RoleUser
	case ai.RoleTool:
		return genai.RoleUser
	default:
		return genai.RoleModel
	}
}

func convertFinishReason(r genai.FinishReason) ai.FinishReason {
	switch r {
	case genai.FinishReasonStop:
		return ai.FinishReasonStop
	case genai.FinishReasonMaxTokens:
		return ai.FinishReasonLength
	default:
		return ai.FinishReasonOther
	}
}
