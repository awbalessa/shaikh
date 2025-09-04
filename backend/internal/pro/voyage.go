package pro

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/awbalessa/shaikh/backend/internal/dom"
	"github.com/hashicorp/go-retryablehttp"
)

const (
	voyageBaseURL            string = "https://api.voyageai.com/v1"
	voyageEmbedV3p5          string = "voyage-3.5"
	inputTypeQuery           string = "query"
	outputDimension1024      int32  = 1024
	outputDimensionTypeFloat string = "float"
	voyageRerankV2           string = "rerank-2"
	voyageRerankV2p5Lite     string = "rerank-2.5-lite"
)

type VoyageEmbedderReranker struct {
	Cli    *retryablehttp.Client
	apiKey string
}

func NewVoyageEmbedderReranker() *VoyageEmbedderReranker {
	client := &http.Client{
		Timeout: 10 * time.Second,
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
	retryClient.HTTPClient = client
	retryClient.RetryMax = 3
	retryClient.CheckRetry = retryablehttp.ErrorPropagatedRetryPolicy
	retryClient.Backoff = retryablehttp.DefaultBackoff

	return &VoyageEmbedderReranker{
		Cli:    retryClient,
		apiKey: os.Getenv("VOYAGE_API_KEY"),
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
	reqBody := voyageEmbeddingRequest{
		Input:               queries,
		Model:               voyageEmbedV3p5,
		InputType:           inputTypeQuery,
		Truncation:          false,
		OutputDimension:     outputDimension1024,
		OutputDimensionType: outputDimensionTypeFloat,
	}

	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := retryablehttp.NewRequestWithContext(
		ctx,
		http.MethodPost,
		voyageBaseURL+"/embeddings",
		payload,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+v.apiKey)

	resp, err := v.Cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("voyage returned non-200 status: %s", resp.Status)
	}

	var result voyageEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	vectors := make([]dom.Vector, len(result.Data))
	for i, item := range result.Data {
		vectors[i] = item.Embedding[:]
	}

	if len(vectors) != len(queries) {
		return nil, dom.ErrQueriesVectorsNot1to1
	}

	return vectors, nil
}

func (v *VoyageEmbedderReranker) RerankDocuments(
	ctx context.Context,
	query string,
	docs []string,
	topk dom.TopK,
) ([]dom.Rank, error) {
	reqBody := voyageRerankingRequest{
		Query:           query,
		Documents:       docs,
		Model:           voyageRerankV2p5Lite,
		TopK:            topk,
		ReturnDocuments: false,
		Truncation:      false,
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := retryablehttp.NewRequestWithContext(
		ctx,
		http.MethodPost,
		voyageBaseURL+"/rerank",
		payload,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+v.apiKey)

	resp, err := v.Cli.Do(req)
	if err != nil {
		return nil, err
	}

	if resp != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("voyage returned non-200 status: %s", resp.Status)
	}

	var result voyageRerankingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	ranks := make([]dom.Rank, len(result.Data))
	for i, item := range result.Data {
		ranks[i] = dom.Rank{
			Index:     int32(item.Index),
			Relevance: item.RelevanceScore,
		}
	}

	return ranks, nil
}

func (v *VoyageEmbedderReranker) Ping(ctx context.Context) error {
	return dom.ErrNotPingable
}

func (v *VoyageEmbedderReranker) Name() string {
	return "ERM"
}
