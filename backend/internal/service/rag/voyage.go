package rag

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/hashicorp/go-retryablehttp"
	"github.com/pgvector/pgvector-go"
)

const (
	voyageBaseURL              string                       = "https://api.voyageai.com/v1"
	voyageTimeoutTenSeconds    time.Duration                = 10 * time.Second
	voyageMaxRetriesThree      int                          = 3
	voyageMaxIdleConns         int                          = 100
	voyageMaxIdleConnsPerHost  int                          = 10
	voyageIdleConnTimeout      time.Duration                = 90 * time.Second
	voyageDialContextTimeout   time.Duration                = 5 * time.Second
	voyageDialContextKeepAlive time.Duration                = 30 * time.Second
	voyageTLSHandshakeTimeout  time.Duration                = 10 * time.Second
	voyageEmbedV3p5            embeddingModel               = "voyage-3.5"
	inputTypeQuery             embeddingInputType           = "query"
	inputTypeDocument          embeddingInputType           = "document"
	outputDimension1024        embeddingOutputDimension     = 1024
	outputDimensionTypeFloat   embeddingOutputDimensionType = "float"
	voyageRerankV2             rerankingModel               = "rerank-2"
	voyageRerankV2p5Lite       rerankingModel               = "rerank-2.5-lite"
)

type voyageClientConfig struct {
	config     *config.Config
	maxRetries int
	timeout    time.Duration
}

type voyageClient struct {
	client *retryablehttp.Client
	apiKey string
	logger *slog.Logger
}

type embeddingModel string
type embeddingInputType string
type embeddingOutputDimension int
type embeddingOutputDimensionType string
type embeddingEncodingFormat *string
type embedding1024 [1024]float32

type voyageEmbeddingRequest struct {
	Input               []string                     `json:"input"`
	Model               embeddingModel               `json:"model"`
	InputType           embeddingInputType           `json:"input_type"`
	Truncation          bool                         `json:"truncation"`
	OutputDimension     embeddingOutputDimension     `json:"output_dimension"`
	OutputDimensionType embeddingOutputDimensionType `json:"output_dtype"`
}

type voyageEmbedding struct {
	ObjectType string        `json:"object"`
	Embedding  embedding1024 `json:"embedding"`
	Index      int           `json:"index"`
}

type usage struct {
	TokensUsed int `json:"total_tokens"`
}

type voyageEmbeddingResponse struct {
	ObjectType string            `json:"object"`
	Data       []voyageEmbedding `json:"data"`
	Model      embeddingModel    `json:"model"`
	Usage      usage             `json:"usage"`
}

type rerankingModel string

type voyageRerankingRequest struct {
	Query           string         `json:"query"`
	Documents       []string       `json:"documents"`
	Model           rerankingModel `json:"model"`
	TopK            TopK           `json:"top_k"`
	ReturnDocuments bool           `json:"return_documents"`
	Truncation      bool           `json:"truncation"`
}

type voyageReranking struct {
	Index          int     `json:"index"`
	RelevanceScore float64 `json:"relevance_score"`
}

type voyageRerankingResponse struct {
	Object string            `json:"object"`
	Data   []voyageReranking `json:"data"`
	Model  string            `json:"model"`
	Usage  usage             `json:"usage"`
}

func newVoyageClient(cfg voyageClientConfig) *voyageClient {
	client := &http.Client{
		Timeout: cfg.timeout,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   voyageDialContextTimeout,
				KeepAlive: voyageDialContextKeepAlive,
			}).DialContext,
			MaxIdleConns:        voyageMaxIdleConns,
			MaxIdleConnsPerHost: voyageMaxIdleConnsPerHost,
			IdleConnTimeout:     voyageIdleConnTimeout,
			TLSHandshakeTimeout: voyageTLSHandshakeTimeout,
		},
	}

	logger := slog.Default().With(
		"component", "voyage",
	)

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = client
	retryClient.Logger = logger
	retryClient.RetryMax = cfg.maxRetries

	return &voyageClient{
		client: retryClient,
		apiKey: cfg.config.VoyageAPIKey,
		logger: logger,
	}
}

func (vc *voyageClient) embedQueries(
	ctx context.Context,
	queries []string,
) ([]pgvector.Vector, error) {
	const method = "embedQuery"
	reqBody := voyageEmbeddingRequest{
		Input:               queries,
		Model:               voyageEmbedV3p5,
		InputType:           inputTypeQuery,
		Truncation:          false,
		OutputDimension:     outputDimension1024,
		OutputDimensionType: outputDimensionTypeFloat,
	}

	log := vc.logger.With(
		slog.String("method", method),
		slog.String("model", string(reqBody.Model)),
		slog.Int("input_len", len(queries)),
	)

	log.DebugContext(ctx, "sending voyage embedding request...")

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
		voyageBaseURL+"/embeddings",
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

	log.With(
		slog.String("duration", time.Since(start).String()),
	).DebugContext(ctx, "voyage response received: decoding response...")

	var result voyageEmbeddingResponse
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

	if len(vectors) != len(queries) {
		return nil, errors.New("error: vectors and queries are one-to-one")
	}

	log.With(
		slog.String("duration", time.Since(start).String()),
		slog.Int("embedding_count", len(vectors)),
	).DebugContext(ctx, "embedding completed: returning...")

	return vectors, nil
}

func (vc *voyageClient) rerankDocuments(
	ctx context.Context,
	query string,
	docs []string,
	topk TopK,
) ([]voyageReranking, error) {
	const method = "rerankDocuments"
	reqBody := voyageRerankingRequest{
		Query:           query,
		Documents:       docs,
		Model:           voyageRerankV2p5Lite,
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

	log.DebugContext(ctx, "sending voyage reranking request...")

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
		voyageBaseURL+"/rerank",
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
			slog.String("status", resp.Status),
		).ErrorContext(
			ctx,
			"voyage returned non-200 status",
		)
		return nil, fmt.Errorf("voyage returned non-200 status")
	}

	log.With(
		slog.String("duration", time.Since(start).String()),
	).DebugContext(ctx, "voyage response received: decoding response...")

	var result voyageRerankingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.With("err", err).ErrorContext(
			ctx,
			"failed to decode voyage response",
		)
		return nil, fmt.Errorf("failed to decode voyage response: %w", err)
	}

	log.With(
		slog.String("duration", time.Since(start).String()),
		slog.Int("index_count", len(result.Data)),
	).DebugContext(ctx, "reranking completed: returning...")

	return result.Data, nil
}
