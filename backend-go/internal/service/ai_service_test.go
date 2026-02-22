package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAIService_ProcessVoice_FileNotFound(t *testing.T) {
	service := NewAIService("", "")

	_, err := service.ProcessVoice("./non_existent_file.ogg")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file or directory")
}

func TestAIService_Chat_EmptyMessage(t *testing.T) {
	service := NewAIService("fake_url", "fake_whisper")

	resp, err := service.Chat("", "", "")

	assert.Error(t, err)
	assert.Nil(t, resp)
}
