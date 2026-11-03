package profiler

import (
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestDatabaseProfiler(t *testing.T) {
	sampleRows := []map[string]any{
		{"order_id": "ord_1", "amount": 100.0, "status": "COMPLETED", "created_at": "2026-01-01T12:00:00Z"},
		{"order_id": "ord_2", "amount": 110.0, "status": "COMPLETED", "created_at": "2026-01-01T12:05:00Z"},
		{"order_id": "ord_3", "amount": 90.0, "status": "FAILED", "created_at": "2026-01-01T12:10:00Z"},
		{"order_id": "ord_4", "amount": 105.0, "status": "COMPLETED", "created_at": "2026-01-01T12:15:00Z"},
	}

	p := NewDatabaseProfiler("orders", sampleRows)
	cfg, err := p.Profile()
	if err != nil {
		t.Fatalf("failed to profile orders table: %v", err)
	}

	if len(cfg.GetFields()) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(cfg.GetFields()))
	}

	// Verify inferred types
	fieldMap := make(map[string]string)
	for _, f := range cfg.GetFields() {
		switch f.GetGenerator().GetFunc().(type) {
		case *pb.FunctionGenerator_Uuid:
			fieldMap[f.GetName()] = "uuid"
		case *pb.FunctionGenerator_Distribution:
			fieldMap[f.GetName()] = "distribution"
		case *pb.FunctionGenerator_WeightedChoice:
			fieldMap[f.GetName()] = "weighted_choice"
		case *pb.FunctionGenerator_Now:
			fieldMap[f.GetName()] = "now"
		}
	}

	if fieldMap["order_id"] != "uuid" {
		t.Fatalf("expected order_id to be uuid, got %s", fieldMap["order_id"])
	}
	if fieldMap["amount"] != "distribution" {
		t.Fatalf("expected amount to be normal distribution, got %s", fieldMap["amount"])
	}
	if fieldMap["status"] != "weighted_choice" {
		t.Fatalf("expected status to be weighted_choice, got %s", fieldMap["status"])
	}
	if fieldMap["created_at"] != "now" {
		t.Fatalf("expected created_at to be now, got %s", fieldMap["created_at"])
	}
}
