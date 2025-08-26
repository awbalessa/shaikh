package infra

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/config"
	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	voyageBaseURL              string        = "https://api.voyageai.com/v1"
	voyageTimeoutTenSeconds    time.Duration = 10 * time.Second
	voyageMaxRetriesThree      int           = 3
	voyageMaxIdleConns         int           = 100
	voyageMaxIdleConnsPerHost  int           = 10
	voyageIdleConnTimeout      time.Duration = 90 * time.Second
	voyageDialContextTimeout   time.Duration = 5 * time.Second
	voyageDialContextKeepAlive time.Duration = 30 * time.Second
	voyageTLSHandshakeTimeout  time.Duration = 10 * time.Second
	voyageEmbedV3p5            string        = "voyage-3.5"
	inputTypeQuery             string        = "query"
	inputTypeDocument          string        = "document"
	outputDimension1024        int32         = 1024
	outputDimensionTypeFloat   string        = "float"
	voyageRerankV2             string        = "rerank-2"
	voyageRerankV2p5Lite       string        = "rerank-2.5-lite"
)

type VoyageEmbedderReranker struct {
	Cli    *retryablehttp.Client
	Log    *slog.Logger
	apiKey string
}

func NewVoyageEmbedderReranker(env *config.Env, maxRetries int, timeout time.Duration, log *slog.Logger) *VoyageEmbedderReranker {
	client := &http.Client{
		Timeout: timeout,
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

	retryClient := retryablehttp.NewClient()
	retryClient.HTTPClient = client
	retryClient.Logger = log
	retryClient.RetryMax = maxRetries
	retryClient.CheckRetry = retryablehttp.ErrorPropagatedRetryPolicy
	retryClient.Backoff = retryablehttp.DefaultBackoff

	return &VoyageEmbedderReranker{
		Cli:    retryClient,
		apiKey: env.VoyageAPIKey,
		Log:    log,
	}
}

type embedding1024 [1024]float32

type voyageEmbeddingRequest struct {
	Input               []string `json:"input"`
	Model               string   `json:"model"`
	InputType           string   `json:"input_type"`
	Truncation          bool     `json:"truncation"`
	OutputDimension     int32    `json:"output_dimension"`
	OutputDimensionType string   `json:"output_dtype"`
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
	Model      string            `json:"model"`
	Usage      usage             `json:"usage"`
}

type voyageRerankingRequest struct {
	Query           string   `json:"query"`
	Documents       []string `json:"documents"`
	Model           string   `json:"model"`
	TopK            dom.TopK `json:"top_k"`
	ReturnDocuments bool     `json:"return_documents"`
	Truncation      bool     `json:"truncation"`
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

func (v *VoyageEmbedderReranker) EmbedQueries(
	ctx context.Context,
	queries []string,
) ([]dom.Vector, error) {
	const method = "EmbedQueries"

	reqBody := voyageEmbeddingRequest{
		Input:               queries,
		Model:               voyageEmbedV3p5,
		InputType:           inputTypeQuery,
		Truncation:          false,
		OutputDimension:     outputDimension1024,
		OutputDimensionType: outputDimensionTypeFloat,
	}

	log := v.Log.With(
		slog.String("method", method),
	)

	log.With(
		slog.String("model", string(reqBody.Model)),
		slog.Int("number_of_queries", len(queries)),
	).DebugContext(ctx, "sending voyage embedding request...")

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
	req.Header.Set("Authorization", "Bearer "+v.apiKey)

	start := time.Now()
	resp, err := v.Cli.Do(req)
	if err != nil {
		log.With(
			slog.Any("err", err),
		).ErrorContext(
			ctx,
			"embedding request failed",
		)
		return nil, fmt.Errorf("embedding request failed: %w", err)
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		log.With(
			slog.Int("status_code", resp.StatusCode),
			slog.String("status", resp.Status),
		).ErrorContext(
			ctx,
			"voyage returned non-200 status",
		)
		return nil, fmt.Errorf("voyage returned non-200 status: %s", resp.Status)
	}

	log.With(
		slog.String("method", method),
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

	vectors := make([]dom.Vector, len(result.Data))
	for i, item := range result.Data {
		vectors[i] = item.Embedding[:]
	}

	if len(vectors) != len(queries) {
		return nil, dom.ErrQueriesVectorsNot1to1
	}

	log.With(
		slog.String("method", method),
		slog.String("duration", time.Since(start).String()),
		slog.Int("number_of_vectors", len(vectors)),
	).DebugContext(ctx, "embedding completed: returning...")

	return vectors, nil
}

func (v *VoyageEmbedderReranker) RerankDocuments(
	ctx context.Context,
	query string,
	docs []string,
	topk dom.TopK,
) ([]dom.Rank, error) {
	const method = "RerankDocuments"

	reqBody := voyageRerankingRequest{
		Query:           query,
		Documents:       docs,
		Model:           voyageRerankV2p5Lite,
		TopK:            topk,
		ReturnDocuments: false,
		Truncation:      false,
	}

	log := v.Log.With(
		slog.String("method", method),
	)

	log.With(
		slog.String("model", string(reqBody.Model)),
		slog.String("query", string(reqBody.Query)),
		slog.Int("number_of_documents", len(docs)),
		slog.Int("topk", int(topk)),
	).DebugContext(ctx, "sending voyage reranking request...")

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
	req.Header.Set("Authorization", "Bearer "+v.apiKey)

	start := time.Now()
	resp, err := v.Cli.Do(req)
	if err != nil {
		log.With(
			slog.Any("err", err),
		).ErrorContext(
			ctx,
			"reranking request failed",
		)
		return nil, fmt.Errorf("reranking request failed: %w", err)
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		log.With(
			slog.Int("status_code", resp.StatusCode),
			slog.String("status", resp.Status),
		).ErrorContext(
			ctx,
			"voyage returned non-200 status",
		)
		return nil, fmt.Errorf("voyage returned non-200 status: %s", resp.Status)
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

	ranks := make([]dom.Rank, len(result.Data))
	for i, item := range result.Data {
		ranks[i] = dom.Rank{
			Index:     int32(item.Index),
			Relevance: item.RelevanceScore,
		}
	}

	return ranks, nil
}
