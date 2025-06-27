package ai

import (
	"context"
	"fmt"

	"github.com/google/generative-ai-go/genai"
)

type AI interface {
	Generate(ctx context.Context, prompt, instruction string, temp float32) (string, error)
}

type ai struct {
	model  string
	client *genai.Client
}

func NewAI(model string, client *genai.Client) *ai {
	return &ai{model: model, client: client}
}

func (a *ai) Generate(ctx context.Context, prompt, instruction string, temp float32) (string, error) {
	model := a.client.GenerativeModel(a.model)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text(instruction),
		},
	}
	model.Temperature = &temp

	response, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return "", err
	}

	return printResponse(response), nil
}

func printResponse(resp *genai.GenerateContentResponse) string {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				return fmt.Sprint(part)
			}
		}
	}
	return ""
}
