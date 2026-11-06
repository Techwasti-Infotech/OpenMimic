package sink

import (
	"context"
	"strings"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestIcebergStreamingSink(t *testing.T) {
	mockClient := NewMemoryLakehouseStorageClient()
	cfg := &pb.IcebergSink{
		CatalogUri:       "http://localhost:8181",
		Warehouse:        "s3://mimic-lakehouse",
		Database:         "analytics",
		Table:            "user_events",
		FormatVersion:    "2",
		CommitIntervalMs: 1000,
	}

	sink := NewIcebergStreamingSink(cfg, mockClient)
	ctx := context.Background()

	_ = sink.WriteEvent(ctx, nil, map[string]any{"id": 1, "action": "LOGIN", "score": 98.5})
	_ = sink.WriteEvent(ctx, nil, map[string]any{"id": 2, "action": "PURCHASE", "score": 450.0})

	if err := sink.Close(); err != nil {
		t.Fatalf("failed to close iceberg sink: %v", err)
	}

	// Verify that data file and metadata JSON snapshot were written
	var hasDataFile, hasMetadataFile bool
	for path, content := range mockClient.Files {
		if strings.Contains(path, "/data/") && strings.HasSuffix(path, ".parquet") {
			hasDataFile = true
			if !strings.HasPrefix(string(content), "PAR1") {
				t.Errorf("expected parquet file to begin with PAR1 header, got %s", path)
			}
		}
		if strings.Contains(path, "/metadata/") && strings.HasSuffix(path, ".metadata.json") {
			hasMetadataFile = true
			if !strings.Contains(string(content), `"format-version": 2`) {
				t.Errorf("expected Iceberg V2 metadata, got %s", string(content))
			}
		}
	}

	if !hasDataFile {
		t.Errorf("expected Iceberg Parquet data file to be written")
	}
	if !hasMetadataFile {
		t.Errorf("expected Iceberg metadata JSON snapshot file to be written")
	}
}

func TestDeltaLakeStreamingSinkWithTxLog(t *testing.T) {
	mockClient := NewMemoryLakehouseStorageClient()
	cfg := &pb.DeltaLakeSink{
		TableUri:         "s3://lakehouse/delta/orders",
		Mode:             "append",
		BatchSize:        2,
		GenerateDeltaLog: true,
	}

	sink := NewDeltaLakeStreamingSink(cfg, mockClient)
	ctx := context.Background()

	// Write 2 events (triggers batch commit)
	_ = sink.WriteEvent(ctx, nil, map[string]any{"order_id": "ord-1", "amount": 100.0})
	_ = sink.WriteEvent(ctx, nil, map[string]any{"order_id": "ord-2", "amount": 250.0})

	// Verify Delta parquet file + _delta_log/00000000000000000000.json
	var hasDeltaParquet, hasDeltaTxLog bool
	for path, content := range mockClient.Files {
		if strings.Contains(path, "part-") && strings.HasSuffix(path, ".parquet") {
			hasDeltaParquet = true
		}
		if strings.Contains(path, "_delta_log/") && strings.HasSuffix(path, ".json") {
			hasDeltaTxLog = true
			if !strings.Contains(string(content), `"operation":"WRITE"`) {
				t.Errorf("expected Delta log to contain WRITE operation, got %s", string(content))
			}
		}
	}

	if !hasDeltaParquet {
		t.Errorf("expected Delta Lake Parquet part to be written")
	}
	if !hasDeltaTxLog {
		t.Errorf("expected Delta Lake transaction commit log to be written")
	}
}
