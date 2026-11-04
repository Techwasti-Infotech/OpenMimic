package timeseries

import (
	"math"
	"testing"
	"time"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

func TestDriftGeneratorSeasonality(t *testing.T) {
	rng := state.NewDeterministicRand(42)
	cfg := &pb.TimeSeriesDriftFunction{
		BaseValue:            100.0,
		TrendSlope:           0.5,
		SeasonalityAmplitude: 20.0,
		SeasonalityPeriod:    40,
		NoiseStdDev:          0.0, // pure deterministic
	}

	gen := NewDriftGenerator(cfg)

	// At t = 0: 100 + 0 + 20*sin(0) = 100
	v0 := gen.Generate(0, rng)
	if math.Abs(v0-100.0) > 0.001 {
		t.Errorf("expected 100 at t=0, got %f", v0)
	}

	// At t = 10 (quarter period): sin(pi/2) = 1.0 -> 100 + 10*0.5 + 20*1 = 125
	v10 := gen.Generate(10, rng)
	if math.Abs(v10-125.0) > 0.001 {
		t.Errorf("expected 125 at t=10, got %f", v10)
	}
}

func TestSessionWindowExpiration(t *testing.T) {
	cfg := &pb.SessionWindowFunction{
		InactivityTimeoutSeconds: 300, // 5 minutes
		SessionPrefix:            "user_sess_",
	}

	tracker := NewSessionTracker(cfg)
	baseTime := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)

	// First event for user_1
	sess1 := tracker.GetOrCreateSession("user_1", baseTime)
	if sess1 != "user_sess_00001001" {
		t.Errorf("unexpected initial session id: %s", sess1)
	}

	// Active interaction 2 minutes later -> same session
	sess2 := tracker.GetOrCreateSession("user_1", baseTime.Add(2*time.Minute))
	if sess2 != sess1 {
		t.Errorf("expected same session within inactivity window, got %s", sess2)
	}

	// Inactivity of 10 minutes (exceeds 5m timeout) -> rotates to new session
	sess3 := tracker.GetOrCreateSession("user_1", baseTime.Add(12*time.Minute))
	if sess3 == sess1 {
		t.Errorf("expected rotated session after timeout, got same %s", sess3)
	}
}

