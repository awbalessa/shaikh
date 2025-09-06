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

type GeminiLLMOut struct {
	Out *genai.GenerateContentResponse
}

func (g *GeminiLLMOut) Text() string {
	return g.Out.Text()
}

func (g *GeminiLLMOut) FunctionCall() *dom.LLMFunctionCall {
	fns := g.Out.FunctionCalls()
	if fns == nil {
		return nil
	}
	return &dom.LLMFunctionCall{
		Name: fns[0].Name,
		Args: fns[0].Args,
	}
}

func (g *GeminiLLMOut) MarshalJSON() ([]byte, error) {
	return g.Out.MarshalJSON()
}

func (g *GeminiLLMOut) UnmarshalJSON(data []byte) error {
	return g.Out.UnmarshalJSON(data)
}

func (g *GeminiLLMOut) TokenUsage() (int32, int32) {
	if g.Out.UsageMetadata == nil {
		return 0, 0
	}

	return g.Out.UsageMetadata.PromptTokenCount, g.Out.UsageMetadata.CandidatesTokenCount
}

func (g *GeminiLLMOut) Finish() (string, string) {
	if len(g.Out.Candidates) == 0 && g.Out.Candidates[0].FinishMessage == "" && g.Out.Candidates[0].FinishReason == "" {
		return "", ""
	}
	return g.Out.Candidates[0].FinishMessage, string(g.Out.Candidates[0].FinishReason)
}

func (g *GeminiLLM) Generate(
	ctx context.Context,
	model string,
	window []*dom.LLMContent,
	cfg *dom.LLMGenConfig,
) (dom.LLMOut, error) {
	gContents := toGenaiContents(window)
	gCfg := toGenaiConfig(cfg)
	resp, err := g.Cli.Models.GenerateContent(ctx, model, gContents, gCfg)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return nil, fmt.Errorf("generate: %w", dom.ErrTimeout)
		}
		return nil, fmt.Errorf("generate: %w", dom.ErrInternal)
	}

	return &GeminiLLMOut{
		Out: resp,
	}, nil
}

func (g *GeminiLLM) Stream(
	ctx context.Context,
	model string,
	window []*dom.LLMContent,
	cfg *dom.LLMGenConfig,
	yield func(dom.LLMOut, error) bool,
) *dom.Inference {
	gContents := toGenaiContents(window)
	gCfg := toGenaiConfig(cfg)

	var (
		textBuf       strings.Builder
		fnCall        *dom.LLMFunctionCall
		inTokens      int32
		outTokens     int32
		finishReason  string
		finishMessage string
	)

	for p, err := range g.Cli.Models.GenerateContentStream(ctx, model, gContents, gCfg) {
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				yield(nil, fmt.Errorf("stream: %w", dom.ErrTimeout))
			} else {
				yield(nil, fmt.Errorf("stream: %w", dom.ErrInternal))
			}
			return nil
		}

		if p.UsageMetadata != nil {
			inTokens = p.UsageMetadata.PromptTokenCount
			outTokens = p.UsageMetadata.CandidatesTokenCount
		}

		if len(p.Candidates) > 0 {
			if r := p.Candidates[0].FinishReason; r != "" {
				finishReason = string(r)
			}
			if m := p.Candidates[0].FinishMessage; m != "" {
				finishMessage = m
			}
		}

		if len(p.Candidates) > 0 && p.Candidates[0].Content != nil {
			for _, part := range p.Candidates[0].Content.Parts {
				if part.Text != "" {
					textBuf.WriteString(part.Text)
				}
				if part.FunctionCall != nil && fnCall == nil {
					fnCall = &dom.LLMFunctionCall{
						Name: part.FunctionCall.Name,
						Args: part.FunctionCall.Args,
					}
				}

				if !yield(&GeminiLLMOut{Out: p}, nil) {
					return nil
				}
			}
		}
	}

	return &dom.Inference{
		Output: &dom.LLMOutput{
			Text:         textBuf.String(),
			FunctionCall: fnCall,
		},
		InputTokens:   inTokens,
		OutputTokens:  outTokens,
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
