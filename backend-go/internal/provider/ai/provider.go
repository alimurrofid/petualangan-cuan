package ai

import "context"

type AIRequest struct {
	Prompt      string
	Base64Image string
	System      string
}

type Provider interface {
	GenerateCompletion(ctx context.Context, req AIRequest) (string, error)
}
