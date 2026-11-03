package dlq

import (
	"context"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestDLQQuarantineRouting(t *testing.T) {
	cfg := &pb.DLQConfig{
		Enabled:             true,
		TargetTopicOrTable: "orders.dlq",
		IncludeErrorStack:  true,
		DropOnOverflow:      true,
	}

	router := NewRouter(cfg)
	ctx := context.Background()

	badPayload := map[string]any{
		"order_id": "bad_1",
		"amount":   -50.0,
	}

	err := router.Quarantine(ctx, "gen-orders", "Predicate violated: amount > 0", 100, badPayload)
	if err != nil {
		t.Fatalf("unexpected quarantine error: %v", err)
	}

	if router.Count() != 1 {
		t.Fatalf("expected 1 quarantined record, got %d", router.Count())
	}

	quarantined := router.GetQuarantined()
	if quarantined[0].Reason != "Predicate violated: amount > 0" {
		t.Fatalf("unexpected reason: %s", quarantined[0].Reason)
	}
	if quarantined[0].Payload["order_id"] != "bad_1" {
		t.Fatalf("unexpected payload in DLQ: %v", quarantined[0].Payload)
	}
}

