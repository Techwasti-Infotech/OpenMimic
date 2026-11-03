package fidelity

import (
	"math"
	"math/rand"
	"testing"
)

func TestKolmogorovSmirnovFidelity(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	n := 1000

	ref := make([]float64, n)
	syntheticMatching := make([]float64, n)
	syntheticDivergent := make([]float64, n)

	for i := 0; i < n; i++ {
		ref[i] = rng.NormFloat64()*15.0 + 100.0
		syntheticMatching[i] = rng.NormFloat64()*15.0 + 100.0
		syntheticDivergent[i] = rng.NormFloat64()*5.0 + 50.0 // clearly different distribution
	}

	dMatching := KolmogorovSmirnovTest(ref, syntheticMatching)
	dDivergent := KolmogorovSmirnovTest(ref, syntheticDivergent)

	if dMatching > 0.15 {
		t.Fatalf("expected low KS distance for matching distributions, got %f", dMatching)
	}

	if dDivergent < 0.70 {
		t.Fatalf("expected high KS distance for divergent distributions, got %f", dDivergent)
	}
}

func TestPearsonCorrelationFidelity(t *testing.T) {
	n := 100
	x := make([]float64, n)
	y := make([]float64, n)

	for i := 0; i < n; i++ {
		x[i] = float64(i)
		y[i] = float64(i)*2.0 + 5.0 // Perfect linear relationship
	}

	r, err := PearsonCorrelation(x, y)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if math.Abs(r-1.0) > 0.0001 {
		t.Fatalf("expected r ~ 1.0, got %f", r)
	}
}

