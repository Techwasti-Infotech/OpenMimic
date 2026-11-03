package traffic

import (
	"math"
	"time"

	"openmimic/engine/internal/state"
	pb "openmimic/engine/pkg/proto/v1"
)

// Shaper determines the delay until the next event generation.
type Shaper interface {
	NextDelay(elapsed time.Duration, rng *state.DeterministicRand) time.Duration
}

// ConstantShaper delivers events at a steady rate.
type ConstantShaper struct {
	rate int32
}

func (s *ConstantShaper) NextDelay(elapsed time.Duration, rng *state.DeterministicRand) time.Duration {
	if s.rate <= 0 {
		return 0
	}
	return time.Second / time.Duration(s.rate)
}

// DiurnalShaper simulates day/night cyclical traffic via sine waves.
type DiurnalShaper struct {
	baseRate int32
	peakRate int32
	period   time.Duration
}

func (s *DiurnalShaper) NextDelay(elapsed time.Duration, rng *state.DeterministicRand) time.Duration {
	if s.period <= 0 {
		s.period = 60 * time.Second
	}
	// Normalized position in cycle [0, 2*PI]
	phase := (float64(elapsed%s.period) / float64(s.period)) * 2 * math.Pi
	// Sine factor between 0.0 and 1.0
	factor := (math.Sin(phase) + 1.0) / 2.0
	currentRate := float64(s.baseRate) + factor*float64(s.peakRate-s.baseRate)
	if currentRate <= 0 {
		currentRate = 1.0
	}
	return time.Duration(float64(time.Second) / currentRate)
}

// PoissonShaper produces exponential inter-arrival times for stochastic burst modeling.
type PoissonShaper struct {
	lambda float64
}

func (s *PoissonShaper) NextDelay(elapsed time.Duration, rng *state.DeterministicRand) time.Duration {
	if s.lambda <= 0 {
		return 0
	}
	// Inverse transform sampling for Exponential distribution: -ln(U) / lambda
	u := rng.Float64()
	if u <= 0.0000001 {
		u = 0.0000001
	}
	interArrivalTimeSeconds := -math.Log(u) / s.lambda
	return time.Duration(interArrivalTimeSeconds * float64(time.Second))
}

// BurstShaper simulates periodic flash spikes (e.g. flash sales, sudden demand surges).
type BurstShaper struct {
	normalRate int32
	burstRate  int32
	interval   time.Duration
	duration   time.Duration
}

func (s *BurstShaper) NextDelay(elapsed time.Duration, rng *state.DeterministicRand) time.Duration {
	if s.interval <= 0 {
		s.interval = 30 * time.Second
	}
	if s.duration <= 0 {
		s.duration = 5 * time.Second
	}

	cyclePos := elapsed % s.interval
	rate := s.normalRate
	if cyclePos < s.duration {
		rate = s.burstRate
	}
	if rate <= 0 {
		return 0
	}
	return time.Second / time.Duration(rate)
}

// NewShaper constructs a traffic shaper from the TrafficProfile or legacy rate limit.
func NewShaper(profile *pb.TrafficProfile, legacyRate int32) Shaper {
	if profile != nil && profile.ProfileType != nil {
		switch p := profile.ProfileType.(type) {
		case *pb.TrafficProfile_Constant:
			return &ConstantShaper{rate: p.Constant.RatePerSecond}
		case *pb.TrafficProfile_Diurnal:
			return &DiurnalShaper{
				baseRate: p.Diurnal.BaseRate,
				peakRate: p.Diurnal.PeakRate,
				period:   time.Duration(p.Diurnal.PeriodSeconds) * time.Second,
			}
		case *pb.TrafficProfile_Poisson:
			return &PoissonShaper{lambda: p.Poisson.LambdaRate}
		case *pb.TrafficProfile_Burst:
			return &BurstShaper{
				normalRate: p.Burst.NormalRate,
				burstRate:  p.Burst.BurstRate,
				interval:   time.Duration(p.Burst.IntervalSeconds) * time.Second,
				duration:   time.Duration(p.Burst.DurationSeconds) * time.Second,
			}
		}
	}

	if legacyRate > 0 {
		return &ConstantShaper{rate: legacyRate}
	}
	return &ConstantShaper{rate: 0}
}

