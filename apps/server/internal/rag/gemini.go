package rag

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"google.golang.org/genai"
)

const (
	GeminiFlashLiteV2p5        string        = "gemini-2.5-flash-lite-preview-06-17"
	GeminiLocation             string        = "global"
	GeminiBackend              genai.Backend = genai.BackendVertexAI
	GCPProjectID               string        = "shaikh-460416"
	GeminiTimeout              time.Duration = 15 * time.Second
	GeminiMaxRetries           int           = 3
	GeminiMaxIdleConns         int           = 100
	GeminiMaxIdleConnsPerHost  int           = 10
	GeminiIdleConnTimeout      time.Duration = 90 * time.Second
	GeminiDialContextTimeout   time.Duration = 5 * time.Second
	GeminiDialContextKeepAlive time.Duration = 30 * time.Second
	GeminiTLSHandshakeTimeout  time.Duration = 10 * time.Second
)

type GeminiClient struct {
	client *genai.Client
	logger *slog.Logger
}

type GeminiClientConfig struct {
	MaxRetries     int
	Timeout        time.Duration
	GCPProjectID   string
	GeminiBackend  genai.Backend
	GeminiLocation string
}

func NewGeminiClient(ctx context.Context, gcc *GeminiClientConfig) (*GeminiClient, error) {
	baseClient := &http.Client{
		Timeout: gcc.Timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   GeminiDialContextTimeout,
				KeepAlive: GeminiDialContextKeepAlive,
			}).DialContext,
			MaxIdleConns:        GeminiMaxIdleConns,
			MaxIdleConnsPerHost: GeminiMaxIdleConnsPerHost,
			IdleConnTimeout:     GeminiIdleConnTimeout,
			TLSHandshakeTimeout: GeminiTLSHandshakeTimeout,
		},
	}

	logger := slog.Default().With(
		"component", "gemini",
	)

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = baseClient
	retryClient.Logger = logger
	retryClient.RetryMax = GeminiMaxRetries

	standard := retryClient.StandardClient()

	cc := &genai.ClientConfig{
		Backend:    gcc.GeminiBackend,
		Project:    gcc.GCPProjectID,
		Location:   gcc.GeminiLocation,
		HTTPClient: standard,
	}

	gc, err := genai.NewClient(ctx, cc)
	if err != nil {
		logger.With(
			"err", err,
		).ErrorContext(ctx, "failed to create new gemini client")
		return nil, fmt.Errorf("failed to create new gemini client: %w", err)
	}

	return &GeminiClient{
		client: gc,
		logger: logger,
	}, nil
}
