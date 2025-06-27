package ai

import (
	"context"

	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func NewClient(ctx context.Context, key string) (*genai.Client, error) {
	client, err := genai.NewClient(ctx, option.WithAPIKey(key))
	if err != nil {
		return nil, err
	}

	return client, nil
}

func ClientCloser(client *genai.Client) error {
	return client.Close()
}
