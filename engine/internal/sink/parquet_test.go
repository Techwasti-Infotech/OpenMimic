package sink

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/binary"
	"encoding/json"
	"io"
	"strings"
	"testing"

	"github.com/klauspost/compress/zstd"
	pb "openmimic/engine/pkg/proto/v1"
)

func TestParquetColumnarEncodingAndCodecs(t *testing.T) {
	records := []map[string]any{
		{"id": 1.0, "status": "active", "amount": 105.50, "verified": true},
		{"id": 2.0, "status": "pending", "amount": 25.00, "verified": false},
		{"id": 3.0, "status": "active", "amount": 320.10, "verified": true},
		{"id": 4.0, "status": "completed", "amount": 80.00, "verified": true},
	}

	codecs := []string{"none", "gzip", "zstd", "snappy"}

	for _, codec := range codecs {
		t.Run("Codec_"+codec, func(t *testing.T) {
			b, err := EncodeColumnarParquet(records, codec)
			if err != nil {
				t.Fatalf("failed to encode parquet with codec %s: %v", codec, err)
			}

			// Validate Parquet 4-byte header magic
			if len(b) < 12 {
				t.Fatalf("output too short: %d bytes", len(b))
			}
			if string(b[:4]) != "PAR1" {
				t.Fatalf("expected PAR1 magic prefix, got %q", string(b[:4]))
			}

			// Validate Parquet 4-byte footer magic
			if string(b[len(b)-4:]) != "PAR1" {
				t.Fatalf("expected PAR1 magic suffix, got %q", string(b[len(b)-4:]))
			}

			// Validate footer length
			footerLen := binary.LittleEndian.Uint32(b[len(b)-8 : len(b)-4])
			footerStart := len(b) - 8 - int(footerLen)
			if footerStart < 4 {
				t.Fatalf("invalid footer length %d", footerLen)
			}

			footerBytes := b[footerStart : len(b)-8]
			var header ParquetHeader
			if err := json.Unmarshal(footerBytes, &header); err != nil {
				t.Fatalf("failed to parse footer JSON: %v", err)
			}

			if header.MagicCode != "PAR1" {
				t.Errorf("footer MagicCode = %q, want PAR1", header.MagicCode)
			}
			if header.NumRows != 4 {
				t.Errorf("footer NumRows = %d, want 4", header.NumRows)
			}
			if header.Compression != codec {
				t.Errorf("footer Compression = %q, want %q", header.Compression, codec)
			}
			if len(header.Columns) != 4 {
				t.Errorf("expected 4 columns, got %d", len(header.Columns))
			}
		})
	}
}

func TestS3SinkParquetCompressedUpload(t *testing.T) {
	mockClient := NewMemoryStorageClient()
	cfg := &pb.S3Sink{
		Bucket:            "telemetry-lake",
		Prefix:            "stream-data/",
		FileFormat:        "parquet",
		CompressionCodec:  "zstd",
		MaxRecordsPerFile: 2,
	}

	s3Sink := NewS3StorageSink(cfg, mockClient)
	ctx := context.Background()

	_ = s3Sink.WriteEvent(ctx, nil, map[string]any{"device_id": "sensor-1", "temp": 24.5})
	_ = s3Sink.WriteEvent(ctx, nil, map[string]any{"device_id": "sensor-2", "temp": 26.1})

	if len(mockClient.Uploads) != 1 {
		t.Fatalf("expected 1 uploaded file part, got %d", len(mockClient.Uploads))
	}

	for k, data := range mockClient.Uploads {
		if !strings.HasSuffix(k, ".parquet") {
			t.Errorf("expected key to end in .parquet: %s", k)
		}
		if !strings.HasPrefix(string(data), "PAR1") {
			t.Errorf("expected PAR1 header")
		}
	}
}

func TestDecompressGzipZstd(t *testing.T) {
	original := []byte("high-throughput synthetic data generation with openmimic")

	// GZIP
	gz, err := CompressData("gzip", original)
	if err != nil {
		t.Fatalf("gzip compress error: %v", err)
	}
	gr, err := gzip.NewReader(bytes.NewReader(gz))
	if err != nil {
		t.Fatalf("gzip reader error: %v", err)
	}
	decompGz, _ := io.ReadAll(gr)
	if string(decompGz) != string(original) {
		t.Errorf("gzip roundtrip mismatch: got %q", string(decompGz))
	}

	// ZSTD
	zs, err := CompressData("zstd", original)
	if err != nil {
		t.Fatalf("zstd compress error: %v", err)
	}
	zr, err := zstd.NewReader(bytes.NewReader(zs))
	if err != nil {
		t.Fatalf("zstd reader error: %v", err)
	}
	decompZs, _ := io.ReadAll(zr)
	if string(decompZs) != string(original) {
		t.Errorf("zstd roundtrip mismatch: got %q", string(decompZs))
	}
}
