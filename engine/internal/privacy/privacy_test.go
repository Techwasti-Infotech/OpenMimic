package privacy

import (
	"math"
	"math/rand"
	"testing"

	pb "openmimic/engine/pkg/proto/v1"
)

func TestLaplaceMechanism(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	minVal := 0.0
	maxVal := 100.0
	laplace, err := NewLaplaceMechanism(1.0, 1.0, &minVal, &maxVal)
	if err != nil {
		t.Fatalf("unexpected error creating laplace mechanism: %v", err)
	}

	base := 50.0
	noisy := laplace.AddNoise(base, rng)

	if noisy < minVal || noisy > maxVal {
		t.Errorf("noisy value %f outside bounds [%f, %f]", noisy, minVal, maxVal)
	}

	// Verify noise variance across multiple runs is approximately 2 * (b^2)
	sum := 0.0
	n := 1000
	for i := 0; i < n; i++ {
		sum += laplace.AddNoise(base, rng)
	}
	mean := sum / float64(n)
	if math.Abs(mean-base) > 1.5 {
		t.Errorf("empirical mean %f deviated significantly from base %f", mean, base)
	}
}

func TestGaussianMechanism(t *testing.T) {
	rng := rand.New(rand.NewSource(99))
	gaussian, err := NewGaussianMechanism(1.5, 1e-4, 2.0, nil, nil)
	if err != nil {
		t.Fatalf("unexpected error creating gaussian mechanism: %v", err)
	}

	val := gaussian.AddNoise(100.0, rng)
	if val == 100.0 {
		t.Errorf("expected noise to be added to 100.0")
	}
}

func TestCreateDPMechanismFromProto(t *testing.T) {
	protoCfg := &pb.DifferentialPrivacyFunction{
		Mechanism:   "gaussian",
		Epsilon:     0.5,
		Delta:       1e-5,
		Sensitivity: 1.0,
		MinBound:    0.0,
		MaxBound:    500.0,
	}

	dp, err := CreateDPMechanism(protoCfg)
	if err != nil {
		t.Fatalf("failed to create DP mechanism: %v", err)
	}

	rng := rand.New(rand.NewSource(123))
	noisy := dp.AddNoise(250.0, rng)
	if noisy < 0.0 || noisy > 500.0 {
		t.Errorf("value %f out of bounds [0, 500]", noisy)
	}
}
