package sink

import (
	"context"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestSnowflakeMicroBatchSink(t *testing.T) {
	mockClient := NewMemoryWarehouseClient()
	cfg := &pb.SnowflakeSink{
		Database:  "ANALYTICS_DB",
		Schema:    "PUBLIC",
		Table:     "EVENTS",
		BatchSize: 3,
	}

	sink := NewSnowflakeSink(cfg, mockClient)

	// Write 5 events -> 1 batch of 3 flushed automatically, 1 batch of 2 on Close
	for i := 1; i <= 5; i++ {
		_ = sink.WriteEvent(context.Background(), nil, map[string]any{"id": i, "amount": 10.5 * float64(i)})
	}

	target := "ANALYTICS_DB.PUBLIC.EVENTS"
	if len(mockClient.Batches[target]) != 1 {
		t.Fatalf("expected 1 flushed batch before close, got %d", len(mockClient.Batches[target]))
	}

	_ = sink.Close()
	if len(mockClient.Batches[target]) != 2 {
		t.Fatalf("expected 2 flushed batches after close, got %d", len(mockClient.Batches[target]))
	}
}

func TestBigQueryStreamingSink(t *testing.T) {
	mockClient := NewMemoryWarehouseClient()
	cfg := &pb.BigQuerySink{
		ProjectId: "my-gcp-project",
		DatasetId: "stream_dataset",
		TableId:   "user_actions",
		BatchSize: 2,
	}

	sink := NewBigQuerySink(cfg, mockClient)
	_ = sink.WriteEvent(context.Background(), nil, map[string]any{"user": "u1"})
	_ = sink.WriteEvent(context.Background(), nil, map[string]any{"user": "u2"})

	target := "my-gcp-project:stream_dataset.user_actions"
	if len(mockClient.Batches[target]) != 1 {
		t.Fatalf("expected 1 batch in BigQuery target, got %d", len(mockClient.Batches[target]))
	}
	_ = sink.Close()
}
