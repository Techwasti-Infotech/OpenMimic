package sink

import (
	"context"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestRedisSink(t *testing.T) {
	mockClient := NewMockRedisClient()
	cfg := &pb.RedisSink{
		Addr:             "localhost:6379",
		KeyPattern:       "user:{id}",
		Mode:             "set",
		StreamOrChannel: "",
	}

	sink := NewRedisSink(cfg, mockClient)
	ctx := context.Background()

	payload := map[string]any{
		"id":   "u123",
		"name": "Alice",
	}

	if err := sink.WriteEvent(ctx, []byte("u123"), payload); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if len(mockClient.Commands) != 1 {
		t.Fatalf("expected 1 redis command, got %d", len(mockClient.Commands))
	}

	if _, ok := mockClient.Records["user:u123"]; !ok {
		t.Fatalf("expected key user:u123 in redis store")
	}
}

func TestClickHouseSink(t *testing.T) {
	mockClient := NewMockClickHouseClient()
	cfg := &pb.ClickHouseSink{
		Addr:     "localhost:9000",
		Database: "analytics",
		Table:    "analytics.events",
	}

	sink := NewClickHouseSink(cfg, mockClient)
	ctx := context.Background()

	payload := map[string]any{
		"event_id": "evt_999",
		"metric":   42.5,
	}

	if err := sink.WriteEvent(ctx, []byte("evt_999"), payload); err != nil {
		t.Fatalf("unexpected write error: %v", err)
	}

	if len(mockClient.Queries) != 1 {
		t.Fatalf("expected 1 clickhouse query, got %d", len(mockClient.Queries))
	}
}

