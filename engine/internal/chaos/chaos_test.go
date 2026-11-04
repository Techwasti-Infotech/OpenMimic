package chaos

import (
	"context"
	"testing"
	"time"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

func TestChaosSimulatorFailureInjection(t *testing.T) {
	rng := state.NewDeterministicRand(42)
	cfg := &pb.ChaosConfig{
		Enabled:     true,
		FailureRate: 1.0, // 100% simulated drop
	}

	sim := NewChaosSimulator(cfg)
	err := sim.Intercept(context.Background(), rng)
	if err == nil {
		t.Fatalf("expected simulated network partition error, got nil")
	}
}

func TestChaosSimulatorLatencyDelay(t *testing.T) {
	rng := state.NewDeterministicRand(99)
	cfg := &pb.ChaosConfig{
		Enabled:      true,
		MinLatencyMs: 15,
		MaxLatencyMs: 25,
	}

	sim := NewChaosSimulator(cfg)
	start := time.Now()
	err := sim.Intercept(context.Background(), rng)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed < 10*time.Millisecond {
		t.Errorf("expected at least ~15ms delay, got %v", elapsed)
	}
}

