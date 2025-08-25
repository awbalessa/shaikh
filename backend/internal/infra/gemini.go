package infra

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
	geminiFlashLiteV2p5         string        = "gemini-2.5-flash-lite"
	geminiFlashV2p5             string        = "gemini-2.5-flash"
	geminiTimeoutFifteenSeconds time.Duration = 15 * time.Second
	geminiMaxRetriesThree       int           = 3
	geminiMaxIdleConns          int           = 100
	geminiMaxIdleConnsPerHost   int           = 10
	geminiIdleConnTimeout       time.Duration = 90 * time.Second
	geminiDialContextTimeout    time.Duration = 5 * time.Second
	geminiDialContextKeepAlive  time.Duration = 30 * time.Second
	geminiTLSHandshakeTimeout   time.Duration = 10 * time.Second
	geminiTemperatureZero       float32       = 0
	geminiResponseMimeJSON      string        = "application/json"
)

type GeminiClient struct {
	Cli *genai.Client
	Log *slog.Logger
}

func NewGeminiClient(ctx context.Context, maxRetries int, timeout time.Duration, log *slog.Logger) (*GeminiClient, error) {
	baseClient := &http.Client{
		Timeout: timeout,
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

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = baseClient
	retryClient.Logger = log
	retryClient.RetryMax = geminiMaxRetriesThree
	retryClient.CheckRetry = retryablehttp.ErrorPropagatedRetryPolicy
	retryClient.Backoff = retryablehttp.DefaultBackoff

	standard := retryClient.StandardClient()

	cc := &genai.ClientConfig{
		HTTPClient: standard,
	}

	gc, err := genai.NewClient(ctx, cc)
	if err != nil {
		log.With(
			"err", err,
		).ErrorContext(ctx, "failed to create new gemini client")
		return nil, fmt.Errorf("failed to create new gemini client: %w", err)
	}

	return &GeminiClient{
		Cli: gc,
		Log: log,
	}, nil
}
