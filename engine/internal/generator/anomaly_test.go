package generator

import (
	"testing"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

func TestAnomalyInjector(t *testing.T) {
	playbooks := []*pb.AnomalyPlaybook{
		{
			Name:          "Nullify Test",
			AnomalyRate:   1.0, // 100% injection for deterministic verification
			InjectionType: "null_field",
		},
	}

	injector := NewAnomalyInjector(playbooks)
	rng := state.NewDeterministicRand(123)

	record := map[string]any{
		"user_id": "test-123",
	}

	mutatedRecord, injected := injector.Apply(record, rng)
	if !injected {
		t.Fatalf("expected anomaly to be injected")
	}

	if mutatedRecord["user_id"] != nil {
		t.Fatalf("expected field to be nullified, got %v", mutatedRecord["user_id"])
	}
}

func TestAnomalyInjectorCorruptPayload(t *testing.T) {
	playbooks := []*pb.AnomalyPlaybook{
		{
			Name:          "Corrupt Test",
			AnomalyRate:   1.0,
			InjectionType: "corrupt_payload",
		},
	}

	injector := NewAnomalyInjector(playbooks)
	rng := state.NewDeterministicRand(456)

	record := map[string]any{
		"amount": 100.5,
	}

	mutated, injected := injector.Apply(record, rng)
	if !injected || mutated["_corrupt"] != true {
		t.Fatalf("expected corrupt anomaly injected")
	}
}

