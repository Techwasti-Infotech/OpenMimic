package sink

import (
	"context"
	"strings"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestS3StorageSinkCSVAndRollover(t *testing.T) {
	mockClient := NewMemoryStorageClient()
	cfg := &pb.S3Sink{
		Bucket:            "test-bucket",
		Prefix:            "data/sales/",
		FileFormat:        "csv",
		MaxRecordsPerFile: 3,
	}

	sink := NewS3StorageSink(cfg, mockClient)

	// Write 5 records (should produce part-00001 with 3 records and part-00002 on Close with 2 records)
	for i := 1; i <= 5; i++ {
		err := sink.WriteEvent(context.Background(), nil, map[string]any{
			"id":     i,
			"status": "PAID",
		})
		if err != nil {
			t.Fatalf("failed to write event %d: %v", i, err)
		}
	}

	if err := sink.Close(); err != nil {
		t.Fatalf("failed to close sink: %v", err)
	}

	if len(mockClient.Uploads) != 2 {
		t.Fatalf("expected 2 uploaded file parts, got %d", len(mockClient.Uploads))
	}

	for path, content := range mockClient.Uploads {
		if !strings.HasPrefix(path, "test-bucket/data/sales/part-") {
			t.Errorf("unexpected s3 upload path: %s", path)
		}
		if !strings.HasSuffix(path, ".csv") {
			t.Errorf("expected .csv suffix, got %s", path)
		}
		if len(content) == 0 {
			t.Errorf("expected non-empty content in upload %s", path)
		}
	}
}

func TestS3StorageSinkParquet(t *testing.T) {
	mockClient := NewMemoryStorageClient()
	cfg := &pb.S3Sink{
		Bucket:            "warehouse",
		Prefix:            "raw/events/",
		FileFormat:        "parquet",
		MaxRecordsPerFile: 10,
	}

	sink := NewS3StorageSink(cfg, mockClient)
	_ = sink.WriteEvent(context.Background(), nil, map[string]any{"metric": 42.5})
	_ = sink.Close()

	if len(mockClient.Uploads) != 1 {
		t.Fatalf("expected 1 parquet file upload, got %d", len(mockClient.Uploads))
	}

	for _, content := range mockClient.Uploads {
		if !strings.HasPrefix(string(content), "PAR1") {
			t.Errorf("expected PAR1 magic bytes header for parquet file")
		}
	}
}
