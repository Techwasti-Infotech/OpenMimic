package traffic

import (
	"testing"
	"time"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

func TestTrafficShaperProfiles(t *testing.T) {
	rng := state.NewDeterministicRand(42)

	// 1. Constant Shaper
	constShaper := NewShaper(&pb.TrafficProfile{
		ProfileType: &pb.TrafficProfile_Constant{
			Constant: &pb.ConstantRate{RatePerSecond: 20},
		},
	}, 0)
	delay := constShaper.NextDelay(0, rng)
	if delay != 50*time.Millisecond {
		t.Fatalf("expected constant delay 50ms, got %v", delay)
	}

	// 2. Diurnal Shaper (test trough vs peak)
	diurnalShaper := NewShaper(&pb.TrafficProfile{
		ProfileType: &pb.TrafficProfile_Diurnal{
			Diurnal: &pb.DiurnalRate{
				BaseRate:      10,
				PeakRate:      100,
				PeriodSeconds: 10,
			},
		},
	}, 0)
	delayTrough := diurnalShaper.NextDelay(10*time.Second, rng)
	delayPeak := diurnalShaper.NextDelay(2500*time.Millisecond, rng)
	if delayPeak >= delayTrough {
		t.Fatalf("expected peak delay (%v) to be shorter than trough delay (%v)", delayPeak, delayTrough)
	}

	// 3. Poisson Shaper
	poissonShaper := NewShaper(&pb.TrafficProfile{
		ProfileType: &pb.TrafficProfile_Poisson{
			Poisson: &pb.PoissonArrival{LambdaRate: 50.0},
		},
	}, 0)
	var totalDelay time.Duration
	for i := 0; i < 100; i++ {
		totalDelay += poissonShaper.NextDelay(0, rng)
	}
	avgDelayMs := float64(totalDelay.Milliseconds()) / 100.0
	// Expected ~20ms
	if avgDelayMs < 10.0 || avgDelayMs > 35.0 {
		t.Fatalf("expected average delay ~20ms, got %fms", avgDelayMs)
	}

	// 4. Burst Shaper
	burstShaper := NewShaper(&pb.TrafficProfile{
		ProfileType: &pb.TrafficProfile_Burst{
			Burst: &pb.BurstSpikeRate{
				NormalRate:      10,
				BurstRate:       100,
				IntervalSeconds: 10,
				DurationSeconds: 2,
			},
		},
	}, 0)
	inBurstDelay := burstShaper.NextDelay(1*time.Second, rng)
	outBurstDelay := burstShaper.NextDelay(5*time.Second, rng)
	if inBurstDelay != 10*time.Millisecond || outBurstDelay != 100*time.Millisecond {
		t.Fatalf("burst delay mismatch: in=%v out=%v", inBurstDelay, outBurstDelay)
	}
}

