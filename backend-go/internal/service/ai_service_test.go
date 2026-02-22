package service

import (
	"context"
	"testing"

	aiprovider "cuan-backend/internal/provider/ai"

	"github.com/stretchr/testify/assert"
)

type stubProvider struct{}

func (s *stubProvider) GenerateCompletion(_ context.Context, _ aiprovider.AIRequest) (string, error) {
	return "", nil
}

func TestAIService_ProcessVoice_FileNotFound(t *testing.T) {
	svc := NewAIService(&stubProvider{}, "")

	_, err := svc.ProcessVoice("./non_existent_file.ogg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestAIService_ProcessVoice_NoWhisperURL(t *testing.T) {
	svc := NewAIService(&stubProvider{}, "")

	_, err := svc.ProcessVoice("./non_existent_file.ogg")
	assert.Error(t, err)
}
