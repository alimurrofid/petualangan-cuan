package utils

import (
	"path/filepath"
	"strings"
)

type AudioMetadata struct {
	Format   string // "ogg", "m4a", "mp3", "wav"
	MimeType string // "audio/ogg", "audio/mp4", "audio/mpeg", "audio/wav"
}

// ResolveAudioMetadata menentukan format audio dan MIME type berdasarkan ekstensi file.
func ResolveAudioMetadata(filename string) AudioMetadata {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	switch ext {
	case "ogg", "opus":
		return AudioMetadata{Format: "ogg", MimeType: "audio/ogg"}
	case "m4a", "mp4", "aac":
		return AudioMetadata{Format: "m4a", MimeType: "audio/mp4"}
	case "mp3":
		return AudioMetadata{Format: "mp3", MimeType: "audio/mpeg"}
	case "wav":
		return AudioMetadata{Format: "wav", MimeType: "audio/wav"}
	default:
		return AudioMetadata{Format: "ogg", MimeType: "audio/ogg"}
	}
}
