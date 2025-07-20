package rag

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/awbalessa/shaikh/apps/server/internal/config"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	GeminiFlashLite           string        = "gemini-2.5-flash-lite-preview-06-17"
	GeminiLocation            string        = "global"
	GCPProjectID              string        = "shaikh-460416"
	VoyageEmbeddings          string        = "voyage-3.5"
	VoyageBaseURL             string        = ""
	VoyageMaxRetries          int           = 3
	VoyageTimeout             time.Duration = 10 * time.Second
	VoyageMaxIdleConns        int           = 100
	VoyageMaxIdleConnsPerHost int           = 10
	VoyageIdleConnTimeout     time.Duration = 90 * time.Second
)

type VoyageClientConfig struct {
	Config     *config.Config
	MaxRetries int
	Timeout    time.Duration
}

type VoyageClient struct {
	Client *retryablehttp.Client
	APIKey string
	Logger *slog.Logger
}

func NewVoyageClient(cfg VoyageClientConfig) *VoyageClient {
	client := &http.Client{
		Timeout: cfg.Timeout,
		Transport: &http.Transport{
			MaxIdleConns:        VoyageMaxIdleConns,
			MaxIdleConnsPerHost: VoyageMaxIdleConnsPerHost,
			IdleConnTimeout:     VoyageIdleConnTimeout,
		},
	}

	logger := slog.Default().With(
		"component", "voyage",
	)

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = client
	retryClient.Logger = logger
	retryClient.RetryMax = cfg.MaxRetries

	return &VoyageClient{
		Client: retryClient,
		APIKey: cfg.Config.VoyageAPIKey,
		Logger: logger,
	}
}
