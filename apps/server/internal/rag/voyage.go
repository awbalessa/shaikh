package rag

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/awbalessa/shaikh/apps/server/internal/config"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pgvector/pgvector-go"
)

const (
	VoyageBaseURL             string                       = "https://api.voyageai.com/v1"
	VoyageTimeout             time.Duration                = 10 * time.Second
	VoyageMaxRetries          int                          = 3
	VoyageMaxIdleConns        int                          = 100
	VoyageMaxIdleConnsPerHost int                          = 10
	VoyageIdleConnTimeout     time.Duration                = 90 * time.Second
	VoyageEmbedV3p5           EmbeddingModel               = "voyage-3.5"
	InputTypeQuery            EmbeddingInputType           = "query"
	InputTypeDocument         EmbeddingInputType           = "document"
	OutputDimension1024       EmbeddingOutputDimension     = 1024
	OutputDimensionTypeFloat  EmbeddingOutputDimensionType = "float"
	VoyageRerankV2            RerankingModel               = "rerank-2"
	Top5Documents             TopK                         = 5
	Top10Documents            TopK                         = 10
	Top20Documents            TopK                         = 20
)

type VoyageClientConfig struct {
	Config     *config.Config
	MaxRetries int
	Timeout    time.Duration
}

type VoyageClient struct {
	client *retryablehttp.Client
	apiKey string
	logger *slog.Logger
}

type EmbeddingModel string
type EmbeddingInputType string
type EmbeddingOutputDimension int
type EmbeddingOutputDimensionType string
type EmbeddingEncodingFormat *string
type Embedding1024 [1024]float32

type VoyageEmbeddingRequest struct {
	Input               []string                     `json:"input"`
	Model               EmbeddingModel               `json:"model"`
	InputType           EmbeddingInputType           `json:"input_type"`
	Truncation          bool                         `json:"truncation"`
	OutputDimension     EmbeddingOutputDimension     `json:"output_dimension"`
	OutputDimensionType EmbeddingOutputDimensionType `json:"output_dtype"`
}

type VoyageEmbedding struct {
	ObjectType string        `json:"object"`
	Embedding  Embedding1024 `json:"embedding"`
	Index      int           `json:"index"`
}

type Usage struct {
	TokensUsed int `json:"total_tokens"`
}

type VoyageEmbeddingResponse struct {
	ObjectType string            `json:"object"`
	Data       []VoyageEmbedding `json:"data"`
	Model      EmbeddingModel    `json:"model"`
	Usage      Usage             `json:"usage"`
}

type RerankingModel string
type TopK int

type VoyageRerankingRequest struct {
	Query           string         `json:"query"`
	Documents       []string       `json:"documents"`
	Model           RerankingModel `json:"model"`
	TopK            TopK           `json:"top_k"`
	ReturnDocuments bool           `json:"return_documents"`
	Truncation      bool           `json:"truncation"`
}

type VoyageReranking struct {
	Index          int `json:"index"`
	RelevanceScore int `json:"relevance_score"`
}

type VoyageRerankingResponse struct {
	Object string            `json:"object"`
	Data   []VoyageReranking `json:"data"`
	Model  string            `json:"model"`
	Usage  Usage             `json:"usage"`
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
		client: retryClient,
		apiKey: cfg.Config.VoyageAPIKey,
		logger: logger,
	}
}

func (vc *VoyageClient) EmbedQuery(
	ctx context.Context,
	texts []string,
) ([]pgvector.Vector, error) {
	const method = "EmbedQuery"
	reqBody := VoyageEmbeddingRequest{
		Input:               texts,
		Model:               VoyageEmbedV3p5,
		InputType:           InputTypeQuery,
		Truncation:          false,
		OutputDimension:     OutputDimension1024,
		OutputDimensionType: OutputDimensionTypeFloat,
	}

	log := vc.logger.With(
		slog.String("method", method),
		slog.String("model", string(reqBody.Model)),
		slog.String("input_type", string(reqBody.InputType)),
		slog.Int("input_len", len(texts)),
	)

	log.InfoContext(ctx, "sending voyage embedding request...")

	payload, err := json.Marshal(reqBody)
	if err != nil {
		log.With("err", err).ErrorContext(
			ctx,
			"failed to marshal embedding request",
		)
		return nil, fmt.Errorf("failed to marshal embedding request: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(
		ctx,
		http.MethodPost,
		VoyageBaseURL+"/embeddings",
		payload,
	)
	if err != nil {
		log.With("err", err).ErrorContext(
			ctx,
			"failed to create request with context",
		)
		return nil, fmt.Errorf("failed to create request with context: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+vc.apiKey)

	start := time.Now()
	resp, err := vc.client.Do(req)
	if err != nil {
		log.With(
			slog.Any("err", err),
			slog.String("status", resp.Status),
		).ErrorContext(
			ctx,
			"embedding request failed",
		)
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.With(
			slog.Any("err", err),
			slog.String("status", resp.Status),
		).ErrorContext(
			ctx,
			"voyage returned non-200 status",
		)
		return nil, fmt.Errorf("voyage returned non-200 status: %w", err)
	}

	duration := time.Since(start)
	log.With(
		slog.Int64("duration_ms", duration.Milliseconds()),
	).InfoContext(ctx, "voyage response received: decoding response...")

	var result VoyageEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.With("err", err).ErrorContext(
			ctx,
			"failed to decode voyage response",
		)
		return nil, fmt.Errorf("failed to decode voyage response: %w", err)
	}

	vectors := make([]pgvector.Vector, len(result.Data))
	for i, item := range result.Data {
		vectors[i] = pgvector.NewVector(
			item.Embedding[:],
		)
	}

	duration = time.Since(start)
	log.With(
		slog.Int64("duration_ms", duration.Milliseconds()),
		slog.Int("embedding_count", len(vectors)),
	).InfoContext(ctx, "decoded response: returning...")

	return vectors, nil
}

func (vc *VoyageClient) RerankDocuments(
	ctx context.Context,
	query string,
	docs []string,
	topk TopK,
) ([]VoyageReranking, error) {
	const method = "RerankDocuments"
	reqBody := VoyageRerankingRequest{
		Query:           query,
		Documents:       docs,
		Model:           VoyageRerankV2,
		TopK:            topk,
		ReturnDocuments: false,
		Truncation:      false,
	}

	log := vc.logger.With(
		slog.String("method", method),
		slog.String("model", string(reqBody.Model)),
		slog.String("query", string(reqBody.Query)),
		slog.Int("documents_len", len(docs)),
		slog.Int("topk", int(topk)),
		slog.Bool("return_documents", false),
	)

	log.InfoContext(ctx, "sending voyage reranking request...")

	payload, err := json.Marshal(reqBody)
	if err != nil {
		log.With("err", err).ErrorContext(
			ctx,
			"failed to marshal reranking request",
		)
		return nil, fmt.Errorf("failed to marshal reranking request: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(
		ctx,
		http.MethodPost,
		VoyageBaseURL+"/rerank",
		payload,
	)
	if err != nil {
		log.With("err", err).ErrorContext(
			ctx,
			"failed to create request with context",
		)
		return nil, fmt.Errorf("failed to create request with context: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+vc.apiKey)

	start := time.Now()
	resp, err := vc.client.Do(req)
	if err != nil {
		log.With(
			slog.Any("err", err),
			slog.String("status", resp.Status),
		).ErrorContext(
			ctx,
			"reranking request failed",
		)
		return nil, fmt.Errorf("reranking request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.With(
			slog.Any("err", err),
			slog.String("status", resp.Status),
		).ErrorContext(
			ctx,
			"voyage returned non-200 status",
		)
		return nil, fmt.Errorf("voyage returned non-200 status: %w", err)
	}

	duration := time.Since(start)
	log.With(
		slog.Int64("duration_ms", duration.Milliseconds()),
	).InfoContext(ctx, "voyage response received: decoding response...")

	var result VoyageRerankingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.With("err", err).ErrorContext(
			ctx,
			"failed to decode voyage response",
		)
		return nil, fmt.Errorf("failed to decode voyage response: %w", err)
	}

	duration = time.Since(start)
	log.With(
		slog.Int64("duration_ms", duration.Milliseconds()),
		slog.Int("index_count", len(result.Data)),
	).InfoContext(ctx, "decoded response: returning...")

	return result.Data, nil
}
