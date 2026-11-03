package sink

import (
	"context"
	"encoding/json"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestKafkaSinkWriteMessages(t *testing.T) {
	mockProducer := NewMockKafkaProducer()
	ks := NewKafkaSinkWithProducer(mockProducer, "test.events")

	ctx := context.Background()
	event1 := map[string]any{
		"event_id": "abc-123",
		"score":    95.5,
	}

	err := ks.WriteEvent(ctx, []byte("key-1"), event1)
	if err != nil {
		t.Fatalf("unexpected WriteEvent error: %v", err)
	}

	if len(mockProducer.Messages) != 1 {
		t.Fatalf("expected 1 message produced, got %d", len(mockProducer.Messages))
	}

	msg := mockProducer.Messages[0]
	if string(msg.Key) != "key-1" {
		t.Fatalf("expected message key 'key-1', got '%s'", string(msg.Key))
	}

	var parsed map[string]any
	if err := json.Unmarshal(msg.Value, &parsed); err != nil {
		t.Fatalf("failed to unmarshal produced message value: %v", err)
	}

	if parsed["event_id"] != "abc-123" || parsed["score"] != 95.5 {
		t.Fatalf("produced payload mismatch: %v", parsed)
	}

	if err := ks.Close(); err != nil {
		t.Fatalf("error closing sink: %v", err)
	}
	if !mockProducer.Closed {
		t.Fatalf("expected mock producer to be closed")
	}
}

func TestCreateSinkFallback(t *testing.T) {
	s, err := CreateSink(&pb.SinkConfig{}, nil)
	if err != nil {
		t.Fatalf("unexpected error creating sink: %v", err)
	}
	if s == nil {
		t.Fatalf("expected non-nil sink")
	}
}
