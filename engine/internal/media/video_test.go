package media

import (
	"encoding/base64"
	"strings"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestVideoStreamGeneratorPatterns(t *testing.T) {
	patterns := []string{"bouncing_ball", "color_bars", "noise"}

	for _, p := range patterns {
		t.Run("Pattern_"+p, func(t *testing.T) {
			cfg := &pb.VideoStreamFunction{
				Codec:           "mjpeg",
				Width:           160,
				Height:          120,
				Fps:             30,
				Pattern:         p,
				RenderTimestamp: true,
			}

			gen := NewVideoStreamGenerator(cfg)

			// Generate 3 sequential frames
			for f := int64(0); f < 3; f++ {
				frame, err := gen.NextFrame()
				if err != nil {
					t.Fatalf("failed to generate frame %d: %v", f, err)
				}
				if frame.FrameIndex != f {
					t.Errorf("expected frame index %d, got %d", f, frame.FrameIndex)
				}
				if frame.Width != 160 || frame.Height != 120 {
					t.Errorf("expected 160x120, got %dx%d", frame.Width, frame.Height)
				}
				if frame.SizeBytes == 0 || frame.DataB64 == "" {
					t.Errorf("empty frame payload")
				}

				// Check that it's a valid JPEG image (SOI 0xFFD8 marker in base64 is /9j/)
				rawBytes, err := base64.StdEncoding.DecodeString(frame.DataB64)
				if err != nil {
					t.Fatalf("failed to decode base64: %v", err)
				}
				if len(rawBytes) < 4 || rawBytes[0] != 0xFF || rawBytes[1] != 0xD8 {
					t.Errorf("expected JPEG SOI marker 0xFFD8, got %X %X", rawBytes[0], rawBytes[1])
				}
				if !strings.HasPrefix(frame.DataB64, "/9j/") {
					t.Errorf("expected JPEG base64 signature /9j/, got %s", frame.DataB64[:5])
				}
			}
		})
	}
}
