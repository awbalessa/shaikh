package gemini

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

func NewClient(ctx context.Context, key string) (*genai.Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: key,
	})
	if err != nil {
		return nil, fmt.Errorf("new gemini client: %w", err)
	}

	return client, nil
}
