package ai

import "context"

type AIRequest struct {
	Prompt        string
	Base64Image   string
	Base64Audio   string
	AudioFormat   string
	AudioMimeType string
	System        string
}

type Provider interface {
	GenerateCompletion(ctx context.Context, req AIRequest) (string, error)
}
