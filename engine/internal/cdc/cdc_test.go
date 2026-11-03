package cdc

import (
	"testing"
	"time"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

func TestDebeziumCDCWrap(t *testing.T) {
	cfg := &pb.CDCEnvelope{
		Enabled:       true,
		ConnectorType: "postgresql",
		DatabaseName:  "ecom_prod",
		SchemaName:    "orders",
		TableName:     "order_items",
		InsertRatio:   1.0, // 100% create for test
	}

	env := NewDebeziumEnvelope(cfg)
	rng := state.NewDeterministicRand(42)
	clock := state.NewClock(time.Now(), 100*time.Millisecond)

	record := map[string]any{
		"order_id": "ord_101",
		"price":    29.99,
	}

	wrapped := env.Wrap(record, 1, rng, clock)

	if wrapped["op"] != "c" {
		t.Fatalf("expected op 'c', got %v", wrapped["op"])
	}

	after, ok := wrapped["after"].(map[string]any)
	if !ok || after["order_id"] != "ord_101" {
		t.Fatalf("expected after record to contain order_id, got %v", wrapped["after"])
	}

	source, ok := wrapped["source"].(map[string]any)
	if !ok || source["connector"] != "postgresql" || source["db"] != "ecom_prod" {
		t.Fatalf("unexpected source metadata: %v", source)
	}
}

func TestDebeziumDeleteOperation(t *testing.T) {
	cfg := &pb.CDCEnvelope{
		Enabled:       true,
		ConnectorType: "mysql",
		DatabaseName:  "inventory",
		TableName:     "stocks",
		InsertRatio:   0.0,
		UpdateRatio:   0.0,
		DeleteRatio:   1.0, // 100% delete
	}

	env := NewDebeziumEnvelope(cfg)
	rng := state.NewDeterministicRand(99)
	clock := state.NewClock(time.Now(), 100*time.Millisecond)

	record := map[string]any{"sku": "SKU-99"}
	wrapped := env.Wrap(record, 2, rng, clock)

	if wrapped["op"] != "d" {
		t.Fatalf("expected op 'd', got %v", wrapped["op"])
	}
	if wrapped["after"] != nil {
		t.Fatalf("expected after to be nil for delete, got %v", wrapped["after"])
	}
	if wrapped["before"] == nil {
		t.Fatalf("expected before to contain deleted payload")
	}
}

