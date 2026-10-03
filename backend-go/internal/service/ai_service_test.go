package service

import (
	"context"
	"testing"

	aiprovider "cuan-backend/internal/provider/ai"

	"github.com/stretchr/testify/assert"
)

type stubProvider struct {
	lastReq aiprovider.AIRequest
}

func (s *stubProvider) GenerateCompletion(_ context.Context, req aiprovider.AIRequest) (string, error) {
	s.lastReq = req
	return `{"reply": "Halo!", "is_transaction": false}`, nil
}

func TestAIService_Chat_Multimodal(t *testing.T) {
	stub := &stubProvider{}
	svc := NewAIService(stub, "")

	resp, err := svc.Chat(ChatParams{
		Message:       "Beli bakso 20rb",
		ImageBase64:   "dGVzdA==",
		AudioBase64:   "YXVkaW8=",
		AudioFormat:   "ogg",
		AudioMimeType: "audio/ogg",
		UserContext:   "Dompet: Cash",
	})
	assert.NoError(t, err)
	assert.Equal(t, "Halo!", resp.Reply)
	assert.Equal(t, "YXVkaW8=", stub.lastReq.Base64Audio)
	assert.Equal(t, "ogg", stub.lastReq.AudioFormat)
	assert.Equal(t, "dGVzdA==", stub.lastReq.Base64Image)
}

func TestAIService_ProcessVoice_FileNotFound(t *testing.T) {
	svc := NewAIService(&stubProvider{}, "http://localhost:8000")

	_, err := svc.ProcessVoice("./non_existent_file.ogg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to open audio file")
}

func TestAIService_ProcessVoice_NoWhisperURL(t *testing.T) {
	svc := NewAIService(&stubProvider{}, "")

	_, err := svc.ProcessVoice("./non_existent_file.ogg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "whisper service tidak dikonfigurasi")
}
