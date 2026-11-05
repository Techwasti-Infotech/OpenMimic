package media

import (
	"strings"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestGenerateSyntheticWAV(t *testing.T) {
	bytes := GenerateSyntheticWAV(100)
	if len(bytes) < 44 {
		t.Fatalf("expected WAV header at least 44 bytes, got %d", len(bytes))
	}
	if string(bytes[:4]) != "RIFF" || string(bytes[8:12]) != "WAVE" {
		t.Errorf("invalid WAV header magic bytes")
	}
}

func TestGenerateSyntheticImagePNG(t *testing.T) {
	bytes, err := GenerateSyntheticImage("png_image", 32, 32)
	if err != nil {
		t.Fatalf("failed to generate PNG: %v", err)
	}
	// PNG Magic Header: 0x89 'P' 'N' 'G'
	if len(bytes) < 8 || bytes[0] != 0x89 || string(bytes[1:4]) != "PNG" {
		t.Errorf("invalid PNG header magic bytes")
	}
}

func TestGenerateSyntheticImageJPEG(t *testing.T) {
	bytes, err := GenerateSyntheticImage("jpeg_image", 32, 32)
	if err != nil {
		t.Fatalf("failed to generate JPEG: %v", err)
	}
	// JPEG SOI marker: 0xFF 0xD8
	if len(bytes) < 4 || bytes[0] != 0xFF || bytes[1] != 0xD8 {
		t.Errorf("invalid JPEG SOI marker")
	}
}

func TestGenerateMediaPayload(t *testing.T) {
	payload := GenerateMediaPayload(&pb.BinaryMediaFunction{
		Format: "wav_audio",
	})

	if payload["mime_type"] != "audio/wav" {
		t.Errorf("expected audio/wav mime type, got %v", payload["mime_type"])
	}
	b64, ok := payload["base64"].(string)
	if !ok || !strings.HasPrefix(b64, "UklGR") { // Base64 of RIFF
		t.Errorf("expected valid RIFF base64 encoding")
	}
}
