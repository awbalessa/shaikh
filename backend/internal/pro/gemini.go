package pro

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/hashicorp/go-retryablehttp"
	"google.golang.org/genai"
)

type GeminiLLM struct {
	Cli *genai.Client
}

func NewGeminiLLM(
	ctx context.Context,
) (*GeminiLLM, error) {
	baseClient := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			TLSHandshakeTimeout: 10 * time.Second,
		},
	}

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = baseClient
	retryClient.RetryMax = 3
	retryClient.CheckRetry = retryablehttp.ErrorPropagatedRetryPolicy
	retryClient.Backoff = retryablehttp.DefaultBackoff

	standard := retryClient.StandardClient()

	cc := &genai.ClientConfig{
		HTTPClient: standard,
	}

	gc, err := genai.NewClient(ctx, cc)
	if err != nil {
		return nil, fmt.Errorf("new gemini client: %w", dom.ErrUnavailable)
	}

	return &GeminiLLM{
		Cli: gc,
	}, nil
}

func (g *GeminiLLM) Generate(
	ctx context.Context,
	model string,
	window []*dom.LLMContent,
	cfg *dom.LLMGenConfig,
	format dom.LLMResponseSchema,
) (*dom.LLMContentResult, error) {
	gWindow := toGenaiContents(window)
	gCfg := toGenaiConfig(cfg)

	resp, err := g.Cli.Models.GenerateContent(ctx, model, gWindow, gCfg)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return nil, fmt.Errorf("gemini generate: %w", dom.ErrTimeout)
		}
		return nil, fmt.Errorf("gemini generate: %w", dom.ErrInternal)
	}

	var returned = &dom.LLMContentResult{}
	if format == dom.ResponseJson {
		data, err := resp.MarshalJSON()
		if err != nil {
			return nil, fmt.Errorf("gemini marshal json: %w", dom.ErrInternal)
		}
		returned.Bytes = data
	} else {
		text := resp.Text()
		returned.Text = &text
	}

	return returned, nil
}

func (g *GeminiLLM) Stream(
	ctx context.Context,
	model string,
	window []*dom.LLMContent,
	cfg *dom.LLMGenConfig,
	yield func(*dom.LLMPart, error) bool,
) *dom.LLMGenResult {
	gWindow := toGenaiContents(window)
	gCfg := toGenaiConfig(cfg)

	var str strings.Builder
	var output dom.ModelOutput
	var usage dom.TokenUsage
	var finishMessage string
	var finishReason dom.FinishReason

	stream := g.Cli.Models.GenerateContentStream(ctx, model, gWindow, gCfg)
	for resp, err := range stream {
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				yield(nil, fmt.Errorf("gemini stream: %w", dom.ErrTimeout))
			} else {
				yield(nil, fmt.Errorf("gemini stream: %w", dom.ErrInternal))
			}
			return &dom.LLMGenResult{
				Output:        &output,
				Usage:         &usage,
				FinishReason:  finishReason,
				FinishMessage: finishMessage,
			}
		}
		if len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
			continue
		}

		if r := resp.Candidates[0].FinishReason; r != "" {
			finishReason = dom.FinishReason(r)
		}
		if m := resp.Candidates[0].FinishMessage; m != "" {
			finishMessage = m
		}

		if resp.UsageMetadata != nil {
			if inp := resp.UsageMetadata.PromptTokenCount; inp != 0 {
				usage.InputTokens = inp
			}
			if op := resp.UsageMetadata.CandidatesTokenCount; op != 0 {
				usage.OutputTokens = op
			}
		}

		for _, p := range resp.Candidates[0].Content.Parts {
			part := dom.LLMPart{}
			if p.Text != "" {
				part.Text = p.Text
				str.WriteString(p.Text)
				output.Text = str.String()
			}
			if p.FunctionCall != nil {
				part.FunctionCall = &dom.LLMFunctionCall{
					Name: p.FunctionCall.Name,
					Args: p.FunctionCall.Args,
				}
				output.FunctionCall = part.FunctionCall
			}
			if !yield(&part, nil) {
				return &dom.LLMGenResult{
					Output:        &output,
					Usage:         &usage,
					FinishReason:  finishReason,
					FinishMessage: finishMessage,
				}
			}
		}
	}

	return &dom.LLMGenResult{
		Output:        &output,
		Usage:         &usage,
		FinishReason:  finishReason,
		FinishMessage: finishMessage,
	}
}

func (g *GeminiLLM) CountTokens(
	ctx context.Context,
	model string,
	window []*dom.LLMContent,
	cfg *dom.LLMCountConfig,
) (int32, error) {
	gWindow := toGenaiContents(window)
	cCfg := &genai.CountTokensConfig{
		SystemInstruction: toGenaiContent(cfg.System),
		Tools:             toGenaiTools(cfg.Tools),
	}
	resp, err := g.Cli.Models.CountTokens(ctx, string(model), gWindow, cCfg)
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
			return 0, fmt.Errorf("gemini count tokens: %w", dom.ErrTimeout)
		}
		return 0, fmt.Errorf("gemini count tokens: %w", dom.ErrInternal)
	}
	return resp.TotalTokens, nil
}

func toGenaiContents(win []*dom.LLMContent) []*genai.Content {
	out := make([]*genai.Content, 0, len(win))
	for _, c := range win {
		out = append(out, toGenaiContent(c))
	}
	return out
}

func toGenaiContent(c *dom.LLMContent) *genai.Content {
	if c == nil {
		return nil
	}
	parts := []*genai.Part{}
	for _, p := range c.Parts {
		switch {
		case p.Text != "":
			parts = append(parts, genai.NewPartFromText(p.Text))
		case p.FunctionCall != nil:
			parts = append(parts, &genai.Part{
				FunctionCall: &genai.FunctionCall{
					Name: p.FunctionCall.Name,
					Args: p.FunctionCall.Args,
				},
			})
		case p.FunctionResponse != nil:
			parts = append(parts, &genai.Part{
				FunctionResponse: &genai.FunctionResponse{
					Name:     p.FunctionResponse.Name,
					Response: p.FunctionResponse.Content,
				},
			})
		}
	}
	return &genai.Content{Role: string(c.Role), Parts: parts}
}

func toGenaiConfig(cfg *dom.LLMGenConfig) *genai.GenerateContentConfig {
	if cfg == nil {
		return nil
	}
	return &genai.GenerateContentConfig{
		SystemInstruction: toGenaiContent(cfg.SystemInstructions),
		Temperature:       &cfg.Temperature,
		CandidateCount:    cfg.CandidateCount,
		Tools:             toGenaiTools(cfg.Tools),
	}
}

func toGenaiTools(tools []*dom.LLMFunctionDecl) []*genai.Tool {
	if len(tools) == 0 {
		return nil
	}
	out := []*genai.Tool{{
		FunctionDeclarations: []*genai.FunctionDeclaration{},
	}}
	for _, t := range tools {
		out[0].FunctionDeclarations = append(out[0].FunctionDeclarations,
			&genai.FunctionDeclaration{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  toGenaiSchema(t.Parameters),
			},
		)
	}
	return out
}

func toGenaiSchema(s *dom.LLMSchema) *genai.Schema {
	if s == nil {
		return nil
	}

	items := toGenaiSchema(s.Items)
	props := make(map[string]*genai.Schema, len(s.Properties))
	for k, v := range s.Properties {
		props[k] = toGenaiSchema(v)
	}

	return &genai.Schema{
		Title:       s.Title,
		Description: s.Description,
		Type:        genai.Type(s.Type),
		Enum:        s.Enum,
		Example:     s.Example,
		Format:      s.Format,
		Required:    s.Required,
		Properties:  props,
		Items:       items,
		MinItems:    s.MinItems,
		MaxItems:    s.MaxItems,
		Minimum:     s.Minimum,
		Maximum:     s.Maximum,
	}
}

func (g *GeminiLLM) Ping(ctx context.Context) error {
	return dom.ErrNotPingable
}

func (g *GeminiLLM) Name() string {
	return "LLM"
}

func (g *GeminiLLM) Close() error {
	if tr, ok := g.Cli.ClientConfig().HTTPClient.Transport.(*http.Transport); ok {
		tr.CloseIdleConnections()
	}
	return nil
}
