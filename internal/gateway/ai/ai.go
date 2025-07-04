package ai

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/generative-ai-go/genai"
	"golang.org/x/time/rate"
)

type AI interface {
	Generate(ctx context.Context, prompt, instruction string, temp float32) (string, error)
}

type ai struct {
	model   string
	client  *genai.Client
	limiter *rate.Limiter
	log     *slog.Logger
}

func NewAI(model string, client *genai.Client, limiter *rate.Limiter, log *slog.Logger) *ai {
	return &ai{
		model:   model,
		client:  client,
		limiter: limiter,
		log:     log,
	}
}

func (a *ai) Generate(ctx context.Context, prompt, instruction string, temp float32) (string, error) {
	const op = "gateway.ai.Generate"
	if err := a.limiter.Wait(ctx); err != nil {
		a.log.Error(op, "error:", err)
		return "", err
	}

	model := a.client.GenerativeModel(a.model)
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{
			genai.Text(instruction),
		},
	}
	model.Temperature = &temp

	response, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		a.log.Error(op, "error:", err)
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
