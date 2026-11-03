package privacy

import (
	"fmt"
	"math"

	pb "openmimic/engine/pkg/proto/v1"
)

// RandomSource defines the minimal random operations needed for noise injection.
type RandomSource interface {
	Float64() float64
	NormFloat64() float64
}

// DPMechanism defines an interface for adding calibrated privacy noise.
type DPMechanism interface {
	AddNoise(val float64, rng RandomSource) float64
}

// LaplaceMechanism implements epsilon-differential privacy using the Laplace distribution:
// Noise ~ Lap(Delta f / epsilon)
type LaplaceMechanism struct {
	Epsilon     float64
	Sensitivity float64
	MinBound    *float64
	MaxBound    *float64
}

// NewLaplaceMechanism creates a calibrated Laplace DP mechanism.
func NewLaplaceMechanism(epsilon, sensitivity float64, minBound, maxBound *float64) (*LaplaceMechanism, error) {
	if epsilon <= 0 {
		return nil, fmt.Errorf("epsilon must be positive, got %f", epsilon)
	}
	if sensitivity <= 0 {
		return nil, fmt.Errorf("sensitivity must be positive, got %f", sensitivity)
	}
	return &LaplaceMechanism{
		Epsilon:     epsilon,
		Sensitivity: sensitivity,
		MinBound:    minBound,
		MaxBound:    maxBound,
	}, nil
}

// AddNoise adds Laplace noise to the given value:
// Scale b = sensitivity / epsilon
func (m *LaplaceMechanism) AddNoise(val float64, rng RandomSource) float64 {
	b := m.Sensitivity / m.Epsilon
	// Sample from Laplace distribution using inverse CDF:
	// u in (-0.5, 0.5)
	u := rng.Float64() - 0.5
	sgn := 1.0
	if u < 0 {
		sgn = -1.0
	}
	noise := -b * sgn * math.Log(1.0-2.0*math.Abs(u))
	res := val + noise

	if m.MinBound != nil && res < *m.MinBound {
		res = *m.MinBound
	}
	if m.MaxBound != nil && res > *m.MaxBound {
		res = *m.MaxBound
	}
	return res
}

// GaussianMechanism implements (epsilon, delta)-differential privacy using Gaussian noise:
// sigma >= sqrt(2 * ln(1.25 / delta)) * (sensitivity / epsilon)
type GaussianMechanism struct {
	Epsilon     float64
	Delta       float64
	Sensitivity float64
	MinBound    *float64
	MaxBound    *float64
}

// NewGaussianMechanism creates a calibrated Gaussian DP mechanism.
func NewGaussianMechanism(epsilon, delta, sensitivity float64, minBound, maxBound *float64) (*GaussianMechanism, error) {
	if epsilon <= 0 {
		return nil, fmt.Errorf("epsilon must be positive, got %f", epsilon)
	}
	if delta <= 0 || delta >= 1.0 {
		return nil, fmt.Errorf("delta must be in (0, 1), got %f", delta)
	}
	if sensitivity <= 0 {
		return nil, fmt.Errorf("sensitivity must be positive, got %f", sensitivity)
	}
	return &GaussianMechanism{
		Epsilon:     epsilon,
		Delta:       delta,
		Sensitivity: sensitivity,
		MinBound:    minBound,
		MaxBound:    maxBound,
	}, nil
}

// AddNoise adds Gaussian noise with scale sigma.
func (m *GaussianMechanism) AddNoise(val float64, rng RandomSource) float64 {
	sigma := math.Sqrt(2.0*math.Log(1.25/m.Delta)) * (m.Sensitivity / m.Epsilon)
	noise := rng.NormFloat64() * sigma
	res := val + noise

	if m.MinBound != nil && res < *m.MinBound {
		res = *m.MinBound
	}
	if m.MaxBound != nil && res > *m.MaxBound {
		res = *m.MaxBound
	}
	return res
}

// CreateDPMechanism builds the mechanism from protobuf configuration.
func CreateDPMechanism(cfg *pb.DifferentialPrivacyFunction) (DPMechanism, error) {
	if cfg == nil {
		return nil, fmt.Errorf("differential privacy config is nil")
	}

	sens := cfg.GetSensitivity()
	if sens <= 0 {
		sens = 1.0 // default global sensitivity
	}

	eps := cfg.GetEpsilon()
	if eps <= 0 {
		eps = 1.0 // standard default epsilon
	}

	var minB, maxB *float64
	if cfg.GetMinBound() != 0 || cfg.GetMaxBound() != 0 {
		min := cfg.GetMinBound()
		max := cfg.GetMaxBound()
		minB = &min
		maxB = &max
	}

	switch cfg.GetMechanism() {
	case "gaussian":
		delta := cfg.GetDelta()
		if delta <= 0 {
			delta = 1e-5 // default delta
		}
		return NewGaussianMechanism(eps, delta, sens, minB, maxB)
	case "laplace":
		fallthrough
	default:
		return NewLaplaceMechanism(eps, sens, minB, maxB)
	}
}
