package agent

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
	geminiFlashLiteV2p5         geminiModel      = "gemini-2.5-flash-lite-preview-06-17"
	geminiFlashV2p5             geminiModel      = "gemini-2.5-flash"
	geminiLocationUSCentral1    string           = "us-central1"
	geminiLocationGlobal        string           = "global"
	geminiBackend               genai.Backend    = genai.BackendVertexAI
	gcpProjectID                string           = "shaikh-460416"
	geminiTimeoutFifteenSeconds time.Duration    = 15 * time.Second
	geminiMaxRetriesThree       int              = 3
	geminiMaxIdleConns          int              = 100
	geminiMaxIdleConnsPerHost   int              = 10
	geminiIdleConnTimeout       time.Duration    = 90 * time.Second
	geminiDialContextTimeout    time.Duration    = 5 * time.Second
	geminiDialContextKeepAlive  time.Duration    = 30 * time.Second
	geminiTLSHandshakeTimeout   time.Duration    = 10 * time.Second
	geminiTemperatureZero       temperature      = 0
	geminiResponseMimeJSON      responseMimeType = "application/json"
)

type geminiModel string
type temperature float32
type responseMimeType string

type geminiClient struct {
	client *genai.Client
	logger *slog.Logger
}

type geminiClientConfig struct {
	maxRetries     int
	timeout        time.Duration
	gcpProjectID   string
	geminiBackend  genai.Backend
	geminiLocation string
}

func newGeminiClient(ctx context.Context, gcc geminiClientConfig) (*geminiClient, error) {
	baseClient := &http.Client{
		Timeout: gcc.timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   geminiDialContextTimeout,
				KeepAlive: geminiDialContextKeepAlive,
			}).DialContext,
			MaxIdleConns:        geminiMaxIdleConns,
			MaxIdleConnsPerHost: geminiMaxIdleConnsPerHost,
			IdleConnTimeout:     geminiIdleConnTimeout,
			TLSHandshakeTimeout: geminiTLSHandshakeTimeout,
		},
	}

	logger := slog.Default().With(
		"component", "gemini",
	)

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = baseClient
	retryClient.Logger = logger
	retryClient.RetryMax = geminiMaxRetriesThree

	standard := retryClient.StandardClient()

	cc := &genai.ClientConfig{
		Backend:    gcc.geminiBackend,
		Project:    gcc.gcpProjectID,
		Location:   gcc.geminiLocation,
		HTTPClient: standard,
	}

	gc, err := genai.NewClient(ctx, cc)
	if err != nil {
		logger.With(
			"err", err,
		).ErrorContext(ctx, "failed to create new gemini client")
		return nil, fmt.Errorf("failed to create new gemini client: %w", err)
	}

	return &geminiClient{
		client: gc,
		logger: logger,
	}, nil
}
