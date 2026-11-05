package traffic

import (
	"testing"
	"time"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

func TestBlackFridayFlashCrowdShaperPhases(t *testing.T) {
	rng := state.NewDeterministicRand(42)
	cfg := &pb.TrafficProfile{
		ProfileType: &pb.TrafficProfile_FlashCrowd{
			FlashCrowd: &pb.BlackFridayFlashCrowd{
				BaselineRate:    10,
				SurgePeakRate:   100,
				RampUpSeconds:   10,
				PlateauSeconds:  20,
				RampDownSeconds: 10,
			},
		},
	}

	shaper := NewShaper(cfg, 0)

	// At 0s (baseline start): 10 eps -> delay 100ms
	d0 := shaper.NextDelay(0, rng)
	if d0 < 90*time.Millisecond || d0 > 110*time.Millisecond {
		t.Errorf("expected ~100ms at baseline start, got %v", d0)
	}

	// At 15s (during peak plateau): 100 eps -> delay 10ms
	dPeak := shaper.NextDelay(15*time.Second, rng)
	if dPeak < 9*time.Millisecond || dPeak > 11*time.Millisecond {
		t.Errorf("expected ~10ms during surge peak plateau, got %v", dPeak)
	}
}
