package media

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"

	pb "openmimic/engine/pkg/proto/v1"
)

// GenerateSyntheticWAV creates a valid RIFF/WAV audio container with PCM audio sine samples.
func GenerateSyntheticWAV(durationMs int) []byte {
	if durationMs <= 0 {
		durationMs = 500
	}

	sampleRate := 44100
	numSamples := (sampleRate * durationMs) / 1000
	numChannels := 1
	bitsPerSample := 16
	bytesPerSample := bitsPerSample / 8
	blockAlign := numChannels * bytesPerSample
	byteRate := sampleRate * blockAlign
	dataSize := numSamples * blockAlign
	chunkSize := 36 + dataSize

	var buf bytes.Buffer

	// RIFF Header
	buf.WriteString("RIFF")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(chunkSize))
	buf.WriteString("WAVE")

	// fmt Subchunk
	buf.WriteString("fmt ")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(16))      // Subchunk1Size for PCM
	_ = binary.Write(&buf, binary.LittleEndian, uint16(1))       // AudioFormat 1 = PCM
	_ = binary.Write(&buf, binary.LittleEndian, uint16(numChannels))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(sampleRate))
	_ = binary.Write(&buf, binary.LittleEndian, uint32(byteRate))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(blockAlign))
	_ = binary.Write(&buf, binary.LittleEndian, uint16(bitsPerSample))

	// data Subchunk
	buf.WriteString("data")
	_ = binary.Write(&buf, binary.LittleEndian, uint32(dataSize))

	// Generate 440 Hz Sine wave audio samples
	freq := 440.0
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(sampleRate)
		sample := int16(math.Sin(2.0*math.Pi*freq*t) * 16384.0)
		_ = binary.Write(&buf, binary.LittleEndian, sample)
	}

	return buf.Bytes()
}

// GenerateSyntheticImage creates valid JPEG or PNG binary images.
func GenerateSyntheticImage(format string, width, height int) ([]byte, error) {
	if width <= 0 {
		width = 64
	}
	if height <= 0 {
		height = 64
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := color.RGBA{
				R: uint8((x * 255) / width),
				G: uint8((y * 255) / height),
				B: 128,
				A: 255,
			}
			img.Set(x, y, c)
		}
	}

	var buf bytes.Buffer
	if format == "jpeg_image" || format == "jpg" {
		if err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: 80}); err != nil {
			return nil, err
		}
	} else {
		// Default to PNG
		if err := png.Encode(&buf, img); err != nil {
			return nil, err
		}
	}
	return buf.Bytes(), nil
}

// GenerateMediaPayload produces base64 or raw bytes payload representation.
func GenerateMediaPayload(cfg *pb.BinaryMediaFunction) map[string]any {
	var raw []byte
	var mime string

	switch cfg.GetFormat() {
	case "wav_audio":
		raw = GenerateSyntheticWAV(int(cfg.GetDurationMs()))
		mime = "audio/wav"
	case "jpeg_image":
		raw, _ = GenerateSyntheticImage("jpeg_image", int(cfg.GetWidth()), int(cfg.GetHeight()))
		mime = "image/jpeg"
	case "png_image":
		fallthrough
	default:
		raw, _ = GenerateSyntheticImage("png_image", int(cfg.GetWidth()), int(cfg.GetHeight()))
		mime = "image/png"
	}

	return map[string]any{
		"mime_type":  mime,
		"size_bytes": len(raw),
		"base64":     base64.StdEncoding.EncodeToString(raw),
	}
}
